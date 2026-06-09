<!--
 * 管理端整合入口頁。
 * 1. 保留 AJO 管理端第三方與舊系統整合頁面邊界。
 * 2. 以正式入口呈現 POS、iBoard、iCCTV、門禁與對講等後續接入能力。
-->
<script setup lang="ts">
interface IntegrationEntry {
  title: string;
  description: string;
}

const integrationEntries: IntegrationEntry[] = [
  {
    title: '支付與物業費',
    description: '保留 POS、物業費與付款紀錄的管理接入邊界。',
  },
  {
    title: '公告與通知',
    description: '保留 iBoard、站內通知與大廈公告的管理接入邊界。',
  },
  {
    title: '安防與門禁',
    description: '保留 CCTV、智能門鎖、門禁與對講的管理接入邊界。',
  },
];
</script>

<template>
  <main class="section-shell py-12 lg:py-16">
    <section class="mx-auto max-w-5xl">
      <p class="text-xs font-semibold uppercase tracking-[0.18em] text-primary">
        Integrations
      </p>
      <h1 class="mt-3 font-display text-3xl font-semibold text-text">
        系統整合
      </h1>
      <p class="mt-4 max-w-2xl text-sm leading-7 text-text-muted">
        AJO 管理端統一保留舊系統與第三方服務的接入入口。
      </p>
    </section>

    <section class="mx-auto mt-8 grid max-w-5xl gap-4 md:grid-cols-3">
      <article
        v-for="entry in integrationEntries"
        :key="entry.title"
        class="app-card p-6"
      >
        <h2 class="text-base font-semibold text-text">
          {{ entry.title }}
        </h2>
        <p class="mt-3 text-sm leading-6 text-text-muted">
          {{ entry.description }}
        </p>
      </article>
    </section>
  </main>
</template>
