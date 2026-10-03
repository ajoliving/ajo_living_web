/*
 * 登入註冊欄位校驗測試。
 * 1. 確認用戶名稱、電郵、姓名、證件、電話及密碼分別返回準確提示。
 * 2. 確認統一登入只校驗必填欄位，電話格式由後端辨識。
 * 3. 確認註冊提交使用已填寫證件，並保留大廈綁定的必填要求。
 */
import { flushPromises, mount } from '@vue/test-utils';
import { defineComponent, nextTick } from 'vue';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import {
  parsePhoneFormInput,
  resolveLoginValidationErrors,
  resolveRegistrationValidationError,
  resolveRegistrationValidationErrors,
  useLoginPage,
} from './login';

const mocks = vi.hoisted(() => ({
  checkRegistrationAvailability: vi.fn(),
  fetchPosBuildings: vi.fn(),
  fetchPosBuildingUnits: vi.fn(),
  fetchLoginHero: vi.fn(),
  push: vi.fn(),
  pushToast: vi.fn(),
  registerEmailAccount: vi.fn(),
}));

vi.mock('vue-router', () => ({
  useRoute: () => ({ query: {} }),
  useRouter: () => ({ push: mocks.push }),
}));

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key }),
}));

vi.mock('@/httpapis/auth', () => ({
  checkRegistrationAvailability: mocks.checkRegistrationAvailability,
}));

vi.mock('@/httpapis/building', () => ({
  fetchPosBuildings: mocks.fetchPosBuildings,
  fetchPosBuildingUnits: mocks.fetchPosBuildingUnits,
}));

vi.mock('@/httpapis/home-content', () => ({
  fetchLoginHero: mocks.fetchLoginHero,
}));

vi.mock('@/stores/feedback', () => ({
  useFeedbackStore: () => ({ pushToast: mocks.pushToast }),
}));

vi.mock('@/stores/session', () => ({
  useSessionStore: () => ({
    isAuthenticated: false,
    me: null,
    registerEmailAccount: mocks.registerEmailAccount,
  }),
}));

// 1. 建立可操作註冊提交流程的測試元件
const TestHost = defineComponent({
  setup: useLoginPage,
  template: '<div />',
});

// 2. 按兩步註冊流程準備個人帳戶及可選大廈單位
const prepareRegistration = async (shouldBindResidence = true) => {
  const wrapper = mount(TestHost);
  wrapper.vm.toggleEmailAction();
  Object.assign(wrapper.vm.formState, {
    email: 'resident@example.com',
    phone: '61234567',
    username: 'resident',
    engName: 'TEST RESIDENT',
    password: 'password123',
    confirmPassword: 'password123',
  });
  await wrapper.vm.handleSubmit();
  wrapper.vm.formState.shouldBindResidence = shouldBindResidence;
  await flushPromises();
  if (shouldBindResidence) {
    wrapper.vm.formState.primaryCommunityID = '0999900';
    await flushPromises();
    wrapper.vm.formState.residenceFloor = '01';
    await nextTick();
    wrapper.vm.formState.residenceUnit = 'B';
    await nextTick();
  }
  return wrapper;
};

describe('resolveRegistrationValidationError', () => {
  it('returns the email error when registration email is missing', () => {
    expect(resolveRegistrationValidationError('', 'stitch', 'stitch', '+86', '15666823185', 'password123', 'password123'))
      .toBe('auth.invalidEmail');
  });

  it('returns the password error for a valid +86 registration with a seven-character password', () => {
    expect(resolveRegistrationValidationError('member@example.com', 'stitch', 'stitch', '+86', '15666823185', '1090119', '1090119'))
      .toBe('auth.registerPasswordTooShort');
  });

  it('returns field-specific name and phone errors', () => {
    expect(resolveRegistrationValidationError('member@example.com', 'stitch', 'A', '+86', '15666823185', 'password123', 'password123', true))
      .toBe('auth.registerEnglishNameInvalid');
    expect(resolveRegistrationValidationError('member@example.com', 'stitch', 'stitch', '+86', 'abc', 'password123', 'password123'))
      .toBe('auth.invalidPhone');
  });

  it('accepts complete valid registration input', () => {
    expect(resolveRegistrationValidationError('member@example.com', 'stitch', 'stitch', '+86', '15666823185', '10901190', '10901190'))
      .toBeNull();
  });

  it('returns every invalid registration field', () => {
    expect(resolveRegistrationValidationErrors('', 'stitch', 'A', '+86', 'abc', '1090119', '1090119', true, 'building', '12', 'A', 'A1234567')).toEqual({
      email: 'auth.invalidEmail',
      engName: 'auth.registerEnglishNameInvalid',
      phone: 'auth.invalidPhone',
      password: 'auth.registerPasswordTooShort',
    });
  });

  it('requires a complete residence selection only when building binding is selected', () => {
    expect(resolveRegistrationValidationErrors(
      'member@example.com',
      'stitch',
      'stitch',
      '+86',
      '15666823185',
      'password123',
      'password123',
      true,
      '',
      '',
      '',
      'A1234567',
    )).toEqual({ residence: 'auth.residenceBindingRequired' });
  });

  it('requires an identity document when building binding is selected', () => {
    expect(resolveRegistrationValidationErrors(
      'member@example.com',
      'stitch',
      'stitch',
      '+86',
      '15666823185',
      'password123',
      'password123',
      true,
      'building',
      '12',
      'A',
    )).toEqual({ idCard: 'auth.idCardRequired' });
  });

  it('returns a dedicated mismatch error for confirmation password', () => {
    expect(resolveRegistrationValidationErrors(
      'member@example.com',
      'stitch',
      'stitch',
      '+86',
      '15666823185',
      'password123',
      'password456',
    )).toEqual({ confirmPassword: 'auth.registerPasswordMismatch' });
  });
});

