package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/go-historical-data/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// HistoricalRepository defines the interface for historical data repository
type HistoricalRepository interface {
	Create(ctx context.Context, data *model.HistoricalData) error
	BulkCreate(ctx context.Context, data []model.HistoricalData, batchSize int) error
	FindBySymbol(ctx context.Context, symbol string, startDate, endDate time.Time) ([]model.HistoricalData, error)
	FindAll(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]model.HistoricalData, int64, error)
	FindByID(ctx context.Context, id uint64) (*model.HistoricalData, error)
	Update(ctx context.Context, data *model.HistoricalData) error
	Delete(ctx context.Context, id uint64) error
	Count(ctx context.Context, filters map[string]interface{}) (int64, error)
}

// historicalRepository implements HistoricalRepository interface
type historicalRepository struct {
	db *gorm.DB
}

// NewHistoricalRepository creates a new historical repository instance
func NewHistoricalRepository(db *gorm.DB) HistoricalRepository {
	return &historicalRepository{
		db: db,
	}
}

// Create creates a new historical data record
func (r *historicalRepository) Create(ctx context.Context, data *model.HistoricalData) error {
	if err := r.db.WithContext(ctx).Create(data).Error; err != nil {
		return fmt.Errorf("failed to create historical data: %w", err)
	}
	return nil
}

// BulkCreate creates multiple historical data records in a single transaction
func (r *historicalRepository) BulkCreate(ctx context.Context, data []model.HistoricalData, batchSize int) error {
	if len(data) == 0 {
		return nil
	}

	// Use batch insert with conflict handling (upsert)
	// If duplicate symbol+date exists, update the record
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "symbol"}, {Name: "date"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"open", "high", "low", "close", "volume", "updated_at",
		}),
	}).CreateInBatches(data, batchSize).Error

	if err != nil {
		return fmt.Errorf("failed to bulk create historical data: %w", err)
	}

	return nil
}

// FindBySymbol retrieves historical data for a specific symbol within a date range
func (r *historicalRepository) FindBySymbol(ctx context.Context, symbol string, startDate, endDate time.Time) ([]model.HistoricalData, error) {
	var data []model.HistoricalData
	query := r.db.WithContext(ctx).Where("symbol = ?", symbol)

	if !startDate.IsZero() {
		query = query.Where("date >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("date <= ?", endDate)
	}

	if err := query.Order("date ASC").Find(&data).Error; err != nil {
		return nil, fmt.Errorf("failed to find historical data by symbol: %w", err)
	}

	return data, nil
}

// FindAll retrieves all historical data with optional filters and pagination
func (r *historicalRepository) FindAll(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]model.HistoricalData, int64, error) {
	var data []model.HistoricalData
	var total int64

	query := r.db.WithContext(ctx).Model(&model.HistoricalData{})

	// Apply filters
	query = r.applyFilters(query, filters)

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count historical data: %w", err)
	}

	// Apply pagination and fetch data
	if err := query.Limit(limit).Offset(offset).Order("date DESC").Find(&data).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find all historical data: %w", err)
	}

	return data, total, nil
}

// FindByID retrieves a single historical data record by ID
func (r *historicalRepository) FindByID(ctx context.Context, id uint64) (*model.HistoricalData, error) {
	var data model.HistoricalData
	if err := r.db.WithContext(ctx).First(&data, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find historical data by id: %w", err)
	}
	return &data, nil
}

// Update updates an existing historical data record
func (r *historicalRepository) Update(ctx context.Context, data *model.HistoricalData) error {
	if err := r.db.WithContext(ctx).Save(data).Error; err != nil {
		return fmt.Errorf("failed to update historical data: %w", err)
	}
	return nil
}

// Delete deletes a historical data record by ID
func (r *historicalRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.HistoricalData{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete historical data: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Count returns the total count of records matching the filters
func (r *historicalRepository) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.HistoricalData{})
	query = r.applyFilters(query, filters)

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count historical data: %w", err)
	}

	return count, nil
}

// applyFilters applies filters to the query
func (r *historicalRepository) applyFilters(query *gorm.DB, filters map[string]interface{}) *gorm.DB {
	if symbol, ok := filters["symbol"].(string); ok && symbol != "" {
		query = query.Where("symbol = ?", symbol)
	}
	if startDate, ok := filters["start_date"].(time.Time); ok && !startDate.IsZero() {
		query = query.Where("date >= ?", startDate)
	}
	if endDate, ok := filters["end_date"].(time.Time); ok && !endDate.IsZero() {
		query = query.Where("date <= ?", endDate)
	}
	return query
}
