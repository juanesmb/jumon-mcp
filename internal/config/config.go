package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Server        ServerConfig
	Auth          AuthConfig
	Gateway       GatewayConfig
	Observability ObservabilityConfig
}

type ServerConfig struct {
	BindAddress string
	Port        string
	Path        string
	PublicURL   string
}

type AuthConfig struct {
	ClerkIssuer            string
	ClerkJWKSURL           string
	ClerkAudience          string
	RequiredScope          string
	AuthorizationServerURL string
	DebugAuth              bool
}

type GatewayConfig struct {
	BaseURL            string
	InternalSecret     string
	ConnectURL         string
	LinkedInAPIBaseURL string
	GoogleAPIVersion   string
}

// ObservabilityConfig controls OpenTelemetry exporters (Cloud Trace / Cloud Monitoring) and log correlation fields.
type ObservabilityConfig struct {
	Enabled          bool
	GCPProjectID     string
	ServiceName      string
	TraceSampleRatio float64
	UserIDHashSalt   string
}

func Read() Config {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}

	host := "0.0.0.0"
	path := "/mcp"

	clerkIssuer := strings.TrimSpace(os.Getenv("CLERK_ISSUER"))
	authServerURL := strings.TrimSpace(os.Getenv("AUTHORIZATION_SERVER_URL"))
	if authServerURL == "" {
		authServerURL = clerkIssuer
	}

	gatewayBaseURL := gatewayBaseURL()
	linkedInBaseURL := strings.TrimSpace(os.Getenv("LINKEDIN_API_BASE_URL"))
	if linkedInBaseURL == "" {
		linkedInBaseURL = "https://api.linkedin.com/rest"
	}
	googleAPIVersion := strings.TrimSpace(os.Getenv("GOOGLE_ADS_API_VERSION"))
	if googleAPIVersion == "" {
		googleAPIVersion = "v22"
	}

	gcpProject := strings.TrimSpace(os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if gcpProject == "" {
		gcpProject = strings.TrimSpace(os.Getenv("GCP_PROJECT"))
	}

	return Config{
		Server: ServerConfig{
			BindAddress: fmt.Sprintf("%s:%s", host, port),
			Port:        port,
			Path:        path,
			PublicURL:   strings.TrimSpace(os.Getenv("PUBLIC_BASE_URL")),
		},
		Auth: AuthConfig{
			ClerkIssuer:            clerkIssuer,
			ClerkJWKSURL:           strings.TrimSpace(os.Getenv("CLERK_JWKS_URL")),
			ClerkAudience:          strings.TrimSpace(os.Getenv("CLERK_AUDIENCE")),
			RequiredScope:          strings.TrimSpace(os.Getenv("MCP_REQUIRED_SCOPE")),
			AuthorizationServerURL: authServerURL,
			DebugAuth:              envTruthy("MCP_DEBUG_AUTH"),
		},
		Gateway: GatewayConfig{
			BaseURL:            gatewayBaseURL,
			InternalSecret:     gatewayInternalSecret(),
			ConnectURL:         deriveConnectURL(gatewayBaseURL),
			LinkedInAPIBaseURL: linkedInBaseURL,
			GoogleAPIVersion:   googleAPIVersion,
		},
		Observability: readObservabilityConfig(gcpProject),
	}
}

func readObservabilityConfig(gcpProjectFromEnv string) ObservabilityConfig {
	enabledOverride := strings.TrimSpace(os.Getenv("OBSERVABILITY_ENABLED"))

	serviceName := strings.TrimSpace(os.Getenv("OTEL_SERVICE_NAME"))
	if serviceName == "" {
		serviceName = "jumon-mcp"
	}

	sampleRatio := 1.0
	if raw := strings.TrimSpace(os.Getenv("OBSERVABILITY_TRACE_SAMPLE_RATIO")); raw != "" {
		if v, err := parsePositiveFloat(raw); err == nil && v >= 0 && v <= 1 {
			sampleRatio = v
		}
	}

	gcpProject := gcpProjectFromEnv
	if gcpProject == "" {
		gcpProject = strings.TrimSpace(os.Getenv("GOOGLE_CLOUD_PROJECT"))
	}
	if gcpProject == "" {
		gcpProject = strings.TrimSpace(os.Getenv("GCP_PROJECT"))
	}

	enabledDefault := false
	switch strings.ToLower(enabledOverride) {
	case "1", "true", "yes", "on":
		enabledDefault = true
	case "0", "false", "no", "off":
		enabledDefault = false
	default:
		// On Cloud Run a project ID is normally present; enabling avoids silent no-export surprises.
		enabledDefault = gcpProject != ""
	}

	return ObservabilityConfig{
		Enabled:          enabledDefault,
		GCPProjectID:     gcpProject,
		ServiceName:      serviceName,
		TraceSampleRatio: sampleRatio,
		UserIDHashSalt:   strings.TrimSpace(os.Getenv("USER_ID_HASH_SALT")),
	}
}

func parsePositiveFloat(raw string) (float64, error) {
	var v float64
	_, err := fmt.Sscanf(raw, "%f", &v)
	return v, err
}

func gatewayBaseURL() string {
	if v := strings.TrimSpace(os.Getenv("GATEWAY_BASE_URL")); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv("JUMON_GATEWAY_BASE_URL"))
}

func gatewayInternalSecret() string {
	if v := strings.TrimSpace(os.Getenv("GATEWAY_INTERNAL_SECRET")); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv("JUMON_GATEWAY_INTERNAL_SECRET"))
}

func deriveConnectURL(gatewayURL string) string {
	if gatewayURL == "" {
		return "/connections"
	}
	return strings.TrimRight(gatewayURL, "/") + "/connections"
}

func envTruthy(key string) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	return value == "1" || value == "true" || value == "yes"
}
