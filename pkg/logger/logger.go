package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger wraps zerolog.Logger
type Logger struct {
	*zerolog.Logger
}

// Config holds logger configuration
type Config struct {
	Level  string
	Format string
}

// New creates a new logger instance
func New(cfg Config) *Logger {
	// Parse log level
	level := parseLogLevel(cfg.Level)
	zerolog.SetGlobalLevel(level)

	// Configure output format
	var logger zerolog.Logger
	if cfg.Format == "console" || cfg.Format == "" {
		logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Caller().Logger()
	} else {
		logger = zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
	}

	return &Logger{Logger: &logger}
}

// parseLogLevel converts string level to zerolog.Level
func parseLogLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

// WithContext returns a new logger with context fields
func (l *Logger) WithContext(fields map[string]interface{}) *Logger {
	ctx := l.Logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	logger := ctx.Logger()
	return &Logger{Logger: &logger}
}

// WithRequestID adds request ID to logger context
func (l *Logger) WithRequestID(requestID string) *Logger {
	logger := l.Logger.With().Str("request_id", requestID).Logger()
	return &Logger{Logger: &logger}
}

// WithTrace adds trace ID and span ID to logger context
func (l *Logger) WithTrace(traceID, spanID string) *Logger {
	logger := l.Logger.With().
		Str("trace_id", traceID).
		Str("span_id", spanID).
		Logger()
	return &Logger{Logger: &logger}
}

// GetGlobalLogger returns the global logger
func GetGlobalLogger() *Logger {
	return &Logger{Logger: &log.Logger}
}
