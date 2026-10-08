/*
 * 全域頂部導航測試。
 * 1. 驗證未登入時，手機頂部與底部導覽保留登入入口。
 * 2. 驗證登入後底部導覽切換至會員中心。
 * 3. 驗證手機抽屜按登入狀態顯示登入或通知入口。
 * 4. 驗證抽屜遮罩點擊可關閉。
 * 5. 驗證三項主題使用太陽、水星軌道及月亮，並可套用選定色彩。
 * 6. 驗證共用導航文案全部經由 i18n 資源輸出。
 * 7. 驗證主題入口關閉時不渲染主題選擇器。
 */
import { mount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { usePreferenceStore } from '@/stores/preferences';
import { useSessionStore } from '@/stores/session';
import type { CurrentMemberProfile } from '@/model/user';

import AppHeader from './AppHeader.vue';
import themeIconsSvg from './theme-icons.svg?raw';

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

  afterEach(() => {
    vi.unstubAllEnvs();
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

  it('opens the three theme choices and applies cyan', async () => {
    const preferenceStore = usePreferenceStore();
    const wrapper = mountHeader();

    await wrapper.find('.theme-toggle').trigger('click');

    expect(wrapper.findAll('.theme-choice')).toHaveLength(3);
    await wrapper.findAll('.theme-choice')[1]?.trigger('click');

    expect(preferenceStore.theme).toBe('cyan');
    expect(document.documentElement.style.getPropertyValue('--color-primary')).toBe('8 127 140');
    expect(wrapper.find('.theme-menu').exists()).toBe(false);
  });

  it('shows a white moon instead of the dark swatch for the night theme', () => {
    const preferenceStore = usePreferenceStore();
    preferenceStore.setTheme('night');
    const wrapper = mountHeader();

    expect(wrapper.find('.theme-toggle-moon').exists()).toBe(true);
    expect(wrapper.find('.theme-toggle-dot').exists()).toBe(false);
    expect(wrapper.find('.theme-toggle-icon use').attributes('href')).toMatch(/#moon$/);
  });

  // 3. 三項主題的選單與入口使用相同圖示，選擇後保留偏好及選取狀態。
  it('uses matching sun, orbit, and moon SVGs in the menu and active theme button', async () => {
    const preferenceStore = usePreferenceStore();
    const wrapper = mountHeader();
    const icons = ['sun', 'orbit', 'moon'];
    const themes = ['orange-05', 'cyan', 'night'];
    const sprite = new DOMParser().parseFromString(themeIconsSvg, 'image/svg+xml');

    expect(sprite.querySelector('parsererror')).toBeNull();
    expect(sprite.querySelectorAll('symbol')).toHaveLength(3);
    expect(wrapper.find('.theme-toggle-icon use').attributes('href')).toMatch(/#sun$/);

    for (const [index, icon] of icons.entries()) {
      expect(sprite.getElementById(icon)?.querySelector('path, circle')).not.toBeNull();
      await wrapper.find('.theme-toggle').trigger('click');

      const choices = wrapper.findAll('.theme-choice');
      expect(choices.map((choice) =>
        new URL(choice.find('use').attributes('href')!, 'http://localhost').hash,
      )).toEqual(['#sun', '#orbit', '#moon']);
      expect(choices[index]!.find('svg').attributes('aria-hidden')).toBe('true');

      await choices[index]!.trigger('click');

      expect(preferenceStore.theme).toBe(themes[index]);
      expect(window.localStorage.getItem('ajoliving.theme')).toBe(themes[index]);
      expect(wrapper.find('.theme-menu').exists()).toBe(false);
      expect(wrapper.find('.theme-toggle-icon use').attributes('href')).toMatch(new RegExp(`#${icon}$`));

      await wrapper.find('.theme-toggle').trigger('click');
      expect(wrapper.findAll('.theme-choice')[index]!.attributes('aria-checked')).toBe('true');
      await wrapper.find('.theme-toggle').trigger('click');
    }
  });

  it('hides the theme picker when the theme entry is disabled', () => {
    vi.stubEnv('VITE_ENABLE_THEME_PICKER', '0');

    const wrapper = mountHeader();

    expect(wrapper.find('.theme-picker').exists()).toBe(false);
    expect(wrapper.find('.theme-toggle').exists()).toBe(false);
    expect(wrapper.find('.nav-login').exists()).toBe(true);
    expect(wrapper.find('.locale-toggle').exists()).toBe(true);
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
