<!--
 * 樓盤租售列表頁。
 * 1. 桌面採左側篩選、中間列表與右側廣告三欄布局。
 * 2. 手機將篩選收進底部彈窗，入口固定在搜尋框左側。
 * 3. 中間含排序欄、後端樓盤卡片與分頁。
 * 4. 卡片含類型堆疊、售/租標識、代理公司、呎價、位置。
 * 5. 右側接入 3 個 16:9 短廣告與 2 個 9:16 長廣告。
 * 6. 使用後端樓盤接口與收藏接口。
 * 7. CSS 變量與 class 名稱嚴格對齊 HTML 設計稿。
-->
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';

import { favoritePropertySale, fetchPropertySaleListings, unfavoritePropertySale } from '@/httpapis/properties';
import { readStoredAccessToken } from '@/httpapis/auth-session';
import {
  propertyAreaRangeFilterOptions,
  propertyBedroomFilterOptions,
  propertyPriceRangeFilterOptions,
  propertyPublisherFilterOptions,
  propertyRegionFilterOptions,
  propertyRenovationFilterOptions,
  propertyTransactionTypeFilterOptions,
  propertyTypeFilterOptions,
} from '@/constants/property';
import type { PaginationMeta } from '@/model/api';
import type { PropertyListParams, PropertyListingSummaryResponse } from '@/model/property';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import FilterTag from '@/shared/components/base/FilterTag.vue';
import ListingSideAds from '@/shared/components/ads/ListingSideAds.vue';
import {
  resolvePropertyArea,
  resolvePropertyCommunityName,
  resolvePropertyCoverImage,
  resolvePropertyDistrict,
  resolvePropertyPrice,
  resolvePropertyPriceText,
  resolvePropertyRooms,
  resolvePropertyTagLabels,
  resolvePropertyTitle,
  resolvePropertyTransactionType,
  resolvePropertyTypeLabel,
} from '@/utils/property';

// 1. 路由
const router = useRouter();
const pageSize = 12;

// 2. 搜尋關鍵字
const keyword = ref('');
const isFilterOpen = ref(false);

// 3. 排序選項
const sortBy = ref('latest');

// 4. 自動補全下拉顯示
const showAutocomplete = ref(false);
const loading = ref(false);
const errorMessage = ref('');
const currentPage = ref(1);
const items = ref<PropertyListingSummaryResponse[]>([]);
const pagination = ref<PaginationMeta>({ page: 1, page_size: pageSize, total: 0 });

// 5. 自動補全項目
interface AutocompleteItem {
  value: string;
  prefix: string;
  suffix: string;
  icon: 'building' | 'home' | 'pin';
}
const autocompleteItems = computed<AutocompleteItem[]>(() =>
  items.value.slice(0, 6).map((listing) => {
    const value = resolvePropertyCommunityName(listing);
    return {
      value,
      prefix: value.slice(0, 1),
      suffix: value.slice(1) || resolvePropertyTitle(listing),
      icon: listing.property_sale?.property_type === 'house' ? 'home' : 'building',
    };
  }),
);

// 6. 篩選群組資料
interface FilterOption {
  label: string;
  value: string;
}
interface FilterGroup {
  key: string;
  title: string;
  options: FilterOption[];
  activeValue: string;
}
const filterGroups = ref<FilterGroup[]>([
  {
    key: 'region',
    title: '地區',
    options: propertyRegionFilterOptions,
    activeValue: '',
  },
  {
    key: 'transaction',
    title: '性質',
    options: propertyTransactionTypeFilterOptions,
    activeValue: '',
  },
  {
    key: 'property_type',
    title: '物業類型',
    options: propertyTypeFilterOptions,
    activeValue: '',
  },
  {
    key: 'price',
    title: '售價範圍',
    options: propertyPriceRangeFilterOptions,
    activeValue: '',
  },
  {
    key: 'area',
    title: '實用面積',
    options: propertyAreaRangeFilterOptions,
    activeValue: '',
  },
  {
    key: 'bedroom',
    title: '房間',
    options: propertyBedroomFilterOptions,
    activeValue: '',
  },
  {
    key: 'renovation',
    title: '裝修',
    options: propertyRenovationFilterOptions,
    activeValue: '',
  },
  {
    key: 'publisher',
    title: '發布者',
    options: propertyPublisherFilterOptions,
    activeValue: '',
  },
]);
const activeFilterCount = computed(() =>
  filterGroups.value.filter((group) => Boolean(group.activeValue)).length,
);
const mobileQuickFilterGroups = computed<FilterGroup[]>(() =>
  ['region', 'property_type', 'bedroom', 'renovation']
    .map((key) => filterGroups.value.find((group) => group.key === key))
    .filter((group): group is FilterGroup => Boolean(group)),
);

// 7. 樓盤卡片資料
interface PropertyCard {
  id: string;
  listing: PropertyListingSummaryResponse;
  typeStack: string[];
  imageUrl?: string;
  tags: { label: string; dark?: boolean }[];
  title: string;
  location: string;
  sub: string;
  agent?: string;
  priceKind: 'sale' | 'rent';
  price: string;
  priceUnit: string;
  area: string;
  areaPrice: string;
  pills: string[];
  favorite: boolean;
}
const cards = computed<PropertyCard[]>(() => items.value.map((listing) => toPropertyCard(listing)));

