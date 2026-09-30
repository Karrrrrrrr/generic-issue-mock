<script setup lang="ts">
import { computed, reactive, ref, watch } from "vue";
import {
  NAlert,
  NButton,
  NCollapse,
  NCollapseItem,
  NDescriptions,
  NDescriptionsItem,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NSelect,
  NTag,
  useMessage,
} from "naive-ui";
import {
  isPositiveSimulationAmount,
  type AuthorizationSimulationResult,
  type AuthorizationSimulationRequest,
  type SimulationCard,
  type SimulationResult,
} from "./simulation";
import { useSimulation } from "./useSimulation";

const props = withDefaults(
  defineProps<{
    cards: SimulationCard[];
    authorize: (request: AuthorizationSimulationRequest) => Promise<SimulationResult | undefined>;
    initialCardId?: number | null;
    selectCard?: boolean;
    merchantDetails?: boolean;
    disabled?: boolean;
  }>(),
  {
    selectCard: true,
    merchantDetails: true,
    disabled: false,
    initialCardId: null,
  },
);
const emit = defineEmits<{
  completed: [];
  busy: [value: boolean];
}>();
const message = useMessage();
const { submitting, result, submitData } = useSimulation();
const authorizationResult = ref<AuthorizationSimulationResult | null>(null);
const form = reactive({
  cardId: null as number | null,
  amount: 100 as number | null,
  currency: "USD",
  merchantName: "Amazon",
  merchantMCC: "5411",
  merchantCountry: "US",
  merchantCity: "Seattle",
});
const options = computed(() => props.cards.map((card) => ({
  label: card.label,
  value: card.id,
  disabled: card.disabled,
})));
const card = computed(() => props.cards.find((item) => item.id === form.cardId));
const currencyOptions = computed(() => [...new Set([
  "USD", "EUR", "GBP", ...props.cards.map((item) => item.currency),
])].map((currency) => ({
  label: currency,
  value: currency,
})));
const failedSideLabel = computed(() => {
  switch (authorizationResult.value?.failed_side) {
    case "mock":
      return "我方失败";
    case "third_party":
      return "三方失败";
    default:
      return "未授权";
  }
});
const failedSideType = computed(() => (
  authorizationResult.value?.failed_side === "mock" ? "warning" : "error"
));
const resultPanelClass = computed(() => {
  if (authorizationResult.value?.approved) {
    return "authorization-result-approved";
  }
  if (authorizationResult.value?.failed_side === "mock") {
    return "authorization-result-mock";
  }
  return "authorization-result-third-party";
});
const resultTitle = computed(() => (
  authorizationResult.value?.approved ? "授权成功" : "授权失败"
));
const resultSummary = computed(() => (
  authorizationResult.value?.message || authorizationResult.value?.reason || "—"
));

watch(() => props.cards, (cards) => {
  const initial = cards.find((item) => item.id === props.initialCardId && !item.disabled);
  if (initial) {
    form.cardId = initial.id;
    return;
  }
  if (!cards.some((item) => item.id === form.cardId && !item.disabled)) {
    form.cardId = cards.find((item) => !item.disabled)?.id ?? null;
  }
}, { immediate: true });
watch(() => props.initialCardId, (cardId) => {
  const initial = props.cards.find((item) => item.id === cardId && !item.disabled);
  if (initial) {
    form.cardId = initial.id;
  }
});
watch(card, (selected) => {
  if (selected) {
    form.currency = selected.currency;
  }
}, { immediate: true });
watch(submitting, (busy) => emit("busy", busy), { flush: "sync" });

async function submitAuthorization() {
  if (props.disabled || submitting.value) {
    return;
  }
  if (!card.value || card.value.disabled || !isPositiveSimulationAmount(form.amount)) {
    message.warning("请选择可用卡并填写正数金额");
    return;
  }
  if (!form.merchantName.trim() || !form.currency || (
    props.merchantDetails && (!form.merchantMCC.trim() || !form.merchantCountry.trim())
  )) {
    message.warning("请填写币种和必填商户信息");
    return;
  }
  const request: AuthorizationSimulationRequest = {
    cardId: card.value.id,
    amount: form.amount,
    currency: form.currency,
    merchantName: form.merchantName.trim(),
    merchantMCC: form.merchantMCC.trim(),
    merchantCountry: form.merchantCountry.trim(),
    merchantCity: form.merchantCity.trim(),
  };
  authorizationResult.value = null;
  const data = await submitData(
    () => props.authorize(request),
    "授权已创建，可在授权管理中清算、撤销或退款。",
    {
      isSuccess: (item) => !item?.authorization_result || item.authorization_result.approved,
      failureMessage: (item) => item?.authorization_result?.message || "授权失败",
    },
  );
  authorizationResult.value = data?.authorization_result ?? null;
  if (!data) {
    return;
  }
  emit("completed");
}

