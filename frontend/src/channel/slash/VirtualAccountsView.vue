<script setup lang="ts">
import { onMounted, ref } from "vue";
import { NCard, NDataTable } from "naive-ui";
import { managementApi, type VirtualAccount } from "./api";
const rows = ref<VirtualAccount[]>([]);
const loading = ref(false);
async function load() { loading.value = true; try { rows.value = await managementApi.virtualAccounts(); } finally { loading.value = false; } }
onMounted(() => void load());
</script>
<template><section><div class="page-heading"><div><h1>虚拟账户</h1><p>共享余额账户及累计支出。</p></div></div><n-card :bordered="false"><n-data-table :loading="loading" :data="rows" :columns="[{ title: '名称', key: 'name' }, { title: '可用余额', key: 'balance' }, { title: '累计支出', key: 'spend' }, { title: '币种', key: 'currency' }, { title: 'ID', key: 'id' }]" /></n-card></section></template>
