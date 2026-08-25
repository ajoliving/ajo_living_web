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
  fetchBuildingAccess: vi.fn(),
  openDoor: vi.fn(),
  fetchICCTV: vi.fn(),
  fetchServiceCase: vi.fn(),
  fetchServiceCases: vi.fn(),
  submitServiceCase: vi.fn(),
  fetchUnpaidInvoices: vi.fn(),
  fetchTransactionsByUnit: vi.fn(),
  fetchTransactionsByDate: vi.fn(),
  updateMe: vi.fn(),
  routerPush: vi.fn(),
  routerReplace: vi.fn(),
  route: {
    path: '/building',
    query: {} as Record<string, string>,
  },
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
  useRoute: () => mocks.route,
  useRouter: () => ({
    push: mocks.routerPush,
    replace: mocks.routerReplace,
  }),
}));

vi.mock('@/httpapis/building', () => ({
  fetchMemberIsmartBuildingInfo: mocks.fetchBuildingInfo,
  fetchMemberIsmartBuildingNotices: mocks.fetchBuildingNotices,
  fetchMemberPosBuildings: mocks.fetchMemberBuildings,
  fetchMemberPosBuildingUnits: mocks.fetchMemberUnits,
  fetchPosBuildings: mocks.fetchPublicBuildings,
  fetchMemberICCTVPublicCameras: mocks.fetchICCTV,
  fetchMemberIsmartBuildingAccess: mocks.fetchBuildingAccess,
  fetchMemberIsmartManagementFees: vi.fn(),
  fetchMemberIsmartOtherFees: vi.fn(),
  fetchMemberIsmartServiceCase: mocks.fetchServiceCase,
  fetchMemberIsmartServiceCases: mocks.fetchServiceCases,
  generateMemberIsmartDoorQRCode: vi.fn(),
  openMemberIsmartDoor: mocks.openDoor,
  submitMemberIsmartBuildingComment: mocks.submitServiceCase,
}));

vi.mock('@/httpapis/me', () => ({
  updateMe: mocks.updateMe,
}));

