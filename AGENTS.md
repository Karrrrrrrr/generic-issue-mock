# Generic Mock Constraints

Read [AI_GUIDELINES.md](AI_GUIDELINES.md) before changing this repository.

- This service simulates channels for downstream systems. Do not use downstream-perspective names such as `ThirdPartyID` in generic models. Expose the UUID string primary key `ID`; do not add a second resource token.
- Generic models are shared by all channels. Persist only channel-neutral concepts required by an implemented channel. Leave unused third-party API fields in the channel service DTO, tagged with an `Invalid:` comment.
- Use `channel/<channel>/{service,biz,data,http}`. Service converts typed HTTP DTOs; biz owns business behavior and repo interfaces; data uses GORM Gen.
- Repositories perform one database operation per method. Use `samber/do` for dependency injection, typed bind/tag DTOs rather than `gin.H`, and no raw GORM `Where("...")` predicates.
- Put every channel route under `/<channel>/...`; do not validate tokens unless explicitly required.
- Run the generator after model changes, then `gofmt` and `go build ./...` with the configured Go cache paths. Do not run unit tests unless asked.
