<!--
 * 物業頻道樓盤詳情頁。
 * 1. 高保真還原 HTML 設計稿 #page-detail 雙欄布局。
 * 2. 左欄：標籤、標題、地址、價格、統計、設施、描述與大廈資料。
 * 3. 右欄：圖集、物業位置地圖、代理聯絡卡與其他樓盤推薦。
 * 4. CSS 變數全部使用 HTML 設計稿原生變量名（--brand、--ink、--sur、--bdr 等）。
 * 5. 使用後端樓盤接口展示公開詳情、圖片、聯絡、收藏與相似樓盤。
-->
<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { readStoredAccessToken } from '@/httpapis/auth-session';
import { createOrReusePropertyChat } from '@/httpapis/chats';
import {
  createPropertyAppointment,
  favoritePropertySale,
  fetchPropertySaleContactAccess,
  fetchPropertySaleDetail,
  fetchSimilarPropertySales,
  reportPropertySale,
  unfavoritePropertySale,
} from '@/httpapis/properties';
import type { ContactAccessResult, PropertyListingDetailResponse, PropertyListingSummaryResponse } from '@/model/property';
import {
  resolvePropertyArea,
  resolvePropertyCommunityName,
  resolvePropertyCoverImage,
  resolvePropertyDistrict,
  resolvePropertyImages,
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
const route = useRoute();
const listingId = computed(() => String(route.params.listingId ?? ''));
const listing = ref<PropertyListingDetailResponse | null>(null);
const similarItems = ref<PropertyListingSummaryResponse[]>([]);
const contactAccess = ref<ContactAccessResult | null>(null);
const loading = ref(false);
const actionMessage = ref('');
const errorMessage = ref('');
const appointmentOpen = ref(route.query.action === 'appointment');
const reportOpen = ref(false);
const openingChat = ref(false);

// 2. 物件標籤
interface PropertyTag {
  label: string;
  dark?: boolean;
}
const tags = computed<PropertyTag[]>(() => {
  if (!listing.value) {
    return [];
  }
  const role = listing.value.publisher_identity_type === 'agent' ? '代理盤' : '業主盤';
  return [
    { label: role, dark: true },
    { label: resolvePropertyDistrict(listing.value, 'zh-HK') },
    { label: resolvePropertyTypeLabel(listing.value, 'zh-HK') },
  ];
});

// 3. 物件統計資料
interface DetailStat {
  value: string;
  label: string;
}
const stats = computed<DetailStat[]>(() => {
  const sale = listing.value?.property_sale;
  return [
    { value: resolvePropertyArea(listing.value as PropertyListingSummaryResponse || emptyListing()).toLocaleString('zh-HK'), label: '實用呎數' },
    { value: String(sale?.bedroom_count ?? 0), label: '睡房' },
    { value: String(sale?.bathroom_count ?? 0), label: '浴室' },
    { value: sale?.floor_level || '-', label: '樓層' },
  ];
});

// 4. 設施配套
const facilities = computed(() => listing.value ? resolvePropertyTagLabels(listing.value, 'zh-HK', 12) : []);

// 5. 大廈資料
interface BuildingInfo {
  label: string;
  value: string;
}
const buildingInfo = computed<BuildingInfo[]>(() => {
  const sale = listing.value?.property_sale;
  return [
    { label: '落成年份', value: sale?.completion_year ? `${sale.completion_year}年` : '-' },
    { label: '樓層數目', value: sale?.building_total_floors ? `${sale.building_total_floors}層` : sale?.total_floors ? `${sale.total_floors}層` : '-' },
    { label: '管理公司', value: sale?.management_company || '-' },
    { label: '管理費', value: sale?.management_fee_hkd ? `${formatHKD(sale.management_fee_hkd)}/月` : '-' },
    { label: '瀏覽', value: String(sale?.view_count ?? 0) },
    { label: '查詢', value: String(sale?.inquiry_count ?? 0) },
  ];
});

// 6. 圖集資料
interface GalleryImage {
  id: string;
  background: string;
  url?: string;
}
const galleryImages = computed<GalleryImage[]>(() => {
  if (!listing.value) {
    return [{ id: 'placeholder', background: 'linear-gradient(160deg,#e8e8e8,#d0d0d0)' }];
  }
  const images = resolvePropertyImages(listing.value);
  if (images.length === 0) {
    return [{ id: 'placeholder', background: 'linear-gradient(160deg,#e8e8e8,#d0d0d0)' }];
  }
  return images.map((image) => ({ id: image.id, background: '#f3f3f3', url: image.url }));
});
const selectedImageIndex = ref(0);
const extraImageCount = computed(() => Math.max(0, galleryImages.value.length - 4));

// 7. 其他樓盤推薦
interface SimilarListing {
  id: string;
  name: string;
  price: string;
  meta: string;
  background: string;
  imageUrl?: string;
}
const similarListings = computed<SimilarListing[]>(() =>
  similarItems.value.map((item) => ({
    id: item.listing_id,
    name: resolvePropertyTitle(item),
    price: resolvePropertyPriceText(item, 'zh-HK'),
    meta: `${resolvePropertyDistrict(item, 'zh-HK')} · ${resolvePropertyArea(item).toLocaleString('zh-HK')}呎`,
    background: 'linear-gradient(135deg,#e4dcd8,#ccc0bc)',
    imageUrl: resolvePropertyCoverImage(item)?.url,
  })),
);

const title = computed(() => listing.value ? resolvePropertyTitle(listing.value) : '');
const address = computed(() => listing.value?.property_sale?.address_text || '');
const priceText = computed(() => listing.value ? resolvePropertyPriceText(listing.value, 'zh-HK') : '-');
const priceSuffix = computed(() => resolvePropertyTransactionType(listing.value || emptyListing()) === 'rent' ? ' / 月' : '');
const unitPriceText = computed(() => {
  if (!listing.value) {
    return '-';
  }
  const area = resolvePropertyArea(listing.value);
  const price = resolvePropertyPrice(listing.value);
  return area > 0 && price > 0 ? `約 ${formatHKD(Math.round(price / area))} / 呎` : '-';
});
const mapSrc = computed(() => {
  const sale = listing.value?.property_sale;
  if (sale?.latitude && sale?.longitude) {
    return `https://www.google.com/maps?q=${sale.latitude},${sale.longitude}&output=embed`;
  }
  return `https://www.google.com/maps?q=${encodeURIComponent(address.value || resolvePropertyCommunityName(listing.value || emptyListing()))}&output=embed`;
});
const ownerName = computed(() =>
  listing.value?.property_sale?.agency_company_name ||
  listing.value?.owner?.display_name ||
  listing.value?.property_sale?.publisher_role_label ||
  '發布者',
);
const contactPhone = computed(() => contactAccess.value?.contact_payload?.phone || '');
const whatsappURL = computed(() => contactAccess.value?.contact_payload?.whatsapp_url || '');
const appointmentForm = reactive({
  contactName: '',
  contactPhone: '',
  preferredTime: '',
  message: '',
});
const reportForm = reactive({
  reason: 'incorrect_info',
  message: '',
});

// 8. 選擇圖片
const selectImage = (index: number): void => {
  selectedImageIndex.value = index;
};

// 9. 跳轉其他樓盤
const goSimilar = (id: string): void => {
  void router.push(`/properties/${id}`);
};

// 10. 跳轉站內聊天
const goChat = async (): Promise<void> => {
  if (!listing.value || openingChat.value) {
    return;
  }
  if (!readStoredAccessToken()) {
    await router.push({ path: '/login', query: { redirect: route.fullPath } });
    return;
  }

  openingChat.value = true;

  try {
    const { data } = await createOrReusePropertyChat('sale', listing.value.listing_id);
    await router.push(`/account/chat/${data.data.chat_id}`);
  } catch {
    actionMessage.value = '暫時無法開啟站內訊息。';
  } finally {
    openingChat.value = false;
  }
};

// 11. 載入詳情
const loadDetail = async (): Promise<void> => {
  loading.value = true;
  errorMessage.value = '';
  try {
    const { data } = await fetchPropertySaleDetail(listingId.value);
    listing.value = data.data;
    selectedImageIndex.value = 0;
    const similar = await fetchSimilarPropertySales(listingId.value, { limit: 8 });
    similarItems.value = similar.data.data.items;
  } catch {
    errorMessage.value = '暫時無法讀取樓盤詳情。';
  } finally {
    loading.value = false;
  }
};

// 12. 切換收藏
const toggleFavorite = async (): Promise<void> => {
  if (!listing.value) return;
  if (!readStoredAccessToken()) {
    await router.push({ path: '/login', query: { redirect: route.fullPath } });
    return;
  }
  if (listing.value.is_favorite) {
    await unfavoritePropertySale(listing.value.listing_id);
    listing.value = { ...listing.value, is_favorite: false };
    return;
  }
  await favoritePropertySale(listing.value.listing_id);
  listing.value = { ...listing.value, is_favorite: true };
};

// 13. 讀取聯絡方式
const revealContact = async (): Promise<void> => {
  if (!listing.value) return;
  if (!readStoredAccessToken()) {
    await router.push({ path: '/login', query: { redirect: route.fullPath } });
    return;
  }
  try {
    const { data } = await fetchPropertySaleContactAccess(listing.value.listing_id);
    contactAccess.value = data.data;
    actionMessage.value = '已顯示聯絡方式。';
  } catch {
    actionMessage.value = '暫時無法讀取聯絡方式。';
  }
};

// 14. 提交睇樓預約
const submitAppointment = async (): Promise<void> => {
  if (!listing.value) return;
  if (!readStoredAccessToken()) {
    await router.push({ path: '/login', query: { redirect: route.fullPath } });
    return;
  }
  try {
    await createPropertyAppointment(listing.value.listing_id, {
      contact_name: appointmentForm.contactName.trim(),
      contact_phone: appointmentForm.contactPhone.trim(),
      preferred_time: appointmentForm.preferredTime.trim() || undefined,
      message: appointmentForm.message.trim() || undefined,
      appointment_type: 'viewing',
    });
    appointmentOpen.value = false;
    actionMessage.value = '睇樓預約已提交。';
  } catch {
    actionMessage.value = '睇樓預約提交失敗。';
  }
};

// 15. 提交舉報
const submitReport = async (): Promise<void> => {
  if (!listing.value) return;
  if (!readStoredAccessToken()) {
    await router.push({ path: '/login', query: { redirect: route.fullPath } });
    return;
  }
  try {
    await reportPropertySale(listing.value.listing_id, {
      reason: reportForm.reason,
      message: reportForm.message.trim() || undefined,
    });
    reportOpen.value = false;
    actionMessage.value = '舉報已提交。';
  } catch {
    actionMessage.value = '舉報提交失敗。';
  }
};

// 16. 格式化港幣
const formatHKD = (value: number): string =>
  `HK$${value.toLocaleString('zh-HK')}`;

// 17. 空資料 fallback
const emptyListing = (): PropertyListingSummaryResponse => ({
  listing_id: '',
  module: 'property_sale',
  title: '',
  summary: '',
  district_code: '',
  publisher_identity_type: 'owner',
  publication_status: 'draft',
  business_status: 'available',
  updated_at: '',
});

onMounted(() => {
  void loadDetail();
});
</script>

<template>
  <main class="detail-page">
    <!-- 1. 麵包屑 -->
    <div class="breadcrumb">
      <router-link class="bc-link" to="/">首頁</router-link>
      <span class="bc-sep">›</span>
      <router-link class="bc-link" to="/properties">樓盤租售</router-link>
      <span class="bc-sep">›</span>
      <span class="bc-current">{{ title || '樓盤詳情' }}</span>
    </div>

    <div class="detail-wrap">
      <p
        v-if="errorMessage"
        class="detail-state detail-state-error"
      >
        {{ errorMessage }}
      </p>
      <p
        v-else-if="loading && !listing"
        class="detail-state"
      >
        正在讀取樓盤詳情。
      </p>
      <p
        v-if="actionMessage"
        class="detail-state"
      >
        {{ actionMessage }}
      </p>
      <div class="detail-main">
      <!-- 2. 雙欄布局 -->
      <div
        v-if="listing"
        class="detail-layout"
      >
          <!-- 2.1 左欄：物件資訊 -->
          <div class="detail-left">
            <div class="gtags">
              <span
                v-for="tag in tags"
                :key="tag.label"
                class="gtag"
                :class="{ dark: tag.dark }"
              >
                {{ tag.label }}
              </span>
            </div>

            <h2 class="detail-title">{{ title }}</h2>
            <div class="detail-address">{{ address }}</div>

            <div class="detail-price-row">
              <div class="detail-price-main">
                {{ priceText }}<span class="detail-price-suffix">{{ priceSuffix }}</span>
              </div>
              <div class="detail-price-unit">{{ unitPriceText }}</div>
            </div>

            <div class="detail-stats">
              <div
                v-for="stat in stats"
                :key="stat.label"
                class="detail-stat"
              >
                <div class="detail-stat-val">{{ stat.value }}</div>
                <div class="detail-stat-label">{{ stat.label }}</div>
              </div>
            </div>

            <section class="detail-section">
              <div class="detail-section-title">設施配套</div>
              <div class="gpills">
                <span
                  v-for="item in facilities"
                  :key="item"
                  class="gpill"
                >
                  {{ item }}
                </span>
                <span
                  v-if="facilities.length === 0"
                  class="gpill"
                >
                  未提供
                </span>
              </div>
            </section>

            <section class="detail-section">
              <div class="detail-section-title">物業描述</div>
              <p class="body-text">
                {{ listing.description || listing.summary || '未提供物業描述。' }}
              </p>
            </section>

            <section class="detail-section">
              <div class="detail-section-title detail-section-title--label">
                大廈資料
              </div>
              <div class="binfo-grid">
                <div
                  v-for="info in buildingInfo"
                  :key="info.label"
                  class="binfo-card"
                >
                  <div class="binfo-label">{{ info.label }}</div>
                  <div class="binfo-val">{{ info.value }}</div>
                </div>
              </div>
            </section>
          </div>

          <!-- 2.2 右欄：圖集、地圖、代理、推薦 -->
          <div class="detail-right">
            <!-- 2.2.1 圖集 -->
            <div class="detail-right-card">
              <div class="detail-gallery">
                <div
                  class="detail-main-img pat"
                  :style="{ background: galleryImages[selectedImageIndex].background }"
                >
                  <img
                    v-if="galleryImages[selectedImageIndex].url"
                    :src="galleryImages[selectedImageIndex].url"
                    :alt="title"
                  >
                  <svg
                    v-else
                    width="120"
                    height="80"
                    viewBox="0 0 120 80"
                    opacity=".15"
                  >
                    <rect
                      x="5"
                      y="10"
                      width="110"
                      height="60"
                      rx="1"
                      stroke="#000"
                      stroke-width="1.5"
                      fill="none"
                    />
                    <rect
                      x="10"
                      y="15"
                      width="30"
                      height="22"
                      rx="1"
                      stroke="#000"
                      fill="none"
                    />
                    <rect
                      x="45"
                      y="15"
                      width="30"
                      height="22"
                      rx="1"
                      stroke="#000"
                      fill="none"
                    />
                    <rect
                      x="80"
                      y="15"
                      width="30"
                      height="22"
                      rx="1"
                      stroke="#000"
                      fill="none"
                    />
                    <rect
                      x="35"
                      y="42"
                      width="50"
                      height="28"
                      rx="1"
                      stroke="#000"
                      fill="none"
                    />
                  </svg>
                </div>
                <div class="detail-thumb-row">
                  <div
                    v-for="(image, index) in galleryImages.slice(1, 4)"
                    :key="image.id"
                    class="detail-thumb pat"
                    :style="{ background: image.background }"
                    @click="selectImage(index + 1)"
                  />
                  <div
                    v-if="extraImageCount > 0"
                    class="detail-thumb detail-thumb-more pat"
                    :style="{ background: galleryImages[4]?.background || '#f3f3f3' }"
                    @click="selectImage(4)"
                  >
                    +{{ extraImageCount }}
                  </div>
                </div>
              </div>
            </div>

            <!-- 2.2.2 物業位置 -->
            <div class="detail-right-card">
              <div class="label-text" style="margin-bottom: 10px;">
                物業位置
              </div>
              <iframe
                class="building-map-frame"
                title="樓盤地圖位置"
                :src="mapSrc"
                allowfullscreen
                loading="lazy"
                referrerpolicy="no-referrer-when-downgrade"
              />
            </div>

            <!-- 2.2.3 代理聯絡卡 -->
            <div class="detail-agent-card">
              <div class="detail-agent-profile">
                <div class="detail-agent-avatar">{{ ownerName.slice(0, 1) }}</div>
                <div class="detail-agent-copy">
                  <div class="detail-agent-kicker">{{ listing.publisher_identity_type === 'agent' ? '代理人' : '發布者' }}</div>
                  <div class="detail-agent-name">{{ ownerName }}</div>
                  <div class="detail-agent-sub">查看聯絡方式、預約睇樓或發送站內訊息</div>
                </div>
              </div>
              <div class="detail-agent-time">
                <div class="detail-agent-label">發布資料</div>
                <div class="detail-agent-value">{{ resolvePropertyCommunityName(listing) }}</div>
                <div class="detail-agent-note">{{ resolvePropertyRooms(listing) }}</div>
              </div>
              <div class="detail-agent-actions">
                <a
                  v-if="contactPhone"
                  class="detail-agent-contact"
                  :href="`tel:${contactPhone}`"
                  aria-label="代理電話"
                >
                  <svg
                    viewBox="0 0 24 24"
                    fill="none"
                    aria-hidden="true"
                  >
                    <path
                      d="M22 16.92v3a2 2 0 0 1-2.18 2A19.8 19.8 0 0 1 11.19 19 19.5 19.5 0 0 1 5 12.81 19.8 19.8 0 0 1 2.08 4.18 2 2 0 0 1 4.06 2h3a2 2 0 0 1 2 1.72c.12.92.33 1.82.62 2.68a2 2 0 0 1-.45 2.11L8 9.73a16 16 0 0 0 6.27 6.27l1.22-1.23a2 2 0 0 1 2.11-.45c.86.29 1.76.5 2.68.62A2 2 0 0 1 22 16.92Z"
                      stroke-width="1.8"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                    />
                  </svg>
                  致電查詢
                </a>
                <button
                  v-else
                  class="detail-agent-contact"
                  type="button"
                  @click="revealContact"
                >
                  查看電話
                </button>
                <a
                  v-if="whatsappURL"
                  class="detail-agent-contact primary"
                  :href="whatsappURL"
                  target="_blank"
                  rel="noopener"
                >
                  <svg
                    viewBox="0 0 24 24"
                    fill="none"
                    aria-hidden="true"
                  >
                    <path
                      d="M7.2 20.2 3 21l.9-4A9 9 0 1 1 7.2 20.2Z"
                      stroke-width="1.8"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                    />
                    <path
                      d="M9.2 8.7c.2-.4.4-.5.7-.5h.5c.2 0 .4.1.5.4l.6 1.4c.1.3.1.5-.1.7l-.4.5c.7 1.2 1.6 2.1 2.8 2.8l.5-.4c.2-.2.4-.2.7-.1l1.4.6c.3.1.4.3.4.5v.5c0 .3-.1.5-.5.7-.6.3-1.4.3-2.4-.1-2.4-.8-4.3-2.7-5.1-5.1-.4-1-.4-1.8-.1-2.4Z"
                      stroke-width="1.8"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                    />
                  </svg>
                  WhatsApp
                </a>
                <button
                  v-else
                  class="detail-agent-contact primary"
                  type="button"
                  @click="revealContact"
                >
                  查看 WhatsApp
                </button>
                <button
                  v-if="listing.contact_summary.show_chat"
                  class="detail-agent-contact"
                  type="button"
                  :disabled="openingChat"
                  @click="goChat"
                >
                  <svg
                    viewBox="0 0 24 24"
                    fill="none"
                    aria-hidden="true"
                  >
                    <path
                      d="M21 11.5a8.4 8.4 0 0 1-9 8.4 8.8 8.8 0 0 1-3.8-.9L3 20l1.1-4.6a8.4 8.4 0 1 1 16.9-3.9Z"
                      stroke-width="1.8"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                    />
                    <path
                      d="M8 11h8M8 14h5"
                      stroke-width="1.8"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                    />
                  </svg>
                  {{ openingChat ? '開啟中' : '站內訊息' }}
                </button>
              </div>
              <div class="detail-agent-actions">
                <button
                  class="detail-agent-contact"
                  type="button"
                  @click="toggleFavorite"
                >
                  {{ listing.is_favorite ? '已收藏' : '收藏' }}
                </button>
                <button
                  class="detail-agent-contact"
                  type="button"
                  @click="appointmentOpen = !appointmentOpen"
                >
                  預約睇樓
                </button>
                <button
                  class="detail-agent-contact"
                  type="button"
                  @click="reportOpen = !reportOpen"
                >
                  舉報
                </button>
              </div>
              <form
                v-if="appointmentOpen"
                class="detail-form"
                @submit.prevent="submitAppointment"
              >
                <input v-model="appointmentForm.contactName" placeholder="聯絡人" required>
                <input v-model="appointmentForm.contactPhone" placeholder="電話" required>
                <input v-model="appointmentForm.preferredTime" placeholder="希望睇樓時間">
                <textarea v-model="appointmentForm.message" placeholder="補充資料" rows="3" />
                <button type="submit">提交預約</button>
              </form>
              <form
                v-if="reportOpen"
                class="detail-form"
                @submit.prevent="submitReport"
              >
                <select v-model="reportForm.reason">
                  <option value="incorrect_info">資料不準確</option>
                  <option value="unavailable">樓盤已不可用</option>
                  <option value="suspicious">可疑內容</option>
                  <option value="other">其他</option>
                </select>
                <textarea v-model="reportForm.message" placeholder="補充說明" rows="3" />
                <button type="submit">提交舉報</button>
              </form>
            </div>

            <!-- 2.2.4 其他樓盤推薦 -->
            <div class="similar-section">
              <div class="label-text" style="margin-bottom: 12px;">
                其他樓盤
              </div>
              <div class="similar-scroll">
                <div
                  v-for="item in similarListings"
                  :key="item.id"
                  class="sim-card"
                  @click="goSimilar(item.id)"
                >
                  <div
                    class="sim-img pat"
                    :style="{ background: item.background }"
                  >
                    <img
                      v-if="item.imageUrl"
                      :src="item.imageUrl"
                      :alt="item.name"
                    >
                  </div>
                  <div class="sim-body">
                    <div class="sim-name">{{ item.name }}</div>
                    <div class="sim-price">{{ item.price }}</div>
                    <div class="sim-meta">{{ item.meta }}</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </main>
</template>

<style scoped>
/*
 * 樓盤詳情頁樣式。
 * 1. 嚴格對齊 HTML 設計稿 #page-detail 結構與數值。
 * 2. CSS 變數全部使用 HTML 設計稿原生變量名（--brand、--ink、--sur、--bdr、--sur-2 等）。
 * 3. 雙欄響應式：桌面雙欄、行動單欄。
 */

.detail-page {
  display: block;
  width: 100%;
  min-height: calc(100vh - var(--nav-h));
  background: var(--sur-2);
}

/* 1. 容器布局 */
.detail-wrap {
  display: block;
  max-width: var(--layout-page-max-width);
  min-height: auto;
  margin: 0 auto;
  padding: var(--sp-5);
}

.detail-main {
  max-width: 100%;
  border-right: 0;
  background: var(--sur-2);
  padding: 0;
  overflow: visible;
}

/* 2. 雙欄布局 */
.detail-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: var(--sp-4);
  align-items: start;
  margin-top: var(--sp-4);
}

