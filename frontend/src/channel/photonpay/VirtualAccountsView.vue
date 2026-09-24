<script setup lang="ts">
import { useClientPagination } from "@/channel/pagination";
import { renderAmountTag } from "@/channel/tableTags";
import { accountApi } from "./api";
import { h, onMounted, ref } from "vue";
import {
  createDiscreteApi,
  NButton,
  NCard,
  NDataTable,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NModal,
  NSelect,
  NSpace,
} from "naive-ui";
import { request } from "@/channel/shared";
import { fundsApi, type Wallet } from "./api";

type VirtualAccount = {
  id: string;
  account_id: string;
  account_name: string;
  name: string;
  wallet_id: string;
};
const baseURL = "/photonpay/ui";
const { message } = createDiscreteApi(["message"]);
const rows = ref<VirtualAccount[]>([]);
const pagination = useClientPagination(rows);
const wallets = ref<Wallet[]>([]);
const fundingAccount = ref<VirtualAccount>();
const fundingAmount = ref<number | null>(null);
const withdrawing = ref(false);
const funding = ref(false);

function openFunding(account: VirtualAccount, withdraw: boolean) {
  fundingAccount.value = account;
  fundingAmount.value = null;
  withdrawing.value = withdraw;
}

async function submitFunding() {
  if (!fundingAccount.value || !fundingAmount.value || fundingAmount.value <= 0) {
    message.warning("请输入正数金额");
    return;
  }
  funding.value = true;
  try {
    await fundsApi.transferResource({
      accountID: fundingAccount.value.account_id,
      walletID: fundingAccount.value.wallet_id,
      amount: fundingAmount.value,
      withdraw: withdrawing.value,
    });
    fundingAccount.value = undefined;
    message.success("资金划转完成");
    await load();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "资金划转失败");
  } finally {
    funding.value = false;
  }
}

const visible = ref(false);
const saving = ref(false);
const name = ref("");
const accountID = ref<string | null>(null);
const accounts = ref<{ label: string; value: string }[]>([]);
const currency = ref("USD");
const currencies = ["USD", "GBP", "JPY", "CNY"].map((value) => ({ label: value, value }));
const columns = [
  {
    title: "账户名称",
    key: "account_name",
  },
  {
    title: "余额",
    key: "balance",
    render: (account: VirtualAccount) => {
      const wallet = wallets.value.find((item) => item.id === account.wallet_id);
      return wallet ? renderAmountTag(wallet.amount, wallet.currency) : "—";
    },
  },
  {
    title: "所属账户 ID",
    key: "account_id",
  },
  {
    title: "虚拟账户 ID",
    key: "id",
  },
  {
    title: "名称",
    key: "name",
  },
  {
    title: "钱包 ID",
    key: "wallet_id",
  },
  {
    title: "操作",
    key: "actions",
    render: (account: VirtualAccount) =>
      h(
        NSpace,
        {},
        {
          default: () => [
            h(
              NButton,
              {
                size: "small",
                onClick: () => openFunding(account, false),
              },
              {
                default: () => "充值",
              },
            ),
            h(
              NButton,
              {
                size: "small",
                onClick: () => openFunding(account, true),
              },
              {
                default: () => "转出",
              },
            ),
          ],
        },
      ),
  },
];

async function load() {
  try {
    const [accountsResponse, walletRows] = await Promise.all([
      request.get<VirtualAccount[]>(baseURL + "/managed-virtual-accounts"),
      fundsApi.list(),
    ]);
    rows.value = accountsResponse.data;
    wallets.value = walletRows;
  } catch (error) {
    message.error(error instanceof Error ? error.message : "加载失败");
  }
}

async function create() {
  if (!name.value.trim() || !accountID.value) {
    message.warning("请填写名称及所属账户");
    return;
  }
  saving.value = true;
  try {
    await request.post(baseURL + "/managed-virtual-accounts", {
      account_id: accountID.value,
      name: name.value.trim(),
      currency: currency.value,
    });
    visible.value = false;
    name.value = "";
    await load();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "创建失败");
  } finally {
    saving.value = false;
  }
}

async function openCreate() {
  try {
    const accountRows = await accountApi.listAll();
    accounts.value = accountRows.map((account: { id: string; name: string }) => ({
      label: account.name + " · " + account.id,
      value: account.id,
    }));
    accountID.value = null;
    visible.value = true;
  } catch (error) {
    message.error(error instanceof Error ? error.message : "加载账户失败");
  }
}

onMounted(load);
</script>

<template>
  <div class="page-heading">
    <h1>虚拟账户</h1>
    <n-space>
      <n-button @click="load">刷新</n-button>
      <n-button type="primary" @click="openCreate">新增虚拟账户</n-button>
    </n-space>
  </div>
  <n-card :bordered="false">
    <n-data-table
      max-height="max(160px, calc(100dvh - 400px))"
      :pagination="pagination"
      :scroll-x="1400"
      table-layout="fixed"
      :columns="columns"
      :data="rows"
    />
  </n-card>
  <n-modal v-model:show="visible" preset="card" title="新增虚拟账户">
    <n-form>
      <n-form-item label="所属账户">
        <n-select v-model:value="accountID" :options="accounts" filterable />
      </n-form-item>
      <n-form-item label="名称">
        <n-input v-model:value="name" />
      </n-form-item>
      <n-form-item label="币种">
        <n-select v-model:value="currency" :options="currencies" />
      </n-form-item>
    </n-form>
    <template #action>
      <n-button type="primary" :loading="saving" @click="create">创建</n-button>
    </template>
  </n-modal>
  <n-modal
    :show="Boolean(fundingAccount)"
    preset="card"
    :title="withdrawing ? '虚拟账户资金转出到账户' : '账户资金充值到虚拟账户'"
    @update:show="
      (shown) => {
        if (!shown) fundingAccount = undefined;
      }
    "
  >
    <p>
      {{ fundingAccount?.name }} · 关联账户 {{ fundingAccount?.account_name || "账户不存在" }} ·
      {{ fundingAccount?.account_id }}
    </p>
    <p>充值从关联账户钱包扣款；转出退回同一账户钱包。</p>
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
