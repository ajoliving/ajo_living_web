<!--
 * 會員通知中心頁。
 * 1. 讀取通知 API 並按分類展示。
 * 2. 對齊設計稿提供側欄篩選與時間分組。
 * 3. 支援單條與全部標記已讀。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import {
  fetchNotifications,
  markAllNotificationsRead,
  markNotificationRead,
} from '@/httpapis/notifications';
import type { NotificationItem } from '@/model/notification';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';

type NotificationFilter = 'all' | 'offers' | 'system' | 'property' | 'payment';
type NotificationTone = 'orange' | 'blue' | 'green' | 'gray';
type NotificationGroup = 'today' | 'yesterday' | 'earlier';

interface NotificationViewItem {
  id: string;
  filter: NotificationFilter;
  group: NotificationGroup;
  tone: NotificationTone;
  title: string;
  description: string;
  createdAt: string;
  unread: boolean;
}

const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const preferenceStore = usePreferenceStore();

const loading = ref(false);
const activeFilter = ref<NotificationFilter>('all');
const notifications = ref<NotificationViewItem[]>([]);

const filterItems = computed<Array<{ key: NotificationFilter; label: string }>>(() => [
  { key: 'all', label: t('account.notifications.all') },
  { key: 'offers', label: t('account.notifications.offers') },
  { key: 'system', label: t('account.notifications.system') },
  { key: 'property', label: t('account.notifications.property') },
  { key: 'payment', label: t('account.notifications.payment') },
]);

const groupOrder: NotificationGroup[] = ['today', 'yesterday', 'earlier'];

const filteredNotifications = computed(() =>
  activeFilter.value === 'all'
    ? notifications.value
    : notifications.value.filter((item) => item.filter === activeFilter.value),
);

const unreadCount = computed(() => notifications.value.filter((item) => item.unread).length);

const activeTitle = computed(() =>
  filterItems.value.find((item) => item.key === activeFilter.value)?.label ?? t('account.notifications.all'),
);

// 1. 映射通知類型
const resolveFilter = (item: NotificationItem): NotificationFilter => {
  const value = `${item.category} ${item.related_type} ${item.title}`.toLowerCase();

  if (value.includes('payment') || value.includes('bill') || value.includes('fee')) {
    return 'payment';
  }
  if (value.includes('offer') || value.includes('coupon') || value.includes('discount')) {
    return 'offers';
  }
  if (value.includes('property') || value.includes('listing') || value.includes('viewing')) {
    return 'property';
  }

  return 'system';
};

// 2. 映射通知圖示
const resolveIcon = (filter: NotificationFilter): string => {
  const iconKeyMap: Record<NotificationFilter, string> = {
    all: 'account.notifications.icons.all',
    offers: 'account.notifications.icons.offers',
    system: 'account.notifications.icons.system',
    property: 'account.notifications.icons.property',
    payment: 'account.notifications.icons.payment',
  };

  return t(iconKeyMap[filter]);
};

// 3. 映射通知色調
const resolveTone = (filter: NotificationFilter): NotificationTone => {
  const toneMap: Record<NotificationFilter, NotificationTone> = {
    all: 'gray',
    offers: 'orange',
    system: 'blue',
    property: 'green',
    payment: 'blue',
  };

  return toneMap[filter];
};

// 4. 格式化時間分組
const resolveGroup = (value: string): NotificationGroup => {
  const createdAt = new Date(value);
  const now = new Date();
  const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const startOfYesterday = new Date(startOfToday);
  startOfYesterday.setDate(startOfToday.getDate() - 1);

  if (createdAt >= startOfToday) {
    return 'today';
  }
  if (createdAt >= startOfYesterday) {
    return 'yesterday';
  }

  return 'earlier';
};

