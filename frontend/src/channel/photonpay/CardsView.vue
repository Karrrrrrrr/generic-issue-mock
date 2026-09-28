<script setup lang="ts">
import SharedView from "@/channel/shared/CardsView.vue";
import { api, accountApi, fundCard } from "./api";
</script>

<template>
  <SharedView
    :api="api"
    :account-api="accountApi"
    :fund-card="fundCard"
    :show-expiry="false"
  />
</template>
