<!--
 * 通知中心頁。
 * 1. 高保真還原 HTML 設計稿雙欄布局（work-shell / work-sidebar / work-main）。
 * 2. 側欄提供全部通知、優惠提醒、物業消息、支付提醒、系統通知分類切換。
 * 3. 主內容區按分類渲染對應 panel，包含 hero 標題區與 notice-list 通知卡片。
 * 4. 使用靜態 mock 資料，不接入真實 API。
 * 5. 點擊通知項可標記已讀並依分類跳轉對應模組路由。
-->
<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';

type NotificationFilter = 'all' | 'offers' | 'property' | 'payment' | 'system';
type NotificationGroup = '今日' | '昨日' | '更早';
type ChipTone = 'brand' | 'warn' | 'default';

interface NotificationItem {
  id: string;
  icon: string;
  title: string;
  description: string;
  time: string;
  unread: boolean;
  group: NotificationGroup;
}

interface CategoryPanel {
  key: NotificationFilter;
  label: string;
  kicker: string;
  title: string;
  desc: string;
  chipTone: ChipTone;
  items: NotificationItem[];
}

const router = useRouter();
const activeFilter = ref<NotificationFilter>('all');

// 1. 分類面板靜態資料
const panels: CategoryPanel[] = [
  {
    key: 'all',
    label: '全部通知',
    kicker: 'Notifications',
    title: '所有通知',
    desc: '未讀通知會優先顯示，方便業主、租客與職員快速處理待辦。',
    chipTone: 'brand',
    items: [
      {
        id: 'offer-toothbrush',
        icon: '惠',
        title: '收藏商品有新優惠',
        description: '高露潔牙刷於 AEON 減價 60%，可前往綜合優惠查看最新價格。',
        time: '5 分鐘前',
        unread: true,
        group: '今日',
      },
      {
        id: 'payment-june',
        icon: '費',
        title: '6月管理費待繳提醒',
        description: '截止日期：2026年6月15日，金額 HK$2,800。',
        time: '1 小時前',
        unread: true,
        group: '今日',
      },
      {
        id: 'property-price-drop',
        icon: '樓',
        title: '查看過的物件降價',
        description: '佐敦高級住宅月租由 HK$38,000 降至 HK$36,000。',
        time: '3 小時前',
        unread: true,
        group: '今日',
      },
      {
        id: 'viewing-confirmed',
        icon: '約',
        title: '預約睇樓確認',
        description: '中環甲級寫字樓睇樓時間：6月8日下午2時。',
        time: '昨天 14:30',
        unread: false,
        group: '昨日',
      },
      {
        id: 'system-lift',
        icon: '告',
        title: '社區公告：電梯保養',
        description: 'A棟電梯將於 6月8日 上午 9-12 時進行年度保養。',
        time: '昨天 09:15',
        unread: false,
        group: '昨日',
      },
    ],
  },
  {
    key: 'offers',
    label: '優惠提醒',
    kicker: 'Offers',
    title: '優惠提醒',
    desc: '查看收藏商品、優惠價格與到貨提醒。',
    chipTone: 'default',
    items: [
      {
        id: 'offer-toothbrush-2',
        icon: '惠',
        title: '收藏商品有新優惠',
        description: '高露潔牙刷於 AEON 減價 60%，可前往綜合優惠查看最新價格。',
        time: '5 分鐘前',
        unread: true,
        group: '今日',
      },
    ],
  },
  {
    key: 'property',
    label: '物業消息',
    kicker: 'Property',
    title: '物業消息',
    desc: '查看睇樓、樓盤變動、社區公告與大廈消息。',
    chipTone: 'default',
    items: [
      {
        id: 'property-price-drop-2',
        icon: '樓',
        title: '查看過的物件降價',
        description: '佐敦高級住宅月租由 HK$38,000 降至 HK$36,000。',
        time: '3 小時前',
        unread: true,
        group: '今日',
      },
      {
        id: 'viewing-confirmed-2',
        icon: '約',
        title: '預約睇樓確認',
        description: '中環甲級寫字樓睇樓時間：6月8日下午2時。',
        time: '昨天 14:30',
        unread: false,
        group: '昨日',
      },
      {
        id: 'system-lift-2',
        icon: '告',
        title: '社區公告：電梯保養',
        description: 'A棟電梯將於 6月8日 上午 9-12 時進行年度保養。',
        time: '昨天 09:15',
        unread: false,
        group: '昨日',
      },
    ],
  },
  {
    key: 'payment',
    label: '支付提醒',
    kicker: 'Payment',
    title: '支付提醒',
    desc: '查看管理費、賬單與支付狀態提醒。',
    chipTone: 'warn',
    items: [
      {
        id: 'payment-june-2',
        icon: '費',
        title: '6月管理費待繳提醒',
        description: '截止日期：2026年6月15日，金額 HK$2,800。',
        time: '1 小時前',
        unread: true,
        group: '今日',
      },
    ],
  },
  {
    key: 'system',
    label: '系統通知',
    kicker: 'System',
    title: '系統通知',
    desc: '查看平台更新、帳戶安全與系統消息。',
    chipTone: 'default',
    items: [
      {
        id: 'system-account-updated',
        icon: '系',
        title: '帳戶資料已更新',
        description: '你的聯絡資料已成功保存。',
        time: '6月5日',
        unread: false,
        group: '更早',
      },
    ],
  },
];

