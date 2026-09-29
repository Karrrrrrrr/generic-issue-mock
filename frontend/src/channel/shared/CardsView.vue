<script setup lang="ts">
import TableFilters, { type FilterField } from "@/channel/TableFilters.vue";
import { useTableFilters } from "@/channel/tableFilters";
import type { ManagementAPI } from "./api";
import { useRemotePagination } from "@/channel/pagination";
import { formatDateTime } from "@/channel/dateTime";
import { cardStatusOptions, type CardStatus } from "@/channel/enums";
import { renderAmountTag, renderEnumTag } from "@/channel/tableTags";
import { h, onMounted, ref } from "vue";
import {
	createDiscreteApi,
	NButton,
	NDataTable,
	NForm,
	NFormItem,
	NInputNumber,
	NModal,
	NSpace,
} from "naive-ui";

import type { Card } from "@/channel/types";

import type { ChannelAPI } from "@/channel/types";
import type { CardFundingRequest } from "./contracts";
import CardTransactionSimulator from "./CardTransactionSimulator.vue";
import type { AuthorizationSimulationRequest, RefundSimulationRequest } from "./simulation";
const {
	api,
	accountApi,
	fundCard,
	showExpiry = true,
	showReserved = false,
	showVirtualAccount = false,
	detailedFilters = true,
	newRequestId,
	simulateAuthorization,
	simulateRefund,
	fundingSourceLabel = "账户",
} = defineProps<{
	api: Pick<ChannelAPI, "listCards" | "updateCardStatus">;
	accountApi: Pick<ManagementAPI["accountApi"], "listAll">;
	fundCard: (request: CardFundingRequest) => Promise<unknown>;
	showReserved?: boolean;
	showVirtualAccount?: boolean;
	detailedFilters?: boolean;
	newRequestId?: () => string;
	simulateAuthorization?: (request: AuthorizationSimulationRequest) => Promise<unknown>;
	simulateRefund?: (request: RefundSimulationRequest) => Promise<unknown>;
	fundingSourceLabel?: string;
	showExpiry?: boolean;
}>();

const { message } = createDiscreteApi(["message"]);
const loading = ref(false);
const requestId = ref<string>();
const rows = ref<Card[]>([]);
const { page, pageSize, total, pagination } = useRemotePagination(load);
const fundingCard = ref<Card>();
const fundingAmount = ref<number | null>(null);
const withdrawing = ref(false);
const funding = ref(false);
const simulationCard = ref<Card>();
const authorizationRequestId = ref<string>();
const refundRequestId = ref<string>();

const filterFields: FilterField[] = [
  {
    "key": "id",
    "label": "卡 ID（精确）"
  },
  ...(detailedFilters ? [{ key: "card_number", label: "卡号（支持部分卡号）" }] : []),
  {
    "key": "card_status",
    "label": "卡状态",
    options: cardStatusOptions
  }
];
const {
  filters,
  dateRange,
  appliedFilters,
  accountOptions,
  accountsLoading,
  loadAccounts,
  search,
  reset,
} = useTableFilters({
  loadAccounts: () => accountApi.listAll(),
  onSearch: () => {
    page.value = 1;
    void load();
  },
});

async function load() {
  loading.value = true;
  try {
    const response = await api.listCards({
      ...appliedFilters.value,
      page_number: page.value,
      page_size: pageSize.value,
    });
    rows.value = response.data;
    total.value = response.total_items;
  } catch (error) {
    message.error(error instanceof Error ? error.message : "加载卡片失败");
  } finally {
    loading.value = false;
  }
}

function statusAction(card: Card): { label: string; status: CardStatus | null } {
  switch (card.card_status) {
    case "active":
      return {
        label: "冻结",
        status: "frozen",
      };
    case "frozen":
      return {
        label: "恢复",
        status: "active",
      };
    case "deleted":
      return {
        label: "已注销",
        status: null,
      };
    default:
      return {
        label: "不可操作",
        status: null,
      };
  }
}

async function changeStatus(card: Card) {
  const action = statusAction(card);
  if (action.status === null) return;
  try {
    await api.updateCardStatus({
      id: card.id,
      account_id: card.account_id,
      card_status: action.status,
    });
    await load();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "更新卡片状态失败");
  }
}

function openFunding(card: Card, withdraw: boolean) {
  if (card.card_status !== "active") {
    message.warning("请先激活卡片，再调整资金");
    return;
  }
  fundingCard.value = card;
  requestId.value = newRequestId?.();
  fundingAmount.value = null;
  withdrawing.value = withdraw;
}

async function submitFunding() {
  if (!fundingCard.value || !fundingAmount.value || !Number.isFinite(fundingAmount.value) || fundingAmount.value <= 0) {
    message.warning("请输入正数金额");
    return;
  }
  if (fundingCard.value.card_status !== "active") {
    message.warning("请先激活卡片，再调整资金");
    return;
  }
  funding.value = true;
  try {
    await fundCard({
      card: fundingCard.value,
      requestId: requestId.value,
      amount: fundingAmount.value,
      withdraw: withdrawing.value,
    });
    fundingCard.value = undefined;
    message.success("资金划转完成");
    await load();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "资金划转失败");
  } finally {
    funding.value = false;
  }
}

function openSimulation(card: Card) {
  if (card.card_status !== "active") {
    message.warning("请先激活卡片，再模拟交易");
    return;
  }
  simulationCard.value = card;
  authorizationRequestId.value = newRequestId?.();
  refundRequestId.value = newRequestId?.();
}

