package observability

import (
	"log/slog"
	"os"
)

// ConfigureGlobalLogger sets the default slog logger to GCP-friendly JSON stdout (severity key for Cloud Logging).
func ConfigureGlobalLogger() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key != slog.LevelKey {
				return a
			}
			lvl := a.Value.Any().(slog.Level)
			val := severityForGCP(lvl)
			return slog.String("severity", val)
		},
	})
	slog.SetDefault(slog.New(handler))
}

func severityForGCP(l slog.Level) string {
	switch {
	case l >= slog.LevelError:
		return "ERROR"
	case l >= slog.LevelWarn:
		return "WARNING"
	case l >= slog.LevelInfo:
		return "INFO"
	default:
		return "DEFAULT"
	}
}
