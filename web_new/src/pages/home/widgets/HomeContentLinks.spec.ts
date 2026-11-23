/*
 * 首頁物業資訊與內容入口回歸測試。
 * 1. 驗證物業資訊原始來源及 English 文案。
 * 2. 驗證樓市資訊與會話使用既有路由。
 */
import { enableAutoUnmount, mount } from '@vue/test-utils';
import { afterEach, describe, expect, it } from 'vitest';
import { createMemoryHistory, createRouter } from 'vue-router';

import i18n, { applyLocale } from '@/i18n';

import HomeContentLinks from './HomeContentLinks.vue';
import HomePropertyNews from './HomePropertyNews.vue';

enableAutoUnmount(afterEach);
afterEach(() => applyLocale('zh-HK'));

// 1. 保留正式資訊來源與 RouterLink，確認公開入口可導覽。
describe('Home editorial links', () => {
  it('renders the three real sources and uses the official English source where available', () => {
    applyLocale('en');
    const wrapper = mount(HomePropertyNews, { global: { plugins: [i18n] } });
    expect(wrapper.findAll('a').map((link) => link.attributes('href'))).toEqual([
      'https://www.hk01.com/article/60400435',
      'https://www.pmsa.org.hk/en/press-release/20260814',
      'https://www.stheadline.com/society/3627484',
    ]);
    expect(wrapper.findAll('time')).toHaveLength(3);
    expect(wrapper.text()).toContain('Property information');
  });

  it('uses the existing market and conversations views', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: '/:pathMatch(.*)*', component: { template: '<div />' } }],
    });
    await router.push('/');
    const wrapper = mount(HomeContentLinks, { global: { plugins: [router, i18n] } });
    expect(wrapper.findAll('a').map((link) => link.attributes('href'))).toEqual([
      '/trend',
      '/notifications?tab=conversations',
    ]);
  });
});
