<script setup lang="ts">
import { renderAmountTag } from "@/channel/tableTags";
import { h, onMounted, ref } from "vue";
import { createDiscreteApi, NButton, NCard, NDataTable, NInputNumber, NModal } from "naive-ui";
import { request } from "@/channel/shared";

type Authorization = {
  id: string;
  account_id: string;
  account_name: string;
  card_id: string;
  amount: string;
  settled: string;
  remaining: string;
  currency: string;
  merchant_name: string;
  created_at: string;
};

const baseURL = "/paynda/ui";
const { message } = createDiscreteApi(["message"]);
const loading = ref(false);
const saving = ref(false);
const rows = ref<Authorization[]>([]);
const selected = ref<Authorization>();
const amount = ref<number | null>(null);
const columns = [
  {
    title: "账户名称",
    key: "account_name",
  },
  {
    title: "所属账户 ID",
    key: "account_id",
  },
  {
    title: "授权 ID",
    key: "id",
  },
  {
    title: "卡 ID",
    key: "card_id",
  },
  {
    title: "授权金额",
    key: "amount",
    render: (row: Authorization) => renderAmountTag(row.amount, row.currency),
  },
  {
    title: "已清算",
    key: "settled",
    render: (row: Authorization) => renderAmountTag(row.settled, row.currency),
  },
  {
    title: "可清算",
    key: "remaining",
    render: (row: Authorization) => renderAmountTag(row.remaining, row.currency),
  },
  {
    title: "商户",
    key: "merchant_name",
  },
  {
    title: "时间",
    key: "created_at",
  },
  {
    title: "操作",
    key: "actions",
    render: (row: Authorization) =>
      h(
        NButton,
        {
          size: "small",
          type: "primary",
          onClick: () => {
            selected.value = row;
            amount.value = Number(row.remaining) > 0 ? Number(row.remaining) : null;
          },
        },
        { default: () => "清算" },
      ),
  },
];

async function load() {
  loading.value = true;
  try {
    rows.value = (await request.get<Authorization[]>(baseURL + "/authorization-balances")).data;
  } catch (error) {
    message.error(error instanceof Error ? error.message : "加载授权失败");
  } finally {
    loading.value = false;
  }
}

async function clear() {
  if (!selected.value || !amount.value || amount.value <= 0) return;
  saving.value = true;
  try {
    await request.post(baseURL + "/authorizations/" + selected.value.id + "/clear", {
      account_id: selected.value.account_id,
      amount: String(amount.value),
    });
    selected.value = undefined;
    message.success("清算成功，已生成清算交易");
    await load();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "清算失败");
  } finally {
    saving.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div class="page-heading">
    <div>
      <h1>授权管理</h1>
      <p>从授权发起部分或全额清算，每次清算生成一笔独立交易。</p>
    </div>
    <n-button :loading="loading" @click="load">刷新</n-button>
  </div>
  <n-card :bordered="false">
    <n-data-table
      :scroll-x="1600"
      table-layout="fixed"
      :loading="loading"
      :columns="columns"
      :data="rows"
    />
  </n-card>
  <n-modal
    :show="Boolean(selected)"
    preset="card"
    title="授权清算"
    @update:show="
      (shown) => {
        if (!shown) selected = undefined;
      }
    "
  >
    <p>授权 {{ selected?.id }} · 剩余 {{ selected?.remaining }} {{ selected?.currency }}</p>
    <n-input-number v-model:value="amount" :min="0.01" :precision="2" />
    <template #action>
      <n-button type="primary" :loading="saving" @click="clear">确认清算</n-button>
    </template>
  </n-modal>
</template>
