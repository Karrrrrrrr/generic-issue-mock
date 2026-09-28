<script setup lang="ts">
import SharedView from "@/channel/shared/TransactionsView.vue";
import { api, accountApi } from "./api";
</script>

<template>
  <SharedView
    :api="api"
    :account-api="accountApi"
  />
</template>
