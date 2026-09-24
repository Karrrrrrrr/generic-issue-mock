# PingPong

## 状态

部分实现（2026-09-24）。已按 `sdk/pingpong` 接入 12 个核心 HTTP 方法和独立管理页面：账户、虚拟账户、产品、卡、授权、资金订单。完成虚拟账户供资的独立卡钱包、随机多 BIN、资金订单幂等及本地授权预占。默认只初始化渠道产品，不创建 PingPong 种子账户。

**Webhook 尚未发送**：SDK 未包含事件、payload、签名和 ACK 契约，当前页面明确显示 `contract_pending`（协议待补齐／未投递），不是待重试投递记录。根账户余额报表、两类交易报表和 3DS 仍返回 HTTP 501；卡备注更新同样返回 501，不伪装为成功。

用户已确认预算组和普通卡的关联语义，见下一节；当前仍不是完整的 16 个 SDK 方法实现。SDK 能力清单不等于 Marxo 生产调用清单；不注册占位成功接口，也不把测试中的真实资源样例当成 mock 的 ID 格式要求。

## 命名边界

「预算组 / budget」仅属于 PingPong OpenAPI 的对外契约。HTTP 路径、SDK DTO 的 `budget_id`、`budget_name`、`target_budget_id` 等字段保持不变；service 在边界转换为 `VirtualAccountID`、`Name` 和 `TargetVirtualAccountID` 后再调用 usecase。

内部 usecase（包括 OpenAPI usecase）、仓储、变量、日志、文件名和管理页面一律称虚拟账户，不使用 `Budget`、`budgetRepo` 等内部名称。UI 使用 `/pingpong/ui/virtual-accounts`、页面 `/pingpong/virtual-accounts`，创建响应及卡关联字段为 `virtual_account_id`；资金类型为 `virtual_account_top_up` / `virtual_account_transfer`。通用资金类型的数据库值原本即为这两个字符串，本次仅重命名 Go 常量，不更改存量数据。

## 已确认：虚拟账户供资的普通卡

- 内部使用 `VirtualAccount`，OpenAPI `budget_id` 对应 `Card.VirtualAccountID`；它表示卡的归属与资金来源，不代表共享余额。
- 此类 PingPong 卡使用新增的通用 `CardType_VirtualAccountSingle`（`virtual_account_single`）。卡有自己的 `WalletType_Card` 钱包，`Card.WalletID` 必须与虚拟账户 `WalletID` 不同。
- `pkg/cardwallet.Prepare` 已支持第三类卡：验证 VA 及钱包在同账户、同渠道、同币种，保留 VA 关联并分配新的零余额卡钱包。调用方在同一事务创建钱包和卡；不得直接复制虚拟账户余额。
- `top_up` 从关联虚拟账户钱包转入独立卡钱包，`withdraw` 从卡钱包退回同一虚拟账户钱包；`FundingWalletID` 已支持从第三类卡的 VA 关系定位供资钱包。普通转出仍检查余额，不允许直接绕过虚拟账户与根账户划转。
- 授权、清算、退款等卡交易只读写卡自己的钱包；不能因为 `VirtualAccountID != nil` 或预加载了 VA 就改用虚拟账户余额。
- `single`（无 VA 的独立卡）与 `share`（共享 VA 钱包的卡）仍保留，不改变现有渠道对外卡类型。SDK 产品的 `share` 字段不是推翻上述用户确认的理由；若以后接入 PingPong 的另一种共享产品，再单独核对其契约。

## 已确认：本地授权与异步通知

