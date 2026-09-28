<script setup lang="ts">
import SharedView from "@/channel/shared/CardholdersView.vue";
import { api } from "./api";
</script>

<template>
  <SharedView
    :api="api"
    channel-name="paynda"
  />
</template>
