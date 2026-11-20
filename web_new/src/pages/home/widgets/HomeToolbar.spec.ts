/*
 * 首頁工具列回歸測試。
 * 1. 驗證公開屋苑選擇不影響其他路由查詢。
 * 2. 驗證目前物業回退與既有搜尋路由。
 */
import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils';
import { afterEach, describe, expect, it } from 'vitest';
import { createMemoryHistory, createRouter } from 'vue-router';

import i18n, { applyLocale } from '@/i18n';

import HomeContentLinks from './HomeContentLinks.vue';
import HomePropertyNews from './HomePropertyNews.vue';
import HomeToolbar from './HomeToolbar.vue';

enableAutoUnmount(afterEach);
afterEach(() => applyLocale('zh-HK'));

// 1. 保留實際 RouterLink 及 query 行為，避免使用導航替身。
const mountToolbar = async (location: string, buildingId?: string) => {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/:pathMatch(.*)*', component: { template: '<div />' } }],
  });
  await router.push(location);
  const wrapper = mount(HomeToolbar, {
    global: { plugins: [router, i18n] },
    props: { buildingId },
  });
  return { router, wrapper };
};

describe('HomeToolbar', () => {
  it('defaults guests to the first public estate and changes only its query', async () => {
    const { router, wrapper } = await mountToolbar('/?from=home#summary');
    expect(wrapper.get('select').element.value).toBe('seaview');
    await wrapper.get('select').setValue('summit');
    await flushPromises();
    expect(router.currentRoute.value.fullPath).toBe('/?from=home&estate=summit#summary');
    expect(wrapper.get('select').element.value).toBe('summit');
  });

  it('allows bound members to return from a public preset to their property', async () => {
    const { router, wrapper } = await mountToolbar('/?estate=fleur', 'building-2');
    expect(wrapper.get('select').element.value).toBe('fleur');
    await wrapper.get('select').setValue('');
    await flushPromises();
    expect(router.currentRoute.value.query.estate).toBeUndefined();
    expect(wrapper.get('select').element.value).toBe('');
  });

  it('does not accept an unknown or repeated estate and opens the real product search', async () => {
    const { router, wrapper } = await mountToolbar('/?estate=fleur&estate=summit');
    expect(wrapper.get('select').element.value).toBe('seaview');
    expect(wrapper.get('a').attributes('href')).toBe('/notifications');
    await wrapper.get('form').trigger('submit');
    expect(router.currentRoute.value.path).toBe('/');
    await wrapper.get('input').setValue('  牛奶  ');
    await wrapper.get('form').trigger('submit');
    await flushPromises();
    expect(router.currentRoute.value.path).toBe('/supermarket-offers');
    expect(router.currentRoute.value.query.q).toBe('牛奶');
  });
});

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
