<!--
 * AJO Living 首頁。
 * 1. 以現有物業及生活服務提供六個可直接進入的快捷入口。
 * 2. 獨立載入超市優惠前三項及最新樓盤，保留各區載入、重試與空態。
 * 3. 社區環境、交通及生活提示沿用會員目前大廈，缺少位置時使用公開資料。
 * 4. 沿用全站主題、語系及登入路由，不展示尚未接入的模組。
-->
<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { RouterLink, useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';

import { fetchHomeEnvironment } from '@/httpapis/home';
import { fetchPropertySaleListings } from '@/httpapis/properties';
import { fetchSupermarketSummary } from '@/httpapis/supermarket_offers';
import type { HomeEnvironment } from '@/model/home';
import type { PropertyListingSummaryResponse } from '@/model/property';
import type { SupermarketProduct } from '@/model/supermarket_offers';
import AppIcon, { type IconName } from '@/shared/components/base/AppIcon.vue';
import { usePreferenceStore } from '@/stores/preferences';
import { useSessionStore } from '@/stores/session';
import {
  resolvePropertyArea,
  resolvePropertyCoverImage,
  resolvePropertyDetailPath,
  resolvePropertyDistrict,
  resolvePropertyPrice,
  resolvePropertyPriceText,
  resolvePropertyPublisherRole,
  resolvePropertyTitle,
  resolvePropertyTransactionType,
  resolvePropertyTypeLabel,
} from '@/utils/property';
import {
  displaySupermarketStore,
  formatSupermarketHKPrice,
  resolveSupermarketProductFullTitle,
  supermarketOfferDisplayText,
} from '@/utils/supermarket_offers';

import HomeEnvironmentPanel from './widgets/HomeEnvironmentPanel.vue';
import HomeLifePanel from './widgets/HomeLifePanel.vue';
import HomeTransportPanel from './widgets/HomeTransportPanel.vue';
import HomeToolbar from './widgets/HomeToolbar.vue';
import HomePropertyNews from './widgets/HomePropertyNews.vue';
import HomeContentLinks from './widgets/HomeContentLinks.vue';
import { resolveHomeEstate } from './home_estate';

interface HomeShortcut {
  key: string;
  to: string;
  icon: IconName;
}

const { t } = useI18n();
const preferenceStore = usePreferenceStore();
const sessionStore = useSessionStore();
const route = useRoute();
const shortcuts: HomeShortcut[] = [
  { key: 'building', to: '/building', icon: 'building' },
  { key: 'properties', to: '/properties', icon: 'home' },
  { key: 'access', to: '/building?tab=access', icon: 'lock' },
  { key: 'offers', to: '/supermarket-offers', icon: 'tag' },
  { key: 'trend', to: '/trend', icon: 'browse' },
  { key: 'conversations', to: '/notifications?tab=conversations', icon: 'message' },
];
const products = ref<SupermarketProduct[]>([]);
const propertyListings = ref<PropertyListingSummaryResponse[]>([]);
const loadingOffers = ref<boolean>(true);
const loadingProperties = ref<boolean>(true);
const offersError = ref<boolean>(false);
const propertiesError = ref<boolean>(false);
const failedImages = ref<Set<string>>(new Set());
const environment = ref<HomeEnvironment | null>(null);
const loadingEnvironment = ref(true);
const environmentRetry = ref(0);
const environmentBuildingID = computed(() =>
  sessionStore.isAuthenticated && sessionStore.me?.member_status === 'active'
    ? sessionStore.me.bound_building_ids?.[0]?.trim() || undefined
    : undefined,
);
const environmentEstate = computed(() => resolveHomeEstate(route.query.estate, environmentBuildingID.value));
const requestedBuilding = computed(() => environmentEstate.value ? undefined : environmentBuildingID.value);

// 1. 保留上游優惠排序，沿用現有商品標題、商店名稱與港幣格式。
const offers = computed(() => products.value.map((product) => ({
  key: product.code,
  title: resolveSupermarketProductFullTitle(product),
  store: displaySupermarketStore(product.bestStore, preferenceStore.locale),
  price: formatSupermarketHKPrice(product.bestEffectiveUnitPrice, preferenceStore.locale),
  imageUrl: product.image_url || product.imageUrl || '',
  terms: supermarketOfferDisplayText(product.bestOffer ?? ''),
  to: `/supermarket-offers/products/${encodeURIComponent(product.code)}`,
})));

// 2. 只使用公開樓盤欄位，不組合會員或聯絡人私有資料。
const properties = computed(() => propertyListings.value.map((listing) => {
  const locale = preferenceStore.locale;
  const area = resolvePropertyArea(listing);
  return {
    key: listing.listing_id,
    title: resolvePropertyTitle(listing, locale),
    district: resolvePropertyDistrict(listing, locale),
    type: resolvePropertyTypeLabel(listing, locale),
    publisher: resolvePropertyPublisherRole(listing, locale),
    imageUrl: resolvePropertyCoverImage(listing)?.url || '',
    price: resolvePropertyPriceText(listing, locale),
    priceUnit: resolvePropertyTransactionType(listing) === 'rent' && resolvePropertyPrice(listing) > 0
      ? t('home.newShell.featured.priceSuffix')
      : '',
    area: area > 0
      ? `${t('property.sale.areaLabel')} ${t('common.unit.sqft', { value: area.toLocaleString(locale) })}`
      : t('property.editor.areaUnverifiedField'),
    to: resolvePropertyDetailPath(listing),
  };
}));

// 3. 超市資料失敗或上游不可用時，只切換本區的重試狀態。
const loadOffers = async (): Promise<void> => {
  loadingOffers.value = true;
  offersError.value = false;
  try {
    const { data } = await fetchSupermarketSummary();
    if (data.data.metadata?.upstreamAvailable === false) {
      offersError.value = true;
      products.value = [];
      return;
    }
    products.value = data.data.bestDiscounts.slice(0, 3);
  } catch {
    offersError.value = true;
    products.value = [];
  } finally {
    loadingOffers.value = false;
  }
};

// 4. 最新樓盤獨立載入，避免另一區塊失敗阻塞首頁入口。
const loadProperties = async (): Promise<void> => {
  loadingProperties.value = true;
  propertiesError.value = false;
  try {
    const { data } = await fetchPropertySaleListings({ page: 1, page_size: 3, sort_by: 'latest' });
    propertyListings.value = data.data.items.slice(0, 3);
  } catch {
    propertiesError.value = true;
    propertyListings.value = [];
  } finally {
    loadingProperties.value = false;
  }
};

// 5. 以響應式狀態切換圖片後備內容，保持卡片原有尺寸。
const markFailedImage = (url: string): void => {
  failedImages.value.add(url);
};

// 6. 切換物業或語系即清除舊讀數，避免稍後完成的請求覆蓋目前資料。
watch(
  [requestedBuilding, environmentEstate, () => preferenceStore.locale, environmentRetry],
  async ([buildingID, estate, locale], _previous, onCleanup) => {
    const request = new AbortController();
    onCleanup(() => request.abort());
    environment.value = null;
    loadingEnvironment.value = true;
    try {
      const { data } = await fetchHomeEnvironment(
        { lang: locale === 'en' ? 'en' : 'tc', ...(estate ? { estate } : buildingID ? { building_id: buildingID } : {}) },
        request.signal,
      );
      if (!request.signal.aborted) environment.value = data.data;
    } catch {
      // 6.1 請求失敗時保留公開查詢入口，不阻塞超市及樓盤。
    } finally {
      if (!request.signal.aborted) loadingEnvironment.value = false;
    }
  },
  { immediate: true },
);

onMounted(() => {
  void loadOffers();
  void loadProperties();
});
</script>

<template>
  <div id="page-home" class="home-page">
    <HomeToolbar class="home-container" :building-id="environmentBuildingID" />

    <!-- 1. 首頁封面及主要物業入口 -->
    <section class="home-cover" aria-labelledby="home-title">
      <img
        class="home-cover-image"
        src="/home_stage/community_cover.webp"
        alt=""
        fetchpriority="high"
        width="1680"
        height="640"
      />
      <div class="home-container home-cover-content">
        <h1 id="home-title">AJO LIVING</h1>
        <p class="home-cover-title">{{ t('home.dashboard.coverTitle') }}</p>
        <p class="home-cover-description">{{ t('home.dashboard.coverDescription') }}</p>
        <RouterLink class="home-cover-action" to="/building">
          {{ t('home.dashboard.shortcuts.building') }}
          <AppIcon name="arrow-right" :size="18" />
        </RouterLink>
      </div>
    </section>

    <!-- 2. 六個真實功能入口 -->
    <nav class="home-container home-shortcuts" :aria-label="t('home.dashboard.services')">
      <RouterLink
        v-for="shortcut in shortcuts"
        :key="shortcut.key"
        :to="shortcut.to"
        class="home-shortcut"
      >
        <span class="home-shortcut-icon"><AppIcon :name="shortcut.icon" :size="24" /></span>
        <span>{{ t(`home.dashboard.shortcuts.${shortcut.key}`) }}</span>
      </RouterLink>
    </nav>

    <!-- 3. 即時超市優惠 -->
    <section
      class="home-container home-content-section home-offers"
      :aria-busy="loadingOffers"
      aria-labelledby="home-offers-title"
    >
      <div class="home-section-heading">
        <h2 id="home-offers-title">{{ t('home.dashboard.offersTitle') }}</h2>
        <RouterLink class="home-more" to="/supermarket-offers">
          {{ t('home.dashboard.moreOffers') }}<AppIcon name="arrow-right" :size="16" />
        </RouterLink>
      </div>
      <div v-if="loadingOffers" class="home-card-grid" role="status">
        <span class="sr-only">{{ t('home.dashboard.loadingOffers') }}</span>
        <div v-for="slot in 3" :key="slot" class="home-skeleton home-offer-skeleton" aria-hidden="true">
          <span class="home-skeleton-image"></span><span class="home-skeleton-line"></span>
          <span class="home-skeleton-line home-skeleton-line-short"></span>
        </div>
      </div>
      <div v-else-if="offersError" class="home-state" role="status">
        <p>{{ t('home.dashboard.offersError') }}</p>
        <button type="button" @click="loadOffers"><AppIcon name="reload" :size="16" />{{ t('home.dashboard.retry') }}</button>
      </div>
      <p v-else-if="offers.length === 0" class="home-state" role="status">{{ t('home.dashboard.offersEmpty') }}</p>
      <div v-else class="home-card-grid">
        <RouterLink v-for="offer in offers" :key="offer.key" :to="offer.to" class="home-offer-card">
          <div class="home-offer-image">
            <img
              v-if="offer.imageUrl && !failedImages.has(offer.imageUrl)"
              :src="offer.imageUrl"
              :alt="offer.title"
              loading="lazy"
              @error="markFailedImage(offer.imageUrl)"
            />
            <AppIcon v-else name="picture" :size="32" />
          </div>
          <div class="home-offer-body">
            <p class="home-offer-store">{{ offer.store }}</p>
            <h3>{{ offer.title }}</h3>
            <p class="home-offer-price">{{ offer.price }}<span>{{ t('home.dashboard.perItem') }}</span></p>
            <p v-if="offer.terms" class="home-offer-terms">{{ offer.terms }}</p>
          </div>
        </RouterLink>
      </div>
    </section>

    <!-- 4. 目前大廈環境；沒有可用位置時顯示香港天氣 -->
    <HomeEnvironmentPanel
      class="home-container home-content-section"
      :environment="environment"
      :loading="loadingEnvironment"
      @retry="environmentRetry++"
    />

    <!-- 5. 港鐵到站及日常公共資訊 -->
    <HomeTransportPanel
      class="home-container home-content-section"
      :transport="environment?.transport"
      :loading="loadingEnvironment"
      :reference-area="environment?.scope === 'reference'"
      @retry="environmentRetry++"
    />
    <HomeLifePanel
      class="home-container home-content-section"
      :environment="environment"
      :loading="loadingEnvironment"
    />

    <!-- 6. 物業資訊及最新公開樓盤 -->
    <HomePropertyNews class="home-container home-content-section" />
    <section
      class="home-container home-content-section home-properties"
      :aria-busy="loadingProperties"
      aria-labelledby="home-properties-title"
    >
      <div class="home-section-heading">
        <h2 id="home-properties-title">{{ t('home.dashboard.propertiesTitle') }}</h2>
        <RouterLink class="home-more" to="/properties">
          {{ t('home.dashboard.moreProperties') }}<AppIcon name="arrow-right" :size="16" />
        </RouterLink>
      </div>
      <div v-if="loadingProperties" class="home-card-grid" role="status">
        <span class="sr-only">{{ t('home.dashboard.loadingProperties') }}</span>
        <div v-for="slot in 3" :key="slot" class="home-skeleton" aria-hidden="true">
          <span class="home-skeleton-image"></span><span class="home-skeleton-line"></span>
          <span class="home-skeleton-line home-skeleton-line-short"></span>
        </div>
      </div>
      <div v-else-if="propertiesError" class="home-state" role="status">
        <p>{{ t('home.dashboard.propertiesError') }}</p>
        <button type="button" @click="loadProperties"><AppIcon name="reload" :size="16" />{{ t('home.dashboard.retry') }}</button>
      </div>
      <p v-else-if="properties.length === 0" class="home-state" role="status">{{ t('home.dashboard.propertiesEmpty') }}</p>
      <div v-else class="home-card-grid home-property-grid">
        <RouterLink v-for="property in properties" :key="property.key" :to="property.to" class="home-property-card">
          <div class="home-property-image">
            <img
              v-if="property.imageUrl && !failedImages.has(property.imageUrl)"
              :src="property.imageUrl"
              :alt="property.title"
              loading="lazy"
              @error="markFailedImage(property.imageUrl)"
            />
            <AppIcon v-else name="picture" :size="40" />
            <span class="home-property-publisher">{{ property.publisher }}</span>
          </div>
          <div class="home-property-body">
            <div class="home-property-tags"><span>{{ property.district }}</span><span>{{ property.type }}</span></div>
            <h3>{{ property.title }}</h3>
            <p class="home-property-price">{{ property.price }}<span>{{ property.priceUnit }}</span></p>
            <p class="home-property-area">{{ property.area }}</p>
          </div>
        </RouterLink>
      </div>
    </section>
    <HomeContentLinks class="home-container home-content-section" />
  </div>
</template>

<style scoped lang="scss" src="./styles.scss"></style>
