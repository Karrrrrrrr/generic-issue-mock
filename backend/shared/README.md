# 通用卡业务组件

`shared/biz` 处理已经归一化的开卡、模拟授权、清算、退款和撤销请求，不解释任何渠道协议，也不在内部查找“默认渠道账户”。渠道层负责解析外部 ID、选择产品/虚拟账户、校验协议参数，以及将共享错误映射为渠道错误码。

Slash、Paynda、PhotonPay、PingPong 的 UI 模拟授权、清算、退款和撤销，以及 PhotonPay 的 OpenAPI sandbox 已接入 `CardTransactionSimulator`。各渠道不再自行写入模拟授权、交易阶段或修改钱包；UI 和 OpenAPI 仍各自拥有请求、协议适配、操作日志及通知适配，不互相调用。`CardIssuer` 尚未替换各渠道开卡入口。

## 渠道接入约定

- Slash、Paynda、PhotonPay 的 `/ui/simulate/authorizations`、`/ui/simulate/refunds` 和交易阶段操作必须显式传入 `account_id`；授权详情上的阶段操作和 PingPong 原有入口保留其账户参数。前端直接使用所选卡/交易返回的账户 ID，不额外查询账户列表。
- 授权详情操作先按账户读取授权以取得卡 ID，交易操作先按账户读取原交易；共享模拟器在自己的事务内再次锁定并验证账户、卡、授权、币种及钱包。渠道不得在外层再包事务。
- UI 独立退款省略 `authorization_id`，关联退款传正数渠道 ID；显式零值、空字符串和跨账户/跨卡引用不能当成独立退款。
- Slash 由 service 实现 `CardTransactionNotificator`，在边界转换 UUID 后交给原 webhook 投递器；Paynda、PhotonPay UI 的通知适配器按账户读取已提交交易并调用原投递流程。原来 service/渠道流程末尾的重复发送已移除。
- PingPong 明确传入 `NoopNotificator`，保持 `contract_pending`，不虚构投递记录。PhotonPay sandbox 保留原来不发送 UI webhook 的行为，也显式使用 `NoopNotificator`；没有借迁移新增同步授权回调。
- PingPong 和 PhotonPay sandbox 将必填 `request_id` / `requestId` 交给共享模拟器处理幂等，重放不重复记账或通知。三个其他 UI 协议原来没有请求 ID，此次不新增。
- 所有渠道通过各自独立 `errors` 包把共享错误转为已有渠道错误。Slash、Paynda、PhotonPay 的旧业务错误从 `biz` 迁入 `errors`，不保留别名，不改变状态码、reason 或 message。

## 依赖与注册

- `CardIssuer`：通用开卡入口。
- `CardTransactionSimulator`：仅暴露模拟授权、清算、退款和撤销四个方法的接口；具体实现 `cardTransactionSimulator` 不导出。请求、`Validate`、实现和复用的私有方法集中在 `biz/card_transaction_simulator.go`，不再拆出只有一次调用的流程方法。
- `BalanceChanger`：通用余额变更接口，具体实现 `balanceChanger` 不导出，编译期断言保证接口实现完整。请求、`Validate` 和实现集中在 `biz/balance_changer.go`。
- `AccountRepo`、`CardRepo`、`CardHolderRepo`、`CardProductRepo`、`VirtualAccountRepo`、`WalletRepo`、`AuthorizationRepo`、`CardTransactionRepo`：按资源拆分，只包含当前共享业务需要的仓储操作。
- `Transaction`：`InTx` 管理真实提交边界，`IsInTx` 判断当前 context 是否已经携带事务。
- `Notificator`：由上层实现并在每次开卡请求中传入。通用层不负责 webhook DTO、签名、HTTP 调用、队列或重试策略。
- `CardTransactionNotificator`：只负责交易通知，不要求实现开卡通知接口。
- 记账规则对所有渠道一致。模拟器决定操作金额、冻结释放上限和流程，由唯一的 `BalanceChanger` 实现钱包金额变更；它不是可按渠道切换的记账策略。

