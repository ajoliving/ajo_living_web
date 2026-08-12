/*
 * 全域頂部導航測試。
 * 1. 驗證未登入時，手機頂部與底部導覽保留登入入口。
 * 2. 驗證登入後底部導覽切換至會員中心。
 * 3. 驗證手機抽屜按登入狀態顯示登入或通知入口。
 * 4. 驗證抽屜遮罩點擊可關閉。
 * 5. 驗證主題選單提供六十組主題色並可套用選定色彩。
 * 6. 驗證共用導航文案全部經由 i18n 資源輸出。
 */
import { mount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { usePreferenceStore } from '@/stores/preferences';
import { useSessionStore } from '@/stores/session';
import type { CurrentMemberProfile } from '@/model/user';

import AppHeader from './AppHeader.vue';

const routeState = vi.hoisted(() => ({
  path: '/',
  fullPath: '/',
}));

vi.mock('vue-router', () => ({
  RouterLink: {
    props: ['to'],
    template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>',
  },
  useRoute: () => routeState,
  useRouter: () => ({
    push: vi.fn(),
  }),
}));

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string): string => {
      const messages: Record<string, string> = {
        'common.action.close': '關閉',
        'common.action.signOut': '登出',
        'common.action.switchLanguage': '切換語系',
        'common.theme.selector': '選擇主題色',
        'common.theme.orangeGroup': '橙色漸進',
        'common.theme.otherGroup': '其他主題',
        'common.locale.zhHkShort': '繁中',
        'common.locale.enShort': 'EN',
        'nav.account': '帳戶',
        'nav.ajoPay': 'AJO Pay',
        'nav.building': '我的大廈',
        'nav.combinedOffers': '綜合優惠',
        'nav.furniture': '家具',
        'nav.home': '首頁',
        'nav.login': '登入',
        'nav.marketplaceManagement': '管理',
        'nav.menu': '選單',
        'nav.memberCenter': '會員中心',
        'nav.myShort': '我的',
        'nav.notifications': '通知中心',
        'nav.offersShort': '優惠',
        'nav.properties': '樓盤租售',
        'nav.propertiesShort': '樓盤',
        'nav.servicedResidences': '服務式住宅',
        'nav.trend': '走勢',
      };

      return messages[key] ?? key;
    },
  }),
}));

// 1. 建立導航測試元件
const mountHeader = () =>
  mount(AppHeader, {
    global: {
      stubs: {
        Teleport: true,
      },
    },
  });

// 2. 建立登入會員資料
const buildMemberProfile = (isStaff = false): CurrentMemberProfile => ({
  public_id: isStaff ? 'staff-member' : 'member',
  email: 'member@example.com',
  phone_country_code: '+852',
  phone_number: '60000000',
  member_status: 'active',
  member_type: 'owner',
  is_staff: isStaff,
  role: isStaff ? 'staff' : 'member',
  roles: isStaff ? ['staff'] : ['member'],
  permissions: [],
  display_name: isStaff ? 'Staff Member' : 'Member',
  avatar_url: '',
  publisher_identity_type: 'owner',
  district_code: '',
  profile_completed: true,
});

