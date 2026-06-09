<!--
 * 通訊入口頁。
 * 1. 承載公告、通知與站內信的模組入口。
 * 2. 保持平台通訊能力獨立於 marketplace。
-->
<script setup lang="ts">
import { RouterLink } from 'vue-router';

const communicationEntries = [
  {
    title: '公告通知',
    description: '集中查看 iBoard 公告、系統通知與物業提示。',
    to: '/communications/notices',
  },
  {
    title: '站內信',
    description: '進入會員站內訊息與交易會話。',
    to: '/communications/messages',
  },
];
</script>

<template>
  <main class="section-shell py-12 lg:py-16">
    <section class="mx-auto max-w-5xl">
      <p class="text-xs font-semibold uppercase tracking-[0.18em] text-primary">
        Communications
      </p>
      <h1 class="mt-3 font-display text-4xl font-semibold text-text">
        通訊中心
      </h1>
      <p class="mt-4 max-w-2xl text-sm leading-7 text-text-muted">
        統一承載 AJO Living 的站內信、公告、通知與日後 iBoard 入口。
      </p>
    </section>

    <section class="mx-auto mt-8 grid max-w-5xl gap-4 md:grid-cols-2">
      <RouterLink
        v-for="entry in communicationEntries"
        :key="entry.to"
        :to="entry.to"
        class="app-card p-6 transition hover:-translate-y-0.5"
      >
        <h2 class="text-lg font-semibold text-text">
          {{ entry.title }}
        </h2>
        <p class="mt-3 text-sm leading-6 text-text-muted">
          {{ entry.description }}
        </p>
      </RouterLink>
    </section>
  </main>
</template>
