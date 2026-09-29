<script setup lang="ts">
import SharedView from "@/channel/shared/WebhookRecordsView.vue";
import { accountApi, webhookRecordApi } from "./api";
</script>

<template>
	<SharedView
		:account-api="accountApi"
		:webhook-record-api="webhookRecordApi"
		event-label="类型"
	/>
</template>
