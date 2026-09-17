<script setup lang="ts">
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
import { type Account, accountApi, fundsApi } from "./api";

const { message } = createDiscreteApi(["message"]);
const loading = ref(false);
const visible = ref(false);
const editingID = ref<string>();
const name = ref("");
const rows = ref<Account[]>([]);
const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const adjustingAccount = ref<Account>();
const adjustment = ref<number | null>(null);
const adjusting = ref(false);

function openAdjustment(account: Account) {
  adjustingAccount.value = account;
  adjustment.value = null;
}

async function submitAdjustment() {
  if (!adjustingAccount.value || !adjustment.value) {
    message.warning("请输入非零调整金额，正数增加、负数扣减");
    return;
  }
  adjusting.value = true;
  try {
    await fundsApi.adjustAccount(adjustingAccount.value, adjustment.value);
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
    const response = await accountApi.list(page.value, pageSize.value);
    rows.value = response.data;
    total.value = response.total_items;
  } catch (error) {
    message.error(error instanceof Error ? error.message : "无法加载账户");
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  editingID.value = undefined;
  name.value = "";
  visible.value = true;
}

function openEdit(account: Account) {
  editingID.value = account.id;
  name.value = account.name;
  visible.value = true;
}

async function submit() {
  if (!name.value) {
    return;
  }

  try {
    if (editingID.value) {
      await accountApi.update(editingID.value, { name: name.value });
    } else {
      await accountApi.create({ name: name.value });
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
  },
  {
    title: "创建时间",
    key: "created_at",
    render: (row) => h("span", new Date(row.created_at).toLocaleString()),
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
            h(NButton, { onClick: () => openEdit(row) }, { default: () => "改名" }),
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
        <h1>账户</h1>
      </div>
      <n-button type="primary" @click="openCreate">新增账户</n-button>
    </div>
    <n-data-table :columns="columns" :data="rows" :loading="loading" :bordered="false" />
    <n-pagination
      v-model:page="page"
      :page-size="pageSize"
      :item-count="total"
      :page-sizes="[10, 20, 50]"
      show-size-picker
      @update:page="load"
      @update:page-size="changePageSize"
    />
    <n-modal
      v-model:show="visible"
      preset="card"
      :title="editingID ? '账户改名' : '新增账户'"
      style="width: min(440px, calc(100vw - 32px))"
    >
      <n-form label-placement="top">
        <n-form-item label="名称">
          <n-input v-model:value="name" />
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
