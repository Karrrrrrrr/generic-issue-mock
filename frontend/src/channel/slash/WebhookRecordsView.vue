<script setup lang="ts">
import SharedView from "@/channel/shared/WebhookRecordsView.vue";
import { accountApi, webhookRecordApi, webhookApi } from "./api";
import { onMounted, ref } from "vue";
import { useMessage } from "naive-ui";
const message = useMessage();
const events = ref<{ label: string; value: string }[]>([]);
onMounted(async () => {
  try {
    events.value = (await webhookApi.events()).map((value) => ({ label: value, value }));
  } catch (error) {
    message.error(error instanceof Error ? error.message : "加载事件失败");
  }
});
</script>

<template>
  <SharedView
    :account-api="accountApi"
    :webhook-record-api="webhookRecordApi"
    extended-filters
    :event-options="events"
  />
</template>
