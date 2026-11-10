<!--
 * 管理列表分頁列。
 * 1. 顯示目前分頁範圍。
 * 2. 提供上一頁與下一頁操作。
-->
<script setup lang="ts">
const props = defineProps<{
  hasNext: boolean;
  hasPrevious: boolean;
  loading: boolean;
  page: number;
  pageSize: number;
  total: number;
  t: (key: string, params?: Record<string, unknown>) => string;
}>();

const emit = defineEmits<{
  next: [];
  previous: [];
}>();

// 1. 格式化目前列表範圍
const formatRange = (): string => {
  if (props.total <= 0) {
    return props.t('marketplace.management.paginationEmpty');
  }

  const start = (props.page - 1) * props.pageSize + 1;
  const end = Math.min(props.page * props.pageSize, props.total);
  return props.t('marketplace.management.paginationRange', {
    start,
    end,
    total: props.total,
  });
};
</script>

<template>
  <div class="management-pagination">
    <p>{{ formatRange() }}</p>
    <div>
      <button
        type="button"
        :disabled="!hasPrevious || loading"
        @click="emit('previous')"
      >
        {{ t('marketplace.management.previous') }}
      </button>
      <button
        type="button"
        :disabled="!hasNext || loading"
        @click="emit('next')"
      >
        {{ t('marketplace.management.next') }}
      </button>
    </div>
  </div>
</template>

