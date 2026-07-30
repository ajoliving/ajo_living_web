<!--
 * 通知中心頁。
 * 1. 嚴格對齊 HTML 設計稿 #page-notif 區段，雙欄 work-shell / work-sidebar / work-main 結構。
 * 2. 側欄提供全部通知、優惠提醒、物業消息、支付提醒、系統通知分類切換。
 * 3. 接入真實通知 API，保留設計稿 work-panel、hero 與 notice-list 呈現。
 * 4. 支援單條與當前分類標記已讀。
 * 5. 樣式使用 HTML 原生 CSS 變量（var(--brand) / var(--ink) / var(--sur) 等）。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import {
  fetchNotifications,
  markAllNotificationsRead,
  markNotificationRead,
} from '@/httpapis/notifications';
import type { NotificationItem } from '@/model/notification';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';

type PanelKey = 'notif-all' | 'notif-offers' | 'notif-property' | 'notif-payment' | 'notif-system';
type ChipTone = 'brand' | 'warn';
type NoticeGroupKey = 'today' | 'yesterday' | 'earlier';

interface NoticeItem {
  id: string;
  panelKey: PanelKey;
  title: string;
  description: string;
  createdAt: string;
  group: NoticeGroupKey;
  unread: boolean;
  targetPath: string;
}

interface NoticeGroup {
  title: NoticeGroupKey;
  items: NoticeItem[];
}

interface PanelConfig {
  key: PanelKey;
  navLabel: string;
  chipTone?: ChipTone;
  kicker: string;
  title: string;
  desc: string;
  showMarkAllRead: boolean;
}

interface Panel extends PanelConfig {
  chipCount?: number;
  groups: NoticeGroup[];
}

const router = useRouter();
const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const preferenceStore = usePreferenceStore();
const activeFilter = ref<PanelKey>('notif-all');
const loading = ref(false);
const notifications = ref<NoticeItem[]>([]);

// 1. 五個分類面板設定
const panelConfigs = computed<PanelConfig[]>(() => [
  {
    key: 'notif-all',
    navLabel: t('account.notifications.allTitle'),
    chipTone: 'brand',
    kicker: t('account.notifications.title'),
    title: t('account.notifications.allTitle'),
    desc: t('account.notifications.filters.allDescription'),
    showMarkAllRead: true,
  },
  {
    key: 'notif-offers',
    navLabel: t('account.notifications.offers'),
    kicker: t('account.notifications.offers'),
    title: t('account.notifications.offers'),
    desc: t('account.notifications.filters.offersDescription'),
    showMarkAllRead: true,
  },
  {
    key: 'notif-property',
    navLabel: t('account.notifications.property'),
    kicker: t('account.notifications.property'),
    title: t('account.notifications.property'),
    desc: t('account.notifications.filters.propertyDescription'),
    showMarkAllRead: true,
  },
  {
    key: 'notif-payment',
    navLabel: t('account.notifications.payment'),
    chipTone: 'warn',
    kicker: t('account.notifications.payment'),
    title: t('account.notifications.payment'),
    desc: t('account.notifications.filters.paymentDescription'),
    showMarkAllRead: true,
  },
  {
    key: 'notif-system',
    navLabel: t('account.notifications.system'),
    kicker: t('account.notifications.system'),
    title: t('account.notifications.system'),
    desc: t('account.notifications.filters.systemDescription'),
    showMarkAllRead: false,
  },
]);

const groupOrder: NoticeGroupKey[] = ['today', 'yesterday', 'earlier'];
const relativeTimeFormatter = computed(() =>
  new Intl.RelativeTimeFormat(preferenceStore.locale, { numeric: 'auto' }),
);

// 2. 根據通知內容判斷分類
const resolvePanelKey = (item: NotificationItem): PanelKey => {
  const value = `${item.category} ${item.related_type} ${item.title} ${item.body}`.toLowerCase();

  if (value.includes('payment') || value.includes('bill') || value.includes('fee') || value.includes('管理費')) {
    return 'notif-payment';
  }
  if (value.includes('offer') || value.includes('coupon') || value.includes('discount') || value.includes('優惠')) {
    return 'notif-offers';
  }
  if (value.includes('property') || value.includes('listing') || value.includes('viewing') || value.includes('樓')) {
    return 'notif-property';
  }

  return 'notif-system';
};

// 3. 取得通知圖示
const resolveIcon = (panelKey: PanelKey): string => {
  const iconKeyMap: Record<PanelKey, string> = {
    'notif-all': 'account.notifications.icons.all',
    'notif-offers': 'account.notifications.icons.offers',
    'notif-property': 'account.notifications.icons.property',
    'notif-payment': 'account.notifications.icons.payment',
    'notif-system': 'account.notifications.icons.system',
  };

  return t(iconKeyMap[panelKey]);
};

