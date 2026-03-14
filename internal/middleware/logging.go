package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// responseWriter wraps gin.ResponseWriter to capture response body
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// LoggingMiddleware logs request and response details
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Log request
		requestBody := logRequest(c)

		// Create custom response writer to capture response
		blw := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = blw

		// Process request
		c.Next()

		// Log response
		logResponse(c, blw, requestBody, startTime)
	}
}

// logRequest captures and logs the request body
func logRequest(c *gin.Context) string {
	// Read request body
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
	}

	// Restore the request body so handlers can read it
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Format request body for logging
	var requestBody string
	if len(bodyBytes) > 0 {
		// Try to format as JSON for readability
		var jsonBody interface{}
		if err := json.Unmarshal(bodyBytes, &jsonBody); err == nil {
			prettyJSON, _ := json.MarshalIndent(jsonBody, "", "  ")
			requestBody = string(prettyJSON)
		} else {
			requestBody = string(bodyBytes)
		}
	} else {
		requestBody = "(empty)"
	}

	log.Printf("📥 REQUEST: %s %s", c.Request.Method, c.Request.URL.Path)
	log.Printf("   Headers: %v", c.Request.Header)
	log.Printf("   Body: %s", requestBody)

	return requestBody
}

// logResponse captures and logs the response
func logResponse(c *gin.Context, w *responseWriter, requestBody string, startTime time.Time) {
	duration := time.Since(startTime)

	// Format response body for logging
	var responseBody string
	if w.body.Len() > 0 {
		// Try to format as JSON for readability
		var jsonBody interface{}
		if err := json.Unmarshal(w.body.Bytes(), &jsonBody); err == nil {
			prettyJSON, _ := json.MarshalIndent(jsonBody, "", "  ")
			responseBody = string(prettyJSON)
		} else {
			responseBody = w.body.String()
		}
	} else {
		responseBody = "(empty)"
	}

	log.Printf("📤 RESPONSE: %s %s | Status: %d | Duration: %v",
		c.Request.Method, c.Request.URL.Path, c.Writer.Status(), duration)
	log.Printf("   Body: %s", responseBody)
	log.Println("---")
}
