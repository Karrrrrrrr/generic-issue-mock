# Generic Mock Engineering Rules

## Model Ownership

- This repository simulates third-party channels for downstream systems. In this repository, a channel is the service being simulated; never carry downstream names such as `ThirdPartyID` into a generic model.
- Generic models use an auto-incrementing `int64` primary key `ID` for relations and transactions. A channel service derives every external resource ID from that key with its channel formatter; do not persist `display_id` or add another resource token.
- The channel formatter follows the downstream DTO field type. For a downstream string ID, format the internal `int64` as its decimal string and parse it back with validation. For a downstream numeric ID, pass the `int64` value. Only when the downstream contract requires another format, such as UUID, encode the ID into that required format and implement the inverse parser. Do not add a prefix, offset, hash, or other encoding merely to avoid decimal representation.
- A channel service parses every incoming mock-owned resource ID with its channel `service/id.go` helper before invoking a usecase that accepts an internal `ID`. The formatter and parser must be reversible and validate the downstream-required format; responses derive IDs from the relevant model or relation ID at the service boundary.

### Implemented Channel ID Formats

| Channel | OpenAPI and UI resource ID format | Webhook resource ID format | Source of truth |
| --- | --- | --- | --- |
| `slash` | UUID string. Encode the `int64` into the last 8 bytes of a canonical UUID; parse only that canonical form. | No implemented mock-owned webhook ID DTO. Use the same UUID rule when one is added. | `channel/slash/service/id.go` and Slash SDK DTOs. |
| `photonpay` | Decimal string: `strconv.FormatInt(ID, 10)` and `strconv.ParseInt`. | Decimal string for `ds-event` resource IDs such as `cardId` and `transactionId`. | `channel/photonpay/service/id.go` and `pkg/dealer/photonpay` DTOs. |
| `paynda` | Decimal string: `strconv.FormatInt(ID, 10)` and `strconv.ParseInt`. | Numeric `int64` for `xm-event` fields whose Marxo protobuf declares `int64`, including `merchantId`, `balanceAccountId`, `cardholderId`, and `cardId`. String webhook IDs retain decimal-string formatting. | `channel/paynda/service/id.go`, Marxo `notify.proto`, and `pkg/dealer/payndapay` DTOs. |

