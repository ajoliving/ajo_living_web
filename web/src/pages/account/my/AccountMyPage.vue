<!--
 * 會員中心布局頁。
 * 1. 提供會員中心側欄與內容區雙欄布局。
 * 2. 承載會員中心各子頁的 RouterView。
-->
<script setup lang="ts">
import { RouterView } from 'vue-router';

import AccountSidebar from './widgets/AccountSidebar.vue';
</script>

<template>
  <main class="account-my-page">
    <AccountSidebar />
    <section class="account-my-content">
      <RouterView />
    </section>
  </main>
</template>

<style scoped>
.account-my-page {
  display: flex;
  width: 100%;
  max-width: 1280px;
  height: calc(100vh - var(--app-header-offset, 0rem));
  gap: 3rem;
  margin: 0 auto;
  padding: 4rem 2rem;
  color: #1a1c1b;
  overflow: hidden;
}

.account-my-content {
  min-width: 0;
  height: 100%;
  flex: 1;
  overflow-y: auto;
  overscroll-behavior: contain;
  padding-right: 0.5rem;
  scrollbar-gutter: stable;
}

@media (max-width: 767px) {
  .account-my-page {
    height: auto;
    min-height: calc(100vh - var(--app-header-offset, 0rem));
    flex-direction: column;
    gap: 2rem;
    overflow: visible;
    padding: 2.5rem 1.25rem;
  }

  .account-my-content {
    height: auto;
    overflow: visible;
    padding-right: 0;
  }
}
</style>
