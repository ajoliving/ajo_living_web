/*
 * 通知中心語系測試。
 * 1. 驗證分類、操作與分組文案會隨 locale 切換。
 * 2. 驗證已載入通知的相對時間不會保留舊語系快取。
 */
import { flushPromises, mount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { nextTick } from 'vue';

import { fetchNotifications } from '@/httpapis/notifications';
import i18n, { applyLocale } from '@/i18n';
import { usePreferenceStore } from '@/stores/preferences';

import Page from './Page.vue';

vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: vi.fn(),
  }),
}));

vi.mock('@/stores/feedback', () => ({
  useFeedbackStore: () => ({
    pushToast: vi.fn(),
  }),
}));

vi.mock('@/httpapis/notifications', () => ({
  fetchNotifications: vi.fn(),
  markAllNotificationsRead: vi.fn(),
  markNotificationRead: vi.fn(),
}));

const mockedFetchNotifications = vi.mocked(fetchNotifications);

// 1. 建立固定通知回應
const createNotificationsResponse = () => ({
  data: {
    data: {
      items: [
        {
          notification_id: 'notification-1',
          category: 'payment',
          related_type: 'bill',
          related_public_id: '',
          title: 'Payment received',
          body: 'The payment has been recorded.',
          is_read: false,
          created_at: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
        },
      ],
    },
  },
}) as unknown as Awaited<ReturnType<typeof fetchNotifications>>;

describe('NotificationCenterPage locale', () => {
  beforeEach(() => {
    window.localStorage.clear();
    applyLocale('zh-HK');
    mockedFetchNotifications.mockReset();
    mockedFetchNotifications.mockResolvedValue(createNotificationsResponse());
  });

  afterEach(() => {
    applyLocale('zh-HK');
  });

  it('updates loaded notification UI and relative time after switching to English', async () => {
    const pinia = createPinia();
    setActivePinia(pinia);
    const wrapper = mount(Page, {
      global: {
        plugins: [pinia, i18n],
      },
    });
    await flushPromises();

    expect(wrapper.text()).toContain('分鐘前');

    const preferenceStore = usePreferenceStore();
    preferenceStore.setLocale('en');
    applyLocale('en');
    await nextTick();

    expect(wrapper.text()).toContain('All notifications');
    expect(wrapper.text()).toContain('Offer alerts');
    expect(wrapper.text()).toContain('Mark all as read');
    expect(wrapper.text()).toMatch(/[45] minutes ago/);
    expect(wrapper.text()).not.toContain('分鐘前');
  });
});
