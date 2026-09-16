<script setup lang="ts">
import { ref } from "vue";
import { NSwitch } from "naive-ui";

const isDark = ref(localStorage.getItem("generic-mock-theme") === "dark");

function updateTheme(value: boolean) {
  isDark.value = value;
  localStorage.setItem("generic-mock-theme", value ? "dark" : "light");
  window.dispatchEvent(new Event("generic-mock-theme-change"));
}
</script>

<template>
  <n-switch :value="isDark" @update:value="updateTheme">
    <template #checked>Dark</template>
    <template #unchecked>Light</template>
  </n-switch>
</template>