/* 3. 卡片共用樣式 */
.detail-left,
.detail-right-card,
.detail-agent-card,
.similar-section {
  border: 1px solid var(--bdr);
  border-radius: var(--r-md);
  background: #fff;
  padding: var(--sp-4);
}

.detail-right-card,
.detail-agent-card,
.similar-section {
  margin-top: 14px;
}

.detail-right-card:first-of-type {
  margin-top: 0;
}

.detail-left {
  position: sticky;
  top: 72px;
}

/* 4. 麵包屑 */
.breadcrumb {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  width: 100%;
  margin: 0;
  padding: 12px max(24px, calc((100vw - 1180px) / 2 + 24px));
  border: 1px solid var(--bdr);
  border-right: 0;
  border-left: 0;
  background: var(--sur);
  color: var(--ink-3);
  font-family: var(--font);
  font-size: 12px;
  line-height: 1.5;
}

.bc-link {
  border: 0;
  background: transparent;
  color: var(--ink-3);
  font-family: var(--font);
  font-size: 12px;
  padding: 0;
  text-decoration: none;
  cursor: pointer;
}

.bc-link:hover {
  color: var(--brand);
}

.bc-sep {
  color: var(--ink-4);
}

.bc-current {
  color: var(--ink);
  font-weight: 600;
}