// 4. 取得通知跳轉目標
const resolveTargetPath = (item: NotificationItem, panelKey: PanelKey): string => {
  if (item.category !== 'system_notice' && item.related_type === 'chat' && item.related_public_id) {
    return `/account/chat/${item.related_public_id}`;
  }

  const routeMap: Partial<Record<PanelKey, string>> = {
    'notif-offers': '/supermarket-offers',
    'notif-property': '/properties',
    'notif-payment': '/payments',
  };

  return routeMap[panelKey] ?? '';
};

// 5. 取得通知分組
const resolveGroup = (value: string): NoticeGroupKey => {
  const createdAt = new Date(value);
  const now = new Date();
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const yesterday = new Date(today);
  yesterday.setDate(today.getDate() - 1);

  if (createdAt >= today) {
    return 'today';
  }
  if (createdAt >= yesterday) {
    return 'yesterday';
  }

  return 'earlier';
};

// 6. 格式化通知時間
const formatNoticeTime = (value: string): string => {
  const createdAt = new Date(value);
  if (Number.isNaN(createdAt.getTime())) {
    return value;
  }

  const diffMs = Date.now() - createdAt.getTime();
  const minuteMs = 60 * 1000;
  const hourMs = 60 * minuteMs;
  if (diffMs >= 0 && diffMs < hourMs) {
    return relativeTimeFormatter.value.format(-Math.max(1, Math.floor(diffMs / minuteMs)), 'minute');
  }
  if (diffMs >= 0 && diffMs < 24 * hourMs) {
    return relativeTimeFormatter.value.format(-Math.floor(diffMs / hourMs), 'hour');
  }

  return createdAt.toLocaleString(preferenceStore.locale, {
    month: 'numeric',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
};

// 7. 轉換通知資料
const mapNotification = (item: NotificationItem): NoticeItem => {
  const panelKey = resolvePanelKey(item);

  return {
    id: item.notification_id,
    panelKey,
    title: item.title,
    description: item.body,
    createdAt: item.created_at,
    group: resolveGroup(item.created_at),
    unread: !item.is_read,
    targetPath: resolveTargetPath(item, panelKey),
  };
};

// 8. 按分類生成設計稿面板資料
const panels = computed<Panel[]>(() =>
  panelConfigs.value.map((config) => {
    const items = config.key === 'notif-all'
      ? notifications.value
      : notifications.value.filter((item) => item.panelKey === config.key);
    const groups = groupOrder
      .map((group) => ({
        title: group,
        items: items.filter((item) => item.group === group),
      }))
      .filter((group) => group.items.length > 0);
    const chipCount = items.filter((item) => item.unread).length;

    return {
      ...config,
      chipCount: chipCount > 0 ? chipCount : undefined,
      groups,
    };
  }),
);

// 9. 讀取通知列表
const loadNotifications = async (): Promise<void> => {
  loading.value = true;

  try {
    const response = await fetchNotifications({ page: 1, page_size: 50 });
    notifications.value = response.data.data.items.map(mapNotification);
  } catch (error: unknown) {
    feedbackStore.pushToast(
      axios.isAxiosError<{ message?: string }>(error)
        ? error.response?.data?.message ?? t('account.notifications.loadError')
        : t('account.notifications.loadError'),
      'error',
    );
    notifications.value = [];
  } finally {
    loading.value = false;
  }
};

// 10. 更新本地已讀狀態
const markLocalRead = (ids: string[]): void => {
  const targetIDs = new Set(ids);
  notifications.value = notifications.value.map((item) =>
    targetIDs.has(item.id) ? { ...item, unread: false } : item,
  );
};

// 11. 標記當前面板通知已讀
const markAllRead = async (panel: Panel): Promise<void> => {
  const unreadIDs = panel.groups.flatMap((group) =>
    group.items.filter((item) => item.unread).map((item) => item.id),
  );
  if (unreadIDs.length === 0) {
    return;
  }

  try {
    if (panel.key === 'notif-all') {
      await markAllNotificationsRead();
      markLocalRead(notifications.value.map((item) => item.id));
      return;
    }

    await Promise.all(unreadIDs.map((id) => markNotificationRead(id)));
    markLocalRead(unreadIDs);
  } catch (error: unknown) {
    feedbackStore.pushToast(
      axios.isAxiosError<{ message?: string }>(error)
        ? error.response?.data?.message ?? t('account.notifications.updateError')
        : t('account.notifications.updateError'),
      'error',
    );
  }
};

// 12. 標記單條通知已讀並跳轉相關模組
const handleItemClick = async (item: NoticeItem): Promise<void> => {
  if (item.unread) {
    try {
      await markNotificationRead(item.id);
      markLocalRead([item.id]);
    } catch (error: unknown) {
      feedbackStore.pushToast(
        axios.isAxiosError<{ message?: string }>(error)
          ? error.response?.data?.message ?? t('account.notifications.updateError')
          : t('account.notifications.updateError'),
        'error',
      );
      return;
    }
  }

  if (item.targetPath) {
    await router.push(item.targetPath);
  }
};

onMounted(() => {
  void loadNotifications();
});
</script>

<template>
  <div
    id="page-notif"
    class="page"
  >
    <div class="work-shell">
      <aside class="work-sidebar">
        <h1>{{ t('account.notifications.title') }}</h1>
        <p>{{ t('account.notifications.description') }}</p>
        <nav class="work-nav">
          <button
            v-for="panel in panels"
            :key="panel.key"
            type="button"
            class="work-nav-item"
            :class="activeFilter === panel.key ? 'on' : ''"
            :data-work-target="panel.key"
            @click="activeFilter = panel.key"
          >
            <span>{{ panel.navLabel }}</span>
            <span
              v-if="panel.chipCount !== undefined"
              class="work-chip"
              :class="panel.chipTone"
            >
              {{ panel.chipCount }}
            </span>
          </button>
        </nav>
      </aside>

      <main class="work-main">
        <div
          v-for="panel in panels"
          :key="panel.key"
          class="work-panel"
          :class="activeFilter === panel.key ? 'on' : ''"
          :data-work-panel="panel.key"
        >
          <section class="work-hero">
            <div>
              <div class="work-kicker">{{ panel.kicker }}</div>
              <h2 class="work-title">{{ panel.title }}</h2>
              <p class="work-desc">{{ panel.desc }}</p>
            </div>
            <button
              v-if="panel.showMarkAllRead"
              type="button"
              class="work-action secondary"
              @click="markAllRead(panel)"
            >
              {{ t('account.notifications.markAllRead') }}
            </button>
          </section>

          <section
            v-if="loading"
            class="work-card"
          >
            <div class="notice-empty">
              {{ t('account.notifications.loading') }}
            </div>
          </section>

          <section
            v-else-if="panel.groups.length === 0"
            class="work-card"
          >
            <div class="notice-empty">
              <strong>{{ t('account.notifications.emptyTitle') }}</strong>
              <span>{{ t('account.notifications.emptyAllDescription') }}</span>
            </div>
          </section>

          <template v-else>
            <section
              v-for="group in panel.groups"
              :key="group.title"
              class="work-card"
            >
              <div class="work-card-title">
                {{ t(`account.notifications.groups.${group.title}`) }}
              </div>
              <div class="notice-list">
                <article
                  v-for="item in group.items"
                  :key="item.id"
                  class="notice-item"
                  :class="item.unread ? 'unread' : ''"
                  @click="handleItemClick(item)"
                >
                  <div
                    class="notice-dot"
                    :class="item.unread ? '' : 'read'"
                  />
                  <div class="notice-icon">{{ resolveIcon(item.panelKey) }}</div>
                  <div>
                    <div class="notice-title">{{ item.title }}</div>
                    <div class="notice-desc">{{ item.description }}</div>
                  </div>
                  <div class="notice-time">{{ formatNoticeTime(item.createdAt) }}</div>
                </article>
              </div>
            </section>
          </template>
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
/* 1. 頁面外框 */
#page-notif {
  min-height: calc(100svh - 48px);
  background: var(--sur);
}

/* 2. 雙欄主版面 */
.work-shell {
  display: grid;
  grid-template-columns: 240px minmax(0, 1fr);
  gap: 16px;
  max-width: var(--layout-page-max-width);
  margin: 0 auto;
  padding: 24px;
  color: var(--ink);
}

#page-notif .work-shell {
  padding-top: 12px;
  padding-bottom: 16px;
}

