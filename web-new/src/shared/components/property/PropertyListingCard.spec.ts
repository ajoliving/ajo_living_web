/*
 * 樓盤搜尋與發佈預覽共用卡片測試。
 * 1. 驗證分區與物業類型標籤位於卡片內容最上方。
 * 2. 驗證標題、屋苑位置及座數樓層單位依序顯示。
 */
import { mount } from '@vue/test-utils';
import { describe, expect, it, vi } from 'vitest';

import type { PropertyListingCardViewModel } from '@/model/property';
import PropertyListingCard from '@/shared/components/property/PropertyListingCard.vue';

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string): string => key,
  }),
}));

// 1. 建立固定卡片資料
const createCard = (): PropertyListingCardViewModel => ({
  id: 'property-1',
  propertyType: '住宅',
  publisherLabel: '業主盤',
  tags: [
    { label: '銅鑼灣' },
    { label: '住宅' },
  ],
  title: '測試樓盤',
  location: '銅鑼灣 · 中國人壽大廈 C.L.I.',
  facts: ['C座', '中層', 'A室'],
  priceKind: 'sale',
  price: 'HK$1,200萬',
  priceUnit: '',
  area: '實用面積 230呎',
  areaPrice: '@HK$52,174/呎',
  pills: ['公屋'],
  favorite: false,
});

describe('PropertyListingCard', () => {
  it('renders tags before title, location and property facts', () => {
    const wrapper = mount(PropertyListingCard, {
      props: {
        card: createCard(),
        preview: true,
      },
    });
    const bodyChildren = Array.from(wrapper.find('.property-listing-card__body').element.children);

    expect(bodyChildren.map((element) => element.className)).toEqual([
      'property-listing-card__tags',
      '',
      'property-listing-card__location',
      'property-listing-card__facts',
      'property-listing-card__price',
      'property-listing-card__area',
      'property-listing-card__pills',
    ]);
    expect(wrapper.find('.property-listing-card__tags').text()).toContain('銅鑼灣');
    expect(wrapper.find('.property-listing-card__tags').text()).toContain('住宅');
    expect(wrapper.find('h3').text()).toBe('測試樓盤');
    expect(wrapper.find('.property-listing-card__location').text()).toBe('銅鑼灣 · 中國人壽大廈 C.L.I.');
    expect(wrapper.find('.property-listing-card__facts').text()).toBe('C座 · 中層 · A室');
  });
});
