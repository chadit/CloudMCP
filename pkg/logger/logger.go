package logger

import (
	"context"
	"io"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger interface {
	Debug(msg string, args ...any)
	DebugContext(ctx context.Context, msg string, args ...any)
	Info(msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	Warn(msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	Error(msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
	With(args ...any) Logger
}

type LogConfig struct {
	Level      string
	FilePath   string // If empty, log to stderr
	MaxSize    int    // MB
	MaxBackups int    // Number of files
	MaxAge     int    // Days
}

type slogWrapper struct {
	logger *slog.Logger
}

// New creates a logger that outputs to stderr (for backwards compatibility).
func New(level string) Logger {
	return NewWithConfig(LogConfig{
		Level: level,
	})
}

// NewWithConfig creates a logger with the specified configuration.
func NewWithConfig(config LogConfig) Logger {
	var logLevel slog.Level

	switch config.Level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	var writer io.Writer
	if config.FilePath != "" {
		// Use lumberjack for file rotation
		writer = &lumberjack.Logger{
			Filename:   config.FilePath,
			MaxSize:    config.MaxSize,    // MB
			MaxBackups: config.MaxBackups, // Number of files
			MaxAge:     config.MaxAge,     // Days
			Compress:   true,              // Compress rotated files
		}
	} else {
		// Default to stderr
		writer = os.Stderr
	}

	logger := slog.New(slog.NewJSONHandler(writer, opts))

	return &slogWrapper{logger: logger}
}

func (l *slogWrapper) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *slogWrapper) DebugContext(ctx context.Context, msg string, args ...any) {
	l.logger.DebugContext(ctx, msg, args...)
}

func (l *slogWrapper) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *slogWrapper) InfoContext(ctx context.Context, msg string, args ...any) {
	l.logger.InfoContext(ctx, msg, args...)
}

func (l *slogWrapper) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *slogWrapper) WarnContext(ctx context.Context, msg string, args ...any) {
	l.logger.WarnContext(ctx, msg, args...)
}

func (l *slogWrapper) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

func (l *slogWrapper) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.logger.ErrorContext(ctx, msg, args...)
}

func (l *slogWrapper) With(args ...any) Logger {
	return &slogWrapper{logger: l.logger.With(args...)}
}
