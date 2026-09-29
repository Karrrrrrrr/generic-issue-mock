<script setup lang="ts">
import SharedView from "@/channel/shared/TransactionsView.vue";
import { api, accountApi, authorizationApi } from "./api";
</script>

<template>
  <SharedView
    :api="api"
    :authorization-api="authorizationApi"
    :account-api="accountApi"
  />
</template>
