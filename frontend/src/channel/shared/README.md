# 共享管理页面

## 枚举与接口边界

UI 使用项目内部枚举，不使用第三方状态值：卡状态为 `inactive/active/freezing/frozen/deleting/deleted`，交易状态为 `pending/authorized/succeed/failed/void`，交易类型为 `auth/clear/void/refund/verification/fund_in/fund_out`。前端类型及筛选项定义在 `../enums.ts`，后端 UI DTO 直接使用 `backend/enums`。OpenAPI 和 webhook 报文继续遵守渠道协议，Webhook 事件列表仍从各渠道接口加载。

`api.ts` 为 Slash、Paynda、PhotonPay 的相同管理契约提供参数化接口实现；渠道文件只传入自己的 UI 路径。`../pingpong/management.ts` 处理 PingPong 的字段、分页、数值金额和请求 ID 差异。组件不通过渠道名分支，不把第三方状态转换逻辑搬到前端。

## 页面与能力

- `ChannelShell.vue`：四渠道共用布局，菜单、标题和当前渠道由上层传入。
- `AccountsView.vue`：账户列表、创建、调整余额；仅在提供 `accountApi.update` 时展示改名。账户/资金账户标题、钱包 ID 列和说明可配置。
- `CardholdersView.vue`：共用持卡人列表与远程分页，仅接入已有此能力的三个渠道。
- `CardProductsView.vue`：四渠道共用只读产品列表，展示渠道级产品 ID 与 BIN；候选 BIN 文案和说明由上层提供。开卡必须指定产品，不提供默认产品标记或兜底选择。
- `VirtualAccountsView.vue`：Slash、PhotonPay、PingPong 共用列表、创建和充值。未提供 `withdraw` 时不展示转出；币种选择、钱包 ID 列和说明可配置。
- `CardsView.vue`：四渠道共用卡列表、内部状态筛选、冻结/恢复和资金操作。可选到期日、冻结余额、虚拟账户列及模拟授权入口；不按虚拟账户关联关系猜测钱包。PingPong 的卡资金操作仍只在卡钱包与所属虚拟账户之间进行。
- `AuthorizationsView.vue`：四渠道共用列表及清算/撤销/退款。仅在提供 `api.detail` 时展示关联交易、原始报文等详情；PingPong 从本地授权与阶段记录返回真实汇总、卡信息和关联交易。PingPong 仍展示 `contract_pending`，不伪造通知已投递。
- `TransactionsView.vue`：四渠道共用内部交易枚举、操作条件及退款表单。PingPong 查询本地交易记录，不依赖第三方 OpenAPI 交易报表。
- `WebhooksView.vue`、`WebhookRecordsView.vue`：三个渠道共用配置、投递详情和重放。事件选项由接口提供，筛选能力按现有接口开放；PingPong 不开放未实现的通知能力。

领域数据契约定义在 `contracts.ts`。差异通过 typed API、可选操作、少量展示属性和说明 slot 提供，不设计万能 CRUD 组件。账户名称使用原响应的 `account_name`；只有账户选择表单加载账户选项。

需要幂等键的资金操作和后续模拟由上层传入 `newRequestID`，打开操作时生成；失败重试保留当前键，成功后为下一次操作生成新键。PingPong 的授权/虚拟账户分页接口在适配层收集完整结果，供现有共享页面进行客户端分页；账户、卡和资金订单仍使用服务端分页。

## 模拟组件

- `CardTransactionSimulator.vue` 提供授权和退款页，接收卡列表、授权、退款的调用函数。Slash、Paynda、PhotonPay 的页面仅提供这些适配并转发 `completed`。
- `CardAuthorizationForm.vue` 负责卡选择、商户信息、金额校验和授权提交。PingPong 从卡管理传入单张卡，隐藏选择器及其协议不需要的商户字段。
- `TransactionStageForm.vue` 负责清算、撤销、退款的金额表单。只校验有限正数，不按余额或授权剩余金额限制清算。授权详情、PingPong 后续交易、Slash 交易退款复用该表单。
- `useSimulation.ts` 统一提交锁、成功提示、错误提示。请求失败保留输入，成功后由组件发出 `completed`，调用方刷新列表或关闭弹窗；`busy` 用于禁止关闭弹窗或并发提交。
- `simulation.ts` 定义共享输入类型、操作说明和金额校验。

共享组件不判断渠道、不拼接渠道 URL、不生成外部 ID 或幂等键。调用方负责字段映射及协议差异，例如金额的字符串/数值类型和 PingPong 的 `request_id`。适配函数必须在请求失败时抛出错误，不得吞掉错误后当作提交成功。

模拟请求不传账户 ID。授权使用卡 ID；清算、撤销和关联退款使用授权 ID；独立退款才传卡 ID 和币种。已有交易入口传原交易 ID，由后端定位授权。PingPong 也接入授权、独立退款和关联退款；授权详情与交易页面提供后续阶段操作，不改变 `contract_pending` 通知语义。
