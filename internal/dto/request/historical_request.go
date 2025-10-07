package request

import (
	"time"
)

// GetDataRequest represents query parameters for retrieving historical data
type GetDataRequest struct {
	Symbol    string    `query:"symbol" validate:"omitempty,min=1,max=20"`
	StartDate time.Time `query:"start_date" validate:"omitempty"`
	EndDate   time.Time `query:"end_date" validate:"omitempty"`
	Page      int       `query:"page" validate:"omitempty,min=1"`
	Limit     int       `query:"limit" validate:"omitempty,min=1,max=1000"`
}

// SetDefaults sets default values for pagination
func (r *GetDataRequest) SetDefaults() {
	if r.Page == 0 {
		r.Page = 1
	}
	if r.Limit == 0 {
		r.Limit = 100
	}
}

// GetOffset calculates the offset for pagination
func (r *GetDataRequest) GetOffset() int {
	return (r.Page - 1) * r.Limit
}

// Validate validates the date range
func (r *GetDataRequest) Validate() error {
	if !r.StartDate.IsZero() && !r.EndDate.IsZero() && r.StartDate.After(r.EndDate) {
		return ErrInvalidDateRange
	}
	return nil
}

// ErrInvalidDateRange is returned when start_date is after end_date
var ErrInvalidDateRange = &ValidationError{
	Field:   "date_range",
	Message: "start_date must be before or equal to end_date",
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
