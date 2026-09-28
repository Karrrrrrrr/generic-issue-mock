# Slash

## 范围

已实现虚拟账户列表/划拨、卡产品、虚拟卡开卡、持卡人、交易查询、管理端授权/清算/冲正/退款模拟和异步 webhook 投递。同步授权配置通过管理面 `GET/PUT /slash/ui/authorization-config` 管理。实体卡、开卡时选择虚拟账户、单次卡、卡组、消费限制和异步 webhook 的 OpenAPI 管理接口尚未实现。

管理端四种模拟统一调用 `shared/biz.CardTransactionSimulator`：授权冻结可用余额，清算释放对应冻结后扣款，退款直接入账，撤销解冻。模拟请求不再传 `account_id`；授权根据卡 ID 读取账户，清算、关联退款和撤销根据授权 ID 读取账户及卡。独立退款必须提供卡 ID 和币种，`authorization_id` 省略，显式零 UUID 不作为省略。Webhook 仅由模拟器在提交后触发原渠道通知适配器；清算允许超额和负余额，旧未冻结授权应在独立测试库重新创建。

## 字段映射

| Slash 字段 | 通用表字段 | 处理 |
| --- | --- | --- |
| `id`、`:id` | `Card.ID` 或 `CardTransaction.ID` | 通过 Slash UUID formatter 输出；入参反解析为内部 ID |
| `X-API-Key` | `Account.ID` | Slash UUID formatter 解析的渠道账户 ID；service 显式传入 usecase 请求 |
| `cardProductId` | `Card.CardProductID` | Slash UUID 解析后关联 `CardProduct`；其 `Prefix` 写入 `Card.CardBin` |
| `pan`、`cvv`、`last4`、`expiryMonth`/`expiryYear` | `Card.CardNumber`、`Card.Cvv`、`Card.ExpireAt` | `last4` 从卡号派生，不单独存储 |
| `status`、`isPhysical` | `Card.Status`、`Card.FormType` | 枚举和布尔值转换；当前只创建虚拟卡 |
| `source`、`destination`、`amountCents` | `VirtualAccount.ID`、`Wallet.Available`、`Wallet.Out` / `In` | UUID 解析虚拟账户；分为 100 转换到 `decimal` 金额 |
| `transaction.id`、`cardId` | `CardTransaction.ID`、`CardTransaction.CardID` | 均使用 Slash UUID 格式 |
| `providerAuthorizationId` | `CardTransaction.AuthorizationID` | 关联授权内部 ID后以 UUID 输出 |
| `amountCents`、`status`、`authorizedAt`、商户描述 | `CardTransaction.TxAmount`、`Status`、`Authorization.CreatedAt`、`MerchantName` | 金额从分转换，状态经 Slash 枚举转换 |

`accountId`、开卡 `virtualAccountId`、`name`、`isSingleUse`、`spendingConstraint`、`cardGroupId`、更新 `userData` 和 cursor 分页目前都没有已实现的中立持久化行为，service DTO 标记为 `Invalid:`。`ApiKey` 不是 `Invalid:`：它选择当前 Slash 账户域。开卡时仅保留 `userData.requestId` 作为 `Card.RequestID`。

## 账户解析

Slash 的 `X-API-Key` 是当前账户 ID，使用 Slash UUID formatter 解析为 `Account.ID`。service 将该值放入每个账户资源的 usecase/repository 请求结构；不得通过中间件或 `context.Context` 保存账户范围。所有卡、持卡人、授权、交易、钱包和虚拟账户查询都以 `(AccountID, Channel=slash)` 过滤。`VirtualAccount` 是该域内的独立余额资源，不能取代 `Account`，也不能作为 `X-API-Key`。

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

Marxo 目前没有 Slash 授权配置的实际调用点，mock 使用账户唯一的 `AuthorizationConfig`，只通过管理面管理 URL、启用状态和超时。同步回调连接失败时模拟授权失败，不设置回退策略。种子为 `aggregated_transaction.create`、`aggregated_transaction.update`、`card_creation.event`、`card.update` 和 `card.delete` 各创建一条配置，目标为 Marxo 的 `POST /api/v1/notify/xz-event`。

提交业务 transaction 后，mock 仅查询同一账户下启用且事件匹配的 `WebhookConfig`，先建立 pending `WebhookRecord`，再投递 JSON：

```json
{
  "entityId": "Slash UUID",
  "event": "card.update",
  "eventId": "Slash UUID"
}
```

`entityId` 和 `eventId` 都由本次卡或交易的内部 `ID` 转成 Slash UUID；每个配置独立投递并记录请求头、响应头、响应体、状态码和错误。2xx 标为成功，网络错误和非 2xx 标为失败。卡创建发送 `card_creation.event`，普通状态更新发送 `card.update`，关闭卡发送 `card.delete`；授权、模拟退款、清算、冲正和退款步骤各创建一条交易并发送 `aggregated_transaction.create`。`aggregated_transaction.update` 保留为可管理、可初始化的渠道事件，但不用于修正既有交易，因为 mock 的交易不可变。

投递记录列表、详情和 replay 管理页尚未实现。
