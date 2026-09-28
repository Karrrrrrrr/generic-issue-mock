<script setup lang="ts">
import SharedView from "@/channel/shared/AuthorizationsView.vue";
import { accountApi, authorizationApi } from "./api";
</script>

<template>
  <SharedView
    :account-api="accountApi"
    :api="authorizationApi"
  />
</template>
