package response

import (
	"time"
)

// HistoricalDataResponse represents a single historical data record in the response
type HistoricalDataResponse struct {
	ID        uint64    `json:"id"`
	Symbol    string    `json:"symbol"`
	Date      string    `json:"date"` // Format: YYYY-MM-DD
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    uint64    `json:"volume"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PaginatedHistoricalDataResponse represents paginated historical data
type PaginatedHistoricalDataResponse struct {
	Data       []HistoricalDataResponse `json:"data"`
	Pagination PaginationMeta           `json:"pagination"`
}

// PaginationMeta contains pagination metadata
type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

// CSVUploadResponse represents the response for CSV file upload
type CSVUploadResponse struct {
	TotalRows      int      `json:"total_rows"`
	SuccessCount   int      `json:"success_count"`
	FailedCount    int      `json:"failed_count"`
	ProcessedBytes int64    `json:"processed_bytes"`
	Errors         []string `json:"errors,omitempty"`
	Message        string   `json:"message"`
}
