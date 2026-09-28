<script setup lang="ts">
import { computed, h, ref } from "vue";
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
} from "naive-ui";
import { formatDateTime } from "@/channel/dateTime";
import { renderAmountTag } from "@/channel/tableTags";
import { api, requestID, type Card } from "./api";
import { useList } from "./list";
import { renderTag } from "./tags";
import CardAuthorizationForm from "@/channel/shared/CardAuthorizationForm.vue";
import type { AuthorizationSimulationRequest, SimulationCard } from "@/channel/shared/simulation";

const { rows, loading, busy, filters, pagination, search, perform, load } = useList<Card>("cards");
const accountFilter = ref("");
const cardFilter = ref("");
const statusFilter = ref<string | null>(null);
const selected = ref<Card>();
const operation = ref("top_up");
const amount = ref<number | null>(null);
const simulationBusy = ref(false);
const orderID = ref("");
const simulationCards = computed<SimulationCard[]>(() => selected.value ? [{
  id: selected.value.id,
  label: `${selected.value.account_name} · ${selected.value.card_number}`,
  currency: selected.value.currency,
  disabled: selected.value.status !== "ACTIVE",
}] : []);
const statuses = [
  {
    label: "正常",
    value: "ACTIVE"
  },
  {
    label: "冻结",
    value: "REVOKED"
  },
  {
    label: "已注销",
    value: "CANCELLED"
  },
];
function open(row: Card, action: string) {
  selected.value = row;
  operation.value = action;
  amount.value = null;
  orderID.value = requestID();
}
function actionButton(row: Card, label: string, action: string) {
  return h(NButton, {
    size: "small",
    disabled: row.status === "CANCELLED" || (action === "authorize" && row.status !== "ACTIVE"),
    onClick: () => open(row, action),
  }, {
    default: () => label
  });
}
const columns = [
  {
    title: "卡 ID",
    key: "id",
    width: 90
  },
  {
    title: "账户",
    key: "account_name",
    width: 150
  },
  {
    title: "虚拟账户 ID",
    key: "virtual_account_id",
    width: 100
  },
  {
    title: "卡号",
    key: "card_number",
    width: 200
  },
  {
    title: "BIN",
    key: "card_bin",
    width: 100
  },
  {
    title: "状态",
    key: "status",
    width: 110,
    render: (row: Card) => renderTag(row.status)
  },
  {
    title: "卡可用余额",
    key: "balance",
    width: 140,
    render: (row: Card) => renderAmountTag(row.balance, row.currency)
  },
  {
    title: "授权预占",
    key: "reserved",
    width: 140,
    render: (row: Card) => renderAmountTag(row.reserved, row.currency)
  },
  {
    title: "创建时间",
    key: "created_at",
    width: 180,
    render: (row: Card) => formatDateTime(row.created_at)
  },
  {
    title: "操作",
    key: "actions",
    width: 320,
    render: (row: Card) => h(NSpace, {
      size: 6
    }, {
      default: () => [
        actionButton(row, "充值", "top_up"),
        actionButton(row, "转回虚拟账户", "withdraw"),
        actionButton(row, "模拟授权", "authorize"),
        h(NButton, {
          size: "small",
          disabled: row.status !== "ACTIVE" && row.status !== "REVOKED",
          onClick: () => perform(() => api.put(`cards/${row.id}/status`, {
            account_id: row.account_id,
            status: row.status === "ACTIVE" ? "REVOKED" : "ACTIVE",
          })),
        }, {
          default: () => row.status === "REVOKED" ? "恢复" : "冻结"
        }),
      ],
    }),
  },
];
function query() {
  filters.account_id = accountFilter.value || undefined;
  filters.card_id = cardFilter.value || undefined;
  filters.status = statusFilter.value || undefined;
  search();
}
async function submit() {
  const card = selected.value;
  if (!card || amount.value === null || amount.value <= 0) {
    return;
  }
  const common = {
    amount: amount.value,
    request_id: orderID.value,
  };
  const done = await perform(() => api.post(`cards/${card.id}/fund`, {
    ...common,
    account_id: card.account_id,
    action: operation.value,
  }));
  if (done) {
    selected.value = undefined;
  }
}

async function simulateAuthorization(request: AuthorizationSimulationRequest) {
  await api.post("simulate/authorizations", {
    card_id: request.cardID,
    amount: request.amount,
    currency: request.currency,
    merchant_name: request.merchantName,
    request_id: orderID.value,
  });
}

function completeSimulation() {
  selected.value = undefined;
  void load();
}

function closeOperation(shown: boolean) {
  if (!shown && !busy.value && !simulationBusy.value) {
    selected.value = undefined;
  }
}
</script>

<template>
  <section class="ping-page">
    <div class="page-heading">
      <h1>
        卡管理
      </h1>
    </div>
    <n-alert type="info" :show-icon="false">
      卡拥有独立钱包。充值从所属虚拟账户扣款，转出返回同一虚拟账户；开卡通过 OpenAPI 完成。
    </n-alert>
    <n-card :bordered="false">
      <n-space class="ping-toolbar">
        <n-input
          v-model:value="accountFilter"
          placeholder="账户 ID"
          clearable
          style="width: 150px"
        />
        <n-input
          v-model:value="cardFilter"
          placeholder="卡 ID"
          clearable
          style="width: 150px"
        />
        <n-select
          v-model:value="statusFilter"
          :options="statuses"
          placeholder="卡状态"
          clearable
          style="width: 160px"
        />
        <n-button @click="query">
          查询
        </n-button>
      </n-space>
      <n-data-table
        :bordered="true"
        remote
        :columns="columns"
        :data="rows"
        :loading="loading"
        :pagination="pagination"
        :scroll-x="1630"
      />
    </n-card>
    <n-modal
      :show="Boolean(selected)"
      preset="card"
      :title="operation === 'authorize' ? '本地模拟授权' : operation === 'withdraw' ? '卡 → 虚拟账户' : '虚拟账户 → 卡'"
      :closable="!busy && !simulationBusy"
      :mask-closable="!busy && !simulationBusy"
      :close-on-esc="!busy && !simulationBusy"
      style="width: min(560px, 90vw)"
      @update:show="closeOperation"
    >
      <n-space vertical :size="20">
        <p>
          {{ selected?.account_name }}
          ·
          {{ selected?.card_number }}
        </p>
        <p>
          卡可用余额：
          {{ selected?.balance }}
          {{ selected?.currency }}
        </p>
        <n-alert
          v-if="operation === 'authorize'"
          type="warning"
          :show-icon="false"
        >
          余额足够即可本地授权，不等待下游同意。Webhook 协议尚未补齐，本轮不会发送通知。
        </n-alert>
        <CardAuthorizationForm
          v-if="selected && operation === 'authorize'"
          :key="orderID"
          :cards="simulationCards"
          :authorize="simulateAuthorization"
          :select-card="false"
          :merchant-details="false"
          @busy="simulationBusy = $event"
          @completed="completeSimulation"
        />
        <n-input-number
          v-else
          v-model:value="amount"
          :min="0.01"
          :precision="2"
          placeholder="金额"
        />
      </n-space>
      <template #action>
        <n-button
          v-if="operation !== 'authorize'"
          type="primary"
          :loading="busy"
          @click="submit"
        >
          确认
        </n-button>
      </template>
    </n-modal>
  </section>
</template>
