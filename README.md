# Jumon MCP Facade

Unified MCP server for Jumon advertising providers. This server exposes a small
facade surface (`explore_platform`, `execute_platform_tool`) and routes
execution to provider-specific handlers (LinkedIn and Google Ads).

## Environment variables

Required:

- `CLERK_JWKS_URL`
- `CLERK_ISSUER`
- `GATEWAY_BASE_URL` (or `JUMON_GATEWAY_BASE_URL`)
- `GATEWAY_INTERNAL_SECRET` (or `JUMON_GATEWAY_INTERNAL_SECRET`)

Optional:

- `CLERK_AUDIENCE`
- `MCP_REQUIRED_SCOPE`
- `AUTHORIZATION_SERVER_URL` (defaults to `CLERK_ISSUER`)
- `PUBLIC_BASE_URL` — **canonical public origin for this MCP deployment, no trailing slash and no `/mcp` path.**  
  MCP clients compare this to the connector URL OAuth metadata (`resource`). If you use a custom domain (e.g. `https://mcp.example.com`), set `PUBLIC_BASE_URL=https://mcp.example.com` on Cloud Run—even though the container still receives requests on `*.run.app`. If unset, the server uses `X-Forwarded-*` headers or `Host` (which may still be `*.run.app` and trigger “protected resource does not match” in Cursor).
- `PORT` (default `8080`)
- `MCP_DEBUG_AUTH` (`true` to enable verbose auth logs)
- `GOOGLE_ADS_API_VERSION` (default `v22`)
- `USER_ID_HASH_SALT` — optional salt combined with hashed Clerk identifiers for correlated `user.hash` fields without logging raw JWT subjects.

GCP / OpenTelemetry knobs (omit or disable when developing without Cloud exporters):

- `GOOGLE_CLOUD_PROJECT` / `GCP_PROJECT` — used for stdout log fields such as `logging.googleapis.com/trace` so entries join Cloud Trace.
- `OBSERVABILITY_ENABLED` — set to `false` to silence Cloud Trace + Cloud Monitoring export while retaining structured slog JSON logs.
- `OTEL_SERVICE_NAME` — resource `service.name` (default `jumon-mcp`).
- `OBSERVABILITY_TRACE_SAMPLE_RATIO` — `0..1` ratio for probabilistic sampling of new trace IDs.

## Google Cloud IAM (production)

Cloud Run workloads should attach a service principal with **`roles/cloudtrace.agent`** plus **`roles/monitoring.metricWriter`** (or narrower equivalents that still allow exporting spans/custom metrics).

## Verification (observability)

After traffic, validate **Structured JSON logs** in Cloud Logging (`event=http_request`, `event=http_upstream`, `severity`). Open **Cloud Trace** to confirm spans: outer HTTP span wrapping `mcp.tool.explore` / nested `mcp.tool.execute`. In **Metrics Explorer**, allow a couple of minutes for OTel-derived series (names include the `mcp_*` prefixes emitted via the Google metric exporter).

## Run locally

```bash
go run ./cmd/jumon-mcp
```

## Verification

```bash
go test ./...
```

## Rollout guidance

1. Deploy this server as the recommended MCP endpoint.
2. Keep `linkedin-mcp` and `google-ads-mcp` running temporarily for legacy users.
3. Update onboarding to point new users to only one connector URL.
4. Deprecate the provider-specific MCP endpoints after migration.
