<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import {
  darkTheme,
  NConfigProvider,
  NDialogProvider,
  NMessageProvider,
  type GlobalThemeOverrides,
} from "naive-ui";

const initialTheme = localStorage.getItem("generic-mock-theme") === "dark";
const isDark = ref(initialTheme);

function syncTheme() {
  const nextTheme = localStorage.getItem("generic-mock-theme") === "dark";
  isDark.value = nextTheme;
}

onMounted(() => {
  window.addEventListener("generic-mock-theme-change", syncTheme);
});
onUnmounted(() => {
  window.removeEventListener("generic-mock-theme-change", syncTheme);
});

const themeOverrides = computed<GlobalThemeOverrides>(() => ({
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
  DataTable: {
    borderColor: isDark.value ? "#3a3a40" : "#e5e7eb",
  },
  Tag: {
    colorBordered: "transparent",
    colorBorderedPrimary: "transparent",
    colorBorderedInfo: "transparent",
    colorBorderedSuccess: "transparent",
    colorBorderedWarning: "transparent",
    colorBorderedError: "transparent",
  },
  Menu: {
    color: isDark.value ? "#242428" : "#ffffff",
    groupTextColor: isDark.value ? "#b4bbc8" : "#475569",
    itemHeight: "44px",
    itemTextColor: isDark.value ? "#e5e7eb" : "#1f2937",
    itemTextColorHover: "#12B89A",
    itemTextColorActive: "#12B89A",
    itemColorHover: isDark.value ? "#2c3334" : "#eef8f6",
    itemColorActive: isDark.value ? "#203c37" : "#e3f6f3",
    itemColorActiveHover: isDark.value ? "#284c44" : "#d4f0e9",
    itemIconColor: isDark.value ? "#e5e7eb" : "#1f2937",
    itemIconColorHover: "#12B89A",
    itemIconColorActive: "#12B89A",
  },
}));
</script>

<template>
  <n-config-provider :theme="isDark ? darkTheme : undefined" :theme-overrides="themeOverrides">
    <n-message-provider>
      <n-dialog-provider>
        <div class="theme-root" :class="{ 'theme-root-dark': isDark }">
          <router-view/>
        </div>
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>
