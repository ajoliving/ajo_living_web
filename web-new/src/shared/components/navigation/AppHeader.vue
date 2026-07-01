<!--
 * 全域頂部導航。
 * 1. 嚴格對齊 HTML 設計稿 chrome.html 的 .nav 結構與樣式。
 * 2. 對齊設計稿 11 個導航項：首頁 + 9 個主模組入口 + 通知中心鈴鐺。
 * 3. 整合主題切換按鈕（月亮/太陽）與登入入口。
 * 4. 保留 mobile 全屏導航與底部主入口。
-->
<script setup lang="ts">
import { computed, watch } from 'vue';
import { RouterLink, useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';

import AppIcon from '@/shared/components/base/AppIcon.vue';
import { usePreferenceStore } from '@/stores/preferences';
import { useSessionStore } from '@/stores/session';
import { type AppThemeName } from '@/utils/theme';

interface NavigationItem {
  key: string;
  to: string;
  label: string;
  match: string[];
}

const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const preferenceStore = usePreferenceStore();
const sessionStore = useSessionStore();

const mobileDrawerOpen = computed({
  get: () => preferenceStore.mobile_menu_open,
  set: (value: boolean) => {
    preferenceStore.setMobileMenuOpen(value);
  },
});
const accountPath = computed(() => (sessionStore.isAuthenticated ? '/profile' : '/login'));
const isDarkTheme = computed(() => preferenceStore.theme === 'dark-neutral');

// 1. 建立主導航入口，對齊 HTML 設計稿 10 個文字導航項（首頁 + 9 個主模組）
const primaryNavigationItems = computed<NavigationItem[]>(() => [
  { key: 'home', to: '/', label: '首頁', match: ['/'] },
  { key: 'listing', to: '/properties', label: '樓盤租售', match: ['/properties'] },
  { key: 'service', to: '/serviced-residences', label: '服務式住宅', match: ['/serviced-residence', '/serviced-residences'] },
  { key: 'market', to: '/furniture', label: '家具', match: ['/furniture'] },
  { key: 'offers', to: '/supermarket-offers', label: '綜合優惠', match: ['/offers', '/supermarket-offers'] },
  { key: 'payment', to: '/payments', label: 'AJO Pay', match: ['/payments'] },
  { key: 'affairs', to: '/building', label: '我的大廈', match: ['/building'] },
  { key: 'profile', to: '/profile', label: '會員中心', match: ['/profile', '/account', '/login'] },
  { key: 'management', to: '/account/marketplace/management', label: '管理', match: ['/account/marketplace/management'] },
  { key: 'trend', to: '/trend', label: '走勢', match: ['/trend'] },
]);

// 2. mobile 全屏導航項（首頁 + 9 個主模組 + 通知中心）
const mobileNavigationItems = computed<NavigationItem[]>(() => [
  ...primaryNavigationItems.value,
  { key: 'notification', to: '/notifications', label: '通知中心', match: ['/notifications', '/account/notifications'] },
]);

// 3. 底部導航 5 個入口（縮寫標籤對齊設計稿 bottom-nav）
const bottomNavigationItems = computed<NavigationItem[]>(() => [
  { key: 'home', to: '/', label: '首頁', match: ['/'] },
  { key: 'listing', to: '/properties', label: '樓盤', match: ['/properties'] },
  { key: 'market', to: '/furniture', label: '家具', match: ['/furniture'] },
  { key: 'offers', to: '/supermarket-offers', label: '格價', match: ['/offers', '/supermarket-offers'] },
  { key: 'profile', to: '/profile', label: '我的', match: ['/profile', '/account', '/login'] },
]);

// 4. 判斷目前路由是否命中導航項
const isRouteActive = (item: NavigationItem): boolean => {
  if (item.key === 'profile') {
    return (
      route.path === '/login' ||
      route.path === '/profile' ||
      (
        route.path.startsWith('/account') &&
        !route.path.startsWith('/account/marketplace') &&
        route.path !== '/account/notifications'
      )
    );
  }
  if (item.key === 'management') {
    return route.path.startsWith('/account/marketplace/management');
  }
  return item.match.some((path) => (path === '/' ? route.path === '/' : route.path.startsWith(path)));
};

// 5. 通知中心鈴鐺活躍狀態
const isNotificationActive = computed(() =>
  route.path.startsWith('/notifications') || route.path === '/account/notifications',
);

// 6. 切換主題：亮色 html-fidelity ↔ 深色 dark-neutral
const handleThemeToggle = (): void => {
  const nextTheme: AppThemeName = isDarkTheme.value ? 'html-fidelity' : 'dark-neutral';
  preferenceStore.setTheme(nextTheme);
};

// 7. 執行帳戶入口操作
const handleAccountAction = async (): Promise<void> => {
  mobileDrawerOpen.value = false;
  await router.push(accountPath.value);
};

// 8. 執行登出
const handleSignOut = async (): Promise<void> => {
  mobileDrawerOpen.value = false;
  await sessionStore.signOut();
  await router.push('/login');
};

// 9. 路由切換時收起 mobile 導航
watch(
  () => route.fullPath,
  () => {
    mobileDrawerOpen.value = false;
  },
);
</script>

<template>
  <nav
    id="mainNav"
    class="nav"
  >
    <!-- logo -->
    <RouterLink
      to="/"
      class="nav-logo"
    >
      AJO LIVING
    </RouterLink>

    <!-- 主導航連結 -->
    <div class="nav-links">
      <RouterLink
        v-for="item in primaryNavigationItems"
        :key="item.key"
        :to="item.to"
        class="nl"
        :class="{ on: isRouteActive(item) }"
      >
        {{ item.label }}
      </RouterLink>
      <RouterLink
        to="/notifications"
        class="nl nav-bell"
        :class="{ on: isNotificationActive }"
        aria-label="通知中心"
        title="通知中心"
      >
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path d="M18 8a6 6 0 0 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9"></path>
          <path d="M10 21h4"></path>
        </svg>
        <span class="nav-bell-dot"></span>
      </RouterLink>
    </div>

    <!-- 右側區域 -->
    <div class="nav-r">
      <span>照映</span><span>繁中</span>
      <button
        type="button"
        class="nav-login"
        :class="{ 'nav-login--icon': sessionStore.isAuthenticated }"
        :aria-label="sessionStore.isAuthenticated ? t('nav.memberCenter') : '登入'"
        :title="sessionStore.isAuthenticated ? t('nav.memberCenter') : '登入'"
        @click="handleAccountAction"
      >
        <AppIcon
          v-if="sessionStore.isAuthenticated"
          name="user"
          :size="18"
          :stroke-width="2"
        />
        <span v-else>登入</span>
      </button>
    </div>
    <button
      type="button"
      class="theme-toggle"
      :aria-label="t('common.action.switchToDark')"
      @click="handleThemeToggle"
    >
      <span class="theme-toggle__icon theme-toggle__moon">
        <svg viewBox="0 0 24 24"><path d="M21 14.2A8.6 8.6 0 0 1 9.8 3a7.4 7.4 0 1 0 11.2 11.2Z"></path></svg>
      </span>
      <span class="theme-toggle__icon theme-toggle__sun">
        <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="4"></circle><path d="M12 2.5v2.2M12 19.3v2.2M4.6 4.6l1.6 1.6M17.8 17.8l1.6 1.6M2.5 12h2.2M19.3 12h2.2M4.6 19.4l1.6-1.6M17.8 6.2l1.6-1.6"></path></svg>
      </span>
    </button>
    <button
      type="button"
      class="hamburger"
      :aria-label="t('nav.account')"
      @click="mobileDrawerOpen = true"
    >
      <span></span>
      <span></span>
      <span></span>
    </button>
  </nav>

  <div class="scroll-progress"></div>
  <div class="toast-container"></div>

  <Teleport to="body">
    <transition name="drawer-fade">
      <div
        v-if="mobileDrawerOpen"
        class="mobile-menu"
      >
        <div class="mobile-menu-header">
          <RouterLink
            to="/"
            class="mobile-menu-logo"
          >
            AJO LIVING
          </RouterLink>
          <button
            type="button"
            class="mobile-close"
            :aria-label="t('common.action.close')"
            @click="mobileDrawerOpen = false"
          >
            <svg viewBox="0 0 24 24"><path d="M6 6l12 12M18 6l-12 12"/></svg>
          </button>
        </div>

        <nav class="mobile-menu-body">
          <RouterLink
            v-for="item in mobileNavigationItems"
            :key="item.key"
            :to="item.to"
            class="mnl"
            :class="{ on: isRouteActive(item) }"
          >
            {{ item.label }}
          </RouterLink>
        </nav>

        <div class="mobile-menu-footer">
          <button
            type="button"
            class="mobile-menu-theme"
            @click="handleThemeToggle"
          >
            {{ isDarkTheme ? t('common.action.switchToLight') : t('common.action.switchToDark') }}
          </button>
          <button
            v-if="sessionStore.isAuthenticated"
            type="button"
            class="mobile-menu-signout"
            @click="handleSignOut"
          >
            {{ t('common.action.signOut') }}
          </button>
        </div>
      </div>
    </transition>
  </Teleport>

  <nav class="bottom-nav">
    <RouterLink
      v-for="item in bottomNavigationItems"
      :key="item.key"
      :to="item.to"
      class="bnav-item"
      :class="{ on: isRouteActive(item) }"
    >
      <span class="bnav-icon"></span>
      {{ item.label }}
    </RouterLink>
  </nav>
</template>

<style scoped>
/*
 * 樣式嚴格對齊 HTML 設計稿 .nav / .nav-logo / .nav-links / .nl / .nav-bell /
 * .nav-r / .nav-login / .theme-toggle / .hamburger / .mobile-menu /
 * .mobile-menu-header / .mobile-menu-logo / .mobile-close / .mobile-menu-body /
 * .mnl / .mobile-menu-footer / .bottom-nav / .bnav-item / .bnav-icon /
 * .scroll-progress / .toast-container，使用原生 CSS 變數名。
 */

.nav {
  height: var(--nav-h);
  background: var(--sur);
  border-bottom: 1px solid var(--bdr);
  padding: 0 var(--sp-5);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--sp-2);
  position: sticky;
  top: 0;
  z-index: 30;
  box-shadow: var(--shadow-sm);
}

