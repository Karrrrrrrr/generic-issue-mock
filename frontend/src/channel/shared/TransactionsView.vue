<script setup lang="ts">
import { formatDateTime } from "@/channel/dateTime";
import TableFilters, { type FilterField } from "@/channel/TableFilters.vue";
import { useTableFilters } from "@/channel/tableFilters";
import type { ManagementAPI } from "./api";
import { useRemotePagination } from "@/channel/pagination";
import { renderAmountTag, renderEnumTag } from "@/channel/tableTags";
import { h, onMounted, ref } from "vue";
import {
  createDiscreteApi,
  NButton,
  NDataTable,
  NSpace,
} from "naive-ui";

import type { Transaction } from "@/channel/types";
import { transactionStatusOptions, transactionTypeOptions } from "@/channel/enums";
import AuthorizationDetailModal from "@/channel/shared/AuthorizationDetailModal.vue";
import type { Authorization, AuthorizationAPI } from "@/channel/shared/contracts";

import type { ChannelAPI } from "@/channel/types";
const {
  api,
  accountApi,
  authorizationApi,
  newRequestId,
} = defineProps<{
  api: Pick<ChannelAPI, "listTransactions">;
  authorizationApi: AuthorizationAPI;
  accountApi: Pick<ManagementAPI["accountApi"], "listAll">;
  newRequestId?: () => string;

}>();

const { message } = createDiscreteApi(["message"]);
const rows = ref<Transaction[]>([]);
const loading = ref(false);
const { page, pageSize, total, pagination } = useRemotePagination(load);
const selectedAuthorization = ref<Authorization>();

const filterFields: FilterField[] = [
  {
    "key": "id",
    "label": "交易 ID（精确）"
  },
  {
    "key": "card_id",
    "label": "卡 ID（精确）"
  },
  {
    "key": "authorization_id",
    "label": "授权 ID（精确）"
  },
  {
    "key": "transaction_type",
    "label": "交易类型",
    options: transactionTypeOptions
  },
  {
    "key": "status",
    "label": "交易状态",
    options: transactionStatusOptions
  }
];
const {
  filters,
  dateRange,
  appliedFilters,
  accountOptions,
  accountsLoading,
  loadAccounts,
  search,
  reset,
} = useTableFilters({
  loadAccounts: () => accountApi.listAll(),
  onSearch: () => {
    page.value = 1;
    void load();
  },
});

async function load() {
  loading.value = true;
  try {
    const response = await api.listTransactions({
      ...appliedFilters.value,
      page_number: page.value,
      page_size: pageSize.value,
    });
    rows.value = response.data;
    total.value = response.total_items;
  } catch (error) {
    message.error(error instanceof Error ? error.message : "加载交易失败");
  } finally {
    loading.value = false;
  }
}

function openOperation(row: Transaction) {
  if (!row.authorization_id) {
    message.warning("这笔交易没有关联授权，不能打开授权详情");
    return;
  }
  selectedAuthorization.value = {
    id: row.authorization_id,
    account_id: row.account_id,
    account_name: row.account_name,
    card_id: row.card_id,
    status: row.status,
    amount: row.amount,
    remaining: row.amount,
    currency: row.currency,
    merchant_name: row.merchant_name,
    created_at: row.transacted_at,
  };
}

function actions(row: Transaction) {
  return [
    h(
      NButton,
      {
        size: "small",
        type: "primary",
        ghost: true,
        onClick: () => openOperation(row),
      },
      { default: () => "操作" },
    ),
  ];
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
  {
    title: "交易 ID",
    key: "id",
  },
  {
    title: "卡 ID",
    key: "card_id",
  },
  {
    title: "授权 ID",
    key: "authorization_id",
  },
  { title: "商户", key: "merchant_name" },
  {
    title: "交易时间",
    key: "transacted_at",
    width: 180,
    render: (row: Transaction) => formatDateTime(row.transacted_at),
  },
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
      <p>从交易记录发起清算、撤销或退款，具体合法性由后端校验。</p>
    </div>
  </div>
  <TableFilters
    v-model:values="filters"
    v-model:date-range="dateRange"
    :fields="filterFields"
    :account-options="accountOptions"
    :accounts-loading="accountsLoading"
    :loading="loading"
    @load-accounts="loadAccounts"
    @search="search"
    @reset="reset"
  />
  <n-data-table
    max-height="max(160px, calc(100dvh - 580px))"
    remote
    :pagination="pagination"
    :scroll-x="1800"
    table-layout="fixed"
    :columns="columns"
    :data="rows"
    :loading="loading"
  />
  <AuthorizationDetailModal
    v-model="selectedAuthorization"
    :api="authorizationApi"
    :new-request-id="newRequestId"
    @refreshed="load"
  />
</template>
