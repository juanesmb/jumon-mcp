#!/usr/bin/env bash
# shellcheck shell=bash
# Loads gitignored env files for local dev. Never commit .env.

load_jumon_env() {
  local repo_root="${1:-.}"
  local loaded=0

  if [[ -n "${ENV_FILE:-}" && -f "${ENV_FILE}" ]]; then
    # shellcheck disable=SC1090
    set -a && source "${ENV_FILE}" && set +a
    echo "loaded env from ${ENV_FILE}"
    loaded=1
  elif [[ -f "${repo_root}/.env" ]]; then
    # shellcheck disable=SC1091
    set -a && source "${repo_root}/.env" && set +a
    echo "loaded env from ${repo_root}/.env"
    loaded=1
  fi

  if [[ -f "${repo_root}/.env.local" ]]; then
    # shellcheck disable=SC1091
    set -a && source "${repo_root}/.env.local" && set +a
    echo "loaded env from ${repo_root}/.env.local"
    loaded=1
  fi

  if [[ "${loaded}" -eq 0 ]]; then
    echo "error: no env file found. Create ${repo_root}/.env from .env.example (see docs/local-dev-env.md)" >&2
    return 1
  fi

  local missing=()
  [[ -z "${CLERK_JWKS_URL:-}" ]] && missing+=("CLERK_JWKS_URL")
  if [[ -z "${GATEWAY_BASE_URL:-}" && -z "${JUMON_GATEWAY_BASE_URL:-}" ]]; then
    missing+=("GATEWAY_BASE_URL")
  fi
  if [[ -z "${GATEWAY_INTERNAL_SECRET:-}" && -z "${JUMON_GATEWAY_INTERNAL_SECRET:-}" ]]; then
    missing+=("GATEWAY_INTERNAL_SECRET")
  fi
  [[ -z "${CLERK_ISSUER:-}" ]] && missing+=("CLERK_ISSUER")

  if ((${#missing[@]} > 0)); then
    echo "error: missing required env vars after load: ${missing[*]}" >&2
    return 1
  fi
}