describe('resolveLoginValidationErrors', () => {
  it('returns separate account and password errors', () => {
    expect(resolveLoginValidationErrors('', '')).toEqual({
      ismartAccount: 'auth.accountRequired',
      password: 'auth.passwordRequired',
    });
  });

  it('accepts phone, email and username formats without guessing the account type', () => {
    expect(resolveLoginValidationErrors('+852 6123 4567', 'password123')).toEqual({});
    expect(resolveLoginValidationErrors('member@example.com', 'password123')).toEqual({});
    expect(resolveLoginValidationErrors('patrick', 'password123')).toEqual({});
  });
});

describe('parsePhoneFormInput', () => {
  it('splits a compact mainland number using the supported +86 code', () => {
    expect(parsePhoneFormInput('+852', '+86156668231883')).toEqual({
      phoneCountryCode: '+86',
      phoneNumber: '156668231883',
    });
  });

  it('rejects unsupported international country codes', () => {
    expect(parsePhoneFormInput('+852', '+85361234567')).toEqual({
      phoneCountryCode: '',
      phoneNumber: '',
    });
  });
});

describe('useLoginPage registration submission', () => {
  beforeEach(() => {
    vi.resetAllMocks();
    mocks.checkRegistrationAvailability.mockResolvedValue({
      data: { data: { email_available: true, phone_available: true } },
    });
    mocks.fetchPosBuildings.mockResolvedValue([
      { building_id: '0999900', buildname_chi: '測試大廈' },
    ]);
    mocks.fetchPosBuildingUnits.mockResolvedValue([
      { unit_id: '099990001B', floor: '01', unit: 'B' },
    ]);
    mocks.fetchLoginHero.mockResolvedValue({ data: { data: { items: [] } } });
    mocks.registerEmailAccount.mockResolvedValue(undefined);
  });

  it.each(['A123456(3)', 'A1234563', ' A123456(3) '])(
    'submits a completed residence registration with identity input %s',
    async (idCard) => {
      const wrapper = await prepareRegistration();
      try {
        wrapper.vm.formState.idCard = idCard;
        await nextTick();
        await wrapper.vm.handleSubmit();

        expect(wrapper.vm.validationErrors).toEqual({});
        expect(mocks.registerEmailAccount).toHaveBeenCalledExactlyOnceWith(expect.objectContaining({
          id_card: idCard.trim(),
          eng_name: 'TEST RESIDENT',
          primary_community_id: '0999900',
          primary_community_name: '測試大廈',
          residence_floor: '01',
          residence_unit: 'B',
        }));
        expect(mocks.push).toHaveBeenCalledWith('/');
      } finally {
        wrapper.unmount();
      }
    },
  );

  it.each(['', '   '])('blocks blank identity input %j and accepts a corrected value', async (idCard) => {
    const wrapper = await prepareRegistration();
    try {
      wrapper.vm.formState.idCard = idCard;
      await nextTick();
      await wrapper.vm.handleSubmit();

      expect(wrapper.vm.validationErrors).toEqual({ idCard: 'auth.idCardRequired' });
      expect(mocks.registerEmailAccount).not.toHaveBeenCalled();
      expect(mocks.push).not.toHaveBeenCalled();

      wrapper.vm.formState.idCard = 'A123456(3)';
      await nextTick();
      expect(wrapper.vm.validationErrors).toEqual({});
      await wrapper.vm.handleSubmit();

      expect(mocks.registerEmailAccount).toHaveBeenCalledExactlyOnceWith(expect.objectContaining({
        id_card: 'A123456(3)',
      }));
    } finally {
      wrapper.unmount();
    }
  });

  it('does not require an identity document without residence binding', async () => {
    const wrapper = await prepareRegistration(false);
    try {
      await wrapper.vm.handleSubmit();

      expect(wrapper.vm.validationErrors).toEqual({});
      expect(mocks.registerEmailAccount).toHaveBeenCalledOnce();
      expect(mocks.registerEmailAccount.mock.calls[0][0]).not.toHaveProperty('id_card');
      expect(mocks.registerEmailAccount.mock.calls[0][0]).not.toHaveProperty('primary_community_id');
    } finally {
      wrapper.unmount();
    }
  });
});
