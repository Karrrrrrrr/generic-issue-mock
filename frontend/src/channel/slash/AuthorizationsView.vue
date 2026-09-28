<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from "vue";
import {
  NAlert,
  NButton,
  NCard,
  NCollapse,
  NCollapseItem,
  NDataTable,
  NEmpty,
  NInputNumber,
  NModal,
  NSpin,
  NTag,
  useMessage,
  type DataTableColumns,
} from "naive-ui";
import TableFilters, { type FilterField } from "@/channel/TableFilters.vue";
import { useTableFilters } from "@/channel/tableFilters";
import { useClientPagination } from "@/channel/pagination";
import { formatDateTime } from "@/channel/dateTime";
import { formatEnumLabel, renderAmountTag, renderEnumTag } from "@/channel/tableTags";
import { request } from "@/channel/shared";
import { accountApi } from "./api";

type Authorization = {
  id: string;
  account_id: string;
  account_name: string;
  card_id: string;
  status: string;
  amount: string;
  settled: string;
  reversed: string;
  refunded: string;
  remaining: string;
  currency: string;
  merchant_name: string;
  created_at: string;
};

type AuthorizationTransaction = {
  id: string;
  transaction_type: string;
  status: string;
  amount: string;
  currency: string;
  created_at: string;
};

type AuthorizationDetail = Authorization & {
  card_number: string;
  authorization_code: string;
  merchant_country: string;
  merchant_mcc: string;
  raw_payload: unknown;
  transactions: AuthorizationTransaction[];
};

type Operation = "clear" | "reverse" | "refund";

const baseURL = "/slash/ui";
const message = useMessage();
const loading = ref(false);
const rows = ref<Authorization[]>([]);
const pagination = useClientPagination(rows);
const selected = ref<Authorization>();
const detail = ref<AuthorizationDetail>();
const detailLoading = ref(false);
const detailError = ref("");
const saving = ref<Operation>();
const amounts = reactive<Record<Operation, number | null>>({
  clear: null,
  reverse: null,
  refund: null,
});
const transactions = computed(() => detail.value?.transactions ?? []);
const transactionPagination = useClientPagination(transactions);
const isOverCleared = computed(() => Number(detail.value?.remaining) < 0);
const hasRawPayload = computed(() => {
  const payload = detail.value?.raw_payload;
  return payload != null && (typeof payload !== "object" || Object.keys(payload).length > 0);
});
const rawPayload = computed(() => JSON.stringify(detail.value?.raw_payload, null, 2));
const operations = [
  {
    key: "clear",
    title: "清算",
    type: "primary",
    description: "扣减钱包余额，允许超额及多次清算。",
  },
  {
    key: "reverse",
    title: "撤销",
    type: "warning",
    description: "减少剩余授权金额，不产生钱包收支。",
  },
  {
    key: "refund",
    title: "退款",
    type: "info",
    description: "退回钱包余额，无需先有清算交易。",
  },
] as const;
const metrics = [
  {
    key: "amount",
    label: "授权金额",
  },
  {
    key: "settled",
    label: "已清算",
  },
  {
    key: "reversed",
    label: "已撤销",
  },
  {
    key: "refunded",
    label: "已退款",
  },
  {
    key: "remaining",
    label: "剩余授权金额",
  },
] as const;
let listRequest = 0;
let detailRequest = 0;

const columns: DataTableColumns<Authorization> = [
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
    render: (row) => renderAmountTag(row.settled, row.currency),
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
const transactionColumns: DataTableColumns<AuthorizationTransaction> = [
  {
    title: "交易 ID",
    key: "id",
    width: 310,
    ellipsis: {
      tooltip: true,
    },
  },
  {
    title: "类型",
    key: "transaction_type",
    width: 100,
    render: (row) => renderEnumTag(row.transaction_type, "transaction"),
  },
  {
    title: "金额",
    key: "amount",
    width: 170,
    render: (row) => renderAmountTag(row.amount, row.currency),
  },
  {
    title: "状态",
    key: "status",
    width: 110,
    render: (row) => renderEnumTag(row.status, "status"),
  },
  {
    title: "时间",
    key: "created_at",
    width: 190,
    render: (row) => formatDateTime(row.created_at),
  },
];
const filterFields: FilterField[] = [
  {
    key: "id",
    label: "授权 ID（精确）",
  },
  {
    key: "card_id",
    label: "卡 ID（精确）",
  },
  {
    key: "merchant_name",
    label: "商户名称（模糊）",
  },
  {
    key: "status",
    label: "授权状态",
    options: [
      {
        label: "已授权",
        value: "authorized",
      },
      {
        label: "已拒绝",
        value: "declined",
      },
      {
        label: "处理中",
        value: "pending",
      },
      {
        label: "已撤销",
        value: "void",
      },
    ],
  },
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
    pagination.value.onUpdatePage(1);
    void load();
  },
});

