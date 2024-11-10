package log

import (
	"io"
	"log/slog"
)

func NewDevelopmentLogger(w io.Writer) *slog.Logger {
	handler := slog.NewTextHandler(w,
		&slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		},
	)

	return slog.New(handler)
}

func NewProductionLogger(w io.Writer) *slog.Logger {
	handler := slog.NewJSONHandler(w,
		&slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		},
	)

	return slog.New(handler)
}
