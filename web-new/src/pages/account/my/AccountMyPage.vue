<!--
 * 會員中心外框頁。
 * 1. 承載會員中心左側導航與各功能頁。
 * 2. 保持 Staff 設定與管理頁的獨立版面。
-->
<script setup lang="ts">
import { computed } from 'vue';
import { RouterView, useRoute } from 'vue-router';

import AccountSidebar from './widgets/AccountSidebar.vue';

const route = useRoute();

// 1. 判斷是否使用會員中心左側導航版面
const usesAccountShell = computed(() =>
  !route.path.startsWith('/account/marketplace/settings') &&
  !route.path.startsWith('/account/marketplace/management'),
);
</script>

<template>
  <main
    class="account-my-page"
    :class="{ 'account-my-page--shell': usesAccountShell }"
  >
    <template v-if="usesAccountShell">
      <AccountSidebar />
      <section class="account-my-page__content">
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
      </section>
    </template>

    <RouterView
      v-else
      v-slot="{ Component }"
    >
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

.account-my-page--shell {
  display: flex;
  width: min(100%, var(--layout-page-max-width));
  gap: 14px;
  margin: 0 auto;
  padding: 18px var(--layout-page-padding-inline) 72px;
}

.account-my-page__content {
  min-width: 0;
  flex: 1;
}

@media (max-width: 767px) {
  .account-my-page {
    min-height: calc(100vh - var(--app-header-offset, 0rem));
  }

  .account-my-page--shell {
    flex-direction: column;
    padding: 18px var(--layout-page-padding-inline) 96px;
  }
}
</style>
