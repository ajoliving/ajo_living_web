<!--
 * 通知與會話中心頁。
 * 1. 桌面以左側模式欄切換「會話」與「通知」，避免通知分類與聊天列表重疊。
 * 2. 通知模式使用資料夾欄加郵件式時間列表；會話模式直接使用現有聊天頁。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

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
  title: string;
  showMarkAllRead: boolean;
}

interface Panel extends PanelConfig {
  unreadCount: number;
  groups: NoticeGroup[];
}

const router = useRouter();
const route = useRoute();
const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const preferenceStore = usePreferenceStore();
const activeFilter = ref<PanelKey>('notif-all');
const activeSection = ref<CommunicationSection>(
  route.query.tab === 'conversations' ? 'conversations' : 'notifications',
);
const loading = ref(false);
const notifications = ref<NoticeItem[]>([]);
const skeletonRows = [1, 2, 3, 4, 5, 6, 7, 8];

// 1. 切換通知分類
const selectNotificationPanel = (panel: PanelKey): void => {
  activeFilter.value = panel;
  activeSection.value = 'notifications';
  if (route.query.tab) {
    void router.replace({ path: '/notifications', query: {} });
  }
};

// 2. 進入會話
const enterConversations = async (): Promise<void> => {
  if (activeSection.value === 'conversations') {
    return;
  }

  activeSection.value = 'conversations';
  await router.replace({
    path: '/notifications',
    query: { tab: 'conversations' },
  });
};

// 3. 進入通知
const enterNotifications = async (): Promise<void> => {
  if (activeSection.value === 'notifications' && !route.query.tab) {
    return;
  }

  activeSection.value = 'notifications';
  await router.replace({ path: '/notifications', query: {} });
};

watch(
  () => route.query.tab,
  (tab) => {
    activeSection.value = tab === 'conversations' ? 'conversations' : 'notifications';
  },
);

// 4. 通知分類設定
const panelConfigs = computed<PanelConfig[]>(() => [
  {
    key: 'notif-all',
    navLabel: t('account.notifications.allTitle'),
    icon: 'bell',
    chipTone: 'brand',
    title: t('account.notifications.allTitle'),
    showMarkAllRead: true,
  },
  {
    key: 'notif-offers',
    navLabel: t('account.notifications.offers'),
    icon: 'tag',
    title: t('account.notifications.offers'),
    showMarkAllRead: true,
  },
  {
    key: 'notif-property',
    navLabel: t('account.notifications.property'),
    icon: 'home',
    title: t('account.notifications.property'),
    showMarkAllRead: true,
  },
  {
    key: 'notif-payment',
    navLabel: t('account.notifications.payment'),
    icon: 'credit-card',
    chipTone: 'warn',
    title: t('account.notifications.payment'),
    showMarkAllRead: true,
  },
  {
    key: 'notif-system',
    navLabel: t('account.notifications.system'),
    icon: 'settings',
    title: t('account.notifications.system'),
    showMarkAllRead: false,
  },
]);

const groupOrder: NoticeGroupKey[] = ['today', 'yesterday', 'earlier'];
const relativeTimeFormatter = computed(() =>
  new Intl.RelativeTimeFormat(preferenceStore.locale, { numeric: 'auto' }),
);

// 5. 根據通知內容判斷分類
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

// 6. 取得通知跳轉目標
const resolveTargetPath = (item: NotificationItem, panelKey: PanelKey): string => {
  if (item.related_type === 'chat' && item.related_public_id) {
    return `/notifications?tab=conversations&conversationId=${encodeURIComponent(item.related_public_id)}`;
  }
  if (item.related_type === 'ismart_subaccount_authorization' && item.related_public_id) {
    return `/auth/ismart-subaccount-authorizations/confirm?authorization_id=${encodeURIComponent(item.related_public_id)}`;
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
    const unreadCount = items.filter((item) => item.unread).length;

    return {
      ...config,
      unreadCount,
      groups,
    };
  }),
);

const activePanel = computed(() => panels.value.find((panel) => panel.key === activeFilter.value) ?? panels.value[0]);
const totalUnreadNotifications = computed(() => notifications.value.filter((item) => item.unread).length);

// 11. 接收聊天頁上報的會話未讀總數
const totalUnreadChats = ref(0);
const handleUnreadCountChange = (count: number): void => {
  totalUnreadChats.value = count;
};

// 12. 讀取通知列表
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

// 16. 取得分類標籤與圖示
const folderLabel = (key: PanelKey): string =>
  panelConfigs.value.find((panel) => panel.key === key)?.navLabel ?? '';