/* 5. 共用花紋佔位 */
.pat {
  position: relative;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 6. 標籤 */
.gtags {
  display: flex;
  flex-wrap: wrap;
  gap: 5px;
  margin-bottom: 10px;
}

.gtag {
  font-size: 11px;
  letter-spacing: 0.8px;
  color: var(--ink-3);
  border: 1px solid var(--bdr);
  padding: 3px 7px;
  border-radius: 1px;
}

.gtag.dark {
  background: var(--brand);
  color: #fff;
  border-color: var(--brand);
  font-weight: 500;
}

/* 7. 標題與地址 */
.detail-title {
  margin: 0 0 6px;
  font-family: var(--font);
  font-size: 28px;
  font-weight: 500;
  line-height: 1.25;
  color: var(--ink);
}

.detail-address {
  margin-bottom: 16px;
  color: var(--ink-3);
  font-size: 14px;
  line-height: 1.6;
}

/* 8. 價格列 */
.detail-price-row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--g2);
}

.detail-price-main {
  font-size: 34px;
  font-weight: 400;
  letter-spacing: 0;
  color: var(--ink);
  line-height: 1.2;
}

.detail-price-suffix {
  font-size: 13px;
  color: var(--ink-3);
  font-weight: 400;
}

.detail-price-unit {
  font-size: 14px;
  color: var(--ink-3);
}

