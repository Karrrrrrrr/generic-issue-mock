<script setup lang="ts">
import { onMounted, ref } from "vue";
import { NCard, NDataTable } from "naive-ui";
import { api } from "./api";
import type { Card } from "@/channel/types";
const rows = ref<Card[]>([]);
const columns = [
  { title: "卡号", key: "card_number" },
  { title: "卡 BIN", key: "card_bin" },
  { title: "币种", key: "card_currency" },
  { title: "状态", key: "card_status" },
];
onMounted(async () => {
  rows.value = (await api.listCards()).data;
});
</script>
<template>
  <div class="page-heading">
    <div>
      <h1>卡片</h1>
      <p>Paynda 卡片列表</p>
    </div>
  </div>
  <n-card :bordered="false">
    <n-data-table :columns="columns" :data="rows" />
  </n-card>
</template>
