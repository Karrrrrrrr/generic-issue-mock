import { h } from "vue";
import { NTag } from "naive-ui";
const labels: Record<string, string> = {
  ACTIVE: "正常",
  REVOKED: "冻结",
  CANCELLED: "已注销",
  INACTIVE: "待激活",
  SUCCESS: "成功",
  authorized: "已授权",
  contract_pending: "协议待补齐 · 未投递",
  card_top_up: "卡充值",
  card_withdraw: "卡转出",
  virtual_account_top_up: "虚拟账户充值",
  virtual_account_transfer: "虚拟账户划转",
};
export function renderTag(value: string) {
  return h(NTag, {
    bordered: true,
    size: "small",
    type: value === "contract_pending" ? "warning" : "info",
    style: {
      backgroundColor: "transparent"
    },
  }, {
    default: () => labels[value] || value
  });
}
