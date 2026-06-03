package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"jumon-mcp/internal/infrastructure/security"
)

type authContextKey struct{}

// orgIDContextKey is the context key for the request-scoped resolved org ID.
type orgIDContextKey struct{}

type TokenVerifier interface {
	Verify(ctx context.Context, token string) (security.AuthClaims, error)
}

func RequireBearerAuth(verifier TokenVerifier, resourceMetadataURL, requiredScope string, debugAuth bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		token := extractBearerToken(authHeader)
		if debugAuth {
			slog.WarnContext(r.Context(), "mcp_debug_auth",
				"http.method", r.Method,
				"http.path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
				"has_authorization_header", authHeader != "",
				"bearer_token_len", len(token),
			)
		}

		if token == "" {
			writeUnauthorized(w, resourceMetadataURL, requiredScope, "missing bearer token")
			return
		}

		claims, err := verifier.Verify(r.Context(), token)
		if err != nil {
			if debugAuth {
				slog.WarnContext(r.Context(), "mcp_debug_auth_verify_failed",
					"error", err.Error(),
					"claims_peek", security.PeekUnverifiedClaims(token),
				)
			}
			writeUnauthorized(w, resourceMetadataURL, requiredScope, err.Error())
			return
		}

		orgID := strings.TrimSpace(r.URL.Query().Get("org"))

		ctx := context.WithValue(r.Context(), authContextKey{}, claims)
		ctx = context.WithValue(ctx, orgIDContextKey{}, orgID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UserIDFromContext returns the Clerk user ID stored by RequireBearerAuth.
// Preserved for backward compatibility with existing callers.
func UserIDFromContext(ctx context.Context) (string, bool) {
	claims, ok := ctx.Value(authContextKey{}).(security.AuthClaims)
	if !ok || claims.UserID == "" {
		return "", false
	}
	return claims.UserID, true
}

// OrgIDFromContext returns the MCP URL ?org= value (empty = personal workspace).
func OrgIDFromContext(ctx context.Context) string {
	orgID, _ := ctx.Value(orgIDContextKey{}).(string)
	return orgID
}

func writeUnauthorized(w http.ResponseWriter, resourceMetadataURL, requiredScope, detail string) {
	challenge := fmt.Sprintf(`Bearer resource_metadata="%s"`, resourceMetadataURL)
	if requiredScope != "" {
		challenge += fmt.Sprintf(`, scope="%s"`, requiredScope)
	}
	w.Header().Set("WWW-Authenticate", challenge)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":   "unauthorized",
		"message": detail,
	})
}

func extractBearerToken(header string) string {
	parts := strings.Fields(strings.TrimSpace(header))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
