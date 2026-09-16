<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import {
  NButton,
  NCard,
  NDataTable,
  NForm,
  NFormItem,
  NGrid,
  NGi,
  NInput,
  NInputNumber,
  NSelect,
  NSpace,
  createDiscreteApi,
} from "naive-ui";
import { api, refundApi } from "./api";
import type { Card, Transaction } from "@/channel/types";

const { message } = createDiscreteApi(["message"]);
const loading = ref(false);
const submitting = ref(false);
const cards = ref<Card[]>([]);
const transactions = ref<Transaction[]>([]);
const cardFilter = ref("");
const merchantFilter = ref("");
const form = ref({
  card_id: "",
  amount: 100,
  currency: "USD",
  merchant_name: "Amazon",
  merchant_category_code: "",
  merchant_country: "",
  merchant_city: "",
});

const cardOptions = computed(() => cards.value.map((card) => ({
  label: `${card.card_number} | ID: ${card.id}`,
  value: card.id,
  disabled: card.card_status.toLowerCase() !== "active",
})));
const refunds = computed(() => transactions.value.filter((item) =>
  item.transaction_type === "refund" &&
  item.card_id.includes(cardFilter.value) &&
  item.merchant_name.toLowerCase().includes(merchantFilter.value.toLowerCase()),
));
const columns = [
  { title: "退款 ID", key: "id", ellipsis: { tooltip: true } },
  { title: "卡片 ID", key: "card_id", ellipsis: { tooltip: true } },
  { title: "退款金额", key: "amount" },
  { title: "币种", key: "currency" },
  { title: "商户", key: "merchant_name" },
  { title: "MCC", key: "merchant_category_code" },
  { title: "状态", key: "status" },
  { title: "退款时间", key: "transacted_at" },
];

async function load() {
  loading.value = true;
  try {
    const [cardResult, transactionResult] = await Promise.all([
      api.listCards(),
      api.listTransactions(),
    ]);
    cards.value = cardResult.data;
    transactions.value = transactionResult.data;
    if (!form.value.card_id) {
      form.value.card_id = cards.value.find((card) => card.card_status.toLowerCase() === "active")?.id ?? "";
    }
  } catch (error) {
    message.error(error instanceof Error ? error.message : "加载退款配置失败");
  } finally {
    loading.value = false;
  }
}

async function simulateRefund() {
  if (!form.value.card_id || !form.value.merchant_category_code || !form.value.merchant_country) {
    message.warning("请填写卡片、MCC 和国家");
    return;
  }
  submitting.value = true;
  try {
    const refund = await refundApi.simulate(form.value);
    message.success(`退款交易已创建：${refund.id}`);
    await load();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "模拟退款失败");
  } finally {
    submitting.value = false;
  }
}

onMounted(() => void load());
</script>

<template>
  <section>
    <div class="page-heading">
      <div>
        <h1>模拟退款</h1>
        <p>直接为活跃卡片创建退款交易。</p>
      </div>
      <n-button :loading="loading" @click="load">刷新</n-button>
    </div>
    <n-card title="退款配置" :bordered="false" style="max-width: 720px">
      <n-form :model="form" label-placement="top">
        <n-form-item label="选择卡片" required>
          <n-select v-model:value="form.card_id" :options="cardOptions" :loading="loading" filterable />
        </n-form-item>
        <n-grid :cols="2" :x-gap="12">
          <n-gi><n-form-item label="退款金额" required><n-input-number v-model:value="form.amount" :min="0.01" :precision="2" style="width: 100%" /></n-form-item></n-gi>
          <n-gi><n-form-item label="货币"><n-select v-model:value="form.currency" :options="[{ label: 'USD', value: 'USD' }, { label: 'EUR', value: 'EUR' }, { label: 'GBP', value: 'GBP' }]" /></n-form-item></n-gi>
        </n-grid>
        <n-form-item label="商户名称"><n-input v-model:value="form.merchant_name" /></n-form-item>
        <n-form-item label="MCC 码" required><n-input v-model:value="form.merchant_category_code" placeholder="例如 5411" /></n-form-item>
        <n-grid :cols="2" :x-gap="12">
          <n-gi><n-form-item label="国家" required><n-input v-model:value="form.merchant_country" placeholder="例如 US" /></n-form-item></n-gi>
          <n-gi><n-form-item label="地区"><n-input v-model:value="form.merchant_city" placeholder="例如 Seattle" /></n-form-item></n-gi>
        </n-grid>
        <n-button type="error" block :loading="submitting" @click="simulateRefund">模拟退款</n-button>
      </n-form>
    </n-card>
    <n-card title="退款列表" :bordered="false" style="margin-top: 16px">
      <n-space style="margin-bottom: 12px"><n-input v-model:value="cardFilter" placeholder="卡片 ID" clearable /><n-input v-model:value="merchantFilter" placeholder="商户名称" clearable /></n-space>
      <n-data-table :loading="loading" :columns="columns" :data="refunds" />
    </n-card>
  </section>
</template>
