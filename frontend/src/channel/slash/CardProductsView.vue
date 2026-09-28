<script setup lang="ts">
import SharedView from "@/channel/shared/CardProductsView.vue";
import { managementApi } from "./api";
</script>

<template>
  <SharedView
    :load-products="managementApi.cardProducts"
  />
</template>
