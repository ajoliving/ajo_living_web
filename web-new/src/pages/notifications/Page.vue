<!--
 * 通知中心頁。
 * 1. 高保真還原桌面參考通知中心雙欄布局。
 * 2. 提供全部通知、優惠提醒與本地標記已讀操作。
 * 3. 先搭建 UI，不接入真實通知資料流。
-->
<script setup lang="ts">
import { computed, ref } from 'vue';

type NotificationFilter = 'all' | 'offers' | 'system' | 'property' | 'payment';
type NotificationTone = 'orange' | 'blue' | 'green' | 'gray';

interface NotificationViewItem {
  id: string;
  filter: NotificationFilter;
  group: '今日' | '昨日' | '更早';
  icon: string;
  tone: NotificationTone;
  title: string;
  description: string;
  time: string;
  unread: boolean;
}

const activeFilter = ref<NotificationFilter>('all');
const notifications = ref<NotificationViewItem[]>([
  {
    id: 'offer-toothbrush',
    filter: 'offers',
    group: '今日',
    icon: '火',
    tone: 'orange',
    title: '你收藏的商品有新優惠',
    description: '高露潔牙刷於 AEON 減價 60%，可前往綜合優惠查看最新價格。',
    time: '5 分鐘前',
    unread: true,
  },
  {
    id: 'payment-june',
    filter: 'payment',
    group: '今日',
    icon: '卡',
    tone: 'blue',
    title: '6月管理費待繳提醒',
    description: '截止日期：2026年6月15日，金額 HK$2,800。',
    time: '1 小時前',
    unread: true,
  },
  {
    id: 'property-price-drop',
    filter: 'property',
    group: '今日',
    icon: '宅',
    tone: 'green',
    title: '你查看過的物件降價了',
    description: '佐敦高級住宅月租由 HK$38,000 降至 HK$36,000。',
    time: '3 小時前',
    unread: true,
  },
  {
    id: 'viewing-confirmed',
    filter: 'property',
    group: '昨日',
    icon: '單',
    tone: 'gray',
    title: '預約睇樓確認',
    description: '中環甲級寫字樓睇樓時間：6月8日下午2時。',
    time: '昨天 14:30',
    unread: false,
  },
  {
    id: 'offer-fitness',
    filter: 'offers',
    group: '昨日',
    icon: '禮',
    tone: 'orange',
    title: '新優惠上架：Pure Fitness 住戶月費',
    description: '全港分店月費 HK$799，限時優惠。',
    time: '昨天 09:15',
    unread: false,
  },
  {
    id: 'payment-received',
    filter: 'payment',
    group: '昨日',
    icon: '收',
    tone: 'green',
    title: '5月管理費已收妥',
    description: 'HK$2,800 已成功入賬，謝謝。',
    time: '昨天 08:02',
    unread: false,
  },
  {
    id: 'system-lift',
    filter: 'system',
    group: '更早',
    icon: '告',
    tone: 'blue',
    title: '社區公告：電梯保養',
    description: 'A棟電梯將於 6月8日 上午 9-12 時進行年度保養。',
    time: '6月5日',
    unread: false,
  },
]);

const filterItems: Array<{ key: NotificationFilter; label: string }> = [
  { key: 'all', label: '全部' },
  { key: 'offers', label: '優惠提醒' },
  { key: 'system', label: '系統通知' },
  { key: 'property', label: '物件更新' },
  { key: 'payment', label: '支付提醒' },
];
const groupOrder: NotificationViewItem['group'][] = ['今日', '昨日', '更早'];
const filteredNotifications = computed(() =>
  activeFilter.value === 'all'
    ? notifications.value
    : notifications.value.filter((item) => item.filter === activeFilter.value),
);
const unreadCount = computed(() => notifications.value.filter((item) => item.unread).length);
const activeTitle = computed(() => filterItems.find((item) => item.key === activeFilter.value)?.label ?? '全部');

