---
name: add-mcp-tool
description: Add a new execute_platform_tool target in jumon-mcp for LinkedIn, Google Ads, or Reddit. Use when registering a provider tool in tools.go with service logic, schema, tests, and optional UI catalog sync.
disable-model-invocation: true
---

# Add MCP tool

Add a new provider tool callable via `execute_platform_tool`.

## Reference template

Copy patterns from `linkedin_get_campaign_groups` in `internal/provider/linkedin/`.

## Checklist

1. **Constant + registration** — In `internal/provider/{provider}/tools.go`:
   - Add tool name constant (`{platform}_{action}`)
   - Append `registry.ToolDefinition` with Summary, Description, InputSchema, `RequiresConnection: true`
   - Wire `Execute` to parse input and call service

2. **Service** — In `service.go`:
   - Define input struct and parser
   - Implement method; build API path/query/body
   - Call gateway via provider `proxy.go` (never hold tokens in Go)

3. **Schema** — Add JSON Schema helper for input parameters. For LinkedIn analytics/reporting tools, update `schema_docs.go` with agent-facing field/pivot guidance.

4. **Tests** — Add `*_test.go` for input parsing, pagination, error cases. Run:
   ```bash
   go test ./internal/provider/{provider}/...
   ```

5. **LinkedIn analytics** — If analytics-related, check [docs/linkedin-analytics-roadmap.md](../../docs/linkedin-analytics-roadmap.md) and [reference.md](reference.md).

6. **Google Ads** — See [reference-google.md](reference-google.md) and [docs/google-ads-tools.md](../../docs/google-ads-tools.md).

7. **UI sync** — Remind to add blurb in mcp-ads-manager `apps/web/lib/mcp-provider-tools.ts` (use `@sync-mcp-tool-catalog` in that repo).

## Tool naming

- Prefix: `linkedin_`, `google_`, or `reddit_`
- Snake case action: `get_campaign_groups`, `list_ad_accounts`

## Gateway

All API calls use `internal/infrastructure/gateway/client.go` → mcp-ads-manager proxy. See [docs/gateway-contract.md](../../docs/gateway-contract.md).

## Do not

- Expose new top-level MCP tools — only register in the internal registry.
- Call provider APIs directly from Go.
