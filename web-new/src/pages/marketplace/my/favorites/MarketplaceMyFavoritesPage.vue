<!--
 * 二手交易收藏頁 - 高保真還原 HTML 設計稿。
 * 1. 靜態 mock 收藏列表，支援排序、移除、清除全部與備註編輯。
 * 2. 收藏卡片網格佈局，含目標價提醒 Modal。
 * 3. 使用 router 跳轉至樓盤詳情與搜尋頁。
 * 4. HTML 設計稿 token 已映射為專案 token（rgb(var(--color-xxx))）。
-->
<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';

// 1. 收藏項目資料模型
interface SavedItem {
  id: string;
  name: string;
  price: string;
  priceValue: number;
  area: string;
  areaValue: number;
  rooms: string;
  district: string;
  bg: string;
  note: string;
}

// 2. 價格提醒資料模型
interface PriceAlert {
  name: string;
  target: number;
  current: string;
}

const router = useRouter();

// 3. 靜態 mock 收藏資料
const items = ref<SavedItem[]>([
  {
    id: '1',
    name: '佐敦 高級住宅',
    price: 'HK$36,000/月',
    priceValue: 36000,
    area: '200呎',
    areaValue: 200,
    rooms: '1房1廁',
    district: '九龍',
    bg: 'linear-gradient(160deg,#e8e8e8,#d0d0d0)',
    note: '',
  },
  {
    id: '2',
    name: '沙田第一城 3房',
    price: 'HK$18,500/月',
    priceValue: 18500,
    area: '650呎',
    areaValue: 650,
    rooms: '3房2廁',
    district: '新界',
    bg: 'linear-gradient(160deg,#e4dcd8,#ccc0bc)',
    note: '',
  },
  {
    id: '3',
    name: '中環 半山壹號',
    price: 'HK$45,000/月',
    priceValue: 45000,
    area: '800呎',
    areaValue: 800,
    rooms: '2房2廁',
    district: '港島',
    bg: 'linear-gradient(160deg,#dcdfe4,#bcc4cc)',
    note: '',
  },
  {
    id: '4',
    name: '將軍澳 海翩別墅',
    price: 'HK$22,800/月',
    priceValue: 22800,
    area: '550呎',
    areaValue: 550,
    rooms: '2房1廁',
    district: '新界',
    bg: 'linear-gradient(160deg,#e8e2d8,#ccc0b0)',
    note: '',
  },
  {
    id: '5',
    name: '銅鑼灣 維景花園',
    price: 'HK$28,500/月',
    priceValue: 28500,
    area: '480呎',
    areaValue: 480,
    rooms: '1房1廁',
    district: '港島',
    bg: 'linear-gradient(160deg,#e0e0e0,#c8c8c8)',
    note: '',
  },
  {
    id: '6',
    name: '元朗 YOHO Town',
    price: 'HK$15,200/月',
    priceValue: 15200,
    area: '420呎',
    areaValue: 420,
    rooms: '2房1廁',
    district: '新界',
    bg: 'linear-gradient(160deg,#e4e0d8,#ccc4b8)',
    note: '',
  },
]);

const sortKey = ref<'date' | 'price-asc' | 'price-desc' | 'area'>('date');
const alerts = ref<PriceAlert[]>([]);
const alertModalOpen = ref(false);
const alertTargetName = ref('');
const alertCurrentPrice = ref('');
const alertTargetInput = ref<number | null>(null);

const hasItems = computed(() => items.value.length > 0);

// 4. 排序收藏列表
const sortedItems = computed<SavedItem[]>(() => {
  const list = [...items.value];
  if (sortKey.value === 'price-asc') {
    list.sort((a, b) => a.priceValue - b.priceValue);
  } else if (sortKey.value === 'price-desc') {
    list.sort((a, b) => b.priceValue - a.priceValue);
  } else if (sortKey.value === 'area') {
    list.sort((a, b) => b.areaValue - a.areaValue);
  }
  return list;
});

// 5. 移除單筆收藏
const removeItem = (id: string): void => {
  items.value = items.value.filter((item) => item.id !== id);
};

// 6. 清除全部收藏
const clearAll = (): void => {
  items.value = [];
};

// 7. 更新備註
const updateNote = (item: SavedItem, value: string): void => {
  item.note = value;
};

// 8. 跳轉至樓盤詳情
const goDetail = (id: string): void => {
  void router.push(`/marketplace/listing/${id}`);
};

// 9. 跳轉至搜尋樓盤
const goBrowse = (): void => {
  void router.push('/marketplace/filter');
};

