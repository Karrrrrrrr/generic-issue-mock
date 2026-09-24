# PingPong

## 状态

未来渠道，待实现。2026-09-24 已阅读新加入的 `sdk/pingpong`：已核对 16 个 HTTP 方法、请求/响应 DTO、枚举、接口包装和测试调用，后续不再把 SDK 缺失列为阻塞项。当前后端仍只有通用渠道标识与产品多前缀能力，没有 PingPong OpenAPI、Webhook、管理菜单或种子账户。

本次只调整实现计划，不修改 SDK 或实现渠道。SDK 能力清单不等于 Marxo 生产调用清单；不注册占位成功接口，也不把测试中的真实资源样例当成 mock 的 ID 格式要求。

## 已检查的依据与缺口

| 依据 | 本次结论 |
| --- | --- |
| `sdk/pingpong/pingpong.go` | 实际发送的 HTTP 方法、路径、查询参数、Bearer 头、POST 签名头及响应解包逻辑。 |
| `sdk/pingpong/request.go` / `response.go` | 字段名称、JSON 类型、分页结构、预算账户与卡资金订单、两种交易报表。 |
| `sdk/pingpong/enums.go` / `pingpong_interface.go` | 卡状态、操作、资金订单状态和 token 失效判定；`CreateCardPre` 仅序列化请求，不是额外 HTTP 接口。 |
| `sdk/pingpong/pingpong_interface_test.go` | 16 个方法的本地 HTTP 示例；其中 Authorization 断言与实际发送逻辑不一致，不能把测试视为已通过的契约。 |
| `sdk/pingpong/pingpong_test.go` | 含真实网关、凭据及开卡、资金和状态变更调用；只阅读，未运行，未复制凭据。 |
| Marxo 生产调用点 | 当前仓库 `marxo` 快照未检索到 PingPong 业务接入；约定的外部 Marxo 路径也不可用。仍需补齐请求构造、响应消费、幂等探测和 token 缓存调用点。 |
| Webhook / 授权回调 | SDK 未提供 payload、事件、签名、成功 ACK 或消费端，保持待确认，不从其他渠道推导。 |

## 已确认的传输边界

- mock 路由统一增加 `/pingpong`，SDK 基础地址配置到该前缀；保留后续 `/v2`、`/api/issuing/v3`、`/api/issuing/v4` 路径，不能混为一种版本。
- 取 token 是 `GET /v2/token/get`，query 为 `app_id`、`app_secret`；业务请求实际发送 `Authorization: Bearer <token>`。POST JSON 另带 `sign`、`sign-version`。按项目规则默认不验密钥、签名或 token 有效性，但账户选择仍必须显式解析并传入 biz，不能省略账户隔离。
- HTTP 层拟统一输出 SDK 已支持的 `{"code":0,"message":"success","data":...}`；service 只返回业务 DTO。SDK 可解析数字或字符串 `code`，只认 `0` 为成功；无 `code` 的直接 DTO 也是其兼容分支，不必据此增加第二套成功响应。
- 错误边界保留 `code`、`message`（SDK 也兼容 `msg`）、`details.reason`。`IsInvalidToken` 识别 `1002` 或 `Invalid Token`，不能把数据库错误、资源缺失等一概翻译成 token 失效。其他业务错误码仍需真实调用点或报文确认。
- SDK 资源 ID 字段为字符串，目前未发现 SDK 对卡、预算、产品代码等资源 ID 施加格式校验。mock 默认按项目规则使用十进制 `int64` 字符串及 `To...` / `From...`；不照抄测试中的 `card...`、`ci...` 或字母产品代码。`app_id` 是应用凭据选择字段，其账户映射不能仅因字符串或数字外观就认定等于 `Account.ID`。
- `request_id`、`unique_order_id` 是调用方请求/幂等键，不当作 mock 资源 ID 转换；`record_id` 和查询字段 `order_id` 的对应关系须单独确认。

## HTTP 能力清单

以下是 SDK 已暴露的候选实现范围，**全部待实现**；所有生产调用到的方法必须覆盖，尚未取得生产调用证据的方法在 P0 明确优先级，不宣称它们已被业务使用。表中路径已包含 mock 渠道前缀。

