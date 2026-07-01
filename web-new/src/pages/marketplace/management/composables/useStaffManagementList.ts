/*
 * Staff 管理列表 - 共用資料流程。
 * 1. 管理四類 Staff 分頁列表的查詢條件與分頁狀態。
 * 2. 封裝狀態操作後的即時刷新流程。
 */
import axios from 'axios';
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import type { PaginatedResult } from '@/model/api';
import { useFeedbackStore } from '@/stores/feedback';

export interface StaffManagementListParams {
  page?: number;
  page_size?: number;
  keyword?: string;
  category_code?: string;
  status?: string;
}

export type StaffManagementListFetcher<T> = (
  params: StaffManagementListParams,
) => Promise<{ data: { data: PaginatedResult<T> } }>;

interface UseStaffManagementListOptions<T> {
  fetcher: StaffManagementListFetcher<T>;
  pageSize?: number;
  loadErrorKey: string;
  updateErrorKey?: string;
  updateSuccessKey?: string;
}

// 1. 建立 Staff 分頁列表流程
export const useStaffManagementList = <T>(options: UseStaffManagementListOptions<T>) => {
  const { t } = useI18n();
  const feedbackStore = useFeedbackStore();
  const loading = ref(false);
  const keyword = ref('');
  const categoryCode = ref('');
  const status = ref('');
  const page = ref(1);
  const pageSize = options.pageSize ?? 20;
  const total = ref(0);
  const items = ref<T[]>([]);
  const hasPrevious = computed(() => page.value > 1);
  const hasNext = computed(() => page.value * pageSize < total.value);

  const loadItems = async (targetPage = page.value): Promise<void> => {
    loading.value = true;

    try {
      const { data } = await options.fetcher({
        page: targetPage,
        page_size: pageSize,
        keyword: keyword.value.trim() || undefined,
        category_code: categoryCode.value || undefined,
        status: status.value || undefined,
      });
      items.value = data.data.items;
      page.value = data.data.pagination.page;
      total.value = data.data.pagination.total;
    } catch (error) {
      feedbackStore.pushToast(readManagementError(error, t(options.loadErrorKey)), 'error');
    } finally {
      loading.value = false;
    }
  };

  const search = async (): Promise<void> => {
    await loadItems(1);
  };

  const previous = async (): Promise<void> => {
    if (!hasPrevious.value) {
      return;
    }
    await loadItems(page.value - 1);
  };

  const next = async (): Promise<void> => {
    if (!hasNext.value) {
      return;
    }
    await loadItems(page.value + 1);
  };

  const runAction = async (action: () => Promise<unknown>): Promise<boolean> => {
    try {
      await action();
      feedbackStore.pushToast(t(options.updateSuccessKey ?? 'marketplace.management.statusUpdated'), 'success');
      await loadItems(page.value);
      return true;
    } catch (error) {
      feedbackStore.pushToast(readManagementError(error, t(options.updateErrorKey ?? 'marketplace.management.statusUpdateError')), 'error');
      return false;
    }
  };

  watch([categoryCode, status], () => {
    void loadItems(1);
  });

  onMounted(() => {
    void loadItems(1);
  });

  return {
    categoryCode,
    hasNext,
    hasPrevious,
    items,
    keyword,
    loadItems,
    loading,
    next,
    page,
    pageSize,
    previous,
    runAction,
    search,
    status,
    t,
    total,
  };
};

// 2. 解析 API 錯誤訊息
const readManagementError = (error: unknown, fallback: string): string =>
  axios.isAxiosError<{ message?: string }>(error)
    ? error.response?.data?.message ?? fallback
    : fallback;
