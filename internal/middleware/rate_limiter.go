package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimiter creates a rate limiter middleware
func RateLimiter(maxRequests int) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        maxRequests,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return fiber.NewError(fiber.StatusTooManyRequests, "Rate limit exceeded")
		},
		Storage: nil, // Use in-memory storage (for production, use Redis)
	})
}