// 5. 格式化時間文字
const resolveTime = (value: string): string => {
  const createdAt = new Date(value);
  if (Number.isNaN(createdAt.getTime())) {
    return value;
  }

  return createdAt.toLocaleString(preferenceStore.locale, {
    month: 'numeric',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
};

// 6. 轉換通知資料
const mapNotificationItem = (item: NotificationItem): NotificationViewItem => {
  const filter = resolveFilter(item);

  return {
    id: item.notification_id,
    filter,
    group: resolveGroup(item.created_at),
    tone: resolveTone(filter),
    title: item.title,
    description: item.body,
    createdAt: item.created_at,
    unread: !item.is_read,
  };
};

// 7. 讀取通知列表
const loadNotifications = async (): Promise<void> => {
  loading.value = true;

  try {
    const response = await fetchNotifications({ page: 1, page_size: 50 });
    notifications.value = response.data.data.items.map(mapNotificationItem);
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

// 8. 取得指定分類未讀數
const unreadCountFor = (filter: NotificationFilter): number =>
  notifications.value.filter((item) => item.unread && (filter === 'all' || item.filter === filter)).length;

// 9. 取得分組通知
const groupItems = (group: NotificationGroup): NotificationViewItem[] =>
  filteredNotifications.value.filter((item) => item.group === group);

// 10. 標記全部通知已讀
const handleMarkAllRead = async (): Promise<void> => {
  try {
    await markAllNotificationsRead();
    notifications.value = notifications.value.map((item) => ({ ...item, unread: false }));
  } catch (error: unknown) {
    feedbackStore.pushToast(
      axios.isAxiosError<{ message?: string }>(error)
        ? error.response?.data?.message ?? t('account.notifications.updateError')
        : t('account.notifications.updateError'),
      'error',
    );
  }
};

// 11. 標記單條通知已讀
const handleMarkRead = async (id: string): Promise<void> => {
  const target = notifications.value.find((item) => item.id === id);
  if (!target?.unread) {
    return;
  }

  try {
    await markNotificationRead(id);
    notifications.value = notifications.value.map((item) =>
      item.id === id ? { ...item, unread: false } : item,
    );
  } catch (error: unknown) {
    feedbackStore.pushToast(
      axios.isAxiosError<{ message?: string }>(error)
        ? error.response?.data?.message ?? t('account.notifications.updateError')
        : t('account.notifications.updateError'),
      'error',
    );
  }
};

onMounted(() => {
  void loadNotifications();
});
</script>

<template>
  <main class="notif-wrap">
    <aside class="notif-sidebar">
      <div class="notif-sidebar-title">
        {{ t('account.notifications.title') }}
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
          <p>{{ t('account.notifications.title') }}</p>
          <h1>{{ activeFilter === 'all' ? t('account.notifications.allTitle') : activeTitle }}</h1>
        </div>
        <button
          type="button"
          :disabled="unreadCount === 0 || loading"
          @click="handleMarkAllRead"
        >
          {{ t('account.notifications.markAllRead') }}
        </button>
      </header>

      <div
        v-if="loading"
        class="notif-empty"
      >
        <div>{{ t('account.notifications.title') }}</div>
        <strong>{{ t('account.notifications.loadingShort') }}</strong>
        <span>{{ t('account.notifications.loadingHint') }}</span>
      </div>

      <template
        v-else
        v-for="group in groupOrder"
        :key="group"
      >
        <section
          v-if="groupItems(group).length > 0"
          class="notif-group"
        >
          <div class="notif-group-title">
            {{ t(`account.notifications.groups.${group}`) }}
          </div>
          <div class="notif-list">
            <article
              v-for="item in groupItems(group)"
              :key="item.id"
              class="notif-item"
              :class="item.unread ? 'notif-item--unread' : ''"
              @click="handleMarkRead(item.id)"
            >
              <div
                class="notif-dot"
                :class="item.unread ? '' : 'notif-dot--read'"
              />
              <div
                class="notif-icon"
                :class="`notif-icon--${item.tone}`"
              >
                {{ resolveIcon(item.filter) }}
              </div>
              <div class="notif-content">
                <div class="notif-title">
                  {{ item.title }}
                </div>
                <div class="notif-desc">
                  {{ item.description }}
                </div>
                <div class="notif-time">
                  {{ resolveTime(item.createdAt) }}
                </div>
              </div>
              <button
                v-if="item.unread"
                type="button"
                class="notif-read-button"
                @click.stop="handleMarkRead(item.id)"
              >
                {{ t('account.notifications.markRead') }}
              </button>
            </article>
          </div>
        </section>
      </template>

      <section
        v-if="!loading && filteredNotifications.length === 0"
        class="notif-empty"
      >
        <div>{{ t('account.notifications.title') }}</div>
        <strong>{{ t('account.notifications.emptyTitle') }}</strong>
        <span>{{ t('account.notifications.emptyAllDescription') }}</span>
      </section>
    </section>
  </main>
</template>

<style scoped>
.notif-wrap {
  display: grid;
  min-height: calc(100svh - var(--app-header-offset, 48px));
  background: #ffffff;
  color: #1a1a1a;
  grid-template-columns: 220px minmax(0, 1fr);
}

.notif-sidebar {
  border-right: 1px solid #e4e4e4;
  background: #ffffff;
  padding: 20px 16px;
}

.notif-sidebar-title {
  margin-bottom: 16px;
  font-family: var(--font-display);
  font-size: 1.2rem;
  font-weight: 400;
}

.notif-filter {
  display: grid;
  gap: 8px;
}

.nf-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border: 1px solid transparent;
  border-radius: 2px;
  background: #ffffff;
  color: #777777;
  cursor: pointer;
  font: inherit;
  font-size: 12px;
  padding: 10px 12px;
  text-align: left;
}

.nf-item--active {
  border-color: #1a1a1a;
  background: #1a1a1a;
  color: #ffffff;
}

.nf-badge {
  min-width: 18px;
  height: 18px;
  border-radius: 999px;
  background: rgb(var(--color-primary));
  color: #ffffff;
  font-size: 10px;
  line-height: 18px;
  text-align: center;
}

.notif-main {
  background: rgb(var(--color-surface-muted));
  padding: 20px;
}

.notif-main-head {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 18px;
}

.notif-main-head p {
  margin: 0 0 6px;
  color: #777777;
  font-size: 10px;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.notif-main-head h1 {
  margin: 0;
  font-family: var(--font-display);
  font-size: 1.9rem;
  font-weight: 400;
}

.notif-main-head button {
  border: 1px solid #cccccc;
  border-radius: 2px;
  background: #ffffff;
  color: #1a1a1a;
  cursor: pointer;
  font: inherit;
  font-size: 12px;
  padding: 9px 12px;
}

.notif-main-head button:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.notif-group {
  margin-top: 14px;
}

.notif-group-title {
  margin-bottom: 10px;
  color: #777777;
  font-size: 11px;
  font-weight: 700;
}

.notif-list {
  display: grid;
  gap: 10px;
}

.notif-item {
  display: grid;
  grid-template-columns: auto auto minmax(0, 1fr) auto;
  gap: 12px;
  align-items: start;
  border: 1px solid #e4e4e4;
  border-radius: 8px;
  background: #ffffff;
  cursor: pointer;
  padding: 14px 16px;
}

.notif-dot {
  width: 10px;
  height: 10px;
  margin-top: 6px;
  border-radius: 999px;
  background: rgb(var(--color-primary));
}

.notif-dot--read {
  background: #d2d2d2;
}

.notif-icon {
  display: flex;
  width: 36px;
  height: 36px;
  align-items: center;
  justify-content: center;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 700;
}

.notif-icon--orange {
  background: #fff0e6;
  color: #c04600;
}

.notif-icon--blue {
  background: #eaf1ff;
  color: #2a5bd7;
}

.notif-icon--green {
  background: #eaf6ec;
  color: #1a7a3a;
}

.notif-icon--gray {
  background: #f0f0f0;
  color: #666666;
}

.notif-title {
  font-size: 13px;
  font-weight: 700;
}

.notif-desc {
  margin-top: 4px;
  color: #555555;
  font-size: 12px;
  line-height: 1.65;
}

.notif-time {
  margin-top: 6px;
  color: #888888;
  font-size: 11px;
}

.notif-read-button {
  align-self: center;
  border: 1px solid #e4e4e4;
  border-radius: 2px;
  background: #ffffff;
  color: #777777;
  cursor: pointer;
  font: inherit;
  font-size: 11px;
  padding: 7px 9px;
}

.notif-empty {
  display: grid;
  place-items: center;
  gap: 8px;
  border: 1px solid #e4e4e4;
  border-radius: 8px;
  background: #ffffff;
  padding: 42px 18px;
  text-align: center;
}

.notif-empty div {
  color: #aaaaaa;
  font-size: 11px;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.notif-empty strong {
  font-size: 15px;
  font-weight: 700;
}

.notif-empty span {
  color: #777777;
  font-size: 12px;
  line-height: 1.6;
}

@media (max-width: 1023px) {
  .notif-wrap {
    grid-template-columns: 1fr;
  }

  .notif-sidebar {
    border-right: 0;
    border-bottom: 1px solid #e4e4e4;
  }
}

@media (max-width: 767px) {
  .notif-main {
    padding: 16px;
  }

  .notif-main-head,
  .notif-item {
    grid-template-columns: 1fr;
  }

  .notif-read-button {
    justify-self: start;
  }
}
</style>
