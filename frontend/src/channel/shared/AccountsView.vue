<script setup lang="ts">
import { formatDateTime } from "@/channel/dateTime";
import { pageSizes } from "@/channel/pagination";
import { renderAmountTag } from "@/channel/tableTags";
import { h, onMounted, ref } from "vue";
import {
  createDiscreteApi,
  type DataTableColumns,
  NButton,
  NDataTable,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NModal,
  NPagination,
  NSpace,
} from "naive-ui";
import type { Account, AccountListRequest, ManagementAPI } from "./api";

const {
  accountApi,
  adjustBalance,
  title = "账户",
  showWalletId = false,
} = defineProps<{
  accountApi: Pick<ManagementAPI["accountApi"], "list" | "create"> & Partial<Pick<ManagementAPI["accountApi"], "update">>;
  adjustBalance: (account: Account, amount: number) => Promise<unknown>;
  title?: string;
  showWalletId?: boolean;
}>();

const { message } = createDiscreteApi(["message"]);
const loading = ref(false);
const visible = ref(false);
const editingId = ref<number>();
const name = ref("");
const currency = ref("USD");
const rows = ref<Account[]>([]);
const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const accountIdFilter = ref("");
const nameFilter = ref("");
const appliedFilters = ref<AccountListRequest>({});
const adjustingAccount = ref<Account>();
const adjustment = ref<number | null>(null);
const adjusting = ref(false);

function openAdjustment(account: Account) {
  adjustingAccount.value = account;
  adjustment.value = null;
}

async function submitAdjustment() {
  if (!adjustingAccount.value || !adjustment.value || !Number.isFinite(adjustment.value)) {
    message.warning("请输入非零调整金额，正数增加、负数扣减");
    return;
  }
  adjusting.value = true;
  try {
    await adjustBalance(adjustingAccount.value, adjustment.value);
    adjustingAccount.value = undefined;
    message.success("账户资金已调整");
    await load();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "资金调整失败");
  } finally {
    adjusting.value = false;
  }
}

async function load() {
  loading.value = true;
  try {
    const response = await accountApi.list(page.value, pageSize.value, appliedFilters.value);
    rows.value = response.data;
    total.value = response.total_items;
  } catch (error) {
    message.error(error instanceof Error ? error.message : "无法加载账户");
  } finally {
    loading.value = false;
  }
}

function updateAccountIdFilter(value: string) {
  accountIdFilter.value = value.replace(/\D/g, "");
}

function buildFilters(): AccountListRequest | null {
  const filters: AccountListRequest = {};
  const id = accountIdFilter.value.trim();
  if (id !== "") {
    const idValue = Number(id);
    if (!Number.isSafeInteger(idValue) || idValue <= 0) {
      message.warning("账户 ID 必须是正整数");
      return null;
    }
    filters.id = idValue;
  }
  const name = nameFilter.value.trim();
  if (name !== "") {
    filters.name = name;
  }
  return filters;
}

function search() {
  const filters = buildFilters();
  if (filters === null) {
    return;
  }
  appliedFilters.value = filters;
  page.value = 1;
  void load();
}

function reset() {
  accountIdFilter.value = "";
  nameFilter.value = "";
  appliedFilters.value = {};
  page.value = 1;
  void load();
}

function openCreate() {
  editingId.value = undefined;
  name.value = "";
  currency.value = "USD";
  visible.value = true;
}

function openEdit(account: Account) {
  editingId.value = account.id;
  name.value = account.name;
  visible.value = true;
}

async function submit() {
  if (!name.value || (!editingId.value && !/^[A-Z]{3}$/.test(currency.value))) {
    message.warning("请填写名称和三位大写币种代码");
    return;
  }

  try {
    if (editingId.value) {
      await accountApi.update!(editingId.value, { name: name.value });
    } else {
      await accountApi.create({ name: name.value, currency: currency.value });
    }
    visible.value = false;
    await load();
  } catch (error) {
    message.error(error instanceof Error ? error.message : "保存失败");
  }
}