// 8. 分頁資料
interface PaginationPage {
  label: string | number;
  active?: boolean;
  disabled?: boolean;
  key: string | number;
  page: number;
}
const totalPages = computed(() => Math.max(1, Math.ceil(pagination.value.total / pageSize)));
const paginationPages = computed<PaginationPage[]>(() => {
  const pages = new Set<number>([1, currentPage.value, totalPages.value]);
  if (currentPage.value > 1) pages.add(currentPage.value - 1);
  if (currentPage.value < totalPages.value) pages.add(currentPage.value + 1);
  const numericPages = [...pages].filter((page) => page >= 1 && page <= totalPages.value).sort((a, b) => a - b);

  return [
    { label: '上一頁', key: 'prev', page: Math.max(1, currentPage.value - 1), disabled: currentPage.value <= 1 },
    ...numericPages.map((page) => ({ label: page, key: page, page, active: page === currentPage.value })),
    { label: '下一頁', key: 'next', page: Math.min(totalPages.value, currentPage.value + 1), disabled: currentPage.value >= totalPages.value },
  ];
});

const resultRangeText = computed(() => {
  if (pagination.value.total === 0) {
    return '暫無樓盤';
  }
  const start = (currentPage.value - 1) * pageSize + 1;
  const end = Math.min(currentPage.value * pageSize, pagination.value.total);
  return `第 ${start}-${end} 筆，共 ${pagination.value.total} 筆`;
});

// 9. 載入樓盤列表
const loadListings = async (): Promise<void> => {
  loading.value = true;
  errorMessage.value = '';
  try {
    const { data } = await fetchPropertySaleListings(buildListParams());
    items.value = data.data.items;
    pagination.value = data.data.pagination;
  } catch {
    errorMessage.value = '暫時無法讀取樓盤列表。';
    items.value = [];
    pagination.value = { page: currentPage.value, page_size: pageSize, total: 0 };
  } finally {
    loading.value = false;
  }
};

// 10. 建立查詢參數
const buildListParams = (): PropertyListParams => {
  const params: PropertyListParams = {
    page: currentPage.value,
    page_size: pageSize,
    keyword: keyword.value.trim() || undefined,
    sort_by: sortBy.value as PropertyListParams['sort_by'],
    has_media: false,
  };
  const region = activeValue('region');
  const transaction = activeValue('transaction');
  const propertyType = activeValue('property_type');
  const price = propertyPriceRangeFilterOptions.find((item) => item.value === activeValue('price'));
  const area = propertyAreaRangeFilterOptions.find((item) => item.value === activeValue('area'));
  const bedroom = activeValue('bedroom');
  const renovation = activeValue('renovation');
  const publisher = activeValue('publisher');

  if (region) params.region_code = region;
  if (transaction === 'sale' || transaction === 'rent') params.transaction_type = transaction;
  if (propertyType) params.property_type = propertyType;
  if (price?.min) params.min_price_hkd = price.min;
  if (price?.max) params.max_price_hkd = price.max;
  if (area?.min) params.min_area_sqft = area.min;
  if (area?.max) params.max_area_sqft = area.max;
  if (bedroom) params.bedroom_count = Number(bedroom);
  if (renovation) params.feature_tags = renovation;
  if (publisher) params.publisher_identity_type = publisher;

  return params;
};

// 11. 取得目前篩選值
const activeValue = (groupKey: string): string =>
  filterGroups.value.find((group) => group.key === groupKey)?.activeValue ?? '';

// 12. 切換篩選標籤（同組互斥）
const handleFilterToggle = (groupKey: string, optionValue: string) => {
  const group = filterGroups.value.find((g) => g.key === groupKey);
  if (group) {
    group.activeValue = optionValue;
  }
  currentPage.value = 1;
  void loadListings();
};

// 13. 開關手機端篩選彈窗
const openFilterSheet = (): void => {
  isFilterOpen.value = true;
};

const closeFilterSheet = (): void => {
  isFilterOpen.value = false;
};

// 14. 清除樓盤篩選條件
const clearFilters = (): void => {
  filterGroups.value.forEach((group) => {
    group.activeValue = '';
  });
  currentPage.value = 1;
  void loadListings();
};

// 15. 點擊卡片跳轉詳情
const handleCardClick = (id: string) => {
  void router.push(`/properties/${id}`);
};

// 16. 隱藏失效樓盤封面
const hideFailedImage = (event: Event): void => {
  (event.currentTarget as HTMLImageElement).style.display = 'none';
};

// 17. 點擊分頁
const handlePageSelect = (page: PaginationPage) => {
  if (page.disabled || page.page === currentPage.value) {
    return;
  }
  currentPage.value = page.page;
  void loadListings();
};

// 18. 提交搜尋
const submitSearch = (): void => {
  currentPage.value = 1;
  showAutocomplete.value = false;
  void loadListings();
};

// 19. 切換排序
const handleSortChange = (): void => {
  currentPage.value = 1;
  void loadListings();
};

// 20. 切換收藏
const toggleFavorite = async (card: PropertyCard): Promise<void> => {
  if (!readStoredAccessToken()) {
    await router.push({ path: '/login', query: { redirect: '/properties' } });
    return;
  }
  try {
    if (card.favorite) {
      await unfavoritePropertySale(card.id);
      patchFavorite(card.id, false);
      return;
    }
    await favoritePropertySale(card.id);
    patchFavorite(card.id, true);
  } catch {
    errorMessage.value = '收藏操作失敗。';
  }
};

// 21. 預約睇樓
const openAppointment = (card: PropertyCard): void => {
  void router.push({ path: `/properties/${card.id}`, query: { action: 'appointment' } });
};

// 22. 更新收藏狀態
const patchFavorite = (listingId: string, isFavorite: boolean): void => {
  items.value = items.value.map((listing) =>
    listing.listing_id === listingId ? { ...listing, is_favorite: isFavorite } : listing,
  );
};

