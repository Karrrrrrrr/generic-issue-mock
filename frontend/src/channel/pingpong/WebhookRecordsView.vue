<script setup lang="ts">
import SharedView from "@/channel/shared/WebhookRecordsView.vue";
import { accountApi, webhookRecordApi } from "./management";
</script>

<template>
  <SharedView :account-api="accountApi" :webhook-record-api="webhookRecordApi" />
</template>
