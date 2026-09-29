<script setup lang="ts">
import { NAlert } from "naive-ui";
import SharedView from "@/channel/shared/VirtualAccountsView.vue";
import { accountApi, virtualAccountApi } from "./management";
import { requestId } from "./api";
</script>

<template>
  <SharedView
    :account-api="accountApi"
    :api="virtualAccountApi"
    :currency-options="['USD']"
    :new-request-id="requestId"
  >
    <template #description>
      <n-alert type="info" :show-icon="false">
        虚拟账户是卡的资金来源，不是卡的消费钱包。当前支持 USD，只提供从账户充值。
      </n-alert>
    </template>
  </SharedView>
</template>
