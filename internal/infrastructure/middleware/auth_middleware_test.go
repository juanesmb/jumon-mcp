package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"jumon-mcp/internal/infrastructure/security"
)

type stubVerifier struct {
	userID string
}

func (s stubVerifier) Verify(_ context.Context, _ string) (security.AuthClaims, error) {
	return security.AuthClaims{UserID: s.userID}, nil
}

func TestRequireBearerAuth_orgContextFromURLOnly(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantOrgID string
	}{
		{name: "personal URL", query: "", wantOrgID: ""},
		{name: "org URL", query: "?org=org_abc", wantOrgID: "org_abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotOrgID string
			handler := RequireBearerAuth(
				stubVerifier{userID: "user_1"},
				"https://example.com/.well-known/oauth-protected-resource",
				"",
				false,
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					gotOrgID = OrgIDFromContext(r.Context())
					w.WriteHeader(http.StatusOK)
				}),
			)

			req := httptest.NewRequest(http.MethodPost, "/mcp"+tt.query, nil)
			req.Header.Set("Authorization", "Bearer test-token")
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}
			if gotOrgID != tt.wantOrgID {
				t.Fatalf("OrgIDFromContext() = %q, want %q", gotOrgID, tt.wantOrgID)
			}
		})
	}
}
