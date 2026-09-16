# Generic Mock Engineering Rules

## Model Ownership

- This repository simulates third-party channels for downstream systems. In this repository, a channel is the service being simulated; never carry downstream names such as `ThirdPartyID` into a generic model.
- Expose the generic model primary key `ID` as the channel-facing resource ID. `ID` defaults to a UUID string. Do not add a second token or external-ID column for resources owned by this mock.
- Generic models are shared by every channel. Add a field only when a current channel implementation needs it and it represents a channel-neutral concept.
- Do not add a field merely because it exists in a downstream SDK request or response. Keep such unused fields in the channel service DTO with an `Invalid:` comment and ignore them.
- Before changing a generic model, search the downstream implementation for actual request construction and response consumption. Do not infer persistence fields from third-party SDK type definitions alone.

## Backend Structure

- Use Gin with Kratos-style layering: `channel/<channel>/service`, `biz`, `data`, and `http`.
- Service methods use `Fn(context.Context, *Request) (*Response, error)`. They bind or convert channel DTOs only; business behavior belongs in `biz`.
- `biz` owns repository interfaces and may inject repositories only. Do not inject another usecase or business object.
- `data` implements repositories. A repository method performs exactly one database operation.
- Use `samber/do` for dependency injection. Do not construct dependency graphs manually.
- Keep each channel in its own package and expose all routes under `/<channel>/...`.

## HTTP, Types, and Persistence

- Define typed request and response structs with `json`, `form`, `uri`, and `binding` tags as appropriate. Never use `gin.H`.
- Prefer a generic helper when identical behavior applies to multiple types; for example, use `pkg/types.Value[T]` for pointer value defaults.
- Use generic enums internally. Convert them in a channel service to that channel's API enums.
- Use GORM Gen query objects for data access. Do not write literal SQL predicates such as `Where("field = ?")`.
- Ignore authentication/signature validation unless a channel-specific requirement explicitly asks for it.
- Preserve unused SDK fields only in service DTOs, label them `Invalid:`, and do not persist or implement their behavior.

## Workflow

- Read existing build and generation instructions before running commands. Use the repository generator for schema/query changes.
- Run `gofmt` and only the requested compilation check. Go commands must set `GOCACHE=/home/kar/.cache/go-build` and `GOTMPDIR=/home/kar/.cache/go-tmp`.
- Review the changed code after implementation for layer violations, raw predicates, magic protocol values, and unnecessary generic-model fields.
