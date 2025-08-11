package utils

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

func SetupLogging(level slog.Level) error {
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	now := time.Now()
	logFilename := fmt.Sprintf("app-%s.log", now.Format("2006-01-02"))
	logPath := filepath.Join(logsDir, logFilename)

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", logPath, err)
	}

	consoleHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.String("time", a.Value.Time().Format("15:04:05"))
			}
			return a
		},
	})

	fileHandler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.String("timestamp", a.Value.Time().Format(time.RFC3339))
			}
			return a
		},
	})

	multiHandler := &MultiHandler{
		handlers: []slog.Handler{consoleHandler, fileHandler},
		file:     logFile,
		filePath: logPath,
	}

	slog.SetDefault(slog.New(multiHandler))

	slog.Info("Logging configured",
		slog.String("console_level", level.String()),
		slog.String("file_level", level.String()),
		slog.String("log_file", logPath),
		slog.String("logs_directory", logsDir),
	)

	return nil
}

type MultiHandler struct {
	handlers []slog.Handler
	file     *os.File
	filePath string
}

func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	var lastErr error
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r); err != nil {
				lastErr = err
			}
		}
	}
	return lastErr
}

func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return &MultiHandler{
		handlers: handlers,
		file:     h.file,
		filePath: h.filePath,
	}
}

func (h *MultiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return &MultiHandler{
		handlers: handlers,
		file:     h.file,
		filePath: h.filePath,
	}
}

func (h *MultiHandler) GetLogFilePath() string {
	return h.filePath
}

func (h *MultiHandler) Close() error {
	if h.file != nil {
		return h.file.Close()
	}
	return nil
}