// 10. 開啟目標價提醒 Modal
const openAlertModal = (name: string, price: string): void => {
  alertTargetName.value = name;
  alertCurrentPrice.value = price;
  alertTargetInput.value = null;
  alertModalOpen.value = true;
};

// 11. 關閉目標價提醒 Modal
const closeAlertModal = (): void => {
  alertModalOpen.value = false;
};

// 12. 儲存目標價提醒
const saveAlert = (): void => {
  if (alertTargetInput.value === null || alertTargetInput.value < 1000) {
    return;
  }
  alerts.value = [
    ...alerts.value,
    {
      name: alertTargetName.value,
      target: alertTargetInput.value,
      current: alertCurrentPrice.value,
    },
  ];
  alertModalOpen.value = false;
};

// 13. 移除目標價提醒
const removeAlert = (index: number): void => {
  alerts.value = alerts.value.filter((_, i) => i !== index);
};
</script>

<template>
  <div class="saved-page">
    <!-- 1. 工具列：標題與篩選 -->
    <div class="saved-toolbar">
      <div>
        <div class="section-eyebrow">我的收藏</div>
        <div class="saved-toolbar-title">已收藏物件</div>
      </div>
      <div class="saved-filters">
        <select
          v-model="sortKey"
          class="saved-sort"
        >
          <option value="date">最近收藏</option>
          <option value="price-asc">租金低至高</option>
          <option value="price-desc">租金高至低</option>
          <option value="area">面積大至小</option>
        </select>
        <button
          type="button"
          class="saved-clear-btn"
          :disabled="!hasItems"
          @click="clearAll"
        >
          清除全部
        </button>
      </div>
    </div>

    <!-- 2. 收藏卡片網格 -->
    <div
      v-if="hasItems"
      class="saved-grid"
    >
      <div
        v-for="item in sortedItems"
        :key="item.id"
        class="saved-card"
      >
        <div
          class="saved-card-img"
          :style="{ background: item.bg }"
        >
          <button
            type="button"
            class="saved-card-remove"
            aria-label="移除收藏"
            @click="removeItem(item.id)"
          >
            <svg
              width="12"
              height="12"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2.5"
              stroke-linecap="round"
            >
              <path d="M18 6L6 18M6 6l12 12" />
            </svg>
          </button>
          <div class="saved-card-note">
            <input
              type="text"
              placeholder="加備注…"
              :value="item.note"
              @input="updateNote(item, ($event.target as HTMLInputElement).value)"
              @click.stop
            />
          </div>
        </div>
        <div class="saved-card-body">
          <div class="saved-card-name">{{ item.name }}</div>
          <div class="saved-card-price">{{ item.price }}</div>
          <div class="saved-card-meta">{{ item.district }} · {{ item.area }} · {{ item.rooms }}</div>
        </div>
        <div class="saved-card-footer">
          <button
            type="button"
            class="saved-card-btn"
            @click="openAlertModal(item.name, item.price)"
          >
            <svg
              width="11"
              height="11"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" />
              <path d="M13.73 21a2 2 0 0 1-3.46 0" />
            </svg>
            提醒
          </button>
          <button
            type="button"
            class="saved-card-btn primary"
            @click="goDetail(item.id)"
          >
            查看詳情
          </button>
        </div>
      </div>
    </div>

    <!-- 3. 空狀態 -->
    <div
      v-else
      class="saved-empty"
    >
      <div class="saved-empty-icon">
        <svg
          width="48"
          height="48"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="1.5"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
        </svg>
      </div>
      <div class="saved-empty-title">收藏清單係空嘅</div>
      <div class="saved-empty-desc">瀏覽樓盤時點擊「收藏」即可加入</div>
      <button
        type="button"
        class="saved-empty-action"
        @click="goBrowse"
      >
        去搜尋樓盤
      </button>
    </div>

    <!-- 4. 目標價提醒 Modal -->
    <div
      v-if="alertModalOpen"
      class="alert-modal-overlay"
      @click.self="closeAlertModal"
    >
      <div class="alert-modal">
        <div class="alert-modal-header">
          <div class="alert-modal-title">
            <svg
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" />
              <path d="M13.73 21a2 2 0 0 1-3.46 0" />
            </svg>
            設定目標價提醒
          </div>
          <button
            type="button"
            class="alert-close"
            aria-label="關閉"
            @click="closeAlertModal"
          >
            <svg
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
            >
              <path d="M18 6L6 18M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="alert-modal-body">
          <div class="alert-prop-name">{{ alertTargetName }}</div>
          <div class="alert-current">現時月租：{{ alertCurrentPrice }}</div>
          <label class="alert-label">目標月租 (HK$)</label>
          <div class="alert-input-row">
            <span class="alert-prefix">HK$</span>
            <input
              v-model.number="alertTargetInput"
              class="alert-input"
              type="number"
              placeholder="例如 30000"
              min="1000"
            />
          </div>
          <div class="alert-hint">當物件租金降至或低於目標價，系統會喺通知中心提醒你。</div>
          <button
            type="button"
            class="alert-confirm-btn"
            @click="saveAlert"
          >
            確認設定提醒
          </button>
          <div
            v-if="alerts.length > 0"
            class="alert-active-list"
          >
            <div class="alert-active-title">已設定嘅提醒</div>
            <div
              v-for="(alert, index) in alerts"
              :key="index"
              class="alert-active-item"
            >
              <span class="alert-active-name">{{ alert.name }}</span>
              <span class="alert-active-price">HK${{ alert.target.toLocaleString() }}</span>
              <button
                type="button"
                class="alert-active-rm"
                aria-label="移除提醒"
                @click="removeAlert(index)"
              >
                <svg
                  width="14"
                  height="14"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                >
                  <path d="M18 6L6 18M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 1. 頁面容器 */
