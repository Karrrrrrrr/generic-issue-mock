<script setup lang="ts">
import CardTransactionSimulator from "@/channel/shared/CardTransactionSimulator.vue";
import { api, refundApi } from "./api";

const emit = defineEmits<{ completed: [] }>();

async function loadCards() {
  return (await api.listCards()).data;
}
</script>

<template>
  <CardTransactionSimulator
    :load-cards="loadCards"
    :authorize="api.simulateAuthorization"
    :refund="refundApi.simulate"
    @completed="emit('completed')"
  />
</template>