/* 9. 統計區 */
.detail-stats {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
  margin-bottom: 16px;
}

.detail-stat {
  text-align: center;
  padding: 12px 8px;
  background: var(--sur-2);
  border-radius: 2px;
}

.detail-stat-val {
  font-size: 21px;
  font-weight: 600;
  line-height: 1.15;
  margin-bottom: 2px;
  color: var(--ink);
}

.detail-stat-label {
  margin-top: 5px;
  font-size: 12px;
  color: var(--ink-3);
}

/* 10. 區段 */
.detail-section {
  border-top: 1px solid var(--g2);
  padding-top: 16px;
  margin-top: 16px;
}

.detail-section-title {
  font-size: 16px;
  font-weight: 700;
  margin-bottom: 12px;
  color: var(--ink);
}

.detail-section-title--label {
  font-size: 16px;
  font-weight: 700;
  letter-spacing: 1.5px;
  text-transform: uppercase;
  color: var(--ink-3);
}

/* 10.1 右欄 label-text 小標題（全域樣式，未被 detail-section 覆蓋） */
.label-text {
  font-size: 9px;
  letter-spacing: 1.5px;
  text-transform: uppercase;
  color: var(--ink-3);
  font-weight: 500;
}

/* 11. 設施標籤 */
.gpills {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 6px;
}

