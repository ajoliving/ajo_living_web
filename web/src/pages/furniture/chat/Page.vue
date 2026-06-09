<!--
 * 二手交易會話轉接頁。
 * 1. 保留二手傢俬會話轉接頁面邊界。
 * 2. 將實際站內信能力交由 communications/messages 承載。
-->
<script setup lang="ts">
import { onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';

const route = useRoute();
const router = useRouter();

// 1. 進入平台通訊中心
onMounted(() => {
  void router.replace({
    path: '/communications/messages',
    query: route.query,
  });
});
</script>

<template>
  <main class="section-shell py-12 lg:py-16">
    <section class="mx-auto max-w-3xl">
      <p class="text-xs font-semibold uppercase tracking-[0.18em] text-primary">
        Messages
      </p>
      <h1 class="mt-3 font-display text-3xl font-semibold text-text">
        站內信
      </h1>
      <p class="mt-4 text-sm leading-7 text-text-muted">
        交易會話已納入 AJO Living 通訊中心。
      </p>
    </section>
  </main>
</template>
