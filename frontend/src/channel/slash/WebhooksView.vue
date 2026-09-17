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
  NSelect,
  NSpace,
  NSwitch,
  NTag,
} from "naive-ui";
import { type Webhook, webhookApi } from "@/channel/slash/api";

const { dialog, message } = createDiscreteApi(["dialog", "message"]);
const loading = ref(false);
const rows = ref<Webhook[]>([]);
const creating = ref(false);
const editing = ref<Webhook | null>(null);
const form = ref({ event: "transaction.created", target_url: "", enabled: true });
const eventOptions = [
  "transaction.created",
  "transaction.updated",
  "authorization.created",
  "authorization.updated",
].map((value) => ({ label: value, value }));

async function load() {
  loading.value = true;
  try {
    rows.value = await webhookApi.list();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "无法加载 Webhook 配置");
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  editing.value = null;
  form.value = { event: "transaction.created", target_url: "", enabled: true };
  creating.value = true;
}

function openEdit(item: Webhook) {
  editing.value = item;
  form.value = { event: item.event, target_url: item.target_url, enabled: item.enabled };
  creating.value = true;
}

async function save() {
  if (!form.value.target_url) {
    message.warning("请填写 Webhook URL");
    return;
  }
  try {
    if (editing.value) {
      await webhookApi.update(editing.value.id, { target_url: form.value.target_url, enabled: form.value.enabled });
    } else {
      await webhookApi.create(form.value);
    }
    creating.value = false;
    await load();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "保存失败");
  }
}

function remove(item: Webhook) {
  dialog.warning({
    title: "删除 Webhook",
    content: item.target_url,
    positiveText: "删除",
    negativeText: "取消",
    onPositiveClick: async () => {
      try {
        await webhookApi.remove(item.id);
        await load();
      } catch (error) {
        message.error(error instanceof Error ? error.message : "删除失败");
      }
    },
  });
}

const columns: DataTableColumns<Webhook> = [
  {
    title: "事件",
    key: "event",
    width: 220,
    render: (row) => h(NTag, { size: "small" }, { default: () => row.event })
  },
  { title: "目标地址", key: "target_url", ellipsis: { tooltip: true } },
  {
    title: "状态",
    key: "enabled",
    width: 100,
    render: (row) => h(NTag, {
      type: row.enabled ? "success" : "default",
      size: "small"
    }, { default: () => row.enabled ? "启用" : "停用" })
  },
  { title: "更新时间", key: "updated_at", width: 180 },
  {
    title: "操作",
    key: "actions",
    width: 150,
    render: (row) => h(NSpace, { size: 6 }, {
      default: () => [h(NButton, {
        size: "small",
        onClick: () => openEdit(row)
      }, { default: () => "编辑" }), h(NButton, {
        size: "small",
        type: "error",
        onClick: () => remove(row)
      }, { default: () => "删除" })]
    })
  },
];

onMounted(() => void load());
</script>

<template>
  <section>
    <div class="page-heading">
      <div><h1>Webhook</h1>
        <p>管理事件回调地址与启用状态。</p></div>
      <n-button type="primary" @click="openCreate">新增 Webhook</n-button>
    </div>
    <n-data-table :columns="columns" :data="rows" :loading="loading" :bordered="false"/>
    <n-modal v-model:show="creating" preset="card" :title="editing ? '编辑 Webhook' : '新增 Webhook'"
             style="width: min(560px, calc(100vw - 32px))">
      <n-form label-placement="top">
        <n-form-item label="事件">
          <n-select v-model:value="form.event" :options="eventOptions" :disabled="Boolean(editing)"/>
        </n-form-item>
        <n-form-item label="Webhook URL">
          <n-input v-model:value="form.target_url" placeholder="https://example.com/webhooks/slash"/>
        </n-form-item>
        <n-form-item label="启用">
          <n-switch v-model:value="form.enabled"/>
        </n-form-item>
      </n-form>
      <template #action>
        <n-space justify="end">
          <n-button @click="creating = false">取消</n-button>
          <n-button type="primary" @click="save">保存</n-button>
        </n-space>
      </template>
    </n-modal>
  </section>
</template>