.gpill {
  font-size: 12px;
  padding: 5px 9px;
  background: var(--sur-2);
  color: var(--ink-2);
  border-radius: 2px;
}

/* 12. 描述文字 */
.body-text {
  margin: 0;
  font-size: 15px;
  line-height: 1.85;
  color: var(--ink-2);
}

/* 13. 大廈資料網格 */
.binfo-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  margin-top: 12px;
}

.binfo-card {
  border: 1px solid var(--g2);
  border-radius: 3px;
  padding: 12px 14px;
  background: var(--sur-2);
}

.binfo-label {
  font-size: 12px;
  letter-spacing: 1px;
  text-transform: uppercase;
  color: var(--ink-3);
  margin-bottom: 4px;
}

.binfo-val {
  font-size: 15px;
  font-weight: 600;
  line-height: 1.45;
  color: var(--ink);
}

/* 14. 圖集 */
.detail-gallery {
  display: grid;
  gap: 8px;
}

.detail-main-img {
  width: 100%;
  height: 300px;
  border-radius: 7px;
  cursor: pointer;
  margin-bottom: 8px;
  background: #f3f3f3;
}

.detail-main-img img,
.detail-thumb img,
.sim-img img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.detail-thumb-row {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 8px;
}

.detail-thumb {
  height: 86px;
  border-radius: 6px;
  cursor: pointer;
  background: #f3f3f3;
}

