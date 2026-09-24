# PhotonPay

## 范围

已实现持卡人、虚拟账户卡开卡、同步 sandbox 授权/冲正/退款、交易列表、卡状态更新和管理 UI。Webhook 已支持 PhotonPay `ds-event` 的授权、验卡、冲正、退款和卡状态投递及记录；账户级同步授权配置由 `AuthorizationConfig` 管理，自动重试仍待实现。

PhotonPay SDK 当前没有被 Marxo 调用的授权配置 OpenAPI。mock 因此只在管理面提供 `GET/PUT /photonpay/ui/authorization-config`，请求按 `account_id` 选择配置；它持久化同步授权目标 URL、启用状态和超时，不伪造尚无调用点的 PhotonPay OpenAPI 路径。回调连接失败时，模拟授权直接失败。

## 字段映射

| PhotonPay 字段 | 通用表字段 | 处理 |
| --- | --- | --- |
| `cardDetail.cardId`、`cardId` | `Card.ID` | 以十进制字符串格式化输出；入参先解析为内部 ID |
| `cardBin` | `Card.CardBin` 与 `Card.CardProductID` | 通过产品前缀匹配产品；卡保存所选产品和派生 BIN |
| `cardNo`、`cvv`、`expirationDate` | `Card.CardNumber`、`Card.Cvv`、`Card.ExpireAt` | 直接转换；到期日转换为 `time.Time` |
| `cardholderId` | `Card.CardHolderID` | PhotonPay 十进制 ID 解析后关联 `CardHolder` |
| `app_id` / `access_token` | `Account.ID` | `app_id` 是 PhotonPay 十进制账户 ID；空 `secret` 时 token 接口返回相同格式的 `access_token`，后续请求据此确定账户域 |
| `cardCurrency`、`cardScheme`、`cardType`、`cardFormFactor`、`cardStatus` | `Card.CardCurrency`、`Card.CardScheme`、`Card.CardType`、`Card.FormType`、`Card.Status` | 通过 PhotonPay 枚举转换函数处理 |
| `requestId` | `Card.RequestID`、`Card.LastOperationRequestID` | 开卡幂等键及最近操作键 |
| `transactionId` | `CardTransaction.ID` | 十进制字符串格式化输出 |
| `originTransactionId` | `CardTransaction.OriginCardTransactionID` | 退款/冲正时解析并关联原交易 |
| `txnAmount`、`txnCurrency`、`mcc`、商户字段 | `CardTransaction.TxAmount`、`TxCurrency`、`MerchantMCC`、`MerchantName`、`MerchantCountry` | sandbox 交易时写入 |
| `transaction status/type` | `CardTransaction.Status`、`CardTransaction.Type` | 经 PhotonPay 枚举转换；清算设置 `SettledAt` |

`memberId`、`matrixAccount`、卡面、限额、充值金额、收件人、商户城市/邮编、CVV 校验和到期日校验没有通用持久化含义，均为 `Invalid:` 协议字段。`app_id` 不是 `Invalid:`：它选择当前 PhotonPay 账户域。

## 账户解析

所有 PhotonPay 资源按 `(AccountID, Channel=photonpay)` 隔离。先调用 token 接口：`app_id` 由 PhotonPay `channel/photonpay/pkg/idconv` 解析为 `Account.ID`，`secret` 必须为空；成功时 `access_token` 返回同一账户 ID 的十进制格式。后续 OpenAPI 请求从 Marxo SDK 使用的 token 字段解析账户，并将 `AccountID` 显式传到 usecase/repository 请求；不得使用固定 token、固定账户、`context.Context` 隐式范围或通过卡片推导账户。

```bash
curl -X POST http://127.0.0.1:8000/photonpay/oauth2/token/accessToken \
  -H 'Content-Type: application/json' \
  -d '{"app_id":"1001","secret":""}'
# {"access_token":"1001", ...}
```

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

对齐 Marxo 的 `PagingVccTradeOrder` 请求后补齐 SDK 使用的 card ID、交易状态、交易时间范围和分页字段。先用 `channel/photonpay/pkg/idconv` 解析传入卡 ID；随后显式构造 repository 查询，按 `CardTransaction.ID DESC` 排序。每个可选过滤字段均应在 DTO 中声明，不能用 `gin.H` 或 raw GORM predicate 拼接。响应交易 ID 与原交易 ID 再格式化为 PhotonPay 十进制字符串。

验证夹具应覆盖：同一张卡的授权、冲正、退款各一笔；使用 `originTransactionId` 查询退款/冲正来源；时间范围边界；空页。清算交易应返回 `settledAt`，而未清算授权不能冒充已清算。

### Webhook

Marxo 的实际入口是 `POST ${MARXO_BASE_URL}/api/v1/notify/ds-event`。这是 PhotonPay 交易、清算、持卡人状态和卡状态事件的下游契约，不是 generic-mock 的管理 UI webhook 配置接口。

