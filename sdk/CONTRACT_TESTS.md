# Local SDK protocol tests

Scope: every exported channel operation in `slash`, `photonpay`, and `payndapay`,
including their integration interfaces. UQPay is intentionally excluded.
Low-level request transports are exercised through the public operations.
After the removal of 22 unused Slash methods, the explicit cases cover 25 Slash,
36 PhotonPay, and 43 Paynda operations (104 in total, counting interface wrappers).

## Run

The tests require local PostgreSQL, Go, `psql`, `curl`, and `flock`. Run from the
repository root; do not start a server manually:

```sh
bash sdk/test-contract.sh
```

Every invocation of the runner, including filtered runs, performs these steps:

1. Build the mock backend and call `sdk/init-test-db.sh`.
2. Recreate the dedicated **`generic_mock_sdk_test`** database, migrate the schema,
   and seed the initial channel accounts, card products, cards, and configuration.
3. Start a managed backend on `127.0.0.1:18000`, explicitly connected to that DB.
4. Run the Slash, PhotonPay, and Paynda SDK tests against that backend.
5. Stop the backend on success, failure, or interruption. Retain the test database
   and log directory for inspection; reset the test database on the next run.

To initialize the test database without starting the server or running tests:

```sh
bash sdk/init-test-db.sh
```

**Initialization deletes the previous contents of `generic_mock_sdk_test`.** It
never resets `generic_mock` and does not accept an arbitrary database name or the
application's inherited `DATABASE_DSN`. The database must be local. An existing
database without the script's identifying comment is rejected, not dropped.
Active connections are not forcibly terminated: stop manual test-db clients
before reinitializing. A shared file lock prevents concurrent runner/init jobs.

The standalone initializer uses `backend/cmd/init`, which requires an explicit
`DATABASE_DSN` and reuses the backend's migration and seed functions. No duplicate
schema SQL is maintained in the shell script.

The runner enables verbose output by default. `[SDK]` logs show actual SDK HTTP
calls, while `[UI setup]` logs show fixture creation, funding, and transaction
simulation. Each request includes its method, URL/query, and body; each response
includes its status, content type, and body. Request/response streams are restored
after logging so the SDK still receives the original payload. Request headers
are not dumped. Bodies contain local mock card details; do not publish these logs
with real credentials or data. When running tests directly, use `go test -v` to
see logs from passing cases.

Overrides (all optional):

| Variable | Default | Purpose |
| --- | --- | --- |
| `TEST_PGHOST` | `127.0.0.1` | Local PostgreSQL host; loopback only |
| `TEST_PGPORT` | `5432` | PostgreSQL port |
| `TEST_PGUSER` | `postgres` | Local role with database creation/deletion privileges |
| `TEST_PGPASSWORD` | `root` | Local database password; may be explicitly empty |
| `TEST_HTTP_PORT` | `18000` | Dedicated HTTP port; an occupied port is rejected |
| `MARXO_ROOT` | `/home/kar/workspace/ptm/marxo` | Existing Marxo checkout |

For example:

```sh
TEST_HTTP_PORT=18001 \
TEST_PGPORT=5432 \
MARXO_ROOT=/path/to/marxo \
bash sdk/test-contract.sh -run TestSlashCreateCard
```

The runner rejects a supplied `MOCK_BASE_URL` rather than connecting to an
unverified server/database. It exports its own URL to the SDK tests. The test
helper's direct-run fallback is also `http://127.0.0.1:18000`, not the normal
development server on port 8000. The ordinary backend's database default remains
unchanged. Each suite still creates isolated fixture accounts, but those accounts
now accumulate only in the dedicated test database. The runner prints the log
directory containing `server.log` and `tests.log` when it exits.

The SDK source imports `tman/enums`, supplied by the existing Marxo checkout.
The runner creates a temporary Go workspace that connects the real Marxo module
and removes it on exit. Resty is declared directly as `resty.dev/v3 v3.0.0-beta.3`
in `sdk/go.mod`; the runner does not override its version. Use this runner for
the three-channel tests while the PhotonPay interface depends on `tman/enums`.

Some PhotonPay SDK list validators reject non-nil optional filters (for example,
`CardType`, `CardFormFactor`, and trade `TransactionType`) before sending any HTTP request. Live tests
omit those filters; offline regression cases record that existing SDK behavior.
This is an SDK-side limitation, not a server protocol fix, and the SDK is unchanged.

## Assertions

- Call each actual SDK method directly in its test case. Assertions use explicit
  `if` checks of errors and named response fields. There is no reflection,
  dynamic method dispatch, generic invocation table, or schema-checking framework.
- Shared helpers only configure the local client, prepare fixtures, and capture
  HTTP exchanges; they do not invoke methods selected by name or hide assertions.
- Record real upstream HTTP responses through a local reverse proxy, without
  modifying request/response payloads or replacing SDK decoding.
- Decode through the original SDK response DTOs on every call. Check identifiers,
  timestamps, amounts, statuses, empty arrays, and pagination fields explicitly.
  Card and transaction cases additionally decode captured response bodies into
  explicit wire DTOs and check HTTP status, content type, and success envelopes.
  In particular, Paynda transaction string IDs are checked on the wire because
  the SDK's `json.Number` also accepts JSON numbers.
- Validate Slash canonical UUID IDs and RFC3339 timestamps; Paynda decimal IDs
  and `2006-01-02 15:04:05` timestamps; PhotonPay decimal IDs and
  `2006-01-02T15:04:05` timestamps. Validate card expiry and CVV formats as well.
- Exercise card creation, empty/populated/paged queries, freeze/unfreeze/close,
  sensitive details, funding, request-result queries, account isolation, malformed
  IDs, missing parameters, and SDK interface wrappers.
- Exercise authorization, over-clearing, independent/linked refunds, reversal,
  transaction detail/list queries, date filtering, pagination, and invalid inputs.
- Decode Paynda's JSON-string request-result envelope and validate its embedded
  card/transfer DTO. Check PhotonPay's missing-request code `VCC1039` through the
  SDK predicate, so idempotency probes preserve downstream control flow.
- Preserve offline signature, disabled-BIN, and nil/empty-input checks using
  direct calls. Adding an SDK method requires adding its explicit test case;
  there is intentionally no reflection-based coverage discovery.

## Protocol-only behavior

These tests target the wire contract, not complete downstream business semantics.
The existing implemented account/card/transaction flows still use the business
and repository layers. Additional SDK-only capabilities intentionally have
limited semantics and do not introduce channel-specific persistence:

- Slash legal-entity entries are account-derived protocol views. Cases for deleted
  group, merchant, webhook, utilization, and spending-control SDK calls are removed.
- PhotonPay quote/recharge, funding-history, billing-address, and SDK subscription
  endpoints provide protocol responses, not a complete funding/subscription engine.
- Paynda cardholder wallet DTOs are views of the owning account wallet. Card controls
  are not persisted. Account/cardholder delete endpoints acknowledge valid resources
  without deleting the underlying business aggregates.

Pagination checks cover field types, page selection, and empty-array encoding,
not complete business-level total-count semantics. PhotonPay sandbox requests use
the explicit origin ID `"0"` for an independent transaction; linked refunds and
reversals use the authorization transaction's decimal ID. An explicitly supplied
empty origin is invalid, while an omitted optional origin is represented by nil
in the server DTO.

Unsupported input fields are marked `Invalid:` at the service boundary. The tests
must not be interpreted as business-correctness or third-party certification tests.
