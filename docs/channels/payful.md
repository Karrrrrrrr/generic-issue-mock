# Payful（已废弃，历史归档）

## 状态

**Payful 已废弃，不再实现，也不属于未来渠道待办。** 不添加 `backend/channel/payful`、路由、管理 UI、种子数据或为它扩展通用模型。

以下内容仅保留此前的 SDK 调研记录；其中「实现」「应」「目标」等表述均为废止的历史方案，不构成当前开发要求。新增未来渠道为 [PingPong（待实现）](pingpong.md)，不能复用或推定 Payful 的协议作为 PingPong 契约。

## 目标字段映射

下表是实现时应采用的映射，不代表当前已经持久化。

| Payful 字段 | 目标通用表字段 | 处理 |
| --- | --- | --- |
| `cardId` | `Card.ID` | 新增 Payful formatter；不得保存另一套渠道资源 ID |
| `binRangeId` | `Card.CardProductID` / `Card.CardBin` | 产品主键关联；前缀从产品派生 |
| `localCurrency`、`startDate`、`endDate` | `Card.CardCurrency`、`Card.ExpireAt` | 起始日无中立字段，不持久化；结束日转换为时间 |
| `cardStatus` | `Card.Status` | 使用 Payful 枚举转换 |
| `userReqNo` | `Card.RequestID`、`Card.LastOperationRequestID` | 幂等和异步操作关联 |
| `authLimitAmount` / 充值金额 | `Wallet.Available` 或 `CardTransaction.TxAmount` | 余额变动创建 `fund_in` / `fund_out` 交易，不将渠道额度作为卡字段保存 |
| 交易 `cardId`、金额、类型、状态、时间 | `CardTransaction.CardID`、`TxAmount`、`Type`、`Status`、`OccurredAt` / `SettledAt` | 清算查询映射 |
| 持卡人资料 | `CardHolder` | 只存姓名、联系方式和渠道中立地址/证件字段 |

`cardAlias`、`cardLabel`、币种校验开关和仅属于 Payful 的共享卡组字段没有对应的中立概念，除非实现的渠道功能确实需要，不加入通用模型。

## Marxo 调用基准

| 业务 | SDK 调用 | HTTP 协议 |
| --- | --- | --- |
| 开卡 | `ApplyCard` / 上层 `CreateCard` | `POST /api/vas/card/apply` |
| 查询开卡操作 | `GetCardUserReqNoInfo` | 卡操作查询接口 |
| 充值 | `RechargeCard` | `POST /api/vas/card/recharge` |
| 退款 | `RefundCard` | `POST /api/vas/card/refund` |
| 清算查询 | `QueryTransactionsV3` | `POST /api/vas/trans/v3/query` |

SDK 位于 Marxo 的 `pkg/dealer/payful/payful.go`；实际调用位于 `app/new/card/service/internal/biz/card_open_third_party_qf.go` 和 `internal/data/payful.go`。

## 目标报文示例

以下是 Marxo SDK 的目标协议，不是当前 mock 可调用路由。

```json
// POST /api/vas/card/apply
{
  "userReqNo":"open-card-001",
  "localCurrency":"USD",
  "startDate":"2026-09-17",
  "endDate":"2028-09-17",
  "authLimitAmount":100.00,
  "enableMultiUse":1,
  "enableCurrencyCheck":0,
  "binRangeId":"bin-001",
  "channelType":"1"
}

// POST /api/vas/card/recharge
{"cardId":"12345","userReqNo":"recharge-001","authLimitAmount":20.00,"channelType":"1"}

// POST /api/vas/card/refund
{"cardId":"12345","userReqNo":"refund-001","refundAmount":12.50,"channelType":"1"}

// POST /api/vas/trans/v3/query
{
  "cardType":"0",
  "cardId":"12345",
  "beginTime":"2026-09-17 00:00:00",
  "endTime":"2026-09-17 23:59:59",
  "timeType":0,
  "commonTransType":1,
  "currentPage":1,
  "pageSize":50
}
```

`commonTransType`：`1` 消费、`2` 冲正、`3` 退款、`4` 清算差额、`5` 冲正后请款、`7` 强制清算。`commonTransStatus=3` 表示已清算。

## 归档调研（不实施）

以下清单是下一位实现者的协议边界。渠道路由必须加 `/payful` 前缀；service 层使用 Payful DTO，业务层只接收已转换的中立请求。Payful SDK 的鉴权/签名由其 `restyRequest` 统一附加；mock 按项目约束无需验证 token，不能把签名字段写入通用表。

### 第一批：开卡和异步订单

