<script setup lang="ts">
import { computed, h, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  NButton,
  NCard,
  NConfigProvider,
  NDataTable,
  NDivider,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NLayout,
  NLayoutContent,
  NLayoutHeader,
  NMenu,
  NModal,
  NSelect,
  NSpace,
  NTag,
  createDiscreteApi,
} from "naive-ui";

import { api } from "@/channel/slash/api";
import AuthorizationView from "@/channel/slash/AuthorizationView.vue";
import type { Card, Cardholder, Transaction } from "@/channel/types";

const channelLabel = "Slash";
const route = useRoute();
const router = useRouter();

const { message } = createDiscreteApi(["message"]);
const selectedPage = ref(route.path.split("/").at(-1) || "cardholders");
const loading = ref(false);
const cardholders = ref<Cardholder[]>([]);
const cards = ref<Card[]>([]);
const transactions = ref<Transaction[]>([]);
const cardholderModalVisible = ref(false);
const holderForm = ref({ firstName: "", lastName: "", email: "", mobile: "" });
const authorizationForm = ref({
  cardID: "",
  amount: 1,
  currency: "USD",
  merchantName: "",
  merchantMCC: "",
  merchantCountry: "US",
});

const menuOptions = computed(() => {
  const options = [
    { label: "持卡人", key: "cardholders" },
    { label: "卡片", key: "cards" },
    { label: "交易", key: "transactions" },
  ];
  options.push({ label: "授权模拟", key: "authorization" });
  return options;
});

const cardholderOptions = computed(() =>
  cardholders.value.map((item) => ({
    label: `${item.first_name} ${item.last_name} (${item.id})`,
    value: item.id,
  })),
);

const cardOptions = computed(() =>
  cards.value.map((item) => ({
    label: `${item.card_number} (${item.id})`,
    value: item.id,
  })),
);

const cardholderColumns = [
  { title: "ID", key: "id", ellipsis: { tooltip: true } },
  {
    title: "姓名",
    key: "name",
    render: (row: Cardholder) => `${row.first_name} ${row.last_name}`,
  },
  { title: "邮箱", key: "email" },
  { title: "手机号", key: "phone_number" },
  {
    title: "状态",
    key: "status",
    render: (row: Cardholder) =>
      h(
        NTag,
        { size: "small", type: "success" },
        { default: () => row.status },
      ),
  },
];

const cardColumns = [
  { title: "ID", key: "id", ellipsis: { tooltip: true } },
  { title: "卡号", key: "card_number" },
  { title: "币种", key: "card_currency" },
  {
    title: "状态",
    key: "card_status",
    render: (row: Card) =>
      h(
        NTag,
        {
          size: "small",
          type: row.card_status === "active" ? "success" : "warning",
        },
        { default: () => row.card_status },
      ),
  },
  { title: "CVV", key: "cvv" },
];

const transactionColumns = [
  { title: "交易 ID", key: "id", ellipsis: { tooltip: true } },
  { title: "类型", key: "transaction_type" },
  {
    title: "金额",
    key: "amount",
    render: (row: Transaction) => `${row.currency} ${row.amount}`,
  },
  { title: "商户", key: "merchant_name" },
  {
    title: "状态",
    key: "status",
    render: (row: Transaction) =>
      h(NTag, { size: "small" }, { default: () => row.status }),
  },
  {
    title: "操作",
    key: "actions",
    render: (row: Transaction) =>
      h(NSpace, { size: 4 }, { default: () => transactionActions(row) }),
  },
];

function transactionActions(row: Transaction) {
  if (row.transaction_type === "auth" && row.status === "authorized") {
    return [
      h(
        NButton,
        {
          size: "tiny",
          type: "primary",
          onClick: () => applyStep(row.id, "clear"),
        },
        { default: () => "清算" },
      ),
      h(
        NButton,
        { size: "tiny", onClick: () => applyStep(row.id, "reverse") },
        { default: () => "撤销" },
      ),
    ];
  }
  if (row.transaction_type === "clear" && row.status === "succeed") {
    return [
      h(
        NButton,
        {
          size: "tiny",
          type: "warning",
          onClick: () => applyStep(row.id, "refund"),
        },
        { default: () => "退款" },
      ),
    ];
  }
  return [
    h(NTag, { size: "small", type: "default" }, { default: () => "已处理" }),
  ];
}

async function loadCardholders() {
  const result = await api.listCardholders();
  cardholders.value = result.data;
}

async function loadCards() {
  const result = await api.listCards();
  cards.value = result.data;
}

async function loadTransactions() {
  const result = await api.listTransactions();
  transactions.value = result.data;
}

