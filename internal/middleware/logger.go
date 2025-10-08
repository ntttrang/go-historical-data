package middleware

import (
	"time"

	"github.com/go-historical-data/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

// Logger middleware logs HTTP requests and responses
func Logger(log *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Start timer
		start := time.Now()

		// Get request ID
		requestID := GetRequestID(c)

		// Create logger with request context
		reqLogger := log.WithRequestID(requestID)

		// Add trace context if available
		if traceID, ok := c.Locals("trace_id").(string); ok {
			if spanID, ok := c.Locals("span_id").(string); ok {
				reqLogger = reqLogger.WithTrace(traceID, spanID)
			}
		}

		// Store logger in context
		c.Locals("logger", reqLogger)

		// Log incoming request
		reqLogger.Info().
			Str("method", c.Method()).
			Str("path", c.Path()).
			Str("ip", c.IP()).
			Str("user_agent", c.Get("User-Agent")).
			Msg("Incoming request")

		// Process request
		err := c.Next()

		// Calculate request duration
		duration := time.Since(start)

		// Log response
		logEvent := reqLogger.Info()
		if err != nil {
			logEvent = reqLogger.Error().Err(err)
		}

		logEvent.
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", c.Response().StatusCode()).
			Dur("duration_ms", duration).
			Int("size", len(c.Response().Body())).
			Msg("Request completed")

		return err
	}
}

// GetLogger retrieves the logger from context
func GetLogger(c *fiber.Ctx) *logger.Logger {
	if log, ok := c.Locals("logger").(*logger.Logger); ok {
		return log
	}
	return logger.GetGlobalLogger()
}