// 1. 取得指定分類未讀數
const unreadCountFor = (filter: NotificationFilter): number =>
  notifications.value.filter((item) => item.unread && (filter === 'all' || item.filter === filter)).length;

// 2. 取得分組通知
const groupItems = (group: NotificationViewItem['group']): NotificationViewItem[] =>
  filteredNotifications.value.filter((item) => item.group === group);

// 3. 標記全部通知已讀
const markAllRead = (): void => {
  notifications.value = notifications.value.map((item) => ({ ...item, unread: false }));
};

// 4. 標記單條通知已讀
const markRead = (id: string): void => {
  notifications.value = notifications.value.map((item) =>
    item.id === id ? { ...item, unread: false } : item,
  );
};
</script>

<template>
  <main class="notif-wrap">
    <aside class="notif-sidebar">
      <div class="notif-sidebar-title">
        通知中心
      </div>
      <div class="notif-filter">
        <button
          v-for="item in filterItems"
          :key="item.key"
          type="button"
          class="nf-item"
          :class="activeFilter === item.key ? 'nf-item--active' : ''"
          @click="activeFilter = item.key"
        >
          <span>{{ item.label }}</span>
          <span
            v-if="unreadCountFor(item.key) > 0"
            class="nf-badge"
          >
            {{ unreadCountFor(item.key) }}
          </span>
        </button>
      </div>
    </aside>

    <section class="notif-main">
      <header class="notif-main-head">
        <div>
          <p>Notifications</p>
          <h1>{{ activeFilter === 'all' ? '所有通知' : activeTitle }}</h1>
        </div>
        <button
          type="button"
          :disabled="unreadCount === 0"
          @click="markAllRead"
        >
          全部標為已讀
        </button>
      </header>

      <template
        v-for="group in groupOrder"
        :key="group"
      >
        <section
          v-if="groupItems(group).length > 0"
          class="notif-group"
        >
          <div class="notif-group-title">
            {{ group }}
          </div>
          <div class="notif-list">
            <article
              v-for="item in groupItems(group)"
              :key="item.id"
              class="notif-item"
              :class="item.unread ? 'notif-item--unread' : ''"
              @click="markRead(item.id)"
            >
              <div
                class="notif-dot"
                :class="item.unread ? '' : 'notif-dot--read'"
              />
              <div
                class="notif-icon"
                :class="`notif-icon--${item.tone}`"
              >
                {{ item.icon }}
              </div>
              <div class="notif-content">
                <div class="notif-title">
                  {{ item.title }}
                </div>
                <div class="notif-desc">
                  {{ item.description }}
                </div>
                <div class="notif-time">
                  {{ item.time }}
                </div>
              </div>
              <button
                v-if="item.unread"
                type="button"
                class="notif-read-button"
                @click.stop="markRead(item.id)"
              >
                標記已讀
              </button>
            </article>
          </div>
        </section>
      </template>

      <section
        v-if="filteredNotifications.length === 0"
        class="notif-empty"
      >
        <div>通知</div>
        <strong>目前沒有通知</strong>
        <span>新的優惠提醒、支付提醒與系統通知會顯示在這裡。</span>
      </section>
    </section>
  </main>
</template>

<style scoped>
.notif-wrap {
  display: grid;
  min-height: calc(100vh - var(--app-header-offset, 48px));
  background: #ffffff;
  color: #1a1a1a;
  grid-template-columns: 240px minmax(0, 1fr);
}

.notif-sidebar {
  border-right: 1px solid #e4e4e4;
  background: #ffffff;
  padding: 20px 16px;
}

.notif-sidebar-title {
  color: #1a1a1a;
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 16px;
}

.notif-filter {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.nf-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border: 0;
  border-radius: 2px;
  background: transparent;
  color: #777777;
  cursor: pointer;
  font: inherit;
  font-size: 12px;
  padding: 9px 12px;
  text-align: left;
}

.nf-item:hover {
  background: #f4f4f4;
  color: #1a1a1a;
}

