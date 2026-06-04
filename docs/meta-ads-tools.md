# Meta Ads MCP tools

**Status:** P0 foundation only — OAuth + gateway. **P1** will register read tools (`meta_list_ad_accounts`, insights, structure).

## P1 capability matrix (planned)

| User need | Tool (planned) | Graph API |
|-----------|----------------|-----------|
| List ad accounts | `meta_list_ad_accounts` | `GET /me?fields=adaccounts{...}` |
| Account details | `meta_get_ad_account` | `GET /{act_id}` |
| List campaigns | `meta_list_campaigns` | `GET /{act_id}/campaigns` |
| Campaign insights | `meta_get_campaign_insights` | `GET /{campaign_id}/insights` |
| Account insights | `meta_get_ad_account_insights` | `GET /{act_id}/insights` |

Insights support `publisher_platform` breakdown for Facebook vs Instagram.

## Workflow (after P1)

1. Connect Meta in Jumon dashboard.
2. `meta_list_ad_accounts` → pick `act_id`.
3. Structure: campaigns → ad sets → ads.
4. Reporting: account or campaign insights with `date_preset` or `time_range`.

## API version

[v25.0](meta-ads-api-version.md) — ODAX `OUTCOME_*` objectives for future writes; ASC/AAC write APIs deprecated.

## Related

- mcp-ads-manager [meta-ads-oauth.md](https://github.com/jumonintelligence/mcp-ads-manager/blob/main/docs/meta-ads-oauth.md)
- [meta-ads-smoke-tests.md](https://github.com/jumonintelligence/mcp-ads-manager/blob/main/docs/meta-ads-smoke-tests.md) (P0)
