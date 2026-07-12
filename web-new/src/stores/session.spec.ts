/*
 * 會員登入狀態測試。
 * 1. 鎖定啟動還原登入狀態的並發等待行為。
 * 2. 避免路由守衛在會員資料尚未完成載入時誤判 staff 權限。
 */
import { createPinia, setActivePinia } from 'pinia';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { CurrentMemberProfile } from '@/model/user';

import { useSessionStore } from './session';

const mocks = vi.hoisted(() => ({
  fetchMe: vi.fn(),
}));

vi.mock('@/httpapis/me', () => ({
  fetchMe: mocks.fetchMe,
}));

// 1. 建立可控制完成時間的 Promise
const createDeferred = <T>() => {
  let resolveDeferred!: (value: T) => void;
  const promise = new Promise<T>((resolve) => {
    resolveDeferred = resolve;
  });

  return {
    promise,
    resolve: resolveDeferred,
  };
};

// 2. 建立 staff 會員資料
const buildStaffMember = (): CurrentMemberProfile => ({
  public_id: 'staff-member',
  email: 'staff@example.com',
  phone_country_code: '+852',
  phone_number: '60000000',
  member_status: 'active',
  member_type: 'owner',
  is_staff: true,
  role: 'staff',
  roles: ['staff'],
  permissions: ['marketplace:manage'],
  display_name: 'Staff Member',
  avatar_url: '',
  publisher_identity_type: 'owner',
  district_code: '',
  profile_completed: true,
});

describe('session store hydration', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    window.localStorage.clear();
    mocks.fetchMe.mockReset();
  });

  it('waits for an in-flight hydrate call before exposing loaded staff state', async () => {
    window.localStorage.setItem('ajoliving.access-token', 'access-token');
    window.localStorage.setItem('ajoliving.refresh-token', 'refresh-token');

    const deferred = createDeferred<{ data: { data: CurrentMemberProfile } }>();
    mocks.fetchMe.mockReturnValue(deferred.promise);

    const sessionStore = useSessionStore();
    const firstHydrate = sessionStore.hydrateSession();
    const secondHydrate = sessionStore.hydrateSession();

    expect(mocks.fetchMe).toHaveBeenCalledTimes(1);
    expect(sessionStore.isLoaded).toBe(false);

    deferred.resolve({ data: { data: buildStaffMember() } });
    await Promise.all([firstHydrate, secondHydrate]);

    expect(sessionStore.isLoaded).toBe(true);
    expect(sessionStore.isHydrating).toBe(false);
    expect(sessionStore.currentUser.is_staff).toBe(true);
  });
});
