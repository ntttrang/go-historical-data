package model

import (
	"time"
)

// HistoricalData represents OHLC historical data entity
type HistoricalData struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Symbol    string    `gorm:"type:varchar(20);not null;index:idx_symbol_date" json:"symbol"`
	Date      time.Time `gorm:"type:date;not null;index:idx_symbol_date" json:"date"`
	Open      float64   `gorm:"type:decimal(20,8);not null" json:"open"`
	High      float64   `gorm:"type:decimal(20,8);not null" json:"high"`
	Low       float64   `gorm:"type:decimal(20,8);not null" json:"low"`
	Close     float64   `gorm:"type:decimal(20,8);not null" json:"close"`
	Volume    uint64    `gorm:"type:bigint unsigned;not null;default:0" json:"volume"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for GORM
func (HistoricalData) TableName() string {
	return "historical_data"
}
