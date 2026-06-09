<!--
 * 會員中心頁。
 * 1. 提供會員中心左側內置導航。
 * 2. 承載 AJO 錢包、個人資料、聊天、樓盤、服務式住宅、二手傢俬與收藏子頁。
-->
<script setup lang="ts">
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';

import AccountProfilePage from '@/pages/member/profile/info/Page.vue';
import AccountWalletPage from '@/pages/member/profile/wallet/Page.vue';
import FurnitureChatPage from '@/pages/communications/messages/Page.vue';
import MarketplaceMyFavoritesPage from '@/pages/member/favorites/Page.vue';
import MarketplaceMyListingsPage from '@/pages/member/listings/Page.vue';
import MarketplaceListingEditorPage from '@/pages/member/listings/editor/Page.vue';
import MarketplaceMyListingPreviewPage from '@/pages/member/listings/preview/Page.vue';
import PropertyMyPage from '@/pages/member/property/Page.vue';
import AppIcon from '@/shared/components/base/AppIcon.vue';

type MemberTab = 'wallet' | 'profile' | 'chat' | 'properties' | 'serviced-residences' | 'listings' | 'listing-editor' | 'listing-preview' | 'favorites';

interface MyNavItem {
  key: string;
  label: string;
  tab: MemberTab;
  icon: 'browse' | 'building' | 'home' | 'star' | 'user' | 'message' | 'wallet';
}

const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const activeTab = computed<MemberTab>(() => resolveMemberTab(route.query.tab));
const activeComponent = computed(() => {
  switch (activeTab.value) {
    case 'profile':
      return AccountProfilePage;
    case 'chat':
      return FurnitureChatPage;
    case 'properties':
      return PropertyMyPage;
    case 'serviced-residences':
      return PropertyMyPage;
    case 'listings':
      return MarketplaceMyListingsPage;
    case 'listing-editor':
      return MarketplaceListingEditorPage;
    case 'listing-preview':
      return MarketplaceMyListingPreviewPage;
    case 'favorites':
      return MarketplaceMyFavoritesPage;
    default:
      return AccountWalletPage;
  }
});
const activeComponentProps = computed(() => {
  if (activeTab.value === 'properties') {
    return { channel: 'sale' };
  }
  if (activeTab.value === 'serviced-residences') {
    return { channel: 'serviced' };
  }
  return {};
});

const navItems = computed<MyNavItem[]>(() => [
  {
    key: 'wallet',
    label: t('marketplace.myHub.wallet'),
    tab: 'wallet',
    icon: 'wallet',
  },
  {
    key: 'profile',
    label: t('marketplace.myHub.profile'),
    tab: 'profile',
    icon: 'user',
  },
  {
    key: 'chat',
    label: t('marketplace.myHub.chat'),
    tab: 'chat',
    icon: 'message',
  },
  {
    key: 'properties',
    label: t('property.sale.myTitle'),
    tab: 'properties',
    icon: 'home',
  },
  {
    key: 'serviced-residences',
    label: t('property.serviced.myTitle'),
    tab: 'serviced-residences',
    icon: 'building',
  },
  {
    key: 'listings',
    label: t('marketplace.myHub.publishedListings'),
    tab: 'listings',
    icon: 'browse',
  },
  {
    key: 'favorites',
    label: t('marketplace.myHub.favorites'),
    tab: 'favorites',
    icon: 'star',
  },
]);

// 1. 讀取目前會員中心 tab
const resolveMemberTab = (value: unknown): MemberTab => {
  const tab = typeof value === 'string' ? value : '';
  if ([
    'wallet',
    'profile',
    'chat',
    'properties',
    'serviced-residences',
    'listings',
    'listing-editor',
    'listing-preview',
    'favorites',
  ].includes(tab)) {
    return tab as MemberTab;
  }
  return 'wallet';
};

// 2. 切換會員中心 tab
const setActiveTab = async (tab: MemberTab): Promise<void> => {
  await router.push({ path: '/member', query: { tab } });
};

// 3. 判斷左側導航是否啟用
const isNavActive = (item: MyNavItem): boolean => {
  if (item.tab === 'listings') {
    return ['listings', 'listing-editor', 'listing-preview'].includes(activeTab.value);
  }
  return activeTab.value === item.tab;
};
</script>

<template>
  <main class="marketplace-my-hub">
    <aside class="marketplace-my-hub__sidebar">
      <div>
        <p class="marketplace-my-hub__kicker">
          {{ t('marketplace.myHub.title') }}
        </p>
        <h1>{{ t('marketplace.myHub.title') }}</h1>
        <p class="marketplace-my-hub__description">
          {{ t('marketplace.myHub.subtitle') }}
        </p>
      </div>

      <nav class="marketplace-my-hub__nav">
        <button
          v-for="item in navItems"
          :key="item.key"
          type="button"
          class="marketplace-my-hub__nav-item"
          :class="isNavActive(item) ? 'marketplace-my-hub__nav-item--active' : ''"
          @click="setActiveTab(item.tab)"
        >
          <AppIcon
            :name="item.icon"
            :size="17"
          />
          <span>{{ item.label }}</span>
        </button>
      </nav>
    </aside>

    <section class="marketplace-my-hub__content">
      <Transition
        name="subroute-slide"
        mode="out-in"
      >
        <component
          :is="activeComponent"
          v-bind="activeComponentProps"
          :key="activeTab"
          class="subroute-transition-shell"
        />
      </Transition>
    </section>
  </main>
