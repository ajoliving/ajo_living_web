<!--
 * 二手交易模組路由中轉頁。
 * 1. 承載二手交易子路由頁面與全局 breadcrumb。
 * 2. 保持 marketplace 路由入口穩定。
-->
<script setup lang="ts">
import { computed } from 'vue';
import { RouterLink, RouterView, useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';

import AppIcon from '@/shared/components/base/AppIcon.vue';

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
    { label: t('nav.marketplace'), to: '/marketplace/discover' },
  ];

  if (path.startsWith('/marketplace/discover')) {
    return [...items, { label: t('nav.discover') }];
  }

  if (path.startsWith('/marketplace/filter')) {
    return [...items, { label: t('nav.filter') }];
  }

  if (
    path.startsWith('/marketplace/my') ||
    path.startsWith('/marketplace/my-listings') ||
    path.startsWith('/marketplace/publish') ||
    path.startsWith('/marketplace/chat')
  ) {
    const myItems = [...items, { label: t('nav.my'), to: '/marketplace/my' }];
    const myListingItems = [...myItems, { label: t('nav.myListings'), to: '/marketplace/my/listings' }];

    if (path.includes('/new') || path.startsWith('/marketplace/publish')) {
      return [...myItems, { label: t('marketplace.editor.createTitle') }];
    }

    if (path.includes('/editor')) {
      return [...myListingItems, { label: t('marketplace.editor.editTitle') }];
    }

    if (path.includes('/listing/') || path.includes('/preview/')) {
      return [...myListingItems, { label: t('marketplace.mine.detailTitle') }];
    }

    if (path.includes('/orders')) {
      return [...myItems, { label: t('marketplace.myHub.orders') }];
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
    <nav class="marketplace-breadcrumb">
      <template
        v-for="(item, index) in breadcrumbItems"
        :key="`${item.label}-${index}`"
      >
        <RouterLink
          v-if="item.to && index < breadcrumbItems.length - 1"
          :to="item.to"
          class="marketplace-breadcrumb__link"
        >
          {{ item.label }}
        </RouterLink>
        <span
          v-else
          class="marketplace-breadcrumb__current"
        >
          {{ item.label }}
        </span>

        <AppIcon
          v-if="index < breadcrumbItems.length - 1"
          name="chevron-down"
          class="marketplace-breadcrumb__separator"
          :size="16"
        />
      </template>
    </nav>

    <RouterView />
  </div>
</template>

<style scoped>
.marketplace-breadcrumb {
  display: flex;
  width: 100%;
  max-width: 1280px;
  align-items: center;
  gap: 0.5rem;
  margin: 0 auto;
  padding: 1rem 2rem;
  color: #414848;
  font-size: 0.875rem;
  line-height: 1.5;
}

.marketplace-breadcrumb__link {
  color: #414848;
  font-weight: 500;
  transition: color 0.2s ease;
}

.marketplace-breadcrumb__link:hover {
  color: #002727;
}

.marketplace-breadcrumb__current {
  color: #002727;
  font-weight: 700;
}

.marketplace-breadcrumb__separator {
  transform: rotate(-90deg);
  color: #c1c8c7;
}

@media (max-width: 767px) {
  .marketplace-breadcrumb {
    padding: 1rem 1.25rem;
  }
}
</style>
