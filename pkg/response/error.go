package response

import (
	"github.com/gofiber/fiber/v2"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   ErrorDetail `json:"error"`
}

// ErrorDetail contains error details
type ErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Error codes
const (
	ErrCodeBadRequest         = "BAD_REQUEST"
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeForbidden          = "FORBIDDEN"
	ErrCodeNotFound           = "NOT_FOUND"
	ErrCodeConflict           = "CONFLICT"
	ErrCodeValidation         = "VALIDATION_ERROR"
	ErrCodeInternalServer     = "INTERNAL_SERVER_ERROR"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrCodeTooManyRequests    = "TOO_MANY_REQUESTS"
	ErrCodeDatabaseError      = "DATABASE_ERROR"
	ErrCodeCacheError         = "CACHE_ERROR"
)

// BadRequest sends a 400 Bad Request error response
func BadRequest(c *fiber.Ctx, message string, details interface{}) error {
	return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    ErrCodeBadRequest,
			Message: message,
			Details: details,
		},
	})
}

// Unauthorized sends a 401 Unauthorized error response
func Unauthorized(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    ErrCodeUnauthorized,
			Message: message,
		},
	})
}

// Forbidden sends a 403 Forbidden error response
func Forbidden(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    ErrCodeForbidden,
			Message: message,
		},
	})
}

// NotFound sends a 404 Not Found error response
func NotFound(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    ErrCodeNotFound,
			Message: message,
		},
	})
}

// Conflict sends a 409 Conflict error response
func Conflict(c *fiber.Ctx, message string, details interface{}) error {
	return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    ErrCodeConflict,
			Message: message,
			Details: details,
		},
	})
}

// ValidationError sends a 422 Unprocessable Entity error response
func ValidationError(c *fiber.Ctx, message string, details interface{}) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    ErrCodeValidation,
			Message: message,
			Details: details,
		},
	})
}

// InternalServerError sends a 500 Internal Server Error response
func InternalServerError(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    ErrCodeInternalServer,
			Message: message,
		},
	})
}

// ServiceUnavailable sends a 503 Service Unavailable error response
func ServiceUnavailable(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusServiceUnavailable).JSON(ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    ErrCodeServiceUnavailable,
			Message: message,
		},
	})
}

// TooManyRequests sends a 429 Too Many Requests error response
func TooManyRequests(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusTooManyRequests).JSON(ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    ErrCodeTooManyRequests,
			Message: message,
		},
	})
}
