/*
 * 登入註冊欄位校驗測試。
 * 1. 確認電郵、英文姓名、電話及密碼分別返回準確提示。
 * 2. 確認中國內地手機號碼可通過格式校驗。
 */
import { describe, expect, it } from 'vitest';

import axios from 'axios';

import {
  canFallbackToIsmartLogin,
  parsePhoneFormInput,
  resolveLoginValidationErrors,
  resolveRegistrationValidationError,
  resolveRegistrationValidationErrors,
} from './login';

describe('resolveRegistrationValidationError', () => {
  it('returns the email error when registration email is missing', () => {
    expect(resolveRegistrationValidationError('', 'stitch', '+86', '15666823185', 'password123'))
      .toBe('auth.invalidEmail');
  });

  it('returns the password error for a valid +86 registration with a seven-character password', () => {
    expect(resolveRegistrationValidationError('member@example.com', 'stitch', '+86', '15666823185', '1090119'))
      .toBe('auth.registerPasswordTooShort');
  });

  it('returns field-specific name and phone errors', () => {
    expect(resolveRegistrationValidationError('member@example.com', 'A', '+86', '15666823185', 'password123'))
      .toBe('auth.registerEnglishNameInvalid');
    expect(resolveRegistrationValidationError('member@example.com', 'stitch', '+86', 'abc', 'password123'))
      .toBe('auth.invalidPhone');
  });

  it('accepts complete valid registration input', () => {
    expect(resolveRegistrationValidationError('member@example.com', 'stitch', '+86', '15666823185', '10901190'))
      .toBeNull();
  });

  it('returns every invalid registration field', () => {
    expect(resolveRegistrationValidationErrors('', 'A', '+86', 'abc', '1090119')).toEqual({
      email: 'auth.invalidEmail',
      engName: 'auth.registerEnglishNameInvalid',
      phone: 'auth.invalidPhone',
      password: 'auth.registerPasswordTooShort',
    });
  });
});

describe('resolveLoginValidationErrors', () => {
  it('returns separate account and password errors', () => {
    expect(resolveLoginValidationErrors('username', '', '+852', '', '')).toEqual({
      ismartAccount: 'auth.accountRequired',
      password: 'auth.passwordRequired',
    });
  });

  it('returns separate phone and password errors', () => {
    expect(resolveLoginValidationErrors('phone', '', '+852', 'abc', '')).toEqual({
      phone: 'auth.invalidPhone',
      password: 'auth.passwordRequired',
    });
  });
});

describe('canFallbackToIsmartLogin', () => {
  it('falls back only for local credential errors', () => {
    expect(canFallbackToIsmartLogin(new axios.AxiosError('username or password is incorrect'))).toBe(true);
    expect(canFallbackToIsmartLogin(new axios.AxiosError('Network Error'))).toBe(false);
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
