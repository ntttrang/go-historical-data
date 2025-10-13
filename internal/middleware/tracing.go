package middleware

import (
	fiber "github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	tracerName = "go-historical-data-api"
)

// Tracing returns a middleware that creates spans for incoming HTTP requests
func Tracing() fiber.Handler {
	tracer := otel.Tracer(tracerName)

	return func(c *fiber.Ctx) error {
		// Extract trace context from incoming request headers
		propagator := otel.GetTextMapPropagator()
		ctx := propagator.Extract(c.Context(), &fiberHeaderCarrier{c: c})

		// Start a new span
		spanName := c.Method() + " " + c.Route().Path
		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.method", c.Method()),
				attribute.String("http.url", c.OriginalURL()),
				attribute.String("http.route", c.Route().Path),
				attribute.String("http.scheme", c.Protocol()),
				attribute.String("http.host", c.Hostname()),
				attribute.String("http.user_agent", c.Get("User-Agent")),
				attribute.String("http.client_ip", c.IP()),
			),
		)
		defer span.End()

		// Store span in fiber context for later use
		c.Locals("trace_id", span.SpanContext().TraceID().String())
		c.Locals("span_id", span.SpanContext().SpanID().String())
		c.Locals("span", span)

		// Update the request context with trace context
		c.SetUserContext(ctx)

		// Process request
		err := c.Next()

		// Record response status
		statusCode := c.Response().StatusCode()
		span.SetAttributes(
			attribute.Int("http.status_code", statusCode),
			attribute.Int64("http.response_size", int64(len(c.Response().Body()))),
		)

		// Mark span as error if status code is 4xx or 5xx
		if statusCode >= 400 {
			span.SetStatus(codes.Error, fiber.ErrInternalServerError.Message)
			if err != nil {
				span.RecordError(err)
			}
		} else {
			span.SetStatus(codes.Ok, "")
		}

		return err
	}
}

// fiberHeaderCarrier adapts fiber.Ctx to satisfy the propagation.TextMapCarrier interface
type fiberHeaderCarrier struct {
	c *fiber.Ctx
}

// Get returns the value associated with the passed key.
func (f *fiberHeaderCarrier) Get(key string) string {
	return f.c.Get(key)
}

// Set stores the key-value pair.
func (f *fiberHeaderCarrier) Set(key string, value string) {
	f.c.Set(key, value)
}

// Keys lists the keys stored in this carrier.
func (f *fiberHeaderCarrier) Keys() []string {
	keys := make([]string, 0)
	f.c.Request().Header.All()(func(key, _ []byte) bool {
		keys = append(keys, string(key))
		return true
	})
	return keys
}
