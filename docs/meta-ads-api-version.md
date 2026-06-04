# Meta Marketing API version policy

Jumon pins the Meta Graph / Marketing API version via `META_GRAPH_API_VERSION` (default **`v25.0`**).

**Authoritative implementation:** mcp-ads-manager [`packages/providers/src/meta/version.ts`](https://github.com/jumonintelligence/mcp-ads-manager/blob/main/packages/providers/src/meta/version.ts) and [`meta-adapter.ts`](https://github.com/jumonintelligence/mcp-ads-manager/blob/main/packages/providers/src/meta-adapter.ts).

This repo documents the default in [`internal/provider/meta/paths.go`](../internal/provider/meta/paths.go) (`DefaultGraphAPIVersion`); all proxied calls use paths relative to the gateway (version applied in the vault).

## Upgrade

See mcp-ads-manager [docs/meta-ads-api-version.md](https://github.com/jumonintelligence/mcp-ads-manager/blob/main/docs/meta-ads-api-version.md) for the full checklist.

## Related

- [meta-ads-tools.md](meta-ads-tools.md)
- [docs/meta-ads-oauth.md](https://github.com/jumonintelligence/mcp-ads-manager/blob/main/docs/meta-ads-oauth.md) (mcp-ads-manager)
