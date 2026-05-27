# Gateway contract

Internal API between **jumon-mcp** (this repo, caller) and **mcp-ads-manager** (OAuth vault + gateway).

Canonical full spec: sibling repo `mcp-ads-manager/docs/gateway-contract.md`.

## Auth

All requests:

```
x-gateway-secret: <GATEWAY_INTERNAL_SECRET>
```

Env: `GATEWAY_BASE_URL` or `JUMON_GATEWAY_BASE_URL` — base URL of the web app (no trailing slash).

## Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/internal/connections/{provider}/current?userId=` | Connection health (`connected`, `usable`, `health`) |
| POST | `/api/internal/providers/{provider}/proxy` | Proxy provider API call (proactive + reactive refresh) |
| POST | `/api/internal/providers/{provider}/refresh` | Refresh OAuth token |

Providers: `linkedin`, `google`, `reddit`.

## Connection health (essential)

```json
{
  "connected": true,
  "usable": false,
  "health": "needs_reconnect",
  "healthReason": "refresh_failed"
}
```

- **`usable`**: gate MCP tools and execution (not `connected` alone).
- **`health`**: `active` | `needs_reconnect` | `disconnected`.

## Proxy auth failure

```json
{
  "code": "TOKEN_REFRESH_FAILED",
  "reconnectRequired": true
}
```

## Implementation

- Client: `internal/infrastructure/gateway/client.go`
- `IsProviderUsable` / `RefreshSucceeded` / `IsTokenRefreshFailed`
- Registry connection check: `internal/provider/registry/connections.go` uses `usable`
- Provider wrappers: each provider's `proxy.go` maps auth failures to reconnect errors

**Never store or decrypt tokens in this repo.**
