#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "$0")" && pwd)"
export GOCACHE="${GOCACHE:-/home/kar/.cache/go-build}"
export GOTMPDIR="${GOTMPDIR:-/home/kar/.cache/go-tmp}"
mkdir -p "$GOCACHE" "$GOTMPDIR"
marxo="${MARXO_ROOT:-/home/kar/workspace/ptm/marxo}"
if [[ ! -f "$marxo/go.mod" ]]; then
  echo 'Set MARXO_ROOT to the existing Marxo (module tman) checkout.' >&2
  exit 1
fi
if [[ -n "${MOCK_BASE_URL:-}" ]]; then
  echo 'This runner starts its own isolated server. Unset MOCK_BASE_URL; use TEST_HTTP_PORT instead.' >&2
  exit 1
fi
test_http_port="${TEST_HTTP_PORT:-18000}"
if [[ ! "$test_http_port" =~ ^[1-9][0-9]{0,4}$ ]] || (( test_http_port > 65535 )); then
  echo 'TEST_HTTP_PORT must be a port between 1 and 65535.' >&2
  exit 1
fi
if ! command -v curl >/dev/null; then
  echo 'curl is required for the isolated server readiness check.' >&2
  exit 1
fi
if (echo >/dev/tcp/127.0.0.1/"$test_http_port") 2>/dev/null; then
  echo "Port $test_http_port is already in use; choose a free TEST_HTTP_PORT." >&2
  exit 1
fi

workspace="$(mktemp -d "$GOTMPDIR/sdk-contract.XXXXXX")"
test_log_dir="$(mktemp -d "$GOTMPDIR/sdk-contract-logs.XXXXXX")"
server_pid=''
cleanup() {
  local status=$?
  trap - EXIT
  if [[ -n "$server_pid" ]]; then
    kill "$server_pid" 2>/dev/null || true
    wait "$server_pid" 2>/dev/null || true
  fi
  rm -rf "$workspace"
  echo "[test] Logs retained in $test_log_dir; database generic_mock_sdk_test retained for inspection."
  exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

export GOWORK="$workspace/go.work"
cd "$workspace"
go work init "$root" "$marxo"
(
  cd "$root/../backend"
  GOWORK=off go build -o "$workspace/mock-server" .
)
source "$root/init-test-db.sh"

export HTTP_ADDR="127.0.0.1:$test_http_port"
export MOCK_BASE_URL="http://$HTTP_ADDR"
echo "[test] Starting isolated server at $MOCK_BASE_URL using generic_mock_sdk_test."
"$workspace/mock-server" >"$test_log_dir/server.log" 2>&1 &
server_pid=$!
server_ready=false
for attempt in {1..100}; do
  if ! kill -0 "$server_pid" 2>/dev/null; then
    cat "$test_log_dir/server.log" >&2
    echo 'Isolated server exited before becoming ready.' >&2
    exit 1
  fi
  if curl --silent --fail --max-time 1 "$MOCK_BASE_URL/slash/ui/accounts" >/dev/null; then
    if kill -0 "$server_pid" 2>/dev/null; then
      server_ready=true
      break
    fi
  fi
  sleep 0.2
done
if [[ "$server_ready" != true ]]; then
  cat "$test_log_dir/server.log" >&2
  echo 'Timed out waiting for the isolated server.' >&2
  exit 1
fi

cd "$root"
go test -v -count=1 -timeout=5m ./internal/contract ./slash ./photonpay ./payndapay "$@" 2>&1 | tee "$test_log_dir/tests.log"
