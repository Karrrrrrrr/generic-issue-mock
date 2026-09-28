# 通用卡业务组件

`shared/biz` 处理已经归一化的开卡、模拟授权、清算、退款和撤销请求，不解释任何渠道协议，也不在内部查找“默认渠道账户”。渠道层负责解析外部 ID、选择产品/虚拟账户、校验协议参数，以及将共享错误映射为渠道错误码。

本次组件没有自动替换现有渠道 usecase，也没有合并 UI/OpenAPI 的业务入口。接入时由具体 usecase 明确调用；UI 和 OpenAPI 仍各自拥有协议适配、操作日志及差异化规则。

## 依赖与注册

- `CardIssuer`：通用开卡入口。
- `CardTransactionSimulator`：仅暴露模拟授权、清算、退款和撤销四个方法的接口；具体实现 `cardTransactionSimulator` 不导出。请求、`Validate`、实现和复用的私有方法集中在 `biz/card_transaction_simulator.go`，不再拆出只有一次调用的流程方法。
- `AccountRepo`、`CardRepo`、`CardHolderRepo`、`CardProductRepo`、`VirtualAccountRepo`、`WalletRepo`、`AuthorizationRepo`、`CardTransactionRepo`：按资源拆分，只包含当前共享业务需要的仓储操作。
- `Transaction`：统一事务边界，返回成功必须表示真实提交完成。
- `Notificator`：由上层实现并在每次开卡请求中传入。通用层不负责 webhook DTO、签名、HTTP 调用、队列或重试策略。
- `CardTransactionNotificator`：只负责交易通知，不要求实现开卡通知接口。
- `CardTransactionAccounting`：由上层为授权、清算、退款和撤销选择一致的记账策略；共享流程不根据渠道名称分支。

已有一个注册了根数据库连接 `*gorm.DB` 的 `samber/do` injector 时：

```go
shared.RegisterProviders(injector)
issuer := do.MustInvoke[*biz.CardIssuer](injector)
simulator := do.MustInvoke[biz.CardTransactionSimulator](injector)
```

`RegisterProviders` 直接注册构造器，包括可选用的 GORM Gen 仓储实现。若上层已有其他存储适配，可按需注册 `biz.NewCardIssuer` / `biz.NewCardTransactionSimulator`，自行提供对应仓储接口和 `Transaction`，不必引用 `shared/data`。

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

开卡入口不统一渠道幂等应答。数据库保留既有 `(account_id, channel, request_id)` 唯一约束，重复请求失败时所有新增记录及产品序号均回滚；已有请求的查询、请求内容冲突判断、返回历史卡片等策略由上层 usecase 按渠道契约实现。

## 模拟卡交易

`CardTransactionSimulator` 的四个入口 `SimulateAuthorization`、`SimulateClearing`、`SimulateRefund` 和 `SimulateReversal` 均返回 `(*CardTransactionSimulationResult, error)`。各入口先调用请求的 `Validate()`，随后直接在自己的事务回调内实现流程。结果包含授权记录、当前交易阶段、带符号的授权剩余金额、是否重放，以及提交后的通知错误；独立退款的 `Authorization` 为 nil、`Remaining` 为零。

```go
accounting := biz.ClearingDebitAccounting{}
notificator := biz.NoopNotificator{}
authorizationRequestID := "authorization-001"
clearingRequestID := "clearing-001"

authorizationResult, err := simulator.SimulateAuthorization(ctx, &biz.SimulateAuthorizationReq{
	AccountID:   accountID,
	Channel:     channel,
	CardID:      cardID,
	Currency:    currency,
	Amount:      decimal.NewFromInt(10),
	RequestID:   &authorizationRequestID,
	Accounting:  accounting,
	Notificator: notificator,
})
if err != nil {
	return err
}

clearingResult, err := simulator.SimulateClearing(ctx, &biz.SimulateClearingReq{
	AccountID:       accountID,
	Channel:         channel,
	CardID:          cardID,
	AuthorizationID: authorizationResult.Authorization.ID,
	Amount:          decimal.NewFromInt(12),
	RequestID:       &clearingRequestID,
	Accounting:      accounting,
	Notificator:     notificator,
})
if err != nil {
	return err
}
```

上例允许对 10 元授权清算 12 元，`clearingResult.Remaining` 为 -2；继续以新的请求 ID 清算仍然有效。

### 退款与撤销

```go
refundRequestID := "refund-001"
refundResult, err := simulator.SimulateRefund(ctx, &biz.SimulateRefundReq{
	AccountID:   accountID,
	Channel:     channel,
	CardID:      cardID,
	Currency:    currency,
	Amount:      decimal.NewFromInt(3),
	RequestID:   &refundRequestID,
	Accounting:  accounting,
	Notificator: notificator,
})
if err != nil {
	return err
}

reversalRequestID := "reversal-001"
reversalResult, err := simulator.SimulateReversal(ctx, &biz.SimulateReversalReq{
	AccountID:       accountID,
	Channel:         channel,
	CardID:          cardID,
	AuthorizationID: authorizationResult.Authorization.ID,
	Amount:          decimal.NewFromInt(2),
	Status:          enums.TransactionStatus_VOID,
	RequestID:       &reversalRequestID,
	Accounting:      accounting,
	Notificator:     notificator,
})
if err != nil {
	return err
}
```