When a downstream DTO has mixed ID field types, apply the row's webhook rule per field. Do not infer a UUID requirement merely because an SDK uses Go `string`.
- Generic models are shared by every channel. Add a field only when a current channel implementation needs it and it represents a channel-neutral concept.
- `Account` is the channel account domain, not a `VirtualAccount` and not a webhook configuration. Keep `Account.Channel`: an account is scoped to the channel that owns it. Every persisted business table other than `Account` itself must carry both `AccountID` and `Channel`; `AccountID` references `Account.ID` and the row `Channel` must equal its account's `Channel`.
- Repositories must scope every `Exist`, `Find`, `List`, `Save`, and `Delete` operation by both `AccountID` and `Channel`. A child aggregate inherits that pair from its owning resource in the same transaction. Do not derive the account from an unrelated current/default account, and do not encode either value into a channel external resource ID.
- Account scope is an explicit business input. Channel service code parses its protocol account selector with `service/id.go`, then puts the internal value in the usecase/repository request struct as `AccountID`; do not put an account ID in `context.Context`, use a middleware-derived account scope, or parse then discard the value. A repository request that identifies an account-owned resource contains both `AccountID` and the resource ID. Its query always includes `channel`, `id`, and `account_id`; list requests include `channel` and add `account_id` when the caller selected an account. UI administration may omit the optional list filter to view all accounts, but create and account-specific mutations must carry a non-zero `AccountID`.
- Every uniqueness rule for account-owned data includes `account_id` and `channel` in addition to its natural key. `Account` itself has no `AccountID`; its `ID` is the domain key and `Channel` remains its channel scope.
- The browser account management UI is a first-class account-domain entry, grouped with but separate from Slash `VirtualAccount` management. It is not a channel configuration screen. A selected account scopes browser resource management; a channel only maps that scope into a downstream field when its actual contract requires one.
- Every implemented channel must resolve an account before accessing account-owned OpenAPI resources. Paynda resolves the decimal `balanceAccountId` path/body field; PhotonPay resolves decimal `app_id` during `oauth2/token/accessToken` and receives the resulting token in `X-PD-TOKEN` on subsequent requests; Slash resolves its UUID-formatted account ID from the Marxo SDK's `X-API-Key` header. Each service uses a local helper to parse that selector and explicitly passes `AccountID` to its usecase request. These are channel-specific account selectors, not interchangeable authentication rules.
- PhotonPay has no OpenAPI virtual-account selector. Creating a PhotonPay `Account` must create one default `VirtualAccount` and its wallet in the same transaction. UI virtual-account management remains available for additional accounts and funding operations. PhotonPay OpenAPI flows that require a shared card wallet resolve the earliest `VirtualAccount` for the selected account (`ID ASC`), then set `Card.VirtualAccountID` and `Card.WalletID`; they must not expose or infer a PhotonPay-specific virtual-account request parameter.
- `AuthorizationConfig` is the account-scoped configuration for synchronous authorization callbacks. It is separate from `WebhookConfig`, which subscribes asynchronous events. Its uniqueness is `(account_id, channel)` and it is managed through channel UI services; do not expose an OpenAPI configuration endpoint unless Marxo has an actual call site. A failed synchronous authorization callback rejects the simulated transaction; do not persist a fallback behavior.
- A card product is the configured card BIN. Every new card stores both `CardProductID` and the selected product prefix in `CardBin`; derive the prefix from the product rather than accepting it as independent card state.
- Do not add a field merely because it exists in a downstream SDK request or response. Keep such unused fields in the channel service DTO with an `Invalid:` comment and ignore them.
- Before changing a generic model, search the downstream implementation for actual request construction and response consumption. Do not infer persistence fields from third-party SDK type definitions alone.

## Backend Structure

- Use Gin with Kratos-style layering: `channel/<channel>/service`, `biz`, `data`, and `http`.
- Service methods use `Fn(context.Context, *Request) (*Response, error)`. They bind or convert channel DTOs only; business behavior belongs in `biz`.
- A service response is the channel DTO itself, never a protocol envelope such as `Response[T]`. The channel HTTP adapter wraps successful DTOs in its API envelope. Service translates Kratos business errors to channel-specific error reasons.
- `biz` owns repository interfaces and may inject repositories only. Do not inject another usecase or business object.
- `data` implements repositories. A repository method performs exactly one database operation.
- Use `samber/do` for dependency injection. Register constructors directly as providers and resolve dependencies inside those constructors; do not construct dependency graphs manually in `main`.
- Keep each channel in its own package and expose all routes under `/<channel>/...`.
- For Slash, browser management flows use `/slash/ui/...`; Marxo integrations use the Slash OpenAPI paths directly below `/slash/...`. UI handlers must not substitute for, or redefine, the OpenAPI contract.

## HTTP, Types, and Persistence