const folderIcon = (key: PanelKey): IconName =>
  panelConfigs.value.find((panel) => panel.key === key)?.icon ?? 'bell';

onMounted(() => {
  void loadNotifications();
});
</script>

<template>
  <div
    class="notifications-page"
    :aria-label="t('account.notifications.title')"
  >
    <nav
      class="inbox-rail"
      :aria-label="t('account.notifications.title')"
    >
      <button
        type="button"
        class="inbox-rail__item"
        :class="activeSection === 'conversations' ? 'is-active' : ''"
        :aria-pressed="activeSection === 'conversations'"
        :title="t('account.notifications.sectionConversations')"
        @click="enterConversations"
      >
        <span class="inbox-rail__icon-wrap">
          <AppIcon
            name="message"
            :size="20"
          />
          <span
            v-if="totalUnreadChats > 0"
            class="inbox-rail__badge"
          >
            {{ totalUnreadChats }}
          </span>
        </span>
        <span>{{ t('account.notifications.sectionConversations') }}</span>
      </button>
      <button
        type="button"
        class="inbox-rail__item"
        :class="activeSection === 'notifications' ? 'is-active' : ''"
        :aria-pressed="activeSection === 'notifications'"
        :title="t('account.notifications.sectionNotifications')"
        @click="enterNotifications"
      >
        <span class="inbox-rail__icon-wrap">
          <AppIcon
            name="bell"
            :size="20"
          />
          <span
            v-if="totalUnreadNotifications > 0"
            class="inbox-rail__badge"
          >
            {{ totalUnreadNotifications }}
          </span>
        </span>
        <span>{{ t('account.notifications.sectionNotifications') }}</span>
      </button>
    </nav>

    <div class="inbox-stage">
      <div class="inbox-switch">
        <button
          type="button"
          class="inbox-switch__item"
          :class="activeSection === 'conversations' ? 'is-active' : ''"
          :aria-pressed="activeSection === 'conversations'"
          @click="enterConversations"
        >
          <span>{{ t('account.notifications.sectionConversations') }}</span>
          <span
            v-if="totalUnreadChats > 0"
            class="inbox-switch__badge"
          >
            {{ totalUnreadChats }}
          </span>
        </button>
        <button
          type="button"
          class="inbox-switch__item"
          :class="activeSection === 'notifications' ? 'is-active' : ''"
          :aria-pressed="activeSection === 'notifications'"
          @click="enterNotifications"
        >
          <span>{{ t('account.notifications.sectionNotifications') }}</span>
          <span
            v-if="totalUnreadNotifications > 0"
            class="inbox-switch__badge"
          >
            {{ totalUnreadNotifications }}
          </span>
        </button>
      </div>

      <main
        v-if="activeSection === 'conversations'"
        class="notifications-chat"
      >
        <MarketplaceChatPage @unread-count-change="handleUnreadCountChange" />
      </main>

      <main
        v-else
        class="notice-layout"
      >
        <nav
          class="notice-folders"
          :aria-label="t('account.notifications.sectionNotifications')"
        >
          <button
            v-for="panel in panels"
            :key="panel.key"
            type="button"
            class="notice-folder"
            :class="activeFilter === panel.key ? 'is-active' : ''"
            :aria-pressed="activeFilter === panel.key"
            @click="selectNotificationPanel(panel.key)"
          >
            <span class="notice-folder__icon">
              <AppIcon
                :name="panel.icon"
                :size="15"
              />
            </span>
            <span class="notice-folder__label">{{ panel.navLabel }}</span>
            <span
              v-if="panel.unreadCount > 0"
              class="notice-folder__count"
              :class="panel.chipTone === 'warn' ? 'is-warn' : ''"
            >
              {{ panel.unreadCount }}
            </span>
          </button>
        </nav>

        <section
          v-if="activePanel"
          class="notice-board"
        >
          <header class="notice-board__header">
            <div class="notice-board__heading">
              <h2>{{ activePanel.title }}</h2>
              <span
                v-if="activePanel.unreadCount > 0"
                class="notice-board__unread"
              >
                {{ activePanel.unreadCount }} {{ t('account.notifications.unread') }}
              </span>
            </div>
            <button
              v-if="activePanel.showMarkAllRead && activePanel.unreadCount > 0"
              type="button"
              class="notice-board__action"
              @click="markAllRead(activePanel)"
            >
              <AppIcon
                name="check"
                :size="15"
              />
              {{ t('account.notifications.markAllRead') }}
            </button>
          </header>

          <div
            v-if="loading"
            class="notice-skeleton"
            role="status"
            :aria-label="t('account.notifications.loading')"
          >
            <div
              v-for="row in skeletonRows"
              :key="row"
              class="notice-skeleton__row"
            >
              <span class="notice-skeleton__icon" />
              <span class="notice-skeleton__lines">
                <span />
                <span />
              </span>
            </div>
          </div>

          <div
            v-else-if="activePanel.groups.length === 0"
            class="notice-empty"
          >
            <span class="notice-empty__icon">
              <AppIcon
                name="inbox"
                :size="28"
              />
            </span>
            <strong>{{ t('account.notifications.emptyTitle') }}</strong>
            <span>{{ t('account.notifications.emptyAllDescription') }}</span>
          </div>

          <div
            v-else
            class="notice-board__scroll"
          >
            <section
              v-for="group in activePanel.groups"
              :key="group.title"
              class="notice-group"
            >
              <h3 class="notice-group__title">
                {{ t(`account.notifications.groups.${group.title}`) }}
              </h3>
              <div class="notice-list">
                <button
                  v-for="item in group.items"
                  :key="item.id"
                  type="button"
                  class="notice-row"
                  :class="item.unread ? 'is-unread' : ''"
                  @click="handleItemClick(item)"
                >
                  <span
                    class="notice-row__accent"
                    :class="item.unread ? 'is-on' : ''"
                  />
                  <span
                    class="notice-row__icon"
                    :class="item.panelKey === 'notif-payment' ? 'is-warn' : ''"
                  >
                    <AppIcon
                      :name="folderIcon(item.panelKey)"
                      :size="16"
                    />
                  </span>
                  <span class="notice-row__body">
                    <span class="notice-row__topline">
                      <small v-if="activeFilter === 'notif-all'">{{ folderLabel(item.panelKey) }}</small>
                      <strong>{{ item.title }}</strong>
                    </span>
                    <span class="notice-row__summary">{{ item.description }}</span>
                  </span>
                  <time>{{ formatNoticeTime(item.createdAt) }}</time>
                </button>
              </div>
            </section>
          </div>
        </section>
      </main>
    </div>
  </div>