const groupOrder: NotificationGroup[] = ['今日', '昨日', '更早'];

// 2. 當前面板
const activePanel = computed<CategoryPanel>(
  () => panels.find((item) => item.key === activeFilter.value) ?? panels[0],
);

// 3. 取得指定分類未讀數
const unreadCountFor = (key: NotificationFilter): number => {
  const panel = panels.find((item) => item.key === key);
  return panel ? panel.items.filter((item) => item.unread).length : 0;
};

// 4. 取得當前面板分組項目
const groupItems = (group: NotificationGroup): NotificationItem[] =>
  activePanel.value.items.filter((item) => item.group === group);

// 5. 判斷當前面板是否需要分組展示
const useGroupedDisplay = computed(() => activePanel.value.key === 'all');

// 6. 標記當前面板全部通知已讀
const markAllRead = (): void => {
  activePanel.value.items.forEach((item) => {
    item.unread = false;
  });
};

// 7. 標記單條通知已讀並依分類跳轉
const handleItemClick = (item: NotificationItem): void => {
  if (item.unread) {
    item.unread = false;
  }
  const routeMap: Partial<Record<NotificationFilter, string>> = {
    offers: '/offers',
    property: '/property',
    payment: '/payments',
  };
  const target = routeMap[activeFilter.value];
  if (target) {
    void router.push(target);
  }
};
</script>

