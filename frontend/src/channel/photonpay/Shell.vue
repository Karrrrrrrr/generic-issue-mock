<script setup lang="ts">
import { NLayout, NLayoutContent, NLayoutHeader, NLayoutSider, NMenu, } from "naive-ui";
import { useRoute, useRouter } from "vue-router";
import ChannelSwitcher from "@/channel/ChannelSwitcher.vue";
import ThemeToggle from "@/channel/ThemeToggle.vue";

const route = useRoute();
const router = useRouter();
const items = [
  {
    type: "group", label: "卡", key: "cards", children: [
      { label: "持卡人", key: "/photonpay/cardholders" },
      { label: "卡片管理", key: "/photonpay/cards" },
    ]
  },
  {
    type: "group", label: "交易", key: "transactions", children: [
      { label: "授权管理", key: "/photonpay/authorizations" },
      { label: "卡交易", key: "/photonpay/transactions" },
      { label: "模拟交易", key: "/photonpay/authorization" },
    ]
  },
  {
    type: "group", label: "配置", key: "configuration", children: [
      { label: "账户", key: "/photonpay/accounts" },
      { label: "Webhook 管理", key: "/photonpay/webhooks" },
    ]
  },
];
</script>
<template>
  <n-layout class="app-shell" has-sider native-scrollbar>
    <n-layout-sider class="sidebar" :width="208" bordered native-scrollbar>
      <div class="channel-logo">PhotonPay Mock</div>
      <n-menu
          :value="route.path"
          :options="items"
          @update:value="(path) => router.push(String(path))"
      />
    </n-layout-sider>
    <n-layout class="main-layout" native-scrollbar>
      <n-layout-header class="app-header">
        <span></span>
        <div class="header-actions">
          <ThemeToggle/>
          <ChannelSwitcher current="photonpay"/>
        </div>
      </n-layout-header>
      <n-layout-content class="content" native-scrollbar>
        <router-view v-slot="{ Component, route: currentRoute }">
          <transition name="slide" mode="out-in">
            <div :key="currentRoute.path" class="route-page">
              <component :is="Component"/>
            </div>
          </transition>
        </router-view>
      </n-layout-content>
    </n-layout>
  </n-layout>
</template>
