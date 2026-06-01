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

		// Resolve org ID: URL ?org= param takes precedence over the JWT org_id claim.
		// This allows per-org MCP URLs (mcp.jumon.ai/mcp?org=org_xxx) to override
		// the JWT claim, which may be absent when using OAuth-based AI agent auth.
		orgID := strings.TrimSpace(r.URL.Query().Get("org"))
		if orgID == "" {
			orgID = claims.OrgID
		}

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

// OrgIDFromContext returns the resolved organization ID.
// The ?org= URL query parameter takes precedence over the JWT org_id claim,
// enabling per-org MCP URLs (mcp.jumon.ai/mcp?org=org_xxx) for context switching.
// Returns "" when the user is acting in personal workspace.
func OrgIDFromContext(ctx context.Context) string {
	if orgID, ok := ctx.Value(orgIDContextKey{}).(string); ok {
		return orgID
	}
	// context key absent: caller built the context directly (e.g. unit tests) without RequireBearerAuth.
	claims, ok := ctx.Value(authContextKey{}).(security.AuthClaims)
	if !ok {
		return ""
	}
	return claims.OrgID
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
