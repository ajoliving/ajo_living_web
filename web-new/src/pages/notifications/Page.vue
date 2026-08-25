<!--
 * 通知與會話中心頁。
 * 1. 左側簡潔導航：通知分類 + 會話入口。
 * 2. 右側主內容區：通知列表或會話頁面。
 * 3. 會話頁面內置群聊/私聊篩選器，並默認打開第一個對話。
 * 4. 現代卡片式設計，優化視覺層次。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import AppIcon from '@/shared/components/base/AppIcon.vue';
import type { IconName } from '@/shared/components/base/AppIcon.vue';
import {
  fetchNotifications,
  markAllNotificationsRead,
  markNotificationRead,
} from '@/httpapis/notifications';
import type { NotificationItem } from '@/model/notification';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';
import MarketplaceChatPage from '@/pages/marketplace/chat/MarketplaceChatPage.vue';

type PanelKey = 'notif-all' | 'notif-offers' | 'notif-property' | 'notif-payment' | 'notif-system';
type CommunicationSection = 'notifications' | 'conversations';
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
  icon: IconName;
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
const route = router.currentRoute ?? ref({ query: {} as Record<string, string | undefined> });
const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const preferenceStore = usePreferenceStore();
const activeFilter = ref<PanelKey>('notif-all');
const activeSection = ref<CommunicationSection>(route.value.query.tab === 'conversations' ? 'conversations' : 'notifications');
const loading = ref(false);
const notifications = ref<NoticeItem[]>([]);

// 1. 切換通知分類
const selectNotificationPanel = (panel: PanelKey): void => {
  activeFilter.value = panel;
  activeSection.value = 'notifications';
  void router.replace({ path: '/notifications', query: {} });
};

// 2. 進入會話頁面
const enterConversations = async (): Promise<void> => {
  activeSection.value = 'conversations';
  await router.replace({ path: '/notifications', query: { tab: 'conversations' } });
};

watch(
  () => route.value.query.tab,
  (tab) => {
    activeSection.value = tab === 'conversations' ? 'conversations' : 'notifications';
  },
);

