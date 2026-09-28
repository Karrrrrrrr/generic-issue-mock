<script setup lang="ts">
import ChannelShell from "@/channel/shared/ChannelShell.vue";
const items = [
  {
    type: "group",
    label: "账户",
    key: "accounts",
    children: [
      { label: "账户", key: "/slash/accounts" },
      { label: "虚拟账户", key: "/slash/virtual-accounts" },
    ],
  },
  {
    type: "group",
    label: "卡",
    key: "cards",
    children: [
      { label: "持卡人", key: "/slash/cardholders" },
      { label: "卡片管理", key: "/slash/cards" },
      { label: "卡产品管理", key: "/slash/card-products" },
    ],
  },
  {
    type: "group",
    label: "交易",
    key: "transactions",
    children: [
      { label: "授权管理", key: "/slash/authorizations" },
      { label: "卡交易", key: "/slash/transactions" },
      { label: "模拟交易", key: "/slash/authorization" },
    ],
  },
  {
    type: "group",
    label: "配置",
    key: "configuration",
    children: [
      { label: "Webhook 管理", key: "/slash/webhooks" },
      { label: "投递记录", key: "/slash/webhook-records" },
    ],
  },
];
</script>

<template>
  <ChannelShell channel="slash" title="Slash Mock" :items="items" />
</template>
