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

## Org context propagation (per-org MCP URL)

Users configure per-org MCP server URLs in their AI agent:

```
https://mcp.jumonintelligence.com/mcp?org=org_aaa   # Agency A
https://mcp.jumonintelligence.com/mcp?org=org_bbb   # Agency B
```

Flow:
1. `RequireBearerAuth` verifies the JWT then reads `r.URL.Query().Get("org")`.
2. URL `?org=` takes precedence over the JWT `org_id` claim (JWT claim is unreliable for AI agent OAuth flows).
3. `OrgIDFromContext(ctx)` returns the resolved org ID.
4. `gateway.Client.ProxyProvider` / `RefreshProvider` include `"orgId"` in the JSON payload.
5. `mcp-ads-manager` proxy route validates `userId` is a member of `orgId` (via `org_memberships` table).
6. Connection lookup uses `(userId, provider, orgId)` — org-scoped OAuth connections.

OAuth connections are now keyed by `(clerk_user_id, provider, clerk_org_id)` allowing different ad accounts per org.

## Implementation

- Client: `internal/infrastructure/gateway/client.go`
- Auth claims: `internal/infrastructure/security/clerk_token_verifier.go` — `AuthClaims{UserID, OrgID}`
- Middleware: `internal/infrastructure/middleware/auth_middleware.go` — `OrgIDFromContext`
- `IsProviderUsable` / `RefreshSucceeded` / `IsTokenRefreshFailed`
- Registry connection check: `internal/provider/registry/connections.go` uses `usable`
- Provider wrappers: each provider's `proxy.go` maps auth failures to reconnect errors

**Never store or decrypt tokens in this repo.**
