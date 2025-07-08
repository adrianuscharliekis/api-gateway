package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func BodyCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				// Handle error, maybe return a 400 Bad Request
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
				return
			}
			// After reading, put the bytes back into the request body so it can be read again
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			// Store the cached body in the context for easy access
			c.Set("cachedBody", bodyBytes)
		}
		c.Next()
	}
}
