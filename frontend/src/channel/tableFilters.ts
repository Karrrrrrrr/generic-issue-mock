import { ref } from "vue";
import dayjs from "dayjs";
import { useMessage, type SelectOption } from "naive-ui";

type TableFilterOptions = {
  onSearch: () => void | Promise<void>;
  loadAccounts: () => Promise<{ id: string; name: string }[]>;
};

export function useTableFilters(options: TableFilterOptions) {
  const message = useMessage();
  const filters = ref<Record<string, string | null>>({});
  const dateRange = ref<[number, number] | null>(null);
  const appliedFilters = ref<Record<string, string>>({});
  const accountOptions = ref<SelectOption[]>([]);
  const accountsLoading = ref(false);

  async function loadAccounts() {
    if (accountsLoading.value || accountOptions.value.length > 0) {
      return;
    }
    accountsLoading.value = true;
    try {
      const accounts = await options.loadAccounts();
      accountOptions.value = accounts.map((account) => ({
        label: `${account.name} (${account.id})`,
        value: account.id,
      }));
    } catch (error) {
      message.error(error instanceof Error ? error.message : "加载账户选项失败");
    } finally {
      accountsLoading.value = false;
    }
  }

  function search() {
    const nextFilters: Record<string, string> = {};
    for (const [key, value] of Object.entries(filters.value)) {
      if (value !== null && value.trim() !== "") {
        nextFilters[key] = value.trim();
      }
    }
    if (dateRange.value !== null) {
      const [start, end] = dateRange.value;
      if (!dayjs(start).isValid() || !dayjs(end).isValid() || start > end) {
        message.warning("请选择有效的时间范围");
        return;
      }
      nextFilters.created_from = dayjs(start).toISOString();
      nextFilters.created_to = dayjs(end).toISOString();
    }
    appliedFilters.value = nextFilters;
    void options.onSearch();
  }

  function reset() {
    filters.value = {};
    dateRange.value = null;
    appliedFilters.value = {};
    void options.onSearch();
  }

  return {
    filters,
    dateRange,
    appliedFilters,
    accountOptions,
    accountsLoading,
    loadAccounts,
    search,
    reset,
  };
}
