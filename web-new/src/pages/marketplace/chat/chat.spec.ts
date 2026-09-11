/*
 * 訊息管理頁狀態測試。
 * 1. 驗證會話列表不會自動開啟第一條會話。
 * 2. 驗證進入、返回與直接連結會話的路由狀態。
 */
import { defineComponent } from 'vue';
import { flushPromises, mount } from '@vue/test-utils';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { useMarketplaceChatPage } from './chat';

const mocks = vi.hoisted(() => ({
  completeUpload: vi.fn(),
  createUploadPresign: vi.fn(),
  fetchRealtimeTicket: vi.fn(),
  fetchChatDetail: vi.fn(),
  fetchChats: vi.fn(),
  fetchMessages: vi.fn(),
  markChatRead: vi.fn(),
  pushToast: vi.fn(),
  replace: vi.fn(),
  sendMessage: vi.fn(),
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
  sendMessage: mocks.sendMessage,
}));

vi.mock('@/httpapis/uploads', () => ({
  abortMultipartUpload: vi.fn(),
  completeMultipartUpload: vi.fn(),
  completeUpload: mocks.completeUpload,
  createUploadPresign: mocks.createUploadPresign,
  initiateMultipartUpload: vi.fn(),
  listMultipartParts: vi.fn(),
  presignMultipartPart: vi.fn(),
}));

