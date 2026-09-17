# 通用模型映射

## 设计原则

`backend/model/generic.go` 的表结构面向所有渠道，因此字段名不会复刻任何一家下游 API。渠道 service 负责把渠道字段转换为这些通用字段，并用各自 `service/id.go` 将内部 `int64 ID` 转为外部资源 ID。

禁止在通用模型中增加 `ThirdPartyID`、`display_id` 或第二套资源 token。渠道没有对应通用含义的 API 字段留在 service DTO，并标记为 `Invalid:`，不持久化。

## 主要表

| 通用表 | 作用 | 主要关联 |
| --- | --- | --- |
| `CardProduct` | 可开卡 BIN/产品配置 | `Card.CardProductID` |
| `CardHolder` | 可复用持卡人资料 | `Card.CardHolderID` |
| `Wallet` | 余额、待入账/待出账和累计入出账 | `Card.WalletID`、`VirtualAccount.WalletID`、`Account.WalletID` |
| `VirtualAccount` | 可共享余额的账户 | `Card.VirtualAccountID` |
| `Card` | 卡片及其渠道无关状态 | BIN、卡号、CVV、到期日、状态、持卡人、钱包、请求 ID |
| `Authorization` | 授权事件 | 卡、金额、商户、授权码、状态、发生时间 |
| `CardTransaction` | 授权、清算、冲正、退款、资金划拨 | 卡、授权、原交易、金额、币种、商户、状态、清算时间 |
| `WebhookConfig` | 渠道 webhook 订阅配置 | 渠道、事件、目标地址、启用状态 |
| `WebhookRecord` | 一次 webhook 投递记录 | 配置、来源资源、请求报文、响应、投递状态和次数 |

## 通用字段语义

| 通用字段 | 语义 |
| --- | --- |
| `BaseModel.ID` | 内部主键；对外必须经渠道 formatter 输出 |
| `Card.CardProductID` / `Card.CardBin` | 产品主键及开卡时从产品派生的 BIN 前缀 |
| `Card.WalletID` | 卡余额钱包；共享卡可指向虚拟账户钱包 |
| `Card.VirtualAccountID` | 非空表示共享余额卡；为空表示独立卡 |
| `Card.RequestID` | 开卡请求的商户幂等键 |
| `Card.LastOperation*` | 最近一次开卡/冻结/更新/销卡操作的幂等键、类型和结果 |
| `CardTransaction.AuthorizationID` | 关联的授权记录；没有授权上下文时为零值 |
| `CardTransaction.OriginCardTransactionID` | 退款或冲正所关联的原交易 |
| `CardTransaction.TxAmount` / `TxCurrency` | 原始交易金额和币种；`Currency` 是结算/卡币种语义 |
| `CardTransaction.RawPayload` / `Authorization.RawPayload` | 保留渠道原始事件或计算所需原始数据，不能代替字段化的中立业务概念 |

下列渠道文档说明具体 API 字段如何映射到该结构：PhotonPay、Paynda、Slash、Payful、UQPay。