async function load() {
  const currentRequest = ++listRequest;
  loading.value = true;
  try {
    const response = await request.get<Authorization[]>(baseURL + "/authorization-balances", {
      params: appliedFilters.value,
    });
    if (currentRequest === listRequest) {
      rows.value = response.data;
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
  detail.value = undefined;
  amounts.clear = null;
  amounts.reverse = null;
  amounts.refund = null;
  transactionPagination.value.onUpdatePage(1);
  void loadDetail();
}

function closeDetail(shown: boolean) {
  if (shown || saving.value) return;
  ++detailRequest;
  selected.value = undefined;
  detail.value = undefined;
  detailError.value = "";
  detailLoading.value = false;
}

async function loadDetail() {
  const row = selected.value;
  if (!row) return;
  const currentRequest = ++detailRequest;
  detailLoading.value = true;
  detailError.value = "";
  try {
    const response = await request.get<AuthorizationDetail>(`${baseURL}/authorizations/${row.id}/detail`, {
      params: { account_id: row.account_id },
    });
    if (currentRequest === detailRequest) {
      detail.value = response.data;
    }
  } catch (error) {
    if (currentRequest === detailRequest) {
      detail.value = undefined;
      detailError.value = error instanceof Error ? error.message : "加载授权详情失败";
    }
  } finally {
    if (currentRequest === detailRequest) {
      detailLoading.value = false;
    }
  }
}

function validAmount(operation: Operation) {
  const amount = amounts[operation];
  return amount !== null && Number.isFinite(amount) && amount > 0;
}

async function submit(operation: Operation) {
  const row = detail.value;
  if (!row || saving.value || detailLoading.value || !validAmount(operation)) return;
  saving.value = operation;
  try {
    await request.post(`${baseURL}/authorizations/${row.id}/${operation}`, {
      amount: String(amounts[operation]),
    });
    amounts[operation] = null;
    message.success(`已创建${operations.find((item) => item.key === operation)?.title}交易`);
    await Promise.all([loadDetail(), load()]);
  } catch (error) {
    message.error(error instanceof Error ? error.message : "创建交易失败");
  } finally {
    saving.value = undefined;
  }
}

onMounted(load);
</script>

<template>
  <div class="page-heading">
    <div>
      <h1>授权管理</h1>
      <p>查看授权汇总和关联交易，从详情发起清算、撤销或退款。</p>
    </div>
    <n-button :loading="loading" @click="load">
      刷新列表
    </n-button>
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
      :pagination="pagination"
      :scroll-x="1900"
      :row-key="(row: Authorization) => row.id"
      table-layout="fixed"
      :loading="loading"
      :columns="columns"
      :data="rows"
    />
  </n-card>
  <n-modal
    :show="Boolean(selected)"
    preset="card"
    title="授权详情"
    class="authorization-detail-modal"
    :closable="!saving"
    :mask-closable="!saving"
    :close-on-esc="!saving"
    :content-style="{ overflowY: 'auto', maxHeight: 'calc(90dvh - 130px)' }"
    @update:show="closeDetail"
  >
    <template #header-extra>
      <n-button
        size="small"
        :loading="detailLoading"
        :disabled="Boolean(saving)"
        @click="loadDetail"
      >
        刷新详情
      </n-button>
    </template>
    <n-spin :show="detailLoading">
      <div class="authorization-detail-content">
        <n-alert v-if="detailError" type="error" title="详情加载失败">
          {{ detailError }}。可点击右上角「刷新详情」重试。
        </n-alert>
        <template v-else-if="detail">
          <section class="authorization-section">
            <div class="authorization-section-heading">
              <h2>授权概览</h2>
              <n-tag
                :bordered="true"
                :type="detail.status === 'declined' ? 'error' : detail.status === 'authorized' ? 'success' : 'default'"
              >
                {{ formatEnumLabel(detail.status, "status") }}
              </n-tag>
            </div>
            <div class="authorization-metrics">
              <div
                v-for="metric in metrics"
                :key="metric.key"
                class="authorization-metric"
                :class="{ 'authorization-metric-negative': metric.key === 'remaining' && isOverCleared }"
              >
                <span>{{ metric.label }}</span>
                <n-tag
                  :bordered="true"
                  :type="metric.key === 'remaining' && isOverCleared ? 'error' : 'info'"
                >
                  {{ detail[metric.key] }} {{ detail.currency }}
                </n-tag>
              </div>
            </div>
            <p class="authorization-hint">剩余授权金额 = 授权金额 − 已清算 − 已撤销；退款单独累计，不恢复授权额度。</p>
            <n-alert v-if="isOverCleared" type="warning" :show-icon="false">
              剩余授权金额为负数，按实际金额展示，不归零。仍可继续模拟清算。
            </n-alert>
            <dl class="authorization-info-grid">
              <div>
                <dt>授权 ID</dt>
                <dd>{{ detail.id }}</dd>
              </div>
              <div>
                <dt>所属账户</dt>
                <dd>
                  {{ detail.account_name || "—" }}
                  <span class="authorization-secondary-id">{{ detail.account_id }}</span>
                </dd>
              </div>
              <div>
                <dt>卡号</dt>
                <dd>{{ detail.card_number || "—" }}</dd>
              </div>
              <div>
                <dt>卡 ID</dt>
                <dd>{{ detail.card_id }}</dd>
              </div>
              <div>
                <dt>商户</dt>
                <dd>{{ detail.merchant_name || "—" }}</dd>
              </div>
              <div>
                <dt>授权时间</dt>
                <dd>{{ formatDateTime(detail.created_at) }}</dd>
              </div>
              <div>
                <dt>授权码</dt>
                <dd>{{ detail.authorization_code || "—" }}</dd>
              </div>
              <div>
                <dt>商户国家 / MCC</dt>
                <dd>{{ detail.merchant_country || "—" }} / {{ detail.merchant_mcc || "—" }}</dd>
              </div>
            </dl>
          </section>
          <section class="authorization-section">
            <div class="authorization-section-heading">
              <h2>创建后续交易</h2>
            </div>
            <p class="authorization-hint">每次操作生成一笔独立交易，金额必须为正数。创建后自动刷新汇总与交易记录。</p>
            <div class="authorization-operations">
              <form
                v-for="operation in operations"
                :key="operation.key"
                class="authorization-operation"
                @submit.prevent="submit(operation.key)"
              >
                <h3>{{ operation.title }}</h3>
                <p class="authorization-hint">{{ operation.description }}</p>
                <label :for="`authorization-${operation.key}`">{{ operation.title }}金额（{{ detail.currency }}）</label>
                <n-input-number
                  v-model:value="amounts[operation.key]"
                  :input-props="{ id: `authorization-${operation.key}` }"
                  :placeholder="`输入${operation.title}金额`"
                  :show-button="false"
                  :disabled="Boolean(saving) || detailLoading"
                />
                <n-button
                  attr-type="submit"
                  :type="operation.type"
                  ghost
                  :loading="saving === operation.key"
                  :disabled="Boolean(saving) || detailLoading || !validAmount(operation.key)"
                >
                  创建{{ operation.title }}
                </n-button>
              </form>
            </div>
          </section>
          <section class="authorization-section">
            <div class="authorization-section-heading">
              <h2>关联卡交易</h2>
              <span class="authorization-hint">共 {{ transactions.length }} 笔</span>
            </div>
            <n-data-table
              :columns="transactionColumns"
              :data="transactions"
              :row-key="(row: AuthorizationTransaction) => row.id"
              :pagination="transactionPagination"
              :scroll-x="880"
              :max-height="320"
              table-layout="fixed"
            />
          </section>
          <section class="authorization-section">
            <div class="authorization-section-heading">
              <h2>授权报文</h2>
            </div>
            <n-collapse v-if="hasRawPayload">
              <n-collapse-item title="已留存的原始报文 JSON" name="payload">
                <pre class="authorization-json">{{ rawPayload }}</pre>
              </n-collapse-item>
            </n-collapse>
            <n-empty v-else description="未留存原始授权报文" size="small" />
            <p class="authorization-hint">当前未记录同步授权回调的请求、响应及响应码，不将交易通知投递记录作为授权回调展示。</p>
          </section>
        </template>
        <div v-else class="authorization-detail-placeholder">正在加载授权详情…</div>
      </div>
    </n-spin>
  </n-modal>
</template>