已有一个注册了根数据库连接 `*gorm.DB` 的 `samber/do` injector 时：

```go
shared.RegisterProviders(injector)
issuer := do.MustInvoke[*biz.CardIssuer](injector)
simulator := do.MustInvoke[biz.CardTransactionSimulator](injector)
balanceChanger := do.MustInvoke[biz.BalanceChanger](injector)
```

`RegisterProviders` 直接注册构造器，包括可选用的 GORM Gen 仓储实现。若上层已有其他存储适配，可按需注册 `biz.NewCardIssuer` / `biz.NewCardTransactionSimulator` / `biz.NewBalanceChanger`，自行提供对应仓储接口和 `Transaction`，不必引用 `shared/data`。模拟器依赖 `BalanceChanger`；其余查询仓储仍由模拟器明确声明。

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

## 通用余额变更

`BalanceChanger` 提供五个方法，全部返回 `error`。参数使用本项目的 `AccountID`、`Channel`、`WalletID` 和 `Currency`，资源 ID 为 `model.ID`（int64），不复制其他项目的商户、外部账户或交易归属字段。余额实体继续使用 `model.Wallet`，没有新增一套 `BalanceChangeBalance`。

| 方法 | 请求 | 金额变更 |
| --- | --- | --- |
| `ChangeBalance` | `BalanceChangeReq` | 显式应用 `AvailableDelta`、`PendingOutDelta`、`InDelta`、`OutDelta` |
| `TryBalanceChange` | `TccBalanceChangeReq` | `Available -= 请求金额`，`PendingOut += 请求金额` |
| `ConfirmBalanceChange` | `ConfirmBalanceChangeReq` | `Available += ReleaseAmount - 请求 Amount`，冻结减 `ReleaseAmount`，`Out` 加请求金额 |
| `CancelBalanceChange` | `TccBalanceChangeReq` | 冻结减请求金额，等额加回可用余额，`In` / `Out` 不变 |
| `ChangeBalanceSimple` | `ChangeBalanceSimpleReq` | 正数增加余额和 `In`，负数扣余额并按绝对值增加 `Out`；不改冻结 |

- 每种请求自行实现 `Validate()`。归属、钱包 ID 和币种必填；Try/Confirm/Cancel 金额必须为正，Simple 金额必须非零，原始 delta 请求不能四项全零。
- Confirm 使用单独的请求，是因为项目允许超额及持续清算，实际扣款额与释放冻结额不一定相等。`ReleaseAmount` 是必填金额值，允许为 0，不得为负或超过本次扣款额。授权剩余额度由模拟器计算，不在余额组件中查询。
- `CheckAvailable=true` 时拒绝变更后的可用余额为负；false 允许负可用余额。授权传 true，模拟清算、退款和撤销传 false；正常钱包资金转出应传 true，不能借用模拟清算的例外。
- 冻结余额和累计 `In` / `Out` 不能变成负数，不受 `CheckAvailable` 影响。释放额超出钱包实际冻结额时返回命名业务错误，不会扣别的字段补齐，也不会把错误数据截成零。
- 所有读写都按 `(AccountID, Channel, WalletID)` 限定，锁定钱包后再次验证币种并根据当前值计算。模拟器预加载卡的钱包只用于检查归属、类型和币种，不使用预加载快照记账。
- 已有事务时，组件通过 `Transaction.IsInTx(ctx)` 自动识别并复用，不提交调用方的事务；没有事务时自己开启并提交。因此不提供 `IsInTx` 请求开关，也不把 GORM 句柄放进 biz 请求。模拟器里的交易阶段和余额修改仍能一起回滚。
- TCC 表示冻结、确认、取消的金额操作，不引入 DTM、TCC 状态表或自动幂等。调用两次组件就可能变更两次；请求重放校验、授权归属、冻结释放上限、交易记录和通知由上层处理。模拟器已经保留这些控制。

