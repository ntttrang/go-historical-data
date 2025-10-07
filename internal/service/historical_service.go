package service

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/go-historical-data/internal/dto/request"
	"github.com/go-historical-data/internal/dto/response"
	"github.com/go-historical-data/internal/model"
	"github.com/go-historical-data/internal/repository"
	"github.com/go-historical-data/pkg/csvparser"
)

// HistoricalService defines the interface for historical data business logic
type HistoricalService interface {
	UploadCSV(ctx context.Context, reader io.Reader, fileSize int64) (*response.CSVUploadResponse, error)
	GetHistoricalData(ctx context.Context, req *request.GetDataRequest) (*response.PaginatedHistoricalDataResponse, error)
	GetHistoricalDataByID(ctx context.Context, id uint64) (*response.HistoricalDataResponse, error)
}

// historicalService implements HistoricalService interface
type historicalService struct {
	repo repository.HistoricalRepository
}

// NewHistoricalService creates a new historical service instance
func NewHistoricalService(repo repository.HistoricalRepository) HistoricalService {
	return &historicalService{
		repo: repo,
	}
}

// GetHistoricalData retrieves historical data
func (s *historicalService) GetHistoricalData(ctx context.Context, req *request.GetDataRequest) (*response.PaginatedHistoricalDataResponse, error) {
	// Set defaults
	req.SetDefaults()

	// Validate date range
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Build filters
	filters := make(map[string]interface{})
	if req.Symbol != "" {
		filters["symbol"] = req.Symbol
	}
	if !req.StartDate.IsZero() {
		filters["start_date"] = req.StartDate
	}
	if !req.EndDate.IsZero() {
		filters["end_date"] = req.EndDate
	}

	// Fetch from database
	data, total, err := s.repo.FindAll(ctx, filters, req.Limit, req.GetOffset())
	if err != nil {
		return nil, fmt.Errorf("failed to get historical data: %w", err)
	}

	// Convert to response
	responseData := make([]response.HistoricalDataResponse, len(data))
	for i, item := range data {
		responseData[i] = s.toHistoricalDataResponse(&item)
	}

	// Calculate pagination metadata
	totalPages := int(total) / req.Limit
	if int(total)%req.Limit > 0 {
		totalPages++
	}

	result := &response.PaginatedHistoricalDataResponse{
		Data: responseData,
		Pagination: response.PaginationMeta{
			Page:       req.Page,
			Limit:      req.Limit,
			TotalItems: total,
			TotalPages: totalPages,
		},
	}

	return result, nil
}

// GetHistoricalDataByID retrieves a single historical data record by ID
func (s *historicalService) GetHistoricalDataByID(ctx context.Context, id uint64) (*response.HistoricalDataResponse, error) {
	// Fetch from database
	data, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical data by id: %w", err)
	}
	if data == nil {
		return nil, nil
	}

	// Convert to response
	result := s.toHistoricalDataResponse(data)

	return &result, nil
}

// UploadCSV processes and stores CSV file data with batch processing
func (s *historicalService) UploadCSV(ctx context.Context, reader io.Reader, fileSize int64) (*response.CSVUploadResponse, error) {
	const batchSize = 1000

	parser := csvparser.NewParser(reader)

	// Parse and validate header
	if err := parser.ParseHeader(); err != nil {
		return nil, fmt.Errorf("invalid CSV header: %w", err)
	}

	var totalRows int
	var successCount int
	var failedCount int
	var errors []string
	batch := make([]model.HistoricalData, 0, batchSize)

	// Process rows in batches
	for {
		row, err := parser.ParseRow()
		if err != nil {
			if err == io.EOF {
				break
			}
			// Collect error but continue processing
			errors = append(errors, err.Error())
			failedCount++
			continue
		}

		totalRows++

		// Validate business rules
		if err := s.validateCSVRow(row); err != nil {
			errors = append(errors, fmt.Sprintf("line %d: %v", parser.GetCurrentLine(), err))
			failedCount++
			continue
		}

		// Add to batch
		batch = append(batch, model.HistoricalData{
			Symbol: row.Symbol,
			Date:   row.Date,
			Open:   row.Open,
			High:   row.High,
			Low:    row.Low,
			Close:  row.Close,
			Volume: row.Volume,
		})

		// Process batch when it reaches the size limit
		if len(batch) >= batchSize {
			if err := s.repo.BulkCreate(ctx, batch, batchSize); err != nil {
				// Log error but continue with next batch
				errors = append(errors, fmt.Sprintf("batch insert error: %v", err))
				failedCount += len(batch)
			} else {
				successCount += len(batch)
			}
			batch = batch[:0] // Clear batch
		}
	}

	// Process remaining batch
	if len(batch) > 0 {
		if err := s.repo.BulkCreate(ctx, batch, batchSize); err != nil {
			errors = append(errors, fmt.Sprintf("final batch insert error: %v", err))
			failedCount += len(batch)
		} else {
			successCount += len(batch)
		}
	}

	// Limit errors to first 100 to avoid huge responses
	if len(errors) > 100 {
		errors = append(errors[:100], fmt.Sprintf("... and %d more errors", len(errors)-100))
	}

	message := "CSV file processed successfully"
	if failedCount > 0 {
		message = fmt.Sprintf("CSV file processed with %d errors", failedCount)
	}

	return &response.CSVUploadResponse{
		TotalRows:      totalRows,
		SuccessCount:   successCount,
		FailedCount:    failedCount,
		ProcessedBytes: fileSize,
		Errors:         errors,
		Message:        message,
	}, nil
}

// validateCSVRow validates business rules for CSV row data
func (s *historicalService) validateCSVRow(row *csvparser.HistoricalDataRow) error {
	// Validate OHLC relationships
	if row.High < row.Low {
		return fmt.Errorf("high price (%.2f) must be greater than or equal to low price (%.2f)", row.High, row.Low)
	}
	if row.Open < row.Low || row.Open > row.High {
		return fmt.Errorf("open price (%.2f) must be between low (%.2f) and high (%.2f)", row.Open, row.Low, row.High)
	}
	if row.Close < row.Low || row.Close > row.High {
		return fmt.Errorf("close price (%.2f) must be between low (%.2f) and high (%.2f)", row.Close, row.Low, row.High)
	}
	// Validate date is not in the future
	if row.Date.After(time.Now()) {
		return fmt.Errorf("date (%s) cannot be in the future", row.Date.Format("2006-01-02"))
	}
	// Validate all prices are positive
	if row.Open <= 0 || row.High <= 0 || row.Low <= 0 || row.Close <= 0 {
		return fmt.Errorf("all prices must be positive")
	}
	return nil
}

// toHistoricalDataResponse converts model to response DTO
func (s *historicalService) toHistoricalDataResponse(data *model.HistoricalData) response.HistoricalDataResponse {
	return response.HistoricalDataResponse{
		ID:        data.ID,
		Symbol:    data.Symbol,
		Date:      data.Date.Format("2006-01-02"),
		Open:      data.Open,
		High:      data.High,
		Low:       data.Low,
		Close:     data.Close,
		Volume:    data.Volume,
		CreatedAt: data.CreatedAt,
		UpdatedAt: data.UpdatedAt,
	}
}
