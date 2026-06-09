# Local dev environment

One-time setup for hot-reload MCP development with Air. Merge this tooling to `main` before starting `feature/{platform}-mcp` branches.

## Prerequisites

```bash
go install github.com/air-verse/air@latest
```

If `air` is not on your `PATH`, `./scripts/dev.sh` falls back to `$(go env GOPATH)/bin/air` (default `~/go/bin/air`). Optionally add that directory to your shell profile:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

## Create `.env` (gitignored)

1. `cp .env.example .env`
2. Goland → Run → Edit Configurations → Environment variables
3. Copy all `KEY=value` entries into `.env` (one per line)
4. Verify: `git status` must **not** list `.env`

Optional: add `.env.local` for overrides (also gitignored). It is sourced after `.env`.

## Run the dev server

```bash
./scripts/dev.sh
```

This loads `.env`, then starts [Air](https://github.com/air-verse/air) on port `8080` (default). The server rebuilds on save under `internal/` and `cmd/`.

Connect Cursor `local-jumon-mcp` once per dev session to `http://localhost:8080/mcp` (or your configured MCP path).

## Without Air

```bash
set -a && source .env && set +a && go run ./cmd/jumon-mcp
```

## Agent safety

Agents must **never** create, read, or commit `.env` or `.env.local`. Reference `.env.example` keys only.
