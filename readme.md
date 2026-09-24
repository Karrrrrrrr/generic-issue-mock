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

浏览器是总管理页面，各渠道的数据跨账户平铺展示，不设全局账户选择器或独立资金管理页。持卡人创建和开卡由 OpenAPI 完成。「账户」页面可直接调整账户余额（正数增加、负数扣减，不要求资金来源）；「虚拟账户」和「卡片」页面提供充值/转出，只与其所属账户钱包双向划转，同币种、同账户且资金总额守恒。「虚拟账户」创建时在表单中指定所属账户。

清算入口位于「授权管理」，支持部分清算和超额清算；剩余授权金额为负数时如实显示，钱包余额或剩余额度不足均不阻止继续模拟授权、清算。本次输入金额仍必须为正数。清算、撤销必须关联同账户同卡的有效授权；退款既可独立发起（持久化 `AuthorizationID = 0`），也可指定授权发起，不要求先有清算。普通资金充值/转出仍校验余额。交易列表不提供清算入口。

关联账户的 UI 接口直接返回 `account_id` 和 `account_name`，前端不再为展示名称额外查询账户列表。账户选项只在创建、编辑或选择账户时加载。独立退款可省略 `authorization_id`、传 `null` 或渠道格式的零 ID；关联退款必须提供该渠道格式的有效授权 ID。

Paynda 不支持虚拟账户，没有虚拟账户页面或管理接口；其资金操作仅包括账户余额调整，以及账户钱包与独立卡钱包之间的充值/转出。

Paynda 的商户级 `requestResults` 和 `merchant/wallets` 尚无经确认的账户映射，当前明确返回 HTTP 501。不能把 `appId` 当作 mock 账户主键，也不能按 `requestId` 跨账户检索。该部分需确认 Marxo 的商户凭据与账户映射后接入。

| 模块 | PhotonPay | Paynda | Slash | Payful | UQPay |
| --- | --- | --- | --- | --- |
| 持卡人 | 已实现 | 已实现 | 已实现 | 未实现 | 未实现 |
| 开卡 | 虚拟账户卡 | 独立余额卡 | 虚拟卡 | 未实现 | 未实现 |
| 授权 | 同步 sandbox 模拟，配置接口已提供 | 异步模拟，交易与卡状态 webhook 投递已提供 | 同步模拟，授权配置 OpenAPI 已提供 | 未实现 | 未实现 |
| 清算/冲正/退款 | 已实现 | 已实现 | 已实现 | 未实现 | 未实现 |
| Webhook 配置 | 已实现 | 已实现 | 已实现 | 未实现 | 未实现 |
| Webhook 投递与记录 | 已实现，自动重试待补 | 已实现，自动重试待补 | 异步投递待实现 | 未实现 | 未实现 |


## 运行方式

后端默认监听 `:8000`，数据库连接通过 `DATABASE_DSN` 配置。前端开发服务器默认监听 `:3000`，代理已实现渠道的 UI 路径。

```bash
cd backend
DATABASE_DSN='postgres://postgres:postgres@127.0.0.1:5432/generic_mock?sslmode=disable' go run .

cd ../frontend
bun dev
```

## 约束

1. 通用模型使用自增 `int64 ID`；渠道 DTO 用本渠道 ID 转换函数（统一位于 `channel/<channel>/pkg/idconv`，公开函数以 `To...` / `From...` 表达转换方向）格式化和反解析 mock 自有资源 ID。
2. 实现或调整 OpenAPI 前，先核对 `./marxo` 的 SDK 和实际调用点。
3. OpenAPI、浏览器 UI、业务 usecase 分层独立；仓储使用 GORM Gen，列表显式按 ID 倒序。
4. 模型变化后运行 generator，再执行 `gofmt` 和 `GOCACHE=/home/kar/.cache/go-build GOTMPDIR=/home/kar/.cache/go-tmp go build ./...`；除非明确要求，不运行单元测试。
5. `Account` 是保留 `Channel` 的渠道账户域，和 `VirtualAccount` 不同；除 `Account`、`CardProduct` 外所有业务表以 `(AccountID, Channel)` 隔离。详细规则见[通用模型映射](docs/channels/model-mapping.md)。

`CardProduct` 是渠道级配置，不包含 `AccountID`；产品查询按 `Channel` 隔离，唯一索引为 `(channel, prefix)`。同渠道账户共享产品及发卡序列，创建账户不再复制产品；卡片仍按账户隔离，且只能引用同渠道产品。

旧版数据库若存在按账户复制的同渠道同 BIN 产品，需先整理重复产品及卡片引用，再移除 `card_products.account_id` 与旧索引；不能直接套用新的唯一索引。本次代码调整不会自动清空或合并现有业务数据。