/* 3. 側欄 */
.work-sidebar {
  position: sticky;
  top: 72px;
  align-self: start;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 16px;
}

.work-sidebar h1 {
  margin: 0;
  color: var(--accent);
  font-family: var(--font-serif);
  font-size: 26px;
  font-weight: 400;
  line-height: 1.2;
}

.work-sidebar p {
  margin: 8px 0 0;
  color: var(--ink-3);
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
  color: var(--ink-2);
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
  background: var(--brand);
  transform: scaleX(0);
  transform-origin: left center;
  transition: transform 0.24s ease;
}

.work-nav-item:hover {
  background: var(--brand-light);
  color: var(--ink);
}

.work-nav-item.on {
  background: transparent;
  color: var(--accent);
  font-weight: 700;
  outline: none;
}

.work-nav-item.on:hover {
  background: var(--brand-light);
}

.work-nav-item.on::after {
  transform: scaleX(1);
}

/* 5. 標籤晶片 */
.work-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  background: var(--sur-2);
  color: var(--ink-2);
  font-size: 11px;
  font-weight: 600;
  padding: 5px 9px;
  white-space: nowrap;
}

.work-chip.brand {
  background: var(--brand-light);
  color: var(--accent);
}

.work-chip.warn {
  background: var(--warning-bg);
  color: var(--warning);
}