.detail-thumb-more {
  color: var(--ink);
  font-size: 16px;
  font-weight: 700;
}

/* 15. 地圖 */
.building-map-frame {
  display: block;
  width: 100%;
  height: 230px;
  border: 0;
  border-radius: var(--r-md);
  background: var(--sur-2);
}

/* 16. 代理聯絡卡 */
.detail-agent-card {
  display: grid;
  grid-template-columns: 1fr;
  gap: 14px;
  align-items: start;
}

.detail-agent-profile {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
  padding-bottom: 14px;
  border-bottom: 1px solid var(--sur-3);
}

.detail-agent-avatar {
  display: flex;
  width: 46px;
  height: 46px;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: var(--brand-light);
  color: var(--brand);
  font-size: 15px;
  font-weight: 800;
  flex: 0 0 auto;
}

.detail-agent-copy {
  min-width: 0;
}

.detail-agent-kicker {
  margin-bottom: 3px;
  color: var(--ink-3);
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 1.2px;
  text-transform: uppercase;
}

.detail-agent-name {
  color: var(--ink);
  font-size: 16px;
  font-weight: 700;
  line-height: 1.25;
}

.detail-agent-sub {
  margin-top: 4px;
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 500;
  line-height: 1.45;
}

.detail-agent-time {
  display: grid;
  gap: 4px;
  border: 1px solid var(--bdr);
  border-radius: 7px;
  background: var(--sur-2);
  padding: 12px;
}

