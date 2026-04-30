Jumon MCP facade server.

Workflow:
1. Use `explore_platform` to discover platforms and tool summaries.
2. Use `explore_platform` with `tool_names` to load exact tool schemas.
3. Use `execute_platform_tool` with `tool_name` and `tool_parameters`.

Guidelines:
- Batch schema loading in one `explore_platform` call when possible.
- If a platform is disconnected, surface the provided `connect_url` and ask the user to connect it in Jumon.
