package meta

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"jumon-mcp/internal/domain/catalog"
	"jumon-mcp/internal/infrastructure/gateway"
	infrahttp "jumon-mcp/internal/infrastructure/http"
)

func TestDecodeMetaProxy_platformNotConnected(t *testing.T) {
	gw := gateway.NewClient(infrahttp.NewClient(nil, nil), "http://gateway", "secret", "https://app/connect")
	resp := &infrahttp.Response{StatusCode: http.StatusNotFound, Body: []byte(`{"code":"CONNECTION_NOT_FOUND"}`)}

	_, err := decodeMetaProxy(resp, nil, gw)
	if err == nil {
		t.Fatal("expected error")
	}
	var notConnected *catalog.PlatformNotConnectedError
	if !errors.As(err, &notConnected) {
		t.Fatalf("expected PlatformNotConnectedError, got %T: %v", err, err)
	}
	if notConnected.Platform != platformName {
		t.Fatalf("platform = %q, want %q", notConnected.Platform, platformName)
	}
}

func TestDecodeMetaProxy_success(t *testing.T) {
	body := []byte(`{"id":"123"}`)
	raw, err := decodeMetaProxy(&infrahttp.Response{StatusCode: 200, Body: body}, nil, gateway.NewClient(nil, "", "", ""))
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != string(body) {
		t.Fatalf("body = %s", raw)
	}
}

func TestMetaGateway_getWithRefresh(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/internal/providers/meta/proxy" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		gotPath = "ok"
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"user"}`))
	}))
	defer srv.Close()

	client := gateway.NewClient(infrahttp.NewClient(nil, nil), srv.URL, "secret", "https://connect")
	gw := newMetaGateway(client)
	raw, err := gw.getWithRefresh(context.Background(), "meta_smoke", "user_1", "me", map[string]string{"fields": "id"})
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "ok" {
		t.Fatal("proxy not called")
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["id"] != "user" {
		t.Fatalf("payload = %v", payload)
	}
}
