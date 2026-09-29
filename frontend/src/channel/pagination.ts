import { computed, ref, watch, type Ref } from "vue";
import type { PaginationProps } from "naive-ui";

export const pageSizes = [10, 20, 50, 100, 200, 500, 1000];

export function useRemotePagination(load: () => Promise<void>) {
  const page = ref(1);
  const pageSize = ref(20);
  const total = ref(0);
  const pagination = computed<PaginationProps>(() => ({
    page: page.value,
    pageSize: pageSize.value,
    itemCount: total.value,
    pageSizes,
    showSizePicker: true,
    pageSlot: 5,
    prefix: () => `共 ${total.value} 条`,
    onUpdatePage: (value) => {
      page.value = value;
      void load();
    },
    onUpdatePageSize: (value) => {
      pageSize.value = value;
      page.value = 1;
      void load();
    },
  }));

  return {
    page,
    pageSize,
    total,
    pagination,
  };
}

export function useClientPagination(rows: Ref<unknown[]>) {
  const page = ref(1);
  const pageSize = ref(20);
  const pagination = computed(() => ({
    page: page.value,
    pageSize: pageSize.value,
    pageSizes,
    showSizePicker: true,
    pageSlot: 5,
    prefix: () => `共 ${rows.value.length} 条`,
    onUpdatePage: (value) => {
      page.value = value;
    },
    onUpdatePageSize: (value) => {
      pageSize.value = value;
      page.value = 1;
    },
  } satisfies PaginationProps));

  watch(() => rows.value, () => {
    const lastPage = Math.max(1, Math.ceil(rows.value.length / pageSize.value));
    page.value = Math.min(page.value, lastPage);
  });

  return pagination;
}
