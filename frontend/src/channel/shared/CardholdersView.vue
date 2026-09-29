<script setup lang="ts">
import { useRemotePagination } from "@/channel/pagination";
import { renderEnumTag } from "@/channel/tableTags";
import { onMounted, ref } from "vue";
import { createDiscreteApi, NDataTable } from "naive-ui";
import type { ChannelAPI } from "@/channel/types";
const { api, channelName } = defineProps<{ api: Pick<ChannelAPI, "listCardholders">; channelName: string }>();
import type { Cardholder } from "@/channel/types";

const { message } = createDiscreteApi(["message"]);
const rows = ref<Cardholder[]>([]);
const { page, pageSize, total, pagination } = useRemotePagination(load);
const loading = ref(false);
const columns = [
  {
    title: "账户名称",
    key: "account_name",
  },
  {
    title: "持卡人 ID",
    key: "id",
  },
  {
    title: "姓名",
    key: "name",
    render: (holder: Cardholder) => holder.first_name + " " + holder.last_name,
  },
  {
    title: "邮箱",
    key: "email",
  },
  {
    title: "手机",
    key: "phone_number",
  },
  {
    title: "状态",
    key: "status",
    render: (holder: Cardholder) => renderEnumTag(holder.status, "status"),
  },
];

async function load() {
  loading.value = true;
  try {
    const response = await api.listCardholders({
      page_number: page.value,
      page_size: pageSize.value,
    });
    rows.value = response.data;
    total.value = response.total_items;
  } catch (error) {
    message.error(error instanceof Error ? error.message : "加载持卡人失败");
  } finally {
    loading.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div class="page-heading">
    <div>
      <h1>持卡人</h1>
      <p>查看 {{ channelName }} 全部账户的持卡人；持卡人创建和开卡由 OpenAPI 完成。</p>
    </div>
  </div>
  <n-data-table
    max-height="max(160px, calc(100dvh - 400px))"
    remote
    :pagination="pagination"
    :scroll-x="1100"
    table-layout="fixed"
    :loading="loading"
    :columns="columns"
    :data="rows"
  />
</template>
