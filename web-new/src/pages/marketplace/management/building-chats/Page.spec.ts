/*
 * 大廈群聊管理頁測試。
 * 1. 驗證選擇群組後載入成員與待審申請。
 * 2. 驗證批准申請後重新載入管理資料。
 */
import { flushPromises, mount } from '@vue/test-utils';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import {
  fetchBuildingChatJoinRequests,
  fetchBuildingChatMembers,
  fetchStaffBuildingChats,
  reviewBuildingChatJoin,
} from '@/httpapis/chats';

import Page from './Page.vue';

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
    locale: { value: 'zh-HK' },
  }),
}));

vi.mock('@/stores/feedback', () => ({
  useFeedbackStore: () => ({ pushToast: vi.fn() }),
}));

vi.mock('@/httpapis/chats', () => ({
  fetchStaffBuildingChats: vi.fn(),
  fetchBuildingChatMembers: vi.fn(),
  fetchBuildingChatJoinRequests: vi.fn(),
  moderateBuildingChatMember: vi.fn(),
  reviewBuildingChatJoin: vi.fn(),
}));

const mockedFetchStaffBuildingChats = vi.mocked(fetchStaffBuildingChats);
const mockedFetchBuildingChatMembers = vi.mocked(fetchBuildingChatMembers);
const mockedFetchBuildingChatJoinRequests = vi.mocked(fetchBuildingChatJoinRequests);
const mockedReviewBuildingChatJoin = vi.mocked(reviewBuildingChatJoin);

const chat = {
  chat_id: 'building-chat-1',
  building_id: 'BLG-001',
  chat_type: 'building_group' as const,
  member_count: 1,
  unread_count: 0,
  last_message_preview: '',
  last_message_at: null,
};

// 1. 建立群聊管理 API 回應。
const createStaffChatsResponse = () => ({
  data: { data: { items: [chat] } },
}) as unknown as Awaited<ReturnType<typeof fetchStaffBuildingChats>>;

const createMembersResponse = () => ({
  data: {
    data: {
      items: [{
        user_id: '101',
        public_id: 'member-101',
        display_name: '住戶甲',
        role_in_chat: 'member',
        membership_status: 'active',
        joined_at: '2026-08-26T08:00:00Z',
      }],
    },
  },
}) as unknown as Awaited<ReturnType<typeof fetchBuildingChatMembers>>;

const createJoinRequestsResponse = () => ({
  data: {
    data: {
      items: [{
        request_id: 'join-request-1',
        chat_id: chat.chat_id,
        user_id: '102',
        user_public_id: 'member-102',
        display_name: '住戶乙',
        status: 'pending',
        reason: '申請加入',
        created_at: '2026-08-26T09:00:00Z',
      }],
    },
  },
}) as unknown as Awaited<ReturnType<typeof fetchBuildingChatJoinRequests>>;

// 2. 尋找具有指定文案的操作按鈕。
const findActionButton = (wrapper: ReturnType<typeof mount>, label: string) => {
  const button = wrapper.findAll('button').find((item) => item.text() === label);

  if (!button) {
    throw new Error(`button not found: ${label}`);
  }

  return button;
};

describe('BuildingChatsManagementPage', () => {
  beforeEach(() => {
    mockedFetchStaffBuildingChats.mockReset();
    mockedFetchBuildingChatMembers.mockReset();
    mockedFetchBuildingChatJoinRequests.mockReset();
    mockedReviewBuildingChatJoin.mockReset();
    mockedFetchStaffBuildingChats.mockResolvedValue(createStaffChatsResponse());
    mockedFetchBuildingChatMembers.mockResolvedValue(createMembersResponse());
    mockedFetchBuildingChatJoinRequests.mockResolvedValue(createJoinRequestsResponse());
    mockedReviewBuildingChatJoin.mockResolvedValue({
      data: { data: { request_id: 'join-request-1' } },
    } as Awaited<ReturnType<typeof reviewBuildingChatJoin>>);
  });

  it('loads members and pending requests before approving a request', async () => {
    const wrapper = mount(Page);
    await flushPromises();

    await findActionButton(wrapper, 'marketplace.management.buildingChats.viewMembers').trigger('click');
    await flushPromises();

    expect(mockedFetchBuildingChatMembers).toHaveBeenCalledWith(chat.chat_id);
    expect(mockedFetchBuildingChatJoinRequests).toHaveBeenCalledWith(chat.chat_id);
    expect(wrapper.text()).toContain('住戶甲');
    expect(wrapper.text()).toContain('住戶乙');

    await findActionButton(wrapper, 'marketplace.management.buildingChats.approve').trigger('click');
    await flushPromises();

    expect(mockedReviewBuildingChatJoin).toHaveBeenCalledWith('join-request-1', 'approved');
    expect(mockedFetchBuildingChatMembers).toHaveBeenCalledTimes(2);
    expect(mockedFetchBuildingChatJoinRequests).toHaveBeenCalledTimes(2);
  });

  it('keeps the page stable when the chat list response has no items array', async () => {
    mockedFetchStaffBuildingChats.mockResolvedValue({
      data: { data: {} },
    } as Awaited<ReturnType<typeof fetchStaffBuildingChats>>);

    const wrapper = mount(Page);
    await flushPromises();

    expect(wrapper.text()).toContain('marketplace.management.buildingChats.emptyTitle');
  });
});
