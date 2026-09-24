#!/usr/bin/env bash
set -euo pipefail

test_script_root="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export GOCACHE="${GOCACHE:-/home/kar/.cache/go-build}"
export GOTMPDIR="${GOTMPDIR:-/home/kar/.cache/go-tmp}"
mkdir -p "$GOCACHE" "$GOTMPDIR"

export PGHOST="${TEST_PGHOST:-127.0.0.1}"
export PGPORT="${TEST_PGPORT:-5432}"
export PGUSER="${TEST_PGUSER:-postgres}"
export PGPASSWORD="${TEST_PGPASSWORD-root}"
export PGSSLMODE=disable
export PGCONNECT_TIMEOUT=5
export PGDATABASE=generic_mock_sdk_test
export DATABASE_DSN='dbname=generic_mock_sdk_test'
unset PGSERVICE PGSERVICEFILE PGHOSTADDR PGOPTIONS

case "$PGHOST" in
  127.0.0.1|localhost|::1) ;;
  *)
    echo 'Test database initialization only permits loopback PostgreSQL hosts.' >&2
    exit 1
    ;;
esac
if [[ ! "$PGPORT" =~ ^[1-9][0-9]{0,4}$ ]] || (( PGPORT > 65535 )); then
  echo 'TEST_PGPORT must be a port between 1 and 65535.' >&2
  exit 1
fi
for required_command in psql flock go; do
  if ! command -v "$required_command" >/dev/null; then
    echo "Required command is unavailable: $required_command" >&2
    exit 1
  fi
done

exec 9>"$GOTMPDIR/generic-mock-sdk-test.lock"
if ! flock -n 9; then
  echo 'Another SDK test run or database initialization is active.' >&2
  exit 1
fi

test_database_marker="$(psql -X -w -v ON_ERROR_STOP=1 -d postgres -Atqc \
  "SELECT COALESCE(shobj_description(oid, 'pg_database'), 'unmarked') FROM pg_database WHERE datname = 'generic_mock_sdk_test';")"
if [[ -n "$test_database_marker" && "$test_database_marker" != 'generic-mock SDK disposable test database' ]]; then
  echo 'Refusing to reset an existing generic_mock_sdk_test database without the SDK test marker.' >&2
  exit 1
fi

echo "[test-db] Recreating generic_mock_sdk_test on $PGHOST:$PGPORT; generic_mock is untouched."
psql -X -w -v ON_ERROR_STOP=1 -d postgres <<'SQL'
DROP DATABASE IF EXISTS generic_mock_sdk_test;
CREATE DATABASE generic_mock_sdk_test TEMPLATE template0;
COMMENT ON DATABASE generic_mock_sdk_test IS 'generic-mock SDK disposable test database';
SQL

(
  cd "$test_script_root/../backend"
  GOWORK=off go run ./cmd/init
)
echo '[test-db] Schema and seed data initialized in generic_mock_sdk_test.'
