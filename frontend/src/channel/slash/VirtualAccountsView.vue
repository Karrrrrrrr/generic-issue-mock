<script setup lang="ts">
import { onMounted, ref } from "vue";
import { NCard, NDataTable, type DataTableColumns } from "naive-ui";
import { managementApi, type VirtualAccount } from "./api";

const rows = ref<VirtualAccount[]>([]);
const loading = ref(false);
const columns: DataTableColumns<VirtualAccount> = [
  {
    title: "名称",
    key: "name",
  },
  {
    title: "资金来源",
    key: "funding_source",
  },
  {
    title: "资金余额",
    key: "balance",
  },
  {
    title: "累计支出",
    key: "spend",
  },
  {
    title: "币种",
    key: "currency",
  },
  {
    title: "ID",
    key: "id",
  },
];

async function load() {
  loading.value = true;
  try {
    rows.value = await managementApi.virtualAccounts();
  } finally {
    loading.value = false;
  }
}

onMounted(() => void load());
</script>
<template>
  <section>
    <div class="page-heading">
      <div>
        <h1>虚拟账户</h1>
        <p>共享余额账户及累计支出。</p>
      </div>
    </div>
    <n-card :bordered="false">
      <n-data-table
          :loading="loading"
          :data="rows"
          :columns="columns"
      />
    </n-card>
  </section>
</template>
