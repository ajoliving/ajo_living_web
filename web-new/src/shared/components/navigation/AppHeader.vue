<!--
 * 全域頂部導航。
 * 1. 還原 HTML 設計稿的 52px 高度導航欄。
 * 2. 提供 11 個主模組入口與通知鈴鐺。
 * 3. 整合主題切換按鈕（月亮/太陽）與登入入口。
 * 4. 提供 mobile 全屏導航與底部主入口。
-->
<script setup lang="ts">
import { computed, watch } from 'vue';
import { RouterLink, useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';

import { usePreferenceStore } from '@/stores/preferences';
import { useSessionStore } from '@/stores/session';
import { useFeedbackStore } from '@/stores/feedback';
import { type AppThemeName } from '@/utils/theme';

interface NavigationItem {
  key: string;
  to: string;
  label: string;
  match: string[];
  isBell?: boolean;
}

const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const preferenceStore = usePreferenceStore();
const sessionStore = useSessionStore();
const feedbackStore = useFeedbackStore();

const mobileDrawerOpen = computed({
  get: () => preferenceStore.mobile_menu_open,
  set: (value: boolean) => {
    preferenceStore.setMobileMenuOpen(value);
  },
});
const accountPath = computed(() => (sessionStore.isAuthenticated ? '/account/profile' : '/login'));
const notificationUnreadCount = computed(() => 3);
const isDarkTheme = computed(() => preferenceStore.theme === 'dark-neutral');

// 1. 建立主導航入口，對齊 HTML 設計稿 11 個導航項
const primaryNavigationItems = computed<NavigationItem[]>(() => {
  const items: NavigationItem[] = [
    { key: 'home', to: '/', label: t('nav.home'), match: ['/'] },
    { key: 'properties', to: '/properties', label: t('nav.properties'), match: ['/properties'] },
    { key: 'servicedResidences', to: '/serviced-residences', label: t('nav.servicedResidences'), match: ['/serviced-residences'] },
    { key: 'furniture', to: '/furniture', label: t('nav.furniture'), match: ['/furniture'] },
    { key: 'offers', to: '/supermarket-offers', label: t('nav.combinedOffers'), match: ['/supermarket-offers'] },
    { key: 'payments', to: '/payments', label: t('nav.ajoPay'), match: ['/payments'] },
    { key: 'building', to: '/building', label: t('nav.building'), match: ['/building'] },
    { key: 'profile', to: accountPath.value, label: t('nav.memberCenter'), match: ['/account', '/login'] },
    { key: 'management', to: '/account/marketplace/management', label: t('nav.marketplaceManagement'), match: ['/account/marketplace/management'] },
    { key: 'trend', to: '/trend', label: t('nav.trend'), match: ['/trend'] },
    { key: 'notifications', to: '/notifications', label: t('nav.notifications'), match: ['/notifications'], isBell: true },
  ];

  return items;
});

const mobileNavigationItems = computed<NavigationItem[]>(() => [
  ...primaryNavigationItems.value.filter((item) => !item.isBell),
]);

// 2. 判斷目前路由是否命中導航項
const isRouteActive = (item: NavigationItem): boolean => {
  if (item.key === 'profile') {
    return (
      route.path === '/login' ||
      (route.path.startsWith('/account') &&
        !route.path.startsWith('/account/marketplace/settings') &&
        !route.path.startsWith('/account/marketplace/management'))
    );
  }

  return item.match.some((path) => (path === '/' ? route.path === '/' : route.path.startsWith(path)));
};

// 3. 切換主題：亮色 html-fidelity ↔ 深色 dark-neutral
const handleThemeToggle = (): void => {
  const nextTheme: AppThemeName = isDarkTheme.value ? 'html-fidelity' : 'dark-neutral';
  preferenceStore.setTheme(nextTheme);
  feedbackStore.pushToast(
    nextTheme === 'dark-neutral' ? t('common.action.darkModeOn') : t('common.action.lightModeOn'),
    'info',
  );
};

// 4. 執行帳戶入口操作
const handleAccountAction = async (): Promise<void> => {
  mobileDrawerOpen.value = false;
  await router.push(accountPath.value);
};

// 5. 執行登出
const handleSignOut = async (): Promise<void> => {
  mobileDrawerOpen.value = false;
  await sessionStore.signOut();
  await router.push('/login');
};

// 6. 路由切換時收起 mobile 導航
watch(
  () => route.fullPath,
  () => {
    mobileDrawerOpen.value = false;
  },
);
</script>

<template>
  <header class="ajo-nav">
    <RouterLink
      to="/"
      class="ajo-nav__logo"
    >
      AJO LIVING
    </RouterLink>

    <nav class="ajo-nav__links">
      <RouterLink
        v-for="item in primaryNavigationItems"
        :key="item.key"
        :to="item.to"
        class="ajo-nav__link"
        :class="{ 'ajo-nav__link--active': isRouteActive(item), 'ajo-nav__link--bell': item.isBell }"
      >
        <template v-if="item.isBell">
          <span
            class="ajo-nav__bell"
            :aria-label="item.label"
            :title="item.label"
          >
            <svg
              viewBox="0 0 24 24"
              aria-hidden="true"
            >
              <path d="M18 8a6 6 0 0 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9"></path>
              <path d="M10 21h4"></path>
            </svg>
            <i v-if="notificationUnreadCount > 0" />
          </span>
        </template>
        <template v-else>
          {{ item.label }}
        </template>
      </RouterLink>
    </nav>

    <div class="ajo-nav__right">
      <button
        type="button"
        class="ajo-nav__login"
        @click="handleAccountAction"
      >
        {{ sessionStore.isAuthenticated ? t('nav.memberCenter') : t('common.action.signIn') }}
      </button>

      <button
        type="button"
        class="ajo-nav__theme-toggle"
        :title="isDarkTheme ? t('common.action.switchToLight') : t('common.action.switchToDark')"
        :aria-label="isDarkTheme ? t('common.action.switchToLight') : t('common.action.switchToDark')"
        :aria-pressed="isDarkTheme"
        @click="handleThemeToggle"
      >
        <span
          class="ajo-nav__theme-icon ajo-nav__theme-moon"
          aria-hidden="true"
        >
          <svg viewBox="0 0 24 24">
            <path d="M21 14.2A8.6 8.6 0 0 1 9.8 3a7.4 7.4 0 1 0 11.2 11.2Z"></path>
          </svg>
        </span>
        <span
          class="ajo-nav__theme-icon ajo-nav__theme-sun"
          aria-hidden="true"
        >
          <svg viewBox="0 0 24 24">
            <circle cx="12" cy="12" r="4"></circle>
            <path d="M12 2.5v2.2M12 19.3v2.2M4.6 4.6l1.6 1.6M17.8 17.8l1.6 1.6M2.5 12h2.2M19.3 12h2.2M4.6 19.4l1.6-1.6M17.8 6.2l1.6-1.6"></path>
          </svg>
        </span>
      </button>

      <button
        type="button"
        class="ajo-nav__menu"
        :aria-label="t('nav.account')"
        @click="mobileDrawerOpen = true"
      >
        <span></span>
        <span></span>
        <span></span>
      </button>
    </div>

    <Teleport to="body">
      <transition name="drawer-fade">
        <div
          v-if="mobileDrawerOpen"
          class="ajo-mobile-menu"
        >
          <div class="ajo-mobile-menu__header">
            <RouterLink
              to="/"
              class="ajo-mobile-menu__logo"
            >
              AJO LIVING
            </RouterLink>
            <button
              type="button"
              class="ajo-mobile-menu__close"
              :aria-label="t('common.action.close')"
              @click="mobileDrawerOpen = false"
            >
              <svg
                viewBox="0 0 24 24"
                aria-hidden="true"
              >
                <path d="M6 6l12 12M18 6l-12 12" />
              </svg>
            </button>
          </div>

          <nav class="ajo-mobile-menu__body">
            <RouterLink
              v-for="item in mobileNavigationItems"
              :key="item.key"
              :to="item.to"
              class="ajo-mobile-menu__link"
              :class="{ 'ajo-mobile-menu__link--active': isRouteActive(item) }"
            >
              {{ item.label }}
            </RouterLink>
          </nav>

          <div class="ajo-mobile-menu__footer">
            <button
              type="button"
              class="ajo-mobile-menu__theme"
              @click="handleThemeToggle"
            >
              {{ isDarkTheme ? t('common.action.switchToLight') : t('common.action.switchToDark') }}
            </button>

            <button
              v-if="sessionStore.isAuthenticated"
              type="button"
              class="ajo-mobile-menu__signout"
              @click="handleSignOut"
            >
              {{ t('common.action.signOut') }}
            </button>
          </div>
        </div>
      </transition>
    </Teleport>
  </header>

  <nav class="ajo-bottom-nav">
    <RouterLink
      v-for="item in mobileNavigationItems.slice(0, 5)"
      :key="item.key"
      :to="item.to"
      class="ajo-bottom-nav__item"
      :class="{ 'ajo-bottom-nav__item--active': isRouteActive(item) }"
    >
      {{ item.label }}
    </RouterLink>
  </nav>
</template>

<style scoped>
.ajo-nav {
  position: fixed;
  z-index: 40;
  top: 0;
  right: 0;
  left: 0;
  display: flex;
  height: var(--nav-h);
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  border-bottom: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface));
  padding: 0 16px;
}

