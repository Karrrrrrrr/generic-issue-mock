<script setup lang="ts">
import { h, onMounted, ref } from "vue";
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
import { api, fundCard } from "./api";
import type { Card } from "@/channel/types";

const { message } = createDiscreteApi(["message"]);
const loading = ref(false);
const rows = ref<Card[]>([]);
const fundingCard = ref<Card>();
const fundingAmount = ref<number | null>(null);

async function load() {
  loading.value = true;
  try {
    rows.value = (await api.listCards()).data;
  } catch (error) {
    message.error(error instanceof Error ? error.message : "加载卡片失败");
  } finally {
    loading.value = false;
  }
}

async function changeStatus(card: Card, status: string) {
  try {
    await api.updateCardStatus(card.id, status);
    await load();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "更新卡片状态失败");
  }
}

async function submitFund() {
  if (!fundingCard.value || !fundingAmount.value || fundingAmount.value <= 0) {
    message.warning("请输入正数充值金额");
    return;
  }
  try {
    await fundCard(fundingCard.value.id, fundingAmount.value);
    fundingCard.value = undefined;
    fundingAmount.value = null;
    await load();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "充值失败");
  }
}

const columns = [{ title: "卡号", key: "card_number" }, { title: "卡 BIN", key: "card_bin" }, {
  title: "币种",
  key: "card_currency"
}, {
  title: "状态",
  key: "card_status",
  render: (card: Card) => h(NTag, {
    type: card.card_status === "normal" ? "success" : "warning",
    size: "small"
  }, { default: () => card.card_status })
}, {
  title: "操作",
  key: "actions",
  render: (card: Card) => h(NSpace, { size: 8 }, {
    default: () => [
      h(NButton, {
        size: "small",
        type: "primary",
        onClick: () => {
          fundingCard.value = card;
          fundingAmount.value = null;
        },
      }, { default: () => "充值" }),
      card.card_status === "normal" ? h(NButton, {
      size: "small",
      type: "warning",
      onClick: () => changeStatus(card, "frozen")
      }, { default: () => "冻结" }) : h(NButton, {
      size: "small",
      type: "primary",
      onClick: () => changeStatus(card, "normal")
      }, { default: () => "恢复" }),
    ]
  })
}];
onMounted(() => void load());
</script>
<template>
  <div class="page-heading">
    <div><h1>卡片管理</h1>
      <p>查看卡片状态并执行冻结或恢复。</p></div>
    <n-button :loading="loading" @click="load">刷新</n-button>
  </div>
  <n-card :bordered="false">
    <n-data-table :loading="loading" :columns="columns" :data="rows"/>
  </n-card>
  <n-modal
      :show="Boolean(fundingCard)"
      preset="card"
      title="账户充值到卡"
      @update:show="(shown) => { if (!shown) fundingCard = undefined }"
  >
    <n-form label-placement="top">
      <n-form-item label="充值金额">
        <n-input-number v-model:value="fundingAmount" :min="0.01" :precision="2" style="width: 100%"/>
      </n-form-item>
    </n-form>
    <template #action>
      <n-space justify="end">
        <n-button @click="fundingCard = undefined">取消</n-button>
        <n-button type="primary" @click="submitFund">充值</n-button>
      </n-space>
    </template>
  </n-modal>
</template>