</template>

<style scoped>
.notifications-page {
  display: grid;
  grid-template-columns: 76px minmax(0, 1fr);
  height: calc(100svh - var(--app-header-offset));
  overflow: hidden;
  background: rgb(var(--color-surface));
}

.inbox-rail {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px 8px;
  border-right: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface-muted));
}

.inbox-rail__item {
  display: grid;
  gap: 6px;
  justify-items: center;
  width: 100%;
  border: 0;
  border-radius: 10px;
  background: transparent;
  padding: 10px 6px;
  color: rgb(var(--color-text-muted));
  cursor: pointer;
  font-family: inherit;
  font-size: 11px;
  font-weight: 700;
  line-height: 1.2;
  text-align: center;
}

.inbox-rail__item:hover,
.inbox-rail__item:focus-visible {
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  outline: none;
}

.inbox-rail__item.is-active {
  background: rgb(var(--color-surface));
  color: rgb(var(--color-primary));
  box-shadow: inset 0 0 0 1px rgb(var(--color-primary) / 0.18);
}

.inbox-rail__icon-wrap {
  position: relative;
  display: inline-flex;
  width: 28px;
  height: 28px;
  align-items: center;
  justify-content: center;
}

.inbox-rail__badge,
.inbox-switch__badge,
.notice-folder__count {
  display: inline-flex;
  min-width: 18px;
  height: 18px;
  align-items: center;
  justify-content: center;
  border-radius: 9px;
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
  font-size: 10px;
  font-weight: 700;
  padding: 0 5px;
}

.inbox-rail__badge {
  position: absolute;
  top: -6px;
  right: -8px;
}

.inbox-stage {
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
}

.inbox-switch {
  display: none;
}

.notice-layout,
.notifications-chat {
  flex: 1;
  min-width: 0;
  min-height: 0;
}

.notice-layout {
  display: grid;
  grid-template-columns: 232px minmax(0, 1fr);
}

.notice-folders {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-height: 0;
  overflow: auto;
  padding: 16px 12px;
  border-right: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface-muted));
}

.notice-folder {
  display: grid;
  grid-template-columns: 28px minmax(0, 1fr) auto;
  gap: 8px;
  align-items: center;
  width: 100%;
  min-height: 40px;
  border: 0;
  border-radius: 8px;
  background: transparent;
  padding: 8px;
  color: rgb(var(--color-text-muted));
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  font-weight: 600;
  text-align: left;
}

