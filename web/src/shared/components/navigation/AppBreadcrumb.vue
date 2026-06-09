<!--
 * 全域麵包屑導航。
 * 1. 呈現詳情頁與深層頁面的返回層級。
 * 2. 支援最後一層純文字與前置層級路由連結。
-->
<script setup lang="ts">
import { RouterLink } from 'vue-router';
import type { RouteLocationRaw } from 'vue-router';

interface BreadcrumbItem {
  label: string;
  to?: RouteLocationRaw;
}

defineProps<{
  items: BreadcrumbItem[];
}>();
</script>

<template>
  <nav
    v-if="items.length > 0"
    class="app-breadcrumb"
    aria-label="Breadcrumb"
  >
    <ol>
      <li
        v-for="(item, index) in items"
        :key="`${item.label}-${index}`"
      >
        <RouterLink
          v-if="item.to && index < items.length - 1"
          :to="item.to"
        >
          {{ item.label }}
        </RouterLink>
        <span v-else>{{ item.label }}</span>
      </li>
    </ol>
  </nav>
</template>

<style scoped>
.app-breadcrumb {
  min-width: 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.82rem;
  font-weight: 800;
}

.app-breadcrumb ol {
  display: flex;
  min-width: 0;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.45rem;
  margin: 0;
  padding: 0;
  list-style: none;
}

.app-breadcrumb li {
  display: inline-flex;
  min-width: 0;
  align-items: center;
  gap: 0.45rem;
}

.app-breadcrumb li + li::before {
  color: rgb(var(--color-border-strong));
  content: "/";
}

.app-breadcrumb a,
.app-breadcrumb span {
  overflow: hidden;
  max-width: min(34rem, 72vw);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.app-breadcrumb a {
  color: rgb(var(--color-text-muted));
  text-decoration: none;
}

.app-breadcrumb a:hover {
  color: rgb(var(--color-primary));
}

.app-breadcrumb li:last-child span {
  color: rgb(var(--color-text));
}
</style>
