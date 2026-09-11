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
  loadCurrentUser: vi.fn(),
  pushToast: vi.fn(),
  changeMemberIsmartPassword: vi.fn(),
  updateMemberIsmartProfile: vi.fn(),
  routerPush: vi.fn(),
  route: {
    path: '/account/profile',
    query: {} as Record<string, string>,
  },
  session: {
    me: {
      display_name: 'Member',
      public_id: 'MEMBER-1',
      primary_community: {
        public_id: '0999900',
        name_zh: '測試1大廈',
        name_en: 'Test Building 1',
        address_text: '測試1大廈',
      },
      bound_building_ids: ['0999900'],
      bound_flat_unit_ids: ['09999000401'],
      residence_floor: '04',
      residence_unit: 'G',
      ismart_username: 'patrick',
      ismart_bound_phone: '+85261980774',
      ismart_msg: {
        username: 'patrick',
        email: 'patrick@seventy2.hk',
        phone: '+85261980774',
        client_building_flat_units_permissions: ['09999000401'],
      },
      ismart_account_profile: {
        account_code: 'patrick',
        account_phone: '+85261980774',
        account_email: 'patrick@seventy2.hk',
        owner_name_en: 'patrick',
        owner_name_zh: '林曉洸',
        account_name: 'patrick',
        identity_number: 'A123456(7)',
        legal_entity: '',
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
  submitMemberIsmartOwnerBindingRequest: vi.fn(),
}));

vi.mock('@/httpapis/me', () => ({
  changeMemberIsmartPassword: mocks.changeMemberIsmartPassword,
  updateMe: vi.fn(),
  updateMemberIsmartProfile: mocks.updateMemberIsmartProfile,
}));

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
    mocks.route.path = '/account/profile';
    mocks.route.query = {};
    mocks.fetchMemberBuildings.mockReset();
    mocks.fetchMemberUnits.mockReset();
    mocks.fetchPublicBuildings.mockReset();
    mocks.fetchPublicUnits.mockReset();
    mocks.loadCurrentUser.mockReset();
    mocks.pushToast.mockReset();
    mocks.changeMemberIsmartPassword.mockReset();
    mocks.updateMemberIsmartProfile.mockReset();
    mocks.routerPush.mockReset();
    mocks.fetchMemberBuildings.mockResolvedValue([
      { building_id: '0999900', buildname_chi: '測試1大廈', buildname: 'Test Building 1' },
    ]);
    mocks.fetchMemberUnits.mockResolvedValue([
      { unit_id: '09999000401', floor: '04', unit: 'G' },
    ]);
    mocks.fetchPublicBuildings.mockResolvedValue([
      { building_id: '0999900', buildname_chi: '測試1大廈', buildname: 'Test Building 1' },
    ]);
    mocks.fetchPublicUnits.mockResolvedValue([
      { unit_id: '09999000401', floor: '04', unit: 'G' },
    ]);
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

  it('opens notification settings from the member centre navigation', async () => {
    const wrapper = await mountAccountPage();

    await clickPanel(wrapper, i18n.global.t('account.center.nav.preferences'));

    expect(mocks.routerPush).toHaveBeenCalledWith('/account/profile/preferences');
  });

  it('updates the iSmart password from the account information action', async () => {
    const wrapper = await mountAccountPage();

    const passwordField = wrapper.findAll('.work-account-field').find((item) =>
      item.text().includes(i18n.global.t('account.profile.localPassword')),
    );
    expect(passwordField?.text()).toContain(i18n.global.t('account.center.account.passwordManaged'));
    const changePasswordButton = passwordField?.find('.work-action.work-compact-action');
    expect(changePasswordButton?.text()).toContain(i18n.global.t('account.center.account.changeIsmartPassword'));

    const residentUnitField = wrapper.findAll('.work-account-field').find((item) =>
      item.text().includes(i18n.global.t('account.center.account.residentUnits')),
    );
    expect(residentUnitField?.find('.work-action').exists()).toBe(false);

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
});
