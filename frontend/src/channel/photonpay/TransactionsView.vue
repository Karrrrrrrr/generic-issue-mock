<script setup lang="ts">
import { renderAmountTag, renderEnumTag } from "@/channel/tableTags";
import { h, onMounted, ref } from "vue";
import { createDiscreteApi, NButton, NCard, NDataTable, NSpace, NTag } from "naive-ui";
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

function actions(r: Transaction) {
  if (r.transaction_type === "auth" && r.status === "authorized")
    return [
      h(
        NButton,
        { size: "small", onClick: () => apply(r.id, "reverse") },
        { default: () => "撤销" },
      ),
    ];
  if (r.transaction_type === "clear" && r.status === "succeed")
    return [
      h(
        NButton,
        { size: "small", onClick: () => apply(r.id, "refund") },
        { default: () => "退款" },
      ),
    ];
  return [h(NTag, { size: "small" }, { default: () => "已处理" })];
}

const columns = [
  {
    title: "账户名称",
    key: "account_name",
  },
  {
    title: "金额",
    key: "amount",
    render: (row: Transaction) => renderAmountTag(row.amount, row.currency),
  },
  { title: "商户", key: "merchant_name" },
  {
    title: "类型",
    key: "transaction_type",
    render: (row: Transaction) => renderEnumTag(row.transaction_type, "transaction"),
  },
  {
    title: "状态",
    key: "status",
    render: (row: Transaction) => renderEnumTag(row.status, "status"),
  },
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
      <p>查看交易并执行撤销或退款；清算请到授权管理。</p>
    </div>
    <n-button @click="load">刷新</n-button>
  </div>
  <n-card :bordered="false">
    <n-data-table
      :scroll-x="1000"
      table-layout="fixed"
      :columns="columns"
      :data="rows"
    />
  </n-card>
</template>
