<script setup lang="ts">
import { computed, reactive, ref, watch } from "vue";
import {
  NAlert,
  NButton,
  NCollapse,
  NCollapseItem,
  NDataTable,
  NEmpty,
  NModal,
  NSpin,
  NTag,
  type DataTableColumns,
} from "naive-ui";
import { formatDateTime } from "@/channel/dateTime";
import { useClientPagination } from "@/channel/pagination";
import { formatEnumLabel, renderAmountTag, renderEnumTag } from "@/channel/tableTags";
import TransactionStageOperations from "@/channel/shared/TransactionStageOperations.vue";
import type { Authorization, AuthorizationAPI, AuthorizationTransaction } from "./contracts";
import type { SimulationStage } from "@/channel/shared/simulation";

const props = defineProps<{
  api: AuthorizationAPI;
  modelValue?: Authorization;
  newRequestId?: () => string;
}>();

const emit = defineEmits<{
  "update:modelValue": [value: Authorization | undefined];
  refreshed: [];
}>();

const requestIds = reactive<Partial<Record<SimulationStage, string>>>({});
const detail = ref<Authorization>();
const detailLoading = ref(false);
const detailError = ref("");
const saving = ref(false);
const transactions = computed(() => detail.value?.transactions ?? []);
const transactionPagination = useClientPagination(transactions);
const isOverCleared = computed(() => Number(detail.value?.remaining) < 0);
const hasRawPayload = computed(() => {
  const payload = detail.value?.raw_payload;
  return payload != null && (typeof payload !== "object" || Object.keys(payload).length > 0);
});
const rawPayload = computed(() => JSON.stringify(detail.value?.raw_payload, null, 2));
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
let detailRequest = 0;

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

watch(
  () => props.modelValue,
  (row) => {
    ++detailRequest;
    if (!row) {
      detail.value = undefined;
      detailError.value = "";
      detailLoading.value = false;
      return;
    }
    resetRequestIds();
    void loadDetail();
  },
);

function resetRequestIds() {
  requestIds.clear = props.newRequestId?.();
  requestIds.reverse = props.newRequestId?.();
  requestIds.refund = props.newRequestId?.();
}

function closeDetail(shown: boolean) {
  if (shown || saving.value) {
    return;
  }
  ++detailRequest;
  emit("update:modelValue", undefined);
  detail.value = undefined;
  detailError.value = "";
  detailLoading.value = false;
}

async function loadDetail() {
  const row = props.modelValue;
  if (!row) {
    return;
  }
  const currentRequest = ++detailRequest;
  detailLoading.value = true;
  detailError.value = "";
  try {
    const response = props.api.detail ? await props.api.detail(row) : row;
    if (currentRequest === detailRequest) {
      detail.value = response;
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

async function simulateStage(stage: SimulationStage, amount: number) {
  const row = detail.value;
  if (!row) {
    throw new Error("请先选择授权");
  }
  await props.api.stage({
    authorization: row,
    stage,
    amount,
    requestId: requestIds[stage],
  });
  requestIds[stage] = props.newRequestId?.();
}

async function refreshSimulation() {
  await loadDetail();
  emit("refreshed");
}
</script>

<template>
  <n-modal
    :show="Boolean(modelValue)"
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
        @click="refreshSimulation"
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
                :type="detail.status === 'failed' ? 'error' : detail.status === 'authorized' ? 'success' : 'default'"
              >
                {{ formatEnumLabel(detail.status, "status") }}
              </n-tag>
            </div>
            <div class="authorization-metrics">
              <div
                v-for="metric in metrics.filter((item) => detail?.[item.key] !== undefined)"
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
              <div v-if="api.detail">
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
              <div v-if="api.detail">
                <dt>授权码</dt>
                <dd>{{ detail.authorization_code || "—" }}</dd>
              </div>
              <div v-if="api.detail">
                <dt>商户国家 / MCC</dt>
                <dd>{{ detail.merchant_country || "—" }} / {{ detail.merchant_mcc || "—" }}</dd>
              </div>
            </dl>
          </section>
          <TransactionStageOperations
            :operation-key="detail.id"
            :currency="detail.currency"
            :disabled="saving || detailLoading || detail.status !== 'authorized'"
            :simulate="simulateStage"
            @busy="saving = $event"
            @completed="refreshSimulation"
          />
          <section v-if="api.detail" class="authorization-section">
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
          <section v-if="api.detail" class="authorization-section">
            <div class="authorization-section-heading">
              <h2>授权报文</h2>
            </div>
            <n-collapse v-if="hasRawPayload">
              <n-collapse-item title="已留存的原始报文 JSON" name="payload">
                <pre class="authorization-json">{{ rawPayload }}</pre>
              </n-collapse-item>
            </n-collapse>
            <n-empty v-else description="未留存原始授权报文" size="small" />
            <p class="authorization-hint">同步授权失败时会留存失败来源、原因、请求报文和响应信息。</p>
          </section>
        </template>
        <div v-else class="authorization-detail-placeholder">正在加载授权详情…</div>
      </div>
    </n-spin>
  </n-modal>
</template>
