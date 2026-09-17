<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
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
  createDiscreteApi,
} from "naive-ui";
import { api, refundApi } from "@/channel/photonpay/api";
import type { Card } from "@/channel/types";
const emit = defineEmits<{ completed: [] }>();
const { message } = createDiscreteApi(["message"]);
const cards = ref<Card[]>([]);
const loading = ref(false);
const refundLoading = ref(false);
const result = ref("");
const refundResult = ref("");
const activeTab = ref("authorization");
const form = ref({
  cardID: "",
  amount: 100,
  currency: "USD",
  merchantName: "Amazon",
  merchantMCC: "5411",
  merchantCountry: "US",
  merchantCity: "Seattle",
});
const refundForm = ref({
  card_id: "",
  amount: 100,
  currency: "USD",
  merchant_name: "Amazon",
  merchant_category_code: "5411",
  merchant_country: "US",
  merchant_city: "Seattle",
});
const options = computed(() =>
  cards.value
    .filter((card) => card.card_status === "active")
    .map((card) => ({
      label: `${card.card_number}  ${card.card_currency}`,
      value: card.id,
    })),
);
async function loadCards() {
  cards.value = (await api.listCards()).data;
  if (!form.value.cardID) {
    form.value.cardID = options.value[0]?.value ?? "";
  }
  if (!refundForm.value.card_id) {
    refundForm.value.card_id = cards.value.find(
      (card) => card.card_status === "active",
    )?.id ?? "";
  }
}
async function submit() {
  if (
    !form.value.cardID ||
    !form.value.merchantName ||
    !form.value.merchantMCC ||
    form.value.amount <= 0
  ) {
    message.warning("请选择活动卡，并填写正数金额、商户名称和 MCC");
    return;
  }
  loading.value = true;
  result.value = "";
  try {
    await api.simulateAuthorization(form.value);
    result.value = "授权已创建，可在交易处理查看并清算或撤销。";
    emit("completed");
  } catch (error) {
    message.error(error instanceof Error ? error.message : "模拟授权失败");
  } finally {
    loading.value = false;
  }
}
async function submitRefund() {
  if (!refundForm.value.card_id || !refundForm.value.merchant_name || !refundForm.value.merchant_category_code || !refundForm.value.merchant_country || refundForm.value.amount <= 0) {
    message.warning("请选择活动卡，并填写正数金额、商户名称、MCC 和国家");
    return;
  }
  refundLoading.value = true;
  refundResult.value = "";
  try {
    const refund = await refundApi.simulate(refundForm.value);
    refundResult.value = `退款交易已创建：${refund.id}`;
    emit("completed");
  } catch (error) {
    message.error(error instanceof Error ? error.message : "模拟退款失败");
  } finally {
    refundLoading.value = false;
  }
}
onMounted(() => {
  void loadCards();
});
</script>
<template>
  <div class="simulation-page">
    <n-space vertical size="large">
      <div class="simulation-heading">
        <h2>模拟交易</h2>
        <p>用于模拟授权和独立退款交易。</p>
      </div>
      <n-tabs v-model:value="activeTab" type="line" animated>
    <n-tab-pane name="authorization" tab="模拟授权">
      <n-card title="授权配置" hover style="max-width: 720px">
        <n-form label-placement="top">
      <div class="form-grid">
        <n-form-item class="form-wide" label="活动卡" required>
          <n-select
            v-model:value="form.cardID"
            :options="options"
            filterable
            placeholder="选择活动卡"
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
        <n-form-item label="币种">
          <n-select
            v-model:value="form.currency"
            :options="[
              { label: 'USD', value: 'USD' },
              { label: 'GBP', value: 'GBP' },
              { label: 'EUR', value: 'EUR' },
            ]"
          />
        </n-form-item>
        <n-form-item label="商户名称" required>
          <n-input
            v-model:value="form.merchantName"
            placeholder="例如 Amazon"
          />
        </n-form-item>
        <n-form-item label="MCC" required>
          <n-input v-model:value="form.merchantMCC" placeholder="例如 5411" />
        </n-form-item>
        <n-form-item label="商户国家">
          <n-input v-model:value="form.merchantCountry" />
        </n-form-item>
        <n-form-item label="地区"><n-input v-model:value="form.merchantCity" placeholder="例如 Seattle" /></n-form-item>
      </div>
      <n-button type="primary" block :loading="loading" @click="submit">模拟授权</n-button>
        </n-form>
        <n-alert
      v-if="result"
      type="success"
      :show-icon="false"
      style="margin-top: 16px"
        >
          {{ result }}
        </n-alert>
      </n-card>
    </n-tab-pane>
    <n-tab-pane name="refund" tab="模拟退款">
      <n-card title="退款配置" hover style="max-width: 720px">
        <n-form label-placement="top">
          <div class="form-grid">
            <n-form-item class="form-wide" label="活动卡" required><n-select v-model:value="refundForm.card_id" :options="options" filterable placeholder="选择活动卡" /></n-form-item>
            <n-form-item label="退款金额" required><n-input-number v-model:value="refundForm.amount" :min="0.01" :precision="2" style="width: 100%" /></n-form-item>
            <n-form-item label="币种"><n-select v-model:value="refundForm.currency" :options="[{ label: 'USD', value: 'USD' }, { label: 'EUR', value: 'EUR' }, { label: 'GBP', value: 'GBP' }]" /></n-form-item>
            <n-form-item label="商户名称" required><n-input v-model:value="refundForm.merchant_name" placeholder="例如 Amazon" /></n-form-item>
            <n-form-item label="MCC" required><n-input v-model:value="refundForm.merchant_category_code" placeholder="例如 5411" /></n-form-item>
            <n-form-item label="商户国家" required><n-input v-model:value="refundForm.merchant_country" placeholder="例如 US" /></n-form-item>
            <n-form-item label="地区"><n-input v-model:value="refundForm.merchant_city" placeholder="例如 Seattle" /></n-form-item>
          </div>
          <n-button type="error" block :loading="refundLoading" @click="submitRefund">模拟退款</n-button>
        </n-form>
        <n-alert v-if="refundResult" type="success" :show-icon="false" style="margin-top: 16px">{{ refundResult }}</n-alert>
      </n-card>
    </n-tab-pane>
      </n-tabs>
    </n-space>
  </div>
</template>
