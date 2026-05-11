<!--
 * 全域應用程式外框。
 * 1. 承載頂部導航、頁面切換與頁腳。
 * 2. 統一消費者 Web 原型的背景與版面節奏。
 * 3. 提供全域提示視窗。
-->
<script setup lang="ts">
import { computed } from 'vue';
import { RouterView, useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';

import ToastViewport from '@/shared/components/feedback/ToastViewport.vue';
import AppHeader from '@/shared/components/navigation/AppHeader.vue';
import { useHeaderSubnav } from '@/shared/navigation/useHeaderSubnav';

const { t } = useI18n();
const route = useRoute();
const { subnavItems } = useHeaderSubnav();

// 1. 判斷目前路由是否使用透明覆蓋導航
const isTransparentHeaderRoute = computed(() => route.meta.headerMode === 'transparent');

// 2. 判斷目前路由是否隱藏頁腳
const isFooterHiddenRoute = computed(() => route.name === 'Login');

// 3. 判斷目前桌面版是否會顯示第二層導航
const hasSubNavigation = computed(() => subnavItems.value.length > 0);

// 4. 組合頁面主內容的導航預留空間狀態
const mainClasses = computed(() => ({
  'app-main--overlay': isTransparentHeaderRoute.value,
  'app-main--solid': !isTransparentHeaderRoute.value,
  'app-main--fullscreen': isFooterHiddenRoute.value,
  'app-main--with-subnav': !isTransparentHeaderRoute.value && hasSubNavigation.value,
  'app-main--with-footer': !isFooterHiddenRoute.value,
}));
</script>

<template>
  <div class="min-h-screen bg-canvas text-text">
    <AppHeader />

    <main
      class="app-main relative"
      :class="mainClasses"
    >
      <div
        v-if="!isFooterHiddenRoute"
        class="pointer-events-none absolute inset-x-0 top-0 h-[28rem] bg-shell-glow opacity-90"
      />
      <div
        v-if="!isFooterHiddenRoute"
        class="pointer-events-none absolute inset-x-0 top-16 h-[32rem] soft-grid opacity-[0.18]"
      />

      <RouterView v-slot="{ Component }">
        <transition
          name="page-fade"
          mode="out-in"
        >
          <component :is="Component" />
        </transition>
      </RouterView>
    </main>

    <ToastViewport />

    <footer
      v-if="!isFooterHiddenRoute"
      class="border-t border-border/60 bg-topbar-surface/82 py-8 backdrop-blur"
    >
      <div class="section-shell flex flex-col gap-4 text-sm text-text-muted md:flex-row md:items-center md:justify-between">
        <div>
          <p class="font-semibold text-text">
            {{ t('common.brand.name') }}
          </p>
          <p>{{ t('common.brand.tagline') }}</p>
        </div>
        <p>{{ t('common.footer.copyright') }}</p>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.app-main {
  --app-header-offset: 0rem;
  box-sizing: border-box;
  min-height: 100vh;
}

.app-main--fullscreen {
  height: 100svh;
  min-height: 100svh;
  overflow: hidden;
}

.app-main--solid {
  --app-header-offset: 3.725rem;
  padding-top: var(--app-header-offset);
}

.app-main--with-footer {
  padding-bottom: 4rem;
}

.page-fade-enter-active,
.page-fade-leave-active {
  transition:
    opacity 0.24s ease,
    transform 0.24s ease;
}

.page-fade-enter-from,
.page-fade-leave-to {
  opacity: 0;
  transform: translateY(12px);
}

@media (min-width: 1024px) {
  .app-main--solid.app-main--with-subnav {
    --app-header-offset: 6.375rem;
  }
}
</style>
