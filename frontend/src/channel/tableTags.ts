import { h } from "vue";
import { NTag, type TagProps } from "naive-ui";

type EnumLabel = {
  label: string;
  type?: TagProps["type"];
};

const statusLabels: Record<string, EnumLabel> = {
  active: {
    label: "正常",
    type: "success",
  },
  normal: {
    label: "正常",
    type: "success",
  },
  inactive: {
    label: "未激活",
    type: "default",
  },
  paused: {
    label: "已冻结",
    type: "warning",
  },
  freezing: {
    label: "冻结中",
    type: "warning",
  },
  frozen: {
    label: "已冻结",
    type: "warning",
  },
  closed: {
    label: "已关闭",
    type: "default",
  },
  cancelled: {
    label: "已注销",
    type: "default",
  },
  deleted: {
    label: "已删除",
    type: "default",
  },
  pending: {
    label: "处理中",
    type: "warning",
  },
  authorized: {
    label: "已授权",
    type: "info",
  },
  posted: {
    label: "已入账",
    type: "success",
  },
  succeed: {
    label: "成功",
    type: "success",
  },
  succeeded: {
    label: "成功",
    type: "success",
  },
  failed: {
    label: "失败",
    type: "error",
  },
  void: {
    label: "已撤销",
    type: "default",
  },
};

const transactionLabels: Record<string, EnumLabel> = {
  auth: {
    label: "授权",
    type: "info",
  },
  clear: {
    label: "清算",
    type: "success",
  },
  void: {
    label: "撤销",
    type: "warning",
  },
  refund: {
    label: "退款",
    type: "warning",
  },
  verification: {
    label: "验证",
    type: "info",
  },
  "transaction.authentication.approved": {
    label: "授权",
    type: "info",
  },
  "transaction.authentication.settled": {
    label: "清算",
    type: "success",
  },
  "transaction.authentication.reversal.settled": {
    label: "撤销",
    type: "warning",
  },
  "transaction.refund.settled": {
    label: "退款",
    type: "warning",
  },
};

const eventLabels: Record<string, EnumLabel> = {
  auth: {
    label: "交易授权",
  },
  verification: {
    label: "交易验证",
  },
  void: {
    label: "交易撤销",
  },
  refund: {
    label: "交易退款",
  },
  card_status_update: {
    label: "卡片状态更新",
  },
  cardholder_status_update: {
    label: "持卡人状态更新",
  },
  card_transaction: {
    label: "卡交易通知",
  },
  card_status: {
    label: "卡片状态通知",
  },
  "aggregated_transaction.create": {
    label: "交易创建",
  },
  "aggregated_transaction.update": {
    label: "交易更新",
  },
  "card_creation.event": {
    label: "卡片创建",
  },
  "card.update": {
    label: "卡片更新",
  },
  "card.delete": {
    label: "卡片删除",
  },
};

const enumLabels: Record<string, Record<string, EnumLabel>> = {
  status: statusLabels,
  transaction: transactionLabels,
  event: eventLabels,
  funding: {
    卡资金: {
      label: "卡资金",
      type: "info",
    },
    虚拟账户共享资金: {
      label: "虚拟账户共享资金",
      type: "info",
    },
  },
  enabled: {
    true: {
      label: "启用",
      type: "success",
    },
    false: {
      label: "停用",
      type: "default",
    },
  },
  defaultProduct: {
    true: {
      label: "是",
      type: "success",
    },
    false: {
      label: "否",
      type: "default",
    },
  },
};

type EnumCategory = "status" | "transaction" | "event" | "funding" | "enabled" | "defaultProduct";

export function formatEnumLabel(value: string | boolean, category: EnumCategory) {
  return enumLabels[category]?.[String(value).toLowerCase()]?.label ?? "未知";
}

export function renderEnumTag(
  value: string | boolean,
  category: EnumCategory,
) {
  const rawValue = String(value);
  const entry = enumLabels[category]?.[rawValue.toLowerCase()];

  return h(
    NTag,
    {
      size: "small",
      bordered: true,
      type: entry?.type ?? "default",
      title: rawValue,
    },
    {
      default: () => formatEnumLabel(value, category),
    },
  );
}

export function renderAmountTag(amount: string | number, currency: string) {
  return h(
    NTag,
    {
      size: "small",
      bordered: true,
      class: "amount-tag",
    },
    {
      default: () => `${amount} ${currency}`,
    },
  );
}
