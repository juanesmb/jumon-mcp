package security

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

type ClerkTokenVerifier struct {
	jwksMu        sync.Mutex
	jwks          *keyfunc.JWKS
	jwksURL       string
	issuer        string
	audience      string
	requiredScope string
}

func NewClerkTokenVerifier(jwksURL, issuer, audience, requiredScope string) (*ClerkTokenVerifier, error) {
	if strings.TrimSpace(jwksURL) == "" {
		return nil, fmt.Errorf("CLERK_JWKS_URL is required")
	}
	if strings.TrimSpace(issuer) == "" {
		return nil, fmt.Errorf("CLERK_ISSUER is required")
	}
	return &ClerkTokenVerifier{
		jwksURL:       jwksURL,
		issuer:        issuer,
		audience:      audience,
		requiredScope: requiredScope,
	}, nil
}

func (v *ClerkTokenVerifier) Verify(ctx context.Context, tokenString string) (string, error) {
	_ = ctx
	jwks, err := v.getJWKS()
	if err != nil {
		return "", err
	}

	options := []jwt.ParserOption{
		jwt.WithValidMethods([]string{"RS256", "RS384", "RS512"}),
		jwt.WithIssuer(v.issuer),
	}
	if v.audience != "" {
		options = append(options, jwt.WithAudience(v.audience))
	}

	token, err := jwt.Parse(tokenString, jwks.Keyfunc, options...)
	if err != nil {
		return "", fmt.Errorf("token validation failed: %w", err)
	}
	if !token.Valid {
		return "", fmt.Errorf("token is invalid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("token claims are not map claims")
	}
	userID, _ := claims["sub"].(string)
	if strings.TrimSpace(userID) == "" {
		return "", fmt.Errorf("token missing subject")
	}
	if v.requiredScope != "" && !hasScope(claims, v.requiredScope) {
		return "", fmt.Errorf("token missing required scope %q", v.requiredScope)
	}
	return userID, nil
}

func (v *ClerkTokenVerifier) getJWKS() (*keyfunc.JWKS, error) {
	v.jwksMu.Lock()
	defer v.jwksMu.Unlock()
	if v.jwks != nil {
		return v.jwks, nil
	}

	jwks, err := keyfunc.Get(v.jwksURL, keyfunc.Options{
		RefreshInterval:   time.Hour,
		RefreshRateLimit:  5 * time.Minute,
		RefreshTimeout:    10 * time.Second,
		RefreshUnknownKID: true,
	})
	if err != nil {
		return nil, fmt.Errorf("initialize JWKS: %w", err)
	}
	v.jwks = jwks
	return jwks, nil
}

func hasScope(claims jwt.MapClaims, required string) bool {
	if scopeRaw, ok := claims["scope"]; ok {
		if scopeString, ok := scopeRaw.(string); ok {
			for _, scope := range strings.Fields(scopeString) {
				if scope == required {
					return true
				}
			}
		}
	}
	if scopeRaw, ok := claims["scp"]; ok {
		if scopeList, ok := scopeRaw.([]interface{}); ok {
			for _, value := range scopeList {
				if scopeString, ok := value.(string); ok && scopeString == required {
					return true
				}
			}
		}
	}
	return false
}
