# Generic Mock Engineering Rules

## Model Ownership

- This repository simulates third-party channels for downstream systems. In this repository, a channel is the service being simulated; never carry downstream names such as `ThirdPartyID` into a generic model.
- Generic models use an auto-incrementing `int64` primary key `ID` for relations and transactions. Every resource exposed by a channel has a `display_id` external ID; do not expose `ID` outside the repository or add any other channel-specific resource token.
- A channel service must resolve every incoming mock-owned resource `display_id` through a channel repository `ExistByDisplayID` then `FindByDisplayID` flow before invoking a usecase that accepts an internal `ID`. Never parse, cast, derive, or otherwise treat `display_id` as `ID`; return the model's stored `DisplayID` in channel DTOs.
- Generic models are shared by every channel. Add a field only when a current channel implementation needs it and it represents a channel-neutral concept.
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
- Use generic enums internally. Convert them in a channel service to that channel's API enums.
- Use GORM Gen query objects for data access. Do not write literal SQL predicates such as `Where("field = ?")`.
- If a flow allows a resource to be absent, use a separate one-query `Exist` repository method before `Find`. Do not convert or inspect `RecordNotFound` to implement optional-resource behavior.
- In biz, log an unexpected repository error exactly once at its first handling point with direct `zap.S().Errorw`, then return a Kratos error. Do not hide logging behind a private error helper. Expected `Exist=false` results do not need error logs.
- Define all Kratos business errors centrally in the channel biz error definitions. Use named errors in usecases instead of constructing them with literal reasons or messages.
- Apart from `context.Context`, methods with multiple inputs accept a request struct.
- Use a dedicated biz `Transaction` interface with `InTx(ctx, func(ctx context.Context) error)` for write flows requiring more than one database operation. Keep aggregate repos as small separate interfaces. Data repositories resolve query objects through `repo.DB(ctx)` and must use the transaction stored in the context.
- Every list query needs explicit ordering. Default to primary `ID DESC` unless the channel contract says otherwise.
- Ignore authentication/signature validation unless a channel-specific requirement explicitly asks for it.
- Preserve unused SDK fields only in service DTOs, label them `Invalid:`, and do not persist or implement their behavior.

## Workflow

- Read existing build and generation instructions before running commands. Use the repository generator for schema/query changes.
- Run `gofmt` and only the requested compilation check. Go commands must set `GOCACHE=/home/kar/.cache/go-build` and `GOTMPDIR=/home/kar/.cache/go-tmp`.
- Review the changed code after implementation for layer violations, raw predicates, magic protocol values, and unnecessary generic-model fields.
