# 通用开卡组件

`shared/biz` 只处理已经归一化的开卡请求，不解释任何渠道协议，也不在内部查找“默认渠道账户”。渠道层负责解析外部 ID、选择产品/虚拟账户、校验协议参数，以及将共享错误映射为渠道错误码。

本次组件没有自动替换现有渠道 usecase，也没有合并 UI/OpenAPI 的业务入口。接入时由具体 usecase 明确调用；UI 和 OpenAPI 仍各自拥有协议适配、操作日志及差异化规则。

## 依赖与注册

- `CardIssuer`：通用开卡入口。
- `AccountRepo`、`CardRepo`、`CardHolderRepo`、`CardProductRepo`、`VirtualAccountRepo`、`WalletRepo`：只包含开卡实际需要的仓储操作。
- `Transaction`：统一事务边界，返回成功必须表示真实提交完成。
- `Notificator`：由上层实现并在每次开卡请求中传入。通用层不负责 webhook DTO、签名、HTTP 调用、队列或重试策略。

已有一个注册了根数据库连接 `*gorm.DB` 的 `samber/do` injector 时：

```go
shared.RegisterProviders(injector)
issuer := do.MustInvoke[biz.CardIssuer](injector)
```

`RegisterProviders` 直接注册构造器，包括可选用的 GORM Gen 仓储实现。若上层已有其他存储适配，可只注册 `biz.NewCardIssuer`，自行提供上述仓储接口和 `Transaction`，不必引用 `shared/data`。

## 调用约定

`Issue` 接受一个 `IssueCardReq`，返回 `(*IssueCardResult, error)`：

```go
result, err := issuer.Issue(ctx, &biz.IssueCardReq{
	Channel:          channel,
	AccountID:        accountID,
	CardProductID:    productID,
	CardType:         enums.CardType_VirtualAccountSingle,
	VirtualAccountID: &virtualAccountID,
	Currency:         currency,
	CardScheme:       cardScheme,
	FormType:         enums.CardFormType_Virtual,
	Status:           enums.CardStatus_Active,
	ExpireAt:         expireAt,
	RequestID:        &requestID,
	Notificator:      channelNotificator,
})
```

- `AccountID`、`CardProductID` 必须为正数，`Channel`、币种、卡组织必须明确提供。
- 卡形态支持 `virtual`、`physical`，开卡状态支持 `active`、`inactive`。有效期必须由调用方提供且晚于当前时间；通用层不默认 100 年、两年或任何渠道期限。
- `VirtualAccountID`、持卡人 ID、`RequestID` 使用指针区分省略和显式传入。显式零 ID、显式空请求号无效。
- `RawRequest` 可省略，默认存 `{}`；提供时必须是合法 JSON，不在公共层解释渠道字段。
- `Notificator` 是必填接口。无需通知时显式传 `biz.NoopNotificator{}`；它仅表示跳过通知，不记录或伪装 webhook 投递成功。
- 调用不会回写请求中的持卡人 ID、虚拟账户 ID、原始 JSON 等数据。使用返回的 `result.Card` 获取新资源 ID。

通知请求包含 `AccountID`、`Channel`、`CardID`，均为内部模型标识。上层适配器实现：

```go
NotifyIssueCard(context.Context, *biz.NotifyIssueCardReq) error
```

适配器可以同步发送，也可以只提交异步投递任务。它负责渠道 ID 编码、事件、载荷、签名、ACK 和重试，返回错误而不重复记录同一个错误；共享层在首次处理通知失败时记录日志。PingPong 在 webhook 契约确认前不能借此构造猜测报文或把 `contract_pending` 当成投递尝试。

## 三种钱包语义

| 卡类型 | 虚拟账户入参 | 实际消费钱包 |
| --- | --- | --- |
| `single` | 必须省略 | 新建零余额独立卡钱包 |
| `share` | 必须提供 | 复用该虚拟账户钱包 |
| `virtual_account_single` | 必须提供 | 新建零余额独立卡钱包，保留虚拟账户供资关系 |

通过现有 `pkg/cardwallet.Prepare` 校验虚拟账户及钱包的账户、渠道、币种和类型。绑定虚拟账户不意味着共享余额，也不会自动扣款、充值或挪用根账户的钱包。开卡充值由上层原有资金流程处理。

该公共能力不会自动开放任何渠道的新卡类型。PhotonPay 的默认虚拟账户选择、Paynda 只支持独立卡等渠道限制仍由渠道 usecase 决定。

## 持卡人与产品

- `IssueCardHolder == nil`：不关联持卡人。
- `IssueCardHolder.ID != nil`：关联同账户、同渠道的现有持卡人；此时不得同时提供新建持卡人的字段，也不修改现有持卡人的 `Shared` 属性。
- 提供 `IssueCardHolder` 且 `ID == nil`：按照明确传入的资料创建本卡独有的持卡人，`Shared=false`。至少提供一项非空资料，渠道必填字段由上层先校验。
- 产品按 `(Channel, ID)` 查找并加行锁，序号跨账户共享；序号递增、持卡人创建、钱包创建、卡片创建在同一事务中。
- 使用 `pkg/cardnumber.Generate` 校验 BIN 和容量，保留已有的 16 位卡号规则。只有 PingPong 可配置多 BIN，生成卡片只保存选中的 BIN；非法 BIN、负序号及序号溢出不会推进序列。

## 事务与通知结果

1. 任一仓储操作失败，整个开卡事务回滚，不发通知。错误在首次处理的位置直接记录一次。
2. 只有事务提交成功后才调用一次 `Notificator`，传入真实卡 ID 和原始调用上下文，不传已经结束的事务上下文。
3. 通知失败不会回滚或伪装成开卡失败：`Issue` 返回成功的卡片，通知错误放在 `result.NotificationError`。上层可告警或重试通知，不能据此再次开卡。
4. 组件必须拥有提交边界。自带事务实现会拒绝已有 `gormx` 事务上下文，避免只释放嵌套 savepoint 就提前通知；注册的数据库也必须是根连接，不是未提交的事务句柄。自定义 `Transaction` 实现须遵守同样的提交语义。
5. 不提供持久化 outbox、自动补偿或“恰好一次”投递保证。若进程在提交之后、通知之前退出，补发与持久化投递任务由上层适配器处理。

本组件不统一渠道幂等应答。数据库保留既有 `(account_id, channel, request_id)` 唯一约束，重复请求失败时所有新增记录及产品序号均回滚；已有请求的查询、请求内容冲突判断、返回历史卡片等策略由上层 usecase 按渠道契约实现。