<template>
  <main class="notif-page">
    <div class="work-shell">
      <aside class="work-sidebar">
        <h1>通知中心</h1>
        <p>集中查看優惠提醒、物業消息、支付提醒與系統通知。</p>
        <nav class="work-nav">
          <button
            v-for="panel in panels"
            :key="panel.key"
            type="button"
            class="work-nav-item"
            :class="activeFilter === panel.key ? 'work-nav-item--on' : ''"
            @click="activeFilter = panel.key"
          >
            <span>{{ panel.label }}</span>
            <span
              v-if="unreadCountFor(panel.key) > 0"
              class="work-chip"
              :class="`work-chip--${panel.chipTone}`"
            >
              {{ unreadCountFor(panel.key) }}
            </span>
          </button>
        </nav>
      </aside>

      <section class="work-main">
        <div class="work-panel work-panel--on">
          <header class="work-hero">
            <div>
              <div class="work-kicker">{{ activePanel.kicker }}</div>
              <h2 class="work-title">{{ activePanel.title }}</h2>
              <p class="work-desc">{{ activePanel.desc }}</p>
            </div>
            <button
              type="button"
              class="work-action work-action--secondary"
              :disabled="unreadCountFor(activeFilter) === 0"
              @click="markAllRead"
            >
              全部標為已讀
            </button>
          </header>

          <template v-if="useGroupedDisplay">
            <section
              v-for="group in groupOrder"
              :key="group"
              class="work-card"
            >
              <div
                v-if="groupItems(group).length > 0"
                class="work-card-title"
              >
                {{ group }}
              </div>
              <div
                v-if="groupItems(group).length > 0"
                class="notice-list"
              >
                <article
                  v-for="item in groupItems(group)"
                  :key="item.id"
                  class="notice-item"
                  :class="item.unread ? 'notice-item--unread' : ''"
                  @click="handleItemClick(item)"
                >
                  <div
                    class="notice-dot"
                    :class="item.unread ? '' : 'notice-dot--read'"
                  />
                  <div class="notice-icon">{{ item.icon }}</div>
                  <div class="notice-content">
                    <div class="notice-title">{{ item.title }}</div>
                    <div class="notice-desc">{{ item.description }}</div>
                  </div>
                  <div class="notice-time">{{ item.time }}</div>
                </article>
              </div>
            </section>
          </template>

          <section
            v-else
            class="work-card"
          >
            <div class="work-card-title">{{ activePanel.label }}</div>
            <div class="notice-list">
              <article
                v-for="item in activePanel.items"
                :key="item.id"
                class="notice-item"
                :class="item.unread ? 'notice-item--unread' : ''"
                @click="handleItemClick(item)"
              >
                <div
                  class="notice-dot"
                  :class="item.unread ? '' : 'notice-dot--read'"
                />
                <div class="notice-icon">{{ item.icon }}</div>
                <div class="notice-content">
                  <div class="notice-title">{{ item.title }}</div>
                  <div class="notice-desc">{{ item.description }}</div>
                </div>
                <div class="notice-time">{{ item.time }}</div>
              </article>
            </div>
          </section>

          <section
            v-if="activePanel.items.length === 0"
            class="work-card work-card--empty"
          >
            <strong>目前沒有通知</strong>
            <span>新的優惠提醒、支付提醒與系統通知會顯示在這裡。</span>
          </section>
        </div>
      </section>
    </div>
  </main>
</template>

<style scoped>
/* 1. 頁面外框 */
.notif-page {
  min-height: calc(100vh - var(--nav-h, 52px));
  background: rgb(var(--color-surface-2));
  color: rgb(var(--color-text));
}

/* 2. 雙欄主版面 */
.work-shell {
  display: grid;
  grid-template-columns: 240px minmax(0, 1fr);
  gap: 16px;
  max-width: 1180px;
  margin: 0 auto;
  padding: 12px var(--layout-page-padding-inline) 16px;
}

/* 3. 側欄 */
.work-sidebar {
  position: sticky;
  top: calc(var(--nav-h, 52px) + 12px);
  align-self: start;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 16px;
}

.work-sidebar h1 {
  margin: 0;
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 26px;
  font-weight: 400;
  line-height: 1.2;
}

.work-sidebar p {
  margin: 8px 0 0;
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  line-height: 1.7;
}

/* 4. 側欄導航 */
.work-nav {
  display: grid;
  gap: 2px;
  margin-top: 18px;
}

.work-nav-item {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 34px;
  border: 0;
  border-radius: 0;
  background: transparent;
  color: rgb(var(--color-ink-2));
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  font-weight: 500;
  padding: 8px 2px;
  text-align: left;
  transition: background-color 0.15s ease, color 0.15s ease;
}

.work-nav-item::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 2px;
  background: rgb(var(--color-primary));
  transform: scaleX(0);
  transform-origin: left center;
  transition: transform 0.24s ease;
}

.work-nav-item:hover {
  background: rgb(var(--color-primary-soft));
  color: rgb(var(--color-text));
}

.work-nav-item--on {
  background: transparent;
  color: rgb(var(--color-primary));
  font-weight: 700;
  outline: none;
}

.work-nav-item--on:hover {
  background: rgb(var(--color-primary-soft));
}

.work-nav-item--on::after {
  transform: scaleX(1);
}