// 23. 轉換卡片資料
const toPropertyCard = (listing: PropertyListingSummaryResponse): PropertyCard => {
  const sale = listing.property_sale;
  const priceKind = resolvePropertyTransactionType(listing);
  const area = resolvePropertyArea(listing);
  const price = resolvePropertyPrice(listing);
  const unitPrice = area > 0 && price > 0 ? `@${formatHKD(Math.round(price / area))}/呎` : '';
  const cover = resolvePropertyCoverImage(listing);
  const district = resolvePropertyDistrict(listing, 'zh-HK');
  const community = resolvePropertyCommunityName(listing);
  const role = listing.publisher_identity_type === 'agent' ? '代理盤' : '業主盤';
  const typeLabel = resolvePropertyTypeLabel(listing, 'zh-HK');
  const agent = sale?.agency_company_name?.trim() || (role === '代理盤' ? sale?.publisher_role_label || role : '業主自讓');

  return {
    id: listing.listing_id,
    listing,
    typeStack: [typeLabel, role],
    imageUrl: cover?.url,
    tags: [{ label: district }, { label: typeLabel, dark: priceKind === 'rent' }],
    title: resolvePropertyTitle(listing),
    location: `${district} · ${community}`,
    sub: [
      sale?.public_location_text,
      resolvePropertyRooms(listing),
      sale?.direction,
    ].filter(Boolean).join(' · '),
    agent,
    priceKind,
    price: resolvePropertyPriceText(listing, 'zh-HK'),
    priceUnit: priceKind === 'rent' ? '/ 月' : '',
    area: `實用面積 ${area.toLocaleString('zh-HK')}呎`,
    areaPrice: unitPrice,
    pills: resolvePropertyTagLabels(listing, 'zh-HK', 4),
    favorite: Boolean(listing.is_favorite),
  };
};

// 24. 格式化港幣
const formatHKD = (value: number): string =>
  `HK$${value.toLocaleString('zh-HK')}`;

// 25. 顯示自動補全
const showAC = () => {
  showAutocomplete.value = keyword.value.length > 0;
};

// 26. 隱藏自動補全
const hideAC = () => {
  setTimeout(() => {
    showAutocomplete.value = false;
  }, 200);
};

// 27. 選擇自動補全項目
const selectAC = (value: string) => {
  keyword.value = value;
  showAutocomplete.value = false;
  submitSearch();
};

onMounted(() => {
  void loadListings();
});
</script>

