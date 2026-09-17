<script setup lang="ts">
import { onMounted, ref } from "vue";
import { createDiscreteApi, NButton, NCard, NDataTable } from "naive-ui";
import { request } from "@/channel/shared";

type Authorization = {
  id: string;
  card_id: string;
  status: string;
  authorized_amount: string;
  currency: string;
  merchant_name: string;
  merchant_category_code: string;
  authorized_at: string;
};
const { message } = createDiscreteApi(["message"]);
const loading = ref(false);
const rows = ref<Authorization[]>([]);
const columns = [
  { title: "授权 ID", key: "id" }, { title: "卡片 ID", key: "card_id" },
  { title: "金额", key: "authorized_amount" }, { title: "币种", key: "currency" },
  { title: "商户", key: "merchant_name" }, { title: "MCC", key: "merchant_category_code" },
  { title: "状态", key: "status" }, { title: "授权时间", key: "authorized_at" },
];

async function load() {
  loading.value = true;
  try {
    rows.value = (await request.get<{ data: Authorization[] }>("/slash/ui/authorizations")).data.data;
  } catch (error) {
    message.error(error instanceof Error ? error.message : "加载授权列表失败");
  } finally {
    loading.value = false;
  }
}

onMounted(() => void load());
</script>
<template>
  <div class="page-heading">
    <div><h1>授权管理</h1>
      <p>查看已模拟的授权记录。</p></div>
    <n-button :loading="loading" @click="load">刷新</n-button>
  </div>
  <n-card :bordered="false">
    <n-data-table :loading="loading" :columns="columns" :data="rows"/>
  </n-card>
</template>
