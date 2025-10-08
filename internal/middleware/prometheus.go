package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP request duration histogram
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets, // 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10
		},
		[]string{"method", "path", "status"},
	)

	// HTTP request counter
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// Active HTTP connections gauge
	httpActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_active_connections",
			Help: "Number of active HTTP connections",
		},
	)

	// HTTP request size histogram
	httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000, 10000000}, // 100B to 10MB
		},
		[]string{"method", "path"},
	)

	// HTTP response size histogram
	httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000, 10000000}, // 100B to 10MB
		},
		[]string{"method", "path"},
	)

	// CSV upload metrics
	csvRowsProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "csv_rows_processed_total",
			Help: "Total number of CSV rows processed",
		},
		[]string{"status"}, // success or error
	)

	csvUploadDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "csv_upload_duration_seconds",
			Help:    "CSV upload processing duration in seconds",
			Buckets: []float64{1, 5, 10, 30, 60, 120, 300}, // 1s to 5min
		},
	)

	csvUploadsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "csv_uploads_total",
			Help: "Total number of CSV uploads",
		},
		[]string{"status"}, // success, partial, error
	)

	// Database metrics
	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1}, // 1ms to 1s
		},
		[]string{"operation"}, // select, insert, update, delete, bulk_insert
	)

	dbErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_errors_total",
			Help: "Total number of database errors",
		},
		[]string{"operation"},
	)
)

// PrometheusMiddleware creates a middleware that collects Prometheus metrics
func PrometheusMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Increment active connections
		httpActiveConnections.Inc()
		defer httpActiveConnections.Dec()

		// Record request size
		requestSize := len(c.Body())
		path := c.Route().Path
		if path == "" {
			path = c.Path()
		}
		httpRequestSize.WithLabelValues(c.Method(), path).Observe(float64(requestSize))

		// Continue to next handler
		err := c.Next()

		// Record metrics after request completion
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Response().StatusCode())

		// Record request duration
		httpRequestDuration.WithLabelValues(c.Method(), path, status).Observe(duration)

		// Increment request counter
		httpRequestsTotal.WithLabelValues(c.Method(), path, status).Inc()

		// Record response size
		responseSize := len(c.Response().Body())
		httpResponseSize.WithLabelValues(c.Method(), path).Observe(float64(responseSize))

		return err
	}
}

// RecordCSVMetrics records metrics for CSV upload operations
func RecordCSVMetrics(successCount, errorCount int, duration time.Duration, uploadStatus string) {
	csvRowsProcessed.WithLabelValues("success").Add(float64(successCount))
	csvRowsProcessed.WithLabelValues("error").Add(float64(errorCount))
	csvUploadDuration.Observe(duration.Seconds())
	csvUploadsTotal.WithLabelValues(uploadStatus).Inc()
}

// RecordDBMetrics records metrics for database operations
func RecordDBMetrics(operation string, duration time.Duration, err error) {
	dbQueryDuration.WithLabelValues(operation).Observe(duration.Seconds())
	if err != nil {
		dbErrorsTotal.WithLabelValues(operation).Inc()
	}
}
