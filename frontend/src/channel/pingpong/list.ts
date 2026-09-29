import {
  onMounted,
  reactive,
  ref,
  shallowRef,
} from "vue";
import { useMessage } from "naive-ui";
import { pageSizes } from "@/channel/pagination";
import { api, type Page } from "./api";
export function useList<Item>(resource: string) {
  const message = useMessage();
  const rows = shallowRef<Item[]>([]);
  const loading = ref(false);
  const busy = ref(false);
  const filters = reactive<Record<string, string | number | undefined>>({});
  let version = 0;
  const pagination = reactive({
    page: 1,
    pageSize: 20,
    itemCount: 0,
    showSizePicker: true,
    pageSizes,
    onUpdatePage(page: number) {
      pagination.page = page;
      void load();
    },
    onUpdatePageSize(size: number) {
      pagination.pageSize = size;
      search();
    },
  });
  async function load() {
    const current = ++version;
    loading.value = true;
    try {
      const result = await api.get<Page<Item>>(resource, {
        ...filters,
        page_number: pagination.page,
        page_size: pagination.pageSize,
      });
      if (current === version) {
        rows.value = result.items;
        pagination.itemCount = result.total;
      }
    }
    catch (error) {
      message.error(error instanceof Error ? error.message : "加载失败");
    }
    finally {
      if (current === version) {
        loading.value = false;
      }
    }
  }
  function search() {
    pagination.page = 1;
    void load();
  }
  async function perform(action: () => Promise<unknown>) {
    busy.value = true;
    try {
      await action();
      message.success("操作完成");
      await load();
      return true;
    }
    catch (error) {
      message.error(error instanceof Error ? error.message : "操作失败");
      return false;
    }
    finally {
      busy.value = false;
    }
  }
  onMounted(() => void load());
  return {
    rows,
    loading,
    busy,
    filters,
    pagination,
    load,
    search,
    perform
  };
}
