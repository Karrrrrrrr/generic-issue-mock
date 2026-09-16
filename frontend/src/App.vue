<script setup lang="ts">
import { defineAsyncComponent } from 'vue'
import { channel } from '@/config'

const consoleByChannel = {
  slash: defineAsyncComponent(() => import('@/channel/slash/Console.vue')),
  photonpay: defineAsyncComponent(() => import('@/channel/photonpay/Console.vue')),
  paynda: defineAsyncComponent(() => import('@/channel/paynda/Console.vue')),
}
</script>

<template>
  <component :is="consoleByChannel[channel]" />
</template>
