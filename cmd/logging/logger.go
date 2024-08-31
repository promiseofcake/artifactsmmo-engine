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
	return ctx.Value(loggerKey).(*slog.Logger)
}