function changePageSize(value: number) {
  pageSize.value = value;
  page.value = 1;
  void load();
}

const columns: DataTableColumns<Account> = [
  ...(showWalletId ? [{ title: "钱包 ID", key: "wallet_id" }] : []),
  {
    title: "名称",
    key: "name",
  },
  {
    title: "账户 ID",
    key: "id",
  },
  {
    title: "余额",
    key: "balance",
    render: (row) => renderAmountTag(row.balance, row.currency ?? "USD"),
  },
  {
    title: "创建时间",
    key: "created_at",
    render: (row) => formatDateTime(row.created_at),
  },
  {
    title: "操作",
    key: "actions",
    render: (row) =>
      h(
        NSpace,
        { size: 8 },
        {
          default: () => [
            ...(accountApi.update ? [h(NButton, { onClick: () => openEdit(row) }, { default: () => "改名" })] : []),
            h(NButton, { onClick: () => openAdjustment(row) }, { default: () => "调整资金" }),
          ],
        },
      ),
  },
];

onMounted(() => void load());
</script>

<template>
  <section>
    <div class="page-heading">
      <div>
        <h1>{{ title }}</h1>
        <slot name="description" />
      </div>
      <n-button type="primary" @click="openCreate">新增账户</n-button>
    </div>
    <form class="table-filters" @submit.prevent="search">
      <div class="compact-filter-grid">
        <n-input
          class="compact-filter-id"
          :value="accountIdFilter"
          placeholder="账户 ID（精确）"
          clearable
          @update:value="updateAccountIdFilter"
        />
        <n-input
          v-model:value="nameFilter"
          class="compact-filter-name"
          placeholder="名称（模糊）"
          clearable
        />
      </div>
      <div class="filter-actions">
        <n-button attr-type="submit" type="primary" :loading="loading">查询</n-button>
        <n-button :disabled="loading" @click="reset">重置</n-button>
      </div>
    </form>
    <n-data-table
      max-height="max(160px, calc(100dvh - 400px))"
      :scroll-x="1000"
      table-layout="fixed"
      :columns="columns"
      :data="rows"
      :loading="loading"
      :bordered="true"
    />
    <n-pagination
      v-model:page="page"
      :page-size="pageSize"
      :item-count="total"
      :page-sizes="pageSizes"
      show-size-picker
      @update:page="load"
      @update:page-size="changePageSize"
    />
    <n-modal
      v-model:show="visible"
      preset="card"
      :title="editingId ? '账户改名' : '新增账户'"
      style="width: min(440px, calc(100vw - 32px))"
    >
      <n-form label-placement="top">
        <n-form-item label="名称">
          <n-input v-model:value="name" />
        </n-form-item>
        <n-form-item v-if="!editingId" label="币种">
          <n-input v-model:value="currency" placeholder="例如 USD" maxlength="3" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space justify="end">
          <n-button @click="visible = false">取消</n-button>
          <n-button type="primary" @click="submit">保存</n-button>
        </n-space>
      </template>
    </n-modal>
    <n-modal
      :show="Boolean(adjustingAccount)"
      preset="card"
      style="width: min(520px, calc(100vw - 32px))"
      title="账户资金调整"
      @update:show="
        (shown) => {
          if (!shown) adjustingAccount = undefined;
        }
      "
    >
      <p>{{ adjustingAccount?.name }} · 当前余额 {{ adjustingAccount?.balance }}</p>
      <p>直接增加或扣减账户余额，不从其他钱包取款。正数增加，负数扣减。</p>
      <n-form>
        <n-form-item label="调整金额">
          <n-input-number v-model:value="adjustment" :precision="2" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-button type="primary" :loading="adjusting" @click="submitAdjustment">确认调整</n-button>
      </template>
    </n-modal>
  </section>
</template>
