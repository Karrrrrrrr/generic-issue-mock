<script setup lang="ts">
import PayndaConsole from '@/channel/paynda/Console.vue'
import PhotonPayConsole from '@/channel/photonpay/Console.vue'
import SlashConsole from '@/channel/slash/Console.vue'
import { channel } from '@/config'

const consoleByChannel = {
  slash: SlashConsole,
  photonpay: PhotonPayConsole,
  paynda: PayndaConsole,
}
</script>

<template>
  <component :is="consoleByChannel[channel]" />
</template>
