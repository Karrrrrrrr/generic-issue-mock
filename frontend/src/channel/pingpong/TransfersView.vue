<script setup lang="ts">
import { ref } from "vue";
import {
  NButton,
  NCard,
  NDataTable,
  NInput,
  NSpace,
} from "naive-ui";
import { formatDateTime } from "@/channel/dateTime";
import { renderAmountTag, renderEnumTag } from "@/channel/tableTags";
import { type Transfer } from "./api";
import { useList } from "./list";
const { rows, loading, filters, pagination, search } = useList<Transfer>("transfers");
const accountFilter = ref("");
const columns = [
  {
    title: "记录 ID",
    key: "id",
    width: 100
  },
  {
    title: "账户",
    key: "account_name",
    width: 160
  },
  {
    title: "请求编号",
    key: "request_id",
    width: 280,
    ellipsis: {
      tooltip: true
    }
  },
  {
    title: "类型",
    key: "kind",
    width: 130,
    render: (row: Transfer) => renderEnumTag(row.kind, "walletTransfer")
  },
  {
    title: "金额",
    key: "amount",
    width: 160,
    render: (row: Transfer) => renderAmountTag(row.amount, row.currency)
  },
  {
    title: "来源钱包",
    key: "source_wallet_id",
    width: 110
  },
  {
    title: "目标钱包",
    key: "target_wallet_id",
    width: 110
  },
  {
    title: "时间",
    key: "created_at",
    width: 180,
    render: (row: Transfer) => formatDateTime(row.created_at)
  },
];
function query() {
  filters.account_id = accountFilter.value || undefined;
  search();
}
</script>

<template>
  <section class="ping-page">
    <div class="page-heading">
      <h1>
        资金订单
      </h1>
    </div>
    <n-card :bordered="false">
      <n-space class="ping-toolbar">
        <n-input
          v-model:value="accountFilter"
          placeholder="账户 ID"
          clearable
          style="width: 180px"
        />
        <n-button @click="query">
          查询
        </n-button>
      </n-space>
      <n-data-table
        :bordered="true"
        remote
        :columns="columns"
        :data="rows"
        :loading="loading"
        :pagination="pagination"
        :scroll-x="1330"
      />
    </n-card>
  </section>
</template>
