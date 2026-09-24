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
  <n-layout
    class="app-shell"
    has-sider
    native-scrollbar
  >
    <n-layout-sider
      class="sidebar"
      :width="208"
      bordered
      native-scrollbar
    >
      <div class="channel-logo">
        PingPong Mock
      </div>
      <n-menu
        :value="route.path"
        :options="items"
        @update:value="(path) => router.push(String(path))"
      />
    </n-layout-sider>
    <n-layout class="main-layout" native-scrollbar>
      <n-layout-header class="app-header">
        <span>
        </span>
        <div class="header-actions">
          <ThemeToggle />
          <ChannelSwitcher current="pingpong" />
        </div>
      </n-layout-header>
      <n-layout-content class="content" native-scrollbar>
        <router-view v-slot="{ Component, route: currentRoute }">
          <transition name="slide" mode="out-in">
            <div :key="currentRoute.path" class="route-page">
              <component :is="Component" />
            </div>
          </transition>
        </router-view>
      </n-layout-content>
    </n-layout>
  </n-layout>
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
