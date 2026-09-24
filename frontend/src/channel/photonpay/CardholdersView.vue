<script setup lang="ts">
import { renderEnumTag } from "@/channel/tableTags";
import { onMounted, ref } from "vue";
import { createDiscreteApi, NButton, NCard, NDataTable } from "naive-ui";
import { api } from "./api";
import type { Cardholder } from "@/channel/types";

const { message } = createDiscreteApi(["message"]);
const rows = ref<Cardholder[]>([]);
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
    rows.value = (await api.listCardholders()).data;
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
      <p>查看 PhotonPay 全部账户的持卡人；持卡人创建和开卡由 OpenAPI 完成。</p>
    </div>
    <n-button :loading="loading" @click="load">刷新</n-button>
  </div>
  <n-card :bordered="false">
    <n-data-table
      :scroll-x="1100"
      table-layout="fixed"
      :loading="loading"
      :columns="columns"
      :data="rows"
    />
  </n-card>
</template>
