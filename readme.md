# Generic Mock

用于模拟发卡渠道的服务。通用模型只保存渠道无关的数据；每个渠道在自己的 `channel/<channel>` 包中实现 OpenAPI、业务规则和 DTO 转换。

通用表结构及字段命名与渠道 API 的映射见 [通用模型映射](docs/channels/model-mapping.md)。

## 渠道菜单

| 渠道 | 菜单功能 | 状态 | 文档 |
| --- | --- | --- | --- |
| PhotonPay | 账户、持卡人、卡片、授权管理、卡交易、模拟交易、Webhook 管理 | 已实现 | [PhotonPay](docs/channels/photonpay.md) |
| Paynda | 账户、持卡人、卡片、授权管理、卡交易、模拟交易、Webhook 管理 | 已实现 | [Paynda](docs/channels/paynda.md) |
| Slash | 账户、虚拟账户、卡产品、持卡人、卡片、授权管理、卡交易、模拟交易、Webhook 管理 | 部分实现 | [Slash](docs/channels/slash.md) |
| PingPong | 尚无菜单 | 待实现；SDK 已阅读，已整理分阶段任务及多前缀能力 | [PingPong](docs/channels/pingpong.md) |
| Payful | 不提供菜单 | 已废弃，不再实现 | [Payful（归档）](docs/channels/payful.md) |
| UQPay | 尚无菜单 | 未实现 | [UQPay](docs/channels/uqpay.md) |


## 功能摘要

浏览器是总管理页面，各渠道的数据跨账户平铺展示，不设全局账户选择器或独立资金管理页。持卡人创建和开卡由 OpenAPI 完成。「账户」页面可直接调整账户余额（正数增加、负数扣减，不要求资金来源）；现有渠道的「虚拟账户」和「卡片」页面提供充值/转出，与其所属根账户钱包双向划转，同币种、同账户且资金总额守恒。「虚拟账户」创建时在表单中指定所属账户。未来 PingPong 的预算组普通卡有独立钱包，其充值/转出对手方是关联 VA 钱包，不是根账户钱包。

清算入口位于「授权管理」，支持部分清算和超额清算；剩余授权金额为负数时如实显示，钱包余额或剩余额度不足不阻止继续模拟清算。现有 Slash、PhotonPay、Paynda 的模拟授权也允许余额不足；PingPong 授权为例外，须检查卡自身可用余额。本次输入金额仍必须为正数。清算、撤销必须关联同账户同卡的有效授权；退款既可独立发起（持久化 `AuthorizationID = 0`），也可指定授权发起，不要求先有清算。普通资金充值/转出仍校验余额。交易列表不提供清算入口。

PingPong 的授权模式已确认、待实现：页面模拟授权在卡自身余额足够时本地成功，再异步推送 Webhook，不需要下游同意，不提供同步授权回调配置。投递失败只记录投递状态，不改变授权结果；重放不重复记账。预算余额不能代替卡余额，具体 Webhook 报文契约仍待补齐。

三个渠道的「授权管理」均可点击授权 ID 或「详情」打开模态框，查看授权、已清算、已撤销、已退款、剩余金额和关联卡交易，并直接创建清算、撤销、退款。剩余金额按「授权金额 − 已成功清算 − 已成功撤销」计算，退款单独累计，不恢复授权额度；例如授权 `12 USD`、清算 `24 USD` 后显示 `-12 USD`。列表支持账户、授权 ID、卡 ID、商户、授权状态和时间范围筛选。详情及后续操作携带该授权所属账户，接口按账户和渠道隔离。

授权详情仅展示实际留存的原始报文。当前未记录同步授权回调的请求、响应及响应码，页面明确提示缺失，不使用异步交易通知的投递记录冒充授权回调。

关联账户的 UI 接口直接返回 `account_id` 和 `account_name`，前端不再为展示名称额外查询账户列表。账户选项只在创建、编辑或选择账户时加载。独立退款可省略 `authorization_id`、传 `null` 或渠道格式的零 ID；关联退款必须提供该渠道格式的有效授权 ID。

注销是卡片的终态：Slash 的 `closed`、PhotonPay 的 `cancelled`、Paynda 的 `DELETED` 均不能恢复正常，也不能先冻结再恢复。管理页面禁用注销卡的状态操作，后端 UI/OpenAPI 在账户和渠道范围内锁定卡片后校验状态；重复注销保持幂等。正常卡的冻结、冻结卡的恢复不受影响。

