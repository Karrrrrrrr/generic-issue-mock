<script setup lang="ts">
import ChannelShell from "@/channel/shared/ChannelShell.vue";
const items = [
  {
    type: "group",
    label: "账户",
    key: "accounts",
    children: [
        { label: "账户", key: "/photonpay/accounts" },
        { label: "虚拟账户", key: "/photonpay/virtual-accounts" },
    ],
  },
  {
    type: "group",
    label: "卡",
    key: "cards",
    children: [
      { label: "持卡人", key: "/photonpay/cardholders" },
      { label: "卡片管理", key: "/photonpay/cards" },
    ],
  },
  {
    type: "group",
    label: "交易",
    key: "transactions",
    children: [
      { label: "授权管理", key: "/photonpay/authorizations" },
      { label: "卡交易", key: "/photonpay/transactions" },
      { label: "模拟交易", key: "/photonpay/authorization" },
    ],
  },
  {
    type: "group",
    label: "配置",
    key: "configuration",
    children: [
      { label: "Webhook 管理", key: "/photonpay/webhooks" },
      { label: "投递记录", key: "/photonpay/webhook-records" },
    ],
  },
];
</script>

<template>
  <ChannelShell channel="photonpay" title="PhotonPay Mock" :items="items" />
</template>
