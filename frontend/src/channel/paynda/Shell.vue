<script setup lang="ts">
import ChannelShell from "@/channel/shared/ChannelShell.vue";
const items = [
  {
    type: "group",
    label: "账户",
    key: "accounts",
    children: [{ label: "资金账户", key: "/paynda/accounts" }],
  },
  {
    type: "group",
    label: "卡",
    key: "cards",
    children: [
      { label: "持卡人", key: "/paynda/cardholders" },
      { label: "卡片管理", key: "/paynda/cards" },
      {
        label: "卡产品管理",
        key: "/paynda/card-products",
      },
    ],
  },
  {
    type: "group",
    label: "交易",
    key: "transactions",
    children: [
      { label: "授权管理", key: "/paynda/authorizations" },
      { label: "卡交易", key: "/paynda/transactions" },
      { label: "模拟交易", key: "/paynda/authorization" },
    ],
  },
  {
    type: "group",
    label: "配置",
    key: "configuration",
    children: [
      { label: "Webhook 管理", key: "/paynda/webhooks" },
      { label: "投递记录", key: "/paynda/webhook-records" },
    ],
  },
];
</script>

<template>
  <ChannelShell channel="paynda" title="Paynda Mock" :items="items" />
</template>
