<!--
 * AJO Pay 支付中心入口頁。
 * 1. 頂部水平子導航：概覽、賬單、歷史、購物車、訂單、會計、單位。
 * 2. 內容區承載子路由概覽卡片與各支付子頁。
 * 3. 對齊 HTML 設計稿 pay-subnav 樣式，並映射至專案 token 體系。
 * 4. 響應式：桌面水平 Tab、行動裝置橫向滾動。
-->
<script setup lang="ts">
import { computed } from 'vue';
import { RouterView, useRoute, useRouter } from 'vue-router';

interface PayNavItem {
  key: string;
  label: string;
  to: string;
}

// 1. 子導航項目（靜態 mock）
const navItems: PayNavItem[] = [
  { key: 'overview', label: '概覽', to: '/payments' },
  { key: 'bills', label: '賬單', to: '/payments/bills' },
  { key: 'history', label: '歷史', to: '/payments/history' },
  { key: 'cart', label: '購物車', to: '/payments/cart' },
  { key: 'orders', label: '訂單', to: '/payments/orders' },
  { key: 'accounting', label: '會計', to: '/payments/accounting' },
  { key: 'units', label: '單位', to: '/payments/units' },
];

const route = useRoute();
const router = useRouter();

// 2. 判斷當前 Tab 是否啟用
const isNavActive = (item: PayNavItem): boolean =>
  route.path === item.to || route.path.startsWith(`${item.to}/`);

// 3. 點擊 Tab 跳轉
const handleNavClick = (item: PayNavItem): void => {
  router.push(item.to);
};

// 4. 當前啟用的 Tab 標籤（用於行動裝置輔助提示）
const activeLabel = computed(() => {
  const active = navItems.find(isNavActive);
  return active?.label ?? '概覽';
});
</script>

<template>
  <main class="ajo-pay">
    <!-- 1. 頂部子導航 -->
    <nav
      class="ajo-pay__subnav"
      aria-label="AJO Pay 子導航"
    >
      <div class="ajo-pay__brand">
        <span class="ajo-pay__brand-ajo">AJO</span>
        <span class="ajo-pay__brand-pay">PAY</span>
      </div>
      <span
        class="ajo-pay__divider"
        aria-hidden="true"
      ></span>
      <div
        class="ajo-pay__tabs"
        role="tablist"
      >
        <button
          v-for="item in navItems"
          :key="item.key"
          type="button"
          role="tab"
          class="ajo-pay__tab"
          :class="isNavActive(item) ? 'ajo-pay__tab--active' : ''"
          :aria-selected="isNavActive(item) ? 'true' : 'false'"
          @click="handleNavClick(item)"
        >
          {{ item.label }}
        </button>
      </div>
    </nav>

    <!-- 2. 內容區：承載子路由 -->
    <section class="ajo-pay__content">
      <RouterView v-slot="{ Component, route: activeRoute }">
        <Transition
          name="ajo-pay-fade"
          mode="out-in"
        >
          <component
            :is="Component"
            :key="activeRoute.fullPath"
          />
        </Transition>
      </RouterView>
    </section>

    <!-- 3. 行動裝置當前 Tab 提示 -->
    <p class="ajo-pay__mobile-hint">
      當前：<strong>{{ activeLabel }}</strong>
    </p>
  </main>
</template>

<style scoped>
/* 1. 主容器 */
.ajo-pay {
  display: flex;
  flex-direction: column;
  width: 100%;
  max-width: 1320px;
  min-height: calc(100vh - var(--nav-h, 52px));
  margin: 0 auto;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
}

/* 2. 頂部子導航 */
.ajo-pay__subnav {
  display: flex;
  align-items: center;
  gap: 0;
  background: rgb(var(--color-surface));
  border-bottom: 1px solid rgb(var(--color-border));
  padding: 0 1.5rem;
  height: 48px;
  position: sticky;
  top: var(--nav-h, 52px);
  z-index: 20;
}

/* 3. 品牌標識 */
.ajo-pay__brand {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-right: 1.5rem;
  flex-shrink: 0;
}

.ajo-pay__brand-ajo {
  font-size: 15px;
  font-weight: 800;
  letter-spacing: -0.5px;
  color: rgb(var(--color-text));
  font-family: var(--font-sans);
}

.ajo-pay__brand-pay {
  font-size: 15px;
  font-weight: 400;
  letter-spacing: 2px;
  color: rgb(var(--color-primary));
  font-family: var(--font-sans);
}

/* 4. 分隔線 */
.ajo-pay__divider {
  width: 1px;
  height: 20px;
  background: rgb(var(--color-border));
  margin-right: 1.5rem;
  flex-shrink: 0;
}

/* 5. Tab 列表 */
.ajo-pay__tabs {
  display: flex;
  gap: 0;
  flex: 1;
  overflow-x: auto;
  scrollbar-width: none;
}

.ajo-pay__tabs::-webkit-scrollbar {
  display: none;
}

/* 6. 單個 Tab */
.ajo-pay__tab {
  font-size: 13px;
  color: rgb(var(--color-ink-3));
  padding: 0 1rem;
  height: 48px;
  display: flex;
  align-items: center;
  cursor: pointer;
  border: none;
  border-bottom: 2px solid transparent;
  white-space: nowrap;
  transition: color 0.15s ease, border-color 0.15s ease;
  font-family: inherit;
  background: none;
  font-weight: 500;
}

.ajo-pay__tab:hover {
  color: rgb(var(--color-ink-2));
}

.ajo-pay__tab--active,
.ajo-pay__tab--active:hover {
  color: rgb(var(--color-primary));
  border-bottom-color: rgb(var(--color-primary));
  font-weight: 600;
}

/* 7. 內容區 */
.ajo-pay__content {
  flex: 1;
  min-width: 0;
  padding: 1.5rem;
}

/* 8. 行動裝置提示（預設隱藏） */
.ajo-pay__mobile-hint {
  display: none;
  margin: 0;
  padding: 0.5rem 1rem 1rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 600;
  line-height: 1.4;
}

.ajo-pay__mobile-hint strong {
  color: rgb(var(--color-primary));
  font-weight: 700;
}

/* 9. 路由切換過渡 */
.ajo-pay-fade-enter-active,
.ajo-pay-fade-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.ajo-pay-fade-enter-from {
  opacity: 0;
  transform: translateY(4px);
}

.ajo-pay-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

/* 10. 響應式：行動裝置 */
@media (max-width: 767px) {
  .ajo-pay__subnav {
    position: static;
    padding: 0 1rem;
    height: 44px;
  }

  .ajo-pay__brand {
    margin-right: 0.75rem;
  }

  .ajo-pay__brand-ajo,
  .ajo-pay__brand-pay {
    font-size: 13px;
  }

  .ajo-pay__divider {
    margin-right: 0.75rem;
  }

  .ajo-pay__tab {
    padding: 0 0.75rem;
    font-size: 12px;
    height: 44px;
  }

  .ajo-pay__content {
    padding: 1rem;
  }

  .ajo-pay__mobile-hint {
    display: block;
  }
}

/* 11. 響應式：平板 */
@media (min-width: 768px) and (max-width: 1023px) {
  .ajo-pay__subnav {
    padding: 0 1.25rem;
  }

  .ajo-pay__content {
    padding: 1.25rem;
  }
}
</style>