- 退款的 `AuthorizationID` 是可选指针。nil 表示独立退款，持久化为 0，不创建或查找虚构授权；非 nil 必须是同账户、同渠道、同一卡的正数授权 ID，传入 0 不会被当作省略。
- 独立退款要求卡处于有效状态；关联已有授权的退款与撤销不因卡被冻结而拒绝。两种退款均不要求存在先前清算，也不限制为已清算金额。
- 关联退款的请求币种必须与授权一致；未提供的商户字段继承授权，明确提供的字段按请求记录。独立退款使用请求的币种和商户字段。
- 退款增加卡实际消费钱包的 `Amount` 和 `In`，不释放占款、不增加或扣减授权剩余金额。
- 撤销的 `AuthorizationID` 必填，币种、商户信息及原授权交易关联从授权继承。撤销扣减授权剩余金额，保留负数，不向钱包退回未实际扣过的资金。
- 撤销请求的 `Status` 必填，仅允许 `TransactionStatus_VOID` 或 `TransactionStatus_SUCCEED`。上层按已有渠道约定传入：Slash、Paynda、PhotonPay 使用 VOID，PingPong 使用 SUCCEED。交易 `Type` 始终为 VOID，共享层不硬编码渠道判断。

### 记账差异

独立的策略定义在 `biz/card_transaction_accounting.go`，不是拆散模拟器的方法：

- `ClearingDebitAccounting`：授权只创建授权与 AUTH 交易，不检查余额、不占款；清算从钱包 `Amount` 扣款、累加 `Out`；退款增加余额和 `In`；撤销不改变钱包余额和占款。对应现有 Slash、Paynda、PhotonPay 的模拟记账行为。
- `AuthorizationHoldAccounting`：授权检查 `Amount - PendingOut` 是否足够，成功才增加 `PendingOut`；清算释放本授权尚未消费的占款，再扣款和累加 `Out`；退款增加余额和 `In`，不释放占款；撤销只释放本授权尚未消费的占款，不增加余额。对应 PingPong 的本地授权规则。
- 两种清算策略均不检查余额或授权剩余额度，允许超额清算、负钱包余额和负授权剩余金额。占款释放最多为本授权的正数剩余额度，不会在剩余为负时继续释放其他授权的占款。
- `Accounting` 必须由上层明确传入，不提供隐式默认值。上层应按渠道固定策略，并对同一授权的后续操作保持一致；不能给没有冻结余额的旧授权切换为占款释放策略。清算和撤销的占款释放都不会超过本授权的正数剩余额度。
- 自定义策略只修改传入钱包的 `Amount`、`PendingOut`、`In`、`Out`，不修改资源身份、归属或币种，不自行访问数据库、发送通知或提交事务。共享流程统一持久化修改。策略不得把授权与清算自动推断成相反的记账方向，也不得收紧清算的超额规则。

### 数据与事务边界

- 四种模拟金额均必须为正数。账户、卡 ID 必填且为正数；清算和撤销的授权 ID 必填，退款的授权 ID 可选。可选请求 ID 和商户字段使用指针，非 nil 的空字符串会被拒绝。
- 授权要求卡有效，卡、钱包和请求币种一致。清算必须引用同一账户、同一渠道、同一卡的有效授权，币种和商户信息从授权继承；卡被冻结后仍可清算已有授权。
- 总是使用 `Card.WalletID` 对应的钱包。`single`、`virtual_account_single` 使用卡钱包，`share` 使用虚拟账户钱包；不会因为卡关联 VA 就改扣 VA 或根账户的钱包。
- 账户行锁串行化同账户的共享模拟请求；授权、卡及钱包按需加行锁。账户、授权、卡、钱包和历史阶段查询均明确限制归属。
- 授权记录、AUTH/CLEAR/REFUND/VOID 阶段和钱包变更在各自操作的同一事务内提交，任一步失败全部回滚。关联操作保留授权交易关联，剩余金额扣除成功清算及已撤销金额，不受退款或失败阶段影响。
- 清算不会以剩余金额大于零作为前提，也不会把负剩余金额截成零。模型和现有渠道协议没有变更；四种模拟能力均已提供，共享组件尚未替换各渠道调用入口。

### 请求重放与通知

- `RequestID == nil` 表示新模拟操作，不进行去重。显式传入请求 ID 后，以 `(AccountID, Channel, RequestID)` 在共享模拟器内去重；授权、清算、退款、撤销共用此键空间，应使用不同的键。
- 同键同业务参数返回原交易并标记 `Replayed=true`，不再次记账、创建记录或调用通知。不同卡、阶段、授权引用、金额、币种/商户信息或撤销状态会返回冲突。独立退款重放不查询授权，也不会重复入账；重放结果的 `Remaining` 是查询时的当前剩余金额，不是首次调用快照。
- 并发去重依赖事务中的账户行锁；自定义仓储必须真正加锁，其他直接写入交易的代码不能绕过同一锁约定。没有新增数据库唯一索引或修改模型。
- 事务提交成功后才调用 `CardTransactionNotificator`，传入账户、渠道、卡、授权、交易 ID 和阶段类型。通知失败仅写入 `NotificationError`，不会改变本地授权或清算结果；重试通知应直接使用这些已提交的 ID，而非再次模拟。
- 上层通知适配器负责异步排队、协议 DTO、签名、ACK 和投递状态。PingPong 的真实 webhook 合约仍待确认；可以显式使用 `NoopNotificator{}`，不能把它当成已经成功投递。
- 与开卡一样，模拟器必须拥有真实提交边界；不支持嵌套未提交事务，不提供持久化 outbox 或恰好一次通知保证。
