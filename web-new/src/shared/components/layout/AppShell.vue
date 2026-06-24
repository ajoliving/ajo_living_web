<!--
 * 全域應用程式外框。
 * 1. 承載頂部導航、主要頁面內容與頁腳。
 * 2. 提供接近參考稿的白底、細線與緊湊版面節奏。
 * 3. 保留全域提示視窗。
-->
<script setup lang="ts">
import { computed } from 'vue';
import { RouterLink, RouterView, useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';

import ToastViewport from '@/shared/components/feedback/ToastViewport.vue';
import AppHeader from '@/shared/components/navigation/AppHeader.vue';

interface FooterLink {
  key: string;
  label: string;
  to?: string;
}

interface FooterColumn {
  key: string;
  title: string;
  links: FooterLink[];
}

const route = useRoute();
const { t } = useI18n();

// 1. 判斷目前路由是否隱藏頁腳
const isFooterHiddenRoute = computed(() => route.name === 'Login');

// 2. 組合頁腳欄目，沒有實際頁面的項目保持不可跳轉，避免死鏈
const footerColumns = computed<FooterColumn[]>(() => [
  {
    key: 'services',
    title: t('common.footer.columns.services.title'),
    links: [
      { key: 'properties', label: t('common.footer.columns.services.links.properties'), to: '/properties' },
      { key: 'servicedResidences', label: t('common.footer.columns.services.links.servicedResidences'), to: '/serviced-residences' },
      { key: 'furniture', label: t('common.footer.columns.services.links.furniture'), to: '/furniture' },
      { key: 'supermarketOffers', label: t('common.footer.columns.services.links.supermarketOffers'), to: '/supermarket-offers' },
      { key: 'rentTrends', label: t('common.footer.columns.services.links.rentTrends') },
    ],
  },
  {
    key: 'support',
    title: t('common.footer.columns.support.title'),
    links: [
      { key: 'faq', label: t('common.footer.columns.support.links.faq') },
      { key: 'contact', label: t('common.footer.columns.support.links.contact') },
      { key: 'payments', label: t('common.footer.columns.support.links.payments'), to: '/payments' },
      { key: 'agentRegistration', label: t('common.footer.columns.support.links.agentRegistration') },
      { key: 'advertising', label: t('common.footer.columns.support.links.advertising') },
    ],
  },
  {
    key: 'about',
    title: t('common.footer.columns.about.title'),
    links: [
      { key: 'aboutAjo', label: t('common.footer.columns.about.links.aboutAjo') },
      { key: 'press', label: t('common.footer.columns.about.links.press') },
      { key: 'careers', label: t('common.footer.columns.about.links.careers') },
      { key: 'privacy', label: t('common.footer.columns.about.links.privacy') },
      { key: 'terms', label: t('common.footer.columns.about.links.terms') },
    ],
  },
]);

const footerSocials = computed(() => [
  { key: 'facebook', label: t('common.footer.socials.facebook'), text: 'f' },
  { key: 'camera', label: t('common.footer.socials.camera'), text: '📸' },
  { key: 'chat', label: t('common.footer.socials.chat'), text: '💬' },
  { key: 'video', label: t('common.footer.socials.video'), text: '▶' },
]);
</script>

<template>
  <div class="app-shell min-h-screen bg-canvas text-text">
    <AppHeader />

    <main class="app-main">
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
      class="app-footer"
    >
      <div class="app-footer__inner">
        <div class="app-footer__grid">
          <section class="app-footer__intro">
            <p class="app-footer__brand">
              AJO LIVING
            </p>
            <p class="app-footer__tagline">
              {{ t('common.footer.taglineLong') }}
            </p>
            <div class="app-footer__socials">
              <button
                v-for="social in footerSocials"
                :key="social.key"
                class="app-footer__social"
                type="button"
                :aria-label="social.label"
              >
                {{ social.text }}
              </button>
            </div>
          </section>

          <section
            v-for="column in footerColumns"
            :key="column.key"
            class="app-footer__column"
          >
            <h2>{{ column.title }}</h2>
            <ul>
              <li
                v-for="link in column.links"
                :key="link.key"
              >
                <RouterLink
                  v-if="link.to"
                  :to="link.to"
                  class="app-footer__link"
                >
                  {{ link.label }}
                </RouterLink>
                <span
                  v-else
                  class="app-footer__link app-footer__link--muted"
                >
                  {{ link.label }}
                </span>
              </li>
            </ul>
          </section>
        </div>

        <div class="app-footer__bottom">
          <span>{{ t('common.footer.copyright') }}</span>
        </div>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.app-shell {
  min-width: 320px;
}

.app-main {
  min-height: calc(100vh - var(--nav-h));
  padding-top: var(--nav-h);
}

.app-footer {
  background: #1a1a1a;
  color: rgb(255 255 255 / 0.7);
  padding: 40px 32px 24px;
}

.app-footer__inner {
  width: min(100%, var(--layout-page-max-width));
  margin: 0 auto;
}

.app-footer__grid {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 1fr;
  gap: 32px;
  margin-bottom: 32px;
}

.app-footer__brand {
  display: inline-block;
  border-bottom: 2px solid rgb(var(--color-primary));
  margin: 0 0 10px;
  padding-bottom: 2px;
  color: #ffffff;
  font-family: var(--font-display);
  font-size: 16px;
  letter-spacing: 2px;
}

.app-footer__tagline {
  max-width: 220px;
  margin: 0;
  font-size: 11px;
  line-height: 1.7;
}

.app-footer__socials {
  display: flex;
  gap: 8px;
  margin-top: 18px;
}

.app-footer__social {
  display: inline-flex;
  width: 28px;
  height: 28px;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(255 255 255 / 0.18);
  border-radius: 999px;
  background: rgb(255 255 255 / 0.06);
  color: #ffffff;
  cursor: pointer;
  font-size: 12px;
  line-height: 1;
  transition:
    border-color 0.15s ease,
    background 0.15s ease,
    color 0.15s ease;
}

.app-footer__social:hover {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
}

.app-footer__column h2 {
  margin: 0 0 12px;
  color: #ffffff;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.8px;
}

.app-footer__column ul {
  display: grid;
  gap: 8px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.app-footer__link {
  color: rgb(255 255 255 / 0.62);
  font-size: 11px;
  line-height: 1.5;
  transition: color 0.15s ease;
}

.app-footer__link:hover {
  color: #ffffff;
}

.app-footer__link--muted {
  cursor: default;
}

.app-footer__bottom {
  border-top: 1px solid rgb(255 255 255 / 0.1);
  padding-top: 18px;
  color: rgb(255 255 255 / 0.45);
  font-size: 10px;
  letter-spacing: 0.6px;
}

.page-fade-enter-active,
.page-fade-leave-active {
  transition:
    opacity 0.18s ease,
    transform 0.18s ease;
}

.page-fade-enter-from,
.page-fade-leave-to {
  opacity: 0;
  transform: translateY(6px);
}

@media (max-width: 1023px) {
  .app-footer__grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 767px) {
  .app-footer {
    padding: 32px 20px 92px;
  }

  .app-footer__grid {
    grid-template-columns: 1fr;
    gap: 26px;
  }
}
</style>