.notice-folder:hover,
.notice-folder:focus-visible {
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  outline: none;
}

.notice-folder.is-active {
  background: rgb(var(--color-surface));
  color: rgb(var(--color-primary));
  box-shadow: inset 0 0 0 1px rgb(var(--color-primary) / 0.16);
}

.notice-folder__icon {
  display: inline-flex;
  width: 28px;
  height: 28px;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  background: rgb(var(--color-surface));
  color: inherit;
}

.notice-folder__label {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.notice-folder__count {
  background: rgb(var(--color-primary) / 0.14);
  color: rgb(var(--color-primary));
}

.notice-folder__count.is-warn {
  background: rgb(var(--color-warning) / 0.16);
  color: rgb(var(--color-warning));
}

.notice-board {
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
  background: rgb(var(--color-surface));
}

.notice-board__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 56px;
  padding: 12px 20px;
  border-bottom: 1px solid rgb(var(--color-border));
}

.notice-board__heading {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  min-width: 0;
}

.notice-board__heading h2 {
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 16px;
  font-weight: 700;
}

.notice-board__unread {
  display: inline-flex;
  align-items: center;
  min-height: 22px;
  border-radius: 999px;
  background: rgb(var(--color-primary) / 0.1);
  padding: 0 8px;
  color: rgb(var(--color-primary));
  font-size: 11px;
  font-weight: 700;
}

.notice-board__action {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-height: 34px;
  flex-shrink: 0;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: rgb(var(--color-surface));
  padding: 6px 12px;
  color: rgb(var(--color-text));
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  font-weight: 700;
}

.notice-board__action:hover,
.notice-board__action:focus-visible {
  border-color: rgb(var(--color-primary) / 0.45);
  color: rgb(var(--color-primary));
  outline: none;
}

.notice-board__scroll {
  flex: 1;
  min-height: 0;
  overflow: auto;
}

.notice-skeleton,
.notice-empty {
  display: flex;
  flex-direction: column;
  flex: 1;
}

.notice-skeleton {
  gap: 0;
  padding: 8px 0;
}

.notice-skeleton__row {
  display: grid;
  grid-template-columns: 36px minmax(0, 1fr);
  gap: 12px;
  align-items: center;
  padding: 14px 20px;
}

.notice-skeleton__icon,
.notice-skeleton__lines span {
  display: block;
  border-radius: 8px;
  background: rgb(var(--color-border));
  animation: notice-pulse 1.2s ease-in-out infinite;
}

.notice-skeleton__icon {
  width: 36px;
  height: 36px;
}

.notice-skeleton__lines {
  display: grid;
  gap: 8px;
}

.notice-skeleton__lines span:first-child {
  width: 42%;
  height: 12px;
}

.notice-skeleton__lines span:last-child {
  width: 76%;
  height: 10px;
}

.notice-empty {
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 40px 20px;
  color: rgb(var(--color-text-muted));
  text-align: center;
}

.notice-empty__icon {
  display: inline-flex;
  width: 56px;
  height: 56px;
  align-items: center;
  justify-content: center;
  border-radius: 16px;
  background: rgb(var(--color-surface-muted));
  color: rgb(var(--color-text-muted));
  margin-bottom: 4px;
}

.notice-empty strong {
  color: rgb(var(--color-text));
  font-size: 15px;
}

.notice-empty span {
  max-width: 28em;
  font-size: 13px;
  line-height: 1.6;
}

.notice-group__title {
  position: sticky;
  top: 0;
  z-index: 1;
  margin: 0;
  background: rgb(var(--color-surface));
  padding: 12px 20px 8px;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  font-weight: 700;
}

.notice-list {
  display: flex;
  flex-direction: column;
}

.notice-row {
  display: grid;
  grid-template-columns: 4px 36px minmax(0, 1fr) auto;
  gap: 12px;
  align-items: start;
  width: 100%;
  border: 0;
  border-bottom: 1px solid rgb(var(--color-border) / 0.7);
  background: transparent;
  padding: 14px 20px 14px 12px;
  cursor: pointer;
  font: inherit;
  text-align: left;
}

.notice-row:hover,
.notice-row:focus-visible {
  background: rgb(var(--color-surface-muted));
  outline: none;
}

.notice-row.is-unread {
  background: rgb(var(--color-primary) / 0.04);
}

.notice-row__accent {
  width: 4px;
  min-height: 36px;
  border-radius: 2px;
  background: transparent;
}

.notice-row__accent.is-on {
  background: rgb(var(--color-primary));
}

