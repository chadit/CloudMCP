package logger

import (
	"context"
	"log/slog"
	"os"
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

type slogWrapper struct {
	logger *slog.Logger
}

func New(level string) Logger {
	var logLevel slog.Level
	switch level {
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

	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
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