- Define typed request and response structs with `json`, `form`, `uri`, and `binding` tags as appropriate. Never use `gin.H`.
- Prefer a generic helper when identical behavior applies to multiple types; for example, use `pkg/types.Value[T]` for pointer value defaults.
- Use generic enums internally. A channel service converts them to the channel enum and declares its DTO field with that enum type directly. Do not declare a service enum field as `string` and cast between strings at a later layer.
- Use GORM Gen query objects for data access. Do not write literal SQL predicates such as `Where("field = ?")`.
- If a flow allows a resource to be absent, use a separate one-query `Exist` repository method before `Find`. Do not convert or inspect `RecordNotFound` to implement optional-resource behavior.
- In biz, log an unexpected repository error exactly once at its first handling point with direct `zap.S().Errorw`, then return a Kratos error. Do not hide logging behind a private error helper. Expected `Exist=false` results do not need error logs.
- Define all Kratos business errors centrally in the channel biz error definitions. Use named errors in usecases instead of constructing them with literal reasons or messages.
- Apart from `context.Context`, methods with multiple inputs accept a request struct.
- Use a dedicated biz `Transaction` interface with `InTx(ctx, func(ctx context.Context) error)` for write flows requiring more than one database operation. Keep aggregate repos as small separate interfaces. Data repositories resolve query objects through `repo.DB(ctx)` and must use the transaction stored in the context.
- Every list query needs explicit ordering. Default to primary `ID DESC` unless the channel contract says otherwise.
- Ignore authentication/signature validation unless a channel-specific requirement explicitly asks for it.
- Preserve unused SDK fields only in service DTOs, label them `Invalid:`, and do not persist or implement their behavior.
- A service DTO uses `time.Time` or `*time.Time` when the downstream JSON time is RFC3339, letting the HTTP codec perform the standard conversion. Use `string` only when the downstream contract requires a non-RFC3339 layout or a non-time textual value; keep that format conversion at the service boundary.

### GORM Model Schema

- Every persisted scalar field in `model` has an explicit GORM `column` and PostgreSQL `type` tag. Associations have no physical column tag; declare their `foreignKey` and `references` explicitly instead.
- Strings and string-backed enums use `type:varchar;not null;default:''`. Numeric IDs and counters use their explicit integer type with `not null;default:0`; decimal values use `type:numeric;not null;default:0`; booleans use `type:boolean;not null;default:false`; JSON values use `type:jsonb;not null;default:'{}'`.
- `time.Time` and `*time.Time` model fields use `type:timestamptz`. Required values are `not null`; optional pointer values use `default:null`.
- Model associations used by GORM Gen `Preload` are read-only: add `gorm:"...;->"` so aggregate saves never persist preloaded related objects.
- A card transaction represents one concrete transaction stage. Its `CreatedAt` is that stage's event time; do not add `OccurredAt` or `SettledAt` to `CardTransaction`. For a transaction DTO's authorization timestamp, preload its `Authorization` and use `Authorization.CreatedAt`, never the transaction's creation time.
- This mock may reset its local business tables when a model schema changes. Do not add compatibility field mappings or retain deprecated columns solely to preserve local mock data unless the user explicitly requires a migration.

## Workflow

- Read existing build and generation instructions before running commands. Use the repository generator for schema/query changes.
- Run `gofmt` and only the requested compilation check. Go commands must set `GOCACHE=/home/kar/.cache/go-build` and `GOTMPDIR=/home/kar/.cache/go-tmp`.
- Frontend code follows the repository's formatted multi-line style: keep imports, declarations, functions, object fields, and template elements legible on separate lines. Do not introduce compressed single-line components, handlers, or API definitions. After frontend changes, run `npm run build` from `frontend` when dependencies are available.
- Review the changed code after implementation for layer violations, raw predicates, magic protocol values, and unnecessary generic-model fields.

### Code Layout

1. Do not compress code merely to reduce line count.
2. A struct literal with multiple fields uses a multi-line layout, with one field per line.
3. Split a function call with many arguments, or one that would become long, across multiple lines.
4. Split a multi-variable assignment across multiple lines when its right side contains multiple fields or expressions.
5. Do not put multiple business operations on one line.
6. Passing `gofmt` is not a reason to retain a compressed one-line layout.
7. Prefer readability over line count.
8. Do not proactively compress existing multi-line code into one line while making a change.
9. New code follows these rules even when a compressed line would be 120 characters or fewer.
10. Before submitting Go changes, review hand-written code matched by `\{.*?,.*?,`. Treat each match as a readability warning: expand obvious multi-field literals, multi-argument calls, and combined business operations; exclude generated code and do not mechanically rewrite a legitimate single-field nested literal.
11. Before submitting repository queries, review hand-written code matched by `Where.*?,.*?.(Find|Count|Create)\(\)`. Split a `Where` call with multiple conditions and its terminal `Find`, `Count`, or `Create` call across lines; exclude generated GORM Gen code.
