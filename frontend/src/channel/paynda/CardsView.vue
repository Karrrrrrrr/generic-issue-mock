<script setup lang="ts">
import { h, onMounted, ref } from "vue";
import { NButton, NCard, NDataTable, NSpace, NTag, createDiscreteApi } from "naive-ui";
import { api } from "./api";
import type { Card } from "@/channel/types";
const { message } = createDiscreteApi(["message"]); const loading = ref(false); const rows = ref<Card[]>([]);
async function load() { loading.value = true; try { rows.value = (await api.listCards()).data; } catch (error) { message.error(error instanceof Error ? error.message : "加载卡片失败"); } finally { loading.value = false; } }
async function changeStatus(card: Card, status: string) { try { await api.updateCardStatus(card.id, status); await load(); } catch (error) { message.error(error instanceof Error ? error.message : "更新卡片状态失败"); } }
const columns = [{ title: "卡号", key: "card_number" }, { title: "卡 BIN", key: "card_bin" }, { title: "币种", key: "card_currency" }, { title: "状态", key: "card_status", render: (card: Card) => h(NTag, { type: card.card_status === "ACTIVE" ? "success" : "warning", size: "small" }, { default: () => card.card_status }) }, { title: "操作", key: "actions", render: (card: Card) => h(NSpace, { size: 8 }, { default: () => card.card_status === "ACTIVE" ? [h(NButton, { size: "small", type: "warning", onClick: () => changeStatus(card, "FROZEN") }, { default: () => "冻结" })] : [h(NButton, { size: "small", type: "primary", onClick: () => changeStatus(card, "ACTIVE") }, { default: () => "恢复" })] }) }];
onMounted(() => void load());
</script>
<template><div class="page-heading"><div><h1>卡片管理</h1><p>查看卡片状态并执行冻结或恢复。</p></div><n-button :loading="loading" @click="load">刷新</n-button></div><n-card :bordered="false"><n-data-table :loading="loading" :columns="columns" :data="rows" /></n-card></template>
