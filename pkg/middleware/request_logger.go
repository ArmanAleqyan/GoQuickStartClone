package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"log"
)

// RequestLogger logs all HTTP requests
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		startTime := time.Now()

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(startTime)

		// Get status code
		statusCode := c.Writer.Status()

		// Get request info
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()

		// Log request
		log.Printf("[%s] %s %s | Status: %d | Latency: %v | IP: %s",
			method,
			path,
			c.Request.Proto,
			statusCode,
			latency,
			clientIP,
		)

		// Log errors if any
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				log.Printf("[ERROR] %s", err.Error())
			}
		}
	}
}
