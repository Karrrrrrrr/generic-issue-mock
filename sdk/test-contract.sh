#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "$0")" && pwd)"
export GOCACHE="${GOCACHE:-/home/kar/.cache/go-build}"
export GOTMPDIR="${GOTMPDIR:-/home/kar/.cache/go-tmp}"
mkdir -p "$GOCACHE" "$GOTMPDIR"
workspace="$(mktemp -d "$GOTMPDIR/sdk-contract.XXXXXX")"
trap 'rm -rf "$workspace"' EXIT
marxo="${MARXO_ROOT:-/home/kar/workspace/ptm/marxo}"
if [[ ! -f "$marxo/go.mod" ]]; then
  echo 'Set MARXO_ROOT to the existing Marxo (module tman) checkout.' >&2
  exit 1
fi
export GOWORK="$workspace/go.work"
cd "$workspace"
go work init "$root" "$marxo"
go work edit -replace=resty.dev/v3=resty.dev/v3@v3.0.0-beta.3
cd "$root"
go test -count=1 -timeout=5m ./internal/contract ./slash ./photonpay ./payndapay "$@"
