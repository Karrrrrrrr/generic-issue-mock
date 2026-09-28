<script setup lang="ts">
import { NAlert } from "naive-ui";
import SharedView from "@/channel/shared/AuthorizationsView.vue";
import { accountApi, authorizationApi } from "./management";
import { requestID } from "./api";
</script>

<template>
  <SharedView
    :account-api="accountApi"
    :api="authorizationApi"
    :detailed-filters="false"
    show-notification-status
    :new-request-i-d="requestID"
  >
    <template #description>
      <n-alert type="warning" :show-icon="false">
        本地授权已生效，通知未投递：尚缺 PingPong Webhook 事件及报文契约。这里不配置同步审批。
      </n-alert>
    </template>
  </SharedView>
</template>