<template>
  <div class="page">
    <div class="lp">
      <button
        v-if="isFilterOpen"
        type="button"
        class="filter-backdrop"
        aria-label="關閉篩選條件"
        @click="closeFilterSheet"
      ></button>

      <!-- 1. 左側篩選欄 -->
      <aside
        id="property-filter-sheet"
        class="lf"
        :class="{ 'is-filter-open': isFilterOpen }"
        :role="isFilterOpen ? 'dialog' : undefined"
        :aria-modal="isFilterOpen ? 'true' : undefined"
        aria-labelledby="property-filter-title"
      >
        <header class="filter-sheet-header">
          <div>
            <span class="filter-sheet-eyebrow">樓盤放售</span>
            <h2 id="property-filter-title">篩選條件</h2>
          </div>
          <button
            type="button"
            class="filter-sheet-close"
            aria-label="關閉篩選條件"
            @click="closeFilterSheet"
          >
            <AppIcon
              name="close"
              :size="20"
            />
          </button>
        </header>

        <!-- 1.1 搜尋欄（含自動補全） -->
        <div class="sbar desktop-filter-search">
          <form
            class="autocomplete-wrap"
            @submit.prevent="submitSearch"
          >
            <input
              v-model="keyword"
              class="sinput"
              placeholder="搜尋樓盤…"
              autocomplete="off"
              @input="showAC"
              @blur="hideAC"
            />
            <div
              v-if="showAutocomplete"
              class="autocomplete-drop"
            >
              <div
                v-for="item in autocompleteItems"
                :key="item.value"
                class="ac-item"
                @mousedown="selectAC(item.value)"
              >
                <span class="ac-icon">
                  <svg
                    v-if="item.icon === 'building'"
                    width="14"
                    height="14"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="1.8"
                  >
                    <rect x="4" y="3" width="16" height="18" rx="1" />
                    <line x1="9" y1="7" x2="9" y2="9" />
                    <line x1="15" y1="7" x2="15" y2="9" />
                    <line x1="9" y1="12" x2="9" y2="14" />
                    <line x1="15" y1="12" x2="15" y2="14" />
                  </svg>
                  <svg
                    v-else-if="item.icon === 'home'"
                    width="14"
                    height="14"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="1.8"
                  >
                    <path d="M3 12l9-9 9 9" stroke-linecap="round" stroke-linejoin="round" />
                    <path d="M5 10v10h14V10" stroke-linecap="round" stroke-linejoin="round" />
                  </svg>
                  <svg
                    v-else
                    width="14"
                    height="14"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="1.8"
                  >
                    <path d="M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7z" />
                    <circle cx="12" cy="9" r="2.5" />
                  </svg>
                </span>
                <span><span class="ac-highlight">{{ item.prefix }}</span>{{ item.suffix }}</span>
              </div>
            </div>
          </form>
          <button
            type="button"
            class="sbtn"
            @click="submitSearch"
          >搜</button>
        </div>

        <!-- 1.2 篩選群組 -->
        <div class="filter-sections">
          <section
            v-for="group in filterGroups"
            :key="group.key"
            class="fs"
          >
            <div class="ft-title">{{ group.title }}</div>
            <div class="ftags">
              <FilterTag
                v-for="option in group.options"
                :key="`${group.key}-${option.value}`"
                :label="option.label"
                :active="option.value === group.activeValue"
                @toggle="handleFilterToggle(group.key, option.value)"
              />
            </div>
          </section>
        </div>

        <div class="filter-sheet-actions">
          <button
            type="button"
            class="filter-sheet-reset"
            @click="clearFilters"
          >全部清除</button>
          <button
            type="button"
            class="filter-sheet-apply"
            @click="closeFilterSheet"
          >查看 {{ pagination.total }} 個結果</button>
        </div>
      </aside>

      <!-- 2. 中間列表區 -->
      <main class="lr">
        <div class="mobile-listing-controls">
          <div class="mobile-search-toolbar">
            <form
              class="mobile-search-form"
              @submit.prevent="submitSearch"
            >
              <div class="autocomplete-wrap">
                <input
                  v-model="keyword"
                  class="sinput"
                  placeholder="搜尋樓盤…"
                  autocomplete="off"
                  @input="showAC"
                  @blur="hideAC"
                />
                <div
                  v-if="showAutocomplete"
                  class="autocomplete-drop"
                >
                  <button
                    v-for="item in autocompleteItems"
                    :key="item.value"
                    type="button"
                    class="ac-item"
                    @mousedown="selectAC(item.value)"
                  >
                    <span><span class="ac-highlight">{{ item.prefix }}</span>{{ item.suffix }}</span>
                  </button>
                </div>
              </div>
              <button
                type="submit"
                class="mobile-search-button"
                aria-label="搜尋樓盤"
              >
                <AppIcon
                  name="search"
                  :size="19"
                />
              </button>
            </form>
          </div>

          <div
            class="mobile-filter-rail"
            aria-label="樓盤快捷篩選"
          >
            <select
              v-for="group in mobileQuickFilterGroups"
              :key="group.key"
              v-model="group.activeValue"
              class="mobile-filter-select"
              :aria-label="group.title"
              @change="handleFilterToggle(group.key, group.activeValue)"
            >
              <option value="">{{ group.title }}</option>
              <option
                v-for="option in group.options"
                :key="option.value"
                :value="option.value"
              >{{ option.label }}</option>
            </select>
            <select
              v-model="sortBy"
              class="mobile-filter-select mobile-sort-select"
              aria-label="樓盤排序"
              @change="handleSortChange"
            >
              <option value="latest">最新更新</option>
              <option value="price_asc">價格低至高</option>
              <option value="price_desc">價格高至低</option>
              <option value="usable_area_desc">面積大至小</option>
            </select>
            <button
              type="button"
              class="mobile-filter-button"
              aria-controls="property-filter-sheet"
              :aria-expanded="isFilterOpen"
              @click="openFilterSheet"
            >
              <span>更多</span>
              <AppIcon
                name="chevron-down"
                :size="16"
              />
              <span
                v-if="activeFilterCount > 0"
                class="mobile-filter-count"
              >{{ activeFilterCount }}</span>
            </button>
            <button
              v-if="activeFilterCount > 0"
              type="button"
              class="mobile-filter-reset"
              @click="clearFilters"
            >重設</button>
          </div>
        </div>

        <!-- 2.1 排序欄 -->
        <div class="sort-row listing-sort-row">
          <div class="listing-result-tools">
            <div class="listing-top-filters" aria-label="樓盤快速篩選">
              <span class="ft">住宅</span><span class="ft">車位</span><span class="ft">工業</span><span class="ft">商廈</span>
              <span class="listing-filter-divider" aria-hidden="true">｜</span>
              <span class="ft">業主</span><span class="ft">代理</span>
            </div>
            <span class="rn">{{ loading ? '讀取中' : `${pagination.total} 個結果` }}</span>
          </div>
          <select
            v-model="sortBy"
            class="ssel"
            @change="handleSortChange"
          >
            <option value="latest">最新更新</option>
            <option value="price_asc">價格低至高</option>
            <option value="price_desc">價格高至低</option>
            <option value="usable_area_desc">面積大至小</option>
          </select>
        </div>

        <p
          v-if="errorMessage"
          class="listing-state listing-state-error"
        >
          {{ errorMessage }}
        </p>
        <p
          v-else-if="loading && cards.length === 0"
          class="listing-state"
        >
          正在讀取樓盤。
        </p>
        <p
          v-else-if="!loading && cards.length === 0"
          class="listing-state"
        >
          暫時未有符合條件的樓盤。
        </p>

        <!-- 2.2 列表視圖 -->
        <div class="list-view">
          <div class="grid">
            <div
              v-for="card in cards"
              :key="card.id"
              class="gc"
              @click="handleCardClick(card.id)"
            >
              <!-- 2.2.1 類型堆疊 -->
              <div class="listing-card-type-stack">
                <span
                  v-for="(label, idx) in card.typeStack"
                  :key="idx"
                >{{ label }}</span>
              </div>
              <!-- 2.2.3 圖片區 -->
              <div
                class="gi pat"
              >
                <img
                  v-if="card.imageUrl"
                  :src="card.imageUrl"
                  :alt="card.title"
                  @error="hideFailedImage"
                >
                <svg
                  v-else
                  width="80"
                  height="60"
                  viewBox="0 0 80 60"
                  opacity=".2"
                >
                  <rect x="5" y="15" width="70" height="40" rx="1" stroke="#000" stroke-width="1.2" fill="none" />
                  <rect x="10" y="20" width="18" height="14" rx="1" stroke="#000" fill="none" />
                  <rect x="31" y="20" width="18" height="14" rx="1" stroke="#000" fill="none" />
                  <rect x="52" y="20" width="18" height="14" rx="1" stroke="#000" fill="none" />
                  <rect x="28" y="37" width="24" height="18" rx="1" stroke="#000" fill="none" />
                </svg>
              </div>
              <!-- 2.2.4 卡片內容 -->
              <div class="gb">
                <div class="gtags">
                  <span
                    v-for="(tag, idx) in card.tags"
                    :key="idx"
                    class="gtag"
                    :class="tag.dark ? 'dark' : ''"
                  >{{ tag.label }}</span>
                </div>
                <div class="gtitle">{{ card.title }}</div>
                <div class="g-location">{{ card.location }}</div>
                <div class="gsub">{{ card.sub }}</div>
                <div
                  v-if="card.agent"
                  class="listing-agent"
                >
                  <svg
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    aria-hidden="true"
                  >
                    <path
                      d="M4 21V7l8-4 8 4v14"
                      stroke-width="1.8"
                      stroke-linejoin="round"
                    />
                    <path
                      d="M9 21v-7h6v7M8 9h.01M12 9h.01M16 9h.01"
                      stroke-width="1.8"
                      stroke-linecap="round"
                    />
                  </svg>
                  {{ card.agent }}
                </div>
                <div class="gprice">
                  <span
                    class="listing-price-kind"
                    :class="card.priceKind"
                  >{{ card.priceKind === 'sale' ? '售' : '租' }}</span>{{ card.price }}<span v-if="card.priceUnit"> {{ card.priceUnit }}</span>
                </div>
                <div class="garea">
                  {{ card.area }} <span class="garea-price">{{ card.areaPrice }}</span>
                </div>
                <div
                  v-if="card.pills.length > 0"
                  class="gpills"
                >
                  <span
                    v-for="(pill, idx) in card.pills"
                    :key="idx"
                    class="gpill"
                  >{{ pill }}</span>
                </div>
              </div>
              <!-- 2.2.5 懸停操作按鈕 -->
              <div class="gc-actions">
                <button
                  type="button"
                  class="gc-action-btn"
                  @click.stop="toggleFavorite(card)"
                >{{ card.favorite ? '已收藏' : '收藏' }}</button>
                <button
                  type="button"
                  class="gc-action-btn"
                  @click.stop="openAppointment(card)"
                >睇樓</button>
                <button
                  type="button"
                  class="gc-action-btn primary"
                  @click.stop="handleCardClick(card.id)"
                >查看</button>
              </div>
            </div>

          </div>
        </div>

        <!-- 2.3 分頁 -->
        <div
          class="listing-pagination"
          aria-label="樓盤租售分頁"
        >
          <span class="listing-pagination-info">{{ resultRangeText }}</span>
          <div class="listing-pagination-actions">
            <button
              v-for="page in paginationPages"
              :key="page.key"
              type="button"
              class="listing-page-btn"
              :class="page.active ? 'on' : ''"
              :disabled="page.disabled"
              @click="handlePageSelect(page)"
            >{{ page.label }}</button>
          </div>
        </div>
      </main>

      <!-- 3. 右側廣告欄 -->
      <aside
        class="listing-ad-aside"
        aria-label="樓盤租售展示廣告"
      >
        <ListingSideAds channel="property_sale" />
      </aside>
    </div>
  </div>
