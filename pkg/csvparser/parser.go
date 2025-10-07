package csvparser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// HistoricalDataRow represents a single row from CSV
type HistoricalDataRow struct {
	Symbol string
	Date   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume uint64
}

// ParseError represents a parsing error with line number
type ParseError struct {
	Line    int
	Field   string
	Value   string
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("line %d, field '%s', value '%s': %s", e.Line, e.Field, e.Value, e.Message)
}

// Parser handles CSV parsing for historical data
type Parser struct {
	reader           *csv.Reader
	headers          []string
	headerIndexes    map[string]int
	currentLine      int
	supportedFormats []string
}

// NewParser creates a new CSV parser
func NewParser(r io.Reader) *Parser {
	csvReader := csv.NewReader(r)
	csvReader.TrimLeadingSpace = true
	csvReader.ReuseRecord = true // Memory optimization

	return &Parser{
		reader:      csvReader,
		currentLine: 0,
		supportedFormats: []string{
			"2006-01-02",
			"01/02/2006",
			"02-01-2006",
			"2006/01/02",
			"01-02-2006",
		},
	}
}

// ParseHeader reads and validates the CSV header
func (p *Parser) ParseHeader() error {
	header, err := p.reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	p.currentLine++
	p.headers = make([]string, len(header))
	p.headerIndexes = make(map[string]int)

	// Normalize header names (lowercase, trim spaces)
	for i, h := range header {
		normalized := strings.ToLower(strings.TrimSpace(h))
		p.headers[i] = normalized
		p.headerIndexes[normalized] = i
	}

	// Validate required headers
	requiredHeaders := []string{"symbol", "date", "open", "high", "low", "close", "volume"}
	for _, required := range requiredHeaders {
		if _, exists := p.headerIndexes[required]; !exists {
			return fmt.Errorf("missing required header: %s", required)
		}
	}

	return nil
}

// ParseRow reads and parses a single row
func (p *Parser) ParseRow() (*HistoricalDataRow, error) {
	record, err := p.reader.Read()
	if err != nil {
		return nil, err
	}

	p.currentLine++

	// Parse each field
	row := &HistoricalDataRow{}

	// Symbol
	symbolIdx := p.headerIndexes["symbol"]
	row.Symbol = strings.TrimSpace(strings.ToUpper(record[symbolIdx]))
	if row.Symbol == "" {
		return nil, &ParseError{
			Line:    p.currentLine,
			Field:   "symbol",
			Value:   record[symbolIdx],
			Message: "symbol cannot be empty",
		}
	}

	// Date
	dateIdx := p.headerIndexes["date"]
	dateStr := strings.TrimSpace(record[dateIdx])
	row.Date, err = p.parseDate(dateStr)
	if err != nil {
		return nil, &ParseError{
			Line:    p.currentLine,
			Field:   "date",
			Value:   dateStr,
			Message: fmt.Sprintf("invalid date format, supported formats: %s", strings.Join(p.supportedFormats, ", ")),
		}
	}

	// Open
	openIdx := p.headerIndexes["open"]
	row.Open, err = p.parseFloat(record[openIdx])
	if err != nil {
		return nil, &ParseError{
			Line:    p.currentLine,
			Field:   "open",
			Value:   record[openIdx],
			Message: "must be a valid number",
		}
	}

	// High
	highIdx := p.headerIndexes["high"]
	row.High, err = p.parseFloat(record[highIdx])
	if err != nil {
		return nil, &ParseError{
			Line:    p.currentLine,
			Field:   "high",
			Value:   record[highIdx],
			Message: "must be a valid number",
		}
	}

	// Low
	lowIdx := p.headerIndexes["low"]
	row.Low, err = p.parseFloat(record[lowIdx])
	if err != nil {
		return nil, &ParseError{
			Line:    p.currentLine,
			Field:   "low",
			Value:   record[lowIdx],
			Message: "must be a valid number",
		}
	}

	// Close
	closeIdx := p.headerIndexes["close"]
	row.Close, err = p.parseFloat(record[closeIdx])
	if err != nil {
		return nil, &ParseError{
			Line:    p.currentLine,
			Field:   "close",
			Value:   record[closeIdx],
			Message: "must be a valid number",
		}
	}

	// Volume
	volumeIdx := p.headerIndexes["volume"]
	row.Volume, err = p.parseUint(record[volumeIdx])
	if err != nil {
		return nil, &ParseError{
			Line:    p.currentLine,
			Field:   "volume",
			Value:   record[volumeIdx],
			Message: "must be a valid non-negative integer",
		}
	}

	return row, nil
}

// ParseAll reads all rows from the CSV
func (p *Parser) ParseAll() ([]HistoricalDataRow, []error) {
	rows := make([]HistoricalDataRow, 0)
	errors := make([]error, 0)

	for {
		row, err := p.ParseRow()
		if err != nil {
			if err == io.EOF {
				break
			}
			errors = append(errors, err)
			continue
		}

		rows = append(rows, *row)
	}

	return rows, errors
}

// GetCurrentLine returns the current line number being processed
func (p *Parser) GetCurrentLine() int {
	return p.currentLine
}

// parseDate tries multiple date formats
func (p *Parser) parseDate(dateStr string) (time.Time, error) {
	for _, format := range p.supportedFormats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse date")
}

// parseFloat parses a float64 value
func (p *Parser) parseFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty value")
	}

	// Remove common currency symbols and commas
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "$", "")

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}

	if val < 0 {
		return 0, fmt.Errorf("negative value not allowed")
	}

	return val, nil
}

// parseUint parses a uint64 value
func (p *Parser) parseUint(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil // Volume can be 0
	}

	// Remove commas
	s = strings.ReplaceAll(s, ",", "")

	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}

	return val, nil
}
