package umamusume

import (
	"context"
	"os"

	"golang.org/x/exp/slog"
)

type loggerCtx struct{}

func NewLogger(level slog.Level) *slog.Logger {
	return slog.New(
		slog.HandlerOptions{
			AddSource: true,
			Level:     level,
		}.NewJSONHandler(os.Stdout))
}

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerCtx{}, logger)
}

func GetLogger(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(loggerCtx{}).(*slog.Logger)
	if !ok {
		panic("logger is not stored")
	}
	return logger
}