.notice-row__icon {
  display: inline-flex;
  width: 36px;
  height: 36px;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  background: rgb(var(--color-primary) / 0.1);
  color: rgb(var(--color-primary));
}

.notice-row__icon.is-warn {
  background: rgb(var(--color-warning) / 0.12);
  color: rgb(var(--color-warning));
}

.notice-row__body {
  display: grid;
  gap: 4px;
  min-width: 0;
}

.notice-row__topline {
  display: grid;
  gap: 2px;
  min-width: 0;
}

.notice-row__topline small {
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  font-weight: 700;
}

.notice-row__topline strong {
  overflow: hidden;
  color: rgb(var(--color-text));
  font-size: 14px;
  font-weight: 600;
  line-height: 1.4;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.notice-row.is-unread .notice-row__topline strong {
  font-weight: 700;
}

.notice-row__summary {
  display: -webkit-box;
  overflow: hidden;
  color: rgb(var(--color-text-muted));
  font-size: 13px;
  line-height: 1.5;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.notice-row time {
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  font-weight: 600;
  white-space: nowrap;
}

.notifications-chat {
  overflow: hidden;
}

.notifications-chat :deep(.marketplace-chat-page),
.notifications-chat :deep(.chat-shell),
.notifications-chat :deep(.chat-layout) {
  height: 100%;
  min-height: 0;
}

.notifications-chat :deep(.chat-shell) {
  padding: 0;
}

.notifications-chat :deep(.chat-layout) {
  max-width: none;
  gap: 0;
  border: 0;
  border-radius: 0;
  box-shadow: none;
}

.notifications-chat :deep(.chat-sidebar),
.notifications-chat :deep(.chat-main) {
  height: 100%;
  align-self: stretch;
  border-radius: 0;
  box-shadow: none;
}

.notifications-chat :deep(.chat-sidebar) {
  border: 0;
  border-right: 1px solid rgb(var(--color-border));
}

.notifications-chat :deep(.chat-main) {
  border: 0;
}

@keyframes notice-pulse {
  0%,
  100% {
    opacity: 0.55;
  }
  50% {
    opacity: 0.22;
  }
}

@media (max-width: 1023px) {
  .notifications-page {
    grid-template-columns: 1fr;
    height: calc(100svh - var(--app-header-offset) - var(--app-mobile-content-bottom));
  }

  .inbox-rail {
    display: none;
  }

  .inbox-switch {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 6px;
    padding: 10px 12px 8px;
    border-bottom: 1px solid rgb(var(--color-border));
    background: rgb(var(--color-surface));
  }

  .inbox-switch__item {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    min-height: 40px;
    border: 0;
    border-radius: 8px;
    background: rgb(var(--color-surface-muted));
    color: rgb(var(--color-text-muted));
    cursor: pointer;
    font-family: inherit;
    font-size: 13px;
    font-weight: 700;
  }

  .inbox-switch__item.is-active {
    background: rgb(var(--color-primary) / 0.1);
    color: rgb(var(--color-primary));
  }

  .notice-layout {
    grid-template-columns: 1fr;
    grid-template-rows: auto minmax(0, 1fr);
  }

  .notice-folders {
    flex-direction: row;
    gap: 8px;
    overflow-x: auto;
    padding: 10px 12px;
    border-right: 0;
    border-bottom: 1px solid rgb(var(--color-border));
    background: rgb(var(--color-surface));
    scrollbar-width: none;
  }

  .notice-folders::-webkit-scrollbar {
    display: none;
  }

  .notice-folder {
    flex: 0 0 auto;
    width: auto;
    grid-template-columns: 28px auto auto;
    min-height: 36px;
    padding: 6px 10px;
    background: rgb(var(--color-surface-muted));
  }

  .notice-folder.is-active {
    background: rgb(var(--color-primary) / 0.08);
  }

  .notifications-chat :deep(.chat-layout),
  .notifications-chat :deep(.chat-layout--thread-open) {
    height: 100%;
    min-height: 0;
  }

  .notifications-chat :deep(.chat-sidebar) {
    height: 100%;
    border-right: 0;
  }
}

@media (max-width: 767px) {
  .notice-board__header {
    align-items: flex-start;
    flex-direction: column;
    padding: 12px;
  }

  .notice-board__action {
    width: 100%;
    justify-content: center;
  }

  .notice-row {
    grid-template-columns: 4px 32px minmax(0, 1fr);
    padding: 12px 12px 12px 8px;
  }

  .notice-row time {
    grid-column: 3;
  }

  .notice-group__title {
    padding: 12px 12px 8px;
  }
}
</style>
