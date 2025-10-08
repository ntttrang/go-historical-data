package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/go-historical-data/internal/middleware"
	"github.com/go-historical-data/internal/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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
	start := time.Now()
	err := r.db.WithContext(ctx).Create(data).Error
	middleware.RecordDBMetrics("insert", time.Since(start), err)

	if err != nil {
		return fmt.Errorf("failed to create historical data: %w", err)
	}
	return nil
}

// BulkCreate creates multiple historical data records in a single transaction
func (r *historicalRepository) BulkCreate(ctx context.Context, data []model.HistoricalData, batchSize int) error {
	tracer := otel.Tracer("historical-repository")
	ctx, span := tracer.Start(ctx, "HistoricalRepository.BulkCreate")
	defer span.End()

	span.SetAttributes(
		attribute.Int("record_count", len(data)),
		attribute.Int("batch_size", batchSize),
	)

	if len(data) == 0 {
		return nil
	}

	// Track database operation time
	start := time.Now()

	// Use batch insert with conflict handling (upsert)
	// If duplicate symbol+date exists, update the record
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "symbol"}, {Name: "date"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"open", "high", "low", "close", "volume", "updated_at",
		}),
	}).CreateInBatches(data, batchSize).Error

	// Record metrics
	middleware.RecordDBMetrics("bulk_insert", time.Since(start), err)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "bulk insert failed")
		return fmt.Errorf("failed to bulk create historical data: %w", err)
	}

	span.SetStatus(codes.Ok, "bulk insert successful")
	return nil
}

// FindBySymbol retrieves historical data for a specific symbol within a date range
func (r *historicalRepository) FindBySymbol(ctx context.Context, symbol string, startDate, endDate time.Time) ([]model.HistoricalData, error) {
	start := time.Now()
	var data []model.HistoricalData
	query := r.db.WithContext(ctx).Where("symbol = ?", symbol)

	if !startDate.IsZero() {
		query = query.Where("date >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("date <= ?", endDate)
	}

	err := query.Order("date ASC").Find(&data).Error
	middleware.RecordDBMetrics("select", time.Since(start), err)

	if err != nil {
		return nil, fmt.Errorf("failed to find historical data by symbol: %w", err)
	}

	return data, nil
}

// FindAll retrieves all historical data with optional filters and pagination
func (r *historicalRepository) FindAll(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]model.HistoricalData, int64, error) {
	tracer := otel.Tracer("historical-repository")
	ctx, span := tracer.Start(ctx, "HistoricalRepository.FindAll")
	defer span.End()

	span.SetAttributes(
		attribute.Int("limit", limit),
		attribute.Int("offset", offset),
	)

	var data []model.HistoricalData
	var total int64

	query := r.db.WithContext(ctx).Model(&model.HistoricalData{})

	// Apply filters
	query = r.applyFilters(query, filters)

	// Count total records
	start := time.Now()
	countErr := query.Count(&total).Error
	middleware.RecordDBMetrics("select", time.Since(start), countErr)

	if countErr != nil {
		span.RecordError(countErr)
		span.SetStatus(codes.Error, "count query failed")
		return nil, 0, fmt.Errorf("failed to count historical data: %w", countErr)
	}

	// Apply pagination and fetch data
	start = time.Now()
	findErr := query.Limit(limit).Offset(offset).Order("date DESC").Find(&data).Error
	middleware.RecordDBMetrics("select", time.Since(start), findErr)

	if findErr != nil {
		span.RecordError(findErr)
		span.SetStatus(codes.Error, "select query failed")
		return nil, 0, fmt.Errorf("failed to find all historical data: %w", findErr)
	}

	span.SetAttributes(
		attribute.Int64("total_count", total),
		attribute.Int("returned_count", len(data)),
	)

	return data, total, nil
}

// FindByID retrieves a single historical data record by ID
func (r *historicalRepository) FindByID(ctx context.Context, id uint64) (*model.HistoricalData, error) {
	tracer := otel.Tracer("historical-repository")
	ctx, span := tracer.Start(ctx, "HistoricalRepository.FindByID")
	defer span.End()

	span.SetAttributes(attribute.String("id", fmt.Sprintf("%d", id)))

	start := time.Now()
	var data model.HistoricalData
	err := r.db.WithContext(ctx).First(&data, id).Error
	middleware.RecordDBMetrics("select", time.Since(start), err)

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			span.SetAttributes(attribute.Bool("found", false))
			return nil, nil
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "query failed")
		return nil, fmt.Errorf("failed to find historical data by id: %w", err)
	}

	span.SetAttributes(attribute.Bool("found", true))
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
