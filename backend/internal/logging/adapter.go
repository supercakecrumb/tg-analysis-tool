package logging

import (
	"fmt"
	"log/slog"
)

// Adapter wraps slog.Logger to implement middleware.LoggerInterface
type Adapter struct {
	logger *slog.Logger
}

func (a *Adapter) Print(v ...interface{}) {
	a.logger.Info(fmt.Sprint(v...))
}

// NewAdapter creates a new logging adapter
func NewAdapter(logger *slog.Logger) *Adapter {
	return &Adapter{logger: logger}
}
