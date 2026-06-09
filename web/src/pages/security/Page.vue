<!--
 * 安防入口頁。
 * 1. 承載 CCTV、門禁與對講入口。
 * 2. 為舊系統受控聚合保留清晰頁面邊界。
-->
<script setup lang="ts">
import { RouterLink } from 'vue-router';

const securityEntries = [
  {
    title: 'CCTV',
    description: '查看授權攝像頭與影像入口。',
    to: '/security/cctv',
  },
  {
    title: '門禁',
    description: '查看門禁、智能門鎖與出入記錄入口。',
    to: '/security/access-control',
  },
  {
    title: '對講',
    description: '查看訪客對講與通行相關入口。',
    to: '/security/intercom',
  },
];
</script>

<template>
  <main class="section-shell py-12 lg:py-16">
    <section class="mx-auto max-w-5xl">
      <p class="text-xs font-semibold uppercase tracking-[0.18em] text-primary">
        Security
      </p>
      <h1 class="mt-3 font-display text-4xl font-semibold text-text">
        安防中心
      </h1>
      <p class="mt-4 max-w-2xl text-sm leading-7 text-text-muted">
        統一承載 iCCTV、門禁、智能門鎖與對講能力的入口。
      </p>
    </section>

    <section class="mx-auto mt-8 grid max-w-5xl gap-4 md:grid-cols-3">
      <RouterLink
        v-for="entry in securityEntries"
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
