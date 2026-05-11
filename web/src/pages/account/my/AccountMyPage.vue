<!--
 * 會員中心外框頁。
 * 1. 承載會員中心頂部子導航下的各子頁。
 * 2. 保持帳戶模組全寬內容容器。
-->
<script setup lang="ts">
import { RouterView } from 'vue-router';
</script>

<template>
  <main class="account-my-page">
    <RouterView v-slot="{ Component }">
      <Transition
        name="subroute-slide"
        mode="out-in"
      >
        <component
          :is="Component"
          class="subroute-transition-shell"
        />
      </Transition>
    </RouterView>
  </main>
</template>

<style scoped>
.account-my-page {
  width: 100%;
  min-height: calc(100vh - var(--app-header-offset, 0rem));
  color: rgb(var(--color-text));
}

@media (max-width: 767px) {
  .account-my-page {
    min-height: calc(100vh - var(--app-header-offset, 0rem));
  }
}
</style>
