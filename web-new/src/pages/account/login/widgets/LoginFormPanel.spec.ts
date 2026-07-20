/*
 * 登入頁表單面板測試。
 * 1. 確認登入狀態只展示單一帳戶輸入，不再展示登入方式分頁。
 * 2. 確認住戶註冊仍保留手提電話欄位。
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
  phone: '',
  phoneCountryCode: '+852',
  publisherIdentityType: 'personal',
  ismartAccount: '',
  primaryCommunityId: '',
  residenceFloor: '',
  residenceUnit: '',
  idCard: '',
  remark: '',
  gender: '' as const,
  isReceiveEmail: true,
  rememberMe: true,
  residenceFloorOptions: [],
  residenceUnitOptions: [],
  submitting: false,
  unitsLoading: false,
  emailPlaceholder: 'resident@example.com',
  footerPrompt: '還沒有帳戶？',
  isAuthenticated: false,
  publisherIdentityOptions: [],
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

  it('keeps the mobile field in resident registration', async () => {
    const wrapper = mount(LoginFormPanel, {
      props: createProps(),
      global: { plugins: [i18n] },
    });

    await wrapper.setProps({ emailAction: 'register' });

    expect(wrapper.find('input[type="tel"]').exists()).toBe(true);
    expect(wrapper.find('input[autocomplete="username"]').exists()).toBe(false);
  });
});
