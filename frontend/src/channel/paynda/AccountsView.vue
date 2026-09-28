<script setup lang="ts">
import SharedView from "@/channel/shared/AccountsView.vue";
import { accountApi, fundsApi } from "./api";
</script>

<template>
  <SharedView
    :account-api="accountApi"
    :adjust-balance="fundsApi.adjustAccount"
    title="资金账户"
    show-wallet-i-d
  />
</template>
