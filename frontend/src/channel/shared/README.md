# 共享模拟组件

- `CardTransactionSimulator.vue` 提供授权和退款页，接收卡列表、授权、退款的调用函数。Slash、Paynda、PhotonPay 的页面仅提供这些适配并转发 `completed`。
- `CardAuthorizationForm.vue` 负责卡选择、商户信息、金额校验和授权提交。PingPong 从卡管理传入单张卡，隐藏选择器及其协议不需要的商户字段。
- `TransactionStageForm.vue` 负责清算、撤销、退款的金额表单。只校验有限正数，不按余额或授权剩余金额限制清算。授权详情、PingPong 后续交易、Slash 交易退款复用该表单。
- `useSimulation.ts` 统一提交锁、成功提示、错误提示。请求失败保留输入，成功后由组件发出 `completed`，调用方刷新列表或关闭弹窗；`busy` 用于禁止关闭弹窗或并发提交。
- `simulation.ts` 定义共享输入类型、操作说明和金额校验。

共享组件不判断渠道、不拼接渠道 URL、不生成外部 ID 或幂等键。调用方负责字段映射及协议差异，例如金额的字符串/数值类型和 PingPong 的 `request_id`。适配函数必须在请求失败时抛出错误，不得吞掉错误后当作提交成功。

模拟请求不传账户 ID。授权使用卡 ID；清算、撤销和关联退款使用授权 ID；独立退款才传卡 ID 和币种。已有交易入口传原交易 ID，由后端定位授权。PingPong 仅复用已有的授权及后续交易能力，不增加独立退款入口，也不改变 `contract_pending` 通知语义。
