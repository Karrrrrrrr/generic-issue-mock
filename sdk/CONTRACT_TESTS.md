# Local SDK protocol tests

Scope: every exported channel operation in `slash`, `photonpay`, and `payndapay`,
including their integration interfaces. UQPay is intentionally excluded.
Low-level request transports are exercised through the public operations.
After the removal of 22 unused Slash methods, the explicit cases cover 25 Slash,
36 PhotonPay, and 43 Paynda operations (104 in total, counting interface wrappers).

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