</template>

<style scoped>
/* 1. 頁面容器 */
.page {
  width: 100%;
  min-height: calc(100svh - var(--nav-h, 52px));
  background: var(--sur-2);
}

/* 2. 三欄布局：對齊全局頁面寬度與左右留白 */
.lp {
  display: grid;
  grid-template-columns: 240px minmax(0, 1fr) 360px;
  justify-content: center;
  width: min(100%, var(--layout-page-max-width));
  margin: 0 auto;
  padding: 0 var(--layout-page-padding-inline);
  background: var(--sur-2);
  min-height: calc(100svh - var(--nav-h, 52px));
  align-items: stretch;
}

/* 3. 左側篩選欄 */
.lf {
  background: var(--sur);
  border-right: 1px solid var(--bdr);
  padding: 18px 18px;
  height: auto;
  min-height: calc(100svh - var(--nav-h, 52px));
  overflow: visible;
  scrollbar-width: none;
}

.lf::-webkit-scrollbar {
  display: none;
}

.filter-backdrop,
.filter-sheet-header,
.filter-sheet-actions,
.mobile-listing-controls {
  display: none;
}

/* 4. 搜尋欄（左側欄內嵌，負邊距撐滿） */
.lf .sbar {
  min-width: 0;
  margin-left: -14px;
  margin-right: -14px;
  margin-top: -18px;
  margin-bottom: 18px;
  padding: 18px 0 14px;
  background: var(--sur);
}

.lf .autocomplete-wrap,
.lf .sinput {
  min-width: 0;
  width: 100%;
}

.sbar {
  display: flex;
  gap: 6px;
  margin-bottom: 14px;
}

.sinput {
  flex: 1;
  border: 1px solid var(--bdr);
  padding: 8px 10px;
  font-size: var(--text-sm);
  font-family: var(--font);
  outline: none;
  border-radius: 2px;
  background: var(--sur);
  color: var(--ink);
}

.sinput:focus {
  border-color: var(--brand);
}

.sbtn {
  background: var(--brand);
  color: #fff;
  border: none;
  padding: 8px 12px;
  font-size: var(--text-sm);
  cursor: pointer;
  border-radius: 2px;
  font-family: var(--font);
}

.sbtn:hover {
  background: var(--brand-dark);
}

/* 5. 篩選群組 */
.fs {
  margin-bottom: var(--sp-4);
}

.ft-title {
  font-size: 9px;
  letter-spacing: 2px;
  color: var(--ink-3);
  text-transform: uppercase;
  margin-bottom: 6px;
}

.ftags {
  display: flex;
  flex-wrap: wrap;
  gap: 5px;
  max-width: 100%;
  overflow: hidden;
}

/* 6. 中間列表區 */
.lr {
  background: var(--sur-2);
  padding: 14px 14px 36px;
  min-width: 0;
  min-height: calc(100svh - var(--nav-h, 52px));
}

/* 7. 排序欄 */
.sort-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  width: 100%;
  margin-left: 0;
  margin-right: auto;
}

.listing-result-tools {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
  flex-wrap: wrap;
}

.rn {
  font-size: var(--text-sm);
  color: var(--ink-3);
}

.ssel {
  border: 1px solid var(--bdr);
  padding: 5px 8px;
  font-size: var(--text-sm);
  font-family: var(--font);
  outline: none;
  border-radius: 2px;
  background: var(--sur);
  color: var(--ink);
}

