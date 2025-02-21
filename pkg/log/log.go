package log

import (
	"io"
	"log/slog"
)

// NewDevelopmentLogger creates a new development logger
func NewDevelopmentLogger(w io.Writer) *slog.Logger {
	handler := slog.NewTextHandler(w,
		&slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		},
	)

	return slog.New(handler)
}

// NewProductionLogger creates a new production logger
func NewProductionLogger(w io.Writer) *slog.Logger {
	handler := slog.NewJSONHandler(w,
		&slog.HandlerOptions{
			AddSource: false,
			Level:     slog.LevelInfo,
		},
	)

	return slog.New(handler)
}
