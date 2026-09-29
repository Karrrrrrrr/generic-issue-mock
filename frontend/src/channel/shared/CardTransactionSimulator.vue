<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import {
  NAlert,
  NButton,
  NCard,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NSelect,
  NSpace,
  NTabPane,
  NTabs,
  useMessage,
} from "naive-ui";
import type { Card } from "@/channel/types";
import CardAuthorizationForm from "./CardAuthorizationForm.vue";
import {
  isPositiveSimulationAmount,
  type AuthorizationSimulationRequest,
  type RefundSimulationRequest,
  type SimulationCard,
  type SimulationResult,
} from "./simulation";
import { useSimulation } from "./useSimulation";

const props = defineProps<{
	loadCards: () => Promise<Card[]>;
	authorize: (request: AuthorizationSimulationRequest) => Promise<SimulationResult | undefined>;
	refund: (request: RefundSimulationRequest) => Promise<unknown>;
	initialCardId?: number | null;
}>();
const emit = defineEmits<{ completed: [] }>();
const message = useMessage();
const cards = ref<Card[]>([]);
const cardsLoading = ref(false);
const cardsError = ref("");
const authorizationBusy = ref(false);
const activeTab = ref("authorization");
const { submitting, result, submit } = useSimulation();
const form = reactive({
  authorizationId: null as number | null,
  cardId: null as number | null,
  amount: 100 as number | null,
  currency: "USD",
  merchantName: "Amazon",
  merchantMCC: "5411",
  merchantCountry: "US",
  merchantCity: "Seattle",
});
const simulationCards = computed<SimulationCard[]>(() => cards.value.map((card) => ({
  id: card.id,
  label: `${card.account_name || "账户不存在"} · ${card.card_number}  ${card.card_currency}`,
  currency: card.card_currency,
  disabled: card.card_status !== "active",
})));
const cardOptions = computed(() => simulationCards.value.map((card) => ({
  label: card.label,
  value: card.id,
  disabled: card.disabled,
})));
const authorizationId = computed(() => form.authorizationId);
const currencyOptions = computed(() => [...new Set([
  "USD", "EUR", "GBP", ...simulationCards.value.map((card) => card.currency),
])].map((currency) => ({
  label: currency,
  value: currency,
})));

function selectRefundCard(cardId: number | null) {
  form.cardId = cardId;
  const card = simulationCards.value.find((item) => item.id === cardId);
  if (card) {
    form.currency = card.currency;
  }
}

async function loadCards() {
  if (cardsLoading.value) {
    return;
  }
  cardsLoading.value = true;
  cardsError.value = "";
  try {
    cards.value = await props.loadCards();
    const initial = simulationCards.value.find((card) => card.id === props.initialCardId && !card.disabled);
    selectRefundCard(initial?.id ?? simulationCards.value.find((card) => !card.disabled)?.id ?? null);
  } catch (error) {
    cardsError.value = error instanceof Error ? error.message : "加载卡失败";
  } finally {
    cardsLoading.value = false;
  }
}

watch(() => props.initialCardId, (cardId) => {
  const initial = simulationCards.value.find((card) => card.id === cardId && !card.disabled);
  if (initial) {
    selectRefundCard(initial.id);
  }
});

async function submitRefund() {
  if (submitting.value) {
    return;
  }
  if (!isPositiveSimulationAmount(form.amount)) {
    message.warning("退款金额必须是有限的正数");
    return;
  }
  if (authorizationId.value !== null && (!Number.isSafeInteger(authorizationId.value) || authorizationId.value <= 0)) {
    message.warning("授权 ID 必须是正整数");
    return;
  }
  const card = simulationCards.value.find((item) => item.id === form.cardId);
  if (!authorizationId.value && (!card || card.disabled || !form.currency)) {
    message.warning("独立退款请选择可用卡和币种");
    return;
  }
  if (!form.merchantName.trim() || !form.merchantMCC.trim() || !form.merchantCountry.trim()) {
    message.warning("请填写商户名称、MCC 和国家");
    return;
  }
  const request: RefundSimulationRequest = {
    amount: form.amount,
    merchant_name: form.merchantName.trim(),
    merchant_category_code: form.merchantMCC.trim(),
    merchant_country: form.merchantCountry.trim(),
    merchant_city: form.merchantCity.trim(),
  };
  if (authorizationId.value) {
    request.authorization_id = authorizationId.value;
  } else {
    request.card_id = card!.id;
    request.currency = form.currency;
  }
  if (await submit(() => props.refund(request), "退款交易已创建。")) {
    emit("completed");
  }
}

onMounted(loadCards);
</script>

<template>
  <div class="simulation-page">
    <n-space vertical size="large">
      <n-alert v-if="cardsError" type="error" :show-icon="false">
        {{ cardsError }}
        <n-button text @click="loadCards">
          重试加载卡
        </n-button>
      </n-alert>
      <n-tabs v-model:value="activeTab" type="line" animated>
        <n-tab-pane name="authorization" tab="模拟授权" :disabled="submitting">
          <n-card title="授权配置" hover style="max-width: 720px">
            <CardAuthorizationForm
              :cards="simulationCards"
              :authorize="authorize"
              :initial-card-id="initialCardId"
              :disabled="cardsLoading"
              @busy="authorizationBusy = $event"
              @completed="emit('completed')"
            />
          </n-card>
        </n-tab-pane>
        <n-tab-pane name="refund" tab="模拟退款" :disabled="authorizationBusy">
          <n-card title="退款配置" hover style="max-width: 720px">
            <n-form
              label-placement="top"
              :disabled="submitting"
              @submit.prevent="submitRefund"
            >
              <div class="form-grid">
                <n-form-item v-if="!authorizationId" class="form-wide" label="卡" required>
                  <n-select
                    :value="form.cardId"
                    :options="cardOptions"
                    :loading="cardsLoading"
                    filterable
                    placeholder="选择卡"
                    @update:value="selectRefundCard"
                  />
                </n-form-item>
                <n-form-item label="关联授权 ID（可选，留空为独立退款）">
                  <n-input-number v-model:value="form.authorizationId" :min="1" :precision="0" placeholder="不要求先清算" />
                </n-form-item>
                <n-form-item label="退款金额" required>
                  <n-input-number
                    v-model:value="form.amount"
                    :min="0.01"
                    :precision="2"
                    style="width: 100%"
                  />
                </n-form-item>
                <n-form-item v-if="!authorizationId" label="币种" required>
                  <n-select v-model:value="form.currency" :options="currencyOptions" />
                </n-form-item>
                <n-form-item label="商户名称" required>
                  <n-input v-model:value="form.merchantName" placeholder="例如 Amazon" />
                </n-form-item>
                <n-form-item label="MCC" required>
                  <n-input v-model:value="form.merchantMCC" placeholder="例如 5411" />
                </n-form-item>
                <n-form-item label="商户国家" required>
                  <n-input v-model:value="form.merchantCountry" placeholder="例如 US" />
                </n-form-item>
                <n-form-item label="地区">
                  <n-input v-model:value="form.merchantCity" placeholder="例如 Seattle" />
                </n-form-item>
              </div>
              <n-button
                attr-type="submit"
                type="error"
                block
                :loading="submitting"
                :disabled="submitting"
              >
                模拟退款
              </n-button>
            </n-form>
            <n-alert v-if="result" type="success" :show-icon="false" style="margin-top: 16px">
              {{ result }}
            </n-alert>
          </n-card>
        </n-tab-pane>
      </n-tabs>
    </n-space>
  </div>
</template>
