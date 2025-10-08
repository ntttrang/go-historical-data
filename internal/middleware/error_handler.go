package middleware

import (
	"errors"

	"github.com/go-historical-data/pkg/response"
	"github.com/gofiber/fiber/v2"
)

// ErrorHandler is a global error handler middleware
func ErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		// Get logger from context
		log := GetLogger(c)

		// Default to 500 Internal Server Error
		code := fiber.StatusInternalServerError
		message := "Internal Server Error"

		// Check if it's a Fiber error
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			code = fiberErr.Code
			message = fiberErr.Message
		}

		// Log error
		log.Error().
			Err(err).
			Int("status", code).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Msg("Request error")

		// Send error response based on status code
		switch code {
		case fiber.StatusBadRequest:
			return response.BadRequest(c, message, nil)
		case fiber.StatusUnauthorized:
			return response.Unauthorized(c, message)
		case fiber.StatusForbidden:
			return response.Forbidden(c, message)
		case fiber.StatusNotFound:
			return response.NotFound(c, message)
		case fiber.StatusTooManyRequests:
			return response.TooManyRequests(c, message)
		default:
			return response.InternalServerError(c, message)
		}
	}
}

// Recover middleware recovers from panics
func Recover() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				log := GetLogger(c)
				log.Error().
					Interface("panic", r).
					Str("method", c.Method()).
					Str("path", c.Path()).
					Msg("Panic recovered")

				// Send error response, ignore any error from the response itself
				// as we're already in a panic recovery situation
				if err := response.InternalServerError(c, "Internal Server Error"); err != nil {
					log.Error().Err(err).Msg("Failed to send panic response")
				}
			}
		}()

		return c.Next()
	}
}