.saved-page {
  padding: 24px 32px;
}

/* 2. 工具列 */
.saved-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  gap: 12px;
}

.section-eyebrow {
  font-size: 9px;
  letter-spacing: 2.5px;
  text-transform: uppercase;
  color: rgb(var(--color-primary));
  margin-bottom: 6px;
  font-weight: 500;
}

.saved-toolbar-title {
  font-family: var(--font-display);
  font-size: 24px;
  font-weight: 400;
  color: rgb(var(--color-text));
}

.saved-filters {
  display: flex;
  gap: 8px;
}

.saved-sort {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  padding: 5px 10px;
  font-size: 12px;
  font-family: var(--font-sans);
  outline: none;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  cursor: pointer;
}

.saved-clear-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  font-size: 12px;
  padding: 6px 14px;
  font-family: var(--font-sans);
  cursor: pointer;
  transition: background 0.15s ease;
}

.saved-clear-btn:hover:not(:disabled) {
  background: rgb(var(--color-surface-2));
}

.saved-clear-btn:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

/* 3. 收藏卡片網格 */
.saved-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.saved-card {
  background: rgb(var(--color-surface));
  border: 1px solid rgb(var(--color-border));
  border-radius: 12px;
  overflow: hidden;
  position: relative;
  cursor: pointer;
  transition: box-shadow 0.15s ease, border-color 0.15s ease;
}

.saved-card:hover {
  box-shadow: var(--shadow-raised);
  border-color: rgb(var(--color-brand-mid));
}

/* 4. 卡片圖片區 */
.saved-card-img {
  height: 110px;
  width: 100%;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.saved-card-remove {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.5);
  color: #fff;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s ease;
}

.saved-card-remove:hover {
  background: rgb(var(--color-danger));
}

.saved-card-note {
  position: absolute;
  bottom: 8px;
  left: 8px;
  right: 8px;
}

.saved-card-note input {
  width: 100%;
  background: rgba(255, 255, 255, 0.9);
  border: none;
  border-radius: 4px;
  padding: 4px 8px;
  font-size: 10px;
  font-family: var(--font-sans);
  outline: none;
  color: rgb(var(--color-text));
}

.saved-card-note input::placeholder {
  color: rgb(var(--color-ink-4));
}

/* 5. 卡片內容區 */
.saved-card-body {
  padding: 10px 12px;
}

.saved-card-name {
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 3px;
  color: rgb(var(--color-text));
}

.saved-card-price {
  font-size: 14px;
  font-weight: 500;
  color: rgb(var(--color-primary));
  margin-bottom: 4px;
}

.saved-card-meta {
  font-size: 10px;
  color: rgb(var(--color-ink-3));
}

/* 6. 卡片底部按鈕 */
.saved-card-footer {
  display: flex;
  gap: 6px;
  padding: 8px 12px;
  border-top: 1px solid rgb(var(--color-surface-3));
}

.saved-card-btn {
  flex: 1;
  font-size: 10px;
  padding: 5px 0;
  border-radius: 4px;
  border: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface));
  cursor: pointer;
  font-family: var(--font-sans);
  text-align: center;
  color: rgb(var(--color-ink-2));
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 3px;
  transition: background 0.15s ease;
}

.saved-card-btn:hover {
  background: rgb(var(--color-surface-2));
}