| Mock 路由（目标） | 下游方法 | 必须接收的 JSON | 成功结果与内部写入 |
| --- | --- | --- | --- |
| `POST /payful/api/vas/card/apply` | `POST /api/vas/card/apply` | `userReqNo`、`localCurrency`、`startDate`、`endDate`、`authLimitAmount`、`enableMultiUse`、`binRangeId`；可选 `enableCurrencyCheck`、`cardAlias`、`channelType`、`groupId`、`cardUserId` | 返回 `orderId`、`cardId`、`localCurrency`、`cardNo`、`cardVerifyNo`、`cardExpiryDate`。新建 `Card`，`RequestID=userReqNo`；`orderId` 只能放 operation DTO/响应，不能加第三方 ID 列。 |
| `POST /payful/api/vas/card/queryCardOperate` | 卡操作订单查询 | 至少 `userReqNo` 或 `cardId`；`status` 为 `0` 待处理、`1` 处理中、`2` 成功、`3` 失败；`opType` 为 `0` 开卡、`1` 充值、`3` 销卡、`4` 退款；`beginTime`/`endTime` 是 `yyyy-MM-dd`；`currentPage`、`pageSize`（最大 50）必传 | 按 `userReqNo` 找 `Card.RequestID` 或 `LastOperationRequestID`。分页必须 `ID DESC`。未命中先 `Exist` 再返回该渠道定义的空结果，不能以 not-found 错误推断。 |
| `POST /payful/api/vas/card/info` | 卡详情 | `cardId` | 解析渠道 ID 后读取 `Card`，返回卡号、CVV、有效期、状态、余额等 SDK 实际使用字段。 |

开卡报文与异步完成轮询应成对实现，不能只创建卡而遗漏 `queryCardOperate`：

```json
// POST /payful/api/vas/card/queryCardOperate
{
  "userReqNo": "open-card-001",
  "opType": 0,
  "currentPage": 1,
  "pageSize": 50
}

// 目标成功项至少应包含能让 Marxo 关联的请求号、订单和卡
{
  "userReqNo": "open-card-001",
  "orderId": "order-201",
  "cardId": "201",
  "status": 2,
  "opType": 0
}
```

### 第二批：资金和卡状态

| Mock 路由（目标） | 下游方法 | 请求和业务写入 |
| --- | --- | --- |
| `POST /payful/api/vas/card/recharge` | 充值 | `cardId`、`userReqNo`、`authLimitAmount`，可选 `channelType`、`groupId`。先按卡 ID 查卡；在 transaction 中增加 `fund_in` 的 `CardTransaction`，更新该卡钱包；`LastOperationRequestID=userReqNo`。响应返回异步 `orderId`。 |
| `POST /payful/api/vas/card/refund` | 退款 | `cardId`、`userReqNo`、`refundAmount`，可选 `channelType`、`groupId`。创建 `fund_out` 交易，金额不得超过可用余额；响应同样返回 `orderId`。 |
| `POST /payful/api/vas/card/modifyCard` | 冻结/解冻/额度更新 | `userReqNo`、`cardId`，可选 `cardAlias`、`cardLabel`、`status`、`authLimitAmount`。只有 `status` 映射 `Card.Status`；别名/标签保留 `Invalid:`，不得扩展通用模型。 |
| `POST /payful/api/vas/card/close` | 销卡 | `cardId`、`userReqNo`。创建销卡 operation，状态完成后转换为通用关闭状态。 |

每个变更类请求都必须用 `userReqNo` 幂等：重复请求返回首次 operation 的相同结果，不重复记账。`authLimitAmount` 是 Payful 授权额度，不作为 `Card` 的新列；只有已实现的入金/出金才改变 `Wallet.Available`。

### 第三批：清算查询

实现 `POST /payful/api/vas/trans/v3/query`，完整过滤 DTO 是：`cardType`（`0` 常规卡，`1` 共享卡）、可选 `cardId`、`beginTime`、`endTime`（`yyyy-MM-dd HH:mm:ss`）、`timeType`（`0` 交易时间，`1` 创建时间）、可选 `commonTransType`、可选 `commonTransStatus`、`currentPage`、`pageSize`（最大 50）、可选 `queryUserNo`。列表查询必须按通用交易 `ID DESC` 排序后分页。

`commonTransType` 的映射：`1` 消费、`2` 冲正、`3` 退款、`4` 清算差额、`5` 冲正后请款、`7` 强制清算；`commonTransStatus`：`1` 批准、`2` 拒绝、`3` 已清算。返回项至少提供交易 ID、卡 ID、原交易 ID、类型、状态、金额/币种、商户、交易时间和清算时间，使 Marxo 不必回读数据库推断清算状态。

### Webhook：需要先定下游契约

已检查的 Marxo Payful SDK 只有轮询操作订单和交易清算，未发现 Payful 入站 webhook DTO 或消费端。因此不要假造 URL、签名头或事件名。若后续确定 Payful webhook 契约，按以下固定步骤实现：

1. 在 Payful service 定义渠道专属 payload（含下游 event ID、operation/order ID、`userReqNo`、`cardId`、状态、发生时间），先由 formatter 转换 mock 卡 ID。
2. 对每个启用的 `WebhookConfig`（渠道和事件匹配）先创建 `WebhookRecord`：`Channel=payful`、`Event`、`TargetURL`、格式化后的 `SourceID`、完整 `Payload`、`AttemptCount=1`、`Status=pending`。
3. 提交业务 transaction 后 HTTP POST；不要在数据库 transaction 内投递。写回 `StatusCode`、`ResponseBody`、`DeliveredAt`，失败写 `ErrorMessage` 和失败状态；重试更新同一 record 的次数和结果。
4. 收到的 `2xx` 才成功。重试次数、退避和签名头必须以确认后的 Payful 协议为准，并补入本节的真实报文。