function formatPayload(value: unknown) {
  if (value === undefined || value === null || value === "") {
    return "—";
  }
  if (typeof value === "string") {
    try {
      return JSON.stringify(JSON.parse(value), null, 2);
    } catch {
      return value;
    }
  }
  return JSON.stringify(value, null, 2);
}
</script>

<template>
  <n-form
    label-placement="top"
    :disabled="submitting || disabled"
    @submit.prevent="submitAuthorization"
  >
    <div class="form-grid">
      <n-form-item v-if="selectCard" class="form-wide" label="卡" required>
        <n-select
          v-model:value="form.cardId"
          :options="options"
          filterable
          placeholder="选择卡"
        />
      </n-form-item>
      <n-form-item label="交易金额" required>
        <n-input-number
          v-model:value="form.amount"
          :min="0.01"
          :precision="2"
          style="width: 100%"
        />
      </n-form-item>
      <n-form-item label="币种" required>
        <n-select
          v-model:value="form.currency"
          :options="currencyOptions"
          :disabled="!selectCard || submitting || disabled"
        />
      </n-form-item>
      <n-form-item label="商户名称" required>
        <n-input v-model:value="form.merchantName" placeholder="例如 Amazon" />
      </n-form-item>
      <template v-if="merchantDetails">
        <n-form-item label="MCC" required>
          <n-input v-model:value="form.merchantMCC" placeholder="例如 5411" />
        </n-form-item>
        <n-form-item label="商户国家" required>
          <n-input v-model:value="form.merchantCountry" placeholder="例如 US" />
        </n-form-item>
        <n-form-item label="地区">
          <n-input v-model:value="form.merchantCity" placeholder="例如 Seattle" />
        </n-form-item>
      </template>
    </div>
    <n-button
      attr-type="submit"
      type="primary"
      block
      :loading="submitting"
      :disabled="disabled || submitting"
    >
      模拟授权
    </n-button>
  </n-form>
  <n-alert v-if="result" type="success" :show-icon="false" style="margin-top: 16px">
    {{ result }}
  </n-alert>
  <section
    v-if="authorizationResult"
    class="authorization-result"
    :class="resultPanelClass"
  >
    <div class="authorization-result-header">
      <div class="authorization-result-state">
        <span class="authorization-result-dot" />
        <div class="authorization-result-title-group">
          <div class="authorization-result-title">
            {{ resultTitle }}
          </div>
          <div class="authorization-result-summary">
            {{ resultSummary }}
          </div>
        </div>
      </div>
      <div class="authorization-result-tags">
        <n-tag
          v-if="!authorizationResult.approved"
          size="small"
          :type="failedSideType"
          round
        >
          {{ failedSideLabel }}
        </n-tag>
      </div>
    </div>
    <div class="authorization-result-meta">
      <span>
        <em>金额</em>
        <strong>{{ authorizationResult.amount }} {{ authorizationResult.currency }}</strong>
      </span>
      <span>
        <em>HTTP 状态</em>
        <strong>{{ authorizationResult.status_code || "—" }}</strong>
      </span>
      <span>
        <em>失败来源</em>
        <strong>{{ authorizationResult.approved ? "—" : failedSideLabel }}</strong>
      </span>
    </div>
    <n-descriptions
      :column="2"
      size="small"
      class="authorization-result-descriptions"
    >
      <n-descriptions-item label="卡 ID">
        {{ authorizationResult.card_id || "—" }}
      </n-descriptions-item>
      <n-descriptions-item label="账户 ID">
        {{ authorizationResult.account_id || "—" }}
      </n-descriptions-item>
      <n-descriptions-item label="金额">
        {{ authorizationResult.amount }} {{ authorizationResult.currency }}
      </n-descriptions-item>
      <n-descriptions-item label="原因">
        {{ authorizationResult.reason || "—" }}
      </n-descriptions-item>
      <n-descriptions-item label="商户">
        {{ authorizationResult.merchant_name || "—" }}
      </n-descriptions-item>
      <n-descriptions-item label="MCC / 国家">
        {{ authorizationResult.merchant_mcc || "—" }} / {{ authorizationResult.merchant_country || "—" }}
      </n-descriptions-item>
      <n-descriptions-item label="授权 URL" :span="2">
        {{ authorizationResult.target_url || "—" }}
      </n-descriptions-item>
      <n-descriptions-item label="HTTP 状态" :span="2">
        {{ authorizationResult.status_code || "—" }}
      </n-descriptions-item>
    </n-descriptions>
    <n-collapse
      class="authorization-result-collapse"
      arrow-placement="right"
    >
      <n-collapse-item
        title="授权请求/响应报文"
        name="payload"
      >
        <div class="authorization-result-payloads">
          <div class="authorization-result-payload-group">
            <div class="authorization-result-payload-group-title">
              请求
            </div>
            <div class="authorization-result-payload">
              <div class="authorization-result-payload-title">
                Body
              </div>
              <pre>{{ formatPayload(authorizationResult.request_payload) }}</pre>
            </div>
            <div class="authorization-result-payload">
              <div class="authorization-result-payload-title">
                Headers
              </div>
              <pre>{{ formatPayload(authorizationResult.request_headers) }}</pre>
            </div>
          </div>
          <div class="authorization-result-payload-group">
            <div class="authorization-result-payload-group-title">
              响应
            </div>
            <div class="authorization-result-payload">
              <div class="authorization-result-payload-title">
                Body
              </div>
              <pre>{{ formatPayload(authorizationResult.response_body) }}</pre>
            </div>
            <div class="authorization-result-payload">
              <div class="authorization-result-payload-title">
                Headers
              </div>
              <pre>{{ formatPayload(authorizationResult.response_headers) }}</pre>
            </div>
          </div>
        </div>
      </n-collapse-item>
    </n-collapse>
  </section>