describe('AppHeader mobile navigation', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    routeState.path = '/';
    routeState.fullPath = '/';
    window.localStorage.clear();
  });

  it('keeps a login entry in mobile navigation when signed out', () => {
    const preferenceStore = usePreferenceStore();
    preferenceStore.setMobileMenuOpen(true);

    const wrapper = mountHeader();
    const bottomItems = wrapper.findAll('.bnav-item');

    expect(bottomItems).toHaveLength(5);
    expect(bottomItems[4]?.attributes('href')).toBe('/login');
    expect(bottomItems[4]?.text()).toContain('登入');
    expect(wrapper.find('.mobile-menu-body').text()).toContain('登入');
    expect(wrapper.find('.mobile-menu-body').text()).not.toContain('通知中心');
    expect(wrapper.find('.mobile-login').attributes('href')).toBe('/login');
    expect(wrapper.find('.mobile-login').text()).toBe('登入');
  });

  it('points mobile account navigation to profile and shows notifications after sign in', () => {
    const preferenceStore = usePreferenceStore();
    const sessionStore = useSessionStore();
    preferenceStore.setMobileMenuOpen(true);
    sessionStore.accessToken = 'access-token';
    sessionStore.me = buildMemberProfile(false);

    const wrapper = mountHeader();
    const bottomItems = wrapper.findAll('.bnav-item');

    expect(bottomItems).toHaveLength(5);
    expect(bottomItems[4]?.attributes('href')).toBe('/profile');
    expect(bottomItems[4]?.text()).toContain('我的');
    expect(wrapper.find('.mobile-menu-body').text()).toContain('通知中心');
    expect(wrapper.find('.mobile-menu-body').text()).not.toContain('登入');
    expect(wrapper.find('.mobile-login').exists()).toBe(false);
  });

  it('keeps authenticated mobile function entries in the drawer', () => {
    const preferenceStore = usePreferenceStore();
    const sessionStore = useSessionStore();
    preferenceStore.setMobileMenuOpen(true);
    sessionStore.accessToken = 'access-token';
    sessionStore.me = buildMemberProfile(false);

    const wrapper = mountHeader();
    const drawerText = wrapper.find('.mobile-menu-body').text();

    expect(drawerText).toContain('AJO Pay');
    expect(drawerText).toContain('我的大廈');
    expect(drawerText).toContain('會員中心');
    expect(drawerText).toContain('通知中心');
    expect(drawerText).not.toContain('管理');
  });

  it('only shows management in the mobile drawer for staff members', () => {
    const preferenceStore = usePreferenceStore();
    const sessionStore = useSessionStore();
    preferenceStore.setMobileMenuOpen(true);
    sessionStore.accessToken = 'access-token';
    sessionStore.me = buildMemberProfile(true);

    const wrapper = mountHeader();
    const drawerText = wrapper.find('.mobile-menu-body').text();

    expect(drawerText).toContain('管理');
  });

  it('keeps menu and locale controls available in the mobile header', async () => {
    const preferenceStore = usePreferenceStore();
    const wrapper = mountHeader();
    const headerChildren = [...wrapper.find('.nav').element.children];
    const hamburgerIndex = headerChildren.findIndex((node) => node.classList.contains('hamburger'));
    const logoIndex = headerChildren.findIndex((node) => node.classList.contains('nav-logo'));
    const localeToggle = wrapper.find('.locale-toggle');

    expect(hamburgerIndex).toBeLessThan(logoIndex);
    expect(wrapper.find('.theme-toggle').exists()).toBe(true);
    expect(wrapper.find('.hamburger').attributes('aria-label')).toBe('選單');
    expect(localeToggle.attributes('aria-label')).toBe('切換語系');
    expect(localeToggle.text()).toBe('EN');

    await localeToggle.trigger('click');

    expect(preferenceStore.locale).toBe('en');
    expect(localeToggle.text()).toBe('繁中');
  });

  it('opens sixty theme choices and applies the selected theme', async () => {
    const preferenceStore = usePreferenceStore();
    const wrapper = mountHeader();

    await wrapper.find('.theme-toggle').trigger('click');

    expect(wrapper.findAll('.theme-swatch')).toHaveLength(60);
    await wrapper.findAll('.theme-swatch')[50]?.trigger('click');

    expect(preferenceStore.theme).toBe('jade');
    expect(document.documentElement.style.getPropertyValue('--color-primary')).toBe('13 139 100');
    expect(wrapper.find('.theme-menu').exists()).toBe(false);
  });

  it('closes the mobile drawer when the backdrop is tapped', async () => {
    const preferenceStore = usePreferenceStore();
    preferenceStore.setMobileMenuOpen(true);

    const wrapper = mountHeader();

    expect(wrapper.find('.mobile-menu-panel').exists()).toBe(true);

    await wrapper.find('.mobile-menu').trigger('click');

    expect(preferenceStore.mobile_menu_open).toBe(false);
    expect(wrapper.find('.mobile-menu').exists()).toBe(false);
  });
});
