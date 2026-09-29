<script setup lang="ts">
import { NAlert } from "naive-ui";
import SharedView from "@/channel/shared/CardsView.vue";
import { accountApi, cardApi, fundCard, simulateAuthorization, simulateRefund } from "./management";
import { requestId } from "./api";
</script>

<template>
  <SharedView
    :api="cardApi"
    :account-api="accountApi"
    :fund-card="fundCard"
    :simulate-authorization="simulateAuthorization"
    :simulate-refund="simulateRefund"
    show-reserved
    show-virtual-account
    :new-request-id="requestId"
    funding-source-label="所属虚拟账户"
  >
    <template #description>
      <n-alert type="info" :show-icon="false">
        卡拥有独立钱包。充值从所属虚拟账户扣款，转出返回同一虚拟账户；开卡通过 OpenAPI 完成。
      </n-alert>
    </template>
  </SharedView>
</template>