/* logo */
.nav-logo {
  font-family: var(--font-serif);
  font-size: 14px;
  letter-spacing: 2px;
  color: var(--ink);
  border-bottom: 2px solid var(--brand);
  padding-bottom: 1px;
  flex-shrink: 0;
  cursor: pointer;
  background: none;
  border-top: none;
  border-left: none;
  border-right: none;
  text-decoration: none;
}

/* 主導航連結容器 */
.nav-links {
  display: flex;
  flex: 1;
  justify-content: center;
  overflow: hidden;
}

/* 單個導航項 */
.nl {
  font-family: inherit;
  font-size: var(--text-sm);
  color: var(--ink-3);
  padding: 0 10px;
  height: var(--nav-h);
  border-bottom: 2px solid transparent;
  cursor: pointer;
  background: none;
  border: none;
  transition: color 0.15s;
  white-space: nowrap;
  text-decoration: none;
  display: flex;
  align-items: center;
}

.nl:hover {
  color: var(--ink);
}

.nl.on {
  color: var(--brand);
  border-bottom-color: var(--brand);
  font-weight: 500;
}

/* 通知中心鈴鐺 */
.nav-bell {
  position: relative;
  width: 34px;
  padding: 0;
  justify-content: center;
  color: var(--ink-3);
}

