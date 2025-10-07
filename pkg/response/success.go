package response

import (
	"github.com/gofiber/fiber/v2"
)

// SuccessResponse represents a standardized success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Success sends a success response with data
func Success(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(SuccessResponse{
		Success: true,
		Data:    data,
	})
}

// SuccessWithMessage sends a success response with message and data
func SuccessWithMessage(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created sends a 201 Created response with data
func Created(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Message: "Resource created successfully",
		Data:    data,
	})
}

// NoContent sends a 204 No Content response
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}
