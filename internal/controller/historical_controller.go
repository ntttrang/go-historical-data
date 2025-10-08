package controller

import (
	"strconv"
	"time"

	"github.com/go-historical-data/internal/dto/request"
	"github.com/go-historical-data/internal/middleware"
	"github.com/go-historical-data/internal/service"
	"github.com/go-historical-data/pkg/response"
	"github.com/go-historical-data/pkg/validator"
	"github.com/gofiber/fiber/v2"
)

// HistoricalController handles historical data endpoints
type HistoricalController struct {
	service   service.HistoricalService
	validator *validator.Validator
}

// NewHistoricalController creates a new historical controller instance
func NewHistoricalController(service service.HistoricalService, validator *validator.Validator) *HistoricalController {
	return &HistoricalController{
		service:   service,
		validator: validator,
	}
}

// GetData handles GET /api/v1/data - Retrieve historical data
func (h *HistoricalController) GetData(c *fiber.Ctx) error {
	var req request.GetDataRequest

	// Parse query parameters
	if err := c.QueryParser(&req); err != nil {
		return response.BadRequest(c, "Invalid query parameters", err.Error())
	}

	// Validate request
	if err := h.validator.Validate(&req); err != nil {
		if validationErr, ok := err.(*validator.ValidationError); ok {
			return response.ValidationError(c, "Validation failed", validationErr.GetErrors())
		}
		return response.BadRequest(c, "Validation failed", err.Error())
	}

	// Validate date range
	if err := req.Validate(); err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}

	// Call service
	result, err := h.service.GetHistoricalData(c.UserContext(), &req)
	if err != nil {
		return response.InternalServerError(c, err.Error())
	}

	return response.Success(c, result)
}

// GetDataByID handles GET /api/v1/data/:id - Retrieve historical data by ID
func (h *HistoricalController) GetDataByID(c *fiber.Ctx) error {
	// Parse ID parameter
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid ID parameter", err.Error())
	}

	// Call service
	result, err := h.service.GetHistoricalDataByID(c.UserContext(), id)
	if err != nil {
		return response.InternalServerError(c, err.Error())
	}

	if result == nil {
		return response.NotFound(c, "Historical data not found")
	}

	return response.Success(c, result)
}

// UploadCSV handles POST /api/v1/data - Upload CSV file
func (h *HistoricalController) UploadCSV(c *fiber.Ctx) error {
	// Parse multipart form
	file, err := c.FormFile("file")
	if err != nil {
		return response.BadRequest(c, "No file uploaded", err.Error())
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	if contentType != "text/csv" && contentType != "application/vnd.ms-excel" && contentType != "application/csv" {
		// Also check file extension as a fallback
		if len(file.Filename) < 4 || file.Filename[len(file.Filename)-4:] != ".csv" {
			return response.BadRequest(c, "Invalid file type", "Only CSV files are allowed")
		}
	}

	// Validate file size (max 50MB)
	// const maxFileSize = 50 * 1024 * 1024 // 50MB
	// if file.Size > maxFileSize {
	// 	return response.BadRequest(c, "File too large", "Maximum file size is 50MB")
	// }

	// Open file
	fileReader, err := file.Open()
	if err != nil {
		return response.InternalServerError(c, "Failed to read file")
	}
	defer fileReader.Close() //nolint:errcheck // Close errors in defer are commonly ignored in HTTP handlers

	// Track CSV upload duration
	startTime := time.Now()

	// Process CSV file
	result, err := h.service.UploadCSV(c.UserContext(), fileReader, file.Size)

	// Record metrics
	duration := time.Since(startTime)
	if err != nil {
		middleware.RecordCSVMetrics(0, 0, duration, "error")
		return response.InternalServerError(c, err.Error())
	}

	// Determine upload status based on errors
	uploadStatus := "success"
	if len(result.Errors) > 0 {
		if result.SuccessCount == 0 {
			uploadStatus = "error"
		} else {
			uploadStatus = "partial"
		}
	}

	middleware.RecordCSVMetrics(result.SuccessCount, result.FailedCount, duration, uploadStatus)

	return response.Success(c, result)
}
