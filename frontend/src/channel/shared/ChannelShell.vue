<script setup lang="ts">
import { NLayout, NLayoutContent, NLayoutHeader, NLayoutSider, NMenu } from "naive-ui";
import { useRoute, useRouter } from "vue-router";
import ChannelSwitcher from "@/channel/ChannelSwitcher.vue";
import ThemeToggle from "@/channel/ThemeToggle.vue";

const route = useRoute();
const router = useRouter();
defineProps<{
  channel: "slash" | "paynda" | "photonpay" | "pingpong";
  title: string;
  items: import("naive-ui").MenuOption[];
}>();
</script>

<template>
  <n-layout class="app-shell" has-sider native-scrollbar>
    <n-layout-sider class="sidebar" :width="208" bordered native-scrollbar>
      <div class="channel-logo">{{ title }}</div>
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
          <ThemeToggle />
          <ChannelSwitcher :current="channel" />
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