例如普通扣款（无冻结）：

```go
err := balanceChanger.ChangeBalanceSimple(ctx, &biz.ChangeBalanceSimpleReq{
	AccountID:      accountID,
	Channel:        channel,
	WalletID:       walletID,
	Currency:       currency,
	Amount:         decimal.NewFromInt(-10),
	CheckAvailable: true,
})
if err != nil {
	return err
}
```

## 模拟卡交易

`CardTransactionSimulator` 的四个入口 `SimulateAuthorization`、`SimulateClearing`、`SimulateRefund` 和 `SimulateReversal` 均返回 `(*CardTransactionSimulationResult, error)`。各入口先调用请求的 `Validate()`，随后直接在自己的事务回调内实现流程。结果包含授权记录、当前交易阶段、带符号的授权剩余金额、是否重放，以及提交后的通知错误；独立退款的 `Authorization` 为 nil、`Remaining` 为零。

```go
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
	Notificator:     notificator,
})
if err != nil {
	return err
}
```

- 退款的 `AuthorizationID` 是可选指针。nil 表示独立退款，持久化为 0，不创建或查找虚构授权；非 nil 必须是同账户、同渠道、同一卡的正数授权 ID，传入 0 不会被当作省略。
- 独立退款要求卡处于有效状态；关联已有授权的退款与撤销不因卡被冻结而拒绝。两种退款均不要求存在先前清算，也不限制为已清算金额。
- 关联退款的请求币种必须与授权一致；未提供的商户字段继承授权，明确提供的字段按请求记录。独立退款使用请求的币种和商户字段。
- 退款增加卡实际消费钱包的 `Available` 和 `In`，不释放占款、不增加或扣减授权剩余金额。
- 撤销的 `AuthorizationID` 必填，币种、商户信息及原授权交易关联从授权继承。撤销扣减授权剩余金额，保留负数，将实际释放的冻结金额转回可用余额，不增加总余额。
- 撤销请求的 `Status` 必填，仅允许 `TransactionStatus_VOID` 或 `TransactionStatus_SUCCEED`。上层按已有渠道约定传入：Slash、Paynda、PhotonPay 使用 VOID，PingPong 使用 SUCCEED。交易 `Type` 始终为 VOID，共享层不硬编码渠道判断。

### 统一记账规则

所有渠道使用同一套规则。旧实现中授权不冻结等行为是不一致的实现，不作为渠道差异保留，也不再由调用方传入策略。模拟器分别调用 `BalanceChanger` 的 Try、Confirm、Simple、Cancel 方法实现授权、清算、退款、撤销；撤销可释放金额为零时不调用余额组件，但仍按原流程保存撤销阶段。

- 钱包 `Available` 是可用余额（数据库列为 `available`），`PendingOut` 是冻结余额，`Available + PendingOut` 才是总余额。展示可用余额或检查可转出金额时直接使用 `Available`，不能再减一次 `PendingOut`。请求中的 `Amount` 仍表示本次操作金额，不是钱包余额。
- 授权：检查 `Available` 是否足够；不足则拒绝且不发成功通知。成功从 `Available` 扣减授权金额并等额增加 `PendingOut`，不增加 `Out`，总余额不变。
- 清算：释放 `min(本次清算金额, max(授权剩余金额, 0))` 的冻结金额。`PendingOut -= 释放额`，`Available += 释放额 - 清算金额`，`Out += 清算金额`。冻结覆盖的部分不重复扣减可用余额，超出冻结的部分直接扣可用余额，总余额减少清算金额。
- 退款：直接增加 `Available` 和 `In`。独立退款与关联授权的退款都不修改 `PendingOut`，也不改变授权剩余金额；退款不是扣款。
- 撤销：释放 `min(本次撤销金额, max(授权剩余金额, 0))` 的冻结金额，等额加回 `Available`，总余额不变，不改变 `In` 或 `Out`；授权剩余金额扣除本次撤销金额。
- 清算不检查余额或授权剩余额度，允许超额清算、负钱包余额和负授权剩余金额。清算和撤销释放冻结的数量不会超过本授权的正数剩余额度，不会在剩余为负时继续释放其他授权的占款。
- 同账户、渠道、请求 ID 的成功重放不重复冻结、扣款、退款或解冻；所有钱包修改与当前交易阶段在同一事务内提交。