- 页面发起模拟授权时，由 mock 本地决定结果：金额必须为正，卡自身可用余额足够即可授权，不需要下游同意；账户归属、币种和卡状态等基本合法性检查仍保留。
- 只检查 `Card.WalletID` 对应的独立卡钱包，不以虚拟账户或根账户余额代替。余额不足则拒绝本次授权，不发送授权成功通知。这是 PingPong 对通用「模拟授权允许余额不足」规则的明确例外，不改变其他渠道。
- 在同一账户/渠道事务中锁定卡钱包，完成余额检查及本地授权、账务和待投递记录的一致性写入；提交后再异步推送 Webhook，不持有数据库事务等待网络应答。
- 不发起同步授权回调，不依赖或提供 PingPong `AuthorizationConfig` 页面。Webhook 表达已发生的授权结果，HTTP ACK 是投递确认，不是业务批准。
- 推送超时、失败或下游拒绝接收只改变投递状态，不能把本地已成功的授权改成拒绝或回滚；重放/重试不得再次授权、预占或扣款。
- 异步模式与本地余额检查已确认，待补充的仅是实际事件名、payload、签名和投递 ACK 等协议细节，不能因此重新引入下游审批。清算仍按通用规则允许超额及负余额。

## 已检查的依据与缺口

| 依据 | 本次结论 |
| --- | --- |
| `sdk/pingpong/pingpong.go` | 实际发送的 HTTP 方法、路径、查询参数、Bearer 头、POST 签名头及响应解包逻辑。 |
| `sdk/pingpong/request.go` / `response.go` | 字段名称、JSON 类型、分页结构、预算账户与卡资金订单、两种交易报表。 |
| `sdk/pingpong/enums.go` / `pingpong_interface.go` | 卡状态、操作、资金订单状态和 token 失效判定；`CreateCardPre` 仅序列化请求，不是额外 HTTP 接口。 |
| `sdk/pingpong/pingpong_interface_test.go` | 16 个方法的本地 HTTP 示例；其中 Authorization 断言与实际发送逻辑不一致，不能把测试视为已通过的契约。 |
| `sdk/pingpong/pingpong_test.go` | 含真实网关、凭据及开卡、资金和状态变更调用；只阅读，未运行，未复制凭据。 |
| Marxo 生产调用点 | 当前仓库 `marxo` 快照未检索到 PingPong 业务接入；约定的外部 Marxo 路径也不可用。仍需补齐请求构造、响应消费、幂等探测和 token 缓存调用点。 |
| Webhook / 授权模式 | 用户已确认本地余额校验后授权、异步通知，无同步授权回调；SDK 未提供 payload、事件、签名、成功 ACK 或消费端，这些报文细节仍待确认，不从其他渠道推导。 |

## 已确认的传输边界

- mock 路由统一增加 `/pingpong`，SDK 基础地址配置到该前缀；保留后续 `/v2`、`/api/issuing/v3`、`/api/issuing/v4` 路径，不能混为一种版本。
- 取 token 是 `GET /v2/token/get`，query 为 `app_id`、`app_secret`；业务请求实际发送 `Authorization: Bearer <token>`。POST JSON 另带 `sign`、`sign-version`。按项目规则默认不验密钥、签名或 token 有效性，但账户选择仍必须显式解析并传入 biz，不能省略账户隔离。
- HTTP 层统一输出 SDK 已支持的 `{"code":0,"message":"success","data":...}`；service 只返回业务 DTO。SDK 可解析数字或字符串 `code`，只认 `0` 为成功；无 `code` 的直接 DTO 也是其兼容分支，不必据此增加第二套成功响应。
- 错误边界保留 `code`、`message`（SDK 也兼容 `msg`）、`details.reason`。`IsInvalidToken` 识别 `1002` 或 `Invalid Token`，不能把数据库错误、资源缺失等一概翻译成 token 失效。其他业务错误码仍需真实调用点或报文确认。
- SDK 资源 ID 字段为字符串，目前未发现 SDK 对卡、预算、产品代码等资源 ID 施加格式校验。mock 默认按项目规则使用十进制 `int64` 字符串及 `To...` / `From...`；不照抄测试中的 `card...`、`ci...` 或字母产品代码。`app_id` 是应用凭据选择字段，其账户映射不能仅因字符串或数字外观就认定等于 `Account.ID`。
- `request_id`、`unique_order_id` 是调用方请求/幂等键，不当作 mock 资源 ID 转换。当前 mock 约定 `record_id` 与预算查询 `order_id` 都指向 `WalletTransfer.ID` 的十进制字符串；这不等同于已核对真实生产语义。

## HTTP 能力清单

