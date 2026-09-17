<script setup lang="ts">
import { h, onMounted, ref } from "vue";
import {
  createDiscreteApi,
  NButton,
  NCard,
  NDataTable,
  NForm,
  NFormItem,
  NInputNumber,
  NModal,
  NSpace,
  NTag,
} from "naive-ui";
import { api, fundsApi } from "./api";
import type { Card } from "@/channel/types";

const { message } = createDiscreteApi(["message"]);
const loading = ref(false);
const rows = ref<Card[]>([]);
const fundingCard = ref<Card>();
const fundingAmount = ref<number | null>(null);
const withdrawing = ref(false);
const funding = ref(false);

async function load() {
  loading.value = true;
  try {
    rows.value = (await api.listCards()).data;
  } catch (error) {
    message.error(error instanceof Error ? error.message : "加载卡片失败");
  } finally {
    loading.value = false;
  }
}

async function changeStatus(card: Card, status: string) {
  try {
    await api.updateCardStatus(card.id, status);
    await load();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "更新卡片状态失败");
  }
}

function openFunding(card: Card, withdraw: boolean) {
  fundingCard.value = card;
  fundingAmount.value = null;
  withdrawing.value = withdraw;
}

async function submitFunding() {
  if (!fundingCard.value || !fundingAmount.value || fundingAmount.value <= 0) {
    message.warning("请输入正数金额");
    return;
  }
  funding.value = true;
  try {
    await fundsApi.transferResource({
      accountID: fundingCard.value.account_id,
      walletID: fundingCard.value.wallet_id,
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

const columns = [
  {
    title: "账户名称",
    key: "account_name",
  },
  {
    title: "卡 BIN",
    key: "card_bin",
  },
  {
    title: "卡号",
    key: "card_number",
  },
  {
    title: "所属账户 ID",
    key: "account_id",
  },
  {
    title: "币种",
    key: "card_currency",
  },
  {
    title: "余额",
    key: "balance",
  },
  {
    title: "资金类型",
    key: "funding_source",
  },
  {
    title: "状态",
    key: "card_status",
    render: (card: Card) =>
      h(
        NTag,
        {
          type: card.card_status === "ACTIVE" ? "success" : "warning",
          size: "small",
        },
        {
          default: () => card.card_status,
        },
      ),
  },
  {
    title: "操作",
    key: "actions",
    render: (card: Card) =>
      h(
        NSpace,
        {
          size: 8,
        },
        {
          default: () => [
            h(
              NButton,
              {
                size: "small",
                onClick: () => openFunding(card, false),
              },
              {
                default: () => "充值",
              },
            ),
            h(
              NButton,
              {
                size: "small",
                onClick: () => openFunding(card, true),
              },
              {
                default: () => "转出",
              },
            ),
            h(
              NButton,
              {
                size: "small",
                onClick: () =>
                  changeStatus(card, card.card_status === "ACTIVE" ? "FROZEN" : "ACTIVE"),
              },
              {
                default: () => (card.card_status === "ACTIVE" ? "冻结" : "恢复"),
              },
            ),
          ],
        },
      ),
  },
];

onMounted(load);
</script>

<template>
  <div class="page-heading">
    <div>
      <h1>卡片管理</h1>
      <p>充值从关联账户钱包扣款；转出退回关联账户钱包。</p>
    </div>
    <n-button :loading="loading" @click="load">刷新</n-button>
  </div>
  <n-card :bordered="false">
    <n-data-table :loading="loading" :columns="columns" :data="rows" />
  </n-card>
  <n-modal
    :show="Boolean(fundingCard)"
    preset="card"
    :title="withdrawing ? '卡资金转出到账户' : '账户资金充值到卡'"
    @update:show="
      (shown) => {
        if (!shown) fundingCard = undefined;
      }
    "
  >
    <p>{{ fundingCard?.card_number }} · {{ fundingCard?.card_currency }}</p>
    <p>
      关联账户：{{ fundingCard?.account_name || "账户不存在" }} ·
      {{ fundingCard?.account_id }}
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
</template>
