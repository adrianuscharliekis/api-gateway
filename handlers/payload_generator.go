package handlers

import (
	"api-gateway/model"
	"api-gateway/request"
	"api-gateway/response"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/url" // Import the 'net/url' package
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// generateSignature is the core logic function.
func generateSignature(clientID, privateKeyPath, redirect string) (*response.SignatureResponse, error) {
	// 1. Generate timestamp in RFC3339 format with timezone
	timestamp := time.Now().Format("2006-01-02T15:04:05-07:00")

	// 2. Generate externalID (UUID)
	externalID := uuid.New().String()

	// 3. String to be signed: clientID|timestamp|externalID
	stringToSign := fmt.Sprintf("%s|%s|%s", clientID, timestamp, externalID)

	// 4. Read the private key
	privateKeyPEM, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	var privateKey *rsa.PrivateKey
	parsedKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err == nil {
		privateKey = parsedKey
	} else {
		pkcs8Key, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("failed to parse private key: %v, %v", err, err2)
		}
		var ok bool
		privateKey, ok = pkcs8Key.(*rsa.PrivateKey)
		if !ok {
			return nil, err2
		}
	}

	// 5. Hash the string using SHA256
	hashed := sha256.Sum256([]byte(stringToSign))

	// 6. Sign the hash with RSA private key
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	// 7. Encode the signature using URL-safe Base64
	encodedSignature := base64.StdEncoding.EncodeToString(signature)

	// 8. Build the secure URL with properly escaped query params
	params := url.Values{}
	params.Set("ca_code", clientID)
	params.Set("signature", encodedSignature)
	params.Set("timestamp", timestamp)
	params.Set("product", redirect)
	params.Set("externalId", externalID)

	link := fmt.Sprintf("/auth/login?%s", params.Encode())
	before := fmt.Sprintf("/auth/login?ca_code=%s&externalId=%s&product=%s&signature=%s&timestamp=%s", clientID, externalID, redirect, encodedSignature, timestamp)

	// 9. Return the response object
	return &response.SignatureResponse{
		ClientID:   clientID,
		Timestamp:  timestamp,
		Signature:  encodedSignature,
		Link:       link,
		BeforeLink: before,
		ExternalID: externalID,
	}, nil
}

// GenerateSignatureHandler is the Gin handler that wraps the core logic.
func GenerateSignatureHandler(c *gin.Context) {
	var req request.PayloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	config, err := model.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load server configuration: " + err.Error()})
		return
	}

	clientConf, ok := config.Clients[req.ClientID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Configuration for client_id '%s' not found", req.ClientID)})
		return
	}

	response, err := generateSignature(req.ClientID, clientConf.PrivateKeyPath, req.Redirect)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	helper := config.Helper
	response.Link = fmt.Sprintf("%s%s", helper["secure_page_port"], response.Link)
	response.BeforeLink = fmt.Sprintf("%s%s", helper["secure_page_port"], response.BeforeLink)

	c.JSON(http.StatusOK, response)
}
