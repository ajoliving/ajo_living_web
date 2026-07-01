/*
 * 聊天頁 - 狀態與資料流程。
 * 1. 讀取會話列表、會話詳情與訊息。
 * 2. 管理會話切換、訊息送出與已讀同步。
 */
import axios from 'axios';
import { computed, nextTick, onMounted, ref, watch } from 'vue';
import type { ComponentPublicInstance } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import type { ChatSummaryResponse } from '@/httpapis/chats';
import { fetchChatDetail, fetchChats, markChatRead } from '@/httpapis/chats';
import type { MessageResponse } from '@/httpapis/messages';
import { fetchMessages, sendMessage } from '@/httpapis/messages';
import type { ChatConversationView, ChatMessageView } from '@/model/chat';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';
import { useSessionStore } from '@/stores/session';
import { formatPrice } from '@/utils/format';

const messageMaxLength = 1000;
const directListingChatType = 'direct_listing_chat';
const hiddenMessageTypes = new Set(['notice_card']);

// 1. 只保留真實聊天會話
const isDisplayableConversation = (item: ChatSummaryResponse): boolean =>
  String(item.chat_type) === directListingChatType;

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
  const selectedChatId = ref('');
  const activeMessages = ref<ChatMessageView[]>([]);
  const participantPublicIdByUserId = ref<Record<string, string>>({});
  const draftMessage = ref('');
  const messageContainerRef = ref<HTMLDivElement | null>(null);

  const activeConversation = computed(() =>
    conversations.value.find((conversation) => conversation.id === selectedChatId.value),
  );
  const activeReferencePrice = computed(() => Number(activeConversation.value?.listing.price_hkd || 0));
  const canSendMessage = computed(() =>
    Boolean(
      selectedChatId.value &&
      activeConversation.value &&
      draftMessage.value.trim().length > 0 &&
      !sendingMessage.value,
    ),
  );

  // 3.1 映射會話資料
  const mapConversation = (item: ChatSummaryResponse): ChatConversationView => ({
    id: item.chat_id,
    type: directListingChatType,
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
      const { data } = await fetchChats({ page: 1, page_size: 50 });
      conversations.value = data.data.items.filter(isDisplayableConversation).map(mapConversation);

      const deepLinkedChatId = String(route.params.conversationId || '').trim();
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
        await router.replace(`/account/chat/${targetConversation.id}`);
      } else if (!selectedChatId.value && conversations.value[0]) {
        selectedChatId.value = conversations.value[0].id;
      } else if (deepLinkedChatId) {
        selectedChatId.value = '';
        await router.replace('/account/chat');
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
      return;
    }
    const selectedConversation = conversations.value.find((item) => item.id === selectedChatId.value);
    if (!selectedConversation) {
      activeMessages.value = [];
      participantPublicIdByUserId.value = {};
      selectedChatId.value = conversations.value[0]?.id || '';
      await router.replace(selectedChatId.value ? `/account/chat/${selectedChatId.value}` : '/account/chat');
      return;
    }

    loadingMessages.value = true;

    try {
      const [{ data: chatData }, { data: messageData }] = await Promise.all([
        fetchChatDetail(selectedChatId.value),
        fetchMessages(selectedChatId.value, { page: 1, page_size: 100 }),
      ]);
      if (String(chatData.data.chat_type) !== directListingChatType) {
        conversations.value = conversations.value.filter((item) => item.id !== selectedChatId.value);
        activeMessages.value = [];
        participantPublicIdByUserId.value = {};
        selectedChatId.value = conversations.value[0]?.id || '';
        await router.replace(selectedChatId.value ? `/account/chat/${selectedChatId.value}` : '/account/chat');
        return;
      }

      participantPublicIdByUserId.value = Object.fromEntries(
        chatData.data.participants.map((participant) => [participant.user_id, participant.public_id || '']),
      );
      activeMessages.value = messageData.data.items.filter(isDisplayableMessage).map(mapMessage);

      const matchedConversation = conversations.value.find((item) => item.id === selectedChatId.value);
      if (matchedConversation?.unread_count) {
        await markChatRead(selectedChatId.value);
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
    selectedChatId.value = chatId;
    await router.replace(`/account/chat/${chatId}`);
  };

  // 3.6 送出訊息
  const handleSendMessage = async (): Promise<void> => {
    const content = draftMessage.value.trim();
    if (!content || !selectedChatId.value) {
      return;
    }

    sendingMessage.value = true;

    try {
      const { data } = await sendMessage(selectedChatId.value, { content });
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

  // 3.7 使用 Enter 送出訊息，Shift + Enter 保留換行
  const handleSendByEnter = (event: KeyboardEvent): void => {
    if (event.shiftKey || event.isComposing || !canSendMessage.value) {
      return;
    }

    event.preventDefault();
    void handleSendMessage();
  };

  // 3.8 綁定訊息滾動容器
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
    () => route.params.conversationId,
    (conversationId) => {
      const nextId = String(conversationId || '').trim();
      if (nextId && nextId !== selectedChatId.value) {
        if (conversations.value.some((conversation) => conversation.id === nextId)) {
          selectedChatId.value = nextId;
          return;
        }

        void router.replace('/account/chat');
      }
    },
  );

  onMounted(() => {
    void loadChats();
  });

  return {
    activeConversation,
    activeReferencePrice,
    activeMessages,
    canSendMessage,
    conversations,
    draftMessage,
    formatPrice,
    handleSelectChat,
    handleSendByEnter,
    handleSendMessage,
    loadingConversations,
    loadingMessages,
    messageMaxLength,
    preferenceStore,
    selectedChatId,
    setMessageContainerRef,
    sendingMessage,
    sessionStore,
    t,
  };
};
