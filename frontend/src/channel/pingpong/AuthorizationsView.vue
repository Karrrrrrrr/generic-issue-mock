<script setup lang="ts">
import { h, ref } from "vue";
import {
  NAlert,
  NButton,
  NCard,
  NDataTable,
  NInput,
  NModal,
  NSpace,
} from "naive-ui";
import { formatDateTime } from "@/channel/dateTime";
import { renderAmountTag } from "@/channel/tableTags";
import { api, requestID, type Authorization } from "./api";
import { useList } from "./list";
import { renderTag } from "./tags";
import TransactionStageForm from "@/channel/shared/TransactionStageForm.vue";
import { simulationStages, type SimulationStage } from "@/channel/shared/simulation";

const { rows, loading, filters, pagination, search, load } = useList<Authorization>("authorizations");
const accountFilter = ref("");
const cardFilter = ref("");
const selected = ref<Authorization>();
const action = ref<SimulationStage>("clear");
const simulationBusy = ref(false);
const orderID = ref("");
const columns = [
  {
    title: "授权 ID",
    key: "id",
    width: 100
  },
  {
    title: "账户",
    key: "account_name",
    width: 160
  },
  {
    title: "卡 ID",
    key: "card_id",
    width: 100
  },
  {
    title: "商户",
    key: "merchant_name",
    width: 170
  },
  {
    title: "授权金额",
    key: "amount",
    width: 140,
    render: (row: Authorization) => renderAmountTag(row.amount, row.currency)
  },
  {
    title: "剩余金额",
    key: "remaining",
    width: 140,
    render: (row: Authorization) => renderAmountTag(row.remaining, row.currency)
  },
  {
    title: "授权状态",
    key: "status",
    width: 110,
    render: (row: Authorization) => renderTag(row.status)
  },
  {
    title: "通知状态",
    key: "notification_status",
    width: 240,
    render: (row: Authorization) => renderTag(row.notification_status)
  },
  {
    title: "时间",
    key: "created_at",
    width: 180,
    render: (row: Authorization) => formatDateTime(row.created_at)
  },
  {
    title: "后续操作",
    key: "actions",
    width: 220,
    render: (row: Authorization) => h(NSpace, {
      size: 6
    }, {
      default: () => simulationStages.map(item => h(NButton, {
        size: "small",
        onClick: () => {
          selected.value = row;
          action.value = item.key;
          orderID.value = requestID();
        },
      }, {
        default: () => item.title
      })),
    }),
  },
];
function query() {
  filters.account_id = accountFilter.value || undefined;
  filters.card_id = cardFilter.value || undefined;
  search();
}
async function simulateStage(amount: number) {
  const authorization = selected.value;
  if (!authorization) {
    throw new Error("请先选择授权");
  }
  await api.post(`authorizations/${authorization.id}/stages`, {
    amount,
    stage: action.value,
    request_id: orderID.value,
  });
}

function completeSimulation() {
  selected.value = undefined;
  void load();
}

function closeSimulation(shown: boolean) {
  if (!shown && !simulationBusy.value) {
    selected.value = undefined;
  }
}
</script>

<template>
  <section class="ping-page">
    <div class="page-heading">
      <h1>
        授权管理
      </h1>
    </div>
    <n-alert type="warning" :show-icon="false">
      本地授权已生效，通知未投递：尚缺 PingPong Webhook 事件及报文契约。这里不配置同步审批。
    </n-alert>
    <n-card :bordered="false">
      <n-space class="ping-toolbar">
        <n-input
          v-model:value="accountFilter"
          placeholder="账户 ID"
          clearable
          style="width: 160px"
        />
        <n-input
          v-model:value="cardFilter"
          placeholder="卡 ID"
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
        :scroll-x="1560"
      />
    </n-card>
    <n-modal
      :show="Boolean(selected)"
      preset="card"
      :title="simulationStages.find(item => item.key === action)?.title"
      :closable="!simulationBusy"
      :mask-closable="!simulationBusy"
      :close-on-esc="!simulationBusy"
      style="width: min(480px, 90vw)"
      @update:show="closeSimulation"
    >
      <p>
        授权
        {{ selected?.id }}
        · 剩余
        {{ selected?.remaining }}
        {{ selected?.currency }}
      </p>
      <p>
        允许超额及继续清算，剩余金额可为负数。退款不恢复授权额度。
      </p>
      <TransactionStageForm
        v-if="selected"
        :key="orderID"
        :stage="action"
        :currency="selected.currency"
        :simulate="simulateStage"
        @busy="simulationBusy = $event"
        @completed="completeSimulation"
      />
    </n-modal>
  </section>
</template>
