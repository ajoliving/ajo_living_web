/*
 * 樓盤詳情頁測試。
 * 1. 驗證預約睇樓缺少必填資料時顯示欄位級錯誤。
 * 2. 驗證校驗失敗時不會建立預約請求。
 */
import { flushPromises, mount } from '@vue/test-utils';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { createPropertyAppointment, fetchPropertySaleDetail, fetchSimilarPropertySales } from '@/httpapis/properties';

import PropertyDetailPage from './PropertyDetailPage.vue';

const testMocks = vi.hoisted(() => ({
  routerPush: vi.fn(),
}));

vi.mock('vue-router', () => ({
  useRoute: () => ({ fullPath: '/properties/listing-1?action=appointment', params: { listingId: 'listing-1' }, query: { action: 'appointment' } }),
  useRouter: () => ({ push: testMocks.routerPush }),
}));

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string): string => ({
      'property.publicDetail.appointmentRequiredError': '請填寫預約睇樓所需資料。',
      'property.publicDetail.contactNameRequired': '請填寫聯絡人姓名。',
      'property.publicDetail.phoneRequired': '請填寫聯絡電話。',
    }[key] ?? key),
  }),
}));

vi.mock('@/httpapis/auth-session', () => ({
  readStoredAccessToken: () => 'test-access-token',
}));

vi.mock('@/httpapis/chats', () => ({
  createOrReusePropertyChat: vi.fn(),
}));

vi.mock('@/httpapis/properties', () => ({
  createPropertyAppointment: vi.fn(),
  favoritePropertySale: vi.fn(),
  fetchPropertySaleContactAccess: vi.fn(),
  fetchPropertySaleDetail: vi.fn(),
  fetchSimilarPropertySales: vi.fn(),
  reportPropertySale: vi.fn(),
  unfavoritePropertySale: vi.fn(),
}));

vi.mock('@/stores/preferences', () => ({
  usePreferenceStore: () => ({ locale: 'zh-HK' }),
}));

const mockedCreateAppointment = vi.mocked(createPropertyAppointment);
const mockedFetchDetail = vi.mocked(fetchPropertySaleDetail);
const mockedFetchSimilar = vi.mocked(fetchSimilarPropertySales);

describe('PropertyDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedFetchDetail.mockResolvedValue({
      data: {
        data: {
          listing_id: 'listing-1',
          module: 'property_sale',
          title: '測試樓盤',
          summary: '',
          description: '',
          district_code: '',
          publisher_identity_type: 'owner',
          publication_status: 'active',
          business_status: 'available',
          updated_at: '',
          images: [],
          contact_summary: {
            show_chat: false,
            show_phone: false,
            show_whatsapp: false,
          },
        },
      },
    } as never);
    mockedFetchSimilar.mockResolvedValue({ data: { data: { items: [] } } } as never);
  });

  it('顯示預約必填欄位錯誤並阻止提交', async () => {
    const wrapper = mount(PropertyDetailPage, {
      global: {
        stubs: {
          RouterLink: true,
        },
      },
    });
    await flushPromises();

    await wrapper.find('.detail-form').trigger('submit');

    expect(wrapper.text()).toContain('請填寫預約睇樓所需資料。');
    expect(wrapper.text()).toContain('請填寫聯絡人姓名。');
    expect(wrapper.text()).toContain('請填寫聯絡電話。');
    expect(wrapper.findAll('.detail-required-mark')).toHaveLength(2);
    expect(mockedCreateAppointment).not.toHaveBeenCalled();
  });
});
