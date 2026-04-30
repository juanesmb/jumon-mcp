package local

import (
	"context"
	"log/slog"
	"os"
)

type Logger struct {
	client *slog.Logger
}

func NewLogger() *Logger {
	return &Logger{
		client: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}
}

func (l *Logger) Info(ctx context.Context, message string, tags map[string]string) {
	l.client.InfoContext(ctx, message, flatten(tags)...)
}

func (l *Logger) Warn(ctx context.Context, message string, tags map[string]string) {
	l.client.WarnContext(ctx, message, flatten(tags)...)
}

func (l *Logger) Error(ctx context.Context, message string, tags map[string]string) {
	l.client.ErrorContext(ctx, message, flatten(tags)...)
}

func flatten(tags map[string]string) []any {
	out := make([]any, 0, len(tags)*2)
	for key, value := range tags {
		out = append(out, key, value)
	}
	return out
}
