<!--
 * 我的物業發布頁。
 * 1. 根據頻道讀取我的樓盤或服務式住宅列表。
 * 2. 支援發布彈窗、重發、成交與下架等狀態操作。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onMounted, ref, watch } from 'vue';
import { RouterLink, useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';

import {
  deactivatePropertySale,
  deactivateServicedApartment,
  fetchMyPropertySaleListings,
  fetchMyServicedApartmentListings,
  markPropertySaleSold,
  publishPropertySale,
  publishServicedApartment,
  republishPropertySale,
  republishServicedApartment,
} from '@/domains/property/api';
import type {
  PropertyChannel,
  PropertyListParams,
  PropertyListingSummaryResponse,
} from '@/domains/property/model';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import { useFeedbackStore } from '@/app/stores/feedback';
import { usePreferenceStore } from '@/app/stores/preferences';
import { useSessionStore } from '@/app/stores/session';
import { formatDate } from '@/shared/utils/format';
import {
  resolvePropertyCommunityName,
  resolvePropertyCoverImage,
  resolvePropertyDetailPath,
  resolvePropertyPriceText,
  resolvePropertyStatus,
  resolvePropertySummary,
  resolvePropertyTitle,
  resolvePropertyTransactionType,
  resolvePropertyArea,
  resolvePropertyRooms,
} from '@/shared/utils/property';
import { formatAjoPoints, resolveWalletChargeCost } from '@/shared/utils/wallet';

import { demoPropertyListings } from '@/domains/property/demo';
import PropertyEditorPage from '@/pages/properties/editor/Page.vue';
import ServicedResidenceEditorPage from '@/pages/serviced-residences/editor/Page.vue';

const props = defineProps<{
  channel: PropertyChannel;
}>();

type MyPropertyTab = 'all' | 'draft' | 'active' | 'hidden' | 'expired' | 'sold';
type MyPropertyAction = 'publish' | 'republish' | 'mark-sold' | 'deactivate';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const feedbackStore = useFeedbackStore();
const preferenceStore = usePreferenceStore();
const sessionStore = useSessionStore();

const loading = ref(false);
const isEditorOpen = ref(false);
const editorListingId = ref('');
const activeTab = ref<MyPropertyTab>('all');
const searchQuery = ref('');
const items = ref<PropertyListingSummaryResponse[]>([]);

const pageTitle = computed(() =>
  props.channel === 'sale' ? t('property.sale.myTitle') : t('property.serviced.myTitle'),
);
const editorComponent = computed(() =>
  props.channel === 'sale' ? PropertyEditorPage : ServicedResidenceEditorPage,
);
const chargeCost = computed(() =>
  resolveWalletChargeCost(props.channel === 'sale' ? 'property_sale' : 'serviced_apartment'),
);
const editorKey = computed(() => `${props.channel}-${editorListingId.value || 'new'}-${isEditorOpen.value ? 'open' : 'closed'}`);
const formatPoints = (value: number): string =>
  formatAjoPoints(value, t('common.brand.pointsName'), preferenceStore.locale);
const tabOptions = computed<Array<{ label: string; value: MyPropertyTab }>>(() => [
  { label: t('property.mine.all'), value: 'all' },
  { label: t('property.mine.draft'), value: 'draft' },
  { label: t('property.mine.active'), value: 'active' },
  { label: t('property.mine.hidden'), value: 'hidden' },
  { label: t('property.mine.expired'), value: 'expired' },
  { label: t('property.mine.sold'), value: 'sold' },
]);
const filteredItems = computed(() => {
  const keyword = searchQuery.value.trim().toLowerCase();
  const sourceItems = items.value.length > 0 ? items.value : props.channel === 'sale' ? demoPropertyListings : items.value;
  if (!keyword) {
    return sourceItems;
  }

  return sourceItems.filter((listing) =>
    [
      resolvePropertyTitle(listing),
      resolvePropertySummary(listing),
      resolvePropertyCommunityName(listing),
    ]
      .join(' ')
      .toLowerCase()
      .includes(keyword),
  );
});
const dashboardMetrics = computed(() => [
  { label: t('property.mine.active'), value: String(filteredItems.value.filter((listing) => listing.publication_status === 'active').length) },
  { label: t('property.mine.draft'), value: String(filteredItems.value.filter((listing) => listing.publication_status === 'draft').length) },
  { label: t('property.mine.expired'), value: String(filteredItems.value.filter((listing) => listing.publication_status === 'expired').length) },
  { label: t('property.mine.sold'), value: String(filteredItems.value.filter((listing) => listing.business_status === 'sold').length) },
]);

// 1. 讀取我的發布
const loadMyListings = async (): Promise<void> => {
  loading.value = true;

  try {
    const params: PropertyListParams = {
      page: 1,
      page_size: 50,
      status: activeTab.value === 'all' ? '' : activeTab.value,
    };
    const response = props.channel === 'sale'
      ? await fetchMyPropertySaleListings(params)
      : await fetchMyServicedApartmentListings(params);

    items.value = response.data.data.items;
  } catch (error: unknown) {
    feedbackStore.pushToast(
      axios.isAxiosError<{ message?: string }>(error)
        ? error.response?.data?.message ?? t('property.mine.loadError')
        : t('property.mine.loadError'),
      'error',
    );
    items.value = props.channel === 'sale' ? demoPropertyListings : [];
  } finally {
    loading.value = false;
  }
};

// 2. 開啟發布彈窗
const openCreateEditor = (): void => {
  editorListingId.value = '';
  isEditorOpen.value = true;
};

// 3. 開啟編輯彈窗
const openEditEditor = (listingId: string): void => {
  editorListingId.value = listingId;
  isEditorOpen.value = true;
};

// 4. 關閉發布彈窗
const closeEditor = async (): Promise<void> => {
  isEditorOpen.value = false;
  editorListingId.value = '';
  if (route.query.propertyEditor) {
    const query = { ...route.query };
    delete query.propertyEditor;
    await router.replace({ path: route.path, query });
  }
};

// 5. 完成發布彈窗操作
const handleEditorDone = async (): Promise<void> => {
  await closeEditor();
  await loadMyListings();
};

// 6. 同步舊路徑跳轉帶入的彈窗狀態
const syncEditorFromRoute = (): void => {
  const target = route.query.propertyEditor;
  if (typeof target !== 'string') {
    return;
  }

  editorListingId.value = target === 'new' ? '' : target;
  isEditorOpen.value = true;
};

// 7. 執行狀態操作
const runAction = async (action: MyPropertyAction, listingId: string): Promise<void> => {
  if ((action === 'publish' || action === 'republish') &&
    !window.confirm(`${t(action === 'publish' ? 'property.mine.confirmPublishCharge' : 'property.mine.confirmRepublishCharge')} ${formatPoints(chargeCost.value)}`)) {
    return;
  }
  if (action === 'deactivate' && !window.confirm(t('property.mine.confirmDeactivate'))) {
    return;
  }
  if (action === 'mark-sold' && !window.confirm(t('property.sale.soldConfirm'))) {
    return;
  }

  try {
    if (props.channel === 'sale') {
      if (action === 'publish') {
        await publishPropertySale(listingId);
      }
      if (action === 'republish') {
        await republishPropertySale(listingId);
      }
      if (action === 'mark-sold') {
        await markPropertySaleSold(listingId);
      }
      if (action === 'deactivate') {
        await deactivatePropertySale(listingId);
      }
    } else {
      if (action === 'publish') {
        await publishServicedApartment(listingId);
      }
      if (action === 'republish') {
        await republishServicedApartment(listingId);
      }
      if (action === 'deactivate') {
        await deactivateServicedApartment(listingId);
      }
    }

    feedbackStore.pushToast(t('property.mine.statusUpdated'), 'success');
    if (action === 'publish' || action === 'republish') {
      await sessionStore.loadCurrentUser();
    }
    await loadMyListings();
  } catch (error: unknown) {
    feedbackStore.pushToast(
      axios.isAxiosError<{ message?: string }>(error)
        ? error.response?.data?.message ?? t('property.mine.updateError')
        : t('property.mine.updateError'),
      'error',
    );
  }
};

// 8. 輸出狀態顯示
const resolveStatusLabel = (listing: PropertyListingSummaryResponse): string =>
  t(`property.mine.${resolvePropertyStatus(listing)}`);

// 9. 輸出價格顯示
const resolveListingPriceText = (listing: PropertyListingSummaryResponse): string => {
  const prefix = props.channel === 'sale' && resolvePropertyTransactionType(listing) === 'rent'
    ? t('property.sale.rentLabel')
    : props.channel === 'sale'
      ? t('property.sale.priceLabel')
      : t('property.serviced.priceLabel');

  return `${prefix} ${resolvePropertyPriceText(listing, preferenceStore.locale)}`;
};

// 10. 輸出廣告等級
const resolveListingAdText = (listing: PropertyListingSummaryResponse): string => {
  const code = listing.property_sale?.ad_package_code;
  if (code === 'premium') {
    return '黃金置頂';
  }
  if (code === 'featured') {
    return '置頂';
  }
  if (code === 'fast_sale') {
    return '即走盤';
  }
  return '普通';
};

watch(activeTab, () => {
  void loadMyListings();
});

watch(
  () => route.query.propertyEditor,
  () => {
    syncEditorFromRoute();
  },
);

onMounted(() => {
  syncEditorFromRoute();
  void loadMyListings();
});
</script>

<template>
  <main class="property-my-page">
    <section class="property-my-heading">
      <div>
        <p class="property-kicker">
          AJO Living
        </p>
        <h1>{{ pageTitle }}</h1>
        <p class="property-my-subtitle">
          {{ t('property.mine.subTitle') }}
        </p>
      </div>
      <button
        type="button"
        class="property-button property-button--primary"
        @click="openCreateEditor"
      >
        <AppIcon
          name="plus-square"
          :size="17"
        />
        {{ t('property.list.publish') }}
      </button>
    </section>

    <section class="property-metric-row">
      <article
        v-for="metric in dashboardMetrics"
        :key="metric.label"
        class="property-metric-card"
      >
        <span>{{ metric.label }}</span>
        <strong>{{ metric.value }}</strong>
      </article>
    </section>

    <section class="property-my-toolbar">
      <label class="property-search-input">
        <AppIcon
          name="search"
          :size="17"
        />
        <input
          v-model="searchQuery"
          type="search"
          :placeholder="t('property.list.keywordPlaceholder')"
        />
      </label>
      <div class="property-tab-row">
        <button
          v-for="tab in tabOptions"
          :key="tab.value"
          type="button"
          class="property-tab"
          :class="activeTab === tab.value ? 'property-tab--active' : ''"
          @click="activeTab = tab.value"
        >
          {{ tab.label }}
        </button>
      </div>
      <div class="property-my-hint-row">
        <span>{{ t('property.mine.chargeHint', { cost: formatPoints(chargeCost) }) }}</span>
        <span>{{ t('property.mine.manageHint') }}</span>
      </div>
    </section>

    <section
      v-if="loading"
      class="property-empty"
    >
      {{ t('common.status.loading') }}
    </section>

    <section
      v-else-if="filteredItems.length === 0"
      class="property-empty"
    >
      {{ t('property.list.noListings') }}
    </section>

    <section
      v-else
      class="property-my-stack"
    >
      <article
        v-for="listing in filteredItems"
        :key="listing.listing_id"
        class="property-my-card"
      >
        <div class="property-my-card__media">
          <img
            v-if="resolvePropertyCoverImage(listing)"
            :src="resolvePropertyCoverImage(listing)?.url"
            :alt="resolvePropertyTitle(listing)"
          />
          <div
            v-else
            class="property-my-card__placeholder"
          >
            <AppIcon
              name="picture"
              :size="38"
            />
          </div>
        </div>

        <div class="property-my-card__body">
          <div class="property-my-card__top">
            <span class="property-status-pill">{{ resolveStatusLabel(listing) }}</span>
            <strong>{{ resolveListingPriceText(listing) }}</strong>
          </div>
          <h2>{{ resolvePropertyTitle(listing) }}</h2>
          <p>{{ resolvePropertySummary(listing) }}</p>
          <span class="property-my-card__community">{{ resolvePropertyCommunityName(listing) }}</span>
          <div class="property-my-card__facts">
            <span v-if="channel === 'sale'">{{ resolveListingAdText(listing) }}</span>
            <span>{{ resolvePropertyRooms(listing) }}</span>
            <span>{{ t('common.unit.sqft', { value: resolvePropertyArea(listing) }) }}</span>
          </div>
          <span class="property-my-card__date">
            {{ formatDate(listing.published_at || listing.updated_at, preferenceStore.locale) }}
          </span>
        </div>

        <div class="property-my-actions">
          <button
            type="button"
            class="property-button property-button--secondary"
            @click="openEditEditor(listing.listing_id)"
          >
            <AppIcon
              name="palette"
              :size="16"
            />
            {{ t('property.mine.edit') }}
          </button>
          <RouterLink
            :to="resolvePropertyDetailPath(listing)"
            class="property-button property-button--secondary"
          >
            <AppIcon
              name="view"
              :size="16"
            />
            {{ t('property.mine.publicDetail') }}
          </RouterLink>
          <button
            v-if="listing.publication_status === 'draft'"
            type="button"
            class="property-button property-button--primary"
            @click="runAction('publish', listing.listing_id)"
          >
            {{ t('property.mine.publish') }}
            · {{ formatPoints(chargeCost) }}
          </button>
          <button
            v-if="listing.publication_status === 'expired'"
            type="button"
            class="property-button property-button--primary"
            @click="runAction('republish', listing.listing_id)"
          >
            {{ t('property.mine.republish') }}
            · {{ formatPoints(chargeCost) }}
          </button>
          <button
            v-if="channel === 'sale' && listing.publication_status === 'active' && listing.business_status !== 'sold'"
            type="button"
            class="property-button property-button--secondary"
            @click="runAction('mark-sold', listing.listing_id)"
          >
            {{ t('property.sale.soldAction') }}
          </button>
          <button
            v-if="listing.publication_status === 'active'"
            type="button"
            class="property-button property-button--secondary"
            @click="runAction('deactivate', listing.listing_id)"
          >
            {{ t('property.mine.deactivate') }}
          </button>
        </div>
      </article>
    </section>

    <Teleport to="body">
      <div
        v-if="isEditorOpen"
        class="property-editor-modal"
        role="dialog"
        aria-modal="true"
      >
        <div
          class="property-editor-modal__backdrop"
          @click="closeEditor"
        />
        <section
          class="property-editor-modal__panel"
          :class="channel === 'serviced' ? 'property-editor-modal__panel--serviced' : ''"
        >
          <component
            :is="editorComponent"
            :key="editorKey"
            :listing-id="editorListingId"
            embedded
            @cancel="closeEditor"
            @saved="handleEditorDone"
            @published="handleEditorDone"
          />
        </section>
      </div>
    </Teleport>
  </main>
</template>

<style scoped>
.property-my-page {
  width: 100%;
  max-width: var(--layout-page-max-width);
  margin: 0 auto;
  padding: 3rem var(--layout-page-padding-inline) 5rem;
  color: rgb(var(--color-text));
}

.property-my-heading,
.property-my-toolbar,
.property-my-card {
  display: grid;
  gap: 1rem;
}

.property-my-heading {
  align-items: end;
  margin-bottom: 1.25rem;
}

.property-my-subtitle {
  max-width: 48rem;
  margin: 0.75rem 0 0;
  color: rgb(var(--color-text-muted));
  line-height: 1.7;
}

.property-kicker {
  margin: 0 0 0.6rem;
  color: rgb(var(--color-primary));
  font-size: 0.75rem;
  font-weight: 900;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.property-my-heading h1 {
  margin: 0;
  font-family: var(--font-display);
  font-size: clamp(2rem, 5vw, 3.8rem);
}

.property-button {
  display: inline-flex;
  min-height: 2.6rem;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  border-radius: 0.5rem;
  padding: 0 0.9rem;
  font-size: 0.86rem;
  font-weight: 900;
}

.property-button--primary {
  border: 1px solid rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.property-button--secondary {
  border: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
}

.property-my-toolbar {
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: rgb(var(--color-surface));
  padding: 1rem;
}

.property-metric-row {
  display: grid;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.property-metric-card {
  display: grid;
  gap: 0.35rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: rgb(var(--color-surface));
  padding: 1rem;
  box-shadow: 0 10px 34px rgb(15 23 42 / 0.05);
}

.property-metric-card span {
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 900;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.property-metric-card strong {
  font-family: var(--font-display);
  font-size: 1.65rem;
  font-weight: 800;
}

.property-my-hint-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.85rem;
  line-height: 1.6;
}

.property-my-hint-row span {
  border-radius: 999px;
  background: rgb(var(--color-surface-raised));
  padding: 0.45rem 0.75rem;
}

.property-search-input {
  display: flex;
  min-height: 2.75rem;
  align-items: center;
  gap: 0.5rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.5rem;
  background: rgb(var(--color-surface-raised));
  padding: 0 0.8rem;
}

.property-search-input input {
  width: 100%;
  border: 0;
  background: transparent;
  color: inherit;
  outline: 0;
}

.property-tab-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.property-tab {
  min-height: 2.25rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.5rem;
  background: rgb(var(--color-surface-raised));
  padding: 0 0.75rem;
  color: rgb(var(--color-text-muted));
  font-weight: 900;
}

.property-tab--active {
  border-color: rgb(var(--color-primary));
  color: rgb(var(--color-primary));
}

.property-my-stack {
  display: grid;
  gap: 1rem;
  margin-top: 1rem;
}

.property-my-card {
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: rgb(var(--color-surface));
  box-shadow: 0 10px 34px rgb(15 23 42 / 0.07);
}

.property-my-card__media {
  aspect-ratio: 4 / 3;
  background: rgb(var(--color-surface-raised));
}

.property-my-card__media img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.property-my-card__placeholder {
  display: grid;
  height: 100%;
  place-items: center;
  color: rgb(var(--color-text-muted));
}

.property-my-card__body,
.property-my-actions {
  display: grid;
  min-width: 0;
  gap: 0.75rem;
  padding: 1rem;
}

.property-my-card__top {
  display: flex;
  min-width: 0;
  justify-content: space-between;
  gap: 1rem;
}

.property-status-pill {
  border-radius: 999px;
  background: rgb(var(--color-primary-soft));
  padding: 0.3rem 0.55rem;
  color: rgb(var(--color-primary));
  font-size: 0.75rem;
  font-weight: 900;
}

.property-my-card h2,
.property-my-card p {
  margin: 0;
}

.property-my-card h2 {
  overflow: hidden;
  font-size: 1.1rem;
  font-weight: 900;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.property-my-card p,
.property-my-card__community,
.property-my-card__date {
  color: rgb(var(--color-text-muted));
  font-size: 0.9rem;
  line-height: 1.55;
}

.property-my-card p {
  display: -webkit-box;
  overflow: hidden;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.property-my-card__community,
.property-my-card__date,
.property-my-card__top strong {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.property-my-card__facts {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
}

.property-my-card__facts span {
  border-radius: 999px;
  background: rgb(var(--color-surface-raised));
  padding: 0.3rem 0.55rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 900;
}

.property-my-actions {
  align-content: start;
}

.property-empty {
  margin-top: 1rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: rgb(var(--color-surface));
  padding: 2rem;
  color: rgb(var(--color-text-muted));
  text-align: center;
}

.property-editor-modal {
  position: fixed;
  z-index: 70;
  inset: 0;
  display: grid;
  place-items: stretch center;
  padding: 1.25rem var(--layout-page-padding-inline);
}

.property-editor-modal__backdrop {
  position: absolute;
  inset: 0;
  background: rgb(15 23 42 / 0.62);
  backdrop-filter: blur(10px);
}

.property-editor-modal__panel {
  position: relative;
  width: min(100%, 72rem);
  max-height: calc(100vh - 2.5rem);
  overflow: auto;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: rgb(var(--color-canvas));
  padding: 1.25rem;
  box-shadow: 0 24px 80px rgb(15 23 42 / 0.28);
}

.property-editor-modal__panel--serviced {
  width: min(100%, 80rem);
  background: rgb(var(--color-canvas));
  padding: 2rem;
}

@media (min-width: 760px) {
  .property-my-heading {
    grid-template-columns: minmax(0, 1fr) auto;
  }

  .property-metric-row {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }

  .property-my-card {
    grid-template-columns: 16rem minmax(0, 1fr) 14rem;
  }
}

@media (max-width: 767px) {
  .property-editor-modal {
    padding: 0.75rem;
  }

  .property-editor-modal__panel {
    max-height: calc(100vh - 1.5rem);
    padding: 1rem;
  }

  .property-editor-modal__panel--serviced {
    padding: 1rem;
  }
}
</style>
