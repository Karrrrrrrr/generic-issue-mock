# Slash

## 范围

已实现虚拟账户列表/划拨、卡产品、虚拟卡开卡、持卡人、交易查询、管理端授权/清算/冲正/退款模拟和管理 UI。实体卡、开卡时选择虚拟账户、单次卡、卡组、消费限制和 webhook OpenAPI 尚未实现。

## 字段映射

| Slash 字段 | 通用表字段 | 处理 |
| --- | --- | --- |
| `id`、`:id` | `Card.ID` 或 `CardTransaction.ID` | 通过 Slash UUID formatter 输出；入参反解析为内部 ID |
| `ApiKey` | `Account.ID` | Slash UUID formatter 解析的渠道账户 ID；中间件写入当前账户域 |
| `cardProductId` | `Card.CardProductID` | Slash UUID 解析后关联 `CardProduct`；其 `Prefix` 写入 `Card.CardBin` |
| `pan`、`cvv`、`last4`、`expiryMonth`/`expiryYear` | `Card.CardNumber`、`Card.Cvv`、`Card.ExpireAt` | `last4` 从卡号派生，不单独存储 |
| `status`、`isPhysical` | `Card.Status`、`Card.FormType` | 枚举和布尔值转换；当前只创建虚拟卡 |
| `source`、`destination`、`amountCents` | `VirtualAccount.ID`、`Wallet.Amount`、`Wallet.Out` / `In` | UUID 解析虚拟账户；分为 100 转换到 `decimal` 金额 |
| `transaction.id`、`cardId` | `CardTransaction.ID`、`CardTransaction.CardID` | 均使用 Slash UUID 格式 |
| `providerAuthorizationId` | `CardTransaction.AuthorizationID` | 关联授权内部 ID后以 UUID 输出 |
| `amountCents`、`status`、`authorizedAt`、商户描述 | `CardTransaction.TxAmount`、`Status`、`OccurredAt`、`MerchantName` | 金额从分转换，状态经 Slash 枚举转换 |

`accountId`、开卡 `virtualAccountId`、`name`、`isSingleUse`、`spendingConstraint`、`cardGroupId`、更新 `userData` 和 cursor 分页目前都没有已实现的中立持久化行为，service DTO 标记为 `Invalid:`。`ApiKey` 不是 `Invalid:`：它选择当前 Slash 账户域。开卡时仅保留 `userData.requestId` 作为 `Card.RequestID`。

## 账户解析

Slash 的 `ApiKey` 是当前账户 ID，使用 Slash UUID formatter 解析为 `Account.ID`。账户中间件校验其 `Channel=slash` 后，将 `(AccountID, Channel)` 写入请求上下文；所有卡、持卡人、授权、交易、钱包和虚拟账户查询都以该域过滤。`VirtualAccount` 是该域内的独立余额资源，不能取代 `Account`，也不能作为 `ApiKey`。

## Marxo 调用基准

| 业务 | SDK 调用 | HTTP 协议 |
| --- | --- | --- |
| 开卡 | `CreateCard` | `POST /card` |
| 虚拟账户划拨 | `VirtualAccountTransfer` | `POST /transfer/virtual-account` |
| 清算列表查询 | `ListTransactions` | `GET /transaction` |
| 单笔交易查询 | `GetTransaction` | `GET /transaction/{id}` |

SDK 位于 Marxo 的 `pkg/dealer/slash/slash.go`，实际清算读取/划拨调用位于 `app/new/card/service/internal/data/slash_data.go`。

## Mock 报文示例

Slash mock 自有资源 ID 是可逆的 UUID 形字符串。Marxo SDK 用 `X-Idempotency-Key` 传递开卡请求 ID；当前 mock 从 `userData.requestId` 读取该值，尚未绑定 header，这是待修正的协议偏差。

```bash
curl -X POST http://127.0.0.1:8000/slash/card \
  -H 'Content-Type: application/json' \
  -H 'ApiKey: 00000000-0000-0000-0000-000000000001' \
  -H 'X-Idempotency-Key: open-card-20260917-003' \
  -d '{
    "accountId":"mock-account",
    "virtualAccountId":"00000000-0000-0000-0000-000000000001",
    "type":"virtual",
    "name":"Mock Card",
    "isSingleUse":false,
    "cardProductId":"00000000-0000-0000-0000-000000000301",
    "userData":{"requestId":"open-card-20260917-003","cardId":"merchant-card-001"}
  }'
```

交易读取和虚拟账户划拨：

```bash
curl 'http://127.0.0.1:8000/slash/transaction?filter:cardId=00000000-0000-0000-0000-000000000401&page_number=1&page_size=50'
curl 'http://127.0.0.1:8000/slash/transaction/00000000-0000-0000-0000-000000000501'

curl -X POST http://127.0.0.1:8000/slash/transfer/virtual-account \
  -H 'Content-Type: application/json' \
  -d '{"source":"00000000-0000-0000-0000-000000000001","destination":"00000000-0000-0000-0000-000000000002","amountCents":1250}'
```

## 待办

### 协议修正：开卡幂等

`POST /slash/card` 必须读取并校验 `X-Idempotency-Key`，将其写入 `Card.RequestID`；`userData.requestId` 只作为 Slash 透传数据，不能作为替代幂等键。重复 key 必须返回首次创建的卡，不再创建卡号或扣减虚拟账户余额。下列输入组合是实现夹具：

```json
// header: X-Idempotency-Key: 80c4bb6e-3ce8-4afd-835a-704f21519f43
{
  "accountId":"mock-account","virtualAccountId":"...0001","type":"virtual",
  "cardProductId":"...0301","isSingleUse":false,
  "spendingConstraint":{"amountCents":10000,"interval":"monthly"},
  "userData":{"requestId":"merchant-visible-only","cardId":"merchant-card-001"}
}
```

虚拟账户选择完成后，解析 `virtualAccountId` 到 `VirtualAccount.ID` 并写入 `Card.VirtualAccountID`/`WalletID`；`accountId`、`userData.cardId` 和消费限制尚无中立字段，维持 `Invalid:`。账户域由 `ApiKey` 确定，不从 body 的 `accountId` 推断。实体卡、单次卡和卡组必须先确认现有通用表是否能表达，再决定是否扩展；不能为了原样保存 Slash 请求而添加渠道字段。

### Webhook

已检查的 Marxo Slash SDK 只定义开卡、划拨和交易查询，未定义 Slash webhook 消费协议。待确认实际协议前，不得猜测事件名、认证头或签名。实现范围应至少覆盖 UI 已配置的交易和授权创建/更新四类事件；交易 payload 含格式化卡/交易/原交易/授权 ID、分单位金额、币种、状态、商户和时间，授权 payload 含授权结果与商户数据。

同样采用固定投递账本：提交业务 transaction 后查询启用 `WebhookConfig`，先建 pending `WebhookRecord`（channel、event、URL、格式化 `SourceID`、payload、attempt 1），再 POST；记录 2xx 成功、其他结果失败，并对同一 record 记录重试次数/响应/错误。协议确定后补入 URL、headers、签名、退避与完整 JSON 夹具。
