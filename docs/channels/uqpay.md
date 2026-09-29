# UQPay

## 状态

UQPay 当前暂不实现，不列入近期 backlog。以下内容仅保留此前调研和可能的协议边界，不能作为当前开发要求；后续如重新启用，仍必须以 Marxo 实际调用点为边界重新核对。

## 目标字段映射

下表是实施边界，具体 DTO 字段以 Marxo 实际请求构造为准。

| UQPay 概念 | 目标通用表字段 | 处理 |
| --- | --- | --- |
| UQPay 卡 ID | `Card.ID` | 新增 UQPay formatter/parser；不持久化下游资源 ID |
| 卡产品和 BIN | `CardProduct.ID`、`Card.CardProductID`、`Card.CardBin` | 产品配置管理 BIN，开卡时派生前缀 |
| 虚拟/实体卡 | `Card.FormType`、`VirtualCard` / `PhysicalCard` | 使用现有表区分卡介质 |
| 共享卡/虚拟账户 | `Card.VirtualAccountID`、`Card.WalletID`、`VirtualAccount` | 共享卡指向虚拟账户的钱包 |
| 持卡人 ID 和资料 | `CardHolder.ID`、`Card.CardHolderID`、`CardHolder` | 仅存渠道中立资料字段 |
| 卡状态 | `Card.Status` | UQPay 枚举转换 |
| 卡交易和同步授权 | `Authorization`、`CardTransaction` | 交易 ID、原交易、金额、币种、商户、状态和时间分别映射；授权回调关联 `AuthorizationID` |
| 余额退回/资金调整 | `Wallet`、`CardTransaction` | 以钱包余额和 `fund_in` / `fund_out` 交易表达 |

PIN、PAN token、配送地址和其他仅 UQPay 业务字段应首先保留在 service DTO；只有已实现功能需要且具备渠道中立语义时才扩展模型。

## Marxo 调用基准

SDK 位于 `pkg/dealer/uqpay`，实际调用位于 `app/new/card/service/internal/data/uqpay.go`。

| 业务 | SDK 调用 |
| --- | --- |
| 鉴权 | `GetAccessToken` |
| 持卡人 | `CreateCardholder`、`UpdateCardholder` |
| 开卡 | `CreateCard` |
| 库存卡分配 | `AssignCard` |
| 卡查询 | `GetCardInfo`、`GetCardPrivateInfo` |
| 卡状态 | `UpdateCardStatus` |
| 清算查询 | `ListTransaction` |
| 后续卡操作 | `RetrieveIssuingBalance`、`ResetCardPin`、`ActivateCard`、`UpdateCard`、`CreatePanToken` |

## 归档调研（暂不实施）

下列内容来自此前对 Marxo `pkg/dealer/uqpay` 的调研，当前不实施。若未来恢复 UQPay，所有路由仍须置于 `/uqpay` 前缀下，并重新确认请求/响应 DTO。

### 公共鉴权和错误规则

| 调用类别 | 必须的 header | 规则 |
| --- | --- | --- |
| Token | `x-api-key`、`x-client-id` | `POST /api/v1/connect/token` 返回 `code`、`message`、`auth_token`、`expired_at`。mock 可固定返回 token，不验证输入凭据。 |
| 读卡/读持卡人/读交易 | `x-auth-token` | mock 不验证 token。 |
| 所有发卡写操作、余额、模拟交易 | `x-auth-token`、`x-idempotency-key` | `x-idempotency-key` 必须是 UUID。先校验格式；以它作为 `Card.RequestID` 或 `LastOperationRequestID` 幂等键，重复请求不得产生第二张卡/第二笔交易。 |

SDK 把 HTTP 非 2xx 视为传输错误；很多成功 HTTP 响应仍以非空 `code` 表示业务失败。因此 mock 成功时返回空字符串 `code`，失败时返回非空 `code` 和 `message`，不要只依赖 HTTP 状态。

### 第一批：产品、持卡人和开卡

| Mock 路由（目标） | 下游请求字段 | 成功响应/内部映射 |
| --- | --- | --- |
| `POST /uqpay/api/v1/connect/token` | 无 JSON body | `{"code":"","message":"","auth_token":"mock-token","expired_at":...}`。不持久化 token。 |
| `GET /uqpay/api/v1/issuing/products?page_size=&page_number=` | 分页参数；SDK 还会附带 idempotency header | 返回 `total_pages`、`total_items`、`data[]`；每个产品有 `product_id`、`mode_type`、`card_bin`、`card_form`、`max_card_quota`、`card_scheme`、`card_currency`、时间和 `product_status`。`product_id` 由 UQPay formatter 表示 `CardProduct.ID`。 |
| `POST /uqpay/api/v1/issuing/cardholders` | `email`、`first_name`、`last_name`、`date_of_birth`、`country_code`、`phone_number`，可选 `document_type`/`document`；`delivery_address.city/country/line1/postal_code` 必填，`state/line2` 可选 | 返回 `cardholder_id`、`cardholder_status`。资料写入 `CardHolder`；配送/证件仅存在于 UQPay DTO，除非已有中立字段可承载。 |
| `GET /uqpay/api/v1/issuing/cardholders/{id}`、`POST .../{id}` | 查询无 body；更新只传需要修改的持卡人字段 | 路径 ID 先解析；更新 `CardHolder`。 |
| `POST /uqpay/api/v1/issuing/cards` | `card_currency`、`cardholder_id`、`card_product_id`；可选 `card_limit`、`spending_controls:[{amount,interval}]`、`risk_controls:{allow_3ds_transactions,allowed_mcc,blocked_mcc}` | 返回 `card_id`、`card_order_id`、`create_time`、`card_status`、`order_status`。创建 `Card`，产品/BIN、持卡人、币种、状态及请求 UUID 都必须写入；`card_order_id` 仅用于 operation DTO/异步关联，不能持久化为第二资源 ID。 |
| `POST /uqpay/api/v1/issuing/cards/assign` | 可选 `cardholder_id`、`card_number`、`card_currency`、`card_mode` | 分配库存实体卡，返回同样的卡/订单/状态字段；卡号关联到已有 `PhysicalCard`，不能新造外部卡 ID。 |

