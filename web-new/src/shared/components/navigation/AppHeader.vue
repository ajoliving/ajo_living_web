<!--
 * 全域頂部導航。
 * 1. 嚴格對齊 HTML 設計稿 chrome.html 的 .nav 結構與樣式。
 * 2. 按登入態輸出公開入口、會員入口與通知中心鈴鐺。
 * 3. 整合主題、語系切換按鈕與登入入口。
 * 4. 保留 mobile 側邊抽屜與底部主入口。
-->
<script setup lang="ts">
import { computed, watch } from 'vue';
import { RouterLink, useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';

import AppIcon from '@/shared/components/base/AppIcon.vue';
import { type AppLocale, usePreferenceStore } from '@/stores/preferences';
import { useSessionStore } from '@/stores/session';
import { type AppThemeName } from '@/utils/theme';

interface NavigationItem {
  key: string;
  to: string;
  label: string;
  match: string[];
  icon?: 'home' | 'building' | 'browse' | 'star' | 'user';
  requiresAuth?: boolean;
  requiresStaff?: boolean;
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
const themeToggleAriaLabel = computed(() =>
  isDarkTheme.value ? t('common.action.switchToLight') : t('common.action.switchToDark'),
);
const localeToggleLabel = computed(() =>
  preferenceStore.locale === 'zh-HK'
    ? t('common.locale.enShort')
    : t('common.locale.zhHkShort'),
);
const isNavigationItemVisible = (item: NavigationItem): boolean => {
  if (item.requiresAuth && !sessionStore.isAuthenticated) {
    return false;
  }

  if (item.requiresStaff && !sessionStore.currentUser.is_staff) {
    return false;
  }

  return true;
};

// 1. 建立主導航入口
const allPrimaryNavigationItems = computed<NavigationItem[]>(() => [
  { key: 'home', to: '/', label: '首頁', match: ['/'] },
  { key: 'listing', to: '/properties', label: '樓盤租售', match: ['/properties'] },
  { key: 'service', to: '/serviced-residences', label: '服務式住宅', match: ['/serviced-residence', '/serviced-residences'] },
  { key: 'market', to: '/furniture', label: '家具', match: ['/furniture'] },
  { key: 'offers', to: '/supermarket-offers', label: '綜合優惠', match: ['/offers', '/supermarket-offers'] },
  { key: 'payment', to: '/payments', label: 'AJO Pay', match: ['/payments'], requiresAuth: true },
  { key: 'affairs', to: '/building', label: '我的大廈', match: ['/building'], requiresAuth: true },
  { key: 'profile', to: '/profile', label: '會員中心', match: ['/profile', '/account'], requiresAuth: true },
  { key: 'management', to: '/account/marketplace/management', label: '管理', match: ['/account/marketplace/management'], requiresAuth: true, requiresStaff: true },
  { key: 'trend', to: '/trend', label: '走勢', match: ['/trend'] },
]);
const primaryNavigationItems = computed<NavigationItem[]>(() =>
  allPrimaryNavigationItems.value.filter(isNavigationItemVisible),
);

// 2. mobile 全屏導航項
const mobileNavigationItems = computed<NavigationItem[]>(() => [
  ...primaryNavigationItems.value,
  ...(
    sessionStore.isAuthenticated
      ? [{ key: 'notification', to: '/notifications', label: '通知中心', match: ['/notifications', '/account/notifications'] }]
      : [{ key: 'login', to: '/login', label: '登入', match: ['/login'] }]
  ),
]);

// 3. 底部導航入口
const bottomNavigationItems = computed<NavigationItem[]>(() => {
  const items: NavigationItem[] = [
    { key: 'home', to: '/', label: '首頁', match: ['/'], icon: 'home' },
    { key: 'listing', to: '/properties', label: '樓盤', match: ['/properties'], icon: 'building' },
    { key: 'market', to: '/furniture', label: '家具', match: ['/furniture'], icon: 'browse' },
    { key: 'offers', to: '/supermarket-offers', label: '優惠', match: ['/offers', '/supermarket-offers'], icon: 'star' },
    {
      key: 'profile',
      to: accountPath.value,
      label: sessionStore.isAuthenticated ? '我的' : '登入',
      match: ['/profile', '/account', '/login'],
      icon: 'user',
    },
  ];

  return items.filter(isNavigationItemVisible);
});

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

// 7. 切換繁中與 English 語系
const handleLocaleToggle = (): void => {
  const nextLocale: AppLocale = preferenceStore.locale === 'zh-HK' ? 'en' : 'zh-HK';
  preferenceStore.setLocale(nextLocale);
};

// 8. 執行帳戶入口操作
const handleAccountAction = async (): Promise<void> => {
  mobileDrawerOpen.value = false;
  await router.push(accountPath.value);
};

// 9. 執行登出
const handleSignOut = async (): Promise<void> => {
  mobileDrawerOpen.value = false;
  await sessionStore.signOut();
  await router.push('/login');
};

// 10. 路由切換時收起 mobile 導航
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
        v-if="sessionStore.isAuthenticated"
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
      :aria-label="themeToggleAriaLabel"
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
      class="locale-toggle"
      :aria-label="t('common.action.switchLanguage')"
      @click="handleLocaleToggle"
    >
      {{ localeToggleLabel }}
    </button>
    <RouterLink
      v-if="!sessionStore.isAuthenticated"
      to="/login"
      class="mobile-login"
    >
      登入
    </RouterLink>
  </nav>

  <div class="scroll-progress"></div>
  <div class="toast-container"></div>

  <Teleport to="body">
    <transition name="drawer-fade">
      <div
        v-if="mobileDrawerOpen"
        class="mobile-menu"
        @click.self="mobileDrawerOpen = false"
      >
        <aside
          class="mobile-menu-panel"
          role="dialog"
          aria-modal="true"
          aria-labelledby="mobile-menu-title"
        >
          <div class="mobile-menu-header">
            <RouterLink
              id="mobile-menu-title"
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
        </aside>
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
      <span class="bnav-icon">
        <AppIcon
          v-if="item.icon"
          :name="item.icon"
          :size="22"
          :stroke-width="2"
        />
      </span>
      {{ item.label }}
    </RouterLink>
  </nav>
</template>

<style scoped>
/*
 * 樣式嚴格對齊 HTML 設計稿 .nav / .nav-logo / .nav-links / .nl / .nav-bell /
 * .nav-r / .nav-login / .theme-toggle / .locale-toggle / .hamburger / .mobile-menu /
 * .mobile-menu-panel /
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

/* 語系切換按鈕（桌面隱藏） */
.locale-toggle,
.mobile-login {
  display: none;
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

/* mobile 側邊抽屜 */
.mobile-menu {
  position: fixed;
  z-index: 100;
  inset: 0;
  background: rgb(0 0 0 / 0.48);
}

.mobile-menu-panel {
  display: flex;
  width: min(76vw, 336px);
  height: 100%;
  flex-direction: column;
  background: var(--sur);
  box-shadow: 12px 0 28px rgb(0 0 0 / 0.18);
  will-change: transform;
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
  display: inline-flex;
  min-height: 44px;
  align-items: center;
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

.drawer-fade-enter-active .mobile-menu-panel,
.drawer-fade-leave-active .mobile-menu-panel {
  transition: transform 0.22s ease;
}

.drawer-fade-enter-from,
.drawer-fade-leave-to {
  opacity: 0;
}

.drawer-fade-enter-from .mobile-menu-panel,
.drawer-fade-leave-to .mobile-menu-panel {
  transform: translateX(-100%);
}

@media (max-width: 1023px) {
  .nav-links,
  .nav-r {
    display: none;
  }

  .nav {
    height: calc(var(--nav-h) + var(--app-safe-top));
    justify-content: flex-start;
    padding: var(--app-safe-top) max(14px, calc(14px + var(--app-safe-right))) 0 max(14px, calc(14px + var(--app-safe-left)));
  }

  .nav-logo {
    position: absolute;
    z-index: 1;
    left: 50%;
    transform: translateX(-50%);
    font-size: 13px;
    border-bottom: none;
    padding-bottom: 0;
  }

  .theme-toggle,
  .locale-toggle,
  .mobile-login,
  .hamburger {
    width: 44px;
    height: 44px;
  }

  .theme-toggle {
    margin-left: auto;
  }

  .locale-toggle {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border: none;
    background: transparent;
    color: var(--ink);
    cursor: pointer;
    font: inherit;
    font-size: 11px;
    font-weight: 700;
  }

  .mobile-login {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    color: var(--brand);
    font-size: 13px;
    font-weight: 700;
    text-decoration: none;
  }

  .locale-toggle:hover {
    color: var(--brand);
  }

  .mobile-login:hover {
    color: var(--brand-dark);
  }

  html[data-theme='dark-neutral'] .locale-toggle {
    color: #E5E7EB;
  }

  html[data-theme='dark-neutral'] .locale-toggle:hover {
    color: var(--brand);
  }

  html[data-theme='dark-neutral'] .mobile-login:hover {
    color: var(--brand-mid);
  }

  .hamburger {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 999px;
  }

  .bottom-nav {
    position: fixed;
    z-index: 20;
    right: 0;
    bottom: 0;
    left: 0;
    display: flex;
    height: calc(var(--app-mobile-bottom-nav-height) + var(--app-safe-bottom));
    align-items: center;
    justify-content: space-around;
    border-top: 1px solid var(--bdr);
    background: var(--sur);
    padding: 6px max(8px, var(--app-safe-right)) calc(6px + var(--app-safe-bottom)) max(8px, var(--app-safe-left));
    box-shadow: 0 -8px 24px rgba(0, 0, 0, 0.08);
  }

  .bnav-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 3px;
    min-width: 0;
    min-height: 52px;
    border-radius: 12px;
    font-size: 10px;
    font-weight: 700;
    color: var(--g4);
    cursor: pointer;
    border: none;
    background: none;
    font-family: inherit;
    flex: 1;
    padding: 4px 0;
    text-decoration: none;
    transition:
      background 0.15s ease,
      color 0.15s ease;
  }

  .bnav-icon {
    display: inline-flex;
    width: 24px;
    height: 24px;
    align-items: center;
    justify-content: center;
  }

  .bnav-item.on {
    background: var(--brand-light);
    color: var(--brand);
  }

  .mobile-menu-panel {
    box-sizing: border-box;
    padding-top: var(--app-safe-top);
    padding-right: var(--app-safe-right);
    padding-bottom: var(--app-safe-bottom);
    padding-left: var(--app-safe-left);
  }

  .mobile-menu-header {
    height: var(--nav-h);
    padding: 0 16px;
  }

  .mobile-close {
    width: 44px;
    height: 44px;
  }

  .mobile-menu-body {
    display: grid;
    align-content: start;
    gap: 8px;
    padding: 12px 16px;
  }

  .mnl {
    min-height: 48px;
    border: 1px solid var(--bdr);
    border-radius: 8px;
    padding: 13px 14px;
  }

  .mnl.on {
    border-color: var(--brand-mid);
    background: var(--brand-light);
  }

  .mobile-menu-footer {
    padding: 14px 16px;
  }

  .mobile-menu-theme,
  .mobile-menu-signout {
    height: 46px;
  }

  .scroll-progress {
    top: var(--app-header-offset);
  }

  .toast-container {
    right: 16px;
    bottom: calc(var(--app-mobile-content-bottom) + 12px);
    left: 16px;
    transform: none;
  }
}
</style>