以下按 SDK 实际发送方法列出实现状态。生产调用点尚缺失，不宣称这些方法均已被 Marxo 使用；取得调用点后仍须逐一核验。表中路径已包含 mock 渠道前缀。

| SDK 方法 | HTTP 与路径 | 关键输入 → `data` 结构 | 当前状态 |
| --- | --- | --- | --- |
| `GetAccessToken` | `GET /pingpong/v2/token/get` | query `app_id`、`app_secret` → `access_token`、`expires_in`。 | 已接入 |
| `QueryAccountsBalances` | `POST /pingpong/api/reporting/v3/account-balance` | `account_type`、`currency_list`、`subaccount_id_list`、`page_no`、`page_size` → `item_list`、`total_num`、分页字段。 | 501，待契约确认 |
| `QueryCardProducts` | `GET /pingpong/api/issuing/v3/card-products` | → `product_list`；保留 `card_product_code`、`bin_range`、`share` 等 SDK 字段。 | 已接入 |
| `CreateCard` | `POST /pingpong/api/issuing/v3/cards/apply` | `request_id`、`card_product_code`、`card_currency`、`budget_id`、可选持卡人/限额等 → `card_id`；不自行增加 `card_type` 请求字段。 | 已接入 |
| `GetCardDetails` | `GET /pingpong/api/issuing/v3/cards/details` | query `card_id` → 卡状态、预算关联、卡号/CVC、币种、销卡标记、限额和时间等详情字段。 | 已接入 |
| `CardFunding` | `POST /pingpong/api/issuing/v3/cards/funding/actions` | `card_id`、`action`、`amount`、`unique_order_id` → `record_id`、`action`、`amount`、`currency`。 | 已接入 |
| `QueryCardFundingOrders` | `GET /pingpong/api/issuing/v3/card/funding/orders` | query 分页、`start_date`、`end_date`、`status`、`unique_order_id`、`card_id` → `list`、`total_num`、分页字段。注意路径是单数 `card`。 | 已接入 |
| `QueryDedicatedCardBalance` | `GET /pingpong/api/issuing/v3/card/balance` | query `card_id` → `card_number`、`available_balance`、`currency`。 | 已接入 |
| `CardAction` | `POST /pingpong/api/issuing/v3/cards/actions` | `card_id`、`action`、可选 `remark` → SDK 不消费 `data`；仍返回可解析的 JSON 成功 envelope。 | 状态操作已接入；备注更新 501 |
| `QueryCardTransactions` | `GET /pingpong/api/issuing/v3/transactions` | query 卡、类型、状态、清算类型、时间与分页 → `transaction_list`、`total_num`、分页字段；金额字段为字符串。 | 501，待契约确认 |
| `Query3DSDetails` | `GET /pingpong/api/issuing/v3/cards/3ds/details` | query `card_id` → 持卡人身份、联系信息和安全问答；没有配套的持卡人创建方法，需确认实际业务需求。 | 501，待契约确认 |
| `CreateBudgetAccount` | `POST /pingpong/api/issuing/v3/budgets` | 仅 `budget_name` → `budget_id`；不增加 `account_type` 或币种请求字段。 | 已接入 |
| `BudgetFunding` | `POST /pingpong/api/issuing/v3/budgets/funding` | `budget_id`、`action`、`amount`、`currency`、按操作提供订单键/目标预算及币种 → `record_id`。 | 已接入 |
| `QueryBudgetFundingOrder` | `GET /pingpong/api/issuing/v3/funding/orders` | query `order_id`、`action` → `order_id`、`action`、`status`。 | 已接入 |
| `QueryBudgetAccountBalance` | `GET /pingpong/api/issuing/v3/budget/balance` | 可选 query `budget_id`，省略查所属账户全部预算 → `balance_list`。注意路径是单数 `budget`。 | 已接入 |
| `QueryAccountTransactions` | `GET /pingpong/api/issuing/v4/account/transactions` | query 预算/卡、入账区间、类型、方向、分页 → `transaction_list`、`total_num`、分页字段；金额为 JSON number，与 v3 卡交易不同。 | 501，待契约确认 |
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
- `NextCardNumber` 在该产品的所有 BIN 和账户间共享，开卡在同一事务内锁定产品、推进序号并创建独立卡钱包与卡片。
- 保留 `(channel, prefix)` 唯一索引，约束完整配置字符串；它不表示不同产品的候选 BIN 集合互斥。既有单前缀产品和已发行卡片不改写。

