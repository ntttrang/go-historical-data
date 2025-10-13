package logger

import (
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

// LogstashWriter implements io.Writer interface for sending logs to Logstash
type LogstashWriter struct {
	conn     net.Conn
	address  string
	mu       sync.Mutex
	fallback io.Writer
}

// NewLogstashWriter creates a new Logstash writer
func NewLogstashWriter(host string, port int) (*LogstashWriter, error) {
	address := fmt.Sprintf("%s:%d", host, port)

	writer := &LogstashWriter{
		address:  address,
		fallback: os.Stdout,
	}

	// Try to establish initial connection
	if err := writer.connect(); err != nil {
		// Log to stdout if Logstash is not available yet
		fmt.Fprintf(os.Stderr, "Warning: Failed to connect to Logstash at %s: %v. Logs will be written to stdout.\n", address, err)
		// Don't return error - allow application to start even if Logstash is not ready
	}

	return writer, nil
}

// connect establishes a TCP connection to Logstash
func (w *LogstashWriter) connect() error {
	conn, err := net.DialTimeout("tcp", w.address, 5*time.Second)
	if err != nil {
		return err
	}
	w.conn = conn
	return nil
}

// Write implements io.Writer interface
func (w *LogstashWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Try to write to Logstash
	if w.conn != nil {
		// Set write deadline to prevent hanging
		if err := w.conn.SetWriteDeadline(time.Now().Add(2 * time.Second)); err == nil {
			n, err = w.conn.Write(p)
			if err == nil {
				return n, nil
			}
		}

		// Connection failed, close it
		_ = w.conn.Close() // Explicitly ignore close error in cleanup path
		w.conn = nil
	}

	// Try to reconnect
	if err := w.connect(); err == nil {
		if err := w.conn.SetWriteDeadline(time.Now().Add(2 * time.Second)); err == nil {
			n, err = w.conn.Write(p)
			if err == nil {
				return n, nil
			}
		}
	}

	// Fallback to stdout if Logstash is unavailable
	return w.fallback.Write(p)
}

// Close closes the connection to Logstash
func (w *LogstashWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.conn != nil {
		err := w.conn.Close()
		w.conn = nil
		return err
	}
	return nil
}

// MultiWriter creates a writer that writes to multiple destinations
type MultiWriter struct {
	writers []io.Writer
}

// NewMultiWriter creates a new multi-writer
func NewMultiWriter(writers ...io.Writer) *MultiWriter {
	return &MultiWriter{
		writers: writers,
	}
}

// Write implements io.Writer interface
func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		n, err = w.Write(p)
		if err != nil {
			return n, err
		}
	}
	return len(p), nil
}
