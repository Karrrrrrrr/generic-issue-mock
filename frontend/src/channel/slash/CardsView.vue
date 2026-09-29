<script setup lang="ts">
import SharedView from "@/channel/shared/CardsView.vue";
import { api, accountApi, fundCard, refundApi } from "./api";
</script>

<template>
  <SharedView
    :api="api"
    :account-api="accountApi"
    :fund-card="fundCard"
    :simulate-authorization="api.simulateAuthorization"
    :simulate-refund="refundApi.simulate"
  />
</template>