| HTTP header | 值 | 用途 |
| --- | --- | --- |
| `Content-Type` | `application/json` | body 必须按发送字节签名。 |
| `X-PD-SIGN` | 原始 JSON body 的 `MD5withRSA` 签名，再 Base64 | Marxo 使用配置的 PhotonPay 公钥验签。验签开关开启时缺失或不合法会被拒绝。 |
| `X-PD-NOTIFICATION-CATAGORY` | `issuing`、`issuing_settlement` 或 `issuing_card` | 决定 body 的消费分支。值必须小写；注意下游常量拼写是 `CATAGORY`。 |
| `X-PD-NOTIFICATION-TYPE` | `auth`、`verification`、`void`、`refund`、`cardholder_status_update`、`card_status_update` 等 | 交易分支目前由前四种触发；卡/持卡人状态各使用对应类型。 |
| `X-PD-PUBLISHED-AT` | 渠道事件唯一时间/键 | OTP 事件以它作为下游幂等关联键；所有事件建议携带。 |

交易或清算事件使用 `issuing` 或 `issuing_settlement`，并在 header 中给出 `auth`、`verification`、`void` 或 `refund`。下面是可直接用于集成测试的最小完整交易报文；所有卡、交易和原交易 ID 必须经 PhotonPay `channel/photonpay/pkg/idconv` 格式化，不能直接序列化数据库 ID。

```bash
# signature = Base64(RSA_sign_MD5(exact_body_bytes, photonpay_private_key))
curl -X POST "$MARXO_BASE_URL/api/v1/notify/ds-event" \
  -H 'Content-Type: application/json' \
  -H "X-PD-SIGN: $signature" \
  -H 'X-PD-NOTIFICATION-CATAGORY: issuing_settlement' \
  -H 'X-PD-NOTIFICATION-TYPE: auth' \
  -H 'X-PD-PUBLISHED-AT: 2026-09-17T10:01:00Z' \
  -d '{
    "memberId":"mock-member","matrixAccount":"mock-account",
    "createdAt":"2026-09-17T10:01:00Z","updatedAt":"2026-09-17T10:01:00Z",
    "cardId":"201","cardType":"share","transactionId":"501","originTransactionId":"",
    "requestId":"txn-20260917-001","transactionType":"auth","status":"succeed","code":"0","msg":"success",
    "mcc":"5812","authCode":"A12345","transactionAmount":"12.50","transactionCurrency":"USD",
    "txnPrincipalChangeAccount":"card","txnPrincipalChangeAmount":"12.50","txnPrincipalChangeCurrency":"USD",
    "feeDeductionAccount":"card","feeDeductionAmount":"0.00","feeDeductionCurrency":"USD",
    "feeDetailJson":{"transactionFeeAmount":"0.00"},
    "arrivalAccount":"merchant","arrivalAmount":"12.50","merchantNameLocation":"Mock Cafe",
    "merchantLocation":"New York","cardBalance":"87.50","availableTransactionLimit":"87.50",
    "merchantName":"Mock Cafe","transactionStatus":"settled","transactionCountry":"US",
    "transactionHappenedAt":"2026-09-17T10:01:00Z","settleAmount":"12.50","settleCurrency":"USD"
  }'
```

退款/冲正必须带格式化 `originTransactionId`。`feeReturn*` 和 `feeReturnDetailJson` 仅在有退费时提供；`memberId`、`matrixAccount`、资金变动账户和手续费明细仍是 PhotonPay 专属 DTO 字段，不能扩展进通用模型。卡状态事件的 body 至少为 `{"cardId":"201","cardStatus":"frozen"}`，headers 为 `issuing_card` / `card_status_update`；持卡人状态事件提供 `cardholderId`、`status`、`cardholderReviewStatus`、`reason`。

Marxo 不使用通用成功信封，必须收到裸响应 `{"roger":true}`；连续八次不规范响应会使 PhotonPay 停止全部事件通知。因此 mock 的投递判定需要同时检查 HTTP 2xx 和该 JSON body 的 `roger=true`。提交业务 transaction 后先创建 `WebhookRecord`（`Channel=photonpay`、事件、目标、格式化 `SourceID`、原始 body、attempt 1、pending），再发送；将状态码、响应 body、投递时间或错误写回同一 record，重试不得重新生成业务交易。

配置事件直接使用当前已实现的 Marxo `X-PD-NOTIFICATION-TYPE` 值：`auth`、`verification`、`void`、`refund`。卡状态和持卡人状态投递尚未实现，因此不在后端事件列表中返回。在 Marxo 开启验签时设置 `PHOTONPAY_WEBHOOK_PRIVATE_KEY`（PKCS#8 RSA PEM）；未设置时只发送事件 headers 和 body，适用于关闭验签的本地环境。投递记录 UI 已支持分页、详情和 replay：详情保存并展示报文、请求头、响应体和响应头；replay 使用原始报文和请求头再次发送，并新建一条记录保留审计历史。自动重试仍待实现。