Paynda 不支持虚拟账户，没有虚拟账户页面或管理接口；其资金操作仅包括账户余额调整，以及账户钱包与独立卡钱包之间的充值/转出。

Paynda 的商户级 `requestResults` 和 `merchant/wallets` 尚无经确认的账户映射，当前明确返回 HTTP 501。不能把 `appId` 当作 mock 账户主键，也不能按 `requestId` 跨账户检索。该部分需确认 Marxo 的商户凭据与账户映射后接入。

| 模块 | PhotonPay | Paynda | Slash | PingPong | Payful | UQPay |
| --- | --- | --- | --- | --- | --- | --- |
| 持卡人 | 已实现 | 已实现 | 已实现 | 待实现 | 已废弃 | 未实现 |
| 开卡 | 虚拟账户卡 | 独立余额卡 | 虚拟卡 | 待实现 | 已废弃 | 未实现 |
| 授权 | 同步 sandbox 模拟，配置接口已提供 | 异步模拟，交易与卡状态 webhook 投递已提供 | 同步模拟，授权配置 OpenAPI 已提供 | 本地校验卡余额后异步通知，待实现 | 已废弃 | 未实现 |
| 清算/冲正/退款 | 已实现 | 已实现 | 已实现 | 待实现 | 已废弃 | 未实现 |
| Webhook 配置 | 已实现 | 已实现 | 已实现 | 待实现 | 已废弃 | 未实现 |
| Webhook 投递与记录 | 已实现，自动重试待补 | 已实现，自动重试待补 | 异步投递待实现 | 待实现 | 已废弃 | 未实现 |


## 运行方式

后端默认监听 `:8000`，数据库连接通过 `DATABASE_DSN` 配置。前端开发服务器默认监听 `:3000`，代理已实现渠道的 UI 路径。

```bash
cd backend
DATABASE_DSN='postgres://postgres:postgres@127.0.0.1:5432/generic_mock?sslmode=disable' go run .

cd ../frontend
bun dev
```

## Webhook 投递日志

后端默认在服务日志中输出三个渠道的所有 Webhook 投递，包括交易通知、卡状态通知和手动重放，无需开启额外的 debug 开关：

- `webhook delivery started`（INFO）：渠道、账户、投递记录 ID、事件、来源资源、尝试次数、目标 URL、HTTP 方法和完整请求报文。
- `webhook delivery succeeded`（INFO）或 `webhook delivery failed`（ERROR）：同一投递记录 ID、请求头、HTTP 状态码、响应头、完整响应报文、耗时（毫秒）及失败原因。
- 非 2xx、连接失败、超时、响应读取失败以及 PhotonPay 未返回有效 `roger: true` 确认，均会打印失败日志。响应读取中断时保留已收到的状态码、响应头和部分响应报文。

可通过 `channel`、`account_id` 和 `webhook_record_id` 关联一次投递的开始与结果。日志包含完整业务报文和请求头（可能含签名等敏感信息），应限制日志访问和保留时间。

## Webhook 接收服务器

独立的调试接收端，不需要数据库或启动主服务。接受任意路径的 `POST`，在标准输出打印接收时间、来源地址、请求路径、请求头及完整报文，不校验签名、不写入数据库。成功接收后返回 HTTP 200 和 `{"roger":true}`，兼容本项目 Slash、PhotonPay、Paynda 的投递成功判定。JSON、纯文本等请求体均可接收，单次上限为 4 MiB。

```bash
cd backend
GOCACHE=/home/kar/.cache/go-build GOTMPDIR=/home/kar/.cache/go-tmp \
  go run ./cmd/webhook-receiver -addr 127.0.0.1:9000
```

将渠道的 Webhook 地址设置为接收端可达的地址，例如 `http://127.0.0.1:9000/slash/webhook`；也可以使用 `/photonpay/webhook` 或 `/paynda/webhook` 区分日志。

```bash
curl -i http://127.0.0.1:9000/slash/webhook \
  -H 'Content-Type: application/json' \
  -d '{"event":"aggregated_transaction.create","id":"demo"}'
```

