<script setup lang="ts">
import { h, onMounted, ref } from "vue";
import {
  NButton,
  NCard,
  NDataTable,
  NSpace,
  NTag,
  createDiscreteApi,
} from "naive-ui";
import { api } from "./api";
import type { Transaction } from "@/channel/types";
const { message } = createDiscreteApi(["message"]);
const rows = ref<Transaction[]>([]);
async function load() {
  rows.value = (await api.listTransactions()).data;
}
async function apply(id: string, action: "clear" | "reverse" | "refund") {
  try {
    await api.applyTransactionStep(id, action);
    await load();
  } catch (e) {
    message.error(e instanceof Error ? e.message : "操作失败");
  }
}
function actions(row: Transaction) {
  if (row.transaction_type === "auth" && row.status === "authorized")
    return [
      h(
        NButton,
        {
          size: "small",
          type: "primary",
          onClick: () => apply(row.id, "clear"),
        },
        { default: () => "清算" },
      ),
      h(
        NButton,
        { size: "small", onClick: () => apply(row.id, "reverse") },
        { default: () => "撤销" },
      ),
    ];
  if (row.transaction_type === "clear" && row.status === "succeed")
    return [
      h(
        NButton,
        {
          size: "small",
          type: "warning",
          onClick: () => apply(row.id, "refund"),
        },
        { default: () => "退款" },
      ),
    ];
  return [h(NTag, { size: "small" }, { default: () => "已处理" })];
}
const columns = [
  {
    title: "金额",
    key: "amount",
    render: (r: Transaction) => `${r.currency} ${r.amount}`,
  },
  { title: "商户", key: "merchant_name" },
  { title: "类型", key: "transaction_type" },
  { title: "状态", key: "status" },
  {
    title: "操作",
    key: "actions",
    render: (r: Transaction) => h(NSpace, null, { default: () => actions(r) }),
  },
];
onMounted(() => void load());
</script>
<template>
  <div class="page-heading">
    <div>
      <h1>交易处理</h1>
      <p>按交易状态执行清算、撤销或退款</p>
    </div>
    <n-button @click="load">刷新</n-button>
  </div>
  <n-card :bordered="false">
    <n-data-table :columns="columns" :data="rows" />
  </n-card>
</template>