// 3. 通知分類面板設定
const panelConfigs = computed<PanelConfig[]>(() => [
  {
    key: 'notif-all',
    navLabel: t('account.notifications.allTitle'),
    icon: 'bell',
    chipTone: 'brand',
    kicker: t('account.notifications.title'),
    title: t('account.notifications.allTitle'),
    desc: t('account.notifications.filters.allDescription'),
    showMarkAllRead: true,
  },
  {
    key: 'notif-offers',
    navLabel: t('account.notifications.offers'),
    icon: 'tag',
    kicker: t('account.notifications.offers'),
    title: t('account.notifications.offers'),
    desc: t('account.notifications.filters.offersDescription'),
    showMarkAllRead: true,
  },
  {
    key: 'notif-property',
    navLabel: t('account.notifications.property'),
    icon: 'home',
    kicker: t('account.notifications.property'),
    title: t('account.notifications.property'),
    desc: t('account.notifications.filters.propertyDescription'),
    showMarkAllRead: true,
  },
  {
    key: 'notif-payment',
    navLabel: t('account.notifications.payment'),
    icon: 'credit-card',
    chipTone: 'warn',
    kicker: t('account.notifications.payment'),
    title: t('account.notifications.payment'),
    desc: t('account.notifications.filters.paymentDescription'),
    showMarkAllRead: true,
  },
  {
    key: 'notif-system',
    navLabel: t('account.notifications.system'),
    icon: 'settings',
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

// 4. 根據通知內容判斷分類
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

// 5. 取得通知圖示
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

// 6. 取得通知跳轉目標
const resolveTargetPath = (item: NotificationItem, panelKey: PanelKey): string => {
  if (item.related_type === 'chat' && item.related_public_id) {
    return `/notifications?tab=conversations&conversationId=${encodeURIComponent(item.related_public_id)}`;
  }

  const routeMap: Partial<Record<PanelKey, string>> = {
    'notif-offers': '/supermarket-offers',
    'notif-property': '/properties',
    'notif-payment': '/payments',
  };

  return routeMap[panelKey] ?? '';
};

// 7. 取得通知分組
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

// 8. 格式化通知時間
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

// 9. 轉換通知資料
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

// 10. 按分類生成面板資料
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

// 11. 讀取通知列表
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

// 13. 更新本地已讀狀態
const markLocalRead = (ids: string[]): void => {
  const targetIDs = new Set(ids);
  notifications.value = notifications.value.map((item) =>
    targetIDs.has(item.id) ? { ...item, unread: false } : item,
  );
};

// 14. 標記當前面板通知已讀
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

// 15. 標記單條通知已讀並跳轉
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
  <div class="notifications-page">
    <div class="notifications-shell">
      <aside class="notifications-sidebar">
        <header class="notifications-sidebar__header">
          <h1>{{ t('account.notifications.title') }}</h1>
          <p>{{ t('account.notifications.description') }}</p>
        </header>

        <nav class="notifications-nav">
          <button
            type="button"
            class="notifications-nav__item notifications-nav__item--conversations"
            :class="activeSection === 'conversations' ? 'active' : ''"
            @click="enterConversations"
          >
            <div class="notifications-nav__item-content">
              <AppIcon
                name="message"
                :size="20"
              />
              <span>{{ t('account.notifications.sectionConversations') }}</span>
            </div>
          </button>

          <div class="notifications-nav__divider" />

          <button
            v-for="panel in panels"
            :key="panel.key"
            type="button"
            class="notifications-nav__item"
            :class="activeSection === 'notifications' && activeFilter === panel.key ? 'active' : ''"
            @click="selectNotificationPanel(panel.key)"
          >
            <div class="notifications-nav__item-content">
              <AppIcon
                :name="panel.icon"
                :size="18"
              />
              <span>{{ panel.navLabel }}</span>
            </div>
            <span
              v-if="panel.chipCount !== undefined"
              class="notifications-nav__badge"
              :class="panel.chipTone"
            >
              {{ panel.chipCount }}
            </span>
          </button>
        </nav>
      </aside>

      <main
        v-if="activeSection === 'notifications'"
        class="notifications-main"
      >
        <div
          v-for="panel in panels"
          :key="panel.key"
          class="notifications-panel"
          :class="activeFilter === panel.key ? 'active' : ''"
        >
          <header class="notifications-panel__header">
            <div>
              <div class="notifications-panel__kicker">{{ panel.kicker }}</div>
              <h2 class="notifications-panel__title">{{ panel.title }}</h2>
              <p class="notifications-panel__desc">{{ panel.desc }}</p>
            </div>
            <button
              v-if="panel.showMarkAllRead && panel.chipCount"
              type="button"
              class="notifications-action"
              @click="markAllRead(panel)"
            >
              <AppIcon
                name="check"
                :size="16"
              />
              {{ t('account.notifications.markAllRead') }}
            </button>
          </header>

          <div
            v-if="loading"
            class="notifications-loading"
          >
            {{ t('account.notifications.loading') }}
          </div>

          <div
            v-else-if="panel.groups.length === 0"
            class="notifications-empty"
          >
            <AppIcon
              name="bell"
              :size="48"
            />
            <strong>{{ t('account.notifications.emptyTitle') }}</strong>
            <span>{{ t('account.notifications.emptyAllDescription') }}</span>
          </div>

          <div
            v-else
            class="notifications-content"
          >
            <section
              v-for="group in panel.groups"
              :key="group.title"
              class="notifications-group"
            >
              <h3 class="notifications-group__title">
                {{ t(`account.notifications.groups.${group.title}`) }}
              </h3>
              <div class="notifications-list">
                <article
                  v-for="item in group.items"
                  :key="item.id"
                  class="notification-card"
                  :class="item.unread ? 'notification-card--unread' : ''"
                  @click="handleItemClick(item)"
                >
                  <div class="notification-card__indicator" />
                  <div class="notification-card__icon">{{ resolveIcon(item.panelKey) }}</div>
                  <div class="notification-card__content">
                    <h4 class="notification-card__title">{{ item.title }}</h4>
                    <p class="notification-card__desc">{{ item.description }}</p>
                  </div>
                  <time class="notification-card__time">{{ formatNoticeTime(item.createdAt) }}</time>
                </article>
              </div>
            </section>
          </div>
        </div>
      </main>

      <main
        v-else
        class="notifications-chat"
      >
        <MarketplaceChatPage />
      </main>
    </div>
  </div>
</template>

<style scoped>
.notifications-page {
  min-height: calc(100svh - 48px);
  background: rgb(var(--color-surface-muted));
}

.notifications-shell {
  display: grid;
  grid-template-columns: 260px minmax(0, 1fr);
  gap: 18px;
  max-width: var(--layout-page-max-width);
  margin: 0 auto;
  padding: 20px var(--layout-page-padding-inline) 48px;
}

.notifications-sidebar {
  position: sticky;
  top: 72px;
  align-self: start;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  box-shadow: 0 1px 3px rgb(0 0 0 / 0.05);
  overflow: hidden;
}

.notifications-sidebar__header {
  padding: 20px 18px 18px;
  border-bottom: 1px solid rgb(var(--color-border));
}

.notifications-sidebar__header h1 {
  margin: 0;
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 22px;
  font-weight: 600;
  line-height: 1.3;
}

.notifications-sidebar__header p {
  margin: 6px 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  line-height: 1.6;
}

.notifications-nav {
  display: grid;
  padding: 8px;
}

.notifications-nav__divider {
  height: 1px;
  background: rgb(var(--color-border));
  margin: 8px 0;
}

.notifications-nav__item {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 34px;
  border: 0;
  border-radius: 0;
  background: transparent;
  padding: 8px 2px;
  color: rgb(var(--color-text-muted));
  cursor: pointer;
  font-family: inherit;
  font-size: 14px;
  font-weight: 600;
  text-align: left;
  transition: background-color 0.15s ease, color 0.15s ease;
}

.notifications-nav__item:hover {
  background: rgb(var(--color-primary) / 0.08);
  color: rgb(var(--color-text));
}

.notifications-nav__item.active {
  background: transparent;
  color: rgb(var(--color-primary));
  font-weight: 700;
  outline: none;
}

.notifications-nav__item::after {
  content: '';
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  height: 2px;
  background: rgb(var(--color-primary));
  transform: scaleX(0);
  transform-origin: left center;
  transition: transform 0.24s ease;
}

.notifications-nav__item.active::after,
.notifications-nav__item:focus-visible::after {
  transform: scaleX(1);
}

.notifications-nav__item:focus-visible {
  color: rgb(var(--color-primary));
  outline: none;
}

.notifications-nav__item.active:hover,
.notifications-nav__item:focus-visible:hover {
  background: rgb(var(--color-primary) / 0.08);
}

.notifications-nav__item-content {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.notifications-nav__badge {
  display: inline-flex;
  min-width: 22px;
  height: 22px;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border-radius: 11px;
  background: rgb(var(--color-surface-raised));
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  font-weight: 700;
  padding: 0 7px;
}

.notifications-nav__badge.brand {
  background: rgb(var(--color-primary) / 0.15);
  color: rgb(var(--color-primary));
}

.notifications-nav__badge.warn {
  background: rgb(255 152 0 / 0.15);
  color: rgb(255 152 0);
}

.notifications-nav__item.active .notifications-nav__badge {
  background: rgb(var(--color-primary) / 0.15);
  color: rgb(var(--color-primary));
}

.notifications-main {
  min-width: 0;
}

.notifications-panel {
  display: none;
}

.notifications-panel.active {
  display: grid;
  gap: 18px;
  align-content: start;
}

.notifications-panel__header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 20px;
  padding-bottom: 16px;
  border-bottom: 2px solid rgb(var(--color-border));
}

.notifications-panel__kicker {
  margin-bottom: 6px;
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.notifications-panel__title {
  margin: 0;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: 26px;
  font-weight: 600;
  line-height: 1.3;
}

.notifications-panel__desc {
  margin: 6px 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 13px;
  line-height: 1.6;
}

.notifications-action {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 38px;
  flex-shrink: 0;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: rgb(var(--color-surface));
  padding: 8px 16px;
  color: rgb(var(--color-text));
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  font-weight: 700;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    transform 0.1s ease;
}

.notifications-action:hover {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary) / 0.05);
  transform: translateY(-1px);
}

.notifications-loading,
.notifications-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  min-height: 320px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 40px 20px;
  color: rgb(var(--color-text-muted));
  text-align: center;
}

.notifications-empty strong {
  color: rgb(var(--color-text));
  font-size: 16px;
  font-weight: 700;
}

.notifications-empty span {
  font-size: 13px;
  line-height: 1.6;
}

.notifications-content {
  display: grid;
  gap: 24px;
}

.notifications-group__title {
  margin: 0 0 12px;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.05em;
  line-height: 1;
  text-transform: uppercase;
}

.notifications-list {
  display: grid;
  gap: 10px;
}

.notification-card {
  position: relative;
  display: grid;
  grid-template-columns: 4px 48px minmax(0, 1fr) auto;
  gap: 14px;
  align-items: start;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 16px 18px;
  cursor: pointer;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    transform 0.1s ease;
}

.notification-card:hover {
  border-color: rgb(var(--color-primary) / 0.3);
  box-shadow: 0 2px 8px rgb(0 0 0 / 0.06);
  transform: translateY(-1px);
}

.notification-card--unread {
  background: rgb(var(--color-primary) / 0.03);
  border-color: rgb(var(--color-primary) / 0.2);
}

.notification-card__indicator {
  width: 4px;
  height: 100%;
  border-radius: 2px;
  background: transparent;
}

.notification-card--unread .notification-card__indicator {
  background: rgb(var(--color-primary));
}

.notification-card__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  flex-shrink: 0;
  border-radius: 6px;
  background: rgb(var(--color-surface-raised));
  font-size: 24px;
}

