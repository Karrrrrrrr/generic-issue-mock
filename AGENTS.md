# Generic Mock Constraints

Read [AI_GUIDELINES.md](AI_GUIDELINES.md) before changing this repository.

- This service simulates channels for downstream systems. Do not use downstream-perspective names such as `ThirdPartyID` in generic models. Expose the PostgreSQL UUIDv7 string primary key `ID`; do not add a second resource token.
- Generic models are shared by all channels. Persist only channel-neutral concepts required by an implemented channel. Leave unused third-party API fields in the channel service DTO, tagged with an `Invalid:` comment.
- Use `channel/<channel>/{service,biz,data,http}`. Service converts typed HTTP DTOs; biz owns business behavior and repo interfaces; data uses GORM Gen.
- Repositories perform one database operation per method. Use `samber/do` for dependency injection, typed bind/tag DTOs rather than `gin.H`, and no raw GORM `Where("...")` predicates.
- Put every channel route under `/<channel>/...`; do not validate tokens unless explicitly required.
- Service methods return their channel DTO directly. The channel HTTP adapter owns the protocol success envelope; service translates business errors to channel error codes with Kratos errors.
- When a business flow permits a resource to be absent, call an explicit one-query `Exist` repository method before `Find`; never infer absence by handling `RecordNotFound` or a sentinel not-found error.
- Log unexpected repository errors once in biz at their first handling point with direct `zap.S().Errorw` calls. Do not wrap logging in a private error helper, log expected absence, or repeat a log while propagating the same error.
- Define all Kratos business errors centrally in the channel biz error definitions. Use those named errors in usecases; do not construct business errors with literal reasons or messages inline.
- Except for `context.Context`, methods with more than one argument use a request struct.
- Register layer constructors directly as `samber/do` providers. Do not manually resolve and thread dependencies through registration closures.
- Define `InTx(ctx, func(ctx context.Context) error)` only on a dedicated biz `Transaction` interface. Keep aggregate repos separate; their data implementations resolve GORM Gen queries through `repo.DB(ctx)` so every operation uses the transaction carried by the context.
- Every list query must include an explicit order. Unless a channel requirement says otherwise, order by primary `ID DESC`.
- Run the generator after model changes, then `gofmt` and `go build ./...` with the configured Go cache paths. Do not run unit tests unless asked.
