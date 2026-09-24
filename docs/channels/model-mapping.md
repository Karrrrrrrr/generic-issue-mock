# 通用模型映射

## 设计原则

`backend/model/generic.go` 的表结构面向所有渠道，因此字段名不会复刻任何一家下游 API。渠道 service 负责把渠道字段转换为这些通用字段，并用各自的 ID 转换函数（统一位于 `channel/<channel>/pkg/idconv`，公开函数以 `To...` / `From...` 表达转换方向）将内部 `int64 ID` 转为外部资源 ID。

禁止在通用模型中增加 `ThirdPartyID`、`display_id` 或第二套资源 token。渠道没有对应通用含义的 API 字段留在 service DTO，并标记为 `Invalid:`，不持久化。

## 主要表

| 通用表 | 作用 | 主要关联 |
| --- | --- | --- |
| `CardProduct` | 渠道级可开卡 BIN/产品配置，同渠道账户共享 | `Card.CardProductID` |
| `Account` | 渠道账户域；浏览器管理资源的隔离边界 | 账户级业务表的 `AccountID` |
| `CardHolder` | 可复用持卡人资料 | `Card.CardHolderID` |
| `Wallet` | 余额、待入账/待出账和累计入出账 | `Card.WalletID`、`VirtualAccount.WalletID`、`Account.WalletID` |
| `VirtualAccount` | 虚拟账户；可供卡片共享钱包，也可仅为独立卡供资 | `Card.VirtualAccountID` |
| `Card` | 卡片及其渠道无关状态 | BIN、卡号、CVV、到期日、状态、持卡人、钱包、请求 ID |
| `Authorization` | 授权事件 | 卡、金额、商户、授权码、状态、发生时间 |
| `CardTransaction` | 授权、清算、冲正、退款、资金划拨 | 卡、授权、原交易、金额、币种、商户、状态、清算时间 |

## 账户域

`Account` 与 `VirtualAccount` 是不同概念。`Account` 是渠道账户域，保留 `Channel`；Slash 的 `VirtualAccount` 只是该账户域内可共享余额的资源，不能替代 `Account`。

除 `Account` 与渠道级 `CardProduct` 外，每张持久化业务表都必须保存 `AccountID` 与 `Channel`。`AccountID` 引用 `Account.ID`，同一行的 `Channel` 必须与账户的渠道一致。创建卡、持卡人、钱包、授权、交易、Webhook 配置和投递记录等子资源时，在同一事务中从所属聚合继承这两个字段。

账户级仓储的 `Exist`、`Find`、`List`、`Save`、`Delete` 都必须以 `(account_id, channel)` 过滤；每个账户拥有资源的唯一索引也必须包含这两个字段及自然键。渠道外部 ID 始终只由本表 `ID` 经 formatter 编码，不能将账户域或渠道拼入 ID。

账户范围是显式参数，不是 `context.Context` 状态。service 使用渠道 ID 转换函数 解析账户选择器后，在 usecase/repository 请求结构中传递 `AccountID`。账户级单资源请求结构包含必填的 `AccountID` 和资源 `ID`，对应查询必须包含 `id`、`account_id`、`channel`；列表的可选账户过滤使用 `AccountID *model.ID`，仅 `nil` 表示未提供，非 `nil` 时验证 ID 并追加过滤，不得用 `!= 0` 判断是否传入。显式传入无效 ID 应报错，不能退化为跨账户查询。UI 是总后台，因此列表可以不传账户以查看全部数据，但创建和指定账户的变更必须提供有效账户 ID。

仓储请求结构在 biz 中按具体资源和操作单独定义，不复用 service/usecase 请求，也不使用 `ManagementScope`、`ResourceRequest` 或渠道级 `ListRequest` 这类含义模糊的通用请求。可以在各自的请求结构中嵌入语义明确的公共字段结构，但不得用类型别名代替独立请求类型。所有可选参数一律使用指针，以 `nil` 区分未提供与显式零值。

| 渠道 | OpenAPI 账户选择器 | 外部格式 | service 到 biz 的显式字段 |
| --- | --- | --- | --- |
| Paynda | 路径或 body 的 `balanceAccountId` | 十进制字符串 | `AccountID` |
| PhotonPay | `oauth2/token/accessToken` 的 `app_id`，后续请求 `X-PD-TOKEN` | 十进制字符串 | `AccountID` |
| Slash | 请求 `X-API-Key` | Slash 可逆 UUID | `AccountID` |

