<script setup lang="ts">
import WebhookRecordsView from "@/channel/photonpay/WebhookRecordsView.vue";
import { webhookRecordApi } from "./api";
</script>

<template>
  <webhook-records-view :api="webhookRecordApi"/>
</template>
