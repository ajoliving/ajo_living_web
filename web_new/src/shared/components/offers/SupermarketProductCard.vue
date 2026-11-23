<!--
 * 首頁與超市優惠共用商品卡。
 * 1. 統一品牌、商品圖片、最低價門店、實際原價及每件節省。
 * 2. 詳情連結與收藏操作分開，保留鍵盤導覽及圖片失效後備。
-->
<script setup lang="ts">
import { computed, ref } from 'vue';
import { RouterLink, type RouteLocationRaw } from 'vue-router';
import { useI18n } from 'vue-i18n';

import type { SupermarketProduct } from '@/model/supermarket_offers';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import { usePreferenceStore } from '@/stores/preferences';
import {
  displaySupermarketStore,
  formatSupermarketHKPrice,
  resolveSupermarketProductBrand,
  resolveSupermarketProductFullTitle,
  resolveSupermarketProductName,
  resolveSupermarketProductUnit,
  supermarketOfferDisplayText,
  supermarketStorePrices,
} from '@/utils/supermarket_offers';

const props = defineProps<{ product: SupermarketProduct; to: RouteLocationRaw }>();
const emit = defineEmits<{ favorite: [] }>();
const { t } = useI18n();
const preferences = usePreferenceStore();
const failedImage = ref('');

// 1. 商品名稱、品牌及圖片沿用同一組正式展示欄位。
const brand = computed(() => resolveSupermarketProductBrand(props.product));
const name = computed(() => [
  resolveSupermarketProductName(props.product),
  resolveSupermarketProductUnit(props.product),
].filter(Boolean).join(' '));
const image = computed(() => props.product.image_url || props.product.imageUrl || '');
const title = computed(() => resolveSupermarketProductFullTitle(props.product));
const initial = computed(() => brand.value.slice(0, 1) || name.value.slice(0, 1));

// 2. 只展示並列最低價門店，原價只取勝出門店的實際資料。
const prices = computed(() => supermarketStorePrices(props.product, preferences.locale));
const featured = computed(() => prices.value[0]!);
const stores = computed(() => prices.value
  .filter((price) => price.effectiveUnitPrice === featured.value.effectiveUnitPrice)
  .map((price) => displaySupermarketStore(price.store, preferences.locale))
  .join(' / '));
const priceText = computed(() => formatSupermarketHKPrice(featured.value.effectiveUnitPrice, preferences.locale));
const originalText = computed(() => featured.value.listPrice > featured.value.effectiveUnitPrice
  ? t('offers.list.originalPrice', {
    price: `$${featured.value.listPrice.toLocaleString(preferences.locale, {
      minimumFractionDigits: 2, maximumFractionDigits: 2,
    })}`,
  })
  : '');

// 3. 優先使用已解析每件節省，再按實際原價或下一個不同現價計算。
const savingsText = computed(() => {
  const price = featured.value;
  const offer = supermarketOfferDisplayText(price.offer);
  const reported = offer.match(/(?:每件慳|saved per item)\s*[$HK]*\s*([\d,.]+)/i);
  const reportedSaving = reported ? Number(reported[1]?.replace(/,/g, '')) : 0;
  const nextPrice = prices.value.find((item) => item.effectiveUnitPrice > price.effectiveUnitPrice);
  const saving = reportedSaving >= 0.01
    ? reportedSaving
    : price.listPrice > price.effectiveUnitPrice
      ? price.listPrice - price.effectiveUnitPrice
      : (nextPrice?.effectiveUnitPrice ?? price.effectiveUnitPrice) - price.effectiveUnitPrice;
  return saving >= 0.01 ? t('offers.list.savedPerItem', {
    amount: `$${saving.toLocaleString(preferences.locale, {
      minimumFractionDigits: Number.isInteger(saving) ? 0 : 2, maximumFractionDigits: 2,
    })}`,
  }) : '';
});
</script>

<template>
  <article class="gp-card">
    <RouterLink class="gp-card-link" :to="to">
      <div class="gp-card-img">
        <span v-if="brand" class="gp-card-brand-chip">{{ brand }}</span>
        <img
          v-if="image && failedImage !== image"
          :src="image"
          :alt="title"
          loading="lazy"
          @error="failedImage = image"
        />
        <span v-else class="gp-card-img-ph">{{ initial }}</span>
      </div>
      <div class="gp-card-body">
        <div class="gp-card-name"><span class="gp-card-name-text">{{ name }}</span></div>
        <div class="gp-card-best-store" :title="stores">{{ stores }}</div>
        <div class="gp-card-best-price">
          <span>{{ priceText }}</span><small>{{ t('offers.list.perItem') }}</small>
        </div>
        <div v-if="originalText" class="gp-card-best-compare">{{ originalText }}</div>
        <div v-if="savingsText" class="gp-card-best-offer">{{ savingsText }}</div>
      </div>
    </RouterLink>
    <button
      type="button"
      class="gp-card-fav"
      :class="{ on: product.isFavorite }"
      :aria-label="product.isFavorite ? t('offers.list.favorited') : t('offers.list.favorite')"
      :title="product.isFavorite ? t('offers.list.favorited') : t('offers.list.favorite')"
      :aria-pressed="Boolean(product.isFavorite)"
      @click.stop="emit('favorite')"
    >
      <AppIcon name="heart" :size="16" />
    </button>
  </article>
