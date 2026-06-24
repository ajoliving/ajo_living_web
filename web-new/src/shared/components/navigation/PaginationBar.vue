<!--
 * 分頁欄元件。
 * 1. 左側顯示分頁資訊，右側顯示分頁按鈕群組。
 * 2. 支援上一頁/下一頁與數字頁碼，點擊觸發 select 事件。
 * 3. 響應式：移動端改為垂直排列。
-->
<script setup lang="ts">
interface PaginationPage {
  label: string | number;
  active?: boolean;
  key: string | number;
}

interface PaginationBarProps {
  info?: string;
  pages: PaginationPage[];
  ariaLabel?: string;
}

const props = defineProps<PaginationBarProps>();

const emit = defineEmits<{
  (event: 'select', page: PaginationPage): void;
}>();

// 1. 點擊分頁按鈕時觸發 select 事件
const handleSelect = (page: PaginationPage) => {
  emit('select', page);
};
</script>

<template>
  <div
    class="pagination-bar"
    :aria-label="props.ariaLabel"
  >
    <span
      v-if="props.info"
      class="pagination-bar-info"
    >
      {{ props.info }}
    </span>
    <div class="pagination-bar-actions">
      <button
        v-for="page in props.pages"
        :key="page.key"
        type="button"
        class="pagination-bar-btn"
        :class="page.active ? 'pagination-bar-btn-active' : ''"
        @click="handleSelect(page)"
      >
        {{ page.label }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.pagination-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-top: 18px;
  padding: 12px 14px;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background: #fff;
}

.pagination-bar-info {
  font-size: 11px;
  color: var(--color-ink-3);
}

.pagination-bar-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.pagination-bar-btn {
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: #fff;
  color: var(--color-ink-3);
  cursor: pointer;
  font-family: inherit;
  font-size: 11px;
  padding: 7px 11px;
}

.pagination-bar-btn-active {
  border-color: var(--color-text);
  color: var(--color-text);
  font-weight: 500;
}

@media (max-width: 767px) {
  .pagination-bar {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
