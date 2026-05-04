package observability

import (
	stdhttp "net/http"
	neturl "net/url"
	"path"
	"strings"
)

const unknownRoute = "(unknown)"

// GatewayRoutePattern maps internal gateway URLs to low-cardinality route templates.
func GatewayRoutePattern(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return unknownRoute
	}
	u, err := neturl.Parse(rawURL)
	if err != nil {
		return unknownRoute
	}
	p := path.Clean(u.Path)
	if !strings.HasPrefix(p, "/") {
		p = "/" + strings.TrimPrefix(p, "/")
	}

	switch {
	case strings.HasPrefix(p, "/api/internal/connections/") && strings.HasSuffix(p, "/current"):
		return "/api/internal/connections/{provider}/current"
	case strings.HasPrefix(p, "/api/internal/providers/") && strings.HasSuffix(p, "/refresh"):
		return "/api/internal/providers/{provider}/refresh"
	case strings.HasPrefix(p, "/api/internal/providers/") && strings.HasSuffix(p, "/proxy"):
		return "/api/internal/providers/{provider}/proxy"
	default:
		if p == "" || p == "." {
			return unknownRoute
		}
		return p
	}
}

// ProviderFromGatewayURL extracts the provider segment from known gateway URL shapes, or "gateway" as a fallback.
func ProviderFromGatewayURL(rawURL string) string {
	u, err := neturl.Parse(strings.TrimSpace(rawURL))
	if err != nil || u.Path == "" {
		return "gateway"
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	for i, p := range parts {
		if p == "connections" && i+1 < len(parts) {
			return parts[i+1]
		}
		if p == "providers" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return "gateway"
}

// StatusClass returns a coarse HTTP status grouping for metrics labels.
func StatusClass(status int) string {
	switch {
	case status >= 500:
		return "5xx"
	case status >= 400:
		return "4xx"
	case status >= 200 && status <= 299:
		return "2xx"
	case status >= 300 && status <= 399:
		return "3xx"
	case status == 0:
		return "error"
	default:
		return "other"
	}
}

// NormalizeHTTPMethod returns an uppercase method or GET if empty.
func NormalizeHTTPMethod(method string) string {
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		return stdhttp.MethodGet
	}
	if allowedMethod(method) {
		return method
	}
	return "OTHER"
}

func allowedMethod(method string) bool {
	switch method {
	case stdhttp.MethodGet, stdhttp.MethodPost, stdhttp.MethodPut, stdhttp.MethodPatch,
		stdhttp.MethodDelete, stdhttp.MethodHead, stdhttp.MethodOptions:
		return true
	default:
		return false
	}
}
