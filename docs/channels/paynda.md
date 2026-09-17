# Paynda

## 范围

已实现持卡人、独立余额卡开卡、余额划拨、交易查询、管理端授权/清算/冲正/退款模拟和管理 UI。Paynda 不支持同步授权；当前尚未实现事件 webhook 投递。

## 字段映射

| Paynda 字段 | 通用表字段 | 处理 |
| --- | --- | --- |
| `cardholder.id`、`cardholderId` | `CardHolder.ID`、`Card.CardHolderID` | 十进制字符串格式化/解析 |
| 姓名、邮箱、手机和 `billingAddressLine1`/城市/国家/邮编/州 | `CardHolder.FirstName`、`LastName`、`Email`、`Mobile*`、`Residential*` | 账单地址映射到渠道中立的居住/账单地址字段 |
| `card.id`、`cardId` | `Card.ID` | 十进制字符串格式化/解析 |
| `cardBinId` | `Card.CardProductID` | 解析产品 ID；开卡时从 `CardProduct.Prefix` 派生 `Card.CardBin` |
| `maskCardNo`、敏感信息中的 `cardNo`/`cvv`/`expirationDate` | `Card.CardNumber`、`Card.Cvv`、`Card.ExpireAt` | 读取时按接口需要掩码或原样返回 |
| `status`、`currency`、`cardScheme` | `Card.Status`、`Card.CardCurrency`、`Card.CardScheme` | 枚举转换 |
| `requestId` header | `Card.RequestID`、`Card.LastOperationRequestID` | 开卡、状态变更和资金操作的幂等关联键 |
| 卡余额 `amount`、`availableAmount`、`amountFrozen` | `Wallet.Amount`、`Wallet.PendingOut` | Paynda 普通卡的 `Card.WalletID` 指向独立钱包 |
| `transaction.id`、`cardId`、`type`、金额、授权码和商户字段 | `CardTransaction.ID`、`CardID`、`Type`、`TxAmount`、`AuthorizationCode`、`Merchant*` | 交易查询和 UI 模拟的映射 |
| `IN` / `OUT` 资金操作 | `CardTransaction.Type` | 分别映射为 `fund_in` / `fund_out`；操作前余额保留在 `RawPayload` |

`billingAddressLine2`、`unlimitedBalance`、`creditLimitType`、开卡 `amount`、`singleUse`、交易次数限制，以及交易列表中的重复持卡人 ID没有中立持久化含义，均为 `Invalid:` 字段。

## Marxo 调用基准

| 业务 | SDK 调用 | HTTP 协议 |
| --- | --- | --- |
| 开卡 | `CreateCard` | `POST /openapi/balanceAccounts/{balanceAccountId}/cards` |
| 单笔交易查询 | `QueryCardTransaction` | `GET /openapi/balanceAccounts/{balanceAccountId}/transactions/{id}` |
| 交易列表 | 交易列表查询 | `GET /openapi/balanceAccounts/{balanceAccountId}/transactions` |
| 请求结果 | `QueryRequestResult` | `GET /openapi/requestResults` |

SDK 位于 Marxo 的 `pkg/dealer/payndapay/payndapay.go`；交易查询的实际调用位于 card service 的 Paynda data repository。

## Mock 报文示例

Paynda mock 自有资源 ID 是十进制字符串。请求 ID 使用 `requestId` header 传递；先查询 card BIN 和创建持卡人取得相关 ID。

```bash
curl -X POST http://127.0.0.1:8000/paynda/openapi/balanceAccounts/mock-account/cards \
  -H 'Content-Type: application/json' \
  -H 'requestId: open-card-20260917-002' \
  -d '{
    "cardholderId":"101",
    "currency":"USD",
    "amount":"0",
    "expirationDate":"09/28",
    "cardBinId":"301",
    "singleUse":false,
    "transactionCountLimitToTotal":0
  }'
```

成功响应使用 `{code,message,data,success}` 信封，`data.id` 是卡 ID。Marxo 清算流程通过交易列表或单笔交易读取状态：

```bash
curl 'http://127.0.0.1:8000/paynda/openapi/balanceAccounts/mock-account/transactions?cardId=401&current=1&pageSize=50'
curl 'http://127.0.0.1:8000/paynda/openapi/balanceAccounts/mock-account/transactions/501'
```

余额入金/出金是清算相关资金操作，`type` 只能使用 `IN` 或 `OUT`：

```bash
curl -X POST http://127.0.0.1:8000/paynda/openapi/balanceAccounts/mock-account/cardBalanceTransfers \
  -H 'Content-Type: application/json' \
  -H 'requestId: balance-20260917-001' \
  -d '{"cardId":"401","amount":"20.00","type":"IN"}'
```

异步请求结果：

```bash
curl 'http://127.0.0.1:8000/paynda/openapi/requestResults?requestId=open-card-20260917-002'
```

## 待办

### Webhook

Marxo 已使用 Paynda 的开卡、请求结果和交易查询 SDK，但已检查的 SDK 中没有 Paynda 入站 webhook DTO 或事件枚举。因此不得编造 `X-Signature`、回调 URL 或厂商事件名。待确认下游协议后，至少需覆盖管理 UI 已允许订阅的 `transaction.created`、`transaction.updated`、`authorization.created`、`authorization.updated`。

建议的渠道 service payload 边界如下：交易事件包含格式化 `cardId`、`transactionId`、可选 `originTransactionId`、`type`、`status`、`amount`、`currency`、`authorizationCode`、商户字段和发生/清算时间；授权事件包含格式化 `cardId`、`authorizationId`、状态、金额/币种、MCC、商户及时间。字段由 `CardTransaction`/`Authorization` 转换，不能从 Paynda 字符串枚举直接泄露通用枚举。

投递流程必须是：业务 transaction 提交后，按渠道/事件读取启用的 `WebhookConfig`；每个目标先插入一个 `WebhookRecord`（channel、event、target URL、格式化 source ID、完整 JSON、attempt 1、pending），再发 HTTP POST；只有 2xx 标记成功并保存状态码、响应和时间；失败保存错误和响应并更新同一 record 重试。实现者需要把最终确认的 Paynda URL、header、签名算法、重试上限和真实示例报文补到本节。