浏览器账户管理是独立菜单组。选中的账户决定 UI 管理范围；渠道只在实际下游协议要求时，把该域映射为下游账户字段，例如 Paynda 的 `balanceAccountId`。不要将某个渠道的账户路径或字段强加给其他渠道。

## 账户域完成度

| 项目 | 状态 | 说明 |
| --- | --- | --- |
| 创建账户并创建账户钱包 | 已完成 | PhotonPay、Paynda 与 Slash 均在同一事务创建 USD 账户钱包并回写 `Account.WalletID`。 |
| 账户 UI 列表、分页与改名 | 已完成 | PhotonPay、Paynda 与 Slash 均提供分页列表、创建与改名。 |
| OpenAPI 账户范围 | 部分完成 | Slash 和 PhotonPay 的账户资源查询已按账户限定。Paynda 的 `balanceAccountId` 路径资源、账户钱包、持卡人、开卡及单笔交易均已按账户限定；`requestResults` 和 `merchant/wallets` 当前协议 DTO 未提供账户选择器，明确返回 HTTP 501，待确认商户凭据与账户映射后接入，不能臆造请求字段或将 `appId` 解析为资源 ID。 |
| Marxo SDK 与调用点核对 | 受阻 | 当前工作区未提供 Marxo 源码；恢复可访问后逐端点核对 DTO、路径、调用点和错误码。 |
| 账户余额直接调整 | 已实现 | 在账户页面直接增加或扣减账户钱包余额，不要求资金来源。 |
| 账户钱包与卡/虚拟账户双向划转 | 余额操作已实现 | 在资源页面充值/转出；同一事务按钱包 ID 顺序锁定来源与目标，验证账户与币种，更新余额和累计入出账。统一资金流水仍待补齐。 |
| `WebhookConfig` | 渠道 webhook 订阅配置 | 渠道、事件、目标地址、启用状态 |
| `AuthorizationConfig` | 支持同步授权的渠道回调配置 | 账户、渠道、目标地址、启用状态和超时；每个账户和渠道唯一，回调失败即交易失败；PingPong 不使用此配置 |
| `WebhookRecord` | 一次 webhook 投递记录 | 配置、来源资源、请求报文、响应、投递状态和次数 |

## 通用字段语义

| 通用字段 | 语义 |
| --- | --- |
| `BaseModel.ID` | 内部主键；对外必须经渠道 formatter 输出 |
| `Account.ID` / `Account.Channel` | 渠道账户域主键和渠道范围；`Account` 不再额外引用 `AccountID` |
| 除 `Account`、`CardProduct` 外业务表的 `AccountID` / `Channel` | 资源所属账户域和渠道；仓储操作与唯一索引必须包含两列 |
| `CardProduct.Prefix` | 英文逗号分隔的数字 BIN 前缀字符串；只有未来 PingPong 支持多前缀，其他渠道仅单前缀 |
| `Card.CardProductID` / `Card.CardBin` | 产品主键及开卡时从产品候选项中选中的单个 BIN；卡号使用同一个前缀，不保存整串候选列表 |
| `Card.WalletID` | 实际余额和消费钱包；`share` 指向 VA 钱包，`single` 与 `virtual_account_single` 指向独立卡钱包 |
| `Card.VirtualAccountID` | VA 关联，不等于共享余额：`share` 与 `virtual_account_single` 均非空，`single` 为空 |
| `Card.CardType` | `single` 普通独立卡；`share` VA 共享卡；`virtual_account_single` VA 供资的独立卡 |
| `Card.RequestID` | 开卡请求的商户幂等键 |
| `Card.LastOperation*` | 最近一次开卡/冻结/更新/销卡操作的幂等键、类型和结果 |
| `CardTransaction.AuthorizationID` | 清算和撤销必须关联同账户同卡的有效授权；退款可为 `0`（独立退款），也可关联授权，无须先清算 |
| `CardTransaction.OriginCardTransactionID` | 从原交易发起操作时记录的原交易；独立退款或直接关联授权的退款不要求该字段 |
| `CardTransaction.TxAmount` / `TxCurrency` | 原始交易金额和币种；`Currency` 是结算/卡币种语义 |
| `CardTransaction.RawPayload` / `Authorization.RawPayload` | 保留渠道原始事件或计算所需原始数据，不能代替字段化的中立业务概念 |

