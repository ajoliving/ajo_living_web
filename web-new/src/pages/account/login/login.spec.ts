/*
 * 登入註冊欄位校驗測試。
 * 1. 確認用戶名稱、電郵、姓名、證件、電話及密碼分別返回準確提示。
 * 2. 確認統一登入只校驗必填欄位，電話格式由後端辨識。
 */
import { describe, expect, it } from 'vitest';

import {
  parsePhoneFormInput,
  resolveLoginValidationErrors,
  resolveRegistrationValidationError,
  resolveRegistrationValidationErrors,
} from './login';

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
