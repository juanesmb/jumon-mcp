package gateway

import (
	"testing"

	infrahttp "jumon-mcp/internal/infrastructure/http"
)

func TestIsProviderUsable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
		want bool
	}{
		{
			name: "usable true",
			body: `{"connected":true,"usable":true,"health":"active"}`,
			want: true,
		},
		{
			name: "needs reconnect",
			body: `{"connected":true,"usable":false,"health":"needs_reconnect"}`,
			want: false,
		},
		{
			name: "legacy connected only",
			body: `{"connected":true,"health":"active"}`,
			want: true,
		},
		{
			name: "legacy connected needs reconnect health",
			body: `{"connected":true,"health":"needs_reconnect"}`,
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resp := &infrahttp.Response{StatusCode: 200, Body: []byte(tc.body)}
			if got := IsProviderUsable(resp); got != tc.want {
				t.Fatalf("IsProviderUsable() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRefreshSucceeded(t *testing.T) {
	t.Parallel()

	ok := &infrahttp.Response{StatusCode: 200, Body: []byte(`{"refreshed":true}`)}
	if !RefreshSucceeded(ok) {
		t.Fatal("expected refresh success")
	}

	fail := &infrahttp.Response{StatusCode: 200, Body: []byte(`{"refreshed":false,"reason":"provider_error"}`)}
	if RefreshSucceeded(fail) {
		t.Fatal("expected refresh failure")
	}
}

func TestIsTokenRefreshFailed(t *testing.T) {
	t.Parallel()

	resp := &infrahttp.Response{
		StatusCode: 401,
		Body:       []byte(`{"code":"TOKEN_REFRESH_FAILED","message":"reconnect"}`),
	}
	if !IsTokenRefreshFailed(resp) {
		t.Fatal("expected token refresh failed")
	}
	if !IsNotConnectedResponse(resp) {
		t.Fatal("expected not connected response")
	}
}
