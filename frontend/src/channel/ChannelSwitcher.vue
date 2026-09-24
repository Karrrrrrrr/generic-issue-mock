<script setup lang="ts">
import { NButton, NDropdown } from "naive-ui";
import { useRouter } from "vue-router";

defineProps<{ current: "slash" | "photonpay" | "paynda" | "pingpong" }>();

const router = useRouter();
const options = [
  { label: "Slash", key: "slash" },
  { label: "PhotonPay", key: "photonpay" },
  { label: "Paynda", key: "paynda" },
  { label: "PingPong", key: "pingpong" },
];

function select(channel: string) {
  void router.push(channel === "pingpong" ? "/pingpong/accounts" : `/${channel}/cardholders`);
}
</script>

<template>
  <n-dropdown :options="options" @select="select">
    <n-button size="small">{{ current }}</n-button>
  </n-dropdown>
</template>
