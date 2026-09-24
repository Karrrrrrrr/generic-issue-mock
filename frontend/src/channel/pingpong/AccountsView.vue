<script setup lang="ts">
import { h, ref } from "vue";
import {
  NAlert,
  NButton,
  NCard,
  NDataTable,
  NInput,
  NInputNumber,
  NModal,
  NSpace,
} from "naive-ui";
import { formatDateTime } from "@/channel/dateTime";
import { renderAmountTag } from "@/channel/tableTags";
import { api, type Account } from "./api";
import { useList } from "./list";
const { rows, loading, busy, pagination, load, perform } = useList<Account>("accounts");
const name = ref("");
const selected = ref<Account>();
const amount = ref<number | null>(null);
const columns = [
  {
    title: "账户 ID",
    key: "id",
    width: 100
  },
  {
    title: "账户名称",
    key: "name",
    width: 180
  },
  {
    title: "余额",
    key: "balance",
    width: 150,
    render: (row: Account) => renderAmountTag(row.balance, row.currency)
  },
  {
    title: "创建时间",
    key: "created_at",
    width: 180,
    render: (row: Account) => formatDateTime(row.created_at)
  },
  {
    title: "操作",
    key: "actions",
    width: 140,
    render: (row: Account) => h(NButton, {
      size: "small",
      onClick: () => {
        selected.value = row;
        amount.value = null;
      },
    }, {
      default: () => "调整余额"
    }),
  },
];
async function create() {
  if (!name.value.trim()) {
    return;
  }
  if (await perform(() => api.post("accounts", {
    name: name.value
  }))) {
    name.value = "";
  }
}
async function adjust() {
  if (!selected.value || amount.value === null || amount.value === 0) {
    return;
  }
  if (await perform(() => api.post(`accounts/${selected.value!.id}/balance`, {
    amount: amount.value
  }))) {
    selected.value = undefined;
  }
}
</script>

<template>
  <section class="ping-page">
    <div class="page-heading">
      <h1>
        PingPong 账户
      </h1>
    </div>
    <n-alert type="info" :show-icon="false">
      OpenAPI 使用 PINGPONG_APP_ACCOUNTS 显式绑定应用与此处账户 ID；创建账户不创建虚拟账户或复制产品。
    </n-alert>
    <n-card :bordered="false">
      <n-space class="ping-toolbar">
        <n-input
          v-model:value="name"
          placeholder="新账户名称"
          style="width: 240px"
        />
        <n-button
          type="primary"
          :loading="busy"
          @click="create"
        >
          创建账户
        </n-button>
        <n-button @click="load">
          刷新
        </n-button>
      </n-space>
      <n-data-table
        :bordered="true"
        remote
        :columns="columns"
        :data="rows"
        :loading="loading"
        :pagination="pagination"
        :scroll-x="750"
      />
    </n-card>
    <n-modal
      :show="Boolean(selected)"
      preset="card"
      title="调整账户余额"
      style="width: min(480px, 90vw)"
      @update:show="shown => { if (!shown) selected = undefined; }"
    >
      <p>
        {{ selected?.name }}
        · 正数增加，负数扣减
      </p>
      <n-input-number v-model:value="amount" :precision="2" />
      <template #action>
        <n-button
          type="primary"
          :loading="busy"
          @click="adjust"
        >
          确认
        </n-button>
      </template>
    </n-modal>
  </section>
</template>
