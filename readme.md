# Generic Mock

用于模拟发卡渠道的服务。通用模型只保存渠道无关的数据；每个渠道在自己的 `channel/<channel>` 包中实现 OpenAPI、业务规则和 DTO 转换。

通用表结构及字段命名与渠道 API 的映射见 [通用模型映射](docs/channels/model-mapping.md)。

## 渠道菜单

| 渠道 | 菜单功能 | 状态 | 文档 |
| --- | --- | --- | --- |
| PhotonPay | 账户、持卡人、卡片、授权管理、卡交易、模拟交易、Webhook 管理 | 已实现 | [PhotonPay](docs/channels/photonpay.md) |
| Paynda | 账户、持卡人、卡片、授权管理、卡交易、模拟交易、Webhook 管理 | 已实现 | [Paynda](docs/channels/paynda.md) |
| Slash | 账户、虚拟账户、卡产品、持卡人、卡片、授权管理、卡交易、模拟交易、Webhook 管理 | 部分实现 | [Slash](docs/channels/slash.md) |
| Payful | 尚无菜单 | 未实现 | [Payful](docs/channels/payful.md) |
| UQPay | 尚无菜单 | 未实现 | [UQPay](docs/channels/uqpay.md) |


## 功能摘要

| 模块 | PhotonPay | Paynda | Slash | Payful | UQPay |
| --- | --- | --- | --- | --- |
| 持卡人 | 已实现 | 已实现 | 已实现 | 未实现 | 未实现 |
| 开卡 | 虚拟账户卡 | 独立余额卡 | 虚拟卡 | 未实现 | 未实现 |
| 授权 | 同步 sandbox 模拟 | 异步模拟，缺 webhook 投递 | 同步模拟 | 未实现 | 未实现 |
| 清算/冲正/退款 | 已实现 | 已实现 | 已实现 | 未实现 | 未实现 |
| Webhook 配置 | 已实现 | 已实现 | 已实现 | 未实现 | 未实现 |
| Webhook 投递与记录 | 已实现，自动重试待补 | 已实现，自动重试待补 | 未实现 | 未实现 | 未实现 |


## 运行方式

后端默认监听 `:8000`，数据库连接通过 `DATABASE_DSN` 配置。前端开发服务器默认监听 `:3000`，代理已实现渠道的 UI 路径。

```bash
cd backend
DATABASE_DSN='postgres://postgres:postgres@127.0.0.1:5432/generic_mock?sslmode=disable' go run .

cd ../frontend
bun dev
```

## 约束

1. 通用模型使用自增 `int64 ID`；渠道 DTO 用本渠道 `service/id.go` 格式化和反解析 mock 自有资源 ID。
2. 实现或调整 OpenAPI 前，先核对 `./marxo` 的 SDK 和实际调用点。
3. OpenAPI、浏览器 UI、业务 usecase 分层独立；仓储使用 GORM Gen，列表显式按 ID 倒序。
4. 模型变化后运行 generator，再执行 `gofmt` 和 `GOCACHE=/home/kar/.cache/go-build GOTMPDIR=/home/kar/.cache/go-tmp go build ./...`；除非明确要求，不运行单元测试。
5. `Account` 是保留 `Channel` 的渠道账户域，和 `VirtualAccount` 不同；除 `Account` 外所有业务表以 `(AccountID, Channel)` 隔离。详细规则见[通用模型映射](docs/channels/model-mapping.md)。
