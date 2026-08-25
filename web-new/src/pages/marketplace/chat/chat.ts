/*
 * 聊天頁 - 狀態與資料流程。
 * 1. 讀取會話列表、會話詳情與訊息。
 * 2. 管理會話列表與聊天內容的單頁切換、訊息送出與已讀同步。
 */
import axios from 'axios';
import { computed, nextTick, onMounted, ref, watch } from 'vue';
import type { ComponentPublicInstance } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import type { BuildingChatSummaryResponse, ChatSummaryResponse } from '@/httpapis/chats';
import {
  fetchBuildingChatDetail,
  fetchBuildingChatMessages,
  fetchBuildingChatMembers,
  fetchBuildingChatJoinRequests,
  fetchBuildingChats,
  fetchChatDetail,
  fetchChats,
  markBuildingChatRead,
  markChatRead,
  leaveBuildingChat,
  moderateBuildingChatMember,
  reviewBuildingChatJoin,
  sendBuildingChatMessage,
} from '@/httpapis/chats';
import type { MessageResponse } from '@/httpapis/messages';
import { fetchMessages, sendMessage } from '@/httpapis/messages';
import type { BuildingChatJoinRequestView, BuildingChatMemberView, ChatConversationView, ChatMessageView } from '@/model/chat';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';
import { useSessionStore } from '@/stores/session';
import { formatPrice } from '@/utils/format';

const messageMaxLength = 1000;
const directListingChatType = 'direct_listing_chat';
const buildingGroupChatType = 'building_group';
const hiddenMessageTypes = new Set(['notice_card']);
export type ChatListFilter = 'all' | 'direct' | 'building_group';

// 1. 只保留真實聊天會話
const isDisplayableConversation = (item: ChatSummaryResponse): boolean =>
  String(item.chat_type) === directListingChatType;

// 2. 只保留大廈群聊摘要
const isDisplayableBuildingConversation = (item: BuildingChatSummaryResponse): boolean =>
  String(item.chat_type) === buildingGroupChatType;

// 2. 過濾歷史通知卡片訊息
const isDisplayableMessage = (item: MessageResponse): boolean =>
  !hiddenMessageTypes.has(String(item.message_type));