.nf-item--active {
  background: #fff0e6;
  color: #f05a00;
  font-weight: 500;
}

.nf-badge {
  border-radius: 10px;
  background: #f05a00;
  color: #ffffff;
  font-size: 10px;
  padding: 1px 6px;
}

.notif-main {
  background: #ffffff;
  padding: 20px 24px 40px;
}

.notif-main-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.notif-main-head p {
  margin: 0 0 5px;
  color: #777777;
  font-size: 10px;
  letter-spacing: 1.5px;
  text-transform: uppercase;
}

.notif-main-head h1 {
  margin: 0;
  color: #1a1a1a;
  font-size: 14px;
  font-weight: 500;
}

.notif-main-head button {
  border: 0;
  background: transparent;
  color: #f05a00;
  cursor: pointer;
  font: inherit;
  font-size: 11px;
}

.notif-main-head button:disabled {
  color: #aaaaaa;
  cursor: not-allowed;
}

.notif-group {
  margin-top: 20px;
}

.notif-group:first-of-type {
  margin-top: 0;
}

.notif-group-title {
  color: #777777;
  font-size: 10px;
  letter-spacing: 1.5px;
  margin-bottom: 8px;
  text-transform: uppercase;
}

.notif-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.notif-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  border: 1px solid #e4e4e4;
  border-radius: 3px;
  background: #ffffff;
  cursor: pointer;
  padding: 12px 14px;
  transition: border-color 0.15s ease;
}

.notif-item:hover {
  border-color: #f05a00;
}

.notif-item--unread {
  border-left: 3px solid #f05a00;
  background: #fffaf7;
}

.notif-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #f05a00;
  flex: 0 0 auto;
  margin-top: 5px;
}

.notif-dot--read {
  background: #e4e4e4;
}

.notif-icon {
  display: flex;
  width: 36px;
  height: 36px;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  flex: 0 0 auto;
  font-size: 13px;
  font-weight: 600;
}

.notif-icon--orange {
  background: #fff0e6;
  color: #c04600;
}

.notif-icon--blue {
  background: #e6f0ff;
  color: #1d4ed8;
}

.notif-icon--green {
  background: #e6f7ee;
  color: #047857;
}

.notif-icon--gray {
  background: #f4f4f4;
  color: #333333;
}

.notif-content {
  min-width: 0;
  flex: 1;
}

.notif-title {
  color: #1a1a1a;
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 3px;
}

.notif-desc {
  color: #777777;
  font-size: 11px;
  line-height: 1.5;
}

.notif-time {
  color: #aaaaaa;
  font-size: 10px;
  margin-top: 4px;
}

.notif-read-button {
  border: 1px solid #e4e4e4;
  border-radius: 2px;
  background: #ffffff;
  color: #f05a00;
  flex: 0 0 auto;
  font-size: 11px;
  padding: 5px 8px;
}

.notif-empty {
  display: grid;
  justify-items: center;
  border: 1px dashed #e4e4e4;
  border-radius: 3px;
  color: #777777;
  gap: 6px;
  margin-top: 20px;
  padding: 60px 24px;
  text-align: center;
}

.notif-empty div {
  color: #aaaaaa;
  font-size: 42px;
}

.notif-empty strong {
  color: #1a1a1a;
  font-size: 15px;
  font-weight: 500;
}

.notif-empty span {
  font-size: 12px;
}

@media (max-width: 760px) {
  .notif-wrap {
    grid-template-columns: 1fr;
  }

  .notif-sidebar {
    border-right: 0;
    border-bottom: 1px solid #e4e4e4;
  }

  .notif-filter {
    flex-direction: row;
    overflow-x: auto;
  }

  .nf-item {
    flex: 0 0 auto;
    gap: 8px;
  }

  .notif-main-head,
  .notif-item {
    align-items: flex-start;
    flex-direction: column;
  }

  .notif-read-button {
    align-self: flex-start;
  }
}
</style>