例如初始可用余额 100、冻结 0：授权 30 后为可用 70、冻结 30、总额 100；清算 20 后为可用 70、冻结 10、总额 80；撤销剩余 10 后为可用 80、冻结 0、总额 80；退款 5 后为可用 85、冻结 0、总额 85。

旧流程生成但没有冻结金额，或冻结时未从可用余额转出的授权数据不满足新规则的前提。本次不自动修补历史钱包数据，也不保留旧钱包 `amount` 列的兼容映射；验证时应初始化独立测试库并新建授权，不能把旧钱包余额直接视作新规则下的可用余额。

### 数据与事务边界

- 四种模拟金额均必须为正数。账户、卡 ID 必填且为正数；清算和撤销的授权 ID 必填，退款的授权 ID 可选。可选请求 ID 和商户字段使用指针，非 nil 的空字符串会被拒绝。
- 授权要求卡有效，卡、钱包和请求币种一致。清算必须引用同一账户、同一渠道、同一卡的有效授权，币种和商户信息从授权继承；卡被冻结后仍可清算已有授权。
- 总是使用 `Card.WalletID` 对应的钱包。`single`、`virtual_account_single` 使用卡钱包，`share` 使用虚拟账户钱包；不会因为卡关联 VA 就改扣 VA 或根账户的钱包。
- 账户行锁串行化同账户的共享模拟请求；授权、卡及钱包按需加行锁。账户、授权、卡、钱包和历史阶段查询均明确限制归属。
- 授权记录、AUTH/CLEAR/REFUND/VOID 阶段和钱包变更在各自操作的同一事务内提交，任一步失败全部回滚。关联操作保留授权交易关联，剩余金额扣除成功清算及已撤销金额，不受退款或失败阶段影响。
- 清算不会以剩余金额大于零作为前提，也不会把负剩余金额截成零。四种模拟能力均由共享组件执行；渠道仅保留入口转换和通知适配。重放授权返回当前阶段记录，PingPong 的剩余金额展示不会恢复为原始授权金额。

### 请求重放与通知

- `RequestID == nil` 表示新模拟操作，不进行去重。显式传入请求 ID 后，以 `(AccountID, Channel, RequestID)` 在共享模拟器内去重；授权、清算、退款、撤销共用此键空间，应使用不同的键。
- 同键同业务参数返回原交易并标记 `Replayed=true`，不再次记账、创建记录或调用通知。不同卡、阶段、授权引用、金额、币种/商户信息或撤销状态会返回冲突。独立退款重放不查询授权，也不会重复入账；重放结果的 `Remaining` 是查询时的当前剩余金额，不是首次调用快照。
- 并发去重依赖事务中的账户行锁；自定义仓储必须真正加锁，其他直接写入交易的代码不能绕过同一锁约定。没有新增数据库唯一索引或修改模型。
- 事务提交成功后才调用 `CardTransactionNotificator`，传入账户、渠道、卡、授权、交易 ID 和阶段类型。通知失败仅写入 `NotificationError`，不会改变本地授权或清算结果；重试通知应直接使用这些已提交的 ID，而非再次模拟。
- 上层通知适配器负责异步排队、协议 DTO、签名、ACK 和投递状态。PingPong 的真实 webhook 合约仍待确认；可以显式使用 `NoopNotificator{}`，不能把它当成已经成功投递。
- 与开卡一样，模拟器必须拥有真实提交边界；不支持嵌套未提交事务，不提供持久化 outbox 或恰好一次通知保证。