.detail-agent-label {
  color: var(--ink-3);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.8px;
}

.detail-agent-value {
  color: var(--ink);
  font-size: 14px;
  font-weight: 700;
  line-height: 1.35;
}

.detail-agent-note {
  color: var(--ink-3);
  font-size: 12px;
  line-height: 1.45;
}

.detail-agent-actions {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}

.detail-agent-contact {
  display: inline-flex;
  min-width: 0;
  min-height: 40px;
  align-items: center;
  justify-content: center;
  gap: 7px;
  border: 1px solid var(--bdr);
  border-radius: 7px;
  background: #fff;
  color: var(--ink);
  font-family: inherit;
  font-size: 12px;
  font-weight: 700;
  text-decoration: none;
  white-space: nowrap;
  padding: 0 10px;
  cursor: pointer;
  transition: border-color 0.15s ease, color 0.15s ease, background 0.15s ease;
}

.detail-agent-contact svg {
  width: 15px;
  height: 15px;
  flex: 0 0 auto;
  stroke: currentColor;
}

.detail-agent-contact:hover {
  border-color: var(--brand-mid);
  color: var(--brand);
}

.detail-agent-contact.primary {
  border-color: var(--brand);
  background: var(--brand);
  color: #fff;
}

.detail-agent-contact.primary:hover {
  border-color: var(--brand-dark);
  background: var(--brand-dark);
  color: #fff;
}

