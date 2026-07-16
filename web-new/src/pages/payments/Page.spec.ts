/*
 * AJO Pay 首頁語系測試。
 * 1. 驗證 English 切換會更新頁面操作與摘要文案。
 * 2. 驗證港幣與數量格式使用目前偏好語系。
 */
import { flushPromises, mount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { nextTick } from 'vue';

import { fetchPOSPaymentBills } from '@/httpapis/payments';
import i18n, { applyLocale } from '@/i18n';
import { usePreferenceStore } from '@/stores/preferences';

import Page from './Page.vue';

vi.mock('vue-router', () => ({
  RouterLink: {
    props: ['to'],
    template: '<a><slot /></a>',
  },
  useRouter: () => ({
    push: vi.fn(),
  }),
}));

vi.mock('@/stores/feedback', () => ({
  useFeedbackStore: () => ({
    pushToast: vi.fn(),
  }),
}));

vi.mock('@/httpapis/payments', () => ({
  fetchPOSPaymentBills: vi.fn(),
}));

const mockedFetchBills = vi.mocked(fetchPOSPaymentBills);

// 1. 建立固定待繳賬單回應
const createBillsResponse = () => ({
  data: {
    data: {
      context: {
        building_id: 'BLG-001',
        building_name: 'AJO Tower',
        unit_id: 'BLG-0010000101',
        unit_label: '1 / 01',
      },
      items: [
        {
          invoice_no: 'INV-001',
          item_name: '管理費',
          net_amount: 1200,
        },
      ],
    },
  },
}) as unknown as Awaited<ReturnType<typeof fetchPOSPaymentBills>>;

describe('AjoPayHomePage locale', () => {
  beforeEach(() => {
    const pinia = createPinia();
    setActivePinia(pinia);
    window.localStorage.clear();
    applyLocale('zh-HK');
    mockedFetchBills.mockReset();
    mockedFetchBills.mockResolvedValue(createBillsResponse());
  });

  afterEach(() => {
    applyLocale('zh-HK');
  });

  it('updates payment copy and formatting after switching to English', async () => {
    const pinia = createPinia();
    setActivePinia(pinia);
    const wrapper = mount(Page, {
      global: {
        plugins: [pinia, i18n],
      },
    });
    await flushPromises();

    const preferenceStore = usePreferenceStore();
    preferenceStore.setLocale('en');
    applyLocale('en');
    await nextTick();

    expect(wrapper.text()).toContain('Hong Kong smart property payment platform');
    expect(wrapper.text()).toContain('Building management fees');
    expect(wrapper.text()).toContain('Payment items due');
    expect(wrapper.text()).toContain('HK$1,200');
    expect(wrapper.text()).not.toContain('目前待繳項目');
  });
});
