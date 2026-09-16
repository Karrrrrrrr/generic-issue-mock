<script setup lang="ts">
import { onMounted, ref } from "vue";
import { NCard, NDataTable } from "naive-ui";
import { managementApi, type CardProduct } from "./api";
const rows = ref<CardProduct[]>([]);
const loading = ref(false);
async function load() { loading.value = true; try { rows.value = await managementApi.cardProducts(); } finally { loading.value = false; } }
onMounted(() => void load());
</script>
<template><section><div class="page-heading"><div><h1>卡产品</h1><p>渠道可用 BIN 与默认开卡产品。</p></div></div><n-card :bordered="false"><n-data-table :loading="loading" :data="rows" :columns="[{ title: 'BIN', key: 'prefix' }, { title: '默认产品', key: 'is_default' }, { title: 'ID', key: 'id' }]" /></n-card></section></template>
