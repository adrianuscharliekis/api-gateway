package handlers

import (
	"api-gateway/services"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ProxyHandler holds dependencies for the proxy logic.
type ProxyHandler struct {
	tracelog services.TracelogServices
}

// NewProxyHandler creates a new instance of the proxy handler.
func NewProxyHandler(s services.TracelogServices) *ProxyHandler {
	return &ProxyHandler{tracelog: s}
}

// buildRequestLogString constructs a single string containing all relevant request details.
func (h *ProxyHandler) buildRequestLogString(c *gin.Context, targetURL string) string {
	var logBuilder strings.Builder

	// 1. Add the target URL
	logBuilder.WriteString(fmt.Sprintf("Target: %s | ", targetURL))

	// 2. Add all headers
	logBuilder.WriteString("Headers: {")
	headerCount := 0
	for k, v := range c.Request.Header {
		if headerCount > 0 {
			logBuilder.WriteString(", ")
		}
		// Format the header value. The `v` is a slice of strings.
		logBuilder.WriteString(fmt.Sprintf(`"%s": "%s"`, k, strings.Join(v, ", ")))
		headerCount++
	}
	logBuilder.WriteString("} | ")

	// 3. Add the request body from the cache
	cachedBody, exists := c.Get("cachedBody")
	logBuilder.WriteString("Body: ")
	if exists {
		// Convert the byte slice from the cache into a string
		logBuilder.WriteString(string(cachedBody.([]byte)))
	} else {
		logBuilder.WriteString("(empty)")
	}

	return logBuilder.String()
}

// buildResponseLogString constructs a single string containing all relevant response details.
func (h *ProxyHandler) buildResponseLogString(respBody []byte) string {
	var logBuilder strings.Builder

	logBuilder.WriteString("Body: ")
	if len(respBody) > 0 {
		logBuilder.WriteString(string(respBody))
	} else {
		logBuilder.WriteString("(empty)")
	}
	return logBuilder.String()
}

// ProxyHandler forwards the request after logging its contents.
func (h *ProxyHandler) ProxyHandler(c *gin.Context) {
	targetURL := c.Query("target")
	if targetURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'target' query parameter"})
		return
	}

	// --- LOGGING INCOMING REQUEST ---
	clientKey := c.GetHeader("X-PARTNER-ID")
	productType := c.GetHeader("X-EXTERNAL-ID")
	requestLogStr := h.buildRequestLogString(c, targetURL)
	go h.tracelog.Log("PROXY_REQUEST", clientKey, productType, requestLogStr)
	// --- END REQUEST LOGGING ---

	// --- PROXY LOGIC ---
	cachedBody, exists := c.Get("cachedBody")
	var requestBody io.Reader
	if exists {
		requestBody = bytes.NewBuffer(cachedBody.([]byte))
	} else {
		requestBody = c.Request.Body // Fallback for GET requests etc.
	}

	req, err := http.NewRequest(c.Request.Method, targetURL, requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create proxy request"})
		return
	}

	req.Header = c.Request.Header
	req.Header.Del("Host")
	req.Host = ""

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to reach target server", "details": err.Error()})
		return
	}
	defer resp.Body.Close()

	// --- LOGGING OUTGOING RESPONSE ---
	// Use a buffer to capture the response body as it's being streamed back to the client.
	var respBodyBuffer bytes.Buffer
	// TeeReader writes to our buffer as data is read from the original response body.
	teeReader := io.TeeReader(resp.Body, &respBodyBuffer)

	// Copy headers from the proxy response to our main response writer.
	for k, v := range resp.Header {
		c.Writer.Header()[k] = v
	}
	// Write the status code to the client. This must be done before writing the body.
	c.Writer.WriteHeader(resp.StatusCode)

	// Stream the response body to the client. This action simultaneously fills respBodyBuffer.
	io.Copy(c.Writer, teeReader)

	// Now that the response has been fully sent, we can log it asynchronously.
	responseLogStr := h.buildResponseLogString(respBodyBuffer.Bytes())
	go h.tracelog.Log("PROXY_RESPONSE", clientKey, productType, responseLogStr)
	// --- END RESPONSE LOGGING ---
}
