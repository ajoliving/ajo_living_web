/*
 * AJO Pay 付款頁測試。
 * 1. 驗證麵包屑使用全站統一元件資料。
 * 2. 驗證全部賬單按鈕可在選擇與取消之間切換。
 * 3. 驗證確認區不再顯示優惠券折扣行，第四步顯示訂單。
 */
import { flushPromises, mount } from '@vue/test-utils';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import {
  fetchPOSPaymentBankAccounts,
  fetchPOSPaymentBills,
  fetchPOSPaymentFees,
  fetchPOSPaymentOrders,
} from '@/httpapis/payments';

import Page from './Page.vue';

vi.mock('vue-router', () => ({
  useRoute: () => ({
    query: {},
  }),
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
  createPOSPaymentOrder: vi.fn(),
  fetchPOSPaymentBankAccounts: vi.fn(),
  fetchPOSPaymentBills: vi.fn(),
  fetchPOSPaymentFees: vi.fn(),
  fetchPOSPaymentOrder: vi.fn(),
  fetchPOSPaymentOrders: vi.fn(),
  queryPOSPaymentOrder: vi.fn(),
  reportPOSPayment: vi.fn(),
}));

const mockedFetchBills = vi.mocked(fetchPOSPaymentBills);
const mockedFetchFees = vi.mocked(fetchPOSPaymentFees);
const mockedFetchBankAccounts = vi.mocked(fetchPOSPaymentBankAccounts);
const mockedFetchOrders = vi.mocked(fetchPOSPaymentOrders);

// 1. 建立固定賬單回應
const createBillsResponse = () => ({
  data: {
    data: {
      context: {
        building_id: 'BLG-001',
        building_name: 'AJO Tower',
        unit_id: 'BLG-0010000101',
        floor: '1',
        unit: '01',
        unit_label: '1 / 01',
      },
      items: [
        {
          invoice_no: 'INV-001',
          item_name: '管理費',
          bill_dt: '2026-06-01',
          net_amount: 1200,
        },
        {
          invoice_no: 'INV-002',
          item_name: '維修費',
          bill_dt: '2026-06-02',
          net_amount: 300,
        },
      ],
      building_options: ['BLG-001'],
      unit_options: ['BLG-0010000101'],
    },
  },
}) as unknown as Awaited<ReturnType<typeof fetchPOSPaymentBills>>;

// 2. 建立付款設定回應
const createListResponse = (items: Record<string, unknown>[] = []) => ({
  data: {
    data: {
      items,
      building_options: ['BLG-001'],
      unit_options: ['BLG-0010000101'],
    },
  },
}) as unknown as Awaited<ReturnType<typeof fetchPOSPaymentFees>>;

describe('AjoPayCheckoutPage', () => {
  beforeEach(() => {
    mockedFetchBills.mockReset();
    mockedFetchFees.mockReset();
    mockedFetchBankAccounts.mockReset();
    mockedFetchOrders.mockReset();
    mockedFetchBills.mockResolvedValue(createBillsResponse());
    mockedFetchFees.mockResolvedValue(createListResponse([
      { pay_type: 'WEBPOS_WECHAT', markup: '2.7%' },
    ]));
    mockedFetchBankAccounts.mockResolvedValue(createListResponse([]));
    mockedFetchOrders.mockResolvedValue(createListResponse([]));
  });

  it('aligns checkout breadcrumb, bill toggle, and order step copy', async () => {
    const wrapper = mount(Page, {
      global: {
        stubs: {
          AppBreadcrumb: {
            props: ['items'],
            template: '<nav class="breadcrumb"><span v-for="item in items" :key="item.label">{{ item.label }}</span></nav>',
          },
          QrCodeImage: true,
        },
      },
    });
    await flushPromises();

    expect(wrapper.find('.breadcrumb').text()).toContain('首頁');
    expect(wrapper.find('.breadcrumb').text()).toContain('AJO Pay');
    expect(wrapper.find('.breadcrumb').text()).toContain('付款');
    expect(wrapper.find('.ajo-pay-checkout__summary-box').text()).not.toContain('優惠券');
    expect(wrapper.text()).toContain('訂單');

    const selectAllButton = wrapper.findAll('button').find((button) => button.text() === '選擇全部賬單');
    expect(selectAllButton).toBeTruthy();
    await selectAllButton?.trigger('click');
    expect(wrapper.find('.ajo-pay-checkout__total').text()).toContain('HK$1,500.00');

    const cancelAllButton = wrapper.findAll('button').find((button) => button.text() === '取消全部賬單');
    expect(cancelAllButton).toBeTruthy();
    await cancelAllButton?.trigger('click');
    expect(wrapper.find('.ajo-pay-checkout__total').text()).toContain('HK$0.00');
  });
});