## 资金流规则

三种卡的关联和资金流：

| 卡类型 | VA 关联 | 消费钱包 | 普通充值/转出对手方 |
| --- | --- | --- | --- |
| `single` | 无 | 独立卡钱包 | 所属根账户钱包 |
| `share` | 有 | 同一个 VA 钱包 | 所属根账户钱包 |
| `virtual_account_single` | 有 | 独立卡钱包，与 VA 钱包不同 | 所关联 VA 钱包 |

PingPong 的预算组映射为 VA，其普通卡使用第三种类型。初始化第三类卡只分配独立零余额钱包，不把预算余额复制到卡上；充值时才从预算钱包转移资金，转出回到同一个预算钱包。业务通过 `pkg/cardwallet.Prepare` 分配钱包、在同一事务持久化，通过 `FundingWalletID` 解析资金来源。模拟授权/清算只用 `Card.WalletID`，不会因存在 VA 关联而再次扣预算钱包。

- 创建账户必须在同一事务创建 `WalletType_Account` 的 USD 钱包，并将其 ID 回写到 `Account.WalletID`。
- 账户余额调整是管理端的直接余额变更，可凭空增加或减少余额，不要求资金来源。
- `single` 卡与虚拟账户充值从所属根账户钱包转出；`virtual_account_single` 卡只能从所关联 VA 钱包供资。所有普通资金划转在同一事务锁定来源与目标钱包，来源余额不足时拒绝操作。
- 充值不是直接修改目标余额：应同时更新两个钱包的 `Amount`、`In`/`Out`，并按已实现渠道的规则创建资金变动交易。

除 PingPong 授权外，模拟授权不校验余额是否充足；各渠道模拟清算允许超过授权金额，钱包余额和剩余授权金额均可为负数，并允许继续清算。每次输入金额仍须为正数，展示剩余额度时不得截断为零。普通账户、卡、虚拟账户资金划转不适用这一例外。

PingPong 的页面模拟授权必须检查卡自身可用余额，足够后由本地授权并异步推送结果；不等待下游批准，也不使用 `AuthorizationConfig`。余额不足不能靠关联预算或根账户余额通过。Webhook 投递状态与授权业务状态分离：超时/失败不撤销本地授权，重放不得重复授权或记账。此规则仅约束 PingPong 授权，不收紧清算规则；具体 Webhook 报文仍待确认。

账户关联资源的 UI DTO 在原接口直接返回 `account_id`、`account_name`。账户名称通过只读账户关联查询获得，不在业务记录中冗余存储；前端不再为了显示名称单独拉取账户列表。

已实现渠道的字段映射见 PhotonPay、Paynda、Slash 文档；UQPay、PingPong 为未来渠道。PingPong SDK 已阅读，用户已确认预算组为 VA、卡片为独立钱包的第三类卡；通用枚举和钱包分配能力已落地，不新增冗余预算 ID 或资金来源列。其余[模型映射与后续任务](pingpong.md#模型映射与必须先解决的问题)仍区分预算与根账户、资金订单和两类交易报表，生产调用与部分协议语义待确认。Payful 已废弃，其文档仅作历史归档。

`CardProduct` 是渠道级配置，不包含 `AccountID`；产品查询按 `Channel` 隔离，唯一索引为 `(channel, prefix)`。同渠道账户共享产品及发卡序列，创建账户不再复制产品；卡片仍按账户隔离，且只能引用同渠道产品。

`Prefix` 保持字符串列，不新增 BIN 表或卡片候选列表字段。开卡统一使用 `pkg/cardnumber.Generate`：按英文逗号拆分并去除候选项两端空白，校验所有项均为非空数字且能容纳当前序号，只有 `pingpong` 可随机选择多个候选之一，其他渠道对多项配置直接报错。产品级 `NextCardNumber` 不按 BIN 拆分，仍在原事务中锁定、递增、保存；现有单前缀与已发卡数据不变。
