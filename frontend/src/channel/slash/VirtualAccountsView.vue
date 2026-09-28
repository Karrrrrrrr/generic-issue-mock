<script setup lang="ts">
import SharedView from "@/channel/shared/VirtualAccountsView.vue";
import { accountApi, virtualAccountApi } from "./api";
</script>

<template>
  <SharedView
    :account-api="accountApi"
    :api="virtualAccountApi"
  />
</template>
