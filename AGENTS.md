# AGENTS.md

## Cursor Cloud specific instructions

This is a pure Go project (single module, single binary). No databases, Docker, or external build tools are needed for local development.

### Quick reference

| Action | Command |
|--------|---------|
| Install deps | `go mod download` |
| Run tests | `go test ./...` |
| Lint | `go vet ./...` |
| Build | `go build ./cmd/jumon-mcp` |
| Run locally | `go run ./cmd/jumon-mcp` |

### Running the server locally

The server requires these env vars to start (set to placeholder values for local dev if you don't have real credentials):

```bash
CLERK_JWKS_URL="https://clerk.example.com/.well-known/jwks.json" \
CLERK_ISSUER="https://clerk.example.com" \
GATEWAY_BASE_URL="http://localhost:9999" \
GATEWAY_INTERNAL_SECRET="dev-secret" \
OBSERVABILITY_ENABLED=false \
go run ./cmd/jumon-mcp
```

The server listens on `PORT` (default `8080`) at path `/mcp`. With placeholder credentials the server starts and accepts HTTP requests, but MCP tool calls will fail JWT verification (401) since no real Clerk JWKS is configured.

### Key endpoints to verify the server is running

- `GET /.well-known/oauth-protected-resource` — returns OAuth metadata JSON.
- `POST /mcp` — MCP Streamable HTTP endpoint (requires `Authorization: Bearer <jwt>`).
- `GET /favicon.png` — returns the embedded icon.

### Tests

Unit tests live alongside provider code in `internal/provider/linkedin/` and `internal/provider/reddit/`. They run entirely offline with no external dependencies.
