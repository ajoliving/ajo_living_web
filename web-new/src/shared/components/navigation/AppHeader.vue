<!--
 * 全域頂部導航。
 * 1. 提供平台主模組入口。
 * 2. 整合主題、語系與登入狀態操作。
 * 3. 提供 mobile 全屏導航與底部主入口。
-->
<script setup lang="ts">
import { computed, watch } from 'vue';
import { RouterLink, useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';

import AppIcon from '@/shared/components/base/AppIcon.vue';
import { type AppLocale, usePreferenceStore } from '@/stores/preferences';
import { useSessionStore } from '@/stores/session';
import { type AppThemeName, THEME_OPTIONS } from '@/utils/theme';

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

const themeOptions = computed<Array<{ label: string; value: AppThemeName }>>(() => [
  ...THEME_OPTIONS.map((option) => ({
    label: t(option.labelKey),
    value: option.value,
  })),
]);
const localeOptions = computed<Array<{ label: string; value: AppLocale }>>(() => [
  { label: t('common.locale.zhHkShort'), value: 'zh-HK' },
  { label: t('common.locale.enShort'), value: 'en' },
]);
const mobileDrawerOpen = computed({
  get: () => preferenceStore.mobile_menu_open,
  set: (value: boolean) => {
    preferenceStore.setMobileMenuOpen(value);
  },
});
const accountPath = computed(() => (sessionStore.isAuthenticated ? '/account/profile' : '/login'));
const notificationUnreadCount = computed(() => 3);

// 1. 建立主導航入口
const primaryNavigationItems = computed<NavigationItem[]>(() => [
  { key: 'home', to: '/', label: t('nav.home'), match: ['/'] },
  { key: 'properties', to: '/properties', label: t('nav.properties'), match: ['/properties'] },
  { key: 'servicedResidences', to: '/serviced-residences', label: t('nav.servicedResidences'), match: ['/serviced-residences'] },
  { key: 'furniture', to: '/furniture', label: t('nav.furniture'), match: ['/furniture'] },
  { key: 'offers', to: '/supermarket-offers', label: t('nav.combinedOffers'), match: ['/supermarket-offers'] },
  { key: 'payments', to: '/payments', label: t('nav.payments'), match: ['/payments'] },
  { key: 'account', to: accountPath.value, label: t('nav.memberCenter'), match: ['/account', '/login'] },
  { key: 'notifications', to: '/notifications', label: t('nav.notifications'), match: ['/notifications'] },
]);

const mobileNavigationItems = computed<NavigationItem[]>(() => [
  ...primaryNavigationItems.value,
]);

// 2. 判斷目前路由是否命中導航項
const isRouteActive = (item: NavigationItem): boolean =>
  item.match.some((path) => (path === '/' ? route.path === '/' : route.path.startsWith(path)));

// 3. 切換主題與語系
const handleThemeChange = (event: Event): void => {
  const target = event.target as HTMLSelectElement;
  preferenceStore.setTheme(target.value as AppThemeName);
};

const handleLocaleChange = (event: Event): void => {
  const target = event.target as HTMLSelectElement;
  preferenceStore.setLocale(target.value as AppLocale);
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
      {{ t('common.brand.name') }}
    </RouterLink>

    <nav class="ajo-nav__links">
      <RouterLink
        v-for="item in primaryNavigationItems"
        :key="item.key"
        :to="item.to"
        class="ajo-nav__link"
        :class="{ 'ajo-nav__link--active': isRouteActive(item) }"
      >
        <template v-if="item.key === 'notifications'">
          <span
            class="ajo-nav__bell"
            :aria-label="item.label"
            :title="item.label"
          >
            <AppIcon
              name="bell"
              :size="17"
            />
            <i v-if="notificationUnreadCount > 0" />
          </span>
        </template>
        <template v-else>
          {{ item.label }}
        </template>
      </RouterLink>
    </nav>

    <div class="ajo-nav__right">
      <label class="ajo-nav__select">
        <span>{{ t('common.action.switchTheme') }}</span>
        <select
          :value="preferenceStore.theme"
          @change="handleThemeChange"
        >
          <option
            v-for="option in themeOptions"
            :key="option.value"
            :value="option.value"
          >
            {{ option.label }}
          </option>
        </select>
      </label>

      <label class="ajo-nav__select ajo-nav__select--compact">
        <span>{{ t('common.action.switchLanguage') }}</span>
        <select
          :value="preferenceStore.locale"
          @change="handleLocaleChange"
        >
          <option
            v-for="option in localeOptions"
            :key="option.value"
            :value="option.value"
          >
            {{ option.label }}
          </option>
        </select>
      </label>

      <button
        type="button"
        class="ajo-nav__login"
        @click="handleAccountAction"
      >
        {{ sessionStore.isAuthenticated ? t('nav.memberCenter') : t('common.action.signIn') }}
      </button>

      <button
        type="button"
        class="ajo-nav__menu"
        :aria-label="t('nav.account')"
        @click="mobileDrawerOpen = true"
      >
        <AppIcon
          name="menu"
          :size="18"
        />
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
              class="ajo-nav__logo"
            >
              {{ t('common.brand.name') }}
            </RouterLink>
            <button
              type="button"
              class="ajo-mobile-menu__close"
              :aria-label="t('common.action.close')"
              @click="mobileDrawerOpen = false"
            >
              <AppIcon
                name="close"
                :size="20"
              />
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
            <label class="ajo-mobile-menu__field">
              <span>{{ t('common.action.switchTheme') }}</span>
              <select
                :value="preferenceStore.theme"
                @change="handleThemeChange"
              >
                <option
                  v-for="option in themeOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ option.label }}
                </option>
              </select>
            </label>

            <label class="ajo-mobile-menu__field">
              <span>{{ t('common.action.switchLanguage') }}</span>
              <select
                :value="preferenceStore.locale"
                @change="handleLocaleChange"
              >
                <option
                  v-for="option in localeOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ option.label }}
                </option>
              </select>
            </label>

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
  height: 48px;
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
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 0.16em;
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
  height: 48px;
  align-items: center;
  border-bottom: 2px solid transparent;
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  font-weight: 500;
  padding: 0 8px;
  white-space: nowrap;
}

.ajo-nav__link:hover,
.ajo-nav__link--active {
  border-bottom-color: rgb(var(--color-primary));
  color: rgb(var(--color-primary));
}

.ajo-nav__bell {
  position: relative;
  display: inline-flex;
  width: 28px;
  height: 28px;
  align-items: center;
  justify-content: center;
}

.ajo-nav__bell i {
  position: absolute;
  top: 5px;
  right: 5px;
  width: 7px;
  height: 7px;
  border: 1px solid rgb(var(--color-surface));
  border-radius: 999px;
  background: #e11d48;
}

.ajo-nav__right {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 8px;
}

.ajo-nav__select {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: rgb(var(--color-text-muted));
  font-size: 10px;
}

.ajo-nav__select span {
  position: absolute;
  overflow: hidden;
  width: 1px;
  height: 1px;
  clip: rect(0 0 0 0);
  white-space: nowrap;
}

.ajo-nav__select select {
  height: 26px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text-muted));
  font: inherit;
  outline: none;
  padding: 0 6px;
}

.ajo-nav__login {
  height: 28px;
  border: 1px solid rgb(var(--color-primary));
  border-radius: 2px;
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
  cursor: pointer;
  font: inherit;
  font-size: 11px;
  font-weight: 600;
  padding: 0 12px;
  white-space: nowrap;
}

.ajo-nav__menu {
  display: none;
  width: 32px;
  height: 32px;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
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
  height: 48px;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid rgb(var(--color-border));
  padding: 0 16px;
}

.ajo-mobile-menu__close {
  display: inline-flex;
  width: 32px;
  height: 32px;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
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

.ajo-mobile-menu__field {
  display: grid;
  gap: 5px;
  color: rgb(var(--color-text-muted));
  font-size: 11px;
}

.ajo-mobile-menu__field select {
  height: 34px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  padding: 0 8px;
}

.ajo-mobile-menu__signout {
  height: 36px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface-muted));
  color: rgb(var(--color-text));
  font-weight: 600;
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
  .ajo-nav__select,
  .ajo-nav__login {
    display: none;
  }

  .ajo-nav__menu {
    display: inline-flex;
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
