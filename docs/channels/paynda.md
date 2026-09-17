# Paynda

## 范围

已实现持卡人、独立余额卡开卡、余额划拨、交易查询、管理端授权/清算/冲正/退款模拟和管理 UI。Paynda 不支持同步授权；已支持 `xm-event` 交易投递和投递记录，自动重试与记录查询 UI 尚未实现。

## 字段映射

| Paynda 字段 | 通用表字段 | 处理 |
| --- | --- | --- |
| `cardholder.id`、`cardholderId` | `CardHolder.ID`、`Card.CardHolderID` | 十进制字符串格式化/解析 |
| 路径/请求中的 `balanceAccountId` | `Account.ID` | Paynda 十进制账户 ID；中间件解析并确定当前账户域 |
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

## 账户解析

`balanceAccountId` 是 Paynda 当前账户域，而非可忽略的路径占位符。中间件使用 Paynda 十进制 ID helper 解析它，校验 `Account.Channel=paynda`，并将 `(AccountID, Channel)` 放入请求上下文；该域用于过滤全部账户资源。路径外的 `balanceAccountId`（例如账户钱包转账 body）也必须用同一解析规则校验，不能选择默认账户。

## Mock 报文示例

Paynda mock 自有资源 ID 是十进制字符串。请求 ID 使用 `requestId` header 传递；先查询 card BIN 和创建持卡人取得相关 ID。

```bash
curl -X POST http://127.0.0.1:8000/paynda/openapi/balanceAccounts/1001/cards \
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
curl 'http://127.0.0.1:8000/paynda/openapi/balanceAccounts/1001/transactions?cardId=401&current=1&pageSize=50'
curl 'http://127.0.0.1:8000/paynda/openapi/balanceAccounts/1001/transactions/501'
```

余额入金/出金是清算相关资金操作，`type` 只能使用 `IN` 或 `OUT`：

```bash
curl -X POST http://127.0.0.1:8000/paynda/openapi/balanceAccounts/1001/cardBalanceTransfers \
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

Marxo 的实际入口是 `POST ${MARXO_BASE_URL}/api/v1/notify/xm-event`。验签 headers 和事件分类均已在 Marxo middleware 中固定：

| HTTP header | 要求 |
| --- | --- |
| `Content-Type` | `application/json` |
| `appId` | Marxo 配置的 Paynda app ID |
| `timestamp` | Unix 秒时间戳；与 Marxo 当前时间相差不得超过 120 秒 |
| `nonce` | 每次投递的新随机字符串 |
| `sign` | 小写 MD5：`md5(appId + appSecret + timestamp + nonce + targetURL.path)` |
| `X-VK-NOTIFICATION-CATEGORY` | `CARD_TRANSACTION` 或 `CARD_STATUS` |

签名不包含 JSON body。mock 从 `targetURL` 解析实际投递请求的 path 参与签名；验签开启时，接收方必须对同一路径、`appId`、`appSecret`、秒级时间戳和 nonce 进行拼接校验。不要使用 `X-Signature` 或包含 body 的 HMAC。

交易事件的 body 外层必须是 `cardTransactionWebhook`。`cardId`、`cardholderId`、`balanceAccountId`、`merchantId` 在下游 DTO 中是整数；mock 必须输出 Paynda service formatter 得到的可逆数值，不能直接泄露内部自增 ID。其余没有通用语义的收单/POS字段保留在 Paynda service DTO。

```bash
timestamp=$(date +%s)
nonce='mock-nonce-20260917-001'
# sign = md5("$appId$appSecret$timestamp$nonce${targetURLPath}")
curl -X POST "$MARXO_BASE_URL/api/v1/notify/xm-event" \
  -H 'Content-Type: application/json' \
  -H "appId: $appId" -H "timestamp: $timestamp" -H "nonce: $nonce" -H "sign: $sign" \
  -H 'X-VK-NOTIFICATION-CATEGORY: CARD_TRANSACTION' \
  -d '{
    "cardTransactionWebhook": {
      "id":"ptx-501","createTime":"2026-09-17T10:01:00Z","updateTime":"2026-09-17T10:02:00Z",
      "merchantId":9001,"balanceAccountId":7001,"cardholderId":101,"cardId":401,
      "maskCardNo":"486699******0001","type":"AUTH","approvalCode":"A12345",
      "preAuthAmount":"12.50","postedAmount":"12.50","currency":"USD",
      "originalCurrencyCode":"USD","transactionAmountInOriginalCurrency":"12.50","reversalFlag":"false",
      "transactionTime":"2026-09-17T10:01:00Z","authorizationTime":"2026-09-17T10:01:00Z",
      "acquirerId":"acquirer-1","merchantMcc":"5812","merchantName":"Mock Cafe",
      "merchantAddressAddressLine1":"1 Main St","merchantAddressCity":"New York",
      "merchantAddressState":"NY","merchantAddressCountry":"US","merchantAddressZip":"10001",
      "posAcceptorId":"terminal-1","posAcceptLocation":"New York","posEntryDescription":"chip",
      "transactionId":"txn-501","supplierTransactionId":"supplier-501",
      "supplierTransactionLinkId":"auth-501","declineMessage":"","walletId":"wallet-701"
    }
  }'
```

卡状态 body 则是 `{"cardStatusWebhook":{"id":"pcs-401","createTime":"...","updateTime":"...","merchantId":9001,"balanceAccountId":7001,"cardholderId":101,"cardId":401,"maskCardNo":"486699******0001","status":"ACTIVE"}}`，header 分类为 `CARD_STATUS`。Marxo 当前只对 `ACTIVE` 和 `FROZEN` 收敛卡状态；其它状态可以记录但不会触发状态更新。

该入口返回空成功响应，投递成功以 HTTP 2xx 为准。提交业务 transaction 后，先建立 `WebhookRecord`（`Channel=paynda`、分类事件、target、格式化 `SourceID`、完整 JSON、attempt 1、pending），再 POST；记录 HTTP 状态、响应、时间或失败原因，并在重试时更新同一 record。交易 body 的 `id`/`transactionId` 应稳定复用，避免 Marxo 将重投视作新清算消息。

配置事件直接使用当前已实现的 Marxo `X-VK-NOTIFICATION-CATEGORY` 值：`CARD_TRANSACTION`。卡状态投递尚未实现，因此 `CARD_STATUS` 不在后端事件列表中返回。当前 UI 模拟授权、退款、清算/冲正均投递 `CARD_TRANSACTION`。若 Marxo 开启验签，设置 `PAYNDA_WEBHOOK_APP_ID` 与 `PAYNDA_WEBHOOK_APP_SECRET`；未设置时仍发送分类 header 与 body，适用于关闭验签的本地环境。投递记录已经持久化，自动重试和记录查询 UI 仍待实现。
