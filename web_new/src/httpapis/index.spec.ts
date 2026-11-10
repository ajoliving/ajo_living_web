/*
 * HTTP 用戶端業務授權測試。
 * 1. 驗證 POS 與 iSmart 業務層授權失效不會被視為 AJO 主登入失效。
 * 2. 驗證一般 AJO 401 仍交由全域登入失效流程處理。
 */
import axios from 'axios';
import { describe, expect, it } from 'vitest';

import { isIntegrationBusinessAuthError } from '@/httpapis';

// 1. 建立 Axios 401 錯誤
const createAuthError = (url: string, message: string, code = 'AUTH_REQUIRED') =>
  new axios.AxiosError(
    'Request failed with status code 401',
    'ERR_BAD_REQUEST',
    { url, headers: new axios.AxiosHeaders() },
    undefined,
    {
      data: { code, message },
      status: 401,
      statusText: 'Unauthorized',
      headers: {},
      config: { url, headers: new axios.AxiosHeaders() },
    },
  );

// 2. 驗證業務層與主登入 401 邊界
describe('isIntegrationBusinessAuthError', () => {
  it.each([
    ['/me/pos/buildings', 'pos token expired'],
    ['/me/payments/pos/history', 'ismart login is required'],
    ['/me/ismart/subaccounts', 'ismart login is required'],
    ['/me/security/icctv/public-cameras', 'ismart login is required'],
  ])('recognizes business authentication failure for %s', (url, message) => {
    expect(isIntegrationBusinessAuthError(createAuthError(url, message))).toBe(true);
  });

  it('does not classify the current-member 401 as a business authentication failure', () => {
    expect(isIntegrationBusinessAuthError(createAuthError('/me', 'login is required'))).toBe(false);
  });
});
