<!--
 * 會員中心樓盤住宅入口頁。
 * 1. 提供我的樓盤與我的服務式住宅入口。
 * 2. 讓會員中心左側導航承接既有樓盤功能。
-->
<script setup lang="ts">
import { RouterLink } from 'vue-router';
import { useI18n } from 'vue-i18n';

import AppIcon from '@/shared/components/base/AppIcon.vue';

const { t } = useI18n();
</script>

<template>
  <main class="account-property-page">
    <header class="account-property-header">
      <p>AJO Living</p>
      <h1>{{ t('nav.propertiesHome') }}</h1>
    </header>

    <section class="account-property-grid">
      <RouterLink
        to="/account/properties/sale"
        class="account-property-card"
      >
        <span>
          <AppIcon
            name="home"
            :size="20"
          />
        </span>
        <div>
          <h2>{{ t('property.sale.myTitle') }}</h2>
          <p>{{ t('property.mine.manageHint') }}</p>
        </div>
      </RouterLink>

      <RouterLink
        to="/account/properties/serviced-residences"
        class="account-property-card"
      >
        <span>
          <AppIcon
            name="building"
            :size="20"
          />
        </span>
        <div>
          <h2>{{ t('property.serviced.myTitle') }}</h2>
          <p>{{ t('property.mine.manageHint') }}</p>
        </div>
      </RouterLink>
    </section>
  </main>
</template>

<style scoped>
.account-property-page {
  display: grid;
  gap: 14px;
  color: rgb(var(--color-text));
}

.account-property-header {
  display: grid;
  gap: 8px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 3px;
  background: rgb(var(--color-surface));
  padding: 14px;
}

.account-property-header p {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.14em;
  line-height: 1;
  text-transform: uppercase;
}

.account-property-header h1 {
  margin: 0;
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 24px;
  font-weight: 400;
  line-height: 1.2;
}

.account-property-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.account-property-card {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: 12px;
  align-items: start;
  border: 1px solid rgb(var(--color-border));
  border-radius: 3px;
  background: rgb(var(--color-surface));
  padding: 16px;
  color: inherit;
  text-decoration: none;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease;
}

.account-property-card:hover {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-surface-muted));
}

.account-property-card > span {
  display: inline-flex;
  width: 38px;
  height: 38px;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 3px;
  color: rgb(var(--color-primary));
}

.account-property-card h2 {
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 15px;
  font-weight: 700;
  line-height: 1.35;
}

.account-property-card p {
  margin: 6px 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  line-height: 1.6;
}

@media (max-width: 767px) {
  .account-property-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
