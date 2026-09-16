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
const route = useRoute();
const router = useRouter();
const items = [
  { label: "持卡人", key: "/paynda/cardholders" },
  { label: "卡片", key: "/paynda/cards" },
  { label: "交易处理", key: "/paynda/transactions" },
  { label: "授权模拟", key: "/paynda/authorization" },
];
</script>
<template>
  <n-layout class="app-shell" has-sider>
    <n-layout-sider class="sidebar" :width="208">
      <strong>Paynda Mock</strong>
      <n-menu
        :value="route.path"
        :options="items"
        @update:value="(path) => router.push(String(path))"
      />
    </n-layout-sider>
    <n-layout>
      <n-layout-header class="app-header">
        <span>Generic Mock</span>
        <ChannelSwitcher current="paynda" />
      </n-layout-header>
      <n-layout-content class="content">
        <router-view />
      </n-layout-content>
    </n-layout>
  </n-layout>
</template>
