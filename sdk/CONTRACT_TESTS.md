# Local SDK protocol tests

Scope: every exported channel operation in `slash`, `photonpay`, and `payndapay`,
including their integration interfaces. UQPay is intentionally excluded.
Low-level request transports are exercised through the public operations.

## Run

The tests require PostgreSQL and the mock server. From the repository root:

```sh
cd backend
export GOCACHE=/home/kar/.cache/go-build
export GOTMPDIR=/home/kar/.cache/go-tmp
mkdir -p "$GOCACHE" "$GOTMPDIR"
HTTP_ADDR=127.0.0.1:18000 go run .
```

In another terminal, from the repository root:

```sh
bash sdk/test-contract.sh -v
```

Overrides:

```sh
MOCK_BASE_URL=http://127.0.0.1:8000 \
MARXO_ROOT=/path/to/marxo \
bash sdk/test-contract.sh -v
```

The server accepts `DATABASE_DSN`; its default is the local `generic_mock`
database configured in `backend/data/postgres.go`. No database reset is needed.
Each suite creates a uniquely named account, funds it through the UI API, and
creates its own cardholders, cards, and transactions. Fixtures are retained for
inspection. Repeated runs do not depend on IDs or data from earlier runs.
Only loopback server URLs are accepted; old remote credentials and remote test
endpoints are no longer used. Do not point the local server at a production DB.

The SDK source imports `tman/errors` and `tman/enums`, supplied by the existing
Marxo checkout. Also, its source uses the Resty beta API, while `sdk/go.mod`
declares an incompatible RC release. The runner creates a temporary Go workspace
that connects the real Marxo module and selects `resty.dev/v3 v3.0.0-beta.3`, the
version declared by Marxo. It removes the workspace on exit. Neither SDK source,
SDK DTOs, SDK module files, nor Marxo files are modified by this workaround.
Consequently, use this runner rather than an unconfigured standalone `go test`.

Some PhotonPay SDK list validators reject non-nil optional filters (for example,
`CardType` and `CardFormFactor`) before sending any HTTP request. Invocation tests
omit those filters; offline regression cases record that existing SDK behavior.
This is an SDK-side limitation, not a server protocol fix, and the SDK is unchanged.

## Assertions

- Invoke the actual SDK methods, rather than replacing them with HTTP stubs.
- Record real upstream HTTP responses through a local reverse proxy, without
  modifying request/response payloads or replacing SDK decoding.
- Check HTTP success, JSON content type, channel envelope codes, and non-null data.
- Recursively check returned field types against the original SDK response types:
  string versus number, integer versus fraction, booleans, arrays, objects, and
  `json:",string"` numeric strings. Optional SDK fields can remain absent; core
  identifiers, timestamps, amounts, and pagination fields have explicit assertions.
- Validate Slash canonical UUID IDs and RFC3339 timestamps; Paynda decimal IDs
  and `2006-01-02 15:04:05` timestamps; PhotonPay decimal IDs and
  `2006-01-02T15:04:05` timestamps. Validate card expiry and CVV formats as well.
- Exercise empty and populated card lists, mutations, funding operations,
  transactions, uploads, and SDK interface wrappers.
- Decode Paynda's JSON-string request-result envelope and validate its embedded
  card/transfer DTO. Check PhotonPay's missing-request code `VCC1039` through the
  SDK predicate, so idempotency probes preserve downstream control flow.
- Reflect over SDK public methods to fail if a method has no invocation case.
- Run negative unit cases against the checker itself, including number/string
  swaps, null arrays, integer overflow, and invalid timestamps. Preserve the
  existing offline signature, disabled-BIN, and nil/empty-input checks.

## Protocol-only behavior

These tests target the wire contract, not complete downstream business semantics.
The existing implemented account/card/transaction flows still use the business
and repository layers. Additional SDK-only capabilities intentionally have
limited semantics and do not introduce channel-specific persistence:

- Slash group, legal-entity, and merchant directory entries are account-derived
  protocol views. Spending controls, modifiers, and fee reporting are placeholders.
  OpenAPI webhook/authorization-webhook DTOs do not change the UI-managed callback
  configuration.
- PhotonPay quote/recharge, funding-history, billing-address, and SDK subscription
  endpoints provide protocol responses, not a complete funding/subscription engine.
- Paynda cardholder wallet DTOs are views of the owning account wallet. Card controls
  are not persisted. Account/cardholder delete endpoints acknowledge valid resources
  without deleting the underlying business aggregates.

Unsupported input fields are marked `Invalid:` at the service boundary. The tests
must not be interpreted as business-correctness or third-party certification tests.