可作为端到端开卡夹具的报文：

```json
// POST /uqpay/api/v1/issuing/cards
// headers: x-auth-token: mock-token; x-idempotency-key: 80c4bb6e-3ce8-4afd-835a-704f21519f43
{
  "card_currency": "USD",
  "cardholder_id": "uqh_101",
  "card_product_id": "uqp_301",
  "card_limit": "100.00",
  "spending_controls": [{"amount": "100.00", "interval": "monthly"}],
  "risk_controls": {"allow_3ds_transactions": true, "allowed_mcc": ["5812"]}
}

// 200 OK
{"code":"","message":"","type":"success","card_id":"uqc_401","card_order_id":"uqo_501","create_time":"2026-09-17T10:00:00Z","card_status":"active","order_status":"completed"}
```

### 第二批：卡详情、状态和敏感数据

实现 `GET /api/v1/issuing/cards/{id}`，响应需覆盖 `card_id`、`card_bin`、`card_scheme`、`card_currency`、`card_number`、`form_factor`、`mode_type`、`card_product_id`、`card_limit`、`available_balance`、`cardholder`、`spending_controls`、`risk_controls`、`card_status`、`update_reason`、`consumed_amount`。`GET /api/v1/issuing/cards/{id}/secure` 返回 `cvv`、`expire_date`、`card_number`、`card_expiration_date`，全部从通用 `Card` 派生。

`POST /api/v1/issuing/cards/{id}/status` 的 body 是 `{"card_id":"...","card_status":"...","update_reason":"..."}`；更新 `Card.Status`，响应为 `card_id`、`card_order_id`、`order_status`、`update_reason`。`POST /api/v1/issuing/cards/{id}` 用于限额/风控更新；`POST /api/v1/issuing/cards/activate` 的 body 是 `activation_code`、`card_id`、`no_pin_payment_amount`、`pin`；`POST /api/v1/issuing/cards/pin` 的 body 是 `card_id`、`pin`。PIN、风控、免密额度保留 `Invalid:`，不得为了回显而加到通用模型。

### 第三批：清算、授权模拟和余额

`GET /uqpay/api/v1/issuing/transactions` 接收可选 `card_id`、`page_size`、`page_number`、`start_time`、`end_time`。响应为 `code/message/type/total_pages/total_items/data`；每条 `data` 至少有：`card_id`、`transaction_id`、`short_transaction_id`、`original_transaction_id`、`transaction_type`、`transaction_status`、`authorization_code`、交易/账单金额和币种、手续费/币种、`transaction_time`、`posted_time`、`merchant_data.{category_code,city,country,name}`、`description`、`wallet_type`。它们分别映射 `CardTransaction` 的卡、原交易、类型/状态、授权码、金额/币种、费用、发生/清算时间和商户字段。

`POST /api/v1/issuing/balances` body 为 `{"currency":"USD"}`，返回 `balance_id`、`available_balance`、`margin_balance`、`frozen_balance`、状态和时间；以 `Wallet` 表达，不持久化 `balance_id`。模拟授权是 `POST /api/v1/simulation/issuing/authorization`，body 为 `card_id`、`transaction_amount`、`transaction_currency`、`merchant_name`、`merchant_category_code`；创建 `Authorization` 和授权交易。撤销为 `POST /api/v1/simulation/issuing/reversal`，body 只含 `transaction_id`，其原交易必须已存在。

### 归档 webhook 夹具

以下为历史调研中的 UQPay webhook 消费模型，不是当前实现要求。若未来恢复 UQPay，mock 需要按确认后的负载生成事件，而不是只写通用 `WebhookRecord`。

```json
// 卡激活码事件的 payload
{"data":{"activation_code":"123456","card_number":"411111******1111","card_id":"uqc_401"}}

// 交易事件的 payload
{
  "data": {
    "card_id":"uqc_401","card_number":"411111******1111","cardholder_id":"uqh_101",
    "transaction_id":"uqt_601","short_transaction_id":"601","original_transaction_id":"",
    "transaction_type":"authorization","card_available_balance":"87.50","authorization_code":"A12345",
    "billing_amount":"12.50","billing_currency":"USD","transaction_amount":"12.50","transaction_currency":"USD",
    "transaction_fee":"0.00","transaction_fee_currency":"USD","fee_pass_through":"false",
    "transaction_time":"2026-09-17T10:01:00Z","category_code":"5812","city":"New York","country":"US",
    "name":"Mock Cafe","description":"purchase","transaction_status":"completed","wallet_type":"card"
  }
}
```

投递时先保存 `WebhookRecord`（格式化的 `SourceID`、完整 payload、`AttemptCount=1`、pending），提交业务 transaction 后 POST，再写回 HTTP 状态、响应、投递时间或错误；2xx 才成功。UQPay 接收端按 `event_id` 去重，重试必须复用同一 `event_id`，不能生成另一笔交易或事件。真实 URL、签名 header 和 event 枚举并未出现在已检查的 Marxo SDK 中，确认后应补充到此节，不能猜测。

实现时所有 mock 自有资源 ID 必须由 UQPay 的 `service/id.go` 格式化和解析，不能持久化下游视角的资源 ID。
