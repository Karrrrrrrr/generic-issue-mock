<script setup lang="ts">
import { onMounted, ref } from "vue";
import {
  NButton,
  NCard,
  NDataTable,
  NForm,
  NFormItem,
  NInput,
  NModal,
  NSpace,
  createDiscreteApi,
} from "naive-ui";
import { api } from "./api";
import type { Cardholder } from "@/channel/types";

const { message } = createDiscreteApi(["message"]);
const rows = ref<Cardholder[]>([]);
const open = ref(false);
const form = ref({ firstName: "", lastName: "", email: "", mobile: "" });
const columns = [
  {
    title: "姓名",
    key: "name",
    render: (row: Cardholder) => `${row.first_name} ${row.last_name}`,
  },
  { title: "邮箱", key: "email" },
  { title: "手机", key: "phone_number" },
  { title: "状态", key: "status" },
];
async function load() {
  try {
    rows.value = (await api.listCardholders()).data;
  } catch (error) {
    message.error(error instanceof Error ? error.message : "加载持卡人失败");
  }
}
async function create() {
  if (!form.value.firstName.trim() || !form.value.lastName.trim()) {
    message.error("请填写姓名");
    return;
  }
  if (!form.value.email.trim()) {
    message.error("请填写邮箱");
    return;
  }
  try {
    await api.createCardholder(form.value);
    open.value = false;
    form.value = { firstName: "", lastName: "", email: "", mobile: "" };
    await load();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "创建失败");
  }
}
onMounted(() => void load());
</script>
<template>
  <div class="page-heading">
    <div>
      <h1>持卡人</h1>
      <p>Paynda 持卡人列表</p>
    </div>
    <n-button type="primary" @click="open = true">新增持卡人</n-button>
  </div>
  <n-card :bordered="false">
    <n-data-table :columns="columns" :data="rows" />
  </n-card>
  <n-modal v-model:show="open" preset="card" title="新增持卡人">
    <n-form>
      <n-form-item label="名字">
        <n-input v-model:value="form.firstName" />
      </n-form-item>
      <n-form-item label="姓氏">
        <n-input v-model:value="form.lastName" />
      </n-form-item>
      <n-form-item label="邮箱">
        <n-input v-model:value="form.email" />
      </n-form-item>
      <n-form-item label="手机">
        <n-input v-model:value="form.mobile" />
      </n-form-item>
    </n-form>
    <n-space justify="end">
      <n-button @click="open = false">取消</n-button>
      <n-button type="primary" @click="create">创建</n-button>
    </n-space>
  </n-modal>
</template>
