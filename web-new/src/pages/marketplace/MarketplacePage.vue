<!--
 * 二手交易模組路由中轉頁。
 * 1. 承載二手交易子路由頁面與全局 breadcrumb。
 * 2. 保持 marketplace 路由入口穩定。
-->
<script setup lang="ts">
import { computed } from 'vue';
import { RouterLink, RouterView, useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';

interface MarketplaceBreadcrumbItem {
  label: string;
  to?: string;
}

const route = useRoute();
const { t } = useI18n();

// 1. 依目前路由輸出二手交易 breadcrumb
const breadcrumbItems = computed<MarketplaceBreadcrumbItem[]>(() => {
  const path = route.path;
  const items: MarketplaceBreadcrumbItem[] = [
    { label: t('nav.marketplace'), to: '/marketplace/filter' },
  ];

  if (path.startsWith('/marketplace/filter')) {
    return [...items, { label: t('nav.filter') }];
  }

  if (path.startsWith('/marketplace/settings') || path.startsWith('/account/marketplace/settings')) {
    return [...items, { label: t('nav.settings') }];
  }

  if (
    path.startsWith('/marketplace/my') ||
    path.startsWith('/marketplace/my-listings') ||
    path.startsWith('/marketplace/publish') ||
    path.startsWith('/marketplace/chat') ||
    path.startsWith('/account/marketplace/my')
  ) {
    const myItems = [...items, { label: t('nav.my'), to: '/account/listings' }];
    const myListingItems = [...myItems, { label: t('nav.myListings'), to: '/account/listings' }];

    if (path.includes('/new') || path.startsWith('/marketplace/publish')) {
      return [...myItems, { label: t('marketplace.editor.createTitle') }];
    }

    if (path.includes('/editor')) {
      return [...myListingItems, { label: t('marketplace.editor.editTitle') }];
    }

    if (path.includes('/listing/') || path.includes('/preview/')) {
      return [...myListingItems, { label: t('marketplace.mine.detailTitle') }];
    }

    if (path.includes('/favorites')) {
      return [...myItems, { label: t('marketplace.myHub.favorites') }];
    }

    if (path.includes('/profile')) {
      return [...myItems, { label: t('marketplace.myHub.profile') }];
    }

    if (path.includes('/chat')) {
      return [...myItems, { label: t('nav.chat') }];
    }

    return [...myItems, { label: t('nav.myListings') }];
  }

  if (path.startsWith('/marketplace/listing')) {
    const source = String(route.query.from ?? '');
    if (source === 'filter') {
      return [...items, { label: t('nav.filter'), to: '/marketplace/filter' }, { label: t('marketplace.detail.title') }];
    }
    if (source === 'furniture') {
      return [{ label: t('nav.furniture'), to: '/furniture' }, { label: t('marketplace.detail.title') }];
    }

    return [...items, { label: t('marketplace.detail.title') }];
  }

  if (path.startsWith('/marketplace/seller')) {
    return [...items, { label: t('marketplace.seller.title') }];
  }

  return items;
});
</script>

<template>
  <div class="marketplace-layout min-h-screen pb-10">
    <nav class="breadcrumb">
      <template
        v-for="(item, index) in breadcrumbItems"
        :key="`${item.label}-${index}`"
      >
        <span v-if="index > 0" class="bc-sep">›</span>
        <span
          v-if="index === breadcrumbItems.length - 1"
          class="bc-current"
        >
          {{ item.label }}
        </span>
        <RouterLink
          v-else-if="item.to"
          :to="item.to"
          class="bc-link"
        >
          {{ item.label }}
        </RouterLink>
      </template>
    </nav>

    <RouterView />
  </div>
</template>

<style scoped>
.breadcrumb {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  width: 100%;
  margin: 0 0 16px;
  padding: 12px var(--layout-page-padding-inline);
  border: 1px solid var(--bdr);
  background: var(--sur);
  color: var(--ink-3);
  font-family: var(--font);
  font-size: 12px;
  line-height: 1.5;
}

.bc-link {
  border: 0;
  background: transparent;
  color: var(--ink-3);
  cursor: pointer;
  font-family: var(--font);
  font-size: 12px;
  padding: 0;
  text-decoration: none;
}

.bc-link:hover {
  color: var(--brand);
}

.bc-sep {
  color: var(--ink-4);
}

.bc-current {
  color: var(--ink);
  font-weight: 600;
}
</style>
