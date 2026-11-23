/*
 * 監控播放框回歸測試。
 * 1. JSON 錯誤不進入 iframe，正常播放器保留原網址。
 * 2. 網絡失敗及切換鏡頭時不顯示舊結果。
 */
import { flushPromises, mount } from '@vue/test-utils';
import { afterEach, describe, expect, it, vi } from 'vitest';

import i18n, { applyLocale } from '@/i18n';
import CameraFrame from './CameraFrame.vue';

afterEach(() => vi.unstubAllGlobals());

describe('CameraFrame', () => {
  it('displays token errors as text without loading a JSON document', async () => {
    applyLocale('zh-HK');
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(
      '{"detail":"Invalid or expired token"}',
      { status: 401, headers: { 'content-type': 'application/json' } },
    )));
    const wrapper = mount(CameraFrame, {
      props: { src: 'https://camera.example/channel1?token=expired', title: 'Camera' },
      global: { plugins: [i18n] },
    });
    await flushPromises();
    expect(wrapper.text()).toBe('Invalid or expired token');
    expect(wrapper.find('iframe').exists()).toBe(false);
    expect(wrapper.find('input').exists()).toBe(false);
    wrapper.unmount();
  });

  it('preserves the URL and permissions for successful player pages', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('', {
      headers: { 'content-type': 'text/html' },
    })));
    const src = 'https://camera.example/channel1?token=valid';
    const wrapper = mount(CameraFrame, {
      props: { src, title: 'Camera' },
      global: { plugins: [i18n] },
    });
    await flushPromises();
    expect(wrapper.get('iframe').attributes('src')).toBe(src);
    expect(wrapper.get('iframe').attributes('allow')).toContain('fullscreen');
    wrapper.unmount();
  });

  it('shows a localized error when the request fails', async () => {
    applyLocale('en');
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')));
    const wrapper = mount(CameraFrame, {
      props: { src: 'https://camera.example/channel1', title: 'Camera' },
      global: { plugins: [i18n] },
    });
    await flushPromises();
    expect(wrapper.text()).toBe('Unable to load the camera view. Please refresh and try again.');
    expect(wrapper.find('iframe').exists()).toBe(false);
    wrapper.unmount();
  });

  it('ignores a failed request after switching to another camera', async () => {
    let rejectOld: (reason: Error) => void = () => {};
    vi.stubGlobal('fetch', vi.fn()
      .mockImplementationOnce(() => new Promise<Response>((_, reject) => { rejectOld = reject; }))
      .mockResolvedValueOnce(new Response('', { headers: { 'content-type': 'text/html' } })));
    const wrapper = mount(CameraFrame, {
      props: { src: 'https://camera.example/channel1', title: 'Camera' },
      global: { plugins: [i18n] },
    });
    await wrapper.setProps({ src: 'https://camera.example/channel2' });
    await flushPromises();
    rejectOld(new Error('Cancelled'));
    await flushPromises();
    expect(wrapper.get('iframe').attributes('src')).toBe('https://camera.example/channel2');
    expect(wrapper.text()).toBe('');
    wrapper.unmount();
  });
});
