package http

import "context"

type attemptContextKey struct{}

// ContextWithAttempt annotates outbound HTTP retries for observability (1-based attempts).
func ContextWithAttempt(parent context.Context, attempt int) context.Context {
	return context.WithValue(parent, attemptContextKey{}, attempt)
}

// AttemptFromContext returns HTTP attempt sequence (defaults to 1 when absent).
func AttemptFromContext(ctx context.Context) int {
	v, ok := ctx.Value(attemptContextKey{}).(int)
	if !ok || v < 1 {
		return 1
	}
	return v
}