| SDK 方法 | HTTP 与路径 | 关键输入 → `data` 结构 |
| --- | --- | --- |
| `GetAccessToken` | `GET /pingpong/v2/token/get` | query `app_id`、`app_secret` → `access_token`、`expires_in`。 |
| `QueryAccountsBalances` | `POST /pingpong/api/reporting/v3/account-balance` | `account_type`、`currency_list`、`subaccount_id_list`、`page_no`、`page_size` → `item_list`、`total_num`、分页字段。 |
| `QueryCardProducts` | `GET /pingpong/api/issuing/v3/card-products` | → `product_list`；保留 `card_product_code`、`bin_range`、`share` 等 SDK 字段。 |
| `CreateCard` | `POST /pingpong/api/issuing/v3/cards/apply` | `request_id`、`card_product_code`、`card_currency`、`budget_id`、可选持卡人/限额等 → `card_id`；不自行增加 `card_type` 请求字段。 |
| `GetCardDetails` | `GET /pingpong/api/issuing/v3/cards/details` | query `card_id` → 卡状态、预算关联、卡号/CVC、币种、销卡标记、限额和时间等详情字段。 |
| `CardFunding` | `POST /pingpong/api/issuing/v3/cards/funding/actions` | `card_id`、`action`、`amount`、`unique_order_id` → `record_id`、`action`、`amount`、`currency`。 |
| `QueryCardFundingOrders` | `GET /pingpong/api/issuing/v3/card/funding/orders` | query 分页、`start_date`、`end_date`、`status`、`unique_order_id`、`card_id` → `list`、`total_num`、分页字段。注意路径是单数 `card`。 |
| `QueryDedicatedCardBalance` | `GET /pingpong/api/issuing/v3/card/balance` | query `card_id` → `card_number`、`available_balance`、`currency`。 |
| `CardAction` | `POST /pingpong/api/issuing/v3/cards/actions` | `card_id`、`action`、可选 `remark` → SDK 不消费 `data`；仍返回可解析的 JSON 成功 envelope。 |
| `QueryCardTransactions` | `GET /pingpong/api/issuing/v3/transactions` | query 卡、类型、状态、清算类型、时间与分页 → `transaction_list`、`total_num`、分页字段；金额字段为字符串。 |
| `Query3DSDetails` | `GET /pingpong/api/issuing/v3/cards/3ds/details` | query `card_id` → 持卡人身份、联系信息和安全问答；没有配套的持卡人创建方法，需确认实际业务需求。 |
| `CreateBudgetAccount` | `POST /pingpong/api/issuing/v3/budgets` | 仅 `budget_name` → `budget_id`；不增加 `account_type` 或币种请求字段。 |
| `BudgetFunding` | `POST /pingpong/api/issuing/v3/budgets/funding` | `budget_id`、`action`、`amount`、`currency`、按操作提供订单键/目标预算及币种 → `record_id`。 |
| `QueryBudgetFundingOrder` | `GET /pingpong/api/issuing/v3/funding/orders` | query `order_id`、`action` → `order_id`、`action`、`status`。 |
| `QueryBudgetAccountBalance` | `GET /pingpong/api/issuing/v3/budget/balance` | 可选 query `budget_id`，省略查所属账户全部预算 → `balance_list`。注意路径是单数 `budget`。 |
| `QueryAccountTransactions` | `GET /pingpong/api/issuing/v4/account/transactions` | query 预算/卡、入账区间、类型、方向、分页 → `transaction_list`、`total_num`、分页字段；金额为 JSON number，与 v3 卡交易不同。 |

### DTO、筛选和状态约束

- v3 卡交易查询字段必须完整保留：`card_id`、`type`、`page_no`、`page_size`、`start_time`、`end_time`、`start_posting_date`、`end_posting_date`、`start_created_date`、`end_created_date`、`clear_type`、`status`。时间格式与交易类型/状态值不能只凭字段名称推断。
- v4 账户交易 query 为 `budget_id`、`card_id`、`posting_start_time`、`posting_end_time`、`transaction_type`、`direction`、`page_no`、`page_size`。SDK 注释要求预算/卡至少一个，区间不超过 31 天、每页最多 100；示例使用带时区的 ISO 8601。两种资源筛选同时给出时必须交集匹配，不忽略其中一个来扩大账户范围。
- SDK 中可选标量常用零值/`omitempty`，mock DTO 仍用指针辨别省略与显式非法值；repo 多值筛选用切片和 `IN`。默认 `ID DESC`，分页计数与列表必须使用相同过滤条件。
- 卡状态为 `INACTIVE`、`ACTIVE`、`REVOKED`（冻结）、`CANCELLED`；注销后不可恢复。卡资金操作为 `top_up` / `withdraw`；预算资金操作为 `top_up` / `transfer`；资金订单状态为 `SUCCESS` / `FAIL` / `PROCESSING`。为 service 定义渠道枚举，不照抄 SDK 的裸 `string` 字段。
- `CardActionRequest` 与实际发送方法的注释还包含 `update_remark`，并要求 `close` / `update_remark` 提供备注，但 `CardAction` 枚举只列出三种状态操作。将这个差异作为待核对项，不直接据枚举删掉 SDK 支持的请求值。
- 财务运算和持久化继续用 decimal；DTO 按 SDK 区分 number 和 string，不能把所有金额统一序列化成字符串。`Amount` 限额对象包含 `amount` 和 `currency`。
- SDK 把日期时间声明成字符串不代表 mock 也应如此。确认实际格式后，RFC3339 用 `time.Time`，其他完整时间使用项目时间封装；卡有效期等部分日期才使用对应格式字符串。不得为报表增加 `CardTransaction.OccurredAt` / `SettledAt`，授权时间从关联授权读取。