</template>

<style scoped>
.marketplace-my-hub {
  display: grid;
  width: 100%;
  max-width: var(--layout-page-max-width);
  gap: 1.25rem;
  margin: 0 auto;
  padding: 1rem var(--layout-page-padding-inline) 5rem;
  color: rgb(var(--color-text));
}

.marketplace-my-hub__sidebar {
  --subroute-nav-active-color: color-mix(in srgb, rgb(var(--color-primary)) 78%, rgb(var(--color-text)) 22%);
  --subroute-nav-active-shadow: 0 0 10px rgb(var(--color-primary) / 0.16);
  --subroute-nav-underline: color-mix(in srgb, rgb(var(--color-primary)) 88%, rgb(var(--color-text)) 12%);
  display: grid;
  align-content: start;
  gap: 1.5rem;
  border: 1px solid rgb(var(--color-border) / 0.3);
  border-radius: 8px;
  background:
    linear-gradient(
      180deg,
      rgb(var(--color-topbar-surface) / 0.76),
      rgb(var(--color-toolbar-surface) / 0.64)
    );
  box-shadow:
    0 16px 40px rgb(15 23 42 / 0.1),
    inset 0 -1px 0 rgb(255 255 255 / 0.06);
  padding: 1rem;
  backdrop-filter: blur(28px) saturate(184%);
  -webkit-backdrop-filter: blur(28px) saturate(184%);
}

.marketplace-my-hub__kicker {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.marketplace-my-hub__sidebar h1 {
  margin: 0.75rem 0 0;
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 1.75rem;
  font-weight: 500;
  line-height: 1.3;
}

.marketplace-my-hub__description {
  margin: 0.65rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.875rem;
  line-height: 1.7;
}

.marketplace-my-hub__nav {
  display: grid;
  gap: 0.5rem;
}

.marketplace-my-hub__nav-item {
  position: relative;
  display: inline-flex;
  min-height: 2.75rem;
  width: 100%;
  align-items: center;
  gap: 0.65rem;
  border: 1px solid rgb(var(--color-border) / 0.24);
  border-radius: 8px;
  background:
    linear-gradient(
      180deg,
      rgb(var(--color-topbar-surface) / 0.62),
      rgb(var(--color-toolbar-surface) / 0.48)
    );
  padding: 0.65rem 0.85rem;
  color: rgb(var(--color-text) / 0.78);
  font-size: 0.9375rem;
  font-weight: 700;
  line-height: 1;
  box-shadow:
    0 12px 28px rgb(15 23 42 / 0.07),
    inset 0 1px 0 rgb(255 255 255 / 0.08);
  backdrop-filter: blur(18px) saturate(160%);
  -webkit-backdrop-filter: blur(18px) saturate(160%);
  transition:
    border-color 0.2s ease,
    background 0.2s ease,
    color 0.2s ease;
}

.marketplace-my-hub__nav-item:hover {
  border-color: rgb(var(--color-border) / 0.38);
  color: rgb(var(--color-text));
}

.marketplace-my-hub__nav-item--active,
.marketplace-my-hub__nav-item--active:hover {
  border-color: rgb(var(--color-primary) / 0.28);
  background:
    linear-gradient(
      180deg,
      rgb(var(--color-topbar-surface) / 0.86),
      rgb(var(--color-primary-soft) / 0.24)
    );
  color: var(--subroute-nav-active-color);
  box-shadow:
    0 14px 32px rgb(var(--color-primary) / 0.1),
    inset 0 1px 0 rgb(255 255 255 / 0.1);
  text-shadow: var(--subroute-nav-active-shadow);
}

.marketplace-my-hub__nav-item::after {
  position: absolute;
  left: 0.85rem;
  right: 0.85rem;
  bottom: 0.42rem;
  height: 1.5px;
  border-radius: 999px;
  background: var(--subroute-nav-underline);
  transform: scaleX(0);
  transform-origin: left center;
  opacity: 0;
  transition:
    transform 0.22s ease,
    opacity 0.22s ease;
  content: '';
}

.marketplace-my-hub__nav-item--active::after {
  transform: scaleX(1);
  opacity: 1;
}

.marketplace-my-hub__content {
  min-width: 0;
}

@media (min-width: 1024px) {
  .marketplace-my-hub {
    grid-template-columns: 15.75rem minmax(0, 1fr);
    align-items: start;
  }

  .marketplace-my-hub__sidebar {
    position: sticky;
    top: calc(var(--app-header-offset, 0rem) + 2rem);
  }
}

@media (max-width: 767px) {
  .marketplace-my-hub {
    padding: 1rem var(--layout-page-padding-inline) 4rem;
  }
}
</style>
