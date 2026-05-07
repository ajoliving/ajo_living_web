<!--
 * 我的物業發布頁。
 * 1. 根據頻道讀取我的樓盤或服務式住宅列表。
 * 2. 支援發布、重發、成交與下架等狀態操作。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onMounted, ref, watch } from 'vue';
import { RouterLink } from 'vue-router';
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
} from '@/httpapis/properties';
import type {
  PropertyChannel,
  PropertyListParams,
  PropertyListingSummaryResponse,
} from '@/model/property';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';
import { formatDate, formatPrice } from '@/utils/format';
import {
  resolvePropertyCommunityName,
  resolvePropertyCoverImage,
  resolvePropertyDetailPath,
  resolvePropertyPrice,
  resolvePropertyStatus,
  resolvePropertySummary,
  resolvePropertyTitle,
} from '@/utils/property';

const props = defineProps<{
  channel: PropertyChannel;
}>();

type MyPropertyTab = 'all' | 'draft' | 'active' | 'hidden' | 'expired' | 'sold';
type MyPropertyAction = 'publish' | 'republish' | 'mark-sold' | 'deactivate';

const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const preferenceStore = usePreferenceStore();

const loading = ref(false);
const activeTab = ref<MyPropertyTab>('all');
const searchQuery = ref('');
const items = ref<PropertyListingSummaryResponse[]>([]);

const pageTitle = computed(() =>
  props.channel === 'sale' ? t('property.sale.myTitle') : t('property.serviced.myTitle'),
);
const publishPath = computed(() =>
  props.channel === 'sale' ? '/properties/my/new' : '/serviced-residences/my/new',
);
const editorBasePath = computed(() =>
  props.channel === 'sale' ? '/properties/my/editor' : '/serviced-residences/my/editor',
);
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
  if (!keyword) {
    return items.value;
  }

  return items.value.filter((listing) =>
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
  } finally {
    loading.value = false;
  }
};

// 2. 執行狀態操作
const runAction = async (action: MyPropertyAction, listingId: string): Promise<void> => {
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

// 3. 輸出狀態顯示
const resolveStatusLabel = (listing: PropertyListingSummaryResponse): string =>
  t(`property.mine.${resolvePropertyStatus(listing)}`);

watch(activeTab, () => {
  void loadMyListings();
});

onMounted(() => {
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
      </div>
      <RouterLink
        :to="publishPath"
        class="property-button property-button--primary"
      >
        <AppIcon
          name="plus-square"
          :size="17"
        />
        {{ t('property.list.publish') }}
      </RouterLink>
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
            <strong>{{ formatPrice(resolvePropertyPrice(listing), preferenceStore.locale) }}</strong>
          </div>
          <h2>{{ resolvePropertyTitle(listing) }}</h2>
          <p>{{ resolvePropertySummary(listing) }}</p>
          <span class="property-my-card__community">{{ resolvePropertyCommunityName(listing) }}</span>
          <span class="property-my-card__date">
            {{ formatDate(listing.published_at || listing.updated_at, preferenceStore.locale) }}
          </span>
        </div>

        <div class="property-my-actions">
          <RouterLink
            :to="`${editorBasePath}/${listing.listing_id}`"
            class="property-button property-button--secondary"
          >
            <AppIcon
              name="palette"
              :size="16"
            />
            {{ t('property.mine.edit') }}
          </RouterLink>
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
          </button>
          <button
            v-if="listing.publication_status === 'expired'"
            type="button"
            class="property-button property-button--primary"
            @click="runAction('republish', listing.listing_id)"
          >
            {{ t('property.mine.republish') }}
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
  color: #fff;
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
  gap: 0.75rem;
  padding: 1rem;
}

.property-my-card__top {
  display: flex;
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
  font-size: 1.1rem;
  font-weight: 900;
}

.property-my-card p,
.property-my-card__community,
.property-my-card__date {
  color: rgb(var(--color-text-muted));
  font-size: 0.9rem;
  line-height: 1.55;
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

@media (min-width: 760px) {
  .property-my-heading {
    grid-template-columns: minmax(0, 1fr) auto;
  }

  .property-my-card {
    grid-template-columns: 16rem minmax(0, 1fr) 14rem;
  }
}
</style>
