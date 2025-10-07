package controller

import (
	"github.com/go-historical-data/pkg/response"
	"github.com/gofiber/fiber/v2"
)

// HealthController handles health check endpoints
type HealthController struct{}

// NewHealthController creates a new health controller instance
func NewHealthController() *HealthController {
	return &HealthController{}
}

// HealthCheckResponse represents the health check response
type HealthCheckResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

// Check handles GET /health endpoint
func (h *HealthController) Check(c *fiber.Ctx) error {
	return response.Success(c, HealthCheckResponse{
		Status:  "healthy",
		Service: "historical-data-api",
		Version: "1.0.0",
	})
}
