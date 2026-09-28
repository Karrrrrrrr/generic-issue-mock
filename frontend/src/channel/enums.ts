export type CardStatus = "inactive" | "active" | "freezing" | "frozen" | "deleting" | "deleted";
export type TransactionStatus = "pending" | "authorized" | "succeed" | "failed" | "void";
export type TransactionType = "auth" | "clear" | "void" | "refund" | "verification" | "fund_in" | "fund_out";
export type WebhookDeliveryStatus = "pending" | "succeeded" | "failed";
export type WalletTransferKind = "card_top_up" | "card_withdraw" | "virtual_account_top_up" | "virtual_account_transfer";

export const cardStatusOptions: { label: string; value: CardStatus }[] = [
  { label: "正常", value: "active" },
  { label: "未激活", value: "inactive" },
  { label: "冻结中", value: "freezing" },
  { label: "已冻结", value: "frozen" },
  { label: "删除中", value: "deleting" },
  { label: "已删除", value: "deleted" },
];

export const transactionStatusOptions: { label: string; value: TransactionStatus }[] = [
  { label: "处理中", value: "pending" },
  { label: "已授权", value: "authorized" },
  { label: "成功", value: "succeed" },
  { label: "失败", value: "failed" },
  { label: "已撤销", value: "void" },
];

export const transactionTypeOptions: { label: string; value: TransactionType }[] = [
  { label: "授权", value: "auth" },
  { label: "清算", value: "clear" },
  { label: "撤销", value: "void" },
  { label: "退款", value: "refund" },
  { label: "验证", value: "verification" },
  { label: "资金转入", value: "fund_in" },
  { label: "资金转出", value: "fund_out" },
];
