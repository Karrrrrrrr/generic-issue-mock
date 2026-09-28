<script setup lang="ts">
import type { MenuOption } from "naive-ui";
import ChannelShell from "@/channel/shared/ChannelShell.vue";

const items: MenuOption[] = [
  {
    type: "group",
    label: "账户",
    key: "accounts",
    children: [
      {
        label: "账户",
        key: "/pingpong/accounts",
      },
      {
        label: "虚拟账户",
        key: "/pingpong/virtual-accounts",
      },
    ],
  },
  {
    type: "group",
    label: "卡",
    key: "cards",
    children: [
      {
        label: "卡管理",
        key: "/pingpong/cards",
      },
      {
        label: "卡产品",
        key: "/pingpong/products",
      },
    ],
  },
  {
    type: "group",
    label: "交易",
    key: "transactions",
    children: [
      {
        label: "授权管理",
        key: "/pingpong/authorizations",
      },
      {
        label: "交易处理",
        key: "/pingpong/transactions",
      },
      {
        label: "交易模拟",
        key: "/pingpong/simulation",
      },
      {
        label: "资金订单",
        key: "/pingpong/transfers",
      },
    ],
  },
  {
    type: "group",
    label: "通知",
    key: "notifications",
    children: [
      { label: "Webhook 配置", key: "/pingpong/webhooks" },
      { label: "投递记录", key: "/pingpong/webhook-records" },
    ],
  },
];
</script>

<template>
  <ChannelShell channel="pingpong" title="PingPong Mock" :items="items" />
</template>
<style>
.ping-page {
  display: flex;
  flex-direction: column;
  gap: 24px;
  min-width: 0;
}
.ping-toolbar {
  margin-bottom: 24px;
}
.ping-page .n-data-table__pagination {
  margin-top: 24px;
  padding-bottom: 8px;
}
</style>