.notification-card__content {
  min-width: 0;
}

.notification-card__title {
  margin: 0 0 6px;
  color: rgb(var(--color-text));
  font-size: 14px;
  font-weight: 700;
  line-height: 1.4;
}

.notification-card__desc {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 13px;
  line-height: 1.5;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.notification-card__time {
  flex-shrink: 0;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  font-weight: 600;
  white-space: nowrap;
}

.notifications-chat {
  min-width: 0;
  background: transparent;
  overflow: hidden;
}

.notifications-chat :deep(.marketplace-chat-page) {
  background: transparent;
}

.notifications-chat :deep(.chat-shell) {
  padding: 0;
}

.notifications-chat :deep(.chat-layout) {
  max-width: none;
  border: 0;
  border-radius: 0;
  box-shadow: none;
  gap: 18px;
}

@media (max-width: 1023px) {
  .notifications-shell {
    grid-template-columns: 1fr;
    gap: 16px;
    padding-bottom: calc(var(--app-mobile-content-bottom) + 16px);
  }

  .notifications-sidebar {
    position: static;
  }

  .notifications-nav__item {
    min-height: 48px;
  }
}

@media (max-width: 767px) {
  .notifications-panel__header {
    flex-direction: column;
    align-items: flex-start;
  }

  .notifications-action {
    width: 100%;
    justify-content: center;
  }

  .notification-card {
    grid-template-columns: 4px 40px minmax(0, 1fr);
    grid-template-areas:
      "indicator icon content"
      "indicator icon time";
    gap: 12px;
    padding: 14px 16px;
  }

  .notification-card__indicator {
    grid-area: indicator;
  }

  .notification-card__icon {
    grid-area: icon;
    width: 40px;
    height: 40px;
    font-size: 20px;
  }

  .notification-card__content {
    grid-area: content;
  }

  .notification-card__time {
    grid-area: time;
    justify-self: start;
  }
}

:deep(.app-button),
:deep(.app-input-shell) {
  border-radius: 6px;
}
</style>
