<script setup lang="ts">
import { NAlert } from "naive-ui";
import SharedView from "@/channel/shared/AuthorizationsView.vue";
import { accountApi, authorizationApi } from "./management";
import { requestId } from "./api";
</script>

<template>
  <SharedView
    :account-api="accountApi"
    :api="authorizationApi"
    :new-request-id="requestId"
  >
    <template #description>
      <n-alert type="warning" :show-icon="false">
        授权在本地完成后按启用的 Webhook 配置异步通知，不等待下游审批。投递结果请查看投递记录；时间、大小写和 HTTP 200 成功规则待确认，暂不签名。
      </n-alert>
    </template>
  </SharedView>
</template>
