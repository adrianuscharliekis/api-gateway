package handlers

import (
	"api-gateway/model"
	"api-gateway/request"
	"api-gateway/response"
	"api-gateway/services"
	"api-gateway/utils"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	tracelog        services.TracelogServices
	externalIDStore *utils.ExternalIDStore
}

func NewAuthHandler(s services.TracelogServices, store *utils.ExternalIDStore) *AuthHandler {
	return &AuthHandler{tracelog: s, externalIDStore: store}
}

// --- Core Verification Logic ---

// verifySignature checks if the provided signature is valid for the given data.
func verifySignature(publicKeyPath, stringToVerify, base64Signature string) error {
	// 1. Read the public key PEM file
	publicKeyPEM, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key file: %w", err)
	}

	// 2. Decode the PEM block to get the key bytes
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return errors.New("failed to decode PEM block containing public key")
	}

	// 3. Parse the key bytes into an RSA public key
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}
	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return errors.New("key is not a valid RSA public key")
	}

	// 4. Decode the Base64 signature from the header
	// *** FIX: Use URLEncoding instead of StdEncoding. This is crucial for data
	// transmitted in HTTP headers, as it avoids issues with '+' and '/' characters.
	// The client that generates the signature MUST also use URLEncoding.
	signature, err := base64.URLEncoding.DecodeString(base64Signature)
	if err != nil {
		return fmt.Errorf("failed to decode base64 signature: %w", err)
	}

	// 5. Hash the original data string using SHA256
	hashed := sha256.Sum256([]byte(stringToVerify))

	// 6. Verify the signature against the hash
	// This is the core cryptographic operation.
	err = rsa.VerifyPKCS1v15(rsaPublicKey, crypto.SHA256, hashed[:], signature)
	if err != nil {
		// This error means the signature is invalid or tampered with.
		return errors.New("signature verification failed")
	}

	return nil // Signature is valid
}

// --- Gin Handler ---

// Login is the handler for the /auth/token endpoint, following ASPI SNAP specs.
func (h AuthHandler) Login(c *gin.Context) {
	// 1. Extract required headers
	timestampStr := c.GetHeader("X-TIMESTAMP")
	clientKey := c.GetHeader("X-CLIENT-KEY")
	signature := c.GetHeader("X-SIGNATURE")
	productType := c.GetHeader("X-PRODUCT-ID")
	externalID := c.GetHeader("X-EXTERNAL-ID")
	logStr := fmt.Sprintf("X-TIMESTAMP=%s | X-CLIENT-KEY=%s | X-SIGNATURE=%s | X-EXTERNAL-ID=%s", timestampStr, clientKey, signature, externalID)
	go h.tracelog.Log("LOGIN", clientKey, externalID, logStr)

	if timestampStr == "" || clientKey == "" || signature == "" || externalID == "" || productType == "" {
		h.tracelog.Log("LOGIN", clientKey, externalID, "Missing Required Headers")
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			ResponseCode:    "401",
			ResponseMessage: "Missing Required Headers (X-TIMESTAMP, X-CLIENT-KEY, X-SIGNATURE, X-EXTERNAL-ID,X-PRODUCT-ID)",
		})
		return
	}

	// 2. Parse request body
	var req request.JwtRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		go h.tracelog.Log("LOGIN", clientKey, productType, "Invalid request body :"+err.Error())
		c.JSON(http.StatusBadRequest, response.ErrorResponse{ResponseCode: "400", ResponseMessage: "Invalid request body: " + err.Error()})
		return
	}

	// 3. Validate timestamp to prevent replay attacks
	requestTime, err := time.Parse("2006-01-02T15:04:05-07:00", timestampStr)
	if err != nil {
		go h.tracelog.Log("LOGIN", clientKey, productType, "Invalid X-TIMESTAMP format")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			ResponseCode:    "400",
			ResponseMessage: "Invalid X-TIMESTAMP format." + err.Error(),
		})
		return
	}
	// Allow a 5-minute window
	if time.Since(requestTime).Abs() > 5*time.Minute {
		go h.tracelog.Log("LOGIN", clientKey, productType, "Request timestamp is too old or too far in the future.")
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{ResponseCode: "401", ResponseMessage: "Request timestamp is too old or too far in the future."})
		return
	}

	// 4. Load configuration and find the correct public key
	config, err := model.LoadConfig()
	if err != nil {
		go h.tracelog.Log("LOGIN", clientKey, productType, "Server configuration error.")
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{ResponseCode: "500", ResponseMessage: "Server configuration error."})
		return
	}
	clientConf, ok := config.Clients[clientKey]
	if !ok {
		go h.tracelog.Log("LOGIN", clientKey, productType, fmt.Sprintf("Client with key '%s' not registered.", clientKey))
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{ResponseCode: "401", ResponseMessage: fmt.Sprintf("Client with key '%s' not registered.", clientKey)})
		return
	}
	// Check if externalID has been seen before
	if h.externalIDStore.ExistsAndValid(externalID) {
		h.tracelog.Log("LOGIN", clientKey, externalID, "Replay attack detected: externalID reused")
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			ResponseCode:    "401",
			ResponseMessage: "Replay attack detected: externalID already used",
		})
		return
	}

	// Store the externalID as new
	h.externalIDStore.Save(externalID)

	// 5. Reconstruct the string that was signed by the client
	stringToVerify := fmt.Sprintf("%s|%s|%s", clientKey, timestampStr, externalID)

	// 6. Verify the digital signature
	err = verifySignature(clientConf.PublicKeyPath, stringToVerify, signature)
	if err != nil {
		// Log the detailed error for debugging, but return a generic error to the user.
		go h.tracelog.Log("LOGIN", clientKey, productType, "Invalid Signature :"+err.Error())
		fmt.Printf("Signature verification failed for client %s: %v\n", clientKey, err)
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{ResponseCode: "401", ResponseMessage: "Invalid Signature : " + err.Error()})
		return
	}

	// 7. If signature is valid, generate the Access Token
	// The token is signed with the API Gateway's own private key (handled by your util).
	// The token's subject ('sub') should be the client_id that was just authenticated.
	accessToken, err := utils.GenerateJWT(clientKey)
	if err != nil {
		go h.tracelog.Log("LOGIN", clientKey, productType, "Failed to generate access token.")
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{ResponseCode: "500", ResponseMessage: "Failed to generate access token."})
		return
	}

	// 8. Return the successful response as per ASPI SNAP spec
	go h.tracelog.Log("LOGIN", clientKey, productType, "Success generate accesstoken :"+accessToken)
	c.JSON(http.StatusOK, response.SuccessResponse{ResponseCode: "200", ResponseMessage: "Successful", AdditionalInfo: map[string]string{
		"accessToken": accessToken,
		"tokenType":   "Bearer",
		"expiresIn":   "900",
	},
	})
}