</template>

<style scoped>
.authorization-result {
  --authorization-result-accent: #d03050;
  display: grid;
  gap: 12px;
  margin-top: 16px;
  padding: 12px 14px;
  border: 1px solid var(--app-border);
  border-left-width: 3px;
  border-radius: 8px;
  background: var(--app-surface);
  color: var(--app-text);
  transition: var(--theme-color-transition), var(--theme-surface-transition);
}

.authorization-result-approved {
  --authorization-result-accent: #18a058;
  border-left-color: #18a058;
}

.authorization-result-mock {
  --authorization-result-accent: #f0a020;
  border-left-color: #f0a020;
}

.authorization-result-third-party {
  --authorization-result-accent: #d03050;
  border-left-color: #d03050;
}

.authorization-result-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.authorization-result-state {
  display: flex;
  min-width: 0;
  align-items: flex-start;
  gap: 10px;
}

.authorization-result-dot {
  width: 8px;
  height: 8px;
  flex: none;
  margin-top: 7px;
  border-radius: 50%;
  background: var(--authorization-result-accent);
}

.authorization-result-title-group {
  min-width: 0;
}

.authorization-result-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--app-text);
}

.authorization-result-summary {
  margin-top: 2px;
  overflow-wrap: anywhere;
  color: var(--app-text-muted);
  font-size: 13px;
  line-height: 1.45;
}

.authorization-result-tags {
  flex: none;
}

.authorization-result-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 18px;
  padding-top: 10px;
  border-top: 1px solid var(--app-border);
}

.authorization-result-meta span {
  display: inline-flex;
  min-width: 0;
  align-items: baseline;
  gap: 6px;
}

.authorization-result-meta em,
.authorization-result-payload-title {
  color: var(--app-text-muted);
  font-size: 12px;
  font-style: normal;
}

.authorization-result-meta strong {
  overflow-wrap: anywhere;
  color: var(--app-text);
  font-size: 13px;
  font-weight: 600;
}

.authorization-result-descriptions {
  width: 100%;
}

.authorization-result-collapse {
  border-top: 1px solid var(--app-border);
}

.authorization-result-payloads {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.authorization-result-payload-group {
  min-width: 0;
}

.authorization-result-payload-group-title {
  margin-bottom: 8px;
  color: var(--app-text);
  font-size: 13px;
  font-weight: 600;
}

.authorization-result-payload {
  min-width: 0;
}

.authorization-result-payload + .authorization-result-payload {
  margin-top: 10px;
}

.authorization-result-payload-title {
  margin-bottom: 5px;
}

.authorization-result-payload pre {
  min-height: 84px;
  max-height: 180px;
  margin: 0;
  overflow: auto;
  padding: 9px 10px;
  border: 1px solid var(--app-border);
  border-radius: 6px;
  background: var(--app-surface-muted);
  color: var(--app-text);
  font-size: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
  transition: var(--theme-color-transition), var(--theme-surface-transition);
}

@media (max-width: 720px) {
  .authorization-result-header {
    align-items: flex-start;
  }

  .authorization-result-payloads {
    grid-template-columns: 1fr;
  }
}
</style>
