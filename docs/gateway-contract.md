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
| GET | `/api/internal/connections/{provider}/current?userId=` | Connection status |
| POST | `/api/internal/providers/{provider}/proxy` | Proxy provider API call |
| POST | `/api/internal/providers/{provider}/refresh` | Refresh OAuth token |

Providers: `linkedin`, `google`, `reddit`.

## Proxy body (essential)

```json
{
  "userId": "user_abc",
  "mcpTool": "linkedin_get_campaigns",
  "method": "GET",
  "path": "adAccounts/123/adCampaigns",
  "query": {},
  "headers": {}
}
```

## Implementation

- Client: `internal/infrastructure/gateway/client.go`
- Auto-refresh on 401: `ProxyProviderOrRefresh`
- Provider wrappers: each provider's `proxy.go`

**Never store or decrypt tokens in this repo.**