## 本轮实现与运行约定

### 应用映射和接入

1. 在浏览器 `/pingpong/accounts` 创建账户，记录返回的账户 ID；按需调整根账户余额。
2. 启动后端时显式配置应用映射，值必须为账户 ID 字符串，例如：

   ```bash
   export PINGPONG_APP_ACCOUNTS='{"demo-app":"4"}'
   ```

3. SDK 基础地址设为 `http://127.0.0.1:8000/pingpong`，通过 `app_id=demo-app` 获取 token。没有映射时返回 `APP_ACCOUNT_NOT_CONFIGURED`，不猜测应用与账户关系。
4. 当前 mock token 是账户主键的十进制字符串；业务头保持 SDK 的 `Authorization: Bearer <token>`。它只是显式账户选择器，不验证密钥、签名或 token 真实性，不适用于真实鉴权。
5. 管理页创建虚拟账户，向虚拟账户充值，再由 SDK OpenAPI 开卡。卡初始余额为零，从所属虚拟账户充值后才能进行本地模拟授权。

映射是本项目的显式 mock 配置，不代表已确认生产凭据映射；更改环境变量后需重启后端。UI 与 OpenAPI 服务/usecase 独立，不互相调用。各自直接注入所需仓储，账户锁、虚拟账户创建/读取、卡状态与资金划转均在各自的领域文件实现；即使当前逻辑相同，也不通过 `dependencies` 或公共业务 helper 复用。两者只复用仓储实现，为后续不同的操作日志和业务规则保留独立演进空间。

### 模型映射与必须先解决的问题

- `Account` 是根账户，`VirtualAccount` 是虚拟账户；卡引用虚拟账户但独立持有钱包，`Card.WalletID` 不与虚拟账户钱包共用。
- 新增中立 `WalletTransfer` 记录卡/虚拟账户间资金划转：账户、渠道、请求幂等键、来源/目标钱包、币种、金额及双方前后余额；卡关联可空，不为虚拟账户资金订单伪造卡。
- `request_id` 控制开卡幂等；`unique_order_id` 控制资金幂等。同账户/渠道相同键及参数返回原记录，参数不同返回 `REQUEST_CONFLICT`。账户行锁串行化同账户写入；产品锁仍在跨账户范围共享发卡序号。
- 虚拟账户 `top_up` 为根账户 → 虚拟账户；`transfer` 为同账户两个虚拟账户间划转。当前只支持 USD，同币种检查明确拒绝未支持币种，不臆造汇率。`transfer` 可以省略幂等键，省略时每次为独立操作。
- 卡 `top_up` / `withdraw` 仅在所属虚拟账户与卡钱包间划转；正常资金操作校验可用余额，预占资金不可转出。失败事务不产生成功订单。当前资金订单同步成功，查询 `FAIL` / `PROCESSING` 返回空列表。
- 当前默认产品为 `424242,424243`、Visa、USD、24 个月、无计费的普通虚拟账户独立卡。创建虚拟账户不接受 SDK 未定义的币种字段，当前默认 USD。这些是已标明的 mock 配置，不是对真实渠道多币种能力的声明。
- SDK 未给出 `card_type` 的值契约，详情保留字段为空并标 `Invalid:`；额度、优惠券、持卡人和备注等未实施字段仅留在 DTO，不扩充通用表。`update_remark` 不返回假成功。
- 详情 `created_at`、订单 `created` 返回 RFC3339，卡有效期按当前 mock 约定返回 `MM/YY`；资金订单起止日期按 UTC `YYYY-MM-DD` 解析，结束日期包含整天。真实非 RFC3339 格式仍需响应样例确认。
- 缺少生产调用证据的账户报表、两类交易报表和 3DS 使用明确的 501；不会以空成功数据隐藏缺失能力。v3/v4 报表不能直接复用同一个 DTO。

