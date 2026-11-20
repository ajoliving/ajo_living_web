<!--
 * 首頁工具列。
 * 1. 以 URL 選擇公開參考屋苑，不變更會員綁定。
 * 2. 沿用超市商品搜尋及通知中心路由。
-->
<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { RouterLink, useRoute, useRouter } from 'vue-router';

import AppIcon from '@/shared/components/base/AppIcon.vue';

import { HOME_ESTATES, resolveHomeEstate } from '../home_estate';

const props = defineProps<{ buildingId?: string }>();
const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const search = ref('');
const selectedEstate = computed(() => resolveHomeEstate(route.query.estate, props.buildingId) ?? '');

// 1. 保留其他查詢條件，只更新首頁的參考選擇。
const selectEstate = (event: Event): void => {
  const value = (event.target as HTMLSelectElement).value;
  const estate = HOME_ESTATES.find((item) => item === value);
  if (!estate && !props.buildingId) return;
  const query = { ...route.query, estate };
  void router.replace({ path: route.path, query, hash: route.hash });
};

// 2. 提交到既有超市搜尋，空白關鍵詞保留在首頁。
const submitSearch = (): void => {
  const keyword = search.value.trim();
  if (!keyword) return;
  void router.push({ path: '/supermarket-offers', query: { q: keyword } });
};
</script>

<template>
  <div class="home-toolbar" role="group" :aria-label="t('home.community.toolbar.label')">
    <div class="home-toolbar-estate">
      <AppIcon name="location" :size="18" />
      <select
        :value="selectedEstate"
        :aria-label="t('home.community.toolbar.estate')"
        @change="selectEstate"
      >
        <option v-if="buildingId" value="">{{ t('home.community.toolbar.current') }}</option>
        <option v-for="estate in HOME_ESTATES" :key="estate" :value="estate">
          {{ t(`home.community.toolbar.estates.${estate}`) }}
        </option>
      </select>
      <AppIcon class="home-toolbar-chevron" name="chevron-down" :size="16" />
    </div>

    <form class="home-toolbar-search" role="search" @submit.prevent="submitSearch">
      <input
        v-model="search"
        type="search"
        maxlength="100"
        :placeholder="t('home.community.toolbar.search')"
        :aria-label="t('home.community.toolbar.search')"
      />
      <button
        class="home-toolbar-icon"
        type="submit"
        :title="t('home.community.toolbar.search')"
        :aria-label="t('home.community.toolbar.search')"
      >
        <AppIcon name="search" :size="19" />
      </button>
    </form>

    <RouterLink
      class="home-toolbar-icon home-toolbar-notification"
      to="/notifications"
      :title="t('nav.notifications')"
      :aria-label="t('nav.notifications')"
    >
      <AppIcon name="bell" :size="20" />
    </RouterLink>
  </div>
</template>

<style scoped lang="scss">
/* 1. 同列工具在窄屏保留 44px 操作目標。 */
.home-toolbar {
  display: grid;
  grid-template-columns: minmax(0, 176px) minmax(0, 304px) 44px;
  align-items: center;
  gap: 12px;
  min-height: 64px;
  padding-block: 8px;
}

.home-toolbar-estate,
.home-toolbar-search,
.home-toolbar-notification {
  display: flex;
  align-items: center;
  min-width: 0;
  min-height: 44px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  color: rgb(var(--color-text));
  background: rgb(var(--color-surface));
}

.home-toolbar-estate {
  position: relative;
}

.home-toolbar-estate > .app-icon {
  position: absolute;
  left: 10px;
  color: rgb(var(--color-primary));
  pointer-events: none;
}

.home-toolbar-estate > .home-toolbar-chevron {
  right: 8px;
  left: auto;
  color: rgb(var(--color-text-muted));
}

.home-toolbar-estate select,
.home-toolbar-search input {
  width: 100%;
  min-width: 0;
  min-height: 42px;
  color: inherit;
  font-size: 14px;
  background: transparent;
}

.home-toolbar-estate select {
  appearance: none;
  cursor: pointer;
  padding-inline: 34px 28px;
  font-weight: 600;
  text-overflow: ellipsis;
}

.home-toolbar-estate option {
  background: rgb(var(--color-surface));
}

.home-toolbar-search input {
  padding-inline: 10px 2px;
}

.home-toolbar-search input::placeholder {
  color: rgb(var(--color-text-muted));
  opacity: 1;
}

.home-toolbar-icon {
  display: grid;
  width: 44px;
  height: 44px;
  flex: 0 0 44px;
  place-items: center;
  cursor: pointer;
}

.home-toolbar-icon:hover {
  color: rgb(var(--color-primary));
}

.home-toolbar :is(input, select):focus-visible {
  outline: 2px solid rgb(var(--color-primary));
  outline-offset: -2px;
}

@media (max-width: 600px) {
  .home-toolbar {
    grid-template-columns: minmax(0, 1fr) minmax(0, 1.2fr) 44px;
    gap: 6px;
  }

  .home-toolbar-estate select {
    padding-inline: 28px 22px;
    font-size: 13px;
  }

  .home-toolbar-estate > .app-icon {
    left: 7px;
  }

  .home-toolbar-estate > .home-toolbar-chevron {
    right: 4px;
    left: auto;
  }

  .home-toolbar-search input {
    padding-inline-start: 8px;
    font-size: 13px;
  }
}

@media (max-width: 400px) {
  .home-toolbar {
    grid-template-columns: minmax(0, 1fr) 44px;
  }

  .home-toolbar-estate {
    grid-column: 1 / -1;
  }
}
</style>
