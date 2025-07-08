package middleware

import (
	"api-gateway/utils"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTAuthMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {
		timestamp := c.GetHeader("X-TIMESTAMP")
		authHeader := c.GetHeader("Authorization")
		clientID := c.GetHeader("X-PARTNER-ID")
		externalID := c.GetHeader("X-EXTERNAL-ID")
		// signature := c.GetHeader("X-SIGNATURE")
		fmt.Println(timestamp, authHeader, clientID, externalID)

		if timestamp == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Timestamp"})
			return
		}
		if clientID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing ClientID"})
			return
		}
		if externalID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing ExternalID"})
			return
		}
		// if signature == "" {
		// 	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Access Token"})
		// 	return
		// }

		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Access Token"})
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be in 'Bearer <token>' format"})
			return
		}
		tokenString := parts[1]
		_, err := utils.VerifyJWT(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Access Token"})
			return
		}
		c.Next()
	}
}
