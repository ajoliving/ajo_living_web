/*
 * 大廈群聊管理頁測試。
 * 1. 驗證選擇群組後載入成員與待審申請。
 * 2. 驗證批准申請後重新載入管理資料。
 * 3. 驗證加入成員成功後只更新該列，不重載候選列表。
 */
import { flushPromises, mount } from '@vue/test-utils';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import {
  addStaffBuildingChatMember,
  fetchBuildingChatJoinRequests,
  fetchBuildingChatMembers,
  fetchStaffBuildingChats,
  reviewBuildingChatJoin,
} from '@/httpapis/chats';
import { fetchStaffUsers } from '@/httpapis/staff';

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
  addStaffBuildingChatMember: vi.fn(),
}));

vi.mock('@/httpapis/staff', () => ({
  fetchStaffUsers: vi.fn(),
}));

const mockedFetchStaffBuildingChats = vi.mocked(fetchStaffBuildingChats);
const mockedFetchBuildingChatMembers = vi.mocked(fetchBuildingChatMembers);
const mockedFetchBuildingChatJoinRequests = vi.mocked(fetchBuildingChatJoinRequests);
const mockedReviewBuildingChatJoin = vi.mocked(reviewBuildingChatJoin);
const mockedFetchStaffUsers = vi.mocked(fetchStaffUsers);
const mockedAddStaffBuildingChatMember = vi.mocked(addStaffBuildingChatMember);

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
    mockedFetchStaffUsers.mockReset();
    mockedAddStaffBuildingChatMember.mockReset();
    mockedFetchStaffBuildingChats.mockResolvedValue(createStaffChatsResponse());
    mockedFetchBuildingChatMembers.mockResolvedValue(createMembersResponse());
    mockedFetchBuildingChatJoinRequests.mockResolvedValue(createJoinRequestsResponse());
    mockedReviewBuildingChatJoin.mockResolvedValue({
      data: { data: { request_id: 'join-request-1' } },
    } as Awaited<ReturnType<typeof reviewBuildingChatJoin>>);
    mockedFetchStaffUsers.mockResolvedValue({
      data: {
        data: {
          items: [{
            public_id: 'member-103',
            email: 'new@example.com',
            phone_country_code: '+852',
            phone_number: '61234567',
            member_status: 'active',
            member_type: 'user',
            is_staff: false,
            role: 'user',
            roles: ['user'],
            permissions: [],
            display_name: '住戶丙',
            publisher_identity_type: 'owner',
            district_code: '',
            created_at: '2026-08-26T10:00:00Z',
            updated_at: '2026-08-26T10:00:00Z',
          }, {
            public_id: 'member-104',
            email: 'next@example.com',
            phone_country_code: '+852',
            phone_number: '69876543',
            member_status: 'active',
            member_type: 'user',
            is_staff: false,
            role: 'user',
            roles: ['user'],
            permissions: [],
            display_name: '住戶丁',
            publisher_identity_type: 'owner',
            district_code: '',
            created_at: '2026-08-26T11:00:00Z',
            updated_at: '2026-08-26T11:00:00Z',
          }],
          pagination: { page: 1, page_size: 10, total: 2 },
        },
      },
    } as unknown as Awaited<ReturnType<typeof fetchStaffUsers>>);
    mockedAddStaffBuildingChatMember.mockResolvedValue({
      data: { data: { user_id: '103', public_id: 'member-103' } },
    } as Awaited<ReturnType<typeof addStaffBuildingChatMember>>);
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

  it('adds a recorded member from the search dialog', async () => {
    const host = document.createElement('div');
    document.body.appendChild(host);
    const wrapper = mount(Page, { attachTo: host });
    await flushPromises();

    await findActionButton(wrapper, 'marketplace.management.buildingChats.viewMembers').trigger('click');
    await flushPromises();
    await findActionButton(wrapper, 'marketplace.management.buildingChats.addMember').trigger('click');
    await flushPromises();

    expect(mockedFetchStaffUsers).toHaveBeenCalled();
    expect(document.body.textContent).toContain('住戶丙');

    const addButton = Array.from(document.body.querySelectorAll('button')).find(
      (item) => item.textContent === 'marketplace.management.buildingChats.addToGroup',
    );
    if (!addButton) {
      throw new Error('add button not found');
    }
    addButton.dispatchEvent(new MouseEvent('click', { bubbles: true }));
    await flushPromises();

    expect(mockedAddStaffBuildingChatMember).toHaveBeenCalledWith(chat.chat_id, { public_id: 'member-103' });
    expect(mockedFetchStaffUsers).toHaveBeenCalledTimes(1);
    expect(mockedFetchBuildingChatMembers).toHaveBeenCalledTimes(1);
    expect(document.body.textContent).toContain('marketplace.management.buildingChats.alreadyInGroup');
    expect(document.body.textContent).toContain('住戶丁');
    expect(Array.from(document.body.querySelectorAll('button')).some(
      (item) => item.textContent === 'marketplace.management.buildingChats.addToGroup',
    )).toBe(true);
    wrapper.unmount();
    host.remove();
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
