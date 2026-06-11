<!--
 * 服務住宅頁。
 * 1. 承接服務住宅主路由入口。
 * 2. 以正式訊息提示目前頻道狀態。
-->
<script setup lang="ts">
import { RouterLink } from 'vue-router';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
</script>

<template>
  <section class="section-shell py-16 lg:py-20">
    <div class="app-card mx-auto max-w-3xl p-8 text-center md:p-10">
      <p class="text-xs font-semibold uppercase tracking-[0.22em] text-primary">
        {{ t('nav.channelPreparing') }}
      </p>
      <h1 class="mt-4 font-display text-4xl font-semibold text-text md:text-5xl">
        {{ t('nav.servicedResidences') }}
      </h1>
      <p class="mx-auto mt-4 max-w-2xl text-base leading-7 text-text-muted">
        {{ t('nav.servicedResidencesDescription') }}
      </p>
      <RouterLink
        to="/marketplace"
        class="mt-8 inline-flex rounded-pill bg-primary px-5 py-3 text-sm font-semibold text-primary-contrast shadow-soft transition hover:-translate-y-0.5"
      >
        {{ t('nav.goMarketplace') }}
      </RouterLink>
    </div>
  </section>
</template>
