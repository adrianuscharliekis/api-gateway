package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func SecureProxy() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		timestamp := c.GetHeader("X-TIMESTAMP")
		clientSignatureB64 := c.GetHeader("X-SIGNATURE")

		if authHeader == "" || timestamp == "" || clientSignatureB64 == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing required headers (Authorization, X-TIMESTAMP, X-SIGNATURE)"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			return
		}
		accessToken := parts[1]

		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		minifiedBody := minifyJSON(bodyBytes)
		bodyHash := sha256.Sum256([]byte(minifiedBody))
		encodedBody := strings.ToLower(hex.EncodeToString(bodyHash[:]))

		httpMethod := c.Request.Method
		endpointURL := c.Request.URL.Path
		stringToSign := fmt.Sprintf("%s:%s:%s:%s:%s", httpMethod, endpointURL, accessToken, encodedBody, timestamp)

		token, _, err := new(jwt.Parser).ParseUnverified(accessToken, jwt.MapClaims{})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid access token"})
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Could not parse token claims"})
			return
		}
		clientID, ok := claims["sub"].(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Client ID not found in token"})
			return
		}

		clientSecret, err := getClientSecret(clientID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Client not registered or secret not found"})
			return
		}

		mac := hmac.New(sha512.New, []byte(clientSecret))
		mac.Write([]byte(stringToSign))
		expectedSignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(expectedSignature), []byte(clientSignatureB64)) {
			fmt.Printf("Signature Mismatch!\nClient Sent: %s\nServer Expected: %s\nString Signed: %s\n", clientSignatureB64, expectedSignature, stringToSign)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Signature"})
			return
		}

		c.Next()
	}
}

func getClientSecret(clientID string) (string, error) {
	// Example static mapping for demonstration purposes.
	secrets := map[string]string{
		"C00001": "super-secret-for-client-001",
		"C00002": "another-secret-for-client-002",
	}
	secret, ok := secrets[clientID]
	if !ok {
		return "", fmt.Errorf("client secret for '%s' not found", clientID)
	}
	return secret, nil
}

func minifyJSON(jsonBytes []byte) string {
	buffer := new(bytes.Buffer)
	if err := json.Compact(buffer, jsonBytes); err != nil {
		s := string(jsonBytes)
		s = strings.ReplaceAll(s, " ", "")
		s = strings.ReplaceAll(s, "\n", "")
		s = strings.ReplaceAll(s, "\r", "")
		s = strings.ReplaceAll(s, "\t", "")
		return s
	}
	return buffer.String()
}
