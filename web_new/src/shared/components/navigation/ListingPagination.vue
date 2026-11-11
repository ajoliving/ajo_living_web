<!--
 * 公開列表分頁。
 * 1. 顯示首尾頁、目前頁附近頁碼及省略號。
 * 2. 提供上一頁及下一頁操作。
-->
<script setup lang="ts">
import { computed } from 'vue';
import AppIcon from '@/shared/components/base/AppIcon.vue';

interface Props {
  currentPage: number;
  totalPages: number;
  previousLabel: string;
  nextLabel: string;
  navigationLabel: string;
}

const props = defineProps<Props>();
const emit = defineEmits<{ (event: 'select', page: number): void }>();

// 1. 保留首三頁、末兩頁及目前頁前後，缺口以省略號表示。
const pages = computed<Array<number | string>>(() => {
  const total = Math.max(1, props.totalPages);
  const values = new Set([1, 2, 3, total - 1, total, props.currentPage - 1, props.currentPage, props.currentPage + 1]);
  const numbers = [...values].filter((page) => page >= 1 && page <= total).sort((a, b) => a - b);
  const result: Array<number | string> = [];
  numbers.forEach((page, index) => {
    const previous = numbers[index - 1];
    if (previous !== undefined && page - previous === 2) result.push(previous + 1);
    else if (previous !== undefined && page - previous > 2) result.push(`gap-${previous}`);
    result.push(page);
  });
  return result;
});
</script>

<template>
  <nav class="listing-pages" :aria-label="navigationLabel">
    <button
      type="button"
      class="listing-pages-button"
      :aria-label="previousLabel"
      :title="previousLabel"
      :disabled="currentPage <= 1"
      @click="emit('select', currentPage - 1)"
    ><AppIcon name="arrow-left" :size="18" /></button>
    <template v-for="item in pages" :key="item">
      <span v-if="typeof item === 'string'" class="listing-pages-gap" aria-hidden="true">...</span>
      <button
        v-else
        type="button"
        class="listing-pages-button"
        :class="{ 'is-active': item === currentPage }"
        :aria-current="item === currentPage ? 'page' : undefined"
        @click="emit('select', item)"
      >{{ item }}</button>
    </template>
    <button
      type="button"
      class="listing-pages-button"
      :aria-label="nextLabel"
      :title="nextLabel"
      :disabled="currentPage >= totalPages"
      @click="emit('select', currentPage + 1)"
    ><AppIcon name="arrow-right" :size="18" /></button>
  </nav>
</template>

<style scoped>
.listing-pages {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.listing-pages-button {
  display: inline-flex;
  min-width: 40px;
  height: 40px;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  padding: 0 8px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  font: inherit;
  cursor: pointer;
}

.listing-pages-button.is-active {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: #fff;
}

.listing-pages-button:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.listing-pages-gap {
  padding: 0 2px;
  color: rgb(var(--color-text));
}
</style>
