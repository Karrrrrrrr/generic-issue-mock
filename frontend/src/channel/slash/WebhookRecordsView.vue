<script setup lang="ts">
import TableFilters, { type FilterField } from "@/channel/TableFilters.vue";
import { useTableFilters } from "@/channel/tableFilters";
import { accountApi } from "./api";
import { formatDateTime } from "@/channel/dateTime";
import { formatEnumLabel, renderEnumTag } from "@/channel/tableTags";
import { h, onMounted, ref } from "vue";
import {
  createDiscreteApi,
  type DataTableColumns,
  NButton,
  NDataTable,
  NDescriptions,
  NDescriptionsItem,
  NInput,
  NModal,
  NPagination,
  NSpace,
  NTag,
} from "naive-ui";
import { type WebhookRecord, webhookRecordApi } from "./api";

const { dialog, message } = createDiscreteApi(["dialog", "message"]);
const loading = ref(false);
const detailVisible = ref(false);
const selected = ref<WebhookRecord>();
const rows = ref<WebhookRecord[]>([]);
const page = ref(1);
const pageSize = ref(20);
const total = ref(0);

const filterFields: FilterField[] = [
  {
    "key": "event",
    "label": "投递事件",
    "options": [
      {
        "label": "交易创建",
        "value": "aggregated_transaction.create"
      },
      {
        "label": "交易更新",
        "value": "aggregated_transaction.update"
      },
      {
        "label": "卡片创建",
        "value": "card_creation.event"
      },
      {
        "label": "卡片更新",
        "value": "card.update"
      },
      {
        "label": "卡片删除",
        "value": "card.delete"
      }
    ]
  },
  {
    "key": "status",
    "label": "投递状态",
    "options": [
      {
        "label": "处理中",
        "value": "pending"
      },
      {
        "label": "成功",
        "value": "succeeded"
      },
      {
        "label": "失败",
        "value": "failed"
      }
    ]
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
    const response = await webhookRecordApi.list({
      ...appliedFilters.value,
      page_number: page.value,
      page_size: pageSize.value,
    });
    rows.value = response.data;
    total.value = response.total_items;
  } catch (error) {
    message.error(error instanceof Error ? error.message : "无法加载投递记录");
  } finally {
    loading.value = false;
  }
}

function openDetail(record: WebhookRecord) {
  selected.value = record;
  detailVisible.value = true;
}

function replay(record: WebhookRecord) {
  dialog.warning({
    title: "重放 Webhook",
    content: "将使用原始报文和请求头再次投递，并生成新的投递记录。",
    positiveText: "重放",
    negativeText: "取消",
    onPositiveClick: async () => {
      try {
        await webhookRecordApi.replay(record);
        page.value = 1;
        await load();
      } catch (error) {
        message.error(error instanceof Error ? error.message : "重放失败");
      }
    },
  });
}

function changePageSize(value: number) {
  pageSize.value = value;
  page.value = 1;
  void load();
}

const columns: DataTableColumns<WebhookRecord> = [
  {
    title: "账户名称",
    key: "account_name",
  },
  {
    title: "事件",
    key: "event",
    render: (row) => renderEnumTag(row.event, "event"),
  },
  {
    title: "目标地址",
    key: "target_url",
    ellipsis: { tooltip: true },
  },
  {
    title: "状态",
    key: "status",
    render: (row) => renderEnumTag(row.status, "status"),
  },
  {
    title: "HTTP",
    key: "status_code",
    width: 90,
  },
  {
    title: "尝试",
    key: "attempt_count",
    width: 70,
  },
  {
    title: "时间",
    key: "created_at",
    render: (row) => formatDateTime(row.created_at),
  },
  {
    title: "操作",
    key: "actions",
    render: (row) =>
      h(
        NSpace,
        { size: 6 },
        {
          default: () => [
            h(
              NButton,
              { size: "small", onClick: () => openDetail(row) },
              { default: () => "详情" },
            ),
            h(
              NButton,
              { size: "small", type: "primary", onClick: () => replay(row) },
              { default: () => "重放" },
            ),
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
        <h1>Webhook 投递记录</h1>
        <p>查看 Slash 投递结果、请求与响应报文，支持重新投递。</p>
      </div>
      <n-button :loading="loading" @click="load">刷新</n-button>
    </div>
    <TableFilters
      v-model:values="filters"
      v-model:date-range="dateRange"
      :fields="filterFields"
      :account-options="accountOptions"
      :accounts-loading="accountsLoading"
      :loading="loading"
      @load-accounts="loadAccounts"
      @search="search"
      @reset="reset"
    />
    <n-data-table
      max-height="max(160px, calc(100dvh - 580px))"
      :scroll-x="1200"
      table-layout="fixed"
      :columns="columns"
      :data="rows"
      :loading="loading"
      :bordered="false"
    />
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
      v-model:show="detailVisible"
      preset="card"
      title="投递详情"
      style="width: min(900px, calc(100vw - 32px))"
    >
      <n-descriptions
        v-if="selected"
        class="webhook-detail-summary"
        :column="2"
        label-placement="left"
      >
        <n-descriptions-item label="记录 ID">{{ selected.id }}</n-descriptions-item>
        <n-descriptions-item label="所属账户">{{ selected.account_name }}</n-descriptions-item>
        <n-descriptions-item label="资源 ID">{{ selected.source_id }}</n-descriptions-item>
        <n-descriptions-item label="创建时间">{{ formatDateTime(selected.created_at) }}</n-descriptions-item>
        <n-descriptions-item label="投递时间">{{ formatDateTime(selected.delivered_at) }}</n-descriptions-item>
        <n-descriptions-item label="投递次数">{{ selected.attempt_count }}</n-descriptions-item>
        <n-descriptions-item label="错误信息">{{ selected.error_message || "—" }}</n-descriptions-item>
        <n-descriptions-item label="事件">
          <n-tag :bordered="true" size="small" :title="selected.event">
            {{ formatEnumLabel(selected.event, "event") }}
          </n-tag>
        </n-descriptions-item>
        <n-descriptions-item label="状态">
          <n-tag :bordered="true" size="small" :title="selected.status">
            {{ formatEnumLabel(selected.status, "status") }}
          </n-tag>
        </n-descriptions-item>
        <n-descriptions-item label="目标地址">{{ selected.target_url }}</n-descriptions-item>
        <n-descriptions-item label="HTTP 状态">{{ selected.status_code }}</n-descriptions-item>
      </n-descriptions>
      <n-space vertical :size="16">
        <div>报文</div>
        <n-input
          :value="selected?.payload"
          type="textarea"
          readonly
          :autosize="{ minRows: 6, maxRows: 12 }"
        />
        <div>请求头</div>
        <n-input
          :value="selected?.request_headers"
          type="textarea"
          readonly
          :autosize="{ minRows: 4, maxRows: 8 }"
        />
        <div>响应体</div>
        <n-input
          :value="selected?.response_body"
          type="textarea"
          readonly
          :autosize="{ minRows: 4, maxRows: 8 }"
        />
        <div>响应头</div>
        <n-input
          :value="selected?.response_headers"
          type="textarea"
          readonly
          :autosize="{ minRows: 4, maxRows: 8 }"
        />
      </n-space>
    </n-modal>
  </section>
</template>