.listing-state {
  margin: 0 0 12px;
  border: 1px solid var(--bdr);
  border-radius: 4px;
  background: var(--sur);
  color: var(--ink-3);
  font-size: var(--text-sm);
  padding: 12px 14px;
}

.listing-state-error {
  border-color: rgba(186, 26, 26, 0.24);
  color: #ba1a1a;
}

/* 8. 頂部快速篩選（隱藏） */
.listing-top-filters {
  display: none;
}

/* 9. 列表視圖 */
.list-view {
  width: 100%;
  margin-left: 0;
  margin-right: auto;
}

.list-view > .grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
}

/* 10. 樓盤卡片 */
.list-view .gc {
  position: relative;
  display: grid;
  grid-template-columns: minmax(300px, 42%) minmax(0, 1fr);
  align-items: stretch;
  min-height: 246px;
  border: 1px solid var(--bdr);
  border-radius: var(--r-md);
  overflow: hidden;
  cursor: pointer;
  background: var(--sur);
  transition: box-shadow 0.15s, border-color 0.15s;
}

.list-view .gc:hover {
  box-shadow: var(--shadow-md);
  border-color: var(--brand-mid);
}

/* 11. 卡片圖片區 */
.list-view .gc .gi {
  grid-column: 1;
  grid-row: 1 / span 2;
  width: 100%;
  height: auto !important;
  min-height: 0;
  aspect-ratio: 600 / 360;
}

.list-view .gc .gi :deep(svg) {
  position: relative;
  z-index: 1;
  width: 36%;
  height: auto;
}

.gi {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f3f3f3;
}

.gi img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.gi.tall {
  min-height: 150px;
}

.gi.short {
  min-height: 90px;
}

/* 12. 斜紋底圖 */
.pat {
  position: relative;
  overflow: hidden;
}

.pat::after {
  content: '';
  position: absolute;
  inset: 0;
  background: repeating-linear-gradient(
    45deg,
    transparent,
    transparent 5px,
    rgba(0, 0, 0, 0.025) 5px,
    rgba(0, 0, 0, 0.025) 10px
  );
  pointer-events: none;
}

/* 13. 卡片內容區 */
.list-view .gc .gb {
  grid-column: 2;
  grid-row: 1;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  min-width: 0;
  padding: 28px 18px 10px;
}

.gb {
  padding: 11px 13px;
}

/* 14. 卡片標籤 */
.gtags {
  display: flex;
  gap: 5px;
  margin-bottom: 10px;
}

.gtag {
  font-size: var(--text-sm);
  padding: 3px 7px;
  letter-spacing: 0.8px;
  color: var(--ink-3);
  border: 1px solid var(--bdr);
  border-radius: 1px;
}

.gtag.dark {
  background: var(--brand);
  color: #fff;
  border-color: var(--brand);
  font-weight: 500;
}

/* 15. 卡片標題 */
.gtitle {
  font-size: var(--text-md);
  font-weight: 600;
  line-height: 1.35;
  margin-bottom: 3px;
  color: var(--ink);
}

.list-view .gc .gtitle {
  font-size: var(--text-md);
  font-weight: 600;
  line-height: 1.35;
}

/* 16. 位置 */
.g-location {
  margin: 2px 0 5px;
  color: var(--ink-2);
  font-size: var(--text-sm);
  font-weight: 700;
  line-height: 1.4;
}

/* 17. 副標題 */
.gsub {
  font-size: var(--text-sm);
  color: var(--ink-3);
  margin-bottom: 7px;
  line-height: 1.55;
}

.list-view .gc .gsub {
  font-size: var(--text-sm);
  line-height: 1.55;
}

/* 18. 代理公司 */
.listing-agent {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 6px;
  color: var(--ink-3);
  font-size: var(--text-sm);
  font-weight: 700;
  line-height: 1.4;
}

.listing-agent svg {
  width: 14px;
  height: 14px;
  flex: 0 0 auto;
  stroke: currentColor;
  fill: none;
}

/* 19. 價格 */
.gprice {
  display: flex;
  align-items: center;
  gap: 7px;
  flex-wrap: wrap;
  letter-spacing: 0;
  margin-top: 6px;
  font-size: var(--text-lg);
  font-weight: 300;
  color: var(--ink);
}

.list-view .gc .gprice {
  margin-top: 6px;
  font-size: 20px;
}

.gprice span {
  font-size: var(--text-sm);
  color: var(--ink-3);
  font-weight: 400;
}

/* 20. 售/租標識 */
.gprice .listing-price-kind {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 7px;
  background: var(--brand);
  color: #fff;
  font-size: var(--text-base);
  font-weight: 800;
  line-height: 1;
  box-shadow: 0 2px 6px rgba(240, 90, 0, 0.22);
}

.gprice .listing-price-kind.rent {
  background: #29B6E8;
  box-shadow: 0 2px 6px rgba(41, 182, 232, 0.22);
}

/* 21. 面積與呎價 */
.garea {
  margin-top: 6px;
  font-size: 10px;
  color: var(--ink-3);
  line-height: 1.4;
}

.garea .garea-price {
  display: inline;
  margin: 0 0 0 8px;
  font-size: 11px;
  color: var(--accent);
  font-weight: 600;
  line-height: 1.4;
  white-space: nowrap;
}

/* 22. 標籤藥丸 */
.gpills {
  display: flex;
  flex-wrap: wrap;
  gap: 3px;
  margin-top: 9px;
}

.list-view .gc .gpills {
  margin-top: 9px;
}

.gpill {
  font-size: 9px;
  padding: 2px 6px;
  background: var(--sur-2);
  color: var(--ink-2);
  border-radius: 2px;
}

/* 23. 類型堆疊標籤 */
.listing-card-type-stack {
  position: absolute;
  top: 10px;
  left: 10px;
  z-index: 2;
  display: grid;
  gap: 4px;
}

