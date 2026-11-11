/*
 * 公開列表分頁測試。
 * 1. 驗證首頁、中間頁與末頁的頁碼及省略號。
 * 2. 驗證前後翻頁及邊界狀態。
 */
import { mount } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import ListingPagination from './ListingPagination.vue';

describe('ListingPagination', () => {
  const props = {
    currentPage: 1,
    totalPages: 28,
    previousLabel: 'Previous',
    nextLabel: 'Next',
    navigationLabel: 'Pagination',
  };

  it('shows the first three and last two pages with a gap', () => {
    const wrapper = mount(ListingPagination, { props });
    expect(wrapper.findAll('.listing-pages-button').map((button) => button.text())).toEqual(['', '1', '2', '3', '27', '28', '']);
    expect(wrapper.find('.listing-pages-gap').text()).toBe('...');
    expect(wrapper.get('[aria-label="Previous"]').attributes('disabled')).toBeDefined();
  });

  it('shows both gaps and emits navigation for a middle page', async () => {
    const wrapper = mount(ListingPagination, { props: { ...props, currentPage: 14 } });
    expect(wrapper.findAll('.listing-pages-gap')).toHaveLength(2);
    expect(wrapper.get('[aria-current="page"]').text()).toBe('14');
    await wrapper.get('[aria-label="Previous"]').trigger('click');
    await wrapper.get('[aria-label="Next"]').trigger('click');
    expect(wrapper.emitted('select')).toEqual([[13], [15]]);
  });

  it('disables next on the last page and fills single-page gaps', () => {
    const wrapper = mount(ListingPagination, { props: { ...props, currentPage: 6, totalPages: 6 } });
    expect(wrapper.findAll('.listing-pages-gap')).toHaveLength(0);
    expect(wrapper.get('[aria-label="Next"]').attributes('disabled')).toBeDefined();
    expect(wrapper.findAll('.listing-pages-button').map((button) => button.text())).toEqual(['', '1', '2', '3', '4', '5', '6', '']);
  });
});
