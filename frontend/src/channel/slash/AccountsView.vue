<script setup lang="ts">
import { h, onMounted, ref } from "vue";
import {
  createDiscreteApi,
  type DataTableColumns,
  NButton,
  NDataTable,
  NForm,
  NFormItem,
  NInput,
  NModal,
  NSpace
} from "naive-ui";
import { type Account, accountApi } from "./api";

const { message } = createDiscreteApi(["message"]);
const loading = ref(false);
const visible = ref(false);
const name = ref("");
const rows = ref<Account[]>([]);

async function load() {
  loading.value = true;
  try {
    rows.value = await accountApi.list();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "无法加载账户");
  } finally {
    loading.value = false;
  }
}

async function create() {
  if (!name.value) return;
  try {
    await accountApi.create({ name: name.value });
    visible.value = false;
    name.value = "";
    await load();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "创建失败");
  }
}

const columns: DataTableColumns<Account> = [{ title: "名称", key: "name" }, {
  title: "账户 ID",
  key: "id"
}, { title: "创建时间", key: "created_at", render: (row) => h("span", new Date(row.created_at).toLocaleString()) }];
onMounted(() => void load());
</script>
<template>
  <section>
    <div class="page-heading">
      <div><h1>账户</h1></div>
      <n-button type="primary" @click="visible = true">新增账户</n-button>
    </div>
    <n-data-table :columns="columns" :data="rows" :loading="loading" :bordered="false"/>
    <n-modal v-model:show="visible" preset="card" title="新增账户" style="width: min(440px, calc(100vw - 32px))">
      <n-form label-placement="top">
        <n-form-item label="名称">
          <n-input v-model:value="name"/>
        </n-form-item>
      </n-form>
      <template #action>
        <n-space justify="end">
          <n-button @click="visible = false">取消</n-button>
          <n-button type="primary" @click="create">创建</n-button>
        </n-space>
      </template>
    </n-modal>
  </section>
</template>