.saved-card-btn.primary {
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
  border-color: rgb(var(--color-primary));
}

.saved-card-btn.primary:hover {
  background: rgb(var(--color-brand-dark));
}

/* 7. 空狀態 */
.saved-empty {
  text-align: center;
  padding: 80px 40px;
  color: rgb(var(--color-ink-3));
}

.saved-empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
  opacity: 0.4;
  color: rgb(var(--color-ink-3));
  display: flex;
  justify-content: center;
}

.saved-empty-title {
  font-size: 15px;
  font-weight: 500;
  color: rgb(var(--color-text));
  margin-bottom: 6px;
}

.saved-empty-desc {
  font-size: 12px;
  line-height: 1.6;
  margin-bottom: 20px;
}

.saved-empty-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
  border: none;
  border-radius: 8px;
  font-size: 12px;
  padding: 8px 16px;
  font-family: var(--font-sans);
  font-weight: 500;
  cursor: pointer;
  transition: background 0.15s ease;
}

.saved-empty-action:hover {
  background: rgb(var(--color-brand-dark));
}

/* 8. 目標價提醒 Modal */
.alert-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  backdrop-filter: blur(4px);
  z-index: 200;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.alert-modal {
  background: rgb(var(--color-surface));
  border-radius: 16px;
  width: 100%;
  max-width: 420px;
  box-shadow: var(--shadow-floating);
  overflow: hidden;
}

.alert-modal-header {
  padding: 18px 20px;
  border-bottom: 1px solid rgb(var(--color-border));
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.alert-modal-title {
  font-size: 15px;
  font-weight: 500;
  color: rgb(var(--color-text));
  display: flex;
  align-items: center;
  gap: 6px;
}

.alert-close {
  background: none;
  border: none;
  cursor: pointer;
  color: rgb(var(--color-ink-3));
  line-height: 1;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.alert-close:hover {
  color: rgb(var(--color-text));
}

.alert-modal-body {
  padding: 20px;
}

.alert-prop-name {
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 4px;
  color: rgb(var(--color-text));
}

.alert-current {
  font-size: 12px;
  color: rgb(var(--color-ink-3));
  margin-bottom: 16px;
}

.alert-label {
  font-size: 10px;
  letter-spacing: 1px;
  text-transform: uppercase;
  color: rgb(var(--color-ink-3));
  margin-bottom: 6px;
  display: block;
}

.alert-input-row {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 14px;
}

.alert-prefix {
  font-size: 13px;
  color: rgb(var(--color-ink-2));
}

.alert-input {
  flex: 1;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  padding: 9px 12px;
  font-size: 14px;
  font-family: var(--font-sans);
  outline: none;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
}

.alert-input:focus {
  border-color: rgb(var(--color-primary));
}

.alert-hint {
  font-size: 11px;
  color: rgb(var(--color-ink-3));
  margin-bottom: 20px;
  line-height: 1.5;
}

.alert-confirm-btn {
  width: 100%;
  padding: 11px;
  font-size: 13px;
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
  border: none;
  border-radius: 8px;
  font-family: var(--font-sans);
  font-weight: 500;
  cursor: pointer;
  transition: background 0.15s ease;
}

.alert-confirm-btn:hover {
  background: rgb(var(--color-brand-dark));
}

.alert-active-list {
  margin-top: 16px;
  border-top: 1px solid rgb(var(--color-border));
  padding-top: 14px;
}

.alert-active-title {
  font-size: 10px;
  letter-spacing: 1px;
  text-transform: uppercase;
  color: rgb(var(--color-ink-3));
  margin-bottom: 8px;
}

.alert-active-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 0;
  border-bottom: 1px solid rgb(var(--color-surface-3));
  font-size: 12px;
}

.alert-active-name {
  flex: 1;
  color: rgb(var(--color-ink-2));
}

.alert-active-price {
  font-weight: 500;
  color: rgb(var(--color-primary));
}

.alert-active-rm {
  background: none;
  border: none;
  color: rgb(var(--color-ink-4));
  cursor: pointer;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.alert-active-rm:hover {
  color: rgb(var(--color-danger));
}

/* 9. 響應式設計 */
@media (max-width: 1024px) {
  .saved-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 767px) {
  .saved-page {
    padding: 16px;
  }

  .saved-toolbar {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .saved-filters {
    width: 100%;
  }

  .saved-sort {
    flex: 1;
  }

  .saved-grid {
    grid-template-columns: 1fr;
  }

  .saved-empty {
    padding: 48px 20px;
  }
}
</style>
