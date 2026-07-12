/*
 * 訊息管理頁狀態測試。
 * 1. 驗證會話列表不會自動開啟第一條會話。
 * 2. 驗證進入、返回與直接連結會話的路由狀態。
 */
import { defineComponent } from 'vue';
import { flushPromises, mount } from '@vue/test-utils';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useMarketplaceChatPage } from './chat';

const mocks = vi.hoisted(() => ({
  fetchChatDetail: vi.fn(),
  fetchChats: vi.fn(),
  fetchMessages: vi.fn(),
  markChatRead: vi.fn(),
  pushToast: vi.fn(),
  replace: vi.fn(),
  route: {
    params: {} as Record<string, string>,
    query: {} as Record<string, string>,
  },
}));

vi.mock('vue-router', () => ({
  useRoute: () => mocks.route,
  useRouter: () => ({ replace: mocks.replace }),
}));

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key }),
}));

vi.mock('@/httpapis/chats', () => ({
  fetchChatDetail: mocks.fetchChatDetail,
  fetchChats: mocks.fetchChats,
  markChatRead: mocks.markChatRead,
}));

vi.mock('@/httpapis/messages', () => ({
  fetchMessages: mocks.fetchMessages,
  sendMessage: vi.fn(),
}));

vi.mock('@/stores/feedback', () => ({
  useFeedbackStore: () => ({ pushToast: mocks.pushToast }),
}));

vi.mock('@/stores/preferences', () => ({
  usePreferenceStore: () => ({ locale: 'zh-HK' }),
}));

vi.mock('@/stores/session', () => ({
  useSessionStore: () => ({
    currentUser: { public_id: 'member-self' },
    isAuthenticated: true,
  }),
}));

const conversation = {
  chat_id: 'chat-1',
  chat_type: 'direct_listing_chat',
  listing_id: 'listing-1',
  listing_title: '測試樓盤',
  last_message_preview: '你好',
  last_message_at: '2026-07-12T12:00:00Z',
  unread_count: 0,
  peer: {
    user_id: 'member-peer',
    public_id: 'member-peer',
    display_name: '測試會員',
    role_in_chat: 'buyer',
  },
};

// 1. 建立可操作聊天狀態的測試元件
const TestHost = defineComponent({
  setup() {
    return useMarketplaceChatPage();
  },
  template: `
    <div>
      <span class="selected-chat">{{ selectedChatId }}</span>
      <button class="select-chat" type="button" @click="handleSelectChat('chat-1')">select</button>
      <button class="back-to-chats" type="button" @click="handleBackToChats">back</button>
    </div>
  `,
});

describe('useMarketplaceChatPage', () => {
  beforeEach(() => {
    mocks.route.params = {};
    mocks.route.query = {};
    mocks.replace.mockReset();
    mocks.pushToast.mockReset();
    mocks.fetchChats.mockReset();
    mocks.fetchChatDetail.mockReset();
    mocks.fetchMessages.mockReset();
    mocks.markChatRead.mockReset();
    mocks.fetchChats.mockResolvedValue({ data: { data: { items: [conversation] } } });
    mocks.fetchChatDetail.mockResolvedValue({
      data: {
        data: {
          chat_type: 'direct_listing_chat',
          participants: [
            { user_id: 'member-self', public_id: 'member-self' },
            { user_id: 'member-peer', public_id: 'member-peer' },
          ],
        },
      },
    });
    mocks.fetchMessages.mockResolvedValue({ data: { data: { items: [] } } });
    mocks.markChatRead.mockResolvedValue(undefined);
  });

  it('keeps the conversation list visible until a conversation is selected', async () => {
    const wrapper = mount(TestHost);
    await flushPromises();

    expect(wrapper.find('.selected-chat').text()).toBe('');
    expect(mocks.replace).not.toHaveBeenCalled();

    await wrapper.find('.select-chat').trigger('click');
    await flushPromises();
    expect(wrapper.find('.selected-chat').text()).toBe('chat-1');
    expect(mocks.replace).toHaveBeenLastCalledWith('/account/chat/chat-1');

    await wrapper.find('.back-to-chats').trigger('click');
    await flushPromises();
    expect(wrapper.find('.selected-chat').text()).toBe('');
    expect(mocks.replace).toHaveBeenLastCalledWith('/account/chat');
  });

  it('opens a valid conversation from a direct link', async () => {
    mocks.route.params = { conversationId: 'chat-1' };
    const wrapper = mount(TestHost);
    await flushPromises();

    expect(wrapper.find('.selected-chat').text()).toBe('chat-1');
    expect(mocks.fetchChatDetail).toHaveBeenCalledWith('chat-1');
  });
});