/* 5. 標籤晶片 */
.work-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  background: rgb(var(--color-surface-2));
  color: rgb(var(--color-ink-2));
  font-size: 11px;
  font-weight: 600;
  padding: 5px 9px;
  white-space: nowrap;
}

.work-chip--brand {
  background: rgb(var(--color-primary-soft));
  color: rgb(var(--color-primary));
}

.work-chip--warn {
  background: rgb(var(--color-warning-bg));
  color: rgb(var(--color-warning));
}

/* 6. 主內容區 */
.work-main {
  display: grid;
  gap: 12px;
  min-width: 0;
  align-content: start;
}

.work-panel {
  display: grid;
  gap: 14px;
  align-content: start;
}

.work-panel--on {
  display: grid;
}

/* 7. 標題區 */
.work-hero {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 18px;
  border: 0;
  border-bottom: 1px solid rgb(var(--color-border));
  border-radius: 0;
  background: transparent;
  margin: 0;
  padding: 0 0 10px;
}

.work-kicker {
  margin-bottom: 4px;
  color: rgb(var(--color-ink-3));
  font-size: 9px;
  letter-spacing: 1.4px;
  text-transform: uppercase;
}

.work-title {
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 20px;
  font-weight: 600;
  line-height: 1.25;
}

.work-desc {
  max-width: 560px;
  margin: 5px 0 0;
  color: rgb(var(--color-ink-3));
  font-size: 13px;
  line-height: 1.5;
}

/* 8. 操作按鈕 */
.work-action {
  border: 0;
  border-radius: 6px;
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  padding: 10px 14px;
  white-space: nowrap;
}

.work-action--secondary {
  border: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
}

.work-action--secondary:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

/* 9. 卡片區 */
.work-card {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 16px;
}

.work-card-title {
  margin-bottom: 10px;
  color: rgb(var(--color-text));
  font-size: 14px;
  font-weight: 600;
}

.work-card--empty {
  display: grid;
  justify-items: center;
  gap: 6px;
  padding: 60px 24px;
  text-align: center;
}

.work-card--empty strong {
  color: rgb(var(--color-text));
  font-size: 15px;
  font-weight: 600;
}

.work-card--empty span {
  color: rgb(var(--color-ink-3));
  font-size: 12px;
}

/* 10. 通知列表 */
.notice-list {
  display: grid;
  gap: 8px;
}

.notice-item {
  display: grid;
  grid-template-columns: 10px 42px minmax(0, 1fr) auto;
  align-items: start;
  gap: 12px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 13px 14px;
  cursor: pointer;
  transition: border-color 0.15s ease;
}

.notice-item:hover {
  border-color: rgb(var(--color-brand-mid));
}

.notice-item--unread {
  border-left: 3px solid rgb(var(--color-primary));
  background: #fffaf7;
}

.notice-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: rgb(var(--color-primary));
  margin-top: 7px;
}

.notice-dot--read {
  background: rgb(var(--color-border));
}

.notice-icon {
  display: flex;
  width: 36px;
  height: 36px;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: rgb(var(--color-primary-soft));
  color: rgb(var(--color-primary));
  font-size: 13px;
  font-weight: 700;
}

.notice-content {
  min-width: 0;
}

.notice-title {
  color: rgb(var(--color-text));
  font-size: 13px;
  font-weight: 600;
}

.notice-desc {
  margin-top: 4px;
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  line-height: 1.6;
}

.notice-time {
  color: rgb(var(--color-ink-4));
  font-size: 11px;
  white-space: nowrap;
}

/* 11. 響應式 - 平板 */
@media (max-width: 900px) {
  .work-shell {
    grid-template-columns: 1fr;
    padding: 14px;
  }

  .work-sidebar {
    position: static;
  }

  .work-nav {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .notice-item {
    grid-template-columns: 8px 34px minmax(0, 1fr);
  }

  .notice-time {
    grid-column: 3;
  }
}

/* 12. 響應式 - 手機 */
@media (max-width: 560px) {
  .work-hero {
    align-items: flex-start;
    flex-direction: column;
  }

  .work-nav {
    grid-template-columns: 1fr;
  }
}
</style>
