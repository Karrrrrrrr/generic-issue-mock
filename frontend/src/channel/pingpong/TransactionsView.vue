<script setup lang="ts">
import SharedView from "@/channel/shared/TransactionsView.vue";
import { requestId } from "./api";
import { accountApi, authorizationApi, transactionApi } from "./management";
</script>

<template>
  <SharedView
    :account-api="accountApi"
    :api="transactionApi"
    :authorization-api="authorizationApi"
    :new-request-id="requestId"
  />
</template>
