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
  fetchPublicBuildings: vi.fn(),
  fetchMemberUnits: vi.fn(),
  fetchICCTV: vi.fn(),
  fetchServiceCase: vi.fn(),
  fetchServiceCases: vi.fn(),
  submitServiceCase: vi.fn(),
  updateMe: vi.fn(),
  session: {
    me: {
      display_name: 'Member',
      primary_community: {
        public_id: '0419900',
        name_zh: '舊大廈',
        name_en: 'Old Building',
        address_text: '舊大廈',
      },
      bound_building_ids: ['0419900'],
      bound_flat_unit_ids: ['04199000112'],
      residence_floor: '01',
      residence_unit: 'B',
      ismart_msg: {
        client_building_permissions: ['0419900', '0999900'],
        client_building_flat_units_permissions: ['04199000112', '09999000012'],
      },
    },
  },
}));

vi.mock('@/stores/session', () => ({
  useSessionStore: () => ({
    get me() {
      return mocks.session.me;
    },
    set me(value) {
      mocks.session.me = value;
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
  fetchMemberPosBuildingUnits: mocks.fetchMemberUnits,
  fetchPosBuildings: mocks.fetchPublicBuildings,
  fetchMemberICCTVPublicCameras: mocks.fetchICCTV,
  fetchMemberIsmartBuildingAccess: vi.fn(),
  fetchMemberIsmartManagementFees: vi.fn(),
  fetchMemberIsmartOtherFees: vi.fn(),
  fetchMemberIsmartServiceCase: mocks.fetchServiceCase,
  fetchMemberIsmartServiceCases: mocks.fetchServiceCases,
  generateMemberIsmartDoorQRCode: vi.fn(),
  openMemberIsmartDoor: vi.fn(),
  submitMemberIsmartBuildingComment: mocks.submitServiceCase,
}));

vi.mock('@/httpapis/me', () => ({
  updateMe: mocks.updateMe,
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
    mocks.fetchPublicBuildings.mockReset();
    mocks.fetchMemberUnits.mockReset();
    mocks.fetchICCTV.mockReset();
    mocks.fetchServiceCase.mockReset();
    mocks.fetchServiceCases.mockReset();
    mocks.submitServiceCase.mockReset();
    mocks.updateMe.mockReset();
    mocks.session.me = {
      display_name: 'Member',
      primary_community: {
        public_id: '0419900',
        name_zh: '舊大廈',
        name_en: 'Old Building',
        address_text: '舊大廈',
      },
      bound_building_ids: ['0419900'],
      bound_flat_unit_ids: ['04199000112'],
      residence_floor: '01',
      residence_unit: 'B',
      ismart_msg: {
        client_building_permissions: ['0419900', '0999900'],
        client_building_flat_units_permissions: ['04199000112', '09999000012'],
      },
    };
    mocks.loadCurrentUser.mockImplementation(async () => {
      mocks.session.me = {
        display_name: 'Member',
        primary_community: {
          public_id: '0999900',
          name_zh: '測試1大廈',
          name_en: 'Test Building 1',
          address_text: '測試1大廈',
        },
        bound_building_ids: ['0999900'],
        bound_flat_unit_ids: ['09999000012'],
        residence_floor: 'G',
        residence_unit: 'B',
        ismart_msg: {
          client_building_permissions: ['0419900', '0999900'],
          client_building_flat_units_permissions: ['04199000112', '09999000012'],
        },
      };
      return mocks.session.me;
    });
    mocks.fetchMemberBuildings.mockResolvedValue([
      { building_id: '0999900', buildname_chi: '測試1大廈', buildname: 'Test Building 1' },
    ]);
    mocks.fetchPublicBuildings.mockResolvedValue([
      { building_id: '0419900', buildname_chi: '時安大廈', buildname: 'Chee On Building' },
      { building_id: '0999900', buildname_chi: '測試1大廈', buildname: 'Test Building 1' },
    ]);
    mocks.fetchMemberUnits.mockImplementation(async (buildingID: string) => (
      buildingID === '0419900'
        ? [{ unit_id: '04199000112', floor: '01', unit: 'B' }]
        : [{ unit_id: '09999000012', floor: 'G', unit: 'B' }]
    ));
    mocks.updateMe.mockImplementation(async (payload: Record<string, unknown>) => {
      const buildingID = String(payload.primary_community_id ?? '');
      const isCheeOn = buildingID === '0419900';
      mocks.session.me = {
        ...mocks.session.me,
        primary_community: {
          public_id: buildingID,
          name_zh: isCheeOn ? '時安大廈' : '測試1大廈',
          name_en: isCheeOn ? 'Chee On Building' : 'Test Building 1',
          address_text: isCheeOn ? '時安大廈' : '測試1大廈',
        },
        bound_building_ids: payload.bound_building_ids as string[],
        bound_flat_unit_ids: payload.bound_flat_unit_ids as string[],
        residence_floor: String(payload.residence_floor ?? ''),
        residence_unit: String(payload.residence_unit ?? ''),
      };
      return { data: { data: mocks.session.me } };
    });
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
    mocks.fetchICCTV.mockResolvedValue({
      selected_building_id: '0999900',
      building_options: ['0999900'],
      orangepis: [],
      cameras: [],
    });
    mocks.fetchServiceCases.mockResolvedValue({
      selected_building_id: '0999900',
      status_choices: [],
      cases: [],
    });
  });

  it('loads the newly bound building instead of the stale session building', async () => {
    const wrapper = mount(BuildingPage, {
      global: { plugins: [i18n], stubs: { RouterLink: true } },
    });
    await flushPromises();

    expect(mocks.loadCurrentUser).toHaveBeenCalledTimes(1);
    expect(mocks.fetchBuildingInfo).toHaveBeenCalledWith('0999900');
    expect(mocks.fetchBuildingNotices).toHaveBeenCalledWith('0999900');
    expect(wrapper.text()).toContain('測試1大廈');
  });

  it('fills every authorized property option with the official POS name and unit', async () => {
    const wrapper = mount(BuildingPage, {
      global: { plugins: [i18n], stubs: { RouterLink: true } },
    });
    await flushPromises();

    expect(mocks.fetchPublicBuildings).toHaveBeenCalledTimes(1);
    const selector = wrapper.get('.building-context-select');
    const labels = selector.findAll('option').map((option) => option.text());
    expect(selector.get('option[value=""]').text()).toBe('測試1大廈 / G / B');
    expect(labels).toContain('時安大廈 / 01 / B');
    expect(labels.some((label) => label === '0419900')).toBe(false);
    expect(wrapper.find('.building-context-sidebar-current').exists()).toBe(false);
    expect(wrapper.find('.building-context-layer').exists()).toBe(false);
    expect(wrapper.find('.building-context-save').exists()).toBe(false);
  });

  it('switches one complete authorized unit and reloads notices and building data together', async () => {
    const wrapper = mount(BuildingPage, {
      global: { plugins: [i18n], stubs: { RouterLink: true } },
    });
    await flushPromises();

    await wrapper.get('.building-context-select').setValue('04199000112');
    await flushPromises();

    expect(mocks.updateMe).toHaveBeenCalledWith(expect.objectContaining({
      primary_community_id: '0419900',
      primary_community_name: '時安大廈',
      bound_building_ids: ['0419900'],
      bound_flat_unit_ids: ['04199000112'],
      residence_floor: '01',
      residence_unit: 'B',
    }));
    expect(mocks.fetchBuildingInfo).toHaveBeenLastCalledWith('0419900');
    expect(mocks.fetchBuildingNotices).toHaveBeenLastCalledWith('0419900');
    expect(wrapper.get('.building-context-select').get('option[value=""]').text()).toBe('時安大廈 / 01 / B');
    expect(wrapper.text()).toContain('目前物業已更新');
  });

  it('loads and opens only the current member service cases', async () => {
    mocks.fetchServiceCases.mockResolvedValue({
      selected_building_id: '0999900',
      status_choices: [{ code: 'processing', label: '處理中' }],
      cases: [{
        case_id: 'case-1',
        building_id: '0999900',
        request_type: 'repair',
        request_type_label: '維修報修',
        category_label: '水務',
        subject: '18樓走廊漏水',
        status: 'processing',
        status_label: '處理中',
        updated_at: '2026-07-30T10:20:30+08:00',
      }],
    });
    mocks.fetchServiceCase.mockResolvedValue({
      case_id: 'case-1',
      content: '18樓走廊漏水，需要管理處跟進。',
      messages: [{ message_id: 'message-1', author_username: '住戶', body: '已提交。' }],
    });
    const wrapper = mount(BuildingPage, {
      global: { plugins: [i18n], stubs: { RouterLink: true } },
    });
    await flushPromises();

    const feedbackButton = wrapper.findAll('.work-nav-item')
      .find((button) => button.text().includes('意見提供/維修報修'));
    await feedbackButton?.trigger('click');
    await flushPromises();

    expect(mocks.fetchServiceCases).toHaveBeenCalledWith('0999900', undefined);
    expect(wrapper.text()).toContain('18樓走廊漏水');
    await wrapper.get('[data-work-panel="affairs-feedback"] .work-mini-btn').trigger('click');
    await flushPromises();
    expect(mocks.fetchServiceCase).toHaveBeenCalledWith('case-1', '0999900');
    expect(wrapper.text()).toContain('需要管理處跟進');
  });

  it('defaults both repair classifications to the first available option', async () => {
    const wrapper = mount(BuildingPage, {
      global: { plugins: [i18n], stubs: { RouterLink: true } },
    });
    await flushPromises();

    const feedbackButton = wrapper.findAll('.work-nav-item')
      .find((button) => button.text().includes('意見提供/維修報修'));
    await feedbackButton?.trigger('click');
    await flushPromises();

    const repairCategory = wrapper.get('#affairs-repair-category');
    const repairSubcategory = wrapper.get('#affairs-repair-subcategory');
    expect((repairCategory.element as HTMLSelectElement).value).toBe('electrical');
    expect((repairSubcategory.element as HTMLSelectElement).value).toBe('corridorLighting');

    await repairCategory.setValue('water');
    expect((repairSubcategory.element as HTMLSelectElement).value).toBe('freshWater');

    const feedbackEntry = wrapper.findAll('.affairs-entry-card')
      .find((button) => button.text().includes('意見反映'));
    await feedbackEntry?.trigger('click');
    await flushPromises();

    expect((wrapper.get('#affairs-feedback-category').element as HTMLSelectElement).value).toBe('environment');
    expect((wrapper.get('#affairs-feedback-subcategory').element as HTMLSelectElement).value).toBe('cleaningSuggestion');
  });

  it('groups cameras by Orange Pi when devices share a channel name', async () => {
    mocks.fetchICCTV.mockResolvedValue({
      selected_building_id: '0999900',
      building_options: ['0999900'],
      orangepis: [
        { orangepi_id: 3, orangepi_name: '青島香橙派', is_active: true, camera_count: 1 },
        { orangepi_id: 7, orangepi_name: '192.168.72.174', is_active: true, camera_count: 1 },
      ],
      cameras: [
        {
          id: '3-channel1-1',
          title: 'channel1',
          channel: 'channel1',
          url: 'https://icctv.example/opi/29003/channel1',
          orangepi_id: 3,
          orangepi_name: '青島香橙派',
          is_active: true,
        },
        {
          id: '7-channel1-1',
          title: '香工後門',
          channel: 'channel1',
          url: 'https://icctv.example/opi/29005/channel1',
          orangepi_id: 7,
          orangepi_name: '192.168.72.174',
          is_active: true,
        },
      ],
    });
    const wrapper = mount(BuildingPage, {
      global: { plugins: [i18n], stubs: { RouterLink: true } },
    });
    await flushPromises();

    const cctvTab = wrapper.findAll('.work-nav-item').find((item) => item.text().includes('視像監控'));
    expect(cctvTab).toBeDefined();
    await cctvTab!.trigger('click');
    await flushPromises();

    expect(mocks.fetchICCTV).toHaveBeenCalledWith('0999900');
    expect(wrapper.text()).toContain('Orange Pi 裝置 #3 · 青島香橙派');
    expect(wrapper.text()).toContain('Orange Pi 裝置 #7 · 192.168.72.174');
    expect(wrapper.text()).toContain('香工後門');
  });
});
