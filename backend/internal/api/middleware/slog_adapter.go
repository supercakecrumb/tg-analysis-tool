package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// SlogAdapter implements middleware.LogFormatter interface for slog
type SlogAdapter struct {
	Logger *slog.Logger
}

// NewLogEntry creates a new log entry
func (a *SlogAdapter) NewLogEntry(r *http.Request) middleware.LogEntry {
	entry := &SlogEntry{
		Logger: a.Logger,
		fields: []any{
			"method", r.Method,
			"path", r.URL.Path,
			"remote", r.RemoteAddr,
			"request_id", middleware.GetReqID(r.Context()),
		},
	}

	return entry
}

// Print implements middleware.LoggerInterface required method
func (a *SlogAdapter) Print(v ...interface{}) {
	a.Logger.Info(fmt.Sprint(v...))
}

// slogEntry implements middleware.LogEntry interface for slog
type SlogEntry struct {
	Logger *slog.Logger
	fields []any
}

// Write writes the log entry
func (e *SlogEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra any) {
	e.fields = append(e.fields,
		"status", status,
		"bytes", bytes,
		"elapsed", elapsed,
	)

	e.Logger.Info("Request completed", e.fields...)
}

// Panic writes the panic log entry
func (e *SlogEntry) Panic(v any, stack []byte) {
	e.fields = append(e.fields,
		"stack", string(stack),
		"panic", v,
	)

	e.Logger.Error("Request panic", e.fields...)
}

// NewLogFormatter creates a new slog-based log formatter for chi
func NewLogFormatter(logger *slog.Logger) *middleware.DefaultLogFormatter {
	return &middleware.DefaultLogFormatter{
		Logger: &SlogAdapter{Logger: logger},
	}
}
