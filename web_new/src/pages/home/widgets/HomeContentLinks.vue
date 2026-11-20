<!--
 * 首頁樓市及社區會話入口。
 * 1. 導向現有租金走勢和通知中心的會話檢視。
-->
<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { RouterLink } from 'vue-router';

import AppIcon from '@/shared/components/base/AppIcon.vue';

const { t } = useI18n();
</script>

<template>
  <div class="home-content-links">
    <section aria-labelledby="home-market-title">
      <div class="home-section-heading">
        <h2 id="home-market-title">{{ t('home.community.links.market') }}</h2>
      </div>
      <RouterLink to="/trend">
        <AppIcon name="browse" :size="20" />
        <span>{{ t('home.community.links.rent') }}</span>
        <AppIcon name="arrow-right" :size="16" />
      </RouterLink>
    </section>
    <section aria-labelledby="home-discussion-title">
      <div class="home-section-heading">
        <h2 id="home-discussion-title">{{ t('home.community.links.discussion') }}</h2>
      </div>
      <RouterLink :to="{ path: '/notifications', query: { tab: 'conversations' } }">
        <AppIcon name="message" :size="20" />
        <span>{{ t('home.community.links.conversations') }}</span>
        <AppIcon name="arrow-right" :size="16" />
      </RouterLink>
    </section>
  </div>
</template>

<style scoped lang="scss">
/* 1. 手機順序維持兩個獨立內容欄。 */
.home-content-links {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.home-content-links a {
  display: flex;
  min-height: 53px;
  align-items: center;
  gap: 10px;
  border-block: 1px solid rgb(var(--color-border));
  padding: 10px 12px;
  font-size: 14px;
}

.home-content-links a:hover {
  color: rgb(var(--color-primary));
}

.home-content-links a span {
  flex: 1;
  min-width: 0;
}

.home-content-links a svg {
  flex-shrink: 0;
  color: rgb(var(--color-primary));
}

@media (max-width: 600px) {
  .home-content-links {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