async function loadSimulationCards(): Promise<Card[]> {
  const result = await api.listCards();
  return result.data;
}

async function submitSimulationAuthorization(request: AuthorizationSimulationRequest) {
  if (!simulateAuthorization) {
    return;
  }
  await simulateAuthorization({
    ...request,
    requestId: authorizationRequestId.value,
  });
}

async function submitSimulationRefund(request: RefundSimulationRequest) {
  if (!simulateRefund) {
    return;
  }
  await simulateRefund({
    ...request,
    request_id: refundRequestId.value,
  });
}

function completeSimulation() {
  simulationCard.value = undefined;
  void load();
}

function renderActions(card: Card) {
  return h(
    NSpace,
    {
      size: 8,
    },
    {
      default: () => [
        ...(simulateAuthorization && simulateRefund ? [h(NButton, {
          size: "small",
          disabled: card.card_status !== "active",
          onClick: () => openSimulation(card),
        }, { default: () => "模拟交易" })] : []),
        h(NButton, {
          size: "small",
          disabled: card.card_status !== "active",
          onClick: () => openFunding(card, false),
        }, { default: () => "充值" }),
        h(NButton, {
          size: "small",
          disabled: card.card_status !== "active",
          onClick: () => openFunding(card, true),
        }, { default: () => "转出" }),
        h(NButton, {
          size: "small",
          disabled: statusAction(card).status === null,
          onClick: () => changeStatus(card),
        }, { default: () => statusAction(card).label }),
      ],
    },
  );
}

const columns = [
  ...(showReserved ? [{
    title: "冻结余额",
    key: "reserved",
    render: (card: Card) => renderAmountTag(card.reserved ?? "0", card.card_currency),
  }] : []),
  ...(showVirtualAccount ? [{ title: "虚拟账户 ID", key: "virtual_account_id" }] : []),
  {
    title: "账户名称",
    key: "account_name",
  },
  {
    title: "卡 ID",
    key: "id",
  },
  {
    title: "卡 BIN",
    key: "card_bin",
  },
  {
    title: "卡号",
    key: "card_number",
    width: 200,
  },
  {
    title: "所属账户 ID",
    key: "account_id",
  },
  {
    title: "余额",
    key: "balance",
    render: (card: Card) => renderAmountTag(card.balance, card.card_currency),
  },
  {
    title: "资金类型",
    key: "funding_source",
    render: (card: Card) => card.funding_source ? renderEnumTag(card.funding_source, "funding") : "—",
  },
  {
    title: "状态",
    key: "card_status",
    render: (card: Card) => renderEnumTag(card.card_status, "status"),
  },
  ...(showExpiry ? [{
    title: "到期日",
    key: "expires_at",
    render: (row: Card) => formatDateTime(row.expires_at),
  }] : []),
  {
    title: "操作",
    key: "actions",
    render: renderActions,
  },
];

onMounted(load);
defineExpose({ reload: load });
</script>

<template>
  <div class="page-heading">
    <div>
      <h1>卡片管理</h1>
      <p>充值从{{ fundingSourceLabel }}钱包扣款；转出退回{{ fundingSourceLabel }}钱包。</p>
    </div>
  </div>
  <slot name="description" />
  <TableFilters
    v-model:values="filters"
    v-model:date-range="dateRange"
    :fields="filterFields"
    :show-date-range="detailedFilters"
    :account-options="accountOptions"
    :accounts-loading="accountsLoading"
    :loading="loading"
    @load-accounts="loadAccounts"
    @search="search"
    @reset="reset"
  />
  <n-data-table
    max-height="max(160px, calc(100dvh - 580px))"
    remote
    :pagination="pagination"
    :scroll-x="1600"
    table-layout="fixed"
    :loading="loading"
    :columns="columns"
    :data="rows"
  />
  <n-modal
    :show="Boolean(fundingCard)"
    preset="card"
    style="width: min(520px, calc(100vw - 32px))"
    :title="withdrawing ? `卡资金转出到${fundingSourceLabel}` : `${fundingSourceLabel}资金充值到卡`"
    @update:show="
      (shown) => {
        if (!shown) fundingCard = undefined;
      }
    "
  >
    <p>{{ fundingCard?.card_number }} · {{ fundingCard?.card_currency }}</p>
    <p>
      资金来源：{{ fundingSourceLabel }}；所属账户：{{ fundingCard?.account_name || "账户不存在" }} ·
      {{ fundingCard?.account_id }}
    </p>
    <p v-if="fundingCard?.funding_source === '虚拟账户共享资金'">
      此卡共享虚拟账户余额，操作会影响该共享钱包。
    </p>
    <n-form>
      <n-form-item label="金额">
        <n-input-number v-model:value="fundingAmount" :min="0.01" :precision="2" />
      </n-form-item>
    </n-form>
    <template #action>
      <n-button type="primary" :loading="funding" @click="submitFunding">确认</n-button>
    </template>
  </n-modal>
  <n-modal
    :show="Boolean(simulationCard)"
    preset="card"
    title="模拟交易"
    style="width: min(760px, calc(100vw - 32px))"
    @update:show="
      (shown) => {
        if (!shown) simulationCard = undefined;
      }
    "
  >
    <CardTransactionSimulator
      v-if="simulationCard && simulateAuthorization && simulateRefund"
      :key="simulationCard.id"
      :initial-card-id="simulationCard.id"
      :load-cards="loadSimulationCards"
      :authorize="submitSimulationAuthorization"
      :refund="submitSimulationRefund"
      @completed="completeSimulation"
    />
  </n-modal>
</template>