// 3. 管理聊天頁資料與動作
export const useMarketplaceChatPage = () => {
  const route = useRoute();
  const router = useRouter();
  const { t } = useI18n();
  const preferenceStore = usePreferenceStore();
  const sessionStore = useSessionStore();
  const feedbackStore = useFeedbackStore();
  const loadingConversations = ref(false);
  const loadingMessages = ref(false);
  const sendingMessage = ref(false);
  const conversations = ref<ChatConversationView[]>([]);
  const conversationFilter = ref<ChatListFilter>(
    String(route.query.chatType || '') === buildingGroupChatType ? 'building_group' : 'all',
  );
  const showBuildingGroups = ref(conversationFilter.value === 'building_group');
  const selectedChatId = ref('');
  const activeMessages = ref<ChatMessageView[]>([]);
  const activeMembers = ref<BuildingChatMemberView[]>([]);
  const activeJoinRequests = ref<BuildingChatJoinRequestView[]>([]);
  const canManageGroup = ref(false);
  const participantPublicIdByUserId = ref<Record<string, string>>({});
  const draftMessage = ref('');
  const messageContainerRef = ref<HTMLDivElement | null>(null);

  const activeConversation = computed(() =>
    conversations.value.find((conversation) => conversation.id === selectedChatId.value),
  );
  const directConversations = computed(() =>
    conversations.value.filter((conversation) => conversation.type === directListingChatType),
  );
  const buildingConversations = computed(() =>
    conversations.value.filter((conversation) => conversation.type === buildingGroupChatType),
  );
  const filteredConversations = computed(() => {
    if (conversationFilter.value === 'building_group') {
      return buildingConversations.value;
    }
    if (conversationFilter.value === 'direct') {
      return directConversations.value;
    }
    return showBuildingGroups.value
      ? [...buildingConversations.value, ...directConversations.value]
      : directConversations.value;
  });
  const activeReferencePrice = computed(() => Number(activeConversation.value?.listing.price_hkd || 0));
  const canSendMessage = computed(() =>
    Boolean(
      selectedChatId.value &&
      activeConversation.value &&
      draftMessage.value.trim().length > 0 &&
      !sendingMessage.value,
    ),
  );

  // 3.1 切換會話列表篩選
  const setConversationFilter = (filter: ChatListFilter): void => {
    conversationFilter.value = filter;
    showBuildingGroups.value = filter === 'building_group';
  };

  // 3.2 展開或收起大廈群聊
  const toggleBuildingGroups = async (): Promise<void> => {
    showBuildingGroups.value = !showBuildingGroups.value;
    if (conversationFilter.value !== 'building_group' || !showBuildingGroups.value) {
      conversationFilter.value = 'all';
    }
    await router.replace({
      path: '/notifications',
      query: conversationFilter.value === 'building_group' && showBuildingGroups.value
        ? { tab: 'conversations', chatType: buildingGroupChatType }
        : { tab: 'conversations' },
    });
  };

  // 3.1 映射私聊資料
  const mapConversation = (item: ChatSummaryResponse): ChatConversationView => ({
    id: item.chat_id,
    type: directListingChatType,
    title: item.peer?.display_name || t('chat.memberFallback'),
    subtitle: item.listing?.title || item.listing_title,
    listing: {
      id: item.listing?.listing_id || item.listing_id,
      title: item.listing?.title || item.listing_title,
      summary: item.listing?.summary || '',
      price_hkd: item.listing?.price_hkd ?? 0,
      status: item.listing?.business_status || 'available',
      published_at: item.listing?.published_at,
      cover_image_url: item.listing?.cover_image?.url || item.cover_image?.url,
    },
    peer: {
      user_id: item.peer?.user_id || '',
      public_id: item.peer?.public_id || '',
      display_name: item.peer?.display_name || t('chat.memberFallback'),
      avatar_url: '',
      role_in_chat: item.peer?.role_in_chat || '',
    },
    unread_count: item.unread_count,
    last_message: item.last_message_preview,
    last_message_at: item.last_message_at || '',
  });

  // 3.2 映射大廈群聊資料
  const mapBuildingConversation = (item: BuildingChatSummaryResponse): ChatConversationView => ({
    id: item.chat_id,
    type: buildingGroupChatType,
    title: `${t('chat.buildingGroup')} ${item.building_id}`,
    subtitle: `${item.member_count} ${t('chat.members')}`,
    building_id: item.building_id,
    member_count: item.member_count,
    listing: {
      id: item.building_id,
      title: `${t('chat.buildingGroup')} ${item.building_id}`,
      summary: '',
      price_hkd: 0,
      status: 'active',
    },
    peer: {
      user_id: '',
      public_id: '',
      display_name: `${t('chat.buildingGroup')} ${item.building_id}`,
      avatar_url: '',
      role_in_chat: 'group',
    },
    unread_count: item.unread_count,
    last_message: item.last_message_preview,
    last_message_at: item.last_message_at || '',
  });

  // 3.2 根據參與者資料映射訊息
  const mapMessage = (item: MessageResponse): ChatMessageView => ({
    id: item.message_id,
    chat_id: selectedChatId.value,
    sender_role:
      participantPublicIdByUserId.value[item.sender_user_id] === sessionStore.currentUser.public_id
        ? 'self'
        : 'peer',
    body: item.content,
    message_type: item.message_type,
    action_label: item.action_label,
    action_url: item.action_url,
    sent_at: item.created_at,
  });

  // 3.3 讀取會話列表
  const loadChats = async (): Promise<void> => {
    loadingConversations.value = true;

    try {
      const { data: directData } = await fetchChats({ page: 1, page_size: 50 });
      let buildingItems: BuildingChatSummaryResponse[] = [];
      try {
        const buildingChatsLoader = fetchBuildingChats;
        if (typeof buildingChatsLoader === 'function') {
          const { data: buildingData } = await buildingChatsLoader();
          buildingItems = buildingData.data.items;
        }
      } catch {
        buildingItems = [];
      }
      conversations.value = [
        ...directData.data.items.filter(isDisplayableConversation).map(mapConversation),
        ...buildingItems.filter(isDisplayableBuildingConversation).map(mapBuildingConversation),
      ];

      const deepLinkedChatId = String(route.params.conversationId || route.query.conversationId || '').trim();
      const targetConversationType = String(route.query.target || '').trim();
      const targetConversation = conversations.value.find(
        (conversation) => conversation.type === targetConversationType,
      );
      const deepLinkedConversation = conversations.value.find(
        (conversation) => conversation.id === deepLinkedChatId,
      );
      if (deepLinkedConversation) {
        selectedChatId.value = deepLinkedChatId;
      } else if (targetConversation) {
        selectedChatId.value = targetConversation.id;
        await router.replace({ path: '/notifications', query: { tab: 'conversations', conversationId: targetConversation.id } });
      } else if (deepLinkedChatId) {
        selectedChatId.value = '';
        await router.replace({ path: '/notifications', query: { tab: 'conversations' } });
      }
    } catch (error) {
      feedbackStore.pushToast(
        axios.isAxiosError(error)
          ? error.response?.data?.message ?? t('chat.loadChatsError')
          : t('chat.loadChatsError'),
        'error',
      );
    } finally {
      loadingConversations.value = false;
    }
  };

  // 3.4 讀取目前會話資料
  const loadActiveChat = async (): Promise<void> => {
    if (!selectedChatId.value) {
      activeMessages.value = [];
      participantPublicIdByUserId.value = {};
      activeMembers.value = [];
      activeJoinRequests.value = [];
      canManageGroup.value = false;
      return;
    }
    const selectedConversation = conversations.value.find((item) => item.id === selectedChatId.value);
    if (!selectedConversation) {
      activeMessages.value = [];
      participantPublicIdByUserId.value = {};
      activeMembers.value = [];
      activeJoinRequests.value = [];
      canManageGroup.value = false;
      selectedChatId.value = '';
      await router.replace({ path: '/notifications', query: { tab: 'conversations' } });
      return;
    }

    loadingMessages.value = true;

    try {
      const isBuildingGroup = selectedConversation.type === buildingGroupChatType;
      const [{ data: chatData }, { data: messageData }] = isBuildingGroup
        ? await Promise.all([
            fetchBuildingChatDetail(selectedChatId.value),
            fetchBuildingChatMessages(selectedChatId.value, { page: 1, page_size: 100 }),
          ])
        : await Promise.all([
            fetchChatDetail(selectedChatId.value),
            fetchMessages(selectedChatId.value, { page: 1, page_size: 100 }),
          ]);

      participantPublicIdByUserId.value = Object.fromEntries(
        chatData.data.participants.map((participant) => [participant.user_id, participant.public_id || '']),
      );
      canManageGroup.value = isBuildingGroup && Boolean(chatData.data.can_manage);
      if (isBuildingGroup) {
        const { data: memberData } = await fetchBuildingChatMembers(selectedChatId.value);
        activeMembers.value = memberData.data.items.map((member) => ({
          user_id: member.user_id,
          display_name: member.display_name,
          role_in_chat: member.role_in_chat,
          membership_status: member.membership_status,
          muted_until: member.muted_until,
          banned_until: member.banned_until,
        }));
        if (canManageGroup.value) {
          const { data: requestData } = await fetchBuildingChatJoinRequests(selectedChatId.value);
          activeJoinRequests.value = requestData.data.items
            .filter((request) => request.status === 'pending')
            .map((request) => ({
              request_id: request.request_id,
              user_id: request.user_id,
              status: request.status,
              reason: request.reason,
              created_at: request.created_at,
            }));
        } else {
          activeJoinRequests.value = [];
        }
      } else {
        activeMembers.value = [];
        activeJoinRequests.value = [];
      }
      activeMessages.value = messageData.data.items.filter(isDisplayableMessage).map(mapMessage);

      const matchedConversation = conversations.value.find((item) => item.id === selectedChatId.value);
      if (matchedConversation?.unread_count) {
        if (isBuildingGroup) {
          await markBuildingChatRead(selectedChatId.value);
        } else {
          await markChatRead(selectedChatId.value);
        }
        conversations.value = conversations.value.map((item) =>
          item.id === selectedChatId.value ? { ...item, unread_count: 0 } : item,
        );
      }
    } catch (error) {
      feedbackStore.pushToast(
        axios.isAxiosError(error)
          ? error.response?.data?.message ?? t('chat.loadThreadError')
          : t('chat.loadThreadError'),
        'error',
      );
    } finally {
      loadingMessages.value = false;
    }
  };

  // 3.5 切換會話
  const handleSelectChat = async (chatId: string): Promise<void> => {
    if (!conversations.value.some((conversation) => conversation.id === chatId)) {
      await loadChats();
    }
    if (!conversations.value.some((conversation) => conversation.id === chatId)) {
      return;
    }
    selectedChatId.value = chatId;
    const selectedConversation = conversations.value.find((conversation) => conversation.id === chatId);
    if (selectedConversation?.type === buildingGroupChatType) {
      await router.replace({
        path: '/notifications',
        query: {
          tab: 'conversations',
          conversationId: chatId,
          chatType: buildingGroupChatType,
        },
      });
      return;
    }

    // 私聊沿用舊路徑，由路由守衛統一轉到通信中心。
    await router.replace(`/account/chat/${chatId}`);
  };

  // 3.6 返回會話列表
  const handleBackToChats = async (): Promise<void> => {
    const selectedConversation = conversations.value.find((conversation) => conversation.id === selectedChatId.value);
    selectedChatId.value = '';
    if (selectedConversation?.type === directListingChatType) {
      await router.replace('/account/chat');
      return;
    }

    await router.replace({
      path: '/notifications',
      query: conversationFilter.value === 'building_group'
        ? { tab: 'conversations', chatType: buildingGroupChatType }
        : { tab: 'conversations' },
    });
  };

  // 3.7 送出訊息
  const handleSendMessage = async (): Promise<void> => {
    const content = draftMessage.value.trim();
    if (!content || !selectedChatId.value) {
      return;
    }

    sendingMessage.value = true;

    try {
      const isBuildingGroup = activeConversation.value?.type === buildingGroupChatType;
      const { data } = isBuildingGroup
        ? await sendBuildingChatMessage(selectedChatId.value, { content })
        : await sendMessage(selectedChatId.value, { content });
      activeMessages.value = [...activeMessages.value, mapMessage(data.data)];
      conversations.value = conversations.value.map((conversation) =>
        conversation.id === selectedChatId.value
          ? {
              ...conversation,
              last_message: content,
              last_message_at: data.data.created_at,
              unread_count: 0,
            }
          : conversation,
      );
      draftMessage.value = '';
    } catch (error) {
      feedbackStore.pushToast(
        axios.isAxiosError(error)
          ? error.response?.data?.message ?? t('chat.sendError')
          : t('chat.sendError'),
        'error',
      );
    } finally {
      sendingMessage.value = false;
    }
  };

  // 3.8 使用 Enter 送出訊息，Shift + Enter 保留換行
  const handleSendByEnter = (event: KeyboardEvent): void => {
    if (event.shiftKey || event.isComposing || !canSendMessage.value) {
      return;
    }

    event.preventDefault();
    void handleSendMessage();
  };

  // 3.9 離開大廈群聊
  const handleLeaveBuildingChat = async (): Promise<void> => {
    if (!selectedChatId.value || activeConversation.value?.type !== buildingGroupChatType) {
      return;
    }
    try {
      await leaveBuildingChat(selectedChatId.value);
      await handleBackToChats();
      await loadChats();
    } catch (error) {
      feedbackStore.pushToast(
        axios.isAxiosError(error)
          ? error.response?.data?.message ?? t('chat.leaveError')
          : t('chat.leaveError'),
        'error',
      );
    }
  };

  // 3.10 執行大廈群聊成員管理
  const handleModerateBuildingMember = async (
    userId: string,
    action: 'mute' | 'unmute' | 'kick' | 'ban' | 'unban',
  ): Promise<void> => {
    if (!selectedChatId.value || !canManageGroup.value) {
      return;
    }
    try {
      await moderateBuildingChatMember(selectedChatId.value, {
        user_id: Number(userId),
        action,
        duration_minutes: action === 'mute' || action === 'ban' ? 60 : undefined,
      });
      const { data } = await fetchBuildingChatMembers(selectedChatId.value);
      activeMembers.value = data.data.items;
    } catch (error) {
      feedbackStore.pushToast(
        axios.isAxiosError(error)
          ? error.response?.data?.message ?? t('chat.moderationError')
          : t('chat.moderationError'),
        'error',
      );
    }
  };

  // 3.11 審核大廈群聊加入申請
  const handleReviewBuildingChatJoin = async (
    requestId: string,
    status: 'approved' | 'rejected',
  ): Promise<void> => {
    if (!selectedChatId.value || !canManageGroup.value) {
      return;
    }
    try {
      await reviewBuildingChatJoin(requestId, status);
      activeJoinRequests.value = activeJoinRequests.value.filter((request) => request.request_id !== requestId);
      if (status === 'approved') {
        const { data } = await fetchBuildingChatMembers(selectedChatId.value);
        activeMembers.value = data.data.items;
      }
    } catch (error) {
      feedbackStore.pushToast(
        axios.isAxiosError(error)
          ? error.response?.data?.message ?? t('chat.reviewError')
          : t('chat.reviewError'),
        'error',
      );
    }
  };

  // 3.12 綁定訊息滾動容器
  const setMessageContainerRef = (element: Element | ComponentPublicInstance | null): void => {
    messageContainerRef.value = element instanceof HTMLDivElement ? element : null;
  };

  watch(
    () => selectedChatId.value,
    () => {
      draftMessage.value = '';
      void loadActiveChat();
    },
  );

  watch(
    () => [selectedChatId.value, activeMessages.value.length],
    async () => {
      await nextTick();

      if (messageContainerRef.value) {
        messageContainerRef.value.scrollTop = messageContainerRef.value.scrollHeight;
      }
    },
    { immediate: true },
  );

  watch(
    () => route.params.conversationId || route.query.conversationId,
    (conversationId) => {
      const nextId = String(conversationId || '').trim();
      if (!nextId) {
        selectedChatId.value = '';
        return;
      }

      if (nextId && nextId !== selectedChatId.value) {
        if (conversations.value.some((conversation) => conversation.id === nextId)) {
          selectedChatId.value = nextId;
          return;
        }

        void router.replace({ path: '/notifications', query: { tab: 'conversations' } });
      }
    },
  );

  watch(
    () => route.query.chatType,
    (chatType) => {
      conversationFilter.value = String(chatType || '') === buildingGroupChatType ? 'building_group' : 'all';
      showBuildingGroups.value = conversationFilter.value === 'building_group';
    },
  );

  onMounted(() => {
    void loadChats();
  });

  return {
    activeConversation,
    activeReferencePrice,
    activeMessages,
    activeMembers,
    activeJoinRequests,
    buildingConversations,
    canSendMessage,
    canManageGroup,
    conversations,
    conversationFilter,
    directConversations,
    draftMessage,
    filteredConversations,
    formatPrice,
    handleBackToChats,
    handleSelectChat,
    handleSendByEnter,
    handleSendMessage,
    handleLeaveBuildingChat,
    handleModerateBuildingMember,
    handleReviewBuildingChatJoin,
    loadingConversations,
    loadingMessages,
    messageMaxLength,
    preferenceStore,
    selectedChatId,
    setConversationFilter,
    setMessageContainerRef,
    showBuildingGroups,
    toggleBuildingGroups,
    sendingMessage,
    sessionStore,
    t,
  };
};