.nav-bell svg {
  width: 19px;
  height: 19px;
  stroke: currentColor;
  stroke-width: 1.8;
  fill: none;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.nav-bell-dot {
  position: absolute;
  top: 10px;
  right: 6px;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #e52b54;
}

.nav-bell:hover,
.nav-bell.on {
  color: var(--ink);
}

/* 右側區域 */
.nav-r {
  display: flex;
  align-items: center;
  gap: var(--sp-2);
  font-size: 11px;
  color: var(--ink-3);
  flex-shrink: 0;
}

.nav-r span {
  cursor: pointer;
  white-space: nowrap;
}

.nav-r span:hover {
  color: var(--black);
}

.nav-login {
  background: var(--accent);
  color: var(--white);
  border: none;
  padding: 5px 11px;
  font-size: 11px;
  cursor: pointer;
  font-family: inherit;
  border-radius: 2px;
  white-space: nowrap;
  transition: background 0.15s ease;
}

.nav-login--icon {
  width: 34px;
  height: 34px;
  border-radius: 999px;
  background: transparent;
  padding: 0;
  color: var(--ink);
  transition:
    color 0.18s ease,
    transform 0.18s ease;
}

.nav-login--icon :deep(svg) {
  display: block;
}

.nav-login:hover {
  background: var(--brand-dark);
}

.nav-login--icon:hover {
  background: transparent;
  color: var(--brand);
  transform: translateY(-1px);
}

html[data-theme='dark-neutral'] .nav-login--icon {
  background: transparent;
  color: #E5E7EB;
}

html[data-theme='dark-neutral'] .nav-login--icon:hover {
  background: transparent;
  color: var(--brand);
}

/* 主題切換按鈕 */
.theme-toggle {
  position: relative;
  display: inline-flex;
  width: 34px;
  height: 34px;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 999px;
  background: transparent;
  color: var(--ink);
  cursor: pointer;
  box-shadow: none;
  transition: color 0.18s ease, transform 0.18s ease;
}

.theme-toggle:hover {
  color: var(--brand);
  transform: translateY(-1px);
}

.theme-toggle__icon {
  display: block;
  width: 21px;
  height: 21px;
}

.theme-toggle__icon svg {
  display: block;
  width: 100%;
  height: 100%;
  stroke: currentColor;
  stroke-width: 2.15;
  stroke-linecap: round;
  stroke-linejoin: round;
  fill: none;
}

.theme-toggle__sun {
  display: none;
}

html[data-theme='dark-neutral'] .theme-toggle {
  background: transparent;
  border: 0;
  color: #E5E7EB;
}

html[data-theme='dark-neutral'] .theme-toggle__moon {
  display: none;
}

html[data-theme='dark-neutral'] .theme-toggle__sun {
  display: block;
}

/* mobile 漢堡按鈕（桌面隱藏） */
.hamburger {
  display: none;
  flex-direction: column;
  gap: 4px;
  cursor: pointer;
  padding: 6px;
  border: none;
  background: none;
}

.hamburger span {
  display: block;
  width: 18px;
  height: 1.5px;
  background: var(--black);
}

/* mobile 全屏導航 */
.mobile-menu {
  position: fixed;
  z-index: 100;
  inset: 0;
  display: flex;
  flex-direction: column;
  background: var(--sur);
}

.mobile-menu-header {
  display: flex;
  height: var(--nav-h);
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--g2);
  padding: 0 16px;
  flex-shrink: 0;
}

