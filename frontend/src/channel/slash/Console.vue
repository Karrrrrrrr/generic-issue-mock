<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue'
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
} from 'naive-ui'

import { api } from '@/channel/slash/api'
import type { Card, Cardholder, Transaction } from '@/channel/types'

const channelLabel = 'Slash'

const { message } = createDiscreteApi(['message'])
const selectedPage = ref('cardholders')
const loading = ref(false)
const cardholders = ref<Cardholder[]>([])
const cards = ref<Card[]>([])
const transactions = ref<Transaction[]>([])
const cardholderModalVisible = ref(false)
const holderForm = ref({ firstName: '', lastName: '', email: '', mobile: '' })
const authorizationForm = ref({ cardID: '', amount: 1, currency: 'USD', merchantName: '', merchantMCC: '', merchantCountry: 'US' })

const menuOptions = computed(() => {
  const options = [
    { label: '持卡人', key: 'cardholders' },
    { label: '卡片', key: 'cards' },
    { label: '交易', key: 'transactions' },
  ]
  options.push({ label: '授权模拟', key: 'simulation' })
  return options
})

const cardholderOptions = computed(() => cardholders.value.map((item) => ({
  label: `${item.first_name} ${item.last_name} (${item.id})`,
  value: item.id,
})))

const cardOptions = computed(() => cards.value.map((item) => ({
  label: `${item.card_number} (${item.id})`,
  value: item.id,
})))

const cardholderColumns = [
  { title: 'ID', key: 'id', ellipsis: { tooltip: true } },
  { title: '姓名', key: 'name', render: (row: Cardholder) => `${row.first_name} ${row.last_name}` },
  { title: '邮箱', key: 'email' },
  { title: '手机号', key: 'phone_number' },
  { title: '状态', key: 'status', render: (row: Cardholder) => h(NTag, { size: 'small', type: 'success' }, { default: () => row.status }) },
]

const cardColumns = [
  { title: 'ID', key: 'id', ellipsis: { tooltip: true } },
  { title: '卡号', key: 'card_number' },
  { title: '币种', key: 'card_currency' },
  { title: '状态', key: 'card_status', render: (row: Card) => h(NTag, { size: 'small', type: row.card_status === 'active' ? 'success' : 'warning' }, { default: () => row.card_status }) },
  { title: 'CVV', key: 'cvv' },
]

const transactionColumns = [
  { title: '交易 ID', key: 'id', ellipsis: { tooltip: true } },
  { title: '类型', key: 'transaction_type' },
  { title: '金额', key: 'amount', render: (row: Transaction) => `${row.currency} ${row.amount}` },
  { title: '商户', key: 'merchant_name' },
  { title: '状态', key: 'status', render: (row: Transaction) => h(NTag, { size: 'small' }, { default: () => row.status }) },
  {
    title: '操作',
    key: 'actions',
    render: (row: Transaction) => h(NSpace, { size: 4 }, {
      default: () => [
        h(NButton, { size: 'tiny', onClick: () => applyStep(row.id, 'clear') }, { default: () => '清算' }),
        h(NButton, { size: 'tiny', onClick: () => applyStep(row.id, 'reverse') }, { default: () => '撤销' }),
        h(NButton, { size: 'tiny', type: 'warning', onClick: () => applyStep(row.id, 'refund') }, { default: () => '退款' }),
      ],
    }),
  },
]

async function loadCardholders() {
  const result = await api.listCardholders()
  cardholders.value = result.data
}

async function loadCards() {
  const result = await api.listCards()
  cards.value = result.data
}

async function loadTransactions() {
  const result = await api.listTransactions()
  transactions.value = result.data
}

async function refresh() {
  loading.value = true
  try {
    await Promise.all([loadCardholders(), loadCards(), loadTransactions()])
  } catch (error) {
    message.error(error instanceof Error ? error.message : '加载控制台数据失败')
  } finally {
    loading.value = false
  }
}

async function createCardholder() {
  try {
    await api.createCardholder(holderForm.value)
    cardholderModalVisible.value = false
    holderForm.value = { firstName: '', lastName: '', email: '', mobile: '' }
    await loadCardholders()
    message.success('持卡人已创建')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '创建持卡人失败')
  }
}

