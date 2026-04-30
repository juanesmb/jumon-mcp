package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"jumon-mcp/internal/config"
	"jumon-mcp/internal/infrastructure/middleware"
	"jumon-mcp/internal/infrastructure/security"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func Start() {
	cfg := config.Read()

	components, err := initComponents(cfg)
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

	httpServer := &http.Server{
		Handler: middleware.LoggingHandler(mux),
	}

	listener, err := net.Listen("tcp", cfg.Server.BindAddress)
	if err != nil {
		log.Fatalf("listen %s: %v", cfg.Server.BindAddress, err)
	}
	log.Printf("Jumon MCP facade listening on %s (path %s)", listener.Addr().String(), cfg.Server.Path)

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
			log.Printf("graceful shutdown failed: %v", err)
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
				scheme = proto
			} else if r.TLS == nil {
				scheme = "http"
			}
			publicURL = fmt.Sprintf("%s://%s", scheme, r.Host)
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

func appendURLPath(baseURL, path string) string {
	base := strings.TrimRight(baseURL, "/")
	normalized := "/" + strings.TrimLeft(path, "/")
	if base == "" {
		return ""
	}
	return base + normalized
}
