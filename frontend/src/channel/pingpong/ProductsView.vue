<script setup lang="ts">
import { onMounted, ref } from "vue";
import {
  NAlert,
  NCard,
  NDataTable,
  useMessage,
} from "naive-ui";
import { useClientPagination } from "@/channel/pagination";
import { renderEnumTag } from "@/channel/tableTags";
import { api, type Page, type Product } from "./api";
const rows = ref<Product[]>([]);
const pagination = useClientPagination(rows);
const loading = ref(false);
const message = useMessage();
const columns = [
  {
    title: "产品代码",
    key: "id",
    width: 180
  },
  {
    title: "候选 BIN",
    key: "prefix",
    width: 300
  },
  {
    title: "默认产品",
    key: "is_default",
    width: 120,
    render: (row: Product) => renderEnumTag(row.is_default, "defaultProduct")
  },
];
onMounted(async () => {
  loading.value = true;
  try {
    rows.value = (await api.get<Page<Product>>("products")).items;
  }
  catch (error) {
    message.error(error instanceof Error ? error.message : "加载失败");
  }
  finally {
    loading.value = false;
  }
});
</script>

<template>
  <section class="ping-page">
    <div class="page-heading">
      <h1>
        卡产品
      </h1>
    </div>
    <n-alert type="info" :show-icon="false">
      产品在渠道内共享；开卡随机选一个候选 BIN，共享发卡序号。当前为 USD、24 个月有效期的虚拟账户关联独立卡。
    </n-alert>
    <n-card :bordered="false">
      <n-data-table
        :bordered="true"
        :columns="columns"
        :data="rows"
        :loading="loading"
        :pagination="pagination"
        :scroll-x="600"
      />
    </n-card>
  </section>
</template>