vi.mock('@/httpapis/realtime', () => ({
  fetchRealtimeTicket: mocks.fetchRealtimeTicket,
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

const secondConversation = {
  ...conversation,
  chat_id: 'chat-2',
  listing_id: 'listing-2',
  listing_title: '第二個測試樓盤',
  peer: {
    user_id: 'member-peer-2',
    public_id: 'member-peer-2',
    display_name: '第二位會員',
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
      <span class="attachment-status">{{ activeMessages[0]?.attachments[0]?.scan_status || '' }}</span>
      <span class="message-body">{{ activeMessages[0]?.body || '' }}</span>
      <span class="message-count">{{ activeMessages.length }}</span>
      <span class="last-message">{{ conversations[0]?.last_message || '' }}</span>
      <button class="select-chat" type="button" @click="handleSelectChat('chat-1')">select</button>
      <button class="select-chat-2" type="button" @click="handleSelectChat('chat-2')">select second</button>
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
    mocks.fetchRealtimeTicket.mockReset();
    mocks.fetchChatDetail.mockReset();
    mocks.fetchMessages.mockReset();
    mocks.markChatRead.mockReset();
    mocks.sendMessage.mockReset();
    mocks.createUploadPresign.mockReset();
    mocks.completeUpload.mockReset();
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
    mocks.fetchRealtimeTicket.mockResolvedValue({ data: { data: { ticket: 'realtime-ticket', expires_in: 90 } } });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('keeps the conversation list visible until a conversation is selected', async () => {
    const wrapper = mount(TestHost);
    await flushPromises();

    expect(wrapper.find('.selected-chat').text()).toBe('');
    expect(mocks.replace).not.toHaveBeenCalled();

    await wrapper.find('.select-chat').trigger('click');
    await flushPromises();
    expect(wrapper.find('.selected-chat').text()).toBe('chat-1');
    expect(mocks.replace).toHaveBeenLastCalledWith({
      path: '/notifications',
      query: { tab: 'conversations', conversationId: 'chat-1' },
    });

    await wrapper.find('.back-to-chats').trigger('click');
    await flushPromises();
    expect(wrapper.find('.selected-chat').text()).toBe('');
    expect(mocks.replace).toHaveBeenLastCalledWith({
      path: '/notifications',
      query: { tab: 'conversations' },
    });
  });

  it('opens a valid conversation from a direct link', async () => {
    mocks.route.params = { conversationId: 'chat-1' };
    const wrapper = mount(TestHost);
    await flushPromises();

    expect(wrapper.find('.selected-chat').text()).toBe('chat-1');
    expect(mocks.fetchChatDetail).toHaveBeenCalledWith('chat-1');
  });

  it('keeps the newer conversation history when an earlier request resolves late', async () => {
    let resolveFirstDetail: ((value: unknown) => void) | undefined;
    let resolveFirstMessages: ((value: unknown) => void) | undefined;
    const firstDetail = new Promise((resolve) => { resolveFirstDetail = resolve; });
    const firstMessages = new Promise((resolve) => { resolveFirstMessages = resolve; });
    mocks.fetchChats.mockResolvedValue({ data: { data: { items: [conversation, secondConversation] } } });
    mocks.fetchChatDetail.mockImplementation((chatId: string) => (
      chatId === 'chat-1'
        ? firstDetail
        : Promise.resolve({
            data: {
              data: {
                chat_type: 'direct_listing_chat',
                participants: [{ user_id: 'member-self', public_id: 'member-self' }, { user_id: 'member-peer-2', public_id: 'member-peer-2' }],
              },
            },
          })
    ));
    mocks.fetchMessages.mockImplementation((chatId: string) => (
      chatId === 'chat-1'
        ? firstMessages
        : Promise.resolve({
            data: {
              data: {
                items: [{ message_id: 'message-2', sender_user_id: 'member-peer-2', content: '第二個會話', message_type: 'text', status: 'sent', created_at: '2026-08-27T12:00:00Z' }],
              },
            },
          })
    ));

    const wrapper = mount(TestHost);
    await flushPromises();
    await wrapper.find('.select-chat').trigger('click');
    await wrapper.find('.select-chat-2').trigger('click');
    await flushPromises();

    resolveFirstDetail?.({
      data: {
        data: {
          chat_type: 'direct_listing_chat',
          participants: [{ user_id: 'member-self', public_id: 'member-self' }, { user_id: 'member-peer', public_id: 'member-peer' }],
        },
      },
    });
    resolveFirstMessages?.({
      data: {
        data: {
          items: [{ message_id: 'message-1', sender_user_id: 'member-peer', content: '第一個會話', message_type: 'text', status: 'sent', created_at: '2026-08-27T12:00:00Z' }],
        },
      },
    });
    await flushPromises();

    expect(wrapper.find('.selected-chat').text()).toBe('chat-2');
    expect(wrapper.find('.message-body').text()).toBe('第二個會話');
  });

  it('recovers a message that arrives after an empty history load and before websocket readiness', async () => {
    mocks.fetchMessages
      .mockResolvedValueOnce({ data: { data: { items: [] } } })
      .mockResolvedValueOnce({
        data: {
          data: {
            items: [{ message_id: 'message-recovered', sender_user_id: 'member-peer', content: '補回訊息', message_type: 'text', status: 'sent', created_at: '2026-08-27T13:00:00Z' }],
          },
        },
      });

    class MockWebSocket {
      static instance: MockWebSocket | null = null;

      onopen: (() => void) | null = null;
      onmessage: ((event: MessageEvent<string>) => void) | null = null;
      onerror: (() => void) | null = null;
      onclose: (() => void) | null = null;
      close = vi.fn();

      constructor() {
        MockWebSocket.instance = this;
      }
    }
    vi.stubGlobal('WebSocket', MockWebSocket);

    const wrapper = mount(TestHost);
    await flushPromises();
    await wrapper.find('.select-chat').trigger('click');
    await flushPromises();
    MockWebSocket.instance?.onopen?.();
    await flushPromises();

    expect(mocks.fetchMessages).toHaveBeenCalledTimes(2);
    expect(mocks.fetchMessages).toHaveBeenLastCalledWith('chat-1', { latest: true, page_size: 100 });
    expect(wrapper.find('.message-body').text()).toBe('補回訊息');
  });

  it('continues cursor recovery until the final partial page', async () => {
    const baseMessage = {
      message_id: 'message-base',
      sender_user_id: 'member-peer',
      content: '原有訊息',
      message_type: 'text',
      status: 'sent',
      created_at: '2026-08-27T14:00:00Z',
    };
    const firstRecoveredPage = Array.from({ length: 100 }, (_, index) => ({
      ...baseMessage,
      message_id: `message-${index + 1}`,
      content: `補回 ${index + 1}`,
    }));
    mocks.fetchMessages
      .mockResolvedValueOnce({ data: { data: { items: [baseMessage] } } })
      .mockResolvedValueOnce({ data: { data: { items: firstRecoveredPage } } })
      .mockResolvedValueOnce({
        data: {
          data: {
            items: [{ ...baseMessage, message_id: 'message-101', content: '補回 101' }],
          },
        },
      });

    class MockWebSocket {
      static instance: MockWebSocket | null = null;

      onopen: (() => void) | null = null;
      onmessage: ((event: MessageEvent<string>) => void) | null = null;
      onerror: (() => void) | null = null;
      onclose: (() => void) | null = null;
      close = vi.fn();

      constructor() {
        MockWebSocket.instance = this;
      }
    }
    vi.stubGlobal('WebSocket', MockWebSocket);

    const wrapper = mount(TestHost);
    await flushPromises();
    await wrapper.find('.select-chat').trigger('click');
    await flushPromises();
    MockWebSocket.instance?.onopen?.();
    await flushPromises();

    expect(mocks.fetchMessages).toHaveBeenNthCalledWith(2, 'chat-1', {
      after_message_id: 'message-base',
      page_size: 100,
    });
    expect(mocks.fetchMessages).toHaveBeenNthCalledWith(3, 'chat-1', {
      after_message_id: 'message-100',
      page_size: 100,
    });
    expect(wrapper.find('.message-count').text()).toBe('102');
  });

  it('replaces an existing message when a realtime attachment-status update arrives', async () => {
    const initialMessage = {
      message_id: 'message-1',
      sender_user_id: 'member-peer',
      content: '',
      message_type: 'attachment',
      status: 'sent',
      created_at: '2026-08-27T10:00:00Z',
      attachments: [{
        media_asset_id: 'asset-1',
        mime_type: 'image/png',
        file_size: 10,
        url: '',
        processing_status: 'ready',
        scan_status: 'pending',
      }],
    };
    mocks.fetchMessages.mockResolvedValue({ data: { data: { items: [initialMessage] } } });

    class MockWebSocket {
      static instance: MockWebSocket | null = null;

      onopen: (() => void) | null = null;
      onmessage: ((event: MessageEvent<string>) => void) | null = null;
      onerror: (() => void) | null = null;
      onclose: (() => void) | null = null;
      close = vi.fn();

      constructor() {
        MockWebSocket.instance = this;
      }
    }
    vi.stubGlobal('WebSocket', MockWebSocket);

    const wrapper = mount(TestHost);
    await flushPromises();
    await wrapper.find('.select-chat').trigger('click');
    await flushPromises();

    MockWebSocket.instance?.onmessage?.(new MessageEvent('message', {
      data: JSON.stringify({
        type: 'message',
        chat_id: 'chat-1',
        message: {
          ...initialMessage,
          attachments: [{
            ...initialMessage.attachments[0],
            url: 'https://media.example/asset-1.png',
            scan_status: 'passed',
          }],
        },
      }),
    }));
    await flushPromises();

    expect(wrapper.find('.attachment-status').text()).toBe('passed');
    expect(wrapper.find('.last-message').text()).toBe('chat.attachmentMessagePreview');
  });

  it('updates the conversation preview after an attachment-only message is sent', async () => {
    mocks.createUploadPresign.mockResolvedValue({
      data: {
        data: {
          object_key: 'chat/asset-1.png',
          upload_token: 'upload-token',
          upload_url: 'https://upload.example/asset-1.png',
          headers: {},
        },
      },
    });
    mocks.completeUpload.mockResolvedValue({ data: { data: { media_asset_id: 'asset-1' } } });
    mocks.sendMessage.mockResolvedValue({
      data: {
        data: {
          message_id: 'message-attachment-1',
          sender_user_id: 'member-self',
          content: '',
          message_type: 'attachment',
          status: 'sent',
          created_at: '2026-08-27T12:00:00Z',
          attachments: [],
        },
      },
    });
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true }));

    const wrapper = mount(TestHost);
    await flushPromises();
    await wrapper.find('.select-chat').trigger('click');
    await flushPromises();

    const page = wrapper.vm as unknown as {
      handleSelectAttachment: (event: Event) => Promise<void>;
    };
    await page.handleSelectAttachment({
      target: { files: [new File(['image'], 'asset.png', { type: 'image/png' })], value: '' },
    } as unknown as Event);
    await flushPromises();

    expect(mocks.sendMessage).toHaveBeenCalledWith('chat-1', expect.objectContaining({
      attachment_ids: ['asset-1'],
      content: '',
    }));
    expect(wrapper.find('.last-message').text()).toBe('chat.attachmentMessagePreview');
  });

  it('rejects unsupported and oversized attachment selections before upload', async () => {
    const wrapper = mount(TestHost);
    await flushPromises();
    await wrapper.find('.select-chat').trigger('click');
    await flushPromises();

    const page = wrapper.vm as unknown as {
      handleSelectAttachment: (event: Event) => Promise<void>;
    };
    await page.handleSelectAttachment({
      target: { files: [new File(['binary'], 'asset.zip', { type: 'application/zip' })], value: '' },
    } as unknown as Event);
    await page.handleSelectAttachment({
      target: { files: [new File([new Uint8Array(50 * 1024 * 1024 + 1)], 'asset.png', { type: 'image/png' })], value: '' },
    } as unknown as Event);

    expect(mocks.createUploadPresign).not.toHaveBeenCalled();
    expect(mocks.pushToast).toHaveBeenNthCalledWith(1, 'chat.attachmentTypeError', 'error');
    expect(mocks.pushToast).toHaveBeenNthCalledWith(2, 'chat.attachmentSizeError', 'error');
  });
});