## 卡产品与 BIN

- `CardProduct` 是渠道级配置，`Channel = pingpong`，没有 `AccountID`；同渠道账户共享产品及发卡序列。
- `Prefix` 是英文逗号分隔的数字前缀字符串，例如 `424242,555555`；单个前缀也有效。**目前只有 PingPong 允许多前缀**，其他渠道仍要求单前缀。
- 开卡使用 `pkg/cardnumber.Generate`，传入渠道、产品 `Prefix` 和递增后的产品序号。先检查所有候选项，空项、非数字或无法容纳当前序号时拒绝，不随机跳过错误配置。
- 从候选项中随机选出一个 BIN，同时用于 `Card.CardBin` 和卡号前缀。`CardProductID` 仍引用原产品，不把逗号列表写进卡片，也不为每个 BIN 复制产品。
- `NextCardNumber` 在该产品的所有 BIN 和账户间共享，未来开卡流程须沿用事务内锁定产品、推进序号、创建卡片的方式。
- 保留 `(channel, prefix)` 唯一索引，约束完整配置字符串；它不表示不同产品的候选 BIN 集合互斥。既有单前缀产品和已发行卡片不改写。

## 模型映射与必须先解决的问题

| 领域 | 后续设计方向 | 尚不能直接拍板的部分 |
| --- | --- | --- |
| 应用与账户 | `Account` 仍是根账户域，所有账户资源显式携带内部 `AccountID` 与 `Channel`。 | `app_id` 到 mock 账户的映射、token 如何携带/解析账户，以及 `subaccount_id` 的含义；不得选默认账户或以全局查找资源反推账户。 |
| 产品与多 BIN | `card_product_code` 拟由 `CardProduct.ID` 转换；`bin_range` 对应 `Prefix`，选中的单个 BIN 用于开卡。 | SDK 只声明 `bin_range` 为字符串，没有拆分规则；逗号格式来自本项目需求。`share`、卡组织、币种等若决定实际开卡行为，实施时再评估必要的中立产品字段。 |
| 预算账户 | 优先评估复用 `VirtualAccount` 与钱包，不能把预算等同于根 `Account`。 | 创建请求只有名字，余额返回可以包含各币种；确认默认币种、是否多币种钱包、预算充值来源与跨币种 transfer 规则，不臆造汇率或新增创建字段。 |
| 独立卡 / 共享卡 | 根据产品的 `share` 与真实消费决定余额来源，`budget_id` 保留业务关联。 | 开卡请求总有 `budget_id`，但 SDK 又提供独立卡余额/资金接口；现有模型约定非空 `Card.VirtualAccountID` 就是共享钱包，不能仅看是否有预算 ID 判断共享。先明确关联与钱包语义，不能让独立卡误扣预算钱包。 |
| 卡资金 / 预算资金订单 | 必须留存可查询的订单状态、调用方幂等键和对应账务记录；重复请求不重复扣款。 | 评估中立资金订单/流水能力，不能用 `Card.LastOperation*` 代替历史订单，也不能为预算流水伪造卡。确认 `record_id`、`order_id`、`unique_order_id` 的区别及重复键不同参数的错误语义。 |
| 两类交易报表 | v3 卡交易与 v4 已入账账户交易分别转换 DTO，共享可复用的账务数据。 | v3 返回 `authorization_id`、无独立 `transaction_id`；v4 返回 `transaction_id`、`direction`、前后余额。确认阶段关联、金额符号、入账时间和余额快照，不能机械复用一个 DTO 或伪造历史余额。 |
| 3DS / 持卡人 / 限额 / 备注 | 只有确认业务使用且具备中立意义时才持久化；其余字段留在 service DTO 并标 `Invalid:`。 | 当前只有 3DS 读取、没有持卡人创建，不能把其他渠道接口搬过来；也不能把明确需要状态变更的功能伪装为成功但不留存。 |

普通充值/转出/预算划转保持账户、币种和余额校验及事务内按钱包 ID 顺序加锁；模拟授权/清算仍遵循项目允许负余额和超额清算的规则，不能混为普通资金转移规则。

