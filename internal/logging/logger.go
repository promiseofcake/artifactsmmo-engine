package logging

import (
	"context"
	"log/slog"
)

const (
	loggerKey = "logger"
)

func ContextWithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

func Get(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(loggerKey).(*slog.Logger)
	if !ok {
		slog.Error("failed to get logger from context")
		return slog.Default()
	}
	return logger
}