async function simulateAuthorization() {
  try {
    await api.simulateAuthorization(authorizationForm.value)
    await loadTransactions()
    message.success('授权交易已创建')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '模拟授权失败')
  }
}

async function applyStep(id: string, action: 'clear' | 'reverse' | 'refund') {
  try {
    await api.applyTransactionStep(id, action)
    await loadTransactions()
    message.success('交易状态已更新')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '交易操作失败')
  }
}

onMounted(refresh)
</script>

<template>
  <n-config-provider>
      <n-layout class="app-shell">
      <n-layout-header class="app-header">
        <div>
          <strong>Generic Mock</strong>
          <span>{{ channelLabel }} Console</span>
        </div>
        <n-button size="small" :loading="loading" @click="refresh">刷新</n-button>
      </n-layout-header>
      <n-layout has-sider>
        <aside class="sidebar">
          <n-menu v-model:value="selectedPage" :options="menuOptions" />
        </aside>
        <n-layout-content class="content">
          <section v-if="selectedPage === 'cardholders'">
            <n-card title="持卡人" :bordered="false">
              <template #header-extra>
                <n-button type="primary" size="small" @click="cardholderModalVisible = true">新增持卡人</n-button>
              </template>
              <n-data-table :columns="cardholderColumns" :data="cardholders" :loading="loading" :bordered="false" />
            </n-card>
          </section>

          <section v-else-if="selectedPage === 'cards'">
            <n-card title="卡片" :bordered="false">
              <n-data-table :columns="cardColumns" :data="cards" :loading="loading" :bordered="false" />
            </n-card>
          </section>

          <section v-else-if="selectedPage === 'transactions'">
            <n-card title="交易" :bordered="false">
              <n-data-table :columns="transactionColumns" :data="transactions" :loading="loading" :bordered="false" />
            </n-card>
          </section>

          <section v-else>
            <n-card title="授权模拟" :bordered="false" class="simulation-card">
              <n-form label-placement="top">
                <n-form-item label="卡片" required>
                  <n-select v-model:value="authorizationForm.cardID" :options="cardOptions" filterable />
                </n-form-item>
                <n-form-item label="金额" required>
                  <n-input-number v-model:value="authorizationForm.amount" :min="0.01" :precision="2" />
                </n-form-item>
                <n-form-item label="币种">
                  <n-select v-model:value="authorizationForm.currency" :options="[{ label: 'USD', value: 'USD' }, { label: 'GBP', value: 'GBP' }, { label: 'CNY', value: 'CNY' }]" />
                </n-form-item>
                <n-form-item label="商户名称" required>
                  <n-input v-model:value="authorizationForm.merchantName" />
                </n-form-item>
                <n-form-item label="MCC" required>
                  <n-input v-model:value="authorizationForm.merchantMCC" placeholder="例如 5411" />
                </n-form-item>
                <n-form-item label="商户国家">
                  <n-input v-model:value="authorizationForm.merchantCountry" />
                </n-form-item>
                <n-button type="primary" @click="simulateAuthorization">创建授权</n-button>
              </n-form>
            </n-card>
          </section>
        </n-layout-content>
      </n-layout>
      </n-layout>
      <n-modal v-model:show="cardholderModalVisible" preset="card" title="新增持卡人" style="width: 480px">
        <n-form label-placement="top">
          <n-form-item label="名字"><n-input v-model:value="holderForm.firstName" /></n-form-item>
          <n-form-item label="姓氏"><n-input v-model:value="holderForm.lastName" /></n-form-item>
          <n-form-item label="邮箱"><n-input v-model:value="holderForm.email" /></n-form-item>
          <n-form-item label="手机号"><n-input v-model:value="holderForm.mobile" /></n-form-item>
        </n-form>
        <n-divider />
        <n-space justify="end"><n-button @click="cardholderModalVisible = false">取消</n-button><n-button type="primary" @click="createCardholder">创建</n-button></n-space>
      </n-modal>
  </n-config-provider>
</template>
