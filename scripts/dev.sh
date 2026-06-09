#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
# shellcheck source=scripts/load-env.sh
source "$(dirname "$0")/load-env.sh"
load_jumon_env "."

resolve_air() {
  if command -v air >/dev/null 2>&1; then
    command -v air
    return 0
  fi
  local gobin
  gobin="$(go env GOBIN 2>/dev/null || true)"
  if [[ -n "${gobin}" && -x "${gobin}/air" ]]; then
    echo "${gobin}/air"
    return 0
  fi
  local gopath_bin
  gopath_bin="$(go env GOPATH)/bin/air"
  if [[ -x "${gopath_bin}" ]]; then
    echo "${gopath_bin}"
    return 0
  fi
  echo "error: air not found. Run: go install github.com/air-verse/air@latest" >&2
  echo "       Then add \$(go env GOPATH)/bin to your PATH, or re-run ./scripts/dev.sh (it checks GOPATH/bin)." >&2
  return 1
}

exec "$(resolve_air)"
