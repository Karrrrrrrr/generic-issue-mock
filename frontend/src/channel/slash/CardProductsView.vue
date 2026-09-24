<script setup lang="ts">
import { renderEnumTag } from "@/channel/tableTags";
import { onMounted, ref } from "vue";
import { NCard, NDataTable } from "naive-ui";
import { type CardProduct, managementApi } from "./api";

const rows = ref<CardProduct[]>([]);
const loading = ref(false);

async function load() {
  loading.value = true;
  try {
    rows.value = await managementApi.cardProducts();
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
        <h1>卡产品</h1>
        <p>渠道可用 BIN 与默认开卡产品。</p>
      </div>
    </div>
    <n-card :bordered="false">
      <n-data-table
        :scroll-x="600"
        table-layout="fixed"
        :loading="loading"
        :data="rows"
        :columns="[
          { title: 'BIN', key: 'prefix' },
          {
            title: '默认产品',
            key: 'is_default',
            render: (row: CardProduct) => renderEnumTag(row.is_default, 'defaultProduct'),
          },
          { title: 'ID', key: 'id' },
        ]"
      />
    </n-card>
  </section>
</template>
