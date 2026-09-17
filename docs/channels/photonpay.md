# PhotonPay

## 范围

已实现持卡人、虚拟账户卡开卡、同步 sandbox 授权/冲正/退款、交易列表、卡状态更新和管理 UI。Webhook 目前只支持配置管理，不会实际投递。

## 字段映射

| PhotonPay 字段 | 通用表字段 | 处理 |
| --- | --- | --- |
| `cardDetail.cardId`、`cardId` | `Card.ID` | 以十进制字符串格式化输出；入参先解析为内部 ID |
| `cardBin` | `Card.CardBin` 与 `Card.CardProductID` | 通过产品前缀匹配产品；卡保存所选产品和派生 BIN |
| `cardNo`、`cvv`、`expirationDate` | `Card.CardNumber`、`Card.Cvv`、`Card.ExpireAt` | 直接转换；到期日转换为 `time.Time` |
| `cardholderId` | `Card.CardHolderID` | PhotonPay 十进制 ID 解析后关联 `CardHolder` |
| `cardCurrency`、`cardScheme`、`cardType`、`cardFormFactor`、`cardStatus` | `Card.CardCurrency`、`Card.CardScheme`、`Card.CardType`、`Card.FormType`、`Card.Status` | 通过 PhotonPay 枚举转换函数处理 |
| `requestId` | `Card.RequestID`、`Card.LastOperationRequestID` | 开卡幂等键及最近操作键 |
| `transactionId` | `CardTransaction.ID` | 十进制字符串格式化输出 |
| `originTransactionId` | `CardTransaction.OriginCardTransactionID` | 退款/冲正时解析并关联原交易 |
| `txnAmount`、`txnCurrency`、`mcc`、商户字段 | `CardTransaction.TxAmount`、`TxCurrency`、`MerchantMCC`、`MerchantName`、`MerchantCountry` | sandbox 交易时写入 |
| `transaction status/type` | `CardTransaction.Status`、`CardTransaction.Type` | 经 PhotonPay 枚举转换；清算设置 `SettledAt` |

`memberId`、`matrixAccount`、卡面、限额、充值金额、收件人、商户城市/邮编、CVV 校验和到期日校验没有通用持久化含义，均为 `Invalid:` 协议字段。

## Marxo 调用基准

| 业务 | SDK 调用 | HTTP 协议 |
| --- | --- | --- |
| 开卡 | `OpenCard` | `POST /vcc/openApi/v4/openCard` |
| 查询开卡结果 | `GetRequestResult` | `GET /vcc/openApi/v4/getRequestResult` |
| 清算查询 | `PagingVccTradeOrder` | `GET /vcc/openApi/v4/pagingVccTradeOrder` |
| 同步交易模拟 | `SandBoxTransaction` | `POST /vcc/open/v2/sandBoxTransaction` |

SDK 位于 Marxo 的 `pkg/dealer/photonpay/photonpay.go`，实际开卡和清算调用由相关 card service usecase 发起。

## Mock 报文示例

先通过 `GET /photonpay/vcc/openApi/v4/getCardBin` 和 `POST /photonpay/vcc/openApi/v4/addCardholder` 获得卡 BIN 与持卡人 ID。PhotonPay mock 自有资源 ID 是十进制字符串。

```bash
curl -X POST http://127.0.0.1:8000/photonpay/vcc/openApi/v4/openCard \
  -H 'Content-Type: application/json' \
  -d '{
    "cardBin":"486699",
    "cardCurrency":"USD",
    "cardExpirationDate":24,
    "cardScheme":"MasterCard",
    "cardType":"share",
    "cardFormFactor":"virtual_card",
    "cardholderId":"101",
    "requestId":"open-card-20260917-001"
  }'
```

成功响应使用 PhotonPay 信封，后续使用 `data.cardDetail.cardId`：

```json
{
  "code": "0",
  "msg": "success",
  "data": {
    "requestId": "open-card-20260917-001",
    "status": "succeed",
    "cardDetail": { "cardId": "201", "cardNo": "486699...", "cvv": "123" }
  }
}
```

清算任务通过交易列表轮询。当前 mock 仅支持分页，尚未实现 SDK 中的卡 ID、交易状态和时间范围过滤。

```bash
curl 'http://127.0.0.1:8000/photonpay/vcc/openApi/v4/pagingVccTradeOrder?pageIndex=1&pageSize=50'
```

同步交易模拟的 `txnType` 为 `auth`、`void` 或 `refund`；退款/冲正使用原交易 ID。

```bash
curl -X POST http://127.0.0.1:8000/photonpay/vcc/open/v2/sandBoxTransaction \
  -H 'Content-Type: application/json' \
  -d '{
    "requestId":"txn-20260917-001",
    "cardID":"201",
    "cvv":"123",
    "expirationDate":"09/28",
    "originTransactionId":"",
    "txnCurrency":"USD",
    "txnAmount":12.50,
    "txnType":"auth",
    "mcc":"5812",
    "merchantName":"Mock Cafe",
    "merchantCountry":"US",
    "merchantCity":"New York",
    "merchantPostcode":"10001"
  }'
```

## 待办

### 交易列表过滤

对齐 Marxo 的 `PagingVccTradeOrder` 请求后补齐 SDK 使用的 card ID、交易状态、交易时间范围和分页字段。先用 `service/id.go` 解析传入卡 ID；随后显式构造 repository 查询，按 `CardTransaction.ID DESC` 排序。每个可选过滤字段均应在 DTO 中声明，不能用 `gin.H` 或 raw GORM predicate 拼接。响应交易 ID 与原交易 ID 再格式化为 PhotonPay 十进制字符串。

验证夹具应覆盖：同一张卡的授权、冲正、退款各一笔；使用 `originTransactionId` 查询退款/冲正来源；时间范围边界；空页。清算交易应返回 `settledAt`，而未清算授权不能冒充已清算。

### Webhook

已检查的 Marxo PhotonPay SDK 没有入站 webhook DTO/消费者，不能伪造 PhotonPay 特有字段或签名。待下游协议确认后，以 `transaction.created`、`transaction.updated`、`authorization.created`、`authorization.updated` 这四个已在管理 UI 配置的事件为边界实现：payload 必须含格式化的卡/交易/授权 ID、事件时间、状态、金额和币种；授权事件再含商户/MCC 与授权码。

每次业务状态提交后，查询启用且事件匹配的 `WebhookConfig`，先创建 `WebhookRecord`（`Channel=photonpay`、`Event`、`TargetURL`、格式化 `SourceID`、payload、`AttemptCount=1`、pending），再在 transaction 外 POST。2xx 写成功、HTTP 状态/响应体/时间；非 2xx 或网络错误写失败和错误信息。重试更新同一记录；协议确定后在本节补上真实 header、签名和完整 JSON 示例。
