<script setup lang="ts">
import { computed, reactive, watch } from "vue";
import {
  NAlert,
  NButton,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NSelect,
  useMessage,
} from "naive-ui";
import {
  isPositiveSimulationAmount,
  type AuthorizationSimulationRequest,
  type SimulationCard,
} from "./simulation";
import { useSimulation } from "./useSimulation";

const props = withDefaults(
  defineProps<{
    cards: SimulationCard[];
    authorize: (request: AuthorizationSimulationRequest) => Promise<unknown>;
    selectCard?: boolean;
    merchantDetails?: boolean;
    disabled?: boolean;
  }>(),
  {
    selectCard: true,
    merchantDetails: true,
    disabled: false,
  },
);
const emit = defineEmits<{
  completed: [];
  busy: [value: boolean];
}>();
const message = useMessage();
const { submitting, result, submit } = useSimulation();
const form = reactive({
  cardID: null as number | null,
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
const card = computed(() => props.cards.find((item) => item.id === form.cardID));
const currencyOptions = computed(() => [...new Set([
  "USD", "EUR", "GBP", ...props.cards.map((item) => item.currency),
])].map((currency) => ({
  label: currency,
  value: currency,
})));

watch(() => props.cards, (cards) => {
  if (!cards.some((item) => item.id === form.cardID && !item.disabled)) {
    form.cardID = cards.find((item) => !item.disabled)?.id ?? null;
  }
}, { immediate: true });
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
    cardID: card.value.id,
    amount: form.amount,
    currency: form.currency,
    merchantName: form.merchantName.trim(),
    merchantMCC: form.merchantMCC.trim(),
    merchantCountry: form.merchantCountry.trim(),
    merchantCity: form.merchantCity.trim(),
  };
  if (await submit(() => props.authorize(request), "授权已创建，可在授权管理中清算、撤销或退款。")) {
    emit("completed");
  }
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
          v-model:value="form.cardID"
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
</template>
