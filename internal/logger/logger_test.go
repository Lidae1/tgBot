package logger

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected slog.Level
	}{
		{"Debug level", "debug", slog.LevelDebug},
		{"Info level", "info", slog.LevelInfo},
		{"Warn level", "warn", slog.LevelWarn},
		{"Error level", "error", slog.LevelError},
		{"Unknow level", "unknow", slog.LevelInfo},
		{"Empty level", "", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.level)
			assert.NotNil(t, logger)

			var buf bytes.Buffer
			handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
				Level: tt.expected,
			})

			testLogger := slog.New(handler)

			assert.NotNil(t, testLogger)
		})
	}
}

func TestLoggerLevels(t *testing.T) {
	logger := NewLogger("debug")
	assert.NotNil(t, logger)

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")
}
