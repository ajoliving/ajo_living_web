<!--
 * 全域頂部導航。
 * 1. 提供平台主路由與當前業務子路由雙層導航。
 * 2. 整合主題切換、語系切換與登入狀態操作。
 * 3. 提供自定義 mobile 抽屜導航。
-->
<script setup lang="ts">
import { computed, watch } from 'vue';
import { RouterLink, useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';

import AppIcon from '@/shared/components/base/AppIcon.vue';
import BaseAvatar from '@/shared/components/base/BaseAvatar.vue';
import BaseButton from '@/shared/components/base/BaseButton.vue';
import BaseSelect from '@/shared/components/base/BaseSelect.vue';
import TopbarDropdown from '@/shared/components/navigation/TopbarDropdown.vue';
import { type AppLocale, usePreferenceStore } from '@/stores/preferences';
import { useHeaderSubnav } from '@/shared/navigation/useHeaderSubnav';
import { useSessionStore } from '@/stores/session';
import { type AppThemeName } from '@/utils/theme';

const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const preferenceStore = usePreferenceStore();
const sessionStore = useSessionStore();
const headerLocaleOptions = computed(() => [
  { label: t('common.locale.zhHkShort'), value: 'zh-HK' as AppLocale },
  { label: t('common.locale.enShort'), value: 'en' as AppLocale },
]);
const mobileLocaleOptions = computed(() => [
  { label: t('common.locale.zhHk'), value: 'zh-HK' as AppLocale },
  { label: t('common.locale.en'), value: 'en' as AppLocale },
]);
const themeOptions = computed(() => [
  { label: t('common.theme.default'), value: 'default' as AppThemeName },
  { label: t('common.theme.copperSun'), value: 'copper-sun' as AppThemeName },
  { label: t('common.theme.darkNeutral'), value: 'dark-neutral' as AppThemeName },
]);
const brandLogoSrc = '/brand/ajo-living-logo.png';

// 1. 依路由配置判斷頂部導航是否使用透明覆蓋模式
const isTransparentRoute = computed(() => route.meta.headerMode === 'transparent');

interface NavigationItem {
  key: string;
  to: string;
  label: string;
  match?: string[];
}

interface PrimaryNavigationItem extends NavigationItem {}

const primaryNavigationItems = computed<PrimaryNavigationItem[]>(() => [
  { key: 'home', to: '/', label: t('nav.home'), match: ['/'] },
  { key: 'properties', to: '/properties', label: t('nav.properties'), match: ['/properties'] },
  { key: 'servicedResidences', to: '/serviced-residences', label: t('nav.servicedResidences'), match: ['/serviced-residences'] },
  {
    key: 'marketplace',
    to: '/marketplace/discover',
    label: t('nav.marketplace'),
    match: ['/marketplace'],
  },
  { key: 'account', to: '/account/profile', label: t('nav.account'), match: ['/account', '/login'] },
]);

const accountNavigationItem = computed<NavigationItem>(() => ({
  key: 'account',
  to: sessionStore.isAuthenticated ? '/account/profile' : '/login',
  label: t('nav.account'),
  match: ['/account', '/login'],
}));

const desktopNavigationItems = computed<PrimaryNavigationItem[]>(() => [
  ...primaryNavigationItems.value.filter((item) => item.key !== 'account'),
  accountNavigationItem.value,
]);

const { subnavItems, isSubnavActive } = useHeaderSubnav();

const mobileNavigationItems = computed<PrimaryNavigationItem[]>(() => [
  ...desktopNavigationItems.value,
]);

const isRouteActive = (item: NavigationItem) => {
  const matchedPaths = item.match ?? [item.to];

  return matchedPaths.some((path) => (path === '/' ? route.path === '/' : route.path.startsWith(path)));
};


const mobileDrawerOpen = computed({
  get: () => preferenceStore.mobile_menu_open,
  set: (value: boolean) => {
    preferenceStore.setMobileMenuOpen(value);
  },
});

// 2. 判斷目前頁面是否有子路由導航
const hasSubNavigation = computed(() => subnavItems.value.length > 0);

// 3. 切換主題
const handleThemeChange = (value: string | number) => {
  preferenceStore.setTheme(value as AppThemeName);
};

// 4. 切換語系
const handleLocaleChange = (value: string | number) => {
  preferenceStore.setLocale(value as AppLocale);
};

// 5. 導向登入頁
const handleNavigateLogin = async () => {
  mobileDrawerOpen.value = false;
  await router.push('/login');
};

// 6. 導向會員中心
const handleNavigateAccount = async () => {
  mobileDrawerOpen.value = false;
  await router.push('/account/profile');
};

// 7. 執行登出
const handleSignOut = async () => {
  mobileDrawerOpen.value = false;
  await sessionStore.signOut();
  await router.push('/login');
};

// 8. 路由切換時自動收起 mobile 抽屜
watch(
  () => route.fullPath,
  () => {
    mobileDrawerOpen.value = false;
  },
);
</script>

<template>
  <header
    class="app-topbar fixed inset-x-0 top-0 z-40"
    :class="isTransparentRoute ? 'app-topbar-home' : 'app-topbar-solid'"
  >
    <div class="app-topbar__inner flex min-h-[3.725rem] items-center gap-4 px-4 sm:px-6 lg:px-8 2xl:px-10">
      <div class="topbar-brand-shell hidden lg:flex lg:flex-1 lg:items-center">
        <RouterLink
          to="/"
          class="brand-link shrink-0"
        >
          <img
            class="brand-logo"
            :src="brandLogoSrc"
            alt="AJO Living"
          />
          <span class="brand-wordmark">AJOLIVING</span>
        </RouterLink>
      </div>

      <div class="topbar-main-nav-shell flex min-w-0 flex-1 justify-center lg:flex-[1.4]">
        <nav class="hidden min-w-0 items-center justify-center gap-1.5 lg:flex">
          <RouterLink
            v-for="item in desktopNavigationItems"
            :key="item.key"
            :to="item.to"
            class="topbar-nav-link inline-flex items-center rounded-pill px-4 py-2.5 text-sm font-semibold whitespace-nowrap transition xl:px-5"
            :class="
              isRouteActive(item)
                ? 'topbar-nav-link-active'
                : 'topbar-nav-link-idle'
            "
          >
            <span>{{ item.label }}</span>
          </RouterLink>
        </nav>
      </div>

      <div class="hidden flex-1 items-center justify-end gap-6 whitespace-nowrap lg:flex xl:gap-7">
        <div class="nav-switch-group">
          <TopbarDropdown
            :model-value="preferenceStore.theme"
            :options="themeOptions"
            align="left"
            panel-width="8.75rem"
            @update:model-value="handleThemeChange"
          />

          <TopbarDropdown
            :model-value="preferenceStore.locale"
            :options="headerLocaleOptions"
            align="left"
            panel-width="5.75rem"
            @update:model-value="handleLocaleChange"
          />
        </div>

        <button
          v-if="sessionStore.isAuthenticated"
          type="button"
          class="nav-account-avatar-button"
          :aria-label="t('nav.account')"
          @click="handleNavigateAccount"
        >
          <BaseAvatar
            :src="sessionStore.currentUser.avatar_url"
            :name="sessionStore.currentUser.display_name"
            :size="34"
          />
        </button>

        <button
          v-else
          type="button"
          class="topbar-auth-link"
          :class="isRouteActive(accountNavigationItem) ? 'topbar-auth-link-active' : 'topbar-auth-link-idle'"
          @click="handleNavigateLogin"
        >
          {{ t('common.action.signIn') }}
        </button>
      </div>

      <BaseButton
        variant="ghost"
        class="lg:hidden"
        @click="mobileDrawerOpen = true"
      >
        <template #leading>
          <AppIcon
            name="menu"
            :size="18"
          />
        </template>
      </BaseButton>
    </div>

    <transition name="subnav-fade">
      <div
        v-if="hasSubNavigation"
        class="app-subnav hidden border-t border-border/40 lg:block"
      >
        <div class="app-subnav__inner">
          <nav class="flex min-w-0 items-center gap-2 overflow-x-auto">
            <RouterLink
              v-for="item in subnavItems"
              :key="item.key"
              :to="item.to"
              class="subnav-link inline-flex items-center rounded-pill px-4 py-2 text-sm font-semibold whitespace-nowrap transition"
              :class="isSubnavActive(item) ? 'subnav-link-active' : 'subnav-link-idle'"
            >
              {{ item.label }}
            </RouterLink>
          </nav>
        </div>
      </div>
    </transition>

    <Teleport to="body">
      <transition name="drawer-fade">
        <div
          v-if="mobileDrawerOpen"
          class="fixed inset-0 z-50 bg-black/35 backdrop-blur-sm lg:hidden"
          @click="mobileDrawerOpen = false"
        />
      </transition>

      <transition name="drawer-slide">
        <aside
          v-if="mobileDrawerOpen"
          class="nav-drawer fixed right-0 top-0 z-[60] flex h-full w-[320px] flex-col gap-4 border-l border-border/80 px-5 py-5 shadow-floating lg:hidden"
        >
          <div class="flex items-center justify-between">
            <img
              class="brand-logo brand-logo--mobile"
              :src="brandLogoSrc"
              alt="AJO Living"
            />
            <BaseButton
              variant="ghost"
              size="sm"
              @click="mobileDrawerOpen = false"
            >
              <template #leading>
                <AppIcon
                  name="close"
                  :size="16"
                />
              </template>
            </BaseButton>
          </div>

          <div class="rounded-3xl border border-border/80 bg-surface p-4">
            <p class="text-sm font-semibold text-text">
              {{ sessionStore.currentUser.display_name }}
            </p>
            <p class="mt-1 text-sm text-text-muted">
              {{ sessionStore.currentUser.primary_community.name }}
            </p>
            <BaseButton
              class="mt-4"
              block
              @click="handleNavigateAccount"
            >
              {{ t('nav.account') }}
            </BaseButton>
          </div>

          <div class="space-y-2">
            <RouterLink
              v-for="item in mobileNavigationItems"
              :key="item.key"
              :to="item.to"
              class="topbar-nav-link-mobile flex items-center rounded-2xl border px-4 py-3 text-base font-semibold whitespace-nowrap"
              :class="
                isRouteActive(item)
                  ? 'border-primary/30 bg-primary/10 text-primary'
                  : 'border-border/80 bg-surface text-text'
              "
            >
              <span>{{ item.label }}</span>
            </RouterLink>
          </div>

          <div
            v-if="hasSubNavigation"
            class="space-y-2 border-t border-border/70 pt-4"
          >
            <p class="px-1 text-xs font-semibold uppercase tracking-[0.18em] text-text-muted">
              {{ t('nav.marketplaceSubroutes') }}
            </p>
            <RouterLink
              v-for="item in subnavItems"
              :key="item.key"
              :to="item.to"
              class="topbar-subnav-link-mobile flex items-center rounded-2xl border px-4 py-2.5 text-sm font-semibold whitespace-nowrap"
              :class="isSubnavActive(item) ? 'topbar-subnav-link-mobile-active' : 'topbar-subnav-link-mobile-idle'"
            >
              <span>{{ item.label }}</span>
            </RouterLink>
          </div>

          <div class="space-y-3 pt-2">
            <div>
              <p class="app-field-label text-xs uppercase tracking-[0.18em] text-text-muted">
                {{ t('common.action.switchTheme') }}
              </p>
              <BaseSelect
                :model-value="preferenceStore.theme"
                :options="themeOptions"
                @update:model-value="handleThemeChange"
              />
            </div>

            <div>
              <p class="app-field-label text-xs uppercase tracking-[0.18em] text-text-muted">
                {{ t('common.action.switchLanguage') }}
              </p>
              <BaseSelect
                :model-value="preferenceStore.locale"
                :options="mobileLocaleOptions"
                @update:model-value="handleLocaleChange"
              />
            </div>
          </div>

          <BaseButton
            v-if="sessionStore.isAuthenticated"
            variant="danger"
            block
            @click="handleSignOut"
          >
            {{ t('common.action.signOut') }}
          </BaseButton>

          <BaseButton
            v-else
            variant="primary"
            block
            @click="handleNavigateLogin"
          >
            {{ t('common.action.signIn') }}
          </BaseButton>
        </aside>
      </transition>
    </Teleport>
  </header>
</template>

<style scoped>
.app-topbar__inner {
  position: relative;
  z-index: 3;
  border-bottom: 1px solid var(--topbar-glass-border);
  transform: none;
  transition:
    border-color 0.28s ease;
}

.topbar-main-nav-shell {
  gap: 0;
}

.topbar-brand-shell {
  min-width: 0;
}

.brand-wordmark {
  display: inline-flex;
  align-items: center;
  font-family: var(--font-display);
  font-size: 0.95rem;
  font-weight: 700;
  letter-spacing: 0.28em;
  text-transform: uppercase;
  color: rgb(var(--color-text));
  text-shadow: 0 8px 24px rgb(var(--color-primary) / 0.12);
  transition:
    color 0.24s ease,
    text-shadow 0.24s ease;
}

.app-topbar {
  --topbar-glass-border: rgb(var(--color-border) / 0.3);
  --topbar-glass-background: linear-gradient(
    180deg,
    rgb(var(--color-topbar-surface) / 0.76),
    rgb(var(--color-toolbar-surface) / 0.64)
  );
  --topbar-glass-shadow:
    0 16px 40px rgb(15 23 42 / 0.1),
    inset 0 -1px 0 rgb(255 255 255 / 0.06);
  --topbar-glass-filter: blur(28px) saturate(184%);
  --topbar-nav-active-color: color-mix(in srgb, rgb(var(--color-primary)) 78%, rgb(var(--color-text)) 22%);
  --topbar-nav-active-shadow: 0 0 10px rgb(var(--color-primary) / 0.16);
  --topbar-nav-underline: color-mix(in srgb, rgb(var(--color-primary)) 88%, rgb(var(--color-text)) 12%);
  --topbar-auth-idle-color: rgb(var(--color-text) / 0.88);
  --topbar-auth-hover-color: rgb(var(--color-text));
  --topbar-auth-active-color: var(--topbar-nav-active-color);
  --topbar-auth-active-shadow: var(--topbar-nav-active-shadow);
  position: fixed;
  inset-inline: 0;
  top: 0;
  z-index: 40;
  background: var(--topbar-glass-background);
  box-shadow: var(--topbar-glass-shadow);
  backdrop-filter: var(--topbar-glass-filter);
  -webkit-backdrop-filter: var(--topbar-glass-filter);
  transition:
    background 0.28s ease,
    border-color 0.28s ease,
    box-shadow 0.28s ease;
}

.app-topbar-home {
  --topbar-glass-border: rgb(var(--color-border) / 0.16);
  --topbar-glass-background: linear-gradient(
    180deg,
    rgb(var(--color-topbar-surface) / 0.5),
    rgb(var(--color-toolbar-surface) / 0.36)
  );
  --topbar-glass-shadow:
    0 12px 30px rgb(15 23 42 / 0.06),
    inset 0 -1px 0 rgb(255 255 255 / 0.08);
  --topbar-glass-filter: blur(22px) saturate(164%);
}

.app-topbar-solid {
  --topbar-glass-border: rgb(var(--color-border) / 0.3);
  --topbar-glass-background: linear-gradient(
    180deg,
    rgb(var(--color-topbar-surface) / 0.76),
    rgb(var(--color-toolbar-surface) / 0.64)
  );
  --topbar-glass-shadow:
    0 16px 40px rgb(15 23 42 / 0.1),
    inset 0 -1px 0 rgb(255 255 255 / 0.06);
  --topbar-glass-filter: blur(28px) saturate(184%);
}

:global(html[data-theme='default']) .app-topbar {
  --topbar-nav-active-color: color-mix(in srgb, rgb(var(--color-primary)) 76%, rgb(var(--color-text)) 24%);
  --topbar-nav-active-shadow: 0 0 10px rgb(var(--color-primary) / 0.12);
  --topbar-nav-underline: color-mix(in srgb, rgb(var(--color-primary)) 82%, rgb(var(--color-text)) 18%);
  --topbar-auth-idle-color: rgb(var(--color-text) / 0.82);
  --topbar-auth-hover-color: rgb(var(--color-text));
  --topbar-auth-active-color: color-mix(in srgb, rgb(var(--color-primary)) 74%, rgb(var(--color-text)) 26%);
  --topbar-auth-active-shadow: 0 0 9px rgb(var(--color-primary) / 0.1);
}

:global(html[data-theme='copper-sun']) .app-topbar {
  --topbar-nav-active-color: color-mix(in srgb, rgb(var(--color-primary)) 84%, rgb(var(--color-text)) 16%);
  --topbar-nav-active-shadow: 0 0 12px rgb(var(--color-primary) / 0.18);
  --topbar-nav-underline: color-mix(in srgb, rgb(var(--color-primary)) 92%, rgb(var(--color-text)) 8%);
  --topbar-auth-idle-color: rgb(var(--color-text) / 0.88);
  --topbar-auth-hover-color: rgb(var(--color-text));
  --topbar-auth-active-color: color-mix(in srgb, rgb(var(--color-primary)) 82%, rgb(var(--color-text)) 18%);
  --topbar-auth-active-shadow: 0 0 10px rgb(var(--color-primary) / 0.14);
}

.nav-drawer {
  background:
    linear-gradient(
      180deg,
      rgb(var(--color-topbar-surface) / 0.84),
      rgb(var(--color-toolbar-surface) / 0.78)
    );
  backdrop-filter: blur(18px) saturate(138%);
  -webkit-backdrop-filter: blur(18px) saturate(138%);
}

.brand-link {
  display: inline-flex;
  align-items: center;
  gap: 0.8rem;
}

.brand-logo {
  display: block;
  width: auto;
  height: 2.45rem;
  object-fit: contain;
}

.brand-logo--mobile {
  height: 2.25rem;
}

:global(html[data-theme='default']) .brand-wordmark {
  color: rgb(var(--color-text));
}

:global(html[data-theme='copper-sun']) .brand-wordmark {
  color: color-mix(in srgb, rgb(var(--color-primary)) 76%, rgb(var(--color-text)) 24%);
  text-shadow: 0 10px 24px rgb(var(--color-primary) / 0.18);
}

.nav-switch-group {
  display: flex;
  align-items: center;
  gap: 0.95rem;
}

.topbar-nav-link {
  position: relative;
  font-family: var(--font-display);
  color: rgb(var(--color-text) / 0.98);
  border: 1px solid transparent;
  background: transparent !important;
  text-shadow: none;
}

.topbar-nav-link::after {
  position: absolute;
  left: 1rem;
  right: 1rem;
  bottom: 0.42rem;
  height: 1.5px;
  border-radius: 999px;
  background: var(--topbar-nav-underline);
  transform: scaleX(0);
  transform-origin: left center;
  opacity: 0;
  transition:
    transform 0.22s ease,
    opacity 0.22s ease;
  content: '';
}

.topbar-nav-link-idle:hover {
  color: rgb(var(--color-text));
}

.topbar-nav-link-active {
  color: var(--topbar-nav-active-color) !important;
  text-shadow: var(--topbar-nav-active-shadow);
}

.topbar-nav-link-active::after {
  transform: scaleX(1);
  opacity: 1;
}

.topbar-nav-link-mobile {
  font-family: var(--font-display);
}

.topbar-subnav-link-mobile {
  border-color: rgb(var(--color-border) / 0.28);
  background:
    linear-gradient(
      180deg,
      rgb(var(--color-topbar-surface) / 0.72),
      rgb(var(--color-toolbar-surface) / 0.56)
    );
  font-family: var(--font-display);
  box-shadow:
    0 12px 28px rgb(15 23 42 / 0.08),
    inset 0 1px 0 rgb(255 255 255 / 0.08);
  backdrop-filter: blur(18px) saturate(160%);
  -webkit-backdrop-filter: blur(18px) saturate(160%);
  transition:
    border-color 0.2s ease,
    background 0.2s ease,
    color 0.2s ease,
    box-shadow 0.2s ease;
}

.topbar-subnav-link-mobile-idle {
  color: rgb(var(--color-text-muted));
}

.topbar-subnav-link-mobile-active,
.topbar-subnav-link-mobile-active:hover {
  border-color: rgb(var(--color-primary) / 0.28);
  background:
    linear-gradient(
      180deg,
      rgb(var(--color-topbar-surface) / 0.86),
      rgb(var(--color-primary-soft) / 0.24)
    );
  color: var(--topbar-nav-active-color);
  box-shadow:
    0 14px 32px rgb(var(--color-primary) / 0.1),
    inset 0 1px 0 rgb(255 255 255 / 0.1);
}

.topbar-subnav-link-mobile-idle:hover {
  border-color: rgb(var(--color-border) / 0.4);
  color: rgb(var(--color-text));
}

.app-subnav {
  position: relative;
  z-index: 1;
  border-bottom: 1px solid var(--topbar-glass-border);
}

.app-subnav__inner {
  display: flex;
  min-height: 2.65rem;
  width: 100%;
  max-width: var(--layout-page-max-width);
  align-items: center;
  margin: 0 auto;
  padding-right: var(--layout-page-padding-inline);
  padding-left: var(--layout-page-padding-inline);
}

.app-subnav__inner nav {
  margin-left: -1rem;
}

.app-topbar-home .app-subnav {
  border-top-color: var(--topbar-glass-border);
  border-bottom-color: var(--topbar-glass-border);
}

.subnav-link {
  position: relative;
  font-family: var(--font-display);
  color: rgb(var(--color-text) / 0.98);
  border: 1px solid transparent;
  background: transparent;
  text-shadow: none;
}

.subnav-link::after {
  position: absolute;
  left: 0.9rem;
  right: 0.9rem;
  bottom: 0.32rem;
  height: 1.5px;
  border-radius: 999px;
  background: var(--topbar-nav-underline);
  transform: scaleX(0);
  transform-origin: left center;
  opacity: 0;
  transition:
    transform 0.22s ease,
    opacity 0.22s ease;
  content: '';
}

.subnav-link-idle:hover {
  color: rgb(var(--color-text));
}

.subnav-link-active {
  color: var(--topbar-nav-active-color) !important;
  text-shadow: var(--topbar-nav-active-shadow);
}

.subnav-link-active::after {
  transform: scaleX(1);
  opacity: 1;
}

.topbar-auth-link {
  position: relative;
  border: 1px solid transparent;
  background: transparent;
  padding: 0.25rem 0;
  font-family: var(--font-display);
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--topbar-auth-idle-color);
  cursor: pointer;
  transition:
    color 0.2s ease,
    text-shadow 0.2s ease;
}