vi.mock('@/httpapis/payments', () => ({
  fetchPOSIntegrationTransactionsByDate: mocks.fetchTransactionsByDate,
  fetchPOSIntegrationTransactionsByUnit: mocks.fetchTransactionsByUnit,
  fetchPOSIntegrationUnpaidInvoices: mocks.fetchUnpaidInvoices,
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
    mocks.fetchBuildingAccess.mockReset();
    mocks.openDoor.mockReset();
    mocks.fetchICCTV.mockReset();
    mocks.fetchServiceCase.mockReset();
    mocks.fetchServiceCases.mockReset();
    mocks.submitServiceCase.mockReset();
    mocks.fetchUnpaidInvoices.mockReset();
    mocks.fetchTransactionsByUnit.mockReset();
    mocks.fetchTransactionsByDate.mockReset();
    mocks.updateMe.mockReset();
    mocks.routerReplace.mockReset();
    mocks.route.path = '/building';
    mocks.route.query = {};
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
    mocks.fetchBuildingAccess.mockResolvedValue({
      selected_building_id: '0999900',
      building_options: ['0999900'],
      doors: [],
      recent_records: [],
    });
    mocks.openDoor.mockResolvedValue({ is_success: true });
    mocks.submitServiceCase.mockResolvedValue({ case_id: 'case-new', status: 'submitted' });
    mocks.fetchServiceCases.mockResolvedValue({
      selected_building_id: '0999900',
      status_choices: [],
      cases: [],
    });
    mocks.fetchUnpaidInvoices.mockResolvedValue([{
      invoice_no: 'INV-001',
      flat_code: '09999000012',
      item_id: 'MANAGEMENT_FEE',
      net_amount: 1200,
    }]);
    const paymentRecords = {
      payment_objs: [{
        input_time: '2026-02-11 02:52:34',
        tran_time: '2026-02-11 02:52:23',
        payment_id: 'payment-001',
        receipt_id: '20132560',
        pay_type: 'POS_CHEQUE',
        status: 'void',
        payment_detail_objs: [
          { floor: 'G', unit: 'B', item_id: '管理費', term: '2024/09', trs_val: 5 },
          { floor: 'G', unit: 'B', item_id: '清潔費', term: '2024/09', trs_val: 3 },
          { floor: 'G', unit: 'I', item_id: '管理費', term: '2024/09', trs_val: 5 },
        ],
      }],
    };
    mocks.fetchTransactionsByUnit.mockResolvedValue(paymentRecords);
    mocks.fetchTransactionsByDate.mockResolvedValue(paymentRecords);
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

  it('shows the simplified notice and building profile content', async () => {
    mocks.fetchBuildingInfo.mockResolvedValue({
      selected_building_id: '0999900',
      building_options: ['0999900'],
      building: { building_id: '0999900', buildname_chi: '測試1大廈' },
      building_info: { year_built: 1977, total_floor: 2 },
      documents: {
        financial_reports: [{
          id: 1,
          title: '六月財務報表',
          file_date: '2026-06-01',
          created_date: '2026-06-14',
          file_month: '2026-06-01',
        }],
      },
    });
    const wrapper = mount(BuildingPage, {
      global: { plugins: [i18n], stubs: { RouterLink: true } },
    });
    await flushPromises();

    expect(wrapper.get('.notice-current span').text()).toBe('目前顯示：測試1大廈');
    expect(wrapper.findAll('.work-kicker')).toHaveLength(0);
    expect(wrapper.find('.building-summary-grid').exists()).toBe(false);

    const fields = wrapper.findAll('[data-work-panel="affairs-building"] .building-field');
    expect(fields[0]?.text()).toContain('落成年份1977');
    expect(fields[1]?.text()).toContain(`樓齡${new Date().getFullYear() - 1977} 年`);
    expect(fields.some((field) => field.text().includes('資料來源'))).toBe(false);
    expect(wrapper.find('[data-work-panel="affairs-finance"] .building-field-grid').exists()).toBe(false);

    const formsPanel = wrapper.get('[data-work-panel="affairs-forms"]');
    expect(formsPanel.find('.work-desc').exists()).toBe(false);
    expect(formsPanel.find('.notice-admin-grid').exists()).toBe(false);

    const financialPanel = wrapper.findAll('[data-work-panel="affairs-finance"] .acct-subpanel')[1];
    const uploadDate = new Intl.DateTimeFormat('zh-HK', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    }).format(new Date(2026, 5, 14));
    expect(financialPanel?.findAll('th').map((header) => header.text())).toContain('上載日期');
    expect(financialPanel?.findAll('.building-file-meta-cell')[0]?.text()).toBe(uploadDate);
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

  it('shows formal owner-account copy and resolves unit codes through the POS directory', async () => {
    await mocks.loadCurrentUser();
    const wrapper = mount(BuildingPage, {
      global: { plugins: [i18n], stubs: { RouterLink: true } },
    });
    await flushPromises();

    const ownerAccountButton = wrapper.findAll('.work-nav-item')
      .find((button) => button.text().includes('業戶帳目'));
    await ownerAccountButton?.trigger('click');
    await flushPromises();

    const panel = wrapper.get('[data-work-panel="affairs-owner-account"]');
    expect(panel.get('.work-desc').text()).toBe('查看未繳賬單及繳費記錄。');
    expect(panel.get('.acct-note').text()).toMatch(/^目前顯示單位：/);
    expect(panel.findAll('.owner-account-table th')[0]?.text()).toBe('號單編碼');
    expect(panel.findAll('.owner-account-table tbody td')[1]?.text()).toBe('G/B');
    expect(panel.text()).not.toContain('09999000012');

    const recordHeaders = panel.findAll('.owner-record-table th').map((header) => header.text());
    const recordRows = panel.findAll('.owner-record-table tbody tr');
    expect(recordHeaders).toContain('收據編號');
    expect(recordHeaders).not.toContain('輸入時間');
    expect(recordHeaders).toContain('繳付金額');
    expect(recordRows).toHaveLength(2);
    expect(recordRows[0]?.text()).toContain('G/B');
    expect(recordRows[0]?.text()).not.toContain('G/I');
    expect(panel.findAll('.acct-toolbar .acct-total')).toHaveLength(1);
    expect(mocks.fetchTransactionsByUnit).toHaveBeenCalledWith(['09999000012']);

    const itemFilter = panel.get('#owner-record-item-filter');
    expect(itemFilter.findAll('option').map((option) => option.text())).toEqual([
      '全部項目',
      '管理費',
      '清潔費',
    ]);
    await itemFilter.setValue('管理費');
    expect(panel.findAll('.owner-record-table tbody tr')).toHaveLength(1);
    expect(panel.get('.owner-record-table tbody tr').text()).toContain('管理費');
    expect(panel.get('.owner-record-table tbody tr').text()).not.toContain('清潔費');

    await panel.get('.acct-filter-row').trigger('submit');
    await flushPromises();
    expect(mocks.fetchTransactionsByDate).toHaveBeenCalledWith(
      expect.objectContaining({ date_type: 'tran_date' }),
      ['09999000012'],
    );
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
    const feedbackPanel = wrapper.get('[data-work-panel="affairs-feedback"]');
    expect(feedbackPanel.get('.work-desc').text()).toBe('一鍵提交意見反映和維修報修事項，直達管理處跟進');
    expect(feedbackPanel.text()).toContain('提交記錄');
    expect(feedbackPanel.get('.affairs-feedback-form .work-card-title').text()).toBe('已綁定大廈');
    expect(feedbackPanel.get('.affairs-feedback-form .work-card-sub').text()).toBe('測試1大廈');
    expect(feedbackPanel.find('.affairs-feedback-form select').exists()).toBe(false);
    expect(wrapper.text()).toContain('18樓走廊漏水');
    await wrapper.get('[data-work-panel="affairs-feedback"] .work-mini-btn').trigger('click');
    await flushPromises();
    expect(mocks.fetchServiceCase).toHaveBeenCalledWith('case-1', '0999900');
    expect(wrapper.text()).toContain('需要管理處跟進');
  });

  it('submits only iSmart-supported legacy classifications', async () => {
    const wrapper = mount(BuildingPage, {
      global: { plugins: [i18n], stubs: { RouterLink: true } },
    });
    await flushPromises();

    const feedbackButton = wrapper.findAll('.work-nav-item')
      .find((button) => button.text().includes('意見提供/維修報修'));
    await feedbackButton?.trigger('click');
    await flushPromises();

    const repairCategory = wrapper.get('#affairs-repair-category');
    expect((repairCategory.element as HTMLSelectElement).value).toBe('門卡報失');
    expect(wrapper.find('#affairs-repair-subcategory').exists()).toBe(false);
    await repairCategory.setValue('電力問題');

    const actionButton = () => wrapper.findAll('.affairs-guided-actions .work-action')[1];
    await actionButton()?.trigger('click');
    await wrapper.get('#affairs-repair-content').setValue('12樓走廊照明故障');
    await wrapper.get('#affairs-location').setValue('12樓走廊');
    await actionButton()?.trigger('click');
    await actionButton()?.trigger('click');
    await flushPromises();

    expect(mocks.submitServiceCase).toHaveBeenCalledWith({
      building_id: '0999900',
      unit_id: '09999000012',
      comment_type: '電力問題',
      comment: '12樓走廊照明故障',
      subject: '電力問題',
      content: '12樓走廊照明故障',
      location_text: '12樓走廊',
    });
    const successDialog = document.body.querySelector('.affairs-success-dialog');
    expect(successDialog?.textContent).toContain('提交成功');
    expect(successDialog?.textContent).toContain('管理處將按提交內容跟進');
    (successDialog?.querySelector('.affairs-success-dialog__action') as HTMLButtonElement).click();
    await flushPromises();
    expect(document.body.querySelector('.affairs-success-dialog')).toBeNull();

    const feedbackEntry = wrapper.findAll('.affairs-entry-card')
      .find((button) => button.text().includes('意見反映'));
    await feedbackEntry?.trigger('click');
    await flushPromises();

    expect((wrapper.get('#affairs-feedback-category').element as HTMLSelectElement).value).toBe('嘈音滋擾');
    expect(wrapper.find('#affairs-feedback-subcategory').exists()).toBe(false);
  });

  it('blocks service-case submission without a current building binding', async () => {
    const unboundMember = {
      ...mocks.session.me,
      bound_building_ids: [],
      bound_flat_unit_ids: [],
      residence_binding_status: '',
    };
    mocks.session.me = unboundMember;
    mocks.loadCurrentUser.mockResolvedValue(unboundMember);

    const wrapper = mount(BuildingPage, {
      global: { plugins: [i18n], stubs: { RouterLink: true } },
    });
    await flushPromises();

    const feedbackButton = wrapper.findAll('.work-nav-item')
      .find((button) => button.text().includes('意見提供/維修報修'));
    await feedbackButton?.trigger('click');
    await flushPromises();

    const feedbackPanel = wrapper.get('[data-work-panel="affairs-feedback"]');
    expect(feedbackPanel.text()).toContain('目前沒有已綁定大廈，無法提交維修或意見個案');
    expect(feedbackPanel.find('.affairs-guided-shell').exists()).toBe(false);
    expect(feedbackPanel.findAll('.affairs-entry-card').every((button) => button.attributes('disabled') !== undefined)).toBe(true);
    expect(mocks.fetchServiceCases).not.toHaveBeenCalled();
    expect(mocks.submitServiceCase).not.toHaveBeenCalled();
  });

  it('lets the open-door API make the final permission decision', async () => {
    mocks.fetchBuildingAccess.mockResolvedValue({
      selected_building_id: '0999900',
      building_options: ['0999900'],
      doors: [{
        door_id: 8,
        title: '地下大堂門',
        door_no: 2,
        building_id: '0999900',
        has_permission: false,
        is_public: true,
        camera: null,
      }],
      recent_records: [],
    });
    const wrapper = mount(BuildingPage, {
      global: { plugins: [i18n], stubs: { RouterLink: true } },
    });
    await flushPromises();

    const accessButton = wrapper.findAll('.work-nav-item')
      .find((button) => button.text().includes('智能門禁'));
    await accessButton?.trigger('click');
    await flushPromises();

    const accessPanel = wrapper.get('[data-work-panel="affairs-access"]');
    expect(accessPanel.get('.work-desc').text()).toBe('遙距開門、查看公用密碼、生成私人密碼、專屬二維碼及查看開門記錄。');
    expect(accessPanel.find('.access-summary-grid').exists()).toBe(false);
    expect(accessPanel.get('.access-door-table thead').text()).not.toContain('門號');
    expect(accessPanel.get('.access-door-table thead').text()).not.toContain('所屬大廈');
    expect(accessPanel.get('.access-door-table thead').text()).not.toContain('遙距開門權限');
    expect(accessPanel.get('.access-door-table tbody').text()).toContain('未升級');
    const openButton = accessPanel.get('.access-table-actions .work-mini-btn.primary');
    expect(openButton.attributes('disabled')).toBeUndefined();

    const confirm = vi.spyOn(window, 'confirm').mockReturnValue(true);
    await openButton.trigger('click');
    await flushPromises();
    expect(mocks.openDoor).toHaveBeenCalledWith({ building_id: '0999900', door_id: 8 });
    confirm.mockRestore();
  });

  it('shows building cameras without exposing iCCTV device names', async () => {
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
    const icctvPanel = wrapper.get('[data-work-panel="affairs-icctv"]');
    expect(icctvPanel.get('.work-desc').text()).toBe('隨時隨地查看已支援的大廈鏡頭實時影像。');
    expect(icctvPanel.get('.icctv-profile-head h3').text()).toBe('測試1大廈');
    expect(icctvPanel.get('.icctv-profile-head p').text()).toBe('目前顯示的大廈鏡頭');
    expect(icctvPanel.text()).not.toContain('ICCTV');
    expect(icctvPanel.text()).not.toContain('Orange Pi');
    expect(icctvPanel.text()).not.toContain('青島香橙派');
    expect(icctvPanel.text()).not.toContain('192.168.72.174');
    expect(icctvPanel.text()).toContain('channel1');
    expect(icctvPanel.text()).toContain('香工後門');
    expect(icctvPanel.text()).toContain('新窗口顯示');
  });
});
