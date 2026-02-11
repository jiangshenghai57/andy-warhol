package logger

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

type Logger struct {
	*slog.Logger
}

// NewLogger creates a structured logger with dual output (file + stdout)
func NewLogger(logDir string) (*Logger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	logFile := filepath.Join(logDir, time.Now().Format("2006-01-02")+".log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	// Dual output: file (JSON) + stdout (text for readability)
	multiWriter := io.MultiWriter(file, os.Stdout)

	handler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true, // Include file:line in logs
	})

	return &Logger{slog.New(handler)}, nil
}

// Usage example
func ExampleUsage() {
	logger, _ := NewLogger("./logs")

	// Structured logging
	logger.Info("processing loan",
		slog.String("loan_id", "LOAN001"),
		slog.Float64("face", 250000),
		slog.Int("wam", 360),
	)

	var err error
	err = errors.New("calculation failed: invalid input")

	logger.Error("calculation failed",
		slog.String("loan_id", "LOAN002"),
		slog.Any("error", err),
	)
}
