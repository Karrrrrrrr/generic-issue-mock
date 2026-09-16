<script setup lang="ts">
import {
  NLayout,
  NLayoutContent,
  NLayoutHeader,
  NLayoutSider,
  NMenu,
} from "naive-ui";
import { useRoute, useRouter } from "vue-router";
import ChannelSwitcher from "@/channel/ChannelSwitcher.vue";
import ThemeToggle from "@/channel/ThemeToggle.vue";
const route = useRoute();
const router = useRouter();
const items = [
  { type: "group", label: "卡", key: "cards", children: [
    { label: "持卡人", key: "/photonpay/cardholders" },
    { label: "卡片管理", key: "/photonpay/cards" },
  ] },
  { type: "group", label: "交易", key: "transactions", children: [
    { label: "卡交易", key: "/photonpay/transactions" },
    { label: "模拟交易", key: "/photonpay/authorization" },
  ] },
];
</script>
<template>
  <n-layout class="app-shell" has-sider>
    <n-layout-sider class="sidebar" :width="208" bordered>
      <div class="channel-logo">PhotonPay Mock</div>
      <n-menu
        :value="route.path"
        :options="items"
        @update:value="(path) => router.push(String(path))"
      />
    </n-layout-sider>
    <n-layout>
      <n-layout-header class="app-header">
        <span></span>
        <div class="header-actions"><ThemeToggle /><ChannelSwitcher current="photonpay" /></div>
      </n-layout-header>
      <n-layout-content class="content">
        <router-view />
      </n-layout-content>
    </n-layout>
  </n-layout>
</template>