</template>

<style scoped>
/* 1. 兩個頁面共用超市卡片的直式版面與固定圖片高度。 */
.gp-card {
  position: relative;
  display: flex;
  min-width: 0;
  margin: 0;
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 16px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  font-family: "Noto Sans TC", sans-serif;
  letter-spacing: 0;
  transition: border-color 0.15s ease;
}

.gp-card:hover {
  border-color: rgb(var(--color-primary));
}

.gp-card-link {
  display: flex;
  width: 100%;
  min-width: 0;
  flex-direction: column;
  color: inherit;
}

.gp-card-link:focus-visible,
.gp-card-fav:focus-visible {
  outline: 2px solid rgb(var(--color-primary));
  outline-offset: -3px;
}

.gp-card-img {
  position: relative;
  display: flex;
  height: 148px;
  flex: 0 0 148px;
  align-items: center;
  justify-content: center;
  background: rgb(var(--color-surface));
}

.gp-card-img img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.gp-card-img-ph {
  display: grid;
  width: 56px;
  height: 56px;
  place-items: center;
  color: rgb(var(--color-text-muted));
  font-size: 24px;
  font-weight: 800;
}

.gp-card-brand-chip {
  position: absolute;
  z-index: 2;
  top: 8px;
  left: 8px;
  max-width: calc(100% - 56px);
  overflow: hidden;
  border-radius: 5px;
  padding: 2px 7px;
  background: #2a7a6e;
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.gp-card-fav {
  position: absolute;
  z-index: 3;
  top: 0;
  right: 0;
  display: grid;
  width: 44px;
  height: 44px;
  place-items: center;
  border: 0;
  border-radius: 50%;
  background: transparent;
  color: rgb(var(--color-text-muted));
  cursor: pointer;
}

.gp-card-fav::before {
  position: absolute;
  z-index: -1;
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: rgb(var(--color-surface));
  box-shadow: 0 2px 8px rgb(26 26 26 / 0.08);
  content: '';
}

.gp-card-fav.on {
  color: rgb(var(--color-primary));
}

.gp-card-fav.on :deep(.app-icon) {
  fill: currentColor;
}

.gp-card-body {
  display: flex;
  min-width: 0;
  flex: 1 1 auto;
  flex-direction: column;
  align-items: stretch;
  gap: 8px;
  border-top: 1px solid rgb(var(--color-border));
  padding: 10px 10px 12px;
  text-align: center;
}

.gp-card-name {
  display: flex;
  height: 40px;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.gp-card-name-text {
  display: -webkit-box;
  min-width: 0;
  max-height: 40px;
  overflow: hidden;
  font-size: 15px;
  font-weight: 800;
  line-height: 20px;
  overflow-wrap: anywhere;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.gp-card-best-store {
  width: 100%;
  overflow: hidden;
  border-radius: 8px;
  padding: 7px 8px;
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
  font-size: 18px;
  font-weight: 800;
  line-height: 1.25;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.gp-card-best-price {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  justify-content: center;
  gap: 4px;
  font-size: 20px;
  font-weight: 800;
  line-height: 1.2;
}

.gp-card-best-price small {
  font-size: 12px;
  font-weight: 700;
}

.gp-card-best-compare {
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  line-height: 1.3;
  text-decoration: line-through;
}

.gp-card-best-offer {
  margin-top: auto;
  color: rgb(var(--color-primary));
  font-size: 18px;
  font-weight: 800;
  line-height: 1.25;
  overflow-wrap: anywhere;
}

/* 2. 手機保留超市頁的緊湊字級與兩列卡片內距。 */
@media (max-width: 800px) {
  .gp-card-body {
    padding-inline: 8px;
  }

  .gp-card-best-store,
  .gp-card-best-offer {
    font-size: 16px;
  }

  .gp-card-best-price {
    font-size: 18px;
  }
}

/* 3. 夜間圖片表面沿用超市頁的中性深色。 */
:global(html[data-theme="night"]) .gp-card-img {
  background: rgb(42 48 56);
}
</style>
