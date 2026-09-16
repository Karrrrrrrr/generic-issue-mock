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
  createDiscreteApi,
} from "naive-ui";
import { api } from "@/channel/paynda/api";
import type { Card } from "@/channel/types";
const emit = defineEmits<{ completed: [] }>();
const { message } = createDiscreteApi(["message"]);
const cards = ref<Card[]>([]);
const loading = ref(false);
const result = ref("");
const form = ref({
  cardID: "",
  amount: 1,
  currency: "USD",
  merchantName: "",
  merchantMCC: "",
  merchantCountry: "US",
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
onMounted(() => {
  void loadCards();
});
</script>
<template>
  <n-card title="授权配置" class="simulation-card" :bordered="false"
    ><n-form label-placement="top"
      ><div class="form-grid">
        <n-form-item class="form-wide" label="活动卡" required
          ><n-select
            v-model:value="form.cardID"
            :options="options"
            filterable
            placeholder="选择活动卡" /></n-form-item
        ><n-form-item label="交易金额" required
          ><n-input-number
            v-model:value="form.amount"
            :min="0.01"
            :precision="2"
            style="width: 100%" /></n-form-item
        ><n-form-item label="币种"
          ><n-select
            v-model:value="form.currency"
            :options="[
              { label: 'USD', value: 'USD' },
              { label: 'GBP', value: 'GBP' },
              { label: 'CNY', value: 'CNY' },
            ]" /></n-form-item
        ><n-form-item label="商户名称" required
          ><n-input
            v-model:value="form.merchantName"
            placeholder="例如 Amazon" /></n-form-item
        ><n-form-item label="MCC" required
          ><n-input
            v-model:value="form.merchantMCC"
            placeholder="例如 5411" /></n-form-item
        ><n-form-item class="form-wide" label="商户国家"
          ><n-input v-model:value="form.merchantCountry"
        /></n-form-item>
      </div>
      <n-space justify="end"
        ><n-button type="primary" :loading="loading" @click="submit"
          >创建授权</n-button
        ></n-space
      ></n-form
    ><n-alert
      v-if="result"
      type="success"
      :show-icon="false"
      style="margin-top: 16px"
      >{{ result }}</n-alert
    ></n-card
  >
</template>