.listing-card-type-stack span {
  display: inline-flex;
  width: max-content;
  max-width: 90px;
  align-items: center;
  border-radius: 4px;
  background: rgba(26, 26, 26, 0.82);
  color: #fff;
  font-size: 10px;
  font-weight: 700;
  line-height: 1;
  padding: 5px 7px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.12);
}

.listing-card-type-stack span + span {
  background: var(--brand);
}

/* 24. 懸停操作按鈕 */
.gc-actions {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
  height: auto;
  max-height: 0;
  overflow: hidden;
  opacity: 0;
  pointer-events: none;
  transform: translateY(-2px);
  padding: 0 12px;
  transition: max-height 0.18s ease, opacity 0.15s ease, transform 0.15s ease, padding 0.18s ease;
}

.list-view .gc .gc-actions {
  grid-column: 2;
  grid-row: 2;
  align-self: end;
}

.gc:hover .gc-actions {
  max-height: 54px;
  opacity: 1;
  pointer-events: auto;
  transform: translateY(0);
  padding: 0 12px 12px;
}

.list-view .gc:hover .gc-actions {
  padding: 0 18px 14px;
}

.gc-action-btn {
  min-height: 34px;
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: #fff;
  color: var(--ink-2);
  cursor: pointer;
  font-family: var(--font);
  font-size: var(--text-sm);
  font-weight: 500;
}

.gc-action-btn:hover {
  border-color: var(--brand-mid);
  color: var(--brand);
}

.gc-action-btn.primary {
  border-color: var(--brand);
  background: var(--brand);
  color: #fff;
}

.gc-action-btn.primary:hover {
  background: var(--brand-dark);
  color: #fff;
}

/* 26. 自動補全下拉 */
.autocomplete-wrap {
  position: relative;
  flex: 1;
  min-width: 0;
}

.autocomplete-drop {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: var(--sur);
  border: 1px solid var(--bdr);
  border-top: none;
  border-radius: 0 0 2px 2px;
  z-index: 20;
  box-shadow: var(--shadow-md);
}

.ac-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 9px;
  font-size: var(--text-sm);
  color: var(--ink-2);
  cursor: pointer;
  border-bottom: 1px solid var(--bdr);
}

.ac-item:last-child {
  border-bottom: none;
}

.ac-item:hover {
  background: var(--sur-2);
}

.ac-icon {
  display: flex;
  align-items: center;
  color: var(--ink-3);
}

.ac-highlight {
  color: var(--brand);
  font-weight: 600;
}

/* 28. 分頁 */
.listing-pagination {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-top: 18px;
  padding: 12px 14px;
  border: 1px solid var(--bdr);
  border-radius: var(--r-md);
  background: var(--sur);
  width: 100%;
  margin-left: 0;
  margin-right: auto;
}

.listing-pagination-info {
  font-size: var(--text-sm);
  color: var(--ink-3);
}

.listing-pagination-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.listing-page-btn {
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: var(--sur);
  color: var(--ink-3);
  cursor: pointer;
  font-family: var(--font);
  font-size: var(--text-sm);
  padding: 7px 11px;
}

.listing-page-btn.on {
  border-color: var(--ink);
  color: var(--ink);
  font-weight: 500;
}

.listing-page-btn:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

/* 29. 右側廣告欄 */
.listing-ad-aside {
  position: sticky;
  top: var(--nav-h, 52px);
  align-self: start;
  overflow: visible;
  border-left: 1px solid var(--bdr);
  background: var(--sur-2);
  padding: 18px 18px 40px;
}

/* 34. 響應式 */
@media (max-width: 1200px) {
  .lp {
    grid-template-columns: 220px minmax(0, 1fr) 300px;
  }
}

@media (max-width: 1100px) {
  .lp {
    grid-template-columns: 1fr;
  }

  .lf {
    position: static;
    height: auto;
    padding: 14px;
  }

  .lf .sbar {
    margin-left: 0;
    margin-right: 0;
    margin-top: 0;
    padding: 0 0 14px;
  }

  .listing-ad-aside {
    position: static;
    height: auto;
    border-left: 0;
    border-top: 1px solid var(--bdr);
    padding: 14px;
  }

  .list-view,
  .listing-pagination,
  .sort-row {
    max-width: none;
  }
}

