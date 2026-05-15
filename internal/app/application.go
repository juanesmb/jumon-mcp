package app

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"jumon-mcp/internal/config"
	"jumon-mcp/internal/infrastructure/middleware"
	"jumon-mcp/internal/infrastructure/observability"
	"jumon-mcp/internal/infrastructure/security"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

//go:embed favicon.png
var faviconPNG []byte

func Start() {
	cfg := config.Read()

	observability.ConfigureGlobalLogger()

	rec, shutdownTelemetry, err := observability.Setup(context.Background(), cfg.Observability)
	if err != nil {
		log.Fatalf("telemetry: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		if shutdownErr := shutdownTelemetry(ctx); shutdownErr != nil {
			slog.Error("telemetry_shutdown", "error", shutdownErr.Error())
		}
	}()

	components, err := initComponents(cfg, rec)
	if err != nil {
		log.Fatal(err)
	}
	server := initServer(components)

	verifier, err := security.NewClerkTokenVerifier(
		cfg.Auth.ClerkJWKSURL,
		cfg.Auth.ClerkIssuer,
		cfg.Auth.ClerkAudience,
		cfg.Auth.RequiredScope,
	)
	if err != nil {
		log.Fatalf("clerk token verifier: %v", err)
	}

	metadataPath := "/.well-known/oauth-protected-resource"
	resourceMetadataURL := appendURLPath(cfg.Server.PublicURL, metadataPath)
	if resourceMetadataURL == "" {
		resourceMetadataURL = metadataPath
	}

	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return server
	}, &mcp.StreamableHTTPOptions{JSONResponse: true})

	mux := http.NewServeMux()
	protected := middleware.RequireBearerAuth(verifier, resourceMetadataURL, cfg.Auth.RequiredScope, cfg.Auth.DebugAuth, handler)
	mux.Handle(cfg.Server.Path, protected)
	mux.HandleFunc(metadataPath, oauthProtectedResourceHandler(cfg))
	mux.HandleFunc(metadataPath+cfg.Server.Path, oauthProtectedResourceHandler(cfg))
	mux.HandleFunc("/favicon.ico", faviconHandler)
	mux.HandleFunc("/favicon.png", faviconHandler)

	handlerChain := otelhttp.NewMiddleware("jumon-mcp")(middleware.LoggingHandler(mux, cfg.Observability.GCPProjectID))

	httpServer := &http.Server{
		Handler: handlerChain,
	}

	listener, err := net.Listen("tcp", cfg.Server.BindAddress)
	if err != nil {
		log.Fatalf("listen %s: %v", cfg.Server.BindAddress, err)
	}
	slog.Info("listen", "address", listener.Addr().String(), "path", cfg.Server.Path)

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrCh := make(chan error, 1)
	go func() {
		if err := httpServer.Serve(listener); err != nil {
			serverErrCh <- err
		}
	}()

	select {
	case <-shutdownCtx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			slog.Error("http_shutdown_failed", "error", err.Error())
		}
	case err := <-serverErrCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}
}

func oauthProtectedResourceHandler(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		publicURL := strings.TrimRight(cfg.Server.PublicURL, "/")
		if publicURL == "" {
			scheme := "https"
			if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
				scheme = strings.TrimSpace(strings.Split(proto, ",")[0])
			} else if r.TLS == nil {
				scheme = "http"
			}
			host := requestPublicHost(r)
			publicURL = fmt.Sprintf("%s://%s", scheme, host)
		}

		resource := appendURLPath(publicURL, cfg.Server.Path)
		response := map[string]any{
			"resource":              resource,
			"authorization_servers": []string{cfg.Auth.AuthorizationServerURL},
			"bearer_methods_supported": []string{
				"header",
			},
		}
		if cfg.Auth.RequiredScope != "" {
			response["scopes_supported"] = []string{cfg.Auth.RequiredScope}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "failed to write metadata response", http.StatusInternalServerError)
		}
	}
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(faviconPNG)
}

func appendURLPath(baseURL, path string) string {
	base := strings.TrimRight(baseURL, "/")
	normalized := "/" + strings.TrimLeft(path, "/")
	if base == "" {
		return ""
	}
	return base + normalized
}

// requestPublicHost returns the host clients use to reach this service. Prefer proxy
// headers so OAuth "resource" matches the public custom domain behind Cloud Run / LB.
func requestPublicHost(r *http.Request) string {
	if fh := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); fh != "" {
		return strings.TrimSpace(strings.Split(fh, ",")[0])
	}
	return r.Host
}
