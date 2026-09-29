<script setup lang="ts">
import { h, onMounted, ref } from "vue";
import {
  NButton,
  NDataTable,
  useMessage,
  type DataTableColumns,
} from "naive-ui";
import TableFilters, { type FilterField } from "@/channel/TableFilters.vue";
import { useTableFilters } from "@/channel/tableFilters";
import { useRemotePagination } from "@/channel/pagination";
import { formatDateTime } from "@/channel/dateTime";
import { renderAmountTag, renderEnumTag } from "@/channel/tableTags";
import type { ManagementAPI } from "./api";
import type { Authorization, AuthorizationAPI } from "./contracts";
import { transactionStatusOptions } from "@/channel/enums";
import AuthorizationDetailModal from "@/channel/shared/AuthorizationDetailModal.vue";

const {
  api,
  accountApi,
  detailedFilters = true,
  newRequestId,
  showNotificationStatus = false,
} = defineProps<{
  api: AuthorizationAPI;
  accountApi: Pick<ManagementAPI["accountApi"], "listAll">;
  detailedFilters?: boolean;
  newRequestId?: () => string;
  showNotificationStatus?: boolean;
}>();
const message = useMessage();
const loading = ref(false);
const rows = ref<Authorization[]>([]);
const { page, pageSize, total, pagination } = useRemotePagination(load);
const selected = ref<Authorization>();
let listRequest = 0;

const columns: DataTableColumns<Authorization> = [
  ...(showNotificationStatus ? [{
    title: "通知状态",
    key: "notification_status",
    width: 220,
    render: (row: Authorization) => renderEnumTag(row.notification_status ?? "", "notification"),
  }] : []),
  {
    title: "授权 ID",
    key: "id",
    width: 310,
    render: (row) => h(
      NButton,
      {
        text: true,
        type: "primary",
        class: "authorization-id-link",
        onClick: () => openDetail(row),
      },
      { default: () => row.id },
    ),
  },
  {
    title: "账户名称",
    key: "account_name",
    width: 150,
    ellipsis: {
      tooltip: true,
    },
  },
  {
    title: "卡 ID",
    key: "card_id",
    width: 310,
    ellipsis: {
      tooltip: true,
    },
  },
  {
    title: "状态",
    key: "status",
    width: 110,
    render: (row) => renderEnumTag(row.status, "status"),
  },
  {
    title: "授权金额",
    key: "amount",
    width: 150,
    render: (row) => renderAmountTag(row.amount, row.currency),
  },
  {
    title: "已清算",
    key: "settled",
    width: 150,
    render: (row) => row.settled === undefined ? "—" : renderAmountTag(row.settled, row.currency),
  },
  {
    title: "剩余授权金额",
    key: "remaining",
    width: 170,
    render: (row) => renderAmountTag(row.remaining, row.currency),
  },
  {
    title: "商户",
    key: "merchant_name",
    width: 160,
    ellipsis: {
      tooltip: true,
    },
  },
  {
    title: "授权时间",
    key: "created_at",
    width: 190,
    render: (row) => formatDateTime(row.created_at),
  },
  {
    title: "操作",
    key: "actions",
    width: 100,
    fixed: "right",
    render: (row) => h(
      NButton,
      {
        size: "small",
        secondary: true,
        onClick: () => openDetail(row),
      },
      { default: () => "详情" },
    ),
  },
];
const filterFields: FilterField[] = detailedFilters ? [
  { key: "id", label: "授权 ID（精确）" },
  { key: "card_id", label: "卡 ID（精确）" },
  { key: "merchant_name", label: "商户名称（模糊）" },
  { key: "status", label: "授权状态", options: transactionStatusOptions },
] : [{ key: "card_id", label: "卡 ID（精确）" }];
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
  const currentRequest = ++listRequest;
  loading.value = true;
  try {
    const response = await api.list({
      ...appliedFilters.value,
      page_number: page.value,
      page_size: pageSize.value,
    });
    if (currentRequest === listRequest) {
      rows.value = response.data;
      total.value = response.total_items;
    }
  } catch (error) {
    if (currentRequest === listRequest) {
      message.error(error instanceof Error ? error.message : "加载授权失败");
    }
  } finally {
    if (currentRequest === listRequest) {
      loading.value = false;
    }
  }
}

function openDetail(row: Authorization) {
  selected.value = row;
}

onMounted(load);
</script>

<template>
  <div class="page-heading">
    <div>
      <h1>授权管理</h1>
      <p>查看授权汇总和关联交易，从详情发起清算、撤销或退款。</p>
    </div>
  </div>
  <slot name="description" />
  <TableFilters
    v-model:values="filters"
    v-model:date-range="dateRange"
    :fields="filterFields"
    :show-date-range="detailedFilters"
    :account-options="accountOptions"
    :accounts-loading="accountsLoading"
    :loading="loading"
    @load-accounts="loadAccounts"
    @search="search"
    @reset="reset"
  />
  <n-data-table
    max-height="max(160px, calc(100dvh - 580px))"
    :pagination="pagination"
    :scroll-x="1900"
    :row-key="(row: Authorization) => row.id"
    table-layout="fixed"
    :loading="loading"
    :columns="columns"
    :data="rows"
  />
  <AuthorizationDetailModal
    v-model="selected"
    :api="api"
    :new-request-id="newRequestId"
    @refreshed="load"
  />
</template>
