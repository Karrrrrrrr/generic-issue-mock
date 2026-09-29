<script setup lang="ts">
import { useRemotePagination } from "@/channel/pagination";
import { onMounted, ref } from "vue";
import { NDataTable, useMessage } from "naive-ui";
import type { CardProduct } from "./api";
import type { ListResponse, PageRequest } from "@/channel/types";
const { loadProducts, prefixLabel = "BIN" } = defineProps<{
  loadProducts: (page: PageRequest) => Promise<ListResponse<CardProduct>>;
  prefixLabel?: string;
}>();

const rows = ref<CardProduct[]>([]);
const { page, pageSize, total, pagination } = useRemotePagination(load);
const loading = ref(false);
const message = useMessage();

async function load() {
  loading.value = true;
  try {
    const response = await loadProducts({
      page_number: page.value,
      page_size: pageSize.value,
    });
    rows.value = response.data;
    total.value = response.total_items;
  } catch (error) {
    message.error(error instanceof Error ? error.message : "加载卡产品失败");
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
        <slot name="description"><p>渠道共享的卡产品与 BIN；开卡必须显式指定产品。</p></slot>
      </div>
    </div>
    <n-data-table
      max-height="max(160px, calc(100dvh - 400px))"
      :pagination="pagination"
      :scroll-x="600"
      table-layout="fixed"
      :loading="loading"
      :data="rows"
      :columns="[
        { title: prefixLabel, key: 'prefix' },
        { title: 'ID', key: 'id' },
      ]"
    />
  </section>
</template>
