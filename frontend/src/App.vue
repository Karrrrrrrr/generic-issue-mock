<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { darkTheme, NConfigProvider, NDialogProvider, NMessageProvider } from "naive-ui";

const isDark = ref(false);

function syncTheme() {
  isDark.value = localStorage.getItem("generic-mock-theme") === "dark";
  document.documentElement.dataset.theme = isDark.value ? "dark" : "light";
}

syncTheme();

onMounted(() => {
  window.addEventListener("generic-mock-theme-change", syncTheme);
});
onUnmounted(() => window.removeEventListener("generic-mock-theme-change", syncTheme));

const themeOverrides = computed(() => ({
  common: {
    primaryColor: "#12B89A",
    primaryColorHover: "#0FA385",
    primaryColorPressed: "#0D8D71",
    borderRadius: "10px",
    borderRadiusSmall: "8px",
    fontSizeMedium: "14px",
    heightMedium: "38px",
  },
  Button: { borderRadiusMedium: "10px", borderRadiusSmall: "8px" },
  Card: { borderRadius: "14px", paddingMedium: "20px" },
  Input: { borderRadius: "10px", heightMedium: "38px" },
  Select: { borderRadius: "10px", heightMedium: "38px" },
  Menu: {
    color: isDark.value ? "#242428" : "#ffffff",
    groupTextColor: isDark.value ? "#a6adbb" : "#8a94a6",
    itemHeightMedium: "44px",
    itemTextColor: isDark.value ? "#e0e0e0" : "#666980",
    itemTextColorHover: "#12B89A",
    itemTextColorActive: "#12B89A",
    itemColorHover: isDark.value ? "#2c3334" : "#edf8f6",
    itemColorActive: isDark.value ? "#203c37" : "#e3f6f3",
    itemColorActiveHover: isDark.value ? "#203c37" : "#e3f6f3",
    itemIconColor: isDark.value ? "#e0e0e0" : "#666980",
    itemIconColorHover: "#12B89A",
    itemIconColorActive: "#12B89A",
  },
}));
</script>

<template>
  <n-config-provider :theme="isDark ? darkTheme : undefined" :theme-overrides="themeOverrides">
    <n-message-provider>
      <n-dialog-provider><router-view /></n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>
