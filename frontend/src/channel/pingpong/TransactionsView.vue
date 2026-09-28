<script setup lang="ts">
import SharedView from "@/channel/shared/TransactionsView.vue";
import { accountApi, transactionApi } from "./management";
</script>

<template>
  <SharedView :account-api="accountApi" :api="transactionApi" />
</template>
