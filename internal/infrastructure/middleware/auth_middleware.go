package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"jumon-mcp/internal/infrastructure/security"
)

type authContextKey struct{}

type TokenVerifier interface {
	Verify(ctx context.Context, token string) (string, error)
}

func RequireBearerAuth(verifier TokenVerifier, resourceMetadataURL, requiredScope string, debugAuth bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		token := extractBearerToken(authHeader)
		if debugAuth {
			log.Printf("[AUTH] %s %s remote=%s has_header=%t token_len=%d", r.Method, r.URL.Path, r.RemoteAddr, authHeader != "", len(token))
		}

		if token == "" {
			writeUnauthorized(w, resourceMetadataURL, requiredScope, "missing bearer token")
			return
		}

		userID, err := verifier.Verify(r.Context(), token)
		if err != nil {
			if debugAuth {
				log.Printf("[AUTH] verify failed: %v | claims=%s", err, security.PeekUnverifiedClaims(token))
			}
			writeUnauthorized(w, resourceMetadataURL, requiredScope, err.Error())
			return
		}

		ctx := context.WithValue(r.Context(), authContextKey{}, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(authContextKey{}).(string)
	return userID, ok && userID != ""
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
