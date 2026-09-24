<script setup lang="ts">
import { NButton, NDatePicker, NInput, NSelect } from "naive-ui";
import type { SelectOption } from "naive-ui";

export interface FilterField {
  key: string;
  label: string;
  options?: SelectOption[];
}

defineProps<{
  fields: FilterField[];
  accountOptions: SelectOption[];
  accountsLoading: boolean;
  loading: boolean;
}>();

const values = defineModel<Record<string, string | null>>("values", { required: true });
const dateRange = defineModel<[number, number] | null>("dateRange", { required: true });
const emit = defineEmits<{
  search: [];
  reset: [];
  loadAccounts: [];
}>();

function showAccountOptions(shown: boolean) {
  if (shown) {
    emit("loadAccounts");
  }
}
</script>

<template>
  <form class="table-filters" @submit.prevent="emit('search')">
    <div class="filter-grid">
      <n-select
        :value="values.account_id ?? null"
        :options="accountOptions"
        :loading="accountsLoading"
        clearable
        filterable
        placeholder="所属账户（名称 / ID）"
        aria-label="所属账户"
        @update:value="(value) => (values.account_id = value)"
        @update:show="showAccountOptions"
      />
      <template v-for="field in fields" :key="field.key">
        <n-select
          v-if="field.options"
          :value="values[field.key] ?? null"
          :options="field.options"
          :placeholder="field.label"
          :aria-label="field.label"
          clearable
          @update:value="(value) => (values[field.key] = value)"
        />
        <n-input
          v-else
          :value="values[field.key] ?? null"
          :placeholder="field.label"
          :aria-label="field.label"
          clearable
          @update:value="(value) => (values[field.key] = value)"
        />
      </template>
      <n-date-picker
        v-model:value="dateRange"
        class="filter-date-range"
        type="datetimerange"
        format="yyyy-MM-dd HH:mm:ss"
        :default-time="['00:00:00', '23:59:59']"
        start-placeholder="开始时间"
        end-placeholder="结束时间"
        clearable
      />
    </div>
    <div class="filter-actions">
      <n-button attr-type="submit" type="primary" :loading="loading">查询</n-button>
      <n-button :disabled="loading" @click="emit('reset')">重置</n-button>
    </div>
  </form>
</template>
