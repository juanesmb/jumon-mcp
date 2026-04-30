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
- `PUBLIC_BASE_URL`
- `PORT` (default `8080`)
- `MCP_DEBUG_AUTH` (`true` to enable verbose auth logs)
- `GOOGLE_ADS_API_VERSION` (default `v22`)

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
