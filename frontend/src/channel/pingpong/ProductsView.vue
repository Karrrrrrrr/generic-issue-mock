<script setup lang="ts">
import { NAlert } from "naive-ui";
import SharedView from "@/channel/shared/CardProductsView.vue";
import { loadProducts } from "./management";
</script>

<template>
  <SharedView :load-products="loadProducts" prefix-label="候选 BIN">
    <template #description>
      <n-alert type="info" :show-icon="false">
        产品在渠道内共享；开卡随机选一个候选 BIN，共享发卡序号。当前为 USD、24 个月有效期的虚拟账户关联独立卡。
      </n-alert>
    </template>
  </SharedView>
</template>
