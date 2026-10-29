/*
 * 會員中心資料載入回歸測試。
 * 1. 驗證帳號管理初始頁不預載物業綁定資料。
 * 2. 驗證首次切換面板才載入對應資料。
 * 3. 驗證相同大廈的會員單位請求在頁面生命週期內只執行一次。
 */
import { flushPromises, mount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import i18n, { applyLocale } from '@/i18n';

import AccountMyPage from './AccountMyPage.vue';

const mocks = vi.hoisted(() => ({
  fetchMemberBuildings: vi.fn(),
  fetchMemberUnits: vi.fn(),
  fetchPublicBuildings: vi.fn(),
  fetchPublicUnits: vi.fn(),
  fetchIsmartFlatBindings: vi.fn(),
  fetchIsmartOwnerBindingRequests: vi.fn(),
  submitOwnerBindingRequest: vi.fn(),
  loadCurrentUser: vi.fn(),
  pushToast: vi.fn(),
  changeMemberIsmartPassword: vi.fn(),
  updateMemberIsmartProfile: vi.fn(),
  updateMe: vi.fn(),
  routerPush: vi.fn(),
  route: {
    path: '/account/profile',
    query: {} as Record<string, string>,
  },
  session: {
    me: {
      display_name: 'Member',
      username: 'patrick',
      email: 'patrick@seventy2.hk',
      public_id: 'MEMBER-1',
      avatar_key: '',
      avatar_url: '',
      primary_community: {
        public_id: '0999900',
        name_zh: '測試1大廈',
        name_en: 'Test Building 1',
        address_text: '測試1大廈',
      },
      bound_building_ids: ['0999900'],
      bound_flat_unit_ids: ['09999000401', '09999000103'],
      residence_floor: '04',
      residence_unit: 'G',
      ismart_username: 'patrick',
      ismart_bound_phone: '+85261980774',
      ismart_msg: {
        username: 'patrick',
        email: 'patrick@seventy2.hk',
        phone: '+85261980774',
        client_building_flat_units_permissions: ['09999000401', '09999000103'],
      },
      ismart_account_profile: {
        account_code: 'patrick',
        account_phone: '+85261980774',
        account_email: 'patrick@seventy2.hk',
        owner_name_en: 'patrick',
        owner_name_zh: '林曉洸',
        account_name: 'patrick',
        identity_number: 'A123456(7)',
        legal_entity: 'SOLE',
        client_type: '員工',
        gender: 'M',
        birth_date: '1980-01-01',
        contact_name: 'patrick',
        contact_phone: '+85261980774',
        emergency_contact_name: '',
        emergency_contact_phone: '',
        billing_phone: '+85261980774',
        billing_email: 'patrick@seventy2.hk',
        billing_address: '',
        billing_address_en: '',
        billing_address_zh: '',
        properties: [],
      },
      roles: ['member'],
      permissions: [],
      role: 'member',
      account_type: 'personal',
      member_status: 'active',
    },
  },
}));

vi.mock('vue-router', () => ({
  RouterView: { template: '<div />' },
  useRoute: () => mocks.route,
  useRouter: () => ({ push: mocks.routerPush }),
}));

vi.mock('@/stores/session', () => ({
  useSessionStore: () => ({
    get me() {
      return mocks.session.me;
    },
    set me(value) {
      mocks.session.me = value;
    },
    get currentUser() {
      return {
        display_name: mocks.session.me.display_name,
        public_id: mocks.session.me.public_id,
        avatar_key: mocks.session.me.avatar_key ?? '',
        avatar_url: mocks.session.me.avatar_url ?? '',
        is_staff: false,
      };
    },
    loadCurrentUser: mocks.loadCurrentUser,
    signOut: vi.fn(),
  }),
}));

vi.mock('@/stores/preferences', () => ({
  usePreferenceStore: () => ({ locale: 'zh-HK' }),
}));

vi.mock('@/stores/feedback', () => ({
  useFeedbackStore: () => ({ pushToast: mocks.pushToast }),
}));

vi.mock('@/httpapis/building', () => ({
  fetchMemberPosBuildings: mocks.fetchMemberBuildings,
  fetchMemberPosBuildingUnits: mocks.fetchMemberUnits,
  fetchPosBuildings: mocks.fetchPublicBuildings,
  fetchPosBuildingUnits: mocks.fetchPublicUnits,
  fetchMemberIsmartFlatBindings: mocks.fetchIsmartFlatBindings,
  fetchMemberIsmartOwnerBindingRequests: mocks.fetchIsmartOwnerBindingRequests,
  submitMemberIsmartOwnerBindingRequest: mocks.submitOwnerBindingRequest,
}));

vi.mock('@/httpapis/me', () => ({
  changeMemberIsmartPassword: mocks.changeMemberIsmartPassword,
  updateMe: mocks.updateMe,
  updateMemberIsmartProfile: mocks.updateMemberIsmartProfile,
}));

const initialMemberSession = JSON.parse(JSON.stringify(mocks.session.me)) as typeof mocks.session.me;

// 0. 建立可覆蓋同樓及跨樓單位的會員物業資料
const configureLinkedPropertyFixture = (): void => {
  const buildingA = '1234500';
  const buildingB = '1234600';
  const unitA = '12345000401';
  const unitG = '12345000407';
  const unitB = '12346000201';
  const me = JSON.parse(JSON.stringify(initialMemberSession)) as typeof initialMemberSession;

  me.primary_community = {
    public_id: buildingA,
    name_zh: '測試甲大廈',
    name_en: 'Test Building A',
    address_text: '測試甲大廈',
  };
  me.bound_building_ids = [buildingA, buildingB];
  me.bound_flat_unit_ids = [unitA, unitG, unitB];
  me.residence_floor = '04';
  me.residence_unit = 'G';
  me.ismart_msg = {
    ...me.ismart_msg,
    client_building_flat_units_permissions: [unitA, unitG, unitB],
  };
  mocks.session.me = me;

  mocks.fetchMemberBuildings.mockResolvedValue([
    { building_id: buildingA, buildname_chi: '測試甲大廈', buildname: 'Test Building A' },
    { building_id: buildingB, buildname_chi: '測試乙大廈', buildname: 'Test Building B' },
  ]);
  mocks.fetchMemberUnits.mockImplementation(async (buildingID: string) => {
    if (buildingID === buildingA) {
      return [
        { unit_id: unitA, floor: '04', unit: 'A' },
        { unit_id: unitG, floor: '04', unit: 'G' },
      ];
    }
    if (buildingID === buildingB) {
      return [{ unit_id: unitB, floor: '02', unit: 'A' }];
    }
    return [];
  });
  mocks.updateMe.mockImplementation(async (payload: {
    primary_community_id: string;
    primary_community_name: string;
    residence_floor: string;
    residence_unit: string;
  }) => ({
    data: {
      data: {
        ...mocks.session.me,
        primary_community: {
          ...mocks.session.me.primary_community,
          public_id: payload.primary_community_id,
          name_zh: payload.primary_community_name,
        },
        residence_floor: payload.residence_floor,
        residence_unit: payload.residence_unit,
      },
    },
  }));
};

// 1. 掛載會員中心並等待初始化請求完成
const mountAccountPage = async () => {
  const wrapper = mount(AccountMyPage, {
    global: {
      plugins: [i18n],
      stubs: {
        AccountWalletPage: true,
        PropertyMyPage: true,
      },
    },
  });
  await flushPromises();
  return wrapper;
};

// 2. 按導覽文字切換會員中心面板
const clickPanel = async (wrapper: Awaited<ReturnType<typeof mountAccountPage>>, label: string) => {
  const button = wrapper.findAll('.work-nav-item').find((item) => item.text().includes(label));
  if (!button) {
    throw new Error(`Missing panel button: ${label}`);
  }
  await button.trigger('click');
  await flushPromises();
};

describe('AccountMyPage lazy data loading', () => {
  afterEach(() => {
    document.body.querySelectorAll('.work-ismart-modal').forEach((element) => element.remove());
  });

  beforeEach(() => {
    setActivePinia(createPinia());
    applyLocale('zh-HK');
    mocks.session.me = JSON.parse(JSON.stringify(initialMemberSession)) as typeof initialMemberSession;
    mocks.route.path = '/account/profile';
    mocks.route.query = {};
    mocks.fetchMemberBuildings.mockReset();
    mocks.fetchMemberUnits.mockReset();
    mocks.fetchPublicBuildings.mockReset();
    mocks.fetchPublicUnits.mockReset();
    mocks.fetchIsmartFlatBindings.mockReset();
    mocks.fetchIsmartOwnerBindingRequests.mockReset();
    mocks.submitOwnerBindingRequest.mockReset();
    mocks.loadCurrentUser.mockReset();
    mocks.pushToast.mockReset();
    mocks.changeMemberIsmartPassword.mockReset();
    mocks.updateMemberIsmartProfile.mockReset();
    mocks.routerPush.mockReset();
    Reflect.deleteProperty(mocks.session.me, 'avatar_key');
    mocks.fetchMemberBuildings.mockResolvedValue([
      { building_id: '0999900', buildname_chi: '測試1大廈', buildname: 'Test Building 1' },
    ]);
    mocks.fetchMemberUnits.mockResolvedValue([
      { unit_id: '09999000401', floor: '04', unit: 'G' },
      { unit_id: '09999000103', floor: '01', unit: 'C' },
    ]);
    mocks.fetchPublicBuildings.mockResolvedValue([
      { building_id: '0999900', buildname_chi: '測試1大廈', buildname: 'Test Building 1' },
    ]);
    mocks.fetchPublicUnits.mockResolvedValue([
      { unit_id: '09999000401', floor: '04', unit: 'G' },
    ]);
    mocks.fetchIsmartFlatBindings.mockResolvedValue({ items: [], count: 0, building_options: [] });
    mocks.fetchIsmartOwnerBindingRequests.mockResolvedValue({ items: [], count: 0, building_options: [] });
    mocks.submitOwnerBindingRequest.mockResolvedValue({ building_options: [] });
    mocks.routerPush.mockResolvedValue(undefined);
    mocks.updateMemberIsmartProfile.mockResolvedValue({
      data: {
        data: {
          ...mocks.session.me,
          ismart_account_profile: {
            ...mocks.session.me.ismart_account_profile,
            account_email: 'new.patrick@seventy2.hk',
          },
        },
      },
    });
    mocks.changeMemberIsmartPassword.mockResolvedValue({ data: { data: {} } });
    mocks.updateMe.mockReset();
    mocks.updateMe.mockResolvedValue({
      data: {
        data: {
          ...mocks.session.me,
          avatar_key: 'avo-03',
        },
      },
    });
  });

  it('loads only account data initially and reuses the current building unit response', async () => {
    const wrapper = await mountAccountPage();

    expect(mocks.fetchMemberBuildings).toHaveBeenCalledTimes(1);
    expect(mocks.fetchMemberUnits).toHaveBeenCalledTimes(1);
    expect(mocks.fetchMemberUnits).toHaveBeenCalledWith('0999900');
    expect(mocks.fetchPublicBuildings).not.toHaveBeenCalled();
    expect(mocks.fetchPublicUnits).not.toHaveBeenCalled();
    expect(wrapper.text()).not.toContain(i18n.global.t('account.center.nav.subaccounts'));
  });

  it('shows the registered username beside the account avatar', async () => {
    const wrapper = await mountAccountPage();

    expect(wrapper.find('.work-account-name').text()).toBe('patrick');
    expect(wrapper.find('.work-account-name').text()).not.toBe('Guest Member');
  });

  it('loads owner-binding directories only after the property-binding panel is opened', async () => {
    const wrapper = await mountAccountPage();

    await clickPanel(wrapper, i18n.global.t('account.center.nav.propertyBinding'));
    expect(mocks.fetchPublicBuildings).toHaveBeenCalledTimes(1);
    expect(mocks.fetchPublicUnits).toHaveBeenCalledTimes(1);

    await clickPanel(wrapper, i18n.global.t('account.center.nav.account'));
    await clickPanel(wrapper, i18n.global.t('account.center.nav.propertyBinding'));
    expect(mocks.fetchPublicBuildings).toHaveBeenCalledTimes(1);
    expect(mocks.fetchPublicUnits).toHaveBeenCalledTimes(1);
  });

  it('loads owner-binding data when entering through the property-binding query', async () => {
    mocks.route.query = { panel: 'property-binding' };
    await mountAccountPage();

    expect(mocks.fetchPublicBuildings).toHaveBeenCalledTimes(1);
    expect(mocks.fetchPublicUnits).toHaveBeenCalledTimes(1);
  });

  // 3.1 iSmart 待審批申請鎖定提交
  it('locks the owner-binding submit action while an iSmart request is pending', async () => {
    mocks.fetchIsmartOwnerBindingRequests.mockResolvedValue({
      items: [
        {
          owner_reg_id: 11,
          building_id: '0999900',
          building_name: '測試1大廈',
          ownedflat: ['09999000401'],
          cli_role: '業主',
          status: 'pending_approval',
          created_date: '2026-09-01',
        },
      ],
      count: 1,
      building_options: [],
    });
    const wrapper = await mountAccountPage();

    await clickPanel(wrapper, i18n.global.t('account.center.nav.propertyBinding'));

    const submitButton = wrapper.findAll('button').find((item) => item.text().includes(i18n.global.t('account.center.binding.pending')));
    expect(submitButton).toBeTruthy();
    expect(wrapper.findAll('.binding-status-item')).toHaveLength(1);
    expect(wrapper.find('.binding-status-item').text()).toContain(i18n.global.t('account.center.binding.pending'));
  });
  // 3.2 iSmart 已拒絕申請不再鎖定提交
  it('reopens the owner-binding submit action after an iSmart request is rejected', async () => {
    mocks.fetchIsmartOwnerBindingRequests.mockResolvedValue({
      items: [
        {
          owner_reg_id: 12,
          building_id: '0999900',
          building_name: '測試1大廈',
          ownedflat: ['09999000401'],
          cli_role: '業主',
          status: 'rejected',
          created_date: '2026-09-01',
          approval_date: '2026-09-10',
        },
      ],
      count: 1,
      building_options: [],
    });
    const wrapper = await mountAccountPage();

    await clickPanel(wrapper, i18n.global.t('account.center.nav.propertyBinding'));

    const submitButton = wrapper.findAll('button').find((item) => item.text().includes(i18n.global.t('account.center.binding.submitApproval')));
    expect(submitButton).toBeTruthy();
    expect(wrapper.find('.binding-status-title').text()).toBe('測試1大廈');
    expect(wrapper.find('.binding-status-item').text()).toContain(i18n.global.t('account.center.binding.rejected'));
  });

  // 3.2.1 已拒絕申請的單位標題沿用頁面中文大廈名稱格式
  it('formats the rejected request unit label with the AJO building name', async () => {
    mocks.fetchIsmartFlatBindings.mockResolvedValue({
      items: [
        {
          relation_info_id: 2851,
          building_id: '0999900',
          building_name: 'DEMO1 BUILDING',
          unit_id: '09999000417',
          unit_display: 'DEMO1 BUILDING 04 G',
          floor: '04',
          unit: 'G',
          cli_role: '登記業主',
        },
      ],
      count: 1,
      building_options: [],
    });
    mocks.fetchIsmartOwnerBindingRequests.mockResolvedValue({
      items: [
        {
          owner_reg_id: 13,
          building_id: '0999900',
          building_name: 'DEMO1 BUILDING',
          ownedflat: ['09999000417'],
          cli_role: '業主',
          status: 'rejected',
          created_date: '2026-09-01',
          approval_date: '2026-09-10',
        },
      ],
      count: 1,
      building_options: [],
    });
    const wrapper = await mountAccountPage();

    await clickPanel(wrapper, i18n.global.t('account.center.nav.propertyBinding'));

    expect(wrapper.find('.binding-status-title').text()).toBe('測試1大廈 / 04 / G');
  });

  // 3.3 關聯物業以 iSmart 即時綁定為準
  it('prefers live iSmart flat bindings over the local login snapshot', async () => {
    mocks.fetchIsmartFlatBindings.mockResolvedValue({
      items: [
        {
          relation_info_id: 266,
          building_id: '0999900',
          building_name: '測試1大廈',
          unit_id: '09999000103',
          unit_display: '測試1大廈 01 C',
          floor: '01',
          unit: 'C',
          cli_role: '登記業主',
        },
      ],
      count: 1,
      building_options: [],
    });
    const wrapper = await mountAccountPage();

    await clickPanel(wrapper, i18n.global.t('account.center.nav.account'));

    const propertyCells = wrapper.findAll('.work-property-table tbody tr td:first-child').map((item) => item.text());
    expect(propertyCells.filter(Boolean)).toEqual(['測試1大廈 / 01 / C']);
  });

  it('opens notification settings from the member centre navigation', async () => {
    const wrapper = await mountAccountPage();

    await clickPanel(wrapper, i18n.global.t('account.center.nav.preferences'));

    expect(mocks.routerPush).toHaveBeenCalledWith('/account/profile/preferences');
  });

  it('updates the login password from the linked property registration action', async () => {
    const wrapper = await mountAccountPage();

    const fieldLabels = wrapper.findAll('.work-account-card .work-account-field span').map((item) => item.text());
    expect(fieldLabels).toEqual([
      i18n.global.t('auth.publisherIdentityType'),
      i18n.global.t('account.center.account.registeredUsername'),
      i18n.global.t('account.center.account.registeredName'),
      i18n.global.t('account.center.account.identityCardNumber'),
      i18n.global.t('account.center.account.registeredPhone'),
      i18n.global.t('account.center.account.registeredEmail'),
    ]);
    expect(wrapper.text()).not.toContain(i18n.global.t('account.profile.districtCode'));
    expect(wrapper.text()).not.toContain('iSmart');
    expect(wrapper.text()).not.toContain(i18n.global.t('account.profile.saveUnit'));
    expect(wrapper.find('.work-account-card').text()).not.toContain(i18n.global.t('account.profile.localPassword'));
    expect(wrapper.find('.work-account-card').text()).not.toContain(i18n.global.t('account.center.account.changeIsmartPassword'));
    expect(wrapper.text()).toContain(i18n.global.t('account.center.account.propertyOwnershipSole'));
    expect(wrapper.text()).toContain(i18n.global.t('account.center.account.propertyFunctionRental'));
    expect(wrapper.text()).toContain(i18n.global.t('account.center.account.propertyFunctionHome'));
    expect(wrapper.text()).toContain(i18n.global.t('account.center.account.propertyFunctionDisabled'));

    const propertyCells = wrapper.findAll('.work-property-table tbody tr td:first-child').map((item) => item.text());
    expect(propertyCells.filter(Boolean)).toEqual([
      '測試1大廈 / 01 / C',
      '測試1大廈 / 04 / G',
    ]);

    const changePasswordButton = wrapper.findAll('button').find((item) =>
      item.text().includes(i18n.global.t('account.center.account.changeIsmartPassword')),
    );
    expect(changePasswordButton).toBeTruthy();
    expect(wrapper.find('.work-section-heading__actions').text()).toContain(i18n.global.t('account.center.account.editIsmartProfile'));

    await changePasswordButton!.trigger('click');
    await flushPromises();

    const dialog = document.body.querySelector('.work-password-dialog');
    expect(dialog).toBeTruthy();
    const inputs = dialog?.querySelectorAll<HTMLInputElement>('input[type="password"]');
    expect(inputs).toHaveLength(3);
    const values = ['current-password', 'new-password', 'new-password'];
    inputs?.forEach((input, index) => {
      input.value = values[index] ?? '';
      input.dispatchEvent(new Event('input', { bubbles: true }));
    });

    (dialog?.querySelector('button[type="submit"]') as HTMLButtonElement).click();
    await flushPromises();

    expect(mocks.changeMemberIsmartPassword).toHaveBeenCalledWith({
      old_password: 'current-password',
      new_password: 'new-password',
      new_password_confirm: 'new-password',
    });
  });

  it('submits the editable iSmart profile fields through the documented proxy', async () => {
    const wrapper = await mountAccountPage();

    const editButton = wrapper.findAll('button').find((item) =>
      item.text().includes(i18n.global.t('account.center.account.editIsmartProfile')),
    );
    expect(editButton).toBeTruthy();
    await editButton!.trigger('click');
    await flushPromises();

    const dialog = document.body.querySelector('.work-ismart-dialog');
    expect(dialog).toBeTruthy();
    const emailInput = dialog?.querySelector('input[type="email"]') as HTMLInputElement | null;
    expect(emailInput).toBeTruthy();
    emailInput!.value = 'new.patrick@seventy2.hk';
    emailInput!.dispatchEvent(new Event('input', { bubbles: true }));

    const saveButton = dialog?.querySelector('button[type="submit"]') as HTMLButtonElement | null;
    expect(saveButton).toBeTruthy();
    saveButton!.click();
    await flushPromises();

    expect(mocks.updateMemberIsmartProfile).toHaveBeenCalledWith(expect.objectContaining({
      account_email: 'new.patrick@seventy2.hk',
      account_phone: '+85261980774',
      account_name: 'patrick',
    }));
  });

  it('saves the selected linked property from the function table bind action', async () => {
    const wrapper = await mountAccountPage();

    const bindButton = wrapper.findAll('.work-property-table button').find((item) =>
      item.text().includes(i18n.global.t('account.center.account.switchProperty')),
    );
    expect(bindButton).toBeTruthy();
    await bindButton!.trigger('click');
    await flushPromises();

    expect(mocks.updateMe).toHaveBeenCalledWith({
      primary_community_id: '0999900',
      primary_community_name: '測試1大廈',
      residence_floor: '01',
      residence_unit: 'C',
    });
  });

  // 4.1 已連結物業以獨立切換欄顯示目前狀態
  it('renders a separate property switch column and marks only the current unit', async () => {
    configureLinkedPropertyFixture();
    const wrapper = await mountAccountPage();
    const switchLabel = i18n.global.t('account.center.account.switchProperty');
    const currentLabel = i18n.global.t('account.center.account.currentProperty');

    const headers = wrapper.findAll('.work-property-table th').map((item) => item.text());
    expect(headers).toContain(switchLabel);
    expect(headers).not.toContain(currentLabel);

    const rows = wrapper.findAll('.work-property-table tbody tr');
    const currentRow = rows.find((item) => item.text().includes('測試甲大廈 / 04 / G'));
    const sameBuildingRow = rows.find((item) => item.text().includes('測試甲大廈 / 04 / A'));
    const crossBuildingRow = rows.find((item) => item.text().includes('測試乙大廈 / 02 / A'));

    expect(currentRow?.text()).toContain(currentLabel);
    expect(sameBuildingRow?.text()).toContain(switchLabel);
    expect(crossBuildingRow?.text()).toContain(switchLabel);
  });

  // 4.2 同樓及跨樓切換沿用完整物業上下文
  it('switches to same-building and cross-building units after a successful save', async () => {
    configureLinkedPropertyFixture();
    const wrapper = await mountAccountPage();
    const switchLabel = i18n.global.t('account.center.account.switchProperty');
    const currentLabel = i18n.global.t('account.center.account.currentProperty');

    const sameBuildingRow = wrapper.findAll('.work-property-table tbody tr')
      .find((item) => item.text().includes('測試甲大廈 / 04 / A'));
    const sameBuildingButton = sameBuildingRow?.find('button');
    expect(sameBuildingButton?.text()).toContain(switchLabel);
    await sameBuildingButton!.trigger('click');
    await flushPromises();

    expect(mocks.updateMe).toHaveBeenLastCalledWith({
      primary_community_id: '1234500',
      primary_community_name: '測試甲大廈',
      residence_floor: '04',
      residence_unit: 'A',
    });
    expect(wrapper.findAll('.work-property-table tbody tr')
      .find((item) => item.text().includes('測試甲大廈 / 04 / A'))?.text()).toContain(currentLabel);

    const crossBuildingRow = wrapper.findAll('.work-property-table tbody tr')
      .find((item) => item.text().includes('測試乙大廈 / 02 / A'));
    const crossBuildingButton = crossBuildingRow?.find('button');
    expect(crossBuildingButton?.text()).toContain(switchLabel);
    await crossBuildingButton!.trigger('click');
    await flushPromises();

    expect(mocks.updateMe).toHaveBeenLastCalledWith({
      primary_community_id: '1234600',
      primary_community_name: '測試乙大廈',
      residence_floor: '02',
      residence_unit: 'A',
    });
    expect(wrapper.findAll('.work-property-table tbody tr')
      .find((item) => item.text().includes('測試乙大廈 / 02 / A'))?.text()).toContain(currentLabel);
  });

  // 4.3 切換失敗不提前改變目前物業標記
  it('keeps the current property marker when switching fails', async () => {
    configureLinkedPropertyFixture();
    mocks.updateMe.mockRejectedValueOnce(new Error('save failed'));
    const wrapper = await mountAccountPage();
    const switchLabel = i18n.global.t('account.center.account.switchProperty');
    const currentLabel = i18n.global.t('account.center.account.currentProperty');
    const rows = () => wrapper.findAll('.work-property-table tbody tr');

    const targetRow = rows().find((item) => item.text().includes('測試甲大廈 / 04 / A'));
    await targetRow?.find('button').trigger('click');
    await flushPromises();

    expect(rows().find((item) => item.text().includes('測試甲大廈 / 04 / G'))?.text()).toContain(currentLabel);
    expect(rows().find((item) => item.text().includes('測試甲大廈 / 04 / A'))?.text()).toContain(switchLabel);
  });

  // 4.4 保存期間忽略重複切換提交
  it('does not submit the same property switch more than once while saving', async () => {
    configureLinkedPropertyFixture();
    let resolveUpdate: ((value: unknown) => void) | undefined;
    mocks.updateMe.mockImplementationOnce(() => new Promise((resolve) => {
      resolveUpdate = resolve;
    }));
    const wrapper = await mountAccountPage();
    const targetRow = wrapper.findAll('.work-property-table tbody tr')
      .find((item) => item.text().includes('測試甲大廈 / 04 / A'));
    const button = targetRow?.find('button');
    expect(button?.exists()).toBe(true);

    await button!.trigger('click');
    await button!.trigger('click');
    expect(mocks.updateMe).toHaveBeenCalledTimes(1);

    resolveUpdate?.({
      data: {
        data: mocks.session.me,
      },
    });
    await flushPromises();
  });

  // 4.5 每個已連結物業只顯示一次解除綁定，未接通時維持原有登記
  it.each(['zh-HK', 'en'] as const)('shows one unbind action per property without changing bindings in %s', async (locale) => {
    configureLinkedPropertyFixture();
    applyLocale(locale);
    const wrapper = await mountAccountPage();
    const unbindLabel = i18n.global.t('account.center.account.unbindProperty');
    const currentLabel = i18n.global.t('account.center.account.currentProperty');
    const switchLabel = i18n.global.t('account.center.account.switchProperty');
    const rows = wrapper.findAll('.work-property-table tbody tr');
    const originalSession = structuredClone(mocks.session.me);
    const originalTableText = wrapper.find('.work-property-table tbody').text();

    expect(rows).toHaveLength(9);
    expect(rows.filter((row) => row.text().includes(currentLabel))).toHaveLength(1);
    expect(rows.flatMap((row) => row.findAll('button')).filter((button) => button.text() === switchLabel)).toHaveLength(2);

    for (const [index, row] of rows.entries()) {
      const unbindButtons = row.findAll('button').filter((button) => button.text() === unbindLabel);
      expect(unbindButtons).toHaveLength(index % 3 === 0 ? 1 : 0);
      if (unbindButtons[0]) {
        await unbindButtons[0].trigger('click');
      }
    }
    await flushPromises();

    expect(mocks.pushToast).toHaveBeenCalledTimes(3);
    expect(mocks.pushToast).toHaveBeenLastCalledWith(
      i18n.global.t('account.center.account.unbindUnavailable'),
      'error',
    );
    expect(mocks.updateMe).not.toHaveBeenCalled();
    expect(mocks.submitOwnerBindingRequest).not.toHaveBeenCalled();
    expect(mocks.session.me).toEqual(originalSession);
    expect(wrapper.find('.work-property-table tbody').text()).toBe(originalTableText);
  });

  it('updates the member avatar from the account centre portrait', async () => {
    const wrapper = await mountAccountPage();

    const avatarButton = wrapper.find('.work-account-avatar');
    expect(avatarButton.exists()).toBe(true);
    expect(avatarButton.element.tagName).toBe('BUTTON');
    expect(avatarButton.find('img').attributes('src')).toBe('/avatars/avo-07.svg');

    await avatarButton.trigger('click');
    await flushPromises();

    const dialog = document.body.querySelector('.work-avatar-dialog');
    expect(dialog).toBeTruthy();
    const options = dialog?.querySelectorAll<HTMLButtonElement>('.work-avatar-picker__item') ?? [];
    expect(options).toHaveLength(16);

    const currentOption = Array.from(options).find((item) => item.classList.contains('is-active'));
    expect(currentOption).toBeTruthy();
    currentOption!.click();
    await flushPromises();
    expect(mocks.updateMe).not.toHaveBeenCalled();
    expect(document.body.querySelector('.work-avatar-dialog')).toBeNull();

    await avatarButton.trigger('click');
    await flushPromises();
    const nextDialog = document.body.querySelector('.work-avatar-dialog');
    const nextOptions = nextDialog?.querySelectorAll<HTMLButtonElement>('.work-avatar-picker__item') ?? [];
    const selected = Array.from(nextOptions).find((item) => item.querySelector('img')?.getAttribute('src') === '/avatars/avo-03.svg');
    expect(selected).toBeTruthy();
    selected!.click();
    await flushPromises();

    expect(mocks.updateMe).toHaveBeenCalledWith({ avatar_key: 'avo-03' });
    expect(mocks.pushToast).toHaveBeenCalledWith(i18n.global.t('account.profile.avatarUpdateSuccess'), 'success');
    expect(document.body.querySelector('.work-avatar-dialog')).toBeNull();
  });
});
