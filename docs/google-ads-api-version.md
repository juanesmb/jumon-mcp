# Google Ads API version policy

Jumon pins the Google Ads API version via environment variable. All gateway proxy paths use this version segment.

## Configuration

| Variable | Default | Location |
|----------|---------|----------|
| `GOOGLE_ADS_API_VERSION` | `v22` | jumon-mcp `.env`, Cloud Run / deployment env |

Set in:

- [`internal/config/config.go`](../internal/config/config.go) → `GatewayConfig.GoogleAPIVersion`
- Passed to [`internal/provider/googleads/tools.go`](../internal/provider/googleads/tools.go) via `wire.go`

## Upgrade checklist

When Google releases a new API version:

1. Read [Google Ads API release notes](https://developers.google.com/google-ads/api/docs/release-notes) for breaking GAQL or field changes affecting Jumon queries.
2. Bump `GOOGLE_ADS_API_VERSION` in deployment env and [`.env.example`](../.env.example).
3. Update field doc links in [`internal/provider/googleads/schema_docs.go`](../internal/provider/googleads/schema_docs.go) (`fields/v22` → new version).
4. Run unit tests: `go test ./internal/provider/googleads/...`
5. Run manual smoke tests: [google-ads-smoke-tests.md](google-ads-smoke-tests.md)
6. Re-sync `gaql_resources.txt` from [official google-ads-mcp](https://github.com/googleads/google-ads-mcp/blob/main/ads_mcp/gaql_resources.txt) if Google added resources.

## Do not

- Auto-bump the default in code without validating queries against the new version.
- Change the gateway contract — version is opaque to mcp-ads-manager (path is proxied as-is).

## Related

- [google-ads-tools.md](google-ads-tools.md)
- mcp-ads-manager gateway Google adapter (developer token + OAuth unchanged)
