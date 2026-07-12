<!--
 * 個人資料外框頁。
 * 1. 提供個人資料與 AJO 錢包的左側內部導航。
 * 2. 承載個人資料模組下的 RouterView。
-->
<script setup lang="ts">
import { computed } from 'vue';
import { RouterLink, RouterView, useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';

import AppIcon from '@/shared/components/base/AppIcon.vue';

type ProfileShellIconName = 'user' | 'star';

interface ProfileShellRouteItem {
  label: string;
  to: string;
  icon: ProfileShellIconName;
}

const route = useRoute();
const { t } = useI18n();

// 1. 建立個人資料內部導航
const profileRoutes = computed<ProfileShellRouteItem[]>(() => [
  {
    label: t('account.sections.profile'),
    to: '/account/profile/info',
    icon: 'user',
  },
  {
    label: t('account.sections.wallet'),
    to: '/account/profile/wallet',
    icon: 'star',
  },
]);

// 2. 判斷目前內部子路由是否啟用
const isActiveRoute = (path: string): boolean =>
  route.path === path || route.path.startsWith(`${path}/`);
</script>

<template>
  <main class="profile-shell">
    <aside class="profile-shell__sidebar">
      <div class="profile-shell__heading">
        <h1>{{ t('account.profile.title') }}</h1>
        <p>{{ t('account.profile.description') }}</p>
      </div>

      <section class="profile-shell__role-card">
        <span class="profile-shell__role-label">{{ t('account.profile.accountRole') }}</span>
        <strong class="profile-shell__role-value">{{ t('account.profile.primaryOwner') }}</strong>
        <p class="profile-shell__role-text">{{ t('account.profile.manageDescription') }}</p>
      </section>

      <div class="profile-shell__nav-heading">{{ t('account.profile.accountSection') }}</div>

      <nav class="profile-shell__nav">
        <RouterLink
          v-for="item in profileRoutes"
          :key="item.to"
          :to="item.to"
          class="profile-shell__nav-item"
          :class="{ 'profile-shell__nav-item--active': isActiveRoute(item.to) }"
        >
          <AppIcon
            :name="item.icon"
            :size="17"
          />
          <span>{{ item.label }}</span>
        </RouterLink>
      </nav>
    </aside>

    <section class="profile-shell__content">
      <RouterView v-slot="{ Component }">
        <Transition
          name="subroute-slide"
          mode="out-in"
        >
          <component
            :is="Component"
            class="subroute-transition-shell"
          />
        </Transition>
      </RouterView>
    </section>
  </main>
</template>

<style scoped>
.profile-shell {
  display: grid;
  width: 100%;
  max-width: var(--layout-page-max-width);
  gap: 14px;
  margin: 0 auto;
  padding: 18px var(--layout-page-padding-inline) 72px;
  color: rgb(var(--color-text));
}

.profile-shell__sidebar {
  display: grid;
  align-content: start;
  gap: 16px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 3px;
  background: rgb(var(--color-surface));
  box-shadow: none;
  padding: 14px;
}

.profile-shell__sidebar h1 {
  margin: 0;
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 24px;
  font-weight: 400;
  line-height: 1.2;
}

.profile-shell__sidebar p {
  margin: 8px 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  line-height: 1.65;
}

.profile-shell__nav {
  display: grid;
  gap: 6px;
}

.profile-shell__nav-item {
  position: relative;
  display: inline-flex;
  min-height: 36px;
  width: 100%;
  align-items: center;
  gap: 8px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  padding: 8px 10px;
  color: rgb(var(--color-text) / 0.78);
  font-size: 12px;
  font-weight: 600;
  line-height: 1;
  text-decoration: none;
  box-shadow: none;
  transition:
    border-color 0.2s ease,
    background 0.2s ease,
    color 0.2s ease;
}

.profile-shell__nav-item:hover {
  border-color: rgb(var(--color-primary));
  color: rgb(var(--color-text));
}

.profile-shell__nav-item--active,
.profile-shell__nav-item--active:hover {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
  box-shadow: none;
  text-shadow: none;
}

.profile-shell__nav-item::after {
  position: absolute;
  left: 0.85rem;
  right: 0.85rem;
  bottom: 0.42rem;
  height: 1.5px;
  border-radius: 2px;
  background: currentColor;
  transform: scaleX(0);
  transform-origin: left center;
  opacity: 0;
  transition:
    transform 0.22s ease,
    opacity 0.22s ease;
  content: '';
}

.profile-shell__nav-item--active::after {
  transform: scaleX(1);
  opacity: 1;
}

.profile-shell__content {
  min-width: 0;
}

@media (min-width: 1024px) {
  .profile-shell {
    grid-template-columns: 15.75rem minmax(0, 1fr);
    align-items: start;
  }

  .profile-shell__sidebar {
    position: sticky;
    top: 66px;
  }
}

@media (max-width: 767px) {
  .profile-shell {
    padding: 18px var(--layout-page-padding-inline) 96px;
  }
}

.profile-shell {
  max-width: min(1240px, calc(100vw - 32px));
  gap: 18px;
  padding: 10px 0 72px;
}

.profile-shell__sidebar {
  gap: 18px;
  border-radius: 8px;
  padding: 18px;
}

.profile-shell__heading {
  display: grid;
  gap: 8px;
}

.profile-shell__sidebar h1 {
  font-size: 30px;
  line-height: 1.15;
}

.profile-shell__sidebar p {
  margin: 0;
  font-size: 13px;
}

.profile-shell__role-card {
  display: grid;
  gap: 6px;
  border-left: 3px solid rgb(var(--color-primary));
  background: rgb(var(--color-surface-raised));
  padding: 12px 14px;
}

.profile-shell__role-label,
.profile-shell__nav-heading {
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  font-weight: 700;
}

.profile-shell__role-value {
  color: rgb(var(--color-text));
  font-size: 17px;
  font-weight: 700;
  line-height: 1.25;
}

.profile-shell__role-text {
  font-size: 12px;
}

.profile-shell__nav {
  gap: 8px;
}

.profile-shell__nav-item {
  min-height: 42px;
  gap: 10px;
  border-radius: 6px;
  padding: 10px 14px;
  color: rgb(var(--color-text) / 0.82);
  font-size: 14px;
  font-weight: 700;
}

.profile-shell__nav-item:hover {
  border-color: rgb(var(--color-primary) / 0.45);
}

.profile-shell__nav-item--active,
.profile-shell__nav-item--active:hover {
  background: rgb(var(--color-primary) / 0.1);
  color: rgb(var(--color-primary));
}

.profile-shell__nav-item::after,
.profile-shell__nav-item--active::after {
  display: none;
}

.profile-shell__content {
  display: grid;
  gap: 16px;
}

@media (min-width: 1024px) {
  .profile-shell {
    grid-template-columns: 260px minmax(0, 1fr);
  }

  .profile-shell__sidebar {
    top: 72px;
  }
}

@media (max-width: 1023px) {
  .profile-shell {
    max-width: 100%;
    padding-bottom: calc(var(--app-mobile-content-bottom) + 16px);
  }

  .profile-shell__nav {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 767px) {
  .profile-shell__nav {
    grid-template-columns: 1fr;
  }
}
</style>