## 后续任务顺序

### P0：补齐调用链与契约差异

- [x] 阅读 SDK 16 个方法、DTO、枚举及现有调用示例，登记输入输出清单。
- [ ] 取得 Marxo PingPong 实际业务调用点，逐方法标明请求构造、响应消费、错误分支及优先级；不能把 SDK 测试当成生产调用证据。
- [ ] 确认应用/账户选择、独立卡与共享卡、预算币种及资金来源、订单 ID/幂等和两类交易报表语义；收集必要响应样例，明确时间布局与错误码。
- [ ] 处理 SDK 集成前置问题：实际请求为 Bearer，但接口测试断言裸 token；`request` 新建 Resty client 而非使用配置过中间件的 `p.resty`，且无条件 `EnableDebug()`。后续另行修正 SDK/测试一致性和敏感日志，不能通过让 mock 偏离实际请求来掩盖。
- [ ] 补齐可用的 `tman` 依赖。当前 SDK 导入 `tman/pkg/dealer/crypto.Sign`，测试引用 `traffic.PingPongTrafficLog`，仓库附带的相关快照未找到这些定义；先确认正确依赖版本，不宣称可直接编译或已通过测试。
- [ ] 隔离真实测试配置：改为显式启用真实网关测试，清理硬编码凭据并在必要时轮换；不得直接运行当前含真实开卡/资金变更的测试来验证 mock。

### P1：账户隔离与最小模型设计

- [ ] 根据 P0 的实际调用确认最小模型变更，优先复用账户、预算、钱包、卡产品、卡、授权、交易；只有资金订单/产品能力/钱包关联确有缺口时才扩充中立概念。
- [ ] 建立 `channel/pingpong/{enums,pkg/idconv,service,biz,data,http}`，按账户、产品、预算、卡、资金订单、交易领域分文件；UI/OpenAPI 分离，biz 仓储请求独立定义。
- [ ] 完成账户选择、双向十进制 ID 转换、typed DTO、HTTP envelope 与渠道错误转换；不实施 token/签名真实性验证。
- [ ] 模型变化后运行生成器、`gofmt` 与配置缓存路径的 `go build ./...`；保持其他渠道单前缀、账户隔离和原有资金行为。

### P2：预算、产品与开卡闭环

- [ ] 实现 token、根账户余额、预算创建/余额、产品列表，以及实际开卡依赖的预算资金操作；保证 SDK 原始字段和路径。
- [ ] 开卡先验证同账户预算、同渠道产品、币种和请求幂等；锁产品共享序号后随机选一个合法 BIN，原子创建卡及正确的钱包关联。
- [ ] 实现卡详情、独立卡余额、冻结/解冻/注销；`CANCELLED` 终态，重复操作及开卡请求保持已确认的幂等语义。
- [ ] 根据实际调用落实 `update_remark`、3DS 与限额支持范围；未实现能力不能以空数据或假成功标为完成。

### P3：资金订单与报表闭环

- [ ] 完成卡 `top_up` / `withdraw`、预算 `top_up` / `transfer` 及对应查询，补齐 P2 所需的资金链路；相同幂等键重试不重复记账，失败不得部分更新钱包。
- [ ] 实现两类交易查询及全部实际使用的过滤条件、分页、方向、number/string 金额差异和授权关联，输出真实留存的账务字段。
- [ ] 覆盖已确认的所有业务调用端点；Webhooks/授权回调在取得契约和消费端后另列任务，不作为当前假定能力。

### P4：管理 UI 与验收

- [ ] 新增独立 PingPong 管理页面：账户、预算、产品、卡、资金订单和交易；授权/模拟页面按已确认业务补齐，不照搬其他渠道不适用的模块。开卡与持卡人创建仍归 OpenAPI。
- [ ] 页面遵循中文 ghost 标签、金额币种同标签、dayjs datetime、账户名称由后端返回、足够间距与可见分页；多 BIN 产品展示候选列表，卡详情展示实际选中 BIN。
- [ ] 用户要求测试时，再补充离线 SDK 协议检查与隔离 mock 集成验收。`sdk/test-contract.sh` 当前仅覆盖既有三个渠道，并未包含 PingPong；必须显式接入，不能直接跑真实网关测试。
- [ ] 验收包括账户越权、多 BIN 选择/单前缀回归、独立/共享钱包、并发幂等、普通转出余额不足、注销不可恢复、分页过滤、错误 envelope 和时间金额格式；全部通过后再把状态从「待实现」更新为实际完成度。

各阶段的未知项只阻塞依赖它的功能，不阻止整理已确认 DTO；但在生产调用点未核对前，不把候选端点清单宣称为完成的 OpenAPI 契约。
