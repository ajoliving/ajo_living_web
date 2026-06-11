<!--
 * 支付中心頁。
 * 1. 提供與會員中心一致的左側支付導航。
 * 2. 在右側承載單元、賬單、購物車、訂單、會計與歷史子頁。
-->
<script setup lang="ts">
import { computed } from 'vue';
import { RouterLink, RouterView, useRoute } from 'vue-router';

import { useSessionStore } from '@/stores/session';
import AppIcon from '@/shared/components/base/AppIcon.vue';

interface PaymentNavItem {
  key: string;
  label: string;
  to: string;
  icon: 'wallet' | 'inbox' | 'building' | 'clock' | 'home';
  staffOnly?: boolean;
}

const route = useRoute();
const sessionStore = useSessionStore();

const canViewAccounting = computed(() =>
  Boolean(sessionStore.currentUser.is_staff || sessionStore.me?.ismart_msg?.is_staff),
);

const navItems = computed<PaymentNavItem[]>(() => [
  {
    key: 'units',
    label: '單元',
    to: '/payments/units',
    icon: 'home',
  },
  {
    key: 'bills',
    label: '賬單',
    to: '/payments/bills',
    icon: 'wallet',
  },
  {
    key: 'cart',
    label: '購物車',
    to: '/payments/cart',
    icon: 'inbox',
  },
  {
    key: 'orders',
    label: '訂單',
    to: '/payments/orders',
    icon: 'inbox',
  },
  ...(canViewAccounting.value
    ? [{
        key: 'accounting',
        label: '會計',
        to: '/payments/accounting',
        icon: 'building' as const,
        staffOnly: true,
      }]
    : []),
  {
    key: 'history',
    label: '歷史',
    to: '/payments/history',
    icon: 'clock',
  },
]);

// 1. 判斷支付導航是否啟用
const isNavActive = (item: PaymentNavItem): boolean =>
  route.path === item.to || route.path.startsWith(`${item.to}/`);
</script>

<template>
  <main class="payments-hub">
    <aside class="payments-hub__sidebar">
      <div>
        <p class="payments-hub__kicker">
          AJO Pay
        </p>
        <h1>AJO Pay</h1>
      </div>

      <nav class="payments-hub__nav">
        <RouterLink
          v-for="item in navItems"
          :key="item.key"
          :to="item.to"
          class="payments-hub__nav-item"
          :class="isNavActive(item) ? 'payments-hub__nav-item--active' : ''"
        >
          <AppIcon
            :name="item.icon"
            :size="17"
          />
          <span>
            <strong>{{ item.label }}</strong>
          </span>
        </RouterLink>
      </nav>
    </aside>

    <section class="payments-hub__content">
      <RouterView v-slot="{ Component, route: activeRoute }">
        <Transition
          name="subroute-slide"
          mode="out-in"
        >
          <component
            :is="Component"
            :key="activeRoute.fullPath"
            class="subroute-transition-shell"
          />
        </Transition>
      </RouterView>
    </section>
  </main>
</template>

<style scoped>
.payments-hub {
  display: grid;
  width: 100%;
  max-width: var(--layout-page-max-width);
  gap: 1.25rem;
  margin: 0 auto;
  padding: 1rem var(--layout-page-padding-inline) 5rem;
  color: rgb(var(--color-text));
}

.payments-hub__sidebar {
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

.payments-hub__kicker {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.payments-hub__sidebar h1 {
  margin: 0.75rem 0 0;
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 1.75rem;
  font-weight: 500;
  line-height: 1.3;
}

.payments-hub__nav {
  display: grid;
  gap: 0.5rem;
}

.payments-hub__nav-item {
  position: relative;
  display: flex;
  min-height: 2.75rem;
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
  padding: 0.7rem 0.85rem;
  color: rgb(var(--color-text) / 0.78);
  text-decoration: none;
  box-shadow:
    0 12px 28px rgb(15 23 42 / 0.07),
    inset 0 1px 0 rgb(255 255 255 / 0.08);
  backdrop-filter: blur(18px) saturate(160%);
  -webkit-backdrop-filter: blur(18px) saturate(160%);
  transition:
    border-color 0.2s ease,
    background 0.2s ease,
    color 0.2s ease,
    padding-left 0.2s ease;
}

.payments-hub__nav-item:hover {
  border-color: rgb(var(--color-border) / 0.38);
  color: rgb(var(--color-text));
  padding-left: 1rem;
}

.payments-hub__nav-item--active,
.payments-hub__nav-item--active:hover {
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

.payments-hub__nav-item::after {
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

.payments-hub__nav-item--active::after {
  transform: scaleX(1);
  opacity: 1;
}

.payments-hub__nav-item span {
  display: grid;
  min-width: 0;
}

.payments-hub__nav-item strong {
  font-size: 0.95rem;
  font-weight: 750;
  line-height: 1.2;
}

.payments-hub__content {
  min-width: 0;
}

@media (min-width: 1024px) {
  .payments-hub {
    grid-template-columns: 15.75rem minmax(0, 1fr);
    align-items: start;
  }

  .payments-hub__sidebar {
    position: sticky;
    top: calc(var(--app-header-offset, 0rem) + 2rem);
  }
}

@media (max-width: 767px) {
  .payments-hub {
    padding: 1rem var(--layout-page-padding-inline) 4rem;
  }
}
</style>
