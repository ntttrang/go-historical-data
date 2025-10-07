package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestIDHeader is the header key for request ID
const RequestIDHeader = "X-Request-ID"

// RequestID middleware adds a unique request ID to each request
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if request ID already exists in header
		requestID := c.Get(RequestIDHeader)
		if requestID == "" {
			// Generate new UUID
			requestID = uuid.New().String()
		}

		// Set request ID in context and response header
		c.Locals("request_id", requestID)
		c.Set(RequestIDHeader, requestID)

		return c.Next()
	}
}

// GetRequestID retrieves the request ID from context
func GetRequestID(c *fiber.Ctx) string {
	if requestID, ok := c.Locals("request_id").(string); ok {
		return requestID
	}
	return ""
}
