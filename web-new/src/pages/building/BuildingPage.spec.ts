/*
 * 我的大廈頁回歸測試。
 * 1. 驗證進入頁面時重新讀取目前會員資料。
 * 2. 驗證新綁定大廈取代舊 session 大廈並顯示正式名稱。
 */
import { flushPromises, mount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import i18n, { applyLocale } from '@/i18n';

import BuildingPage from './BuildingPage.vue';

const mocks = vi.hoisted(() => ({
  loadCurrentUser: vi.fn(),
  fetchBuildingInfo: vi.fn(),
  fetchBuildingNotices: vi.fn(),
  fetchMemberBuildings: vi.fn(),
  session: {
    me: {
      primary_community: {
        public_id: '0419900',
        name_zh: '舊大廈',
        name_en: 'Old Building',
        address_text: '舊大廈',
      },
      bound_building_ids: ['0419900'],
      residence_floor: '01',
      residence_unit: 'A',
      ismart_msg: { client_building_permissions: ['0419900', '0999900'] },
    },
  },
}));

vi.mock('@/stores/session', () => ({
  useSessionStore: () => ({
    get me() {
      return mocks.session.me;
    },
    loadCurrentUser: mocks.loadCurrentUser,
  }),
}));

vi.mock('vue-router', () => ({
  RouterLink: {
    props: ['to'],
    template: '<a><slot /></a>',
  },
}));

vi.mock('@/httpapis/building', () => ({
  fetchMemberIsmartBuildingInfo: mocks.fetchBuildingInfo,
  fetchMemberIsmartBuildingNotices: mocks.fetchBuildingNotices,
  fetchMemberPosBuildings: mocks.fetchMemberBuildings,
  fetchPosBuildings: vi.fn(),
  fetchMemberICCTVPublicCameras: vi.fn(),
  fetchMemberIsmartBuildingAccess: vi.fn(),
  fetchMemberIsmartManagementFees: vi.fn(),
  fetchMemberIsmartOtherFees: vi.fn(),
  generateMemberIsmartDoorQRCode: vi.fn(),
  openMemberIsmartDoor: vi.fn(),
}));

vi.mock('@/httpapis/payments', () => ({
  fetchPOSIntegrationTransactionsByDate: vi.fn(),
  fetchPOSIntegrationTransactionsByUnit: vi.fn(),
  fetchPOSIntegrationUnpaidInvoices: vi.fn(),
}));

describe('BuildingPage binding refresh', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    applyLocale('zh-HK');
    mocks.loadCurrentUser.mockReset();
    mocks.fetchBuildingInfo.mockReset();
    mocks.fetchBuildingNotices.mockReset();
    mocks.fetchMemberBuildings.mockReset();
    mocks.session.me = {
      primary_community: {
        public_id: '0419900',
        name_zh: '舊大廈',
        name_en: 'Old Building',
        address_text: '舊大廈',
      },
      bound_building_ids: ['0419900'],
      residence_floor: '01',
      residence_unit: 'A',
      ismart_msg: { client_building_permissions: ['0419900', '0999900'] },
    };
    mocks.loadCurrentUser.mockImplementation(async () => {
      mocks.session.me = {
        primary_community: {
          public_id: '0999900',
          name_zh: '測試1大廈',
          name_en: 'Test Building 1',
          address_text: '測試1大廈',
        },
        bound_building_ids: ['0999900'],
        residence_floor: 'G',
        residence_unit: 'B',
        ismart_msg: { client_building_permissions: ['0419900', '0999900'] },
      };
      return mocks.session.me;
    });
    mocks.fetchMemberBuildings.mockResolvedValue([
      { building_id: '0999900', buildname_chi: '測試1大廈', buildname: 'Test Building 1' },
    ]);
    mocks.fetchBuildingInfo.mockResolvedValue({
      selected_building_id: '0999900',
      building_options: ['0419900', '0999900'],
      building: { building_id: '0999900', buildname_chi: '測試1大廈' },
      building_info: {},
      documents: {},
    });
    mocks.fetchBuildingNotices.mockResolvedValue({
      selected_building_id: '0999900',
      building_options: ['0419900', '0999900'],
      result: [],
    });
  });

  it('loads the newly bound building instead of the stale session building', async () => {
    const wrapper = mount(BuildingPage, {
      global: { plugins: [i18n] },
    });
    await flushPromises();

    expect(mocks.loadCurrentUser).toHaveBeenCalledTimes(1);
    expect(mocks.fetchBuildingInfo).toHaveBeenCalledWith('0999900');
    expect(mocks.fetchBuildingNotices).toHaveBeenCalledWith('0999900');
    expect(wrapper.text()).toContain('測試1大廈');
  });
});
