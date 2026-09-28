# Shared UI service 与通用仓储

## 接入边界

`shared/service` 提供独立的浏览器管理 DTO 和 service，`shared/biz/ui_*.go` 实现 UI 业务。当前没有修改现有渠道路由、service、前端或 `main` 来调用它们；后续确认需求后再迁移。

`shared/biz/*_repository.go` 和 `shared/data/*_repository.go` 是按资源划分的通用持久化接口与实现，不属于 UI。UI、OpenAPI 和共享卡组件均可注入同一套 Repo；OpenAPI 不应为了复用仓储而调用 UI service/usecase。

## Repo 设计约定

- 一个资源一套 Repo，例如 `CardRepo`，不保留 `UICardRepo`、`OpenAPICardRepo` 或重复的查询/锁定实现。
- 每个操作拥有独立请求，例如 `CardListRequest`、`CardCountRequest`、`CardFindRequest`、`CardCreateRequest`；只共享明确的过滤字段组，不复用上层 DTO。
- 输入使用内部主键、项目枚举、原生时间和金额类型。Repo 不转换渠道 ID、不产生协议信封、不决定 UI 分页默认值、不编排业务。
- `Channel` 是请求中的必需归属参数，不固定在 Repo 实例中。单资源查询和修改必须提供 `AccountID`；账户和渠道产品遵循各自的归属模型。
- 列表和计数使用同一组过滤条件。多值条件用切片和 `IN`，空切片表示不应用该筛选；时间和模糊查询等可选单值条件使用指针。上层负责验证显式零值并确定可访问范围。
- `List` 显式按 `ID DESC` 排序。`Count` 单独执行；响应所需关联由 ORM `Preload` 加载，不在循环中补查。
- `Exist` 与 `Find` 分开。授权的 `ExistForCard` 额外要求卡归属，不能用普通 `Exist` 替代该校验。锁定查询继续要求完整归属参数。
- 模拟归属解析的 `ExistForSimulation` / `FindForSimulation` 是已批准的渠道加主键查询例外，不用于绕过普通业务的账户限制。
- 每个方法执行一次数据库操作，通过 `repo.DB(ctx)` 加入上下文事务；事务编排、金额规则、状态检查、错误日志和通知留在业务层。
- 私有 Repo 实现都有编译期接口断言。所有 Repo 由 `shared.RegisterProviders` 注册一次，不依赖 `RegisterUIProviders`。

## 创建 UI service

已有注册 `*gorm.DB` 的 injector 时，先注册共享组件，再按需启用 UI 工厂。应用已经调用过 `RegisterProviders` 时不要重复注册。

```go
shared.RegisterProviders(injector)
shared.RegisterUIProviders(injector)

factory := do.MustInvoke[*service.Factory](injector)
management, err := factory.New(&service.NewRequest{
    Channel: enums.Channel_PingPong,
    Notificator: biz.NoopNotificator{},
})
```

处理构造错误后，使用 `management.ListCards` 等具名方法及各自请求。工厂实例可以创建不同渠道的 service，构造时复制配置，各实例不会共享可变渠道状态。上面的 Noop 仅适用于明确不投递的场景，不代表 PingPong webhook 已实现或投递成功。

`RegisterUIProviders` 只注册业务和服务工厂，不注册 Repo，也不注册 HTTP 路由。OpenAPI 可以在只调用 `RegisterProviders` 后直接注入 `biz.CardRepo` 等资源仓储，无需依赖 UI。

## 当前提供的能力

- 账户：列表、创建、改名和余额调整；创建账户不创建或复制产品。
- 产品、持卡人、钱包：列表查询。
- 虚拟账户：列表、创建、充值、提现及账户内转账。
- 卡：列表、详情、状态修改、充值和提现；仅激活卡可调资，支持三个通用钱包类型及其正确资金来源。
- 授权及交易：列表、详情、阶段记录；授权剩余金额保留负数。
- 模拟：授权、清算、退款、撤销及基于原交易的阶段操作，统一调用 `CardTransactionSimulator`。
- 调资记录：查询、请求幂等、双方余额快照；资金转移复用 `BalanceChanger`，不绕过正常余额检查。
- Webhook：通用配置 CRUD、记录查询；事件选项和重放由上层适配器提供。

这些是共享能力，不代表每个渠道都应暴露全部接口。路由接入时仍需遵守渠道能力边界，例如不向 Paynda 暴露虚拟账户，不向 OpenAPI 暴露模拟接口。本次没有补入同步授权回调配置。

## 类型与差异化配置

- UI ID 使用 `model.ID`，JSON 数字；可选 ID 用指针。请求键和 webhook 原始协议引用仍为字符串。
- UI 状态、交易类型及调资种类使用项目 `enums`；时间使用 `time.Time` / `*time.Time`，金额使用 `decimal.Decimal`。账户关联响应包含 `account_name`。
- 列表响应为 `items` 和 `total`，分页默认第 1 页、每页 20 条，上限 200 条。HTTP 成功信封由后续接入的适配层负责。
- `CreateVirtualAccountOnAccountCreation` 是显式的可选构造参数；默认不自动建虚拟账户，也不引入默认产品。
- `CardTransactionNotificator`、`UIWebhookEventCatalog`、`UIWebhookReplayer` 由上层实现。缺少通知适配器时拒绝模拟；缺少事件目录或重放适配器时，对应操作返回未支持错误，不伪造事件或投递成功。
- 渠道 webhook 的报文、签名、ACK 和实际发送仍在渠道适配器中，不下沉到 Repo 或通用 UI 流程。

## 当前验证范围

需求梳理期间仅执行格式化和编译检查，不自动运行单元测试或集成测试；编译通过不代表现有渠道已经完成 Shared UI 接入。
