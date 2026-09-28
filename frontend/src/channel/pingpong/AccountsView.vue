<script setup lang="ts">
import { NAlert } from "naive-ui";
import SharedView from "@/channel/shared/AccountsView.vue";
import { accountApi, adjustAccountBalance } from "./management";
</script>

<template>
  <SharedView :account-api="accountApi" :adjust-balance="adjustAccountBalance" title="PingPong 账户">
    <template #description>
      <n-alert type="info" :show-icon="false">
        OpenAPI 使用 PINGPONG_APP_ACCOUNTS 显式绑定应用与账户 ID；创建账户不创建虚拟账户或复制产品。
      </n-alert>
    </template>
  </SharedView>
</template>
