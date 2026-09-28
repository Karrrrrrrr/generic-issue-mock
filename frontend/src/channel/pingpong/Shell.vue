<script setup lang="ts">
import ChannelShell from "@/channel/shared/ChannelShell.vue";
const items = [
  {
    label: "账户",
    key: "/pingpong/accounts"
  },
  {
    label: "虚拟账户",
    key: "/pingpong/virtual-accounts"
  },
  {
    label: "卡产品",
    key: "/pingpong/products"
  },
  {
    label: "卡管理",
    key: "/pingpong/cards"
  },
  {
    label: "授权管理",
    key: "/pingpong/authorizations"
  },
  {
    label: "资金订单",
    key: "/pingpong/transfers"
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
