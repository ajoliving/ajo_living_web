/*
 * 登入頁表單面板測試。
 * 1. 確認登入狀態只展示單一帳戶輸入，不再展示登入方式分頁。
 * 2. 確認用戶註冊的綁定大廈欄位按勾選狀態展開。
 */
import { mount } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';

import i18n from '@/i18n';

import LoginFormPanel from './LoginFormPanel.vue';

// 1. 建立表單面板必要屬性
const createProps = () => ({
  emailAction: 'login' as const,
  emailActionSwitchLabel: '註冊',
  buildingOptions: [],
  buildingsLoading: false,
  engName: '',
  chiName: '',
  email: '',
  password: '',
  confirmPassword: '',
  username: '',
  phone: '',
  phoneCountryCode: '+852',
  publisherIdentityType: 'personal',
  ismartAccount: '',
  primaryCommunityId: '',
  residenceFloor: '',
  residenceUnit: '',
  idCard: '',
  agencyLicenseNumber: '',
  agencyLicenseFile: null,
  agencyContactName: '',
  shouldBindResidence: false,
  rememberMe: true,
  residenceFloorOptions: [],
  residenceUnitOptions: [],
  submitting: false,
  unitsLoading: false,
  emailPlaceholder: 'resident@example.com',
  footerPrompt: '還沒有帳戶？',
  isAuthenticated: false,
  publisherIdentityOptions: [],
  registrationStep: 1 as const,
  submitLabel: '登入',
  validationErrors: {},
});

describe('LoginFormPanel', () => {
  it('renders one unified account field without sign-in mode tabs', () => {
    const wrapper = mount(LoginFormPanel, {
      props: createProps(),
      global: { plugins: [i18n] },
    });

    expect(wrapper.find('[role="tablist"]').exists()).toBe(false);
    expect(wrapper.findAll('input[autocomplete="username"]')).toHaveLength(1);
    expect(wrapper.find('input[type="tel"]').exists()).toBe(false);
    expect(wrapper.text()).toContain('手提電話 / 電郵 / iSmart username');
  });

  it('expands the building-binding fields only after the checkbox is selected', async () => {
    const wrapper = mount(LoginFormPanel, {
      props: createProps(),
      global: { plugins: [i18n] },
    });

    await wrapper.setProps({ emailAction: 'register', registrationStep: 2 });

    expect(wrapper.find('input[type="tel"]').exists()).toBe(false);
    expect(wrapper.find('input[autocomplete="email"]').exists()).toBe(false);
    expect(wrapper.find('input[autocomplete="username"]').exists()).toBe(true);
    expect(wrapper.text()).toContain('綁定大廈');
    expect(wrapper.text()).not.toContain('英文姓名');

    await wrapper.setProps({ shouldBindResidence: true });

    expect(wrapper.text()).toContain('英文姓名');
    expect(wrapper.text()).toContain('大廈');
  });

  it('hides first-step account fields and renders equal-width actions on step two', async () => {
    const wrapper = mount(LoginFormPanel, {
      props: createProps(),
      global: { plugins: [i18n] },
    });

    await wrapper.setProps({ emailAction: 'register', registrationStep: 2, publisherIdentityType: 'agency_company' });

    expect(wrapper.text()).not.toContain('帳戶類型');
    expect(wrapper.find('input[type="tel"]').exists()).toBe(false);
    expect(wrapper.find('input[autocomplete="email"]').exists()).toBe(false);
    expect(wrapper.find('.login-registration-actions').exists()).toBe(true);
    expect(wrapper.findAll('.login-registration-actions > button')).toHaveLength(2);
  });

  it('uses licence registration fields for an individual agent', async () => {
    const wrapper = mount(LoginFormPanel, {
      props: createProps(),
      global: { plugins: [i18n] },
    });

    await wrapper.setProps({ emailAction: 'register', publisherIdentityType: 'individual_agent', registrationStep: 2 });

    expect(wrapper.find('input[autocomplete="username"]').exists()).toBe(false);
    expect(wrapper.text()).toContain('牌照號碼');
    expect(wrapper.text()).toContain('上載個人牌照');
    expect(wrapper.text()).toContain('審批通過後方可正常登入及使用');
  });

  it('keeps the first registration step limited to account type, mobile, and email', async () => {
    const wrapper = mount(LoginFormPanel, {
      props: createProps(),
      global: { plugins: [i18n] },
    });

    await wrapper.setProps({ emailAction: 'register' });

    expect(wrapper.text()).toContain('註冊步驟 1 / 2');
    expect(wrapper.find('input[type="tel"]').exists()).toBe(true);
    expect(wrapper.find('input[autocomplete="username"]').exists()).toBe(false);
    expect(wrapper.find('input[autocomplete="new-password"]').exists()).toBe(false);
  });
});