.mobile-menu-logo {
  font-family: var(--font-serif);
  font-size: 13px;
  letter-spacing: 2px;
  color: var(--ink);
  text-decoration: none;
}

.mobile-close {
  display: inline-flex;
  width: 32px;
  height: 32px;
  align-items: center;
  justify-content: center;
  background: none;
  border: none;
  color: var(--ink);
  cursor: pointer;
}

.mobile-close svg {
  width: 18px;
  height: 18px;
  stroke: currentColor;
  stroke-width: 2;
  fill: none;
  stroke-linecap: round;
}

.mobile-menu-body {
  flex: 1;
  overflow-y: auto;
}

.mnl {
  display: block;
  width: 100%;
  text-align: left;
  font-size: 15px;
  color: var(--ink);
  padding: 16px 20px;
  border: none;
  border-bottom: 1px solid var(--g1);
  background: none;
  font-family: inherit;
  cursor: pointer;
  text-decoration: none;
}

.mnl.on {
  color: var(--brand);
  font-weight: 600;
}

.mobile-menu-footer {
  display: grid;
  gap: 12px;
  border-top: 1px solid var(--g2);
  padding: 16px 20px;
}

.mobile-menu-theme,
.mobile-menu-signout {
  height: 40px;
  border: 1px solid var(--bdr);
  border-radius: var(--r-md);
  background: var(--sur-2);
  color: var(--ink);
  cursor: pointer;
  font: inherit;
  font-weight: 600;
}

/* 底部導航欄（桌面隱藏） */
.bottom-nav {
  display: none;
}

/* 滾動進度條 */
.scroll-progress {
  position: fixed;
  top: var(--nav-h);
  left: 0;
  height: 2px;
  background: var(--brand);
  z-index: 25;
  transition: width 0.1s;
  width: 0%;
}

/* 全域 toast 容器 */
.toast-container {
  position: fixed;
  bottom: 70px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 300;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  pointer-events: none;
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
  .nav-links,
  .nav-login {
    display: none;
  }

  .hamburger {
    display: flex;
  }

  .bottom-nav {
    position: fixed;
    z-index: 20;
    right: 0;
    bottom: 0;
    left: 0;
    display: flex;
    height: 54px;
    align-items: center;
    justify-content: space-around;
    border-top: 1px solid var(--bdr);
    background: var(--sur);
    box-shadow: 0 -2px 12px rgba(0, 0, 0, 0.06);
  }

  .bnav-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
    font-size: 9px;
    color: var(--g4);
    cursor: pointer;
    border: none;
    background: none;
    font-family: inherit;
    flex: 1;
    padding: 4px 0;
    text-decoration: none;
  }

  .bnav-icon {
    font-size: 19px;
    line-height: 1;
  }

  .bnav-item.on {
    color: var(--brand);
  }
}
</style>