默认仅监听本机。跨机器或容器联调时可使用 `-addr :9000`，并将回调地址中的主机替换成发送端可以访问的 IP 或域名；发送端不在同一网络命名空间时，不能使用 `127.0.0.1`。该工具没有鉴权，且会原样打印可能包含敏感信息的请求头和报文，仅用于受控开发环境，不要直接暴露到公网。

## SDK 联调

全量 SDK 联调使用独立数据库 `generic_mock_sdk_test`，不使用日常开发库 `generic_mock`：

```bash
bash sdk/test-contract.sh
```

脚本自动重建测试库、初始化表结构和种子数据、启动独立后端（默认 `127.0.0.1:18000`），测试结束后停止后端。测试库和日志保留，下一次运行时重置测试库。只初始化测试库、不运行测试时执行 `bash sdk/init-test-db.sh`。本地 PostgreSQL 的连接配置及安全限制见 `sdk/CONTRACT_TESTS.md`。

## 约束

1. 通用模型使用自增 `int64 ID`；渠道 DTO 用本渠道 ID 转换函数（统一位于 `channel/<channel>/pkg/idconv`，公开函数以 `To...` / `From...` 表达转换方向）格式化和反解析 mock 自有资源 ID。
2. 实现或调整 OpenAPI 前，先核对 `./marxo` 的 SDK 和实际调用点。
3. OpenAPI、浏览器 UI、业务 usecase 分层独立；仓储使用 GORM Gen，列表显式按 ID 倒序。
4. 模型变化后运行 generator，再执行 `gofmt` 和 `GOCACHE=/home/kar/.cache/go-build GOTMPDIR=/home/kar/.cache/go-tmp go build ./...`；除非明确要求，不运行单元测试。
5. `Account` 是保留 `Channel` 的渠道账户域，和 `VirtualAccount` 不同；除 `Account`、`CardProduct` 外所有业务表以 `(AccountID, Channel)` 隔离。详细规则见[通用模型映射](docs/channels/model-mapping.md)。

`CardProduct` 是渠道级配置，不包含 `AccountID`；产品查询按 `Channel` 隔离，唯一索引为 `(channel, prefix)`。同渠道账户共享产品及发卡序列，创建账户不再复制产品；卡片仍按账户隔离，且只能引用同渠道产品。

`CardProduct.Prefix` 使用英文逗号分隔的数字前缀字符串，例如 `424242,555555`。**只有未来的 PingPong 渠道允许多前缀**；Slash、PhotonPay、Paynda 及其他渠道仍只允许单前缀（例如 `424242`），配置逗号列表会拒绝开卡。开卡前校验所有候选项，从中随机选择一个作为 `Card.CardBin`，卡号也使用同一个前缀；产品的 `NextCardNumber` 仍在所有候选 BIN、所有账户间共享。空项、非数字或超出卡号长度容量的配置不会被静默忽略。现有单前缀数据和已发行卡片无需改写。

通用卡类型有三种：`single` 为无 VA 的独立卡，`share` 与 VA 共用钱包，`virtual_account_single` 关联 VA 但持有独立卡钱包。PingPong 的预算组对应 VA，其普通卡采用第三种类型，VA 仅为资金来源；余额展示和交易扣款始终使用 `Card.WalletID`，不能根据 VA 关联判定共享。`pkg/cardwallet` 已提供钱包分配与供资钱包解析，现有两类卡保留原语义。

PingPong 已具备通用渠道标识、多前缀和第三类卡钱包基础，但未提供 OpenAPI、Webhook 或管理页面。新加入的 `sdk/pingpong` 已核对 16 个 HTTP 方法，包含预算账户、卡与预算资金订单、v3 卡交易和 v4 账户交易；[后续任务](docs/channels/pingpong.md#后续任务顺序)按调用链确认、账户建模、开卡、资金报表和 UI 验收分阶段推进。Marxo 生产调用点和 Webhook 契约仍待补齐，不能把 SDK 示例当成生产接入证明。Payful 已废弃，其文档仅保留历史研究，不再作为待办。

旧版数据库若存在按账户复制的同渠道同 BIN 产品，需先整理重复产品及卡片引用，再移除 `card_products.account_id` 与旧索引；不能直接套用新的唯一索引。本次代码调整不会自动清空或合并现有业务数据。
