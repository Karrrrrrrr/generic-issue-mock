<script setup lang="ts">
import SharedView from "@/channel/shared/WebhooksView.vue";
import { accountApi, webhookApi } from "./management";
</script>

<template>
  <SharedView :account-api="accountApi" :webhook-api="webhookApi" />
</template>
