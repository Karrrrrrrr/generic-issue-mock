<script setup lang="ts">
import SharedView from "@/channel/shared/WebhooksView.vue";
import { accountApi, webhookApi } from "./api";
</script>

<template>
	<SharedView
		:account-api="accountApi"
		:webhook-api="webhookApi"
		event-label="类型"
	/>
</template>