.detail-form {
  display: grid;
  gap: 8px;
  border-top: 1px solid var(--sur-3);
  padding-top: 12px;
}

.detail-form input,
.detail-form select,
.detail-form textarea {
  width: 100%;
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: #fff;
  color: var(--ink);
  font-family: inherit;
  font-size: 13px;
  padding: 9px 10px;
  outline: none;
}

.detail-form button {
  min-height: 38px;
  border: 1px solid var(--brand);
  border-radius: 6px;
  background: var(--brand);
  color: #fff;
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  font-weight: 700;
}

.detail-state {
  margin: 0 0 12px;
  border: 1px solid var(--bdr);
  border-radius: 4px;
  background: var(--sur);
  color: var(--ink-3);
  font-size: 13px;
  padding: 12px 14px;
}

.detail-state-error {
  border-color: rgba(186, 26, 26, 0.24);
  color: #ba1a1a;
}

/* 17. 其他樓盤推薦 */
.similar-section {
  padding: 14px;
}

.similar-scroll {
  display: flex;
  gap: 10px;
  max-width: 100%;
  overflow-x: auto;
  overflow-y: hidden;
  min-width: 0;
  padding-bottom: 4px;
  scrollbar-width: thin;
}

.similar-scroll::-webkit-scrollbar {
  height: 6px;
}

.similar-scroll::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: var(--bdr-2);
}

.sim-card {
  flex: 0 0 138px;
  min-width: 0;
  overflow: hidden;
}

.sim-img {
  height: 76px;
  border-radius: 7px 7px 0 0;
  background: #f3f3f3;
}

.sim-body {
  min-width: 0;
}

.sim-name,
.sim-price,
.sim-meta {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sim-name {
  font-size: 11px;
  font-weight: 500;
  margin-bottom: 3px;
}

.sim-price {
  font-size: 12px;
  font-weight: 500;
  color: var(--brand);
}

.sim-meta {
  margin-top: 2px;
  font-size: 10px;
  color: var(--ink-3);
}

/* 18. 響應式 */
@media (max-width: 900px) {
  .breadcrumb {
    padding-right: 24px;
    padding-left: 24px;
  }

  .detail-wrap {
    padding: var(--sp-4);
  }

  .detail-layout {
    grid-template-columns: 1fr;
  }

  .detail-left {
    position: static;
  }

  .detail-agent-actions {
    grid-template-columns: 1fr;
  }

  .similar-scroll {
    overflow-x: auto;
  }
}

@media (max-width: 560px) {
  .breadcrumb {
    padding-right: 16px;
    padding-left: 16px;
  }

  .detail-stats {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .binfo-grid {
    grid-template-columns: 1fr;
  }

  .detail-main-img {
    height: 220px;
  }

  .detail-thumb {
    height: 64px;
  }

  .detail-price-main {
    font-size: 28px;
  }
}
</style>