### 授权和后续交易的当前实现

- 授权可用余额为卡钱包 `Amount - PendingOut`。金额为正且足够时，在同一事务增加 `PendingOut`、创建 `Authorization` 与 AUTH 阶段记录；没有同步回调或网络审批。
- 清算释放本次覆盖的有效预占并扣减卡钱包全额；允许超额、负余额及继续清算。撤销释放对应预占，退款记入卡钱包，不要求先清算。
- 剩余授权为授权金额减去成功清算与撤销；保留负号，退款不恢复授权额度。各阶段使用独立请求键，重试不重复记账。
- 当前 UI 的清算、撤销、退款从已有授权发起，均验证同账户、同渠道和同卡；独立退款入口尚未加入。
- 本轮仅完成本地授权与账务；尚未创建真实 Webhook 待投递记录。`contract_pending` 是能力状态，不是网络投递状态。拿到契约后需在本地提交时原子写入通知/outbox，再由异步 worker 投递；投递失败与重放不能再次修改账务。

### 管理页面

六个页面为账户、虚拟账户、卡产品、卡管理／模拟授权、授权管理、资金订单。开卡仅通过 OpenAPI；页面保留中文有边框 tag、金额币种同 tag、dayjs datetime、后端账户名称、表格边框及独立分页间距。账户、卡、状态按页面适用条件筛选，未添加同步授权配置或伪造的 Webhook 管理页。

## 后续任务顺序

### 已完成

- [x] 核对 supplied SDK 的 16 个发送方法、DTO、枚举与示例；独立创建 `channel/pingpong/{enums,pkg/idconv,service,biz,data,http}`。
- [x] 显式应用/账户映射、双向十进制 ID、HTTP envelope、独立 UI/OpenAPI 与 scoped repository。
- [x] 虚拟账户、产品、开卡、卡详情/余额/状态、虚拟账户与卡资金操作、订单查询的 12 个核心方法。
- [x] 虚拟账户供资的独立卡钱包、多 BIN 与共享产品序列、注销终态、事务与请求幂等。
- [x] 本地余额检查、授权预占、清算/撤销/关联退款、负数剩余金额；六个独立管理页面。
- [x] 模型生成、gofmt、前后端构建，以及隔离数据库上的本地 HTTP 验证。

### 下一阶段

- [ ] 补齐 Marxo 生产调用点，确认应用映射、USD 默认策略、产品能力、订单 ID 语义、卡类型与真实时间/错误码格式；不能把 SDK 测试当成生产调用证明。
- [ ] 收集 Webhook 事件、payload、签名、ACK 和消费端后，实现原子待投递记录、异步通知、日志、投递管理及重放。失败仅影响投递状态，不重复授权或扣款。
- [ ] 完成根账户余额报表、v3 卡交易与 v4 账户交易；确认金额符号、方向、时间及快照来源，保留两个 DTO 的字段与过滤差异。
- [ ] 根据生产需求确认 3DS、持卡人、限额与备注的持久化范围；补充独立退款入口与更多管理筛选。
- [ ] 单独修正 SDK 的 Bearer 测试断言、Resty 实例复用及无条件 debug；补齐 `tman` 依赖并隔离真实网关样例中的凭据/变更操作。当前只阅读 SDK，没有运行其真实网关测试。
- [ ] 用户明确要求测试后，再接入离线 SDK 契约测试；`sdk/test-contract.sh` 尚不包含 PingPong。

### 本轮验证范围

使用独立临时 PostgreSQL 数据库启动真实 HTTP 服务，验证账户映射、虚拟账户供资、开卡重试、卡充值并发幂等、授权并发幂等、跨账户隔离、分页/非法筛选、注销不可恢复及订单查询。验证卡余额不足时虚拟账户余额无法替代、恰好足够可授权、预占不能转出、超额清算与负余额继续清算。未运行单元测试或 SDK 真实网关测试；Webhook 网络行为与生产联调仍未验收。
