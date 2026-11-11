/*
 * 管理中心手機導覽回歸測試。
 * 1. 驗證功能首頁、明確面板路由與瀏覽器返回。
 */
import { flushPromises, mount } from '@vue/test-utils';
import { createMemoryHistory, createRouter } from 'vue-router';
import { afterEach, describe, expect, it, vi } from 'vitest';
import i18n from '@/i18n';
import ManagementPage from './Page.vue';

vi.mock('./members/Page.vue', () => ({ default: { template: '<div>Members</div>' } }));
vi.mock('./agency_companies/Page.vue', () => ({ default: { template: '<div>Agency</div>' } }));
vi.mock('./property_sales/Page.vue', () => ({ default: { template: '<div>Properties</div>' } }));
vi.mock('./reward_ads/Page.vue', () => ({ default: { template: '<div>Ads</div>' } }));
vi.mock('./secondhand_listings/Page.vue', () => ({ default: { template: '<div>Furniture</div>' } }));
vi.mock('./serviced_apartments/Page.vue', () => ({ default: { template: '<div>Homes</div>' } }));
vi.mock('./system_notices/Page.vue', () => ({ default: { template: '<div>Notices</div>' } }));
vi.mock('./wallet_transactions/Page.vue', () => ({ default: { template: '<div>Transactions</div>' } }));
vi.mock('./building_chats/Page.vue', () => ({ default: { template: '<div>Chats</div>' } }));
vi.mock('./widgets/ManagementIcctvPanel.vue', () => ({ default: { template: '<div>ICCTV</div>' } }));
vi.mock('./widgets/ManagementAdSettingsPanel.vue', () => ({ default: { template: '<div>Settings</div>' } }));

describe('management mobile navigation', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('opens the default panel with an explicit tab and returns to the home', async () => {
    vi.stubGlobal('matchMedia', () => ({ matches: true }));
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: '/account/marketplace/management', component: ManagementPage }],
    });
    await router.push('/account/marketplace/management');
    await router.isReady();
    const wrapper = mount(ManagementPage, { global: { plugins: [router, i18n] } });
    expect(wrapper.classes()).toContain('is-mobile-home');
    expect(wrapper.findAll('.work-nav-item')).toHaveLength(11);
    await wrapper.findAll('.work-nav-item')[0]!.trigger('click');
    await flushPromises();
    expect(router.currentRoute.value.query.tab).toBe('admin-furniture');
    expect(wrapper.classes()).not.toContain('is-mobile-home');
    await wrapper.get('.management-mobile-back button').trigger('click');
    await flushPromises();
    expect(router.currentRoute.value.query.tab).toBeUndefined();
    expect(wrapper.classes()).toContain('is-mobile-home');
  });
});