/* 6. 主內容區 */
.work-main {
  display: grid;
  gap: 12px;
  min-width: 0;
  align-content: start;
}

/* 7. 面板 */
.work-panel {
  display: none;
}

.work-panel.on {
  display: grid;
  gap: 14px;
  align-content: start;
}

/* 8. 標題區 */
.work-hero {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 18px;
  border: 0;
  border-bottom: 1px solid var(--bdr);
  border-radius: 0;
  background: transparent;
  margin: 0;
  padding: 0 0 12px;
}

#page-notif .work-hero {
  margin-top: 0;
  padding-bottom: 12px;
}

#page-notif [data-work-panel^="notif-"] .work-hero {
  padding-bottom: 10px;
}

.work-kicker {
  margin-bottom: 6px;
  color: var(--ink-3);
  font-size: 10px;
  letter-spacing: 1.6px;
  text-transform: uppercase;
}

#page-notif .work-kicker {
  margin-bottom: 4px;
  font-size: 9px;
  letter-spacing: 1.4px;
}

.work-title {
  margin: 0;
  color: var(--ink);
  font-size: 22px;
  font-weight: 600;
  line-height: 1.25;
}

#page-notif .work-title {
  font-size: 20px;
}

.work-desc {
  max-width: 560px;
  margin: 8px 0 0;
  color: var(--ink-3);
  font-size: 13px;
  line-height: 1.7;
}

#page-notif .work-desc {
  margin-top: 5px;
  line-height: 1.5;
}

/* 9. 操作按鈕 */
.work-action {
  border: 0;
  border-radius: 6px;
  background: var(--accent);
  color: #fff;
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  padding: 10px 14px;
  white-space: nowrap;
}

.work-action.secondary {
  border: 1px solid var(--bdr);
  background: rgb(var(--color-surface));
  color: var(--ink);
}

/* 10. 卡片區 */
.work-card {
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 16px;
}

.work-card-title {
  margin-bottom: 10px;
  color: var(--ink);
  font-size: 14px;
  font-weight: 600;
}

/* 11. 通知列表 */
.notice-list {
  display: grid;
  gap: 8px;
}

.notice-item {
  display: grid;
  grid-template-columns: 10px 42px minmax(0, 1fr) auto;
  align-items: start;
  gap: 12px;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
  padding: 13px 14px;
  cursor: pointer;
}

.notice-item:hover {
  border-color: var(--brand-mid);
}

.notice-item.unread {
  border-left: 3px solid var(--accent);
  background: #fffaf7;
}

.notice-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--accent);
  margin-top: 7px;
}

.notice-dot.read {
  background: var(--bdr);
}

.notice-icon {
  display: flex;
  width: 36px;
  height: 36px;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: var(--brand-light);
  color: var(--accent);
  font-size: 13px;
  font-weight: 700;
}

.notice-title {
  color: var(--ink);
  font-size: 13px;
  font-weight: 600;
}

.notice-desc {
  margin-top: 4px;
  color: var(--ink-3);
  font-size: 12px;
  line-height: 1.6;
}

.notice-time {
  color: var(--ink-4);
  font-size: 11px;
  white-space: nowrap;
}

/* 12. 空狀態 */
.notice-empty {
  display: grid;
  min-height: 128px;
  align-content: center;
  justify-items: center;
  gap: 8px;
  color: var(--ink-3);
  font-size: 13px;
  line-height: 1.6;
  text-align: center;
}

.notice-empty strong {
  color: var(--ink);
  font-size: 15px;
  font-weight: 600;
}

.notice-empty span {
  max-width: 360px;
}

/* 13. 響應式 - 平板 */
@media (max-width: 1023px) {
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

/* 14. 響應式 - 手機 */
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