.topbar-auth-link::after {
  position: absolute;
  left: 0;
  right: 0;
  bottom: -0.06rem;
  height: 1.5px;
  border-radius: 999px;
  background: var(--topbar-nav-underline);
  transform: scaleX(0);
  transform-origin: left center;
  opacity: 0;
  transition:
    transform 0.22s ease,
    opacity 0.22s ease;
  content: '';
}

.topbar-auth-link-idle:hover {
  color: var(--topbar-auth-hover-color);
}

.topbar-auth-link-active {
  color: var(--topbar-auth-active-color);
  text-shadow: var(--topbar-auth-active-shadow);
}

.topbar-auth-link-active::after {
  transform: scaleX(1);
  opacity: 1;
}

.topbar-auth-link:hover {
  color: var(--topbar-auth-hover-color);
}

.nav-account-avatar-button {
  display: inline-flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: var(--radius-pill);
  background: transparent;
  padding: 0;
  cursor: pointer;
  transition:
    opacity 0.2s ease,
    transform 0.2s ease;
}

.nav-account-avatar-button:hover {
  opacity: 0.86;
  transform: translateY(-1px);
}

.nav-account-avatar-button:focus-visible {
  outline: 2px solid rgb(var(--color-primary));
  outline-offset: 3px;
}

.drawer-fade-enter-active,
.drawer-fade-leave-active,
.drawer-slide-enter-active,
.drawer-slide-leave-active {
  transition:
    opacity 0.22s ease,
    transform 0.22s ease;
}

.drawer-fade-enter-from,
.drawer-fade-leave-to {
  opacity: 0;
}

.drawer-slide-enter-from,
.drawer-slide-leave-to {
  opacity: 0;
  transform: translateX(20px);
}

.subnav-fade-enter-active,
.subnav-fade-leave-active {
  transition:
    opacity 0.2s ease,
    transform 0.2s ease;
}

.subnav-fade-enter-from,
.subnav-fade-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

@media (min-width: 1280px) {
  .topbar-main-nav-shell {
    gap: 2.5rem;
  }
}
</style>
