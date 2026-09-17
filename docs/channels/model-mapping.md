# 通用模型映射

## 设计原则

`backend/model/generic.go` 的表结构面向所有渠道，因此字段名不会复刻任何一家下游 API。渠道 service 负责把渠道字段转换为这些通用字段，并用各自 `service/id.go` 将内部 `int64 ID` 转为外部资源 ID。

禁止在通用模型中增加 `ThirdPartyID`、`display_id` 或第二套资源 token。渠道没有对应通用含义的 API 字段留在 service DTO，并标记为 `Invalid:`，不持久化。

## 主要表

| 通用表 | 作用 | 主要关联 |
| --- | --- | --- |
| `CardProduct` | 可开卡 BIN/产品配置 | `Card.CardProductID` |
| `Account` | 渠道账户域；浏览器管理资源的隔离边界 | 所有业务表的 `AccountID` |
| `CardHolder` | 可复用持卡人资料 | `Card.CardHolderID` |
| `Wallet` | 余额、待入账/待出账和累计入出账 | `Card.WalletID`、`VirtualAccount.WalletID`、`Account.WalletID` |
| `VirtualAccount` | 可共享余额的账户 | `Card.VirtualAccountID` |
| `Card` | 卡片及其渠道无关状态 | BIN、卡号、CVV、到期日、状态、持卡人、钱包、请求 ID |
| `Authorization` | 授权事件 | 卡、金额、商户、授权码、状态、发生时间 |
| `CardTransaction` | 授权、清算、冲正、退款、资金划拨 | 卡、授权、原交易、金额、币种、商户、状态、清算时间 |

## 账户域

`Account` 与 `VirtualAccount` 是不同概念。`Account` 是渠道账户域，保留 `Channel`；Slash 的 `VirtualAccount` 只是该账户域内可共享余额的资源，不能替代 `Account`。

除 `Account` 自身外，每张持久化业务表都必须保存 `AccountID` 与 `Channel`。`AccountID` 引用 `Account.ID`，同一行的 `Channel` 必须与账户的渠道一致。创建卡、持卡人、钱包、授权、交易、Webhook 配置和投递记录等子资源时，在同一事务中从所属聚合继承这两个字段。

所有仓储的 `Exist`、`Find`、`List`、`Save`、`Delete` 都必须以 `(account_id, channel)` 过滤；每个账户拥有资源的唯一索引也必须包含这两个字段及自然键。渠道外部 ID 始终只由本表 `ID` 经 formatter 编码，不能将账户域或渠道拼入 ID。

账户范围是显式参数，不是 `context.Context` 状态。service 使用渠道 `service/id.go` 解析账户选择器后，在 usecase/repository 请求结构中传递 `AccountID`。单资源请求结构包含 `AccountID` 和资源 `ID`，对应查询必须包含 `id`、`account_id`、`channel`；列表请求包含可选 `AccountID`，非零时直接追加过滤。UI 是总后台，因此列表可以不传账户以查看全部数据，但创建和指定账户的变更不得传 `0`。

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
| OpenAPI 账户范围 | 部分完成 | Slash 和 PhotonPay 的账户资源查询已按账户限定。Paynda 的 `balanceAccountId` 路径资源、账户钱包、持卡人、开卡及单笔交易均已按账户限定；`requestResults` 和 `merchant/wallets` 当前协议 DTO 未提供账户选择器，待 Marxo 调用点核对后处理，不能臆造请求字段。 |
| Marxo SDK 与调用点核对 | 受阻 | 当前工作区未提供 Marxo 源码；恢复可访问后逐端点核对 DTO、路径、调用点和错误码。 |
| 账户余额直接调整 | 待实现 | 管理端可直接变更账户钱包余额，不要求资金来源。 |
| 账户钱包充值普通卡与虚拟账户 | 待实现 | 事务中锁定来源与目标钱包，更新余额和累计入出账，并创建资金交易。 |
| `WebhookConfig` | 渠道 webhook 订阅配置 | 渠道、事件、目标地址、启用状态 |
| `WebhookRecord` | 一次 webhook 投递记录 | 配置、来源资源、请求报文、响应、投递状态和次数 |

## 通用字段语义

| 通用字段 | 语义 |
| --- | --- |
| `BaseModel.ID` | 内部主键；对外必须经渠道 formatter 输出 |
| `Account.ID` / `Account.Channel` | 渠道账户域主键和渠道范围；`Account` 不再额外引用 `AccountID` |
| 所有非 `Account` 业务表的 `AccountID` / `Channel` | 资源所属账户域和渠道；仓储操作与唯一索引必须包含两列 |
| `Card.CardProductID` / `Card.CardBin` | 产品主键及开卡时从产品派生的 BIN 前缀 |
| `Card.WalletID` | 卡余额钱包；共享卡可指向虚拟账户钱包 |

## 资金流待实现

- 创建账户必须在同一事务创建 `WalletType_Account` 的 USD 钱包，并将其 ID 回写到 `Account.WalletID`。
- 账户余额调整是管理端的直接余额变更，可凭空增加或减少余额，不要求资金来源。
- 普通卡和虚拟账户充值均从所属账户的钱包转出，并在同一事务锁定来源与目标钱包；来源余额不足时拒绝操作。
- 充值不是直接修改目标余额：应同时更新两个钱包的 `Amount`、`In`/`Out`，并按已实现渠道的规则创建资金变动交易。
| `Card.VirtualAccountID` | 非空表示共享余额卡；为空表示独立卡 |
| `Card.RequestID` | 开卡请求的商户幂等键 |
| `Card.LastOperation*` | 最近一次开卡/冻结/更新/销卡操作的幂等键、类型和结果 |
| `CardTransaction.AuthorizationID` | 关联的授权记录；没有授权上下文时为零值 |
| `CardTransaction.OriginCardTransactionID` | 退款或冲正所关联的原交易 |
| `CardTransaction.TxAmount` / `TxCurrency` | 原始交易金额和币种；`Currency` 是结算/卡币种语义 |
| `CardTransaction.RawPayload` / `Authorization.RawPayload` | 保留渠道原始事件或计算所需原始数据，不能代替字段化的中立业务概念 |

下列渠道文档说明具体 API 字段如何映射到该结构：PhotonPay、Paynda、Slash、Payful、UQPay。
