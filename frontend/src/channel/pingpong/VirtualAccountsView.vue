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
  NSelect,
  NSpace,
  useMessage,
} from "naive-ui";
import { formatDateTime } from "@/channel/dateTime";
import { renderAmountTag } from "@/channel/tableTags";
import {
  api,
  accountOptions,
  requestID,
  type VirtualAccount,
} from "./api";
import { useList } from "./list";
const { rows, loading, busy, filters, pagination, load, search, perform } = useList<VirtualAccount>("virtual-accounts");
const accountFilter = ref("");
const creating = ref(false);
const options = ref<{
  label: string;
  value: string;
}[]>([]);
const accountID = ref<string | null>(null);
const name = ref("");
const selected = ref<VirtualAccount>();
const amount = ref<number | null>(null);
const orderID = ref("");
const message = useMessage();
const columns = [
  {
    title: "虚拟账户 ID",
    key: "id",
    width: 100
  },
  {
    title: "虚拟账户名称",
    key: "name",
    width: 170
  },
  {
    title: "账户",
    key: "account_name",
    width: 180
  },
  {
    title: "虚拟账户余额",
    key: "balance",
    width: 150,
    render: (row: VirtualAccount) => renderAmountTag(row.balance, row.currency)
  },
  {
    title: "创建时间",
    key: "created_at",
    width: 180,
    render: (row: VirtualAccount) => formatDateTime(row.created_at)
  },
  {
    title: "操作",
    key: "actions",
    width: 150,
    render: (row: VirtualAccount) => h(NButton, {
      size: "small",
      onClick: () => {
        selected.value = row;
        amount.value = null;
        orderID.value = requestID();
      },
    }, {
      default: () => "从账户充值"
    })
  },
];
async function openCreate() {
  try {
    options.value = await accountOptions();
    creating.value = true;
  }
  catch (error) {
    message.error(error instanceof Error ? error.message : "加载账户失败");
  }
}
async function create() {
  if (!accountID.value || !name.value.trim()) {
    return;
  }
  if (await perform(() => api.post("virtual-accounts", {
    account_id: accountID.value,
    name: name.value
  }))) {
    creating.value = false;
    name.value = "";
  }
}
async function fund() {
  if (!selected.value || amount.value === null || amount.value <= 0) {
    return;
  }
  if (await perform(() => api.post(`virtual-accounts/${selected.value!.id}/fund`, {
    account_id: selected.value!.account_id,
    amount: amount.value,
    request_id: orderID.value,
  }))) {
    selected.value = undefined;
  }
}
function query() {
  filters.account_id = accountFilter.value || undefined;
  search();
}
</script>

<template>
  <section class="ping-page">
    <div class="page-heading">
      <h1>
        虚拟账户
      </h1>
    </div>
    <n-alert type="info" :show-icon="false">
      虚拟账户是卡的资金来源，不是卡的消费钱包。当前支持 USD。
    </n-alert>
    <n-card :bordered="false">
      <n-space class="ping-toolbar">
        <n-input
          v-model:value="accountFilter"
          placeholder="账户 ID"
          clearable
          style="width: 180px"
        />
        <n-button @click="query">
          查询
        </n-button>
        <n-button type="primary" @click="openCreate">
          创建虚拟账户
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
        :scroll-x="930"
      />
    </n-card>
    <n-modal
      v-model:show="creating"
      preset="card"
      title="创建虚拟账户"
      style="width: min(480px, 90vw)"
    >
      <n-space vertical :size="20">
        <n-select
          v-model:value="accountID"
          :options="options"
          placeholder="所属账户"
        />
        <n-input v-model:value="name" placeholder="虚拟账户名称" />
      </n-space>
      <template #action>
        <n-button
          type="primary"
          :loading="busy"
          @click="create"
        >
          创建
        </n-button>
      </template>
    </n-modal>
    <n-modal
      :show="Boolean(selected)"
      preset="card"
      title="账户 → 虚拟账户"
      style="width: min(480px, 90vw)"
      @update:show="shown => { if (!shown) selected = undefined; }"
    >
      <p>
        {{ selected?.account_name }}
        →
        {{ selected?.name }}
      </p>
      <n-input-number
        v-model:value="amount"
        :min="0.01"
        :precision="2"
      />
      <template #action>
        <n-button
          type="primary"
          :loading="busy"
          @click="fund"
        >
          确认充值
        </n-button>
      </template>
    </n-modal>
  </section>
</template>
