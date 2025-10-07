package middleware

import (
	"strings"

	"github.com/go-historical-data/pkg/config"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORS creates a CORS middleware with configuration
func CORS(cfg config.CORSConfig) fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     strings.Join(cfg.AllowedOrigins, ","),
		AllowMethods:     strings.Join(cfg.AllowedMethods, ","),
		AllowHeaders:     strings.Join(cfg.AllowedHeaders, ","),
		AllowCredentials: true,
		ExposeHeaders:    "X-Request-ID",
	})
}
