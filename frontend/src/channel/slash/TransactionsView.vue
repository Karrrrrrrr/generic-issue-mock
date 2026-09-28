<script setup lang="ts">
import { formatDateTime } from "@/channel/dateTime";
import TableFilters, { type FilterField } from "@/channel/TableFilters.vue";
import { useTableFilters } from "@/channel/tableFilters";
import { accountApi } from "./api";
import { useRemotePagination } from "@/channel/pagination";
import { renderAmountTag, renderEnumTag } from "@/channel/tableTags";
import { computed, h, onMounted, ref } from "vue";
import {
  createDiscreteApi,
  NButton,
  NCard,
  NDataTable,
  NForm,
  NFormItem,
  NInputNumber,
  NModal,
  NSpace,
  NTag,
} from "naive-ui";
import { api, applyTransactionAmount } from "./api";
import type { Transaction } from "@/channel/types";

const { message } = createDiscreteApi(["message"]);
const rows = ref<Transaction[]>([]);
const loading = ref(false);
const { page, pageSize, total, pagination } = useRemotePagination(load);
const refunding = ref<Transaction | null>(null);
const refundAmount = ref<number | null>(null);
const refundVisible = computed({
  get: () => refunding.value !== null,
  set: (value) => {
    if (!value) refunding.value = null;
  },
});

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
    "options": [
      {
        "label": "授权",
        "value": "auth"
      },
      {
        "label": "清算",
        "value": "clear"
      },
      {
        "label": "撤销",
        "value": "void"
      },
      {
        "label": "退款",
        "value": "refund"
      }
    ]
  },
  {
    "key": "status",
    "label": "交易状态",
    "options": [
      {
        "label": "处理中",
        "value": "pending"
      },
      {
        "label": "已入账",
        "value": "posted"
      },
      {
        "label": "失败",
        "value": "failed"
      }
    ]
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

async function apply(transaction: Transaction, action: "clear" | "reverse" | "refund") {
  try {
    await api.applyTransactionStep(transaction, action);
    await load();
  } catch (e) {
    message.error(e instanceof Error ? e.message : "操作失败");
  }
}

function openRefund(row: Transaction) {
  refunding.value = row;
  refundAmount.value = Number(row.amount);
}

async function submitRefund() {
  if (!refunding.value || !refundAmount.value || refundAmount.value <= 0) return;
  try {
    await applyTransactionAmount(refunding.value, "refund", refundAmount.value);
    refunding.value = null;
    await load();
  } catch (e) {
    message.error(e instanceof Error ? e.message : "退款失败");
  }
}

function actions(row: Transaction) {
  if (row.transaction_type === "auth" && row.status === "authorized")
    return [
      h(
        NButton,
        { size: "small", onClick: () => apply(row, "reverse") },
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
          onClick: () => openRefund(row),
        },
        { default: () => "退款" },
      ),
    ];
  return [h(
    NTag,
    {
      size: "small",
      bordered: true,
    },
    { default: () => "已处理" },
  )];
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
      <p>按交易状态执行撤销或退款；清算请到授权管理。</p>
    </div>
    <n-button @click="load">刷新</n-button>
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
  <n-card :bordered="false">
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
  </n-card>
  <n-modal
    v-model:show="refundVisible"
    preset="card"
    title="创建退款"
    style="width: min(440px, calc(100vw - 32px))"
  >
    <n-form label-placement="top">
      <n-form-item label="原交易"
        ><span
          >{{ refunding?.merchant_name }} · {{ refunding?.currency }} {{ refunding?.amount }}</span
        ></n-form-item
      >
      <n-form-item label="退款金额">
        <n-input-number
          v-model:value="refundAmount"
          :min="0.01"
          :max="Number(refunding?.amount || 0)"
          :precision="2"
          style="width: 100%"
        />
      </n-form-item>
    </n-form>
    <template #action>
      <n-space justify="end">
        <n-button @click="refunding = null">取消</n-button>
        <n-button type="warning" @click="submitRefund">创建退款</n-button>
      </n-space>
    </template>
  </n-modal>
</template>
