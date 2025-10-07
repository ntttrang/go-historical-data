package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/timeout"
)

// Timeout creates a timeout middleware
func Timeout(duration int) fiber.Handler {
	return timeout.NewWithContext(func(c *fiber.Ctx) error {
		return c.Next()
	}, time.Duration(duration)*time.Second)
}