async function refresh() {
  loading.value = true;
  try {
    await Promise.all([loadCardholders(), loadCards(), loadTransactions()]);
  } catch (error) {
    message.error(
      error instanceof Error ? error.message : "加载控制台数据失败",
    );
  } finally {
    loading.value = false;
  }
}

async function createCardholder() {
  try {
    await api.createCardholder(holderForm.value);
    cardholderModalVisible.value = false;
    holderForm.value = { firstName: "", lastName: "", email: "", mobile: "" };
    await loadCardholders();
    message.success("持卡人已创建");
  } catch (error) {
    message.error(error instanceof Error ? error.message : "创建持卡人失败");
  }
}

async function simulateAuthorization() {
  if (
    !authorizationForm.value.cardID ||
    !authorizationForm.value.merchantName ||
    !authorizationForm.value.merchantMCC ||
    authorizationForm.value.amount <= 0
  ) {
    message.warning("请选择卡片并填写正数金额、商户名称和 MCC");
    return;
  }
  try {
    await api.simulateAuthorization(authorizationForm.value);
    await loadTransactions();
    message.success("授权交易已创建");
  } catch (error) {
    message.error(error instanceof Error ? error.message : "模拟授权失败");
  }
}

async function applyStep(id: string, action: "clear" | "reverse" | "refund") {
  try {
    await api.applyTransactionStep(id, action);
    await loadTransactions();
    message.success("交易状态已更新");
  } catch (error) {
    message.error(error instanceof Error ? error.message : "交易操作失败");
  }
}

async function handleAuthorizationComplete() {
  try {
    await loadTransactions();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "刷新交易失败");
  }
}

onMounted(refresh);
watch(
  () => route.path,
  (path) => {
    selectedPage.value = path.split("/").at(-1) || "cardholders";
  },
);
function navigate(page: string) {
  void router.push(`/slash/${page}`);
}
</script>

<template>
  <n-config-provider>
    <n-layout class="app-shell">
      <n-layout-header class="app-header">
        <div>
          <strong>Generic Mock</strong>
          <span>{{ channelLabel }} Console</span>
        </div>
        <n-button size="small" :loading="loading" @click="refresh"
          >刷新</n-button
        >
      </n-layout-header>
      <n-layout has-sider>
        <aside class="sidebar">
          <n-menu
            :value="selectedPage"
            :options="menuOptions"
            @update:value="navigate"
          />
        </aside>
        <n-layout-content class="content">
          <div class="page-heading">
            <div>
              <h1>
                {{
                  selectedPage === "authorization"
                    ? "授权模拟"
                    : selectedPage === "transactions"
                      ? "交易处理"
                      : selectedPage === "cards"
                        ? "卡片"
                        : "持卡人"
                }}
              </h1>
              <p>Slash 渠道模拟控制台</p>
            </div>
          </div>
          <section v-if="selectedPage === 'cardholders'">
            <n-card title="持卡人" :bordered="false">
              <template #header-extra>
                <n-button
                  type="primary"
                  size="small"
                  @click="cardholderModalVisible = true"
                  >新增持卡人</n-button
                >
              </template>
              <n-data-table
                :columns="cardholderColumns"
                :data="cardholders"
                :loading="loading"
                :bordered="false"
              />
            </n-card>
          </section>

          <section v-else-if="selectedPage === 'cards'">
            <n-card title="卡片" :bordered="false">
              <n-data-table
                :columns="cardColumns"
                :data="cards"
                :loading="loading"
                :bordered="false"
              />
            </n-card>
          </section>

          <section v-else-if="selectedPage === 'transactions'">
            <n-card title="交易" :bordered="false">
              <n-data-table
                :columns="transactionColumns"
                :data="transactions"
                :loading="loading"
                :bordered="false"
              />
            </n-card>
          </section>

          <section v-else>
            <authorization-view @completed="handleAuthorizationComplete" />
          </section>
        </n-layout-content>
      </n-layout>
    </n-layout>
    <n-modal
      v-model:show="cardholderModalVisible"
      preset="card"
      title="新增持卡人"
      style="width: 480px"
    >
      <n-form label-placement="top">
        <n-form-item label="名字"
          ><n-input v-model:value="holderForm.firstName"
        /></n-form-item>
        <n-form-item label="姓氏"
          ><n-input v-model:value="holderForm.lastName"
        /></n-form-item>
        <n-form-item label="邮箱"
          ><n-input v-model:value="holderForm.email"
        /></n-form-item>
        <n-form-item label="手机号"
          ><n-input v-model:value="holderForm.mobile"
        /></n-form-item>
      </n-form>
      <n-divider />
      <n-space justify="end"
        ><n-button @click="cardholderModalVisible = false">取消</n-button
        ><n-button type="primary" @click="createCardholder"
          >创建</n-button
        ></n-space
      >
    </n-modal>
  </n-config-provider>
</template>