@media (max-width: 1023px) {
  .filter-backdrop {
    position: fixed;
    z-index: 120;
    inset: 0;
    display: block;
    width: 100%;
    height: 100%;
    border: 0;
    background: rgb(0 0 0 / 0.44);
    cursor: pointer;
    touch-action: none;
  }

  .lf {
    position: fixed;
    z-index: 121;
    top: auto;
    right: 0;
    bottom: 0;
    left: 0;
    display: flex;
    width: 100%;
    max-height: min(82svh, 720px);
    min-height: 0;
    flex-direction: column;
    border: 0;
    border-radius: 8px 8px 0 0;
    padding: 0;
    background: var(--sur);
    box-shadow: 0 -12px 32px rgb(0 0 0 / 0.18);
    opacity: 0;
    overflow-y: auto;
    pointer-events: none;
    touch-action: pan-y;
    transform: translateY(100%);
    transition: transform 0.2s ease, opacity 0.18s ease, visibility 0.2s;
    visibility: hidden;
  }

  .lf.is-filter-open {
    opacity: 1;
    pointer-events: auto;
    transform: translateY(0);
    visibility: visible;
  }

  .desktop-filter-search {
    display: none;
  }

  .filter-sheet-header {
    position: sticky;
    z-index: 2;
    top: 0;
    display: flex;
    min-height: 58px;
    align-items: center;
    justify-content: space-between;
    border-bottom: 1px solid var(--bdr);
    padding: 8px 12px 8px 14px;
    background: var(--sur);
  }

  .filter-sheet-header h2 {
    margin: 1px 0 0;
    color: var(--ink);
    font-size: 16px;
    font-weight: 700;
    letter-spacing: 0;
  }

  .filter-sheet-eyebrow {
    display: block;
    color: var(--ink-3);
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 0;
  }

  .filter-sheet-close {
    display: inline-flex;
    width: 44px;
    height: 44px;
    align-items: center;
    justify-content: center;
    border: 0;
    border-radius: 6px;
    background: transparent;
    color: var(--ink);
    cursor: pointer;
  }

  .filter-sections {
    padding: 10px 14px 2px;
  }

  .filter-sections .fs {
    margin-bottom: 10px;
  }

  .filter-sections .ft-title {
    margin-bottom: 4px;
    color: var(--ink-3);
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0;
  }

  .filter-sections .ftags {
    gap: 4px;
    overflow: visible;
  }

  .filter-sections :deep(.ft) {
    min-height: 36px;
    padding: 4px 10px;
  }

  .filter-sheet-actions {
    position: sticky;
    z-index: 2;
    bottom: 0;
    display: grid;
    grid-template-columns: minmax(0, 0.8fr) minmax(0, 1.4fr);
    gap: 8px;
    border-top: 1px solid var(--bdr);
    padding: 10px max(14px, calc(14px + var(--app-safe-right))) calc(10px + var(--app-safe-bottom)) max(14px, calc(14px + var(--app-safe-left)));
    background: var(--sur);
  }

  .filter-sheet-reset,
  .filter-sheet-apply {
    min-height: 44px;
    border-radius: 6px;
    cursor: pointer;
    font: inherit;
    font-size: 13px;
    font-weight: 700;
  }

  .filter-sheet-reset {
    border: 1px solid var(--bdr);
    background: var(--sur);
    color: var(--ink-2);
  }

  .filter-sheet-apply {
    border: 1px solid var(--brand);
    background: var(--brand);
    color: #fff;
  }

  .mobile-listing-controls {
    display: block;
    box-sizing: border-box;
    margin: -12px calc(-1 * var(--layout-page-padding-inline)) 12px;
    border-bottom: 1px solid var(--bdr);
    padding: 12px var(--layout-page-padding-inline) 10px;
    background: var(--sur);
  }

  .mobile-search-toolbar {
    margin-bottom: 10px;
  }

  .mobile-search-form {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 48px;
    min-width: 0;
  }

  .mobile-search-form .sinput {
    height: 48px;
    border-radius: 8px 0 0 8px;
    background: var(--sur);
    font-size: 16px;
  }

  .mobile-search-button {
    display: inline-flex;
    width: 48px;
    height: 48px;
    align-items: center;
    justify-content: center;
    border: 1px solid var(--brand);
    border-radius: 0 8px 8px 0;
    background: var(--brand);
    color: #fff;
    cursor: pointer;
  }

  .mobile-filter-rail {
    display: flex;
    gap: 8px;
    margin: 0 -2px;
    overflow-x: auto;
    overscroll-behavior-x: contain;
    padding: 2px;
    scrollbar-width: none;
  }

  .mobile-filter-rail::-webkit-scrollbar {
    display: none;
  }

  .mobile-filter-select,
  .mobile-filter-button,
  .mobile-filter-reset {
    min-height: 40px;
    border: 1px solid var(--bdr);
    border-radius: 999px;
    background: var(--sur);
    color: var(--ink-2);
    font: inherit;
    font-size: 13px;
  }

  .mobile-filter-select {
    width: auto;
    min-width: 98px;
    flex: 0 0 auto;
    padding: 0 30px 0 13px;
  }

  .mobile-sort-select {
    min-width: 108px;
  }

  .mobile-filter-button {
    display: inline-flex;
    flex: 0 0 auto;
    align-items: center;
    gap: 4px;
    padding: 0 12px;
    cursor: pointer;
  }

  .mobile-filter-reset {
    flex: 0 0 auto;
    border-color: transparent;
    padding: 0 4px;
    background: transparent;
    color: var(--ink-3);
    cursor: pointer;
  }

  .mobile-filter-count {
    display: inline-flex;
    min-width: 18px;
    height: 18px;
    align-items: center;
    justify-content: center;
    border-radius: 999px;
    background: var(--brand);
    color: #fff;
    font-size: 10px;
    line-height: 1;
  }

  .sort-row {
    align-items: stretch;
    flex-direction: column;
    gap: 6px;
    margin: 0 0 10px;
  }

  .sort-row .ssel {
    width: 100%;
    min-height: 42px;
    border: 1px solid var(--bdr);
    border-radius: 8px;
    padding: 0 12px;
    background: var(--sur);
    color: var(--ink);
    font: inherit;
    font-size: 14px;
  }

  .listing-sort-row > .ssel {
    display: none;
  }

  .mobile-search-form .autocomplete-drop {
    border-top: 1px solid var(--bdr);
    border-radius: 0 0 6px 6px;
  }

  .mobile-search-form .ac-item {
    width: 100%;
    min-height: 36px;
    border: 0;
    border-bottom: 1px solid var(--bdr);
    background: var(--sur);
    text-align: left;
  }

  .lr {
    padding: 0 0 36px;
  }
}

@media (max-width: 767px) {
  .listing-pagination {
    flex-direction: column;
    align-items: flex-start;
  }

  .sort-row {
    align-items: flex-start;
    flex-direction: column;
  }

  .listing-result-tools {
    align-items: flex-start;
    gap: 8px;
  }

  .list-view .gc {
    grid-template-columns: 1fr;
    min-height: 0;
  }

  .list-view .gc .gi,
  .list-view .gc .gb,
  .list-view .gc .gc-actions {
    grid-column: 1;
    grid-row: auto;
  }

  .list-view .gc .gb {
    padding: 14px;
  }

}

</style>
