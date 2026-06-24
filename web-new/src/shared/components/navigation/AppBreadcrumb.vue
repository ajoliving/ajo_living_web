<!--
 * 麵包屑導航元件。
 * 1. 接收 items 陣列，渲染層級路徑。
 * 2. 最後一項為當前頁（不可點擊），其餘項可點擊跳轉。
 * 3. 還原 HTML 設計稿 .breadcrumb 樣式。
-->
<script setup lang="ts">
import { RouterLink } from 'vue-router';

interface BreadcrumbItem {
  label: string;
  to?: string;
}

interface AppBreadcrumbProps {
  items: BreadcrumbItem[];
}

defineProps<AppBreadcrumbProps>();
</script>

<template>
  <nav class="breadcrumb">
    <template v-for="(item, index) in items" :key="index">
      <span v-if="index > 0" class="bc-sep">/</span>
      <span v-if="index === items.length - 1" class="bc-current">{{ item.label }}</span>
      <RouterLink v-else-if="item.to" :to="item.to" class="bc-link">{{ item.label }}</RouterLink>
      <button v-else class="bc-link" type="button">{{ item.label }}</button>
    </template>
  </nav>
</template>

<style scoped>
.breadcrumb {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  color: var(--color-ink-3);
  font-size: 12px;
  line-height: 1.5;
}

.bc-link {
  border: 0;
  background: transparent;
  color: var(--color-ink-3);
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  padding: 0;
}

.bc-link:hover {
  color: var(--color-primary);
}

.bc-sep {
  color: var(--color-ink-4);
}

.bc-current {
  color: var(--color-text);
  font-weight: 600;
}
</style>