.ajo-nav__logo {
  flex: 0 0 auto;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  white-space: nowrap;
}

.ajo-nav__links {
  display: flex;
  min-width: 0;
  flex: 1;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.ajo-nav__link {
  display: inline-flex;
  height: var(--nav-h);
  align-items: center;
  border-bottom: 2px solid transparent;
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  font-weight: 500;
  padding: 0 10px;
  white-space: nowrap;
  transition: color 0.15s ease, border-color 0.15s ease;
}

.ajo-nav__link:hover,
.ajo-nav__link--active {
  border-bottom-color: rgb(var(--color-primary));
  color: rgb(var(--color-primary));
}

.ajo-nav__link--bell {
  padding: 0 6px;
}

.ajo-nav__bell {
  position: relative;
  display: inline-flex;
  width: 28px;
  height: 28px;
  align-items: center;
  justify-content: center;
  color: rgb(var(--color-ink-3));
}

.ajo-nav__bell svg {
  width: 19px;
  height: 19px;
  stroke: currentColor;
  stroke-width: 1.8;
  fill: none;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.ajo-nav__bell i {
  position: absolute;
  top: 5px;
  right: 5px;
  width: 7px;
  height: 7px;
  border: 1px solid rgb(var(--color-surface));
  border-radius: 999px;
  background: #e52b54;
}

.ajo-nav__right {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 8px;
}

.ajo-nav__login {
  height: 30px;
  border: 1px solid rgb(var(--color-primary));
  border-radius: var(--radius-md);
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
  cursor: pointer;
  font: inherit;
  font-size: 12px;
  font-weight: 600;
  padding: 0 14px;
  white-space: nowrap;
  transition: background 0.15s ease;
}

.ajo-nav__login:hover {
  background: rgb(var(--color-brand-dark));
  border-color: rgb(var(--color-brand-dark));
}

.ajo-nav__theme-toggle {
  position: relative;
  display: inline-flex;
  width: 34px;
  height: 34px;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 999px;
  background: transparent;
  color: rgb(var(--color-text));
  cursor: pointer;
  box-shadow: none;
  transition: color 0.18s ease, transform 0.18s ease;
}

.ajo-nav__theme-toggle:hover {
  color: rgb(var(--color-primary));
  transform: translateY(-1px);
}

.ajo-nav__theme-icon {
  display: block;
  width: 21px;
  height: 21px;
}

.ajo-nav__theme-icon svg {
  display: block;
  width: 100%;
  height: 100%;
  stroke: currentColor;
  stroke-width: 2.15;
  stroke-linecap: round;
  stroke-linejoin: round;
  fill: none;
}

.ajo-nav__theme-sun {
  display: none;
}

html[data-theme='dark-neutral'] .ajo-nav__theme-moon {
  display: none;
}

html[data-theme='dark-neutral'] .ajo-nav__theme-sun {
  display: block;
}

.ajo-nav__menu {
  display: none;
  flex-direction: column;
  width: 32px;
  height: 32px;
  align-items: center;
  justify-content: center;
  gap: 4px;
  border: 1px solid rgb(var(--color-border));
  border-radius: var(--radius-md);
  background: rgb(var(--color-surface));
  cursor: pointer;
}

.ajo-nav__menu span {
  display: block;
  width: 16px;
  height: 2px;
  background: rgb(var(--color-text));
}

.ajo-mobile-menu {
  position: fixed;
  z-index: 80;
  inset: 0;
  display: flex;
  flex-direction: column;
  background: rgb(var(--color-surface));
}

.ajo-mobile-menu__header {
  display: flex;
  height: var(--nav-h);
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid rgb(var(--color-border));
  padding: 0 16px;
}

.ajo-mobile-menu__logo {
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.ajo-mobile-menu__close {
  display: inline-flex;
  width: 32px;
  height: 32px;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: var(--radius-md);
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  cursor: pointer;
}

.ajo-mobile-menu__close svg {
  width: 18px;
  height: 18px;
  stroke: currentColor;
  stroke-width: 2;
  fill: none;
  stroke-linecap: round;
}

.ajo-mobile-menu__body {
  flex: 1;
  overflow-y: auto;
}

.ajo-mobile-menu__link {
  display: block;
  border-bottom: 1px solid rgb(var(--color-border));
  color: rgb(var(--color-text));
  font-size: 15px;
  padding: 16px 20px;
}

.ajo-mobile-menu__link--active {
  color: rgb(var(--color-primary));
  font-weight: 600;
}

.ajo-mobile-menu__footer {
  display: grid;
  gap: 12px;
  border-top: 1px solid rgb(var(--color-border));
  padding: 16px 20px;
}

.ajo-mobile-menu__theme {
  height: 40px;
  border: 1px solid rgb(var(--color-border));
  border-radius: var(--radius-md);
  background: rgb(var(--color-surface-muted));
  color: rgb(var(--color-text));
  cursor: pointer;
  font: inherit;
  font-weight: 600;
}

.ajo-mobile-menu__signout {
  height: 40px;
  border: 1px solid rgb(var(--color-border));
  border-radius: var(--radius-md);
  background: rgb(var(--color-surface-muted));
  color: rgb(var(--color-text));
  font-weight: 600;
  cursor: pointer;
}

.ajo-bottom-nav {
  display: none;
}

.drawer-fade-enter-active,
.drawer-fade-leave-active {
  transition: opacity 0.18s ease;
}

.drawer-fade-enter-from,
.drawer-fade-leave-to {
  opacity: 0;
}

@media (max-width: 1023px) {
  .ajo-nav__links,
  .ajo-nav__login,
  .ajo-nav__theme-toggle {
    display: none;
  }

  .ajo-nav__menu {
    display: flex;
  }

  .ajo-bottom-nav {
    position: fixed;
    z-index: 35;
    right: 0;
    bottom: 0;
    left: 0;
    display: flex;
    height: 54px;
    align-items: center;
    justify-content: space-around;
    border-top: 1px solid rgb(var(--color-border));
    background: rgb(var(--color-surface));
  }

  .ajo-bottom-nav__item {
    display: inline-flex;
    min-width: 0;
    flex: 1;
    align-items: center;
    justify-content: center;
    color: rgb(var(--color-text-muted));
    font-size: 10px;
    font-weight: 500;
    padding: 0 4px;
    text-align: center;
  }

  .ajo-bottom-nav__item--active {
    color: rgb(var(--color-primary));
  }
}
</style>
