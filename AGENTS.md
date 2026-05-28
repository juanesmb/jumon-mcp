# Jumon — jumon-mcp

Unified MCP facade for Jumon ad platforms. Exposes two tools to AI clients; routes execution to provider handlers.

## MCP surface

| Tool | Purpose |
|------|---------|
| `explore_platform` | Discover platforms, list tools, load JSON schemas |
| `execute_platform_tool` | Run a named provider tool with validated parameters |

Workflow: explore → load schemas → execute. See `internal/app/instructions/server_instructions.md`.

## Architecture map

| Path | Role |
|------|------|
| `cmd/jumon-mcp` | Entry point |
| `internal/app` | HTTP server, wiring, OAuth metadata |
| `internal/transport/mcp` | Facade tool handlers |
| `internal/usecase/catalog` | `explore_platform` logic |
| `internal/usecase/execution` | `execute_platform_tool` logic |
| `internal/provider/registry` | Tool registration + connection checks |
| `internal/provider/linkedin` | LinkedIn tools (richest: analytics, creatives, lead gen) |
| `internal/provider/googleads` | Google Ads GAQL tools |
| `internal/provider/reddit` | Reddit Ads tools |
| `internal/infrastructure/gateway` | HTTP client → mcp-ads-manager internal API |

## Key invariants

- Clerk JWT on every MCP request; user ID extracted from token context.
- **No token storage in Go.** All provider API calls go through the gateway in mcp-ads-manager.
- MCP tool gating uses gateway **`usable`**, not `connected` alone (`IsProviderUsable`).
- Gateway proxy refreshes tokens proactively and on 401; jumon-mcp `ProxyProviderOrRefresh` retries only when refresh returns `refreshed: true`.
- Auth failures surface as `platform_not_connected` or `TOKEN_REFRESH_FAILED` with `connect_url`.
- Tool names: `{platform}_{action}` (e.g. `linkedin_get_campaigns`).

## Gateway contract

See [docs/gateway-contract.md](docs/gateway-contract.md). OAuth + token decryption live in **mcp-ads-manager**. Production OAuth monitoring: mcp-ads-manager [docs/oauth-production-monitoring.md](https://github.com/jumonintelligence/mcp-ads-manager/blob/main/docs/oauth-production-monitoring.md).

## Google Ads docs

- [docs/google-ads-tools.md](docs/google-ads-tools.md) — capability matrix (14 tools)
- [docs/google-ads-smoke-tests.md](docs/google-ads-smoke-tests.md) — manual regression checklist
- [docs/google-ads-api-version.md](docs/google-ads-api-version.md) — API version upgrade policy

## Start here

| Task | Entry files |
|------|-------------|
| New LinkedIn tool | `internal/provider/linkedin/tools.go`, `service.go`, `schema_docs.go`, `*_test.go` |
| New Google tool | `internal/provider/googleads/tools.go`, `accounts.go` / `reports.go` / `field_service.go` / `generic_search.go`, `schemas.go`, `schema_docs.go`, `*_test.go` — see [docs/google-ads-tools.md](docs/google-ads-tools.md) |
| New Reddit tool | `internal/provider/reddit/tools.go`, `service.go` |
| Analytics work | `docs/linkedin-analytics-roadmap.md`, `analytics_pagination.go` |
| Gateway calls | `internal/infrastructure/gateway/client.go`, provider `proxy.go` |
| Tool registration | `internal/app/wire.go`, `internal/provider/registry/registry.go` |

Reference templates: LinkedIn `linkedin_get_campaign_groups`; Google curated `google_search_keywords`; Google P2 `google_search_gaql` + `google_get_resource_metadata`.

## Skills (invoke with @)

| Skill | When |
|-------|------|
| `add-mcp-tool` | New `execute_platform_tool` target for any provider |

After adding a tool, sync the UI blurb in mcp-ads-manager `apps/web/lib/mcp-provider-tools.ts`.

## Feature prompt template

```text
Feature: [one line]
Repo(s): [mcp-ads-manager | jumon-mcp | both]
Similar to: [existing tool or file]
Constraints: [e.g. gateway unchanged, LinkedIn only]
Skill: @add-mcp-tool (optional)
```
