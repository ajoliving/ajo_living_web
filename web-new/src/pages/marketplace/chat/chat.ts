/*
 * 聊天頁 - 狀態與資料流程。
 * 1. 讀取會話列表、會話詳情與訊息，統一管理一對一與大廈群聊。
 * 2. 管理會話列表與聊天內容的單頁切換、訊息送出與已讀同步。
 * 3. 提供大廈群聊的只讀成員名單與離開群組流程。
 */
import axios from 'axios';
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import type { ComponentPublicInstance } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import type { BuildingChatSummaryResponse, ChatSummaryResponse } from '@/httpapis/chats';
import {
  fetchBuildingChatDetail,
  fetchBuildingChatMessages,
  fetchBuildingChatMembers,
  fetchBuildingChats,
  fetchChatDetail,
  fetchChats,
  leaveBuildingChat,
  markBuildingChatRead,
  markChatRead,
  sendBuildingChatMessage,
} from '@/httpapis/chats';
import type { MessageResponse } from '@/httpapis/messages';
import { fetchMessages, sendMessage } from '@/httpapis/messages';
import { fetchRealtimeTicket } from '@/httpapis/realtime';
import {
  abortMultipartUpload,
  completeMultipartUpload,
  completeUpload,
  createUploadPresign,
  initiateMultipartUpload,
  listMultipartParts,
  presignMultipartPart,
} from '@/httpapis/uploads';
import { buildUploadHeaders } from '@/utils/upload';
import type { BuildingChatMemberView, ChatConversationView, ChatMessageView } from '@/model/chat';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';
import { useSessionStore } from '@/stores/session';
import { formatPrice } from '@/utils/format';

const messageMaxLength = 1000;
const directListingChatType = 'direct_listing_chat';
const buildingGroupChatType = 'building_group';
const hiddenMessageTypes = new Set(['notice_card']);
const multipartThreshold = 10 * 1024 * 1024;
const chatAttachmentMaxCount = 5;
const chatAttachmentMaxFileSize = 50 * 1024 * 1024;

// 0. 建立一次發送的客戶端幂等鍵，網絡重試不會重複落庫。
const newClientMessageID = (): string => {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID();
  }
  return `chat-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;
};

// 0.1 檢查瀏覽器選取的檔案是否符合聊天附件類型。
const isSupportedChatAttachment = (file: File): boolean => {
  const mimeType = file.type.trim().toLowerCase();
  return mimeType.startsWith('image/')
    || mimeType.startsWith('video/')
    || mimeType === 'application/pdf'
    || mimeType === 'text/plain';
};

// 1. 只保留真實聊天會話
const isDisplayableConversation = (item: ChatSummaryResponse): boolean =>
  String(item.chat_type) === directListingChatType;

// 2. 只保留大廈群聊摘要
const isDisplayableBuildingConversation = (item: BuildingChatSummaryResponse): boolean =>
  String(item.chat_type) === buildingGroupChatType;

// 3. 過濾歷史通知卡片訊息
const isDisplayableMessage = (item: MessageResponse): boolean =>
  !hiddenMessageTypes.has(String(item.message_type));

// 4. 管理聊天頁資料與動作
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
  const uploadingAttachment = ref(false);
  const conversations = ref<ChatConversationView[]>([]);
  const selectedChatId = ref('');
  const activeMessages = ref<ChatMessageView[]>([]);
  const participantPublicIdByUserId = ref<Record<string, string>>({});
  const participantNameByUserId = ref<Record<string, string>>({});
  const draftMessage = ref('');
  const messageContainerRef = ref<HTMLDivElement | null>(null);
  const groupMembersOpen = ref(false);
  const loadingGroupMembers = ref(false);
  const groupMembers = ref<BuildingChatMemberView[]>([]);
  const leavingGroup = ref(false);
  const leaveConfirmOpen = ref(false);
  const realtimeSocket = ref<WebSocket | null>(null);
  let realtimeReconnectTimer: ReturnType<typeof window.setTimeout> | null = null;
  let realtimeReconnectAttempt = 0;

  const activeConversation = computed(() =>
    conversations.value.find((conversation) => conversation.id === selectedChatId.value),
  );
  const totalUnreadCount = computed(() =>
    conversations.value.reduce((total, conversation) => total + conversation.unread_count, 0),
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

  // 4.1 映射私聊資料
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

  // 4.2 映射大廈群聊資料
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

  // 4.3 根據參與者資料映射訊息
  const mapMessage = (item: MessageResponse): ChatMessageView => ({
    id: item.message_id,
    chat_id: selectedChatId.value,
    sender_role:
      participantPublicIdByUserId.value[item.sender_user_id] === sessionStore.currentUser.public_id
        ? 'self'
        : 'peer',
    sender_name: participantNameByUserId.value[item.sender_user_id] || '',
    body: item.content,
    message_type: item.message_type,
    action_label: item.action_label,
    action_url: item.action_url,
    sent_at: item.created_at,
    attachments: item.attachments || [],
  });

  // 4.4 以訊息 ID 寫入或更新目前會話，避免 REST 與 WebSocket 重複顯示同一條消息。
  const upsertActiveMessage = (message: ChatMessageView): void => {
    const existingMessageIndex = activeMessages.value.findIndex((item) => item.id === message.id);
    if (existingMessageIndex >= 0) {
      activeMessages.value = activeMessages.value.map((item, index) =>
        index === existingMessageIndex ? message : item,
      );
      return;
    }
    activeMessages.value = [...activeMessages.value, message];
  };

  // 4.4.1 直接或透過可續傳分片上傳一個聊天附件。
  const uploadChatAttachment = async (file: File): Promise<string> => {
    if (file.size <= multipartThreshold) {
      const { data } = await createUploadPresign({ file_name: file.name, mime_type: file.type, file_size: file.size, purpose: 'chat_attachment' });
      const presign = data.data;
      const uploadResponse = await fetch(presign.upload_url, { method: 'PUT', headers: buildUploadHeaders(presign.headers, file.type), body: file });
      if (!uploadResponse.ok) {
        throw new Error('upload failed');
      }
      const { data: completed } = await completeUpload({ object_key: presign.object_key, upload_token: presign.upload_token, mime_type: file.type, file_size: file.size });
      return completed.data.media_asset_id;
    }

    const { data: initiatedResponse } = await initiateMultipartUpload({ file_name: file.name, mime_type: file.type, file_size: file.size, purpose: 'chat_attachment' });
    const initiated = initiatedResponse.data;
    const uploadedParts: Array<{ part_number: number; etag: string }> = [];
    try {
      const { data: listedResponse } = await listMultipartParts({ object_key: initiated.object_key, upload_id: initiated.upload_id, upload_token: initiated.upload_token, mime_type: file.type, file_size: file.size });
      const existingParts = new Map(listedResponse.data.parts.map((part) => [part.part_number, part.etag]));
      for (let partNumber = 1; partNumber <= initiated.part_count; partNumber += 1) {
        const existingETag = existingParts.get(partNumber);
        if (existingETag) {
          uploadedParts.push({ part_number: partNumber, etag: existingETag });
          continue;
        }
        const start = (partNumber - 1) * initiated.part_size;
        const end = Math.min(start + initiated.part_size, file.size);
        let etag = '';
        for (let attempt = 1; attempt <= 3; attempt += 1) {
          const { data: partResponse } = await presignMultipartPart({ object_key: initiated.object_key, upload_id: initiated.upload_id, upload_token: initiated.upload_token, mime_type: file.type, file_size: file.size, part_number: partNumber });
          const part = partResponse.data;
          const uploadResponse = await fetch(part.upload_url, { method: 'PUT', headers: buildUploadHeaders(part.headers, file.type), body: file.slice(start, end) });
          if (uploadResponse.ok) {
            etag = uploadResponse.headers.get('ETag') || uploadResponse.headers.get('etag') || '';
            if (etag) break;
          }
          if (attempt === 3) throw new Error('multipart part upload failed');
        }
        uploadedParts.push({ part_number: partNumber, etag });
      }
      const { data: completed } = await completeMultipartUpload({ object_key: initiated.object_key, upload_id: initiated.upload_id, upload_token: initiated.upload_token, mime_type: file.type, file_size: file.size, parts: uploadedParts });
      return completed.data.media_asset_id;
    } catch (error) {
      await abortMultipartUpload({ object_key: initiated.object_key, upload_id: initiated.upload_id, upload_token: initiated.upload_token, mime_type: file.type, file_size: file.size }).catch(() => undefined);
      throw error;
    }
  };

  // 4.4.2 上傳並發送已選附件。
  const handleSelectAttachment = async (event: Event): Promise<void> => {
    const input = event.target as HTMLInputElement;
    const files = Array.from(input.files || []);
    input.value = '';
    if (!selectedChatId.value || files.length === 0 || uploadingAttachment.value) return;
    if (files.length > chatAttachmentMaxCount) {
      feedbackStore.pushToast(t('chat.attachmentLimitError'), 'error');
      return;
    }
    if (files.some((file) => file.size > chatAttachmentMaxFileSize)) {
      feedbackStore.pushToast(t('chat.attachmentSizeError'), 'error');
      return;
    }
    if (files.some((file) => !isSupportedChatAttachment(file))) {
      feedbackStore.pushToast(t('chat.attachmentTypeError'), 'error');
      return;
    }
    const chatId = selectedChatId.value;
    const content = draftMessage.value.trim();
    const isBuildingGroup = activeConversation.value?.type === buildingGroupChatType;
    uploadingAttachment.value = true;
    sendingMessage.value = true;
    try {
      const attachmentIDs = await Promise.all(files.map(uploadChatAttachment));
      const { data } = isBuildingGroup
        ? await sendBuildingChatMessage(chatId, { content, attachment_ids: attachmentIDs, client_message_id: newClientMessageID() })
        : await sendMessage(chatId, { content, attachment_ids: attachmentIDs, client_message_id: newClientMessageID() });
      if (selectedChatId.value === chatId) {
        upsertActiveMessage(mapMessage(data.data));
        if (draftMessage.value.trim() === content) {
          draftMessage.value = '';
        }
      }
      conversations.value = conversations.value.map((conversation) =>
        conversation.id === chatId
          ? {
              ...conversation,
              last_message: content || t('chat.attachmentMessagePreview'),
              last_message_at: data.data.created_at,
              unread_count: 0,
            }
          : conversation,
      );
    } catch (error) {
      feedbackStore.pushToast(
        axios.isAxiosError<{ message?: string }>(error)
          ? error.response?.data?.message ?? t('chat.sendError')
          : t('chat.sendError'),
        'error',
      );
    } finally {
      uploadingAttachment.value = false;
      sendingMessage.value = false;
    }
  };

  // 4.4 讀取會話列表
  const loadChats = async (): Promise<void> => {
    loadingConversations.value = true;

    try {
      const { data: directData } = await fetchChats({ page: 1, page_size: 50 });
      let buildingItems: BuildingChatSummaryResponse[] = [];
      try {
        const { data: buildingData } = await fetchBuildingChats();
        buildingItems = buildingData.data.items;
      } catch {
        buildingItems = [];
      }
      conversations.value = [
        ...directData.data.items.filter(isDisplayableConversation).map(mapConversation),
        ...buildingItems.filter(isDisplayableBuildingConversation).map(mapBuildingConversation),
      ];

      const deepLinkedChatId = String(route.params.conversationId || route.query.conversationId || '').trim();
      const deepLinkedConversation = conversations.value.find(
        (conversation) => conversation.id === deepLinkedChatId,
      );
      if (deepLinkedConversation) {
        selectedChatId.value = deepLinkedChatId;
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

  // 4.5 讀取目前會話資料
  const loadActiveChat = async (): Promise<void> => {
    const chatId = selectedChatId.value;
    if (!chatId) {
      activeMessages.value = [];
      participantPublicIdByUserId.value = {};
      participantNameByUserId.value = {};
      groupMembers.value = [];
      return;
    }
    const selectedConversation = conversations.value.find((item) => item.id === chatId);
    if (!selectedConversation) {
      activeMessages.value = [];
      participantPublicIdByUserId.value = {};
      participantNameByUserId.value = {};
      if (selectedChatId.value === chatId) {
        selectedChatId.value = '';
      }
      await router.replace({ path: '/notifications', query: { tab: 'conversations' } });
      return;
    }

    loadingMessages.value = true;

    try {
      const isBuildingGroup = selectedConversation.type === buildingGroupChatType;
      const [{ data: chatData }, { data: messageData }] = isBuildingGroup
        ? await Promise.all([
            fetchBuildingChatDetail(chatId),
            fetchBuildingChatMessages(chatId, { latest: true, page_size: 100 }),
          ])
        : await Promise.all([
            fetchChatDetail(chatId),
            fetchMessages(chatId, { latest: true, page_size: 100 }),
          ]);

      if (selectedChatId.value !== chatId) {
        return;
      }

      participantPublicIdByUserId.value = Object.fromEntries(
        chatData.data.participants.map((participant) => [participant.user_id, participant.public_id || '']),
      );
      participantNameByUserId.value = Object.fromEntries(
        chatData.data.participants.map((participant) => [participant.user_id, participant.display_name || '']),
      );
      activeMessages.value = messageData.data.items.filter(isDisplayableMessage).map(mapMessage);

      if (selectedConversation.unread_count) {
        if (isBuildingGroup) {
          await markBuildingChatRead(chatId);
        } else {
          await markChatRead(chatId);
        }
        conversations.value = conversations.value.map((item) =>
          item.id === chatId ? { ...item, unread_count: 0 } : item,
        );
      }
    } catch (error) {
      if (selectedChatId.value !== chatId) {
        return;
      }
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

  // 4.6 切換會話
  const handleSelectChat = async (chatId: string): Promise<void> => {
    if (!conversations.value.some((conversation) => conversation.id === chatId)) {
      await loadChats();
    }
    if (!conversations.value.some((conversation) => conversation.id === chatId)) {
      return;
    }
    selectedChatId.value = chatId;
    await router.replace({
      path: '/notifications',
      query: { tab: 'conversations', conversationId: chatId },
    });
  };

  // 4.7 返回會話列表
  const handleBackToChats = async (): Promise<void> => {
    selectedChatId.value = '';
    await router.replace({ path: '/notifications', query: { tab: 'conversations' } });
  };

  // 4.8 送出訊息
  const handleSendMessage = async (): Promise<void> => {
    const content = draftMessage.value.trim();
    if (!content || !selectedChatId.value) {
      return;
    }

    const chatId = selectedChatId.value;
    const isBuildingGroup = activeConversation.value?.type === buildingGroupChatType;
    sendingMessage.value = true;

    try {
      const { data } = isBuildingGroup
        ? await sendBuildingChatMessage(chatId, { content, client_message_id: newClientMessageID() })
        : await sendMessage(chatId, { content, client_message_id: newClientMessageID() });
      if (selectedChatId.value === chatId) {
        upsertActiveMessage(mapMessage(data.data));
      }
      conversations.value = conversations.value.map((conversation) =>
        conversation.id === chatId
          ? {
              ...conversation,
              last_message: content,
              last_message_at: data.data.created_at,
              unread_count: 0,
          }
          : conversation,
      );
      if (selectedChatId.value === chatId && draftMessage.value.trim() === content) {
        draftMessage.value = '';
      }
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

  // 4.9 使用 Enter 送出訊息，Shift + Enter 保留換行
  const handleSendByEnter = (event: KeyboardEvent): void => {
    if (event.shiftKey || event.isComposing || !canSendMessage.value) {
      return;
    }

    event.preventDefault();
    void handleSendMessage();
  };

  // 4.11 關閉目前即時連線並取消重連計時器
  const closeRealtime = (): void => {
    if (realtimeReconnectTimer !== null) {
      window.clearTimeout(realtimeReconnectTimer);
      realtimeReconnectTimer = null;
    }
    realtimeReconnectAttempt = 0;
    const socket = realtimeSocket.value;
    realtimeSocket.value = null;
    if (socket) {
      socket.onopen = null;
      socket.onmessage = null;
      socket.onerror = null;
      socket.onclose = null;
      socket.close();
    }
  };

  // 4.12 處理即時消息事件，並以消息 ID 去重
  const handleRealtimeMessage = (event: MessageEvent<string>): void => {
    let payload: {
      type?: string;
      chat_id?: string;
      message?: MessageResponse;
    };
    try {
      payload = JSON.parse(event.data) as typeof payload;
    } catch {
      return;
    }
    if (payload.type !== 'message' || payload.chat_id !== selectedChatId.value || !payload.message) {
      return;
    }
    const message = payload.message;
    upsertActiveMessage(mapMessage(message));
    conversations.value = conversations.value.map((conversation) =>
      conversation.id === selectedChatId.value
        ? {
            ...conversation,
            last_message:
              message.content ||
              (message.attachments?.length ? t('chat.attachmentMessagePreview') : conversation.last_message),
            last_message_at: message.created_at || conversation.last_message_at,
            unread_count: 0,
          }
        : conversation,
    );
  };

  // 4.13 連線恢復後按最後消息游標補回離線消息；空歷史會讀取目前最新一頁。
  const syncRealtimeMessages = async (chatId: string): Promise<void> => {
    const lastMessage = activeMessages.value[activeMessages.value.length - 1];
    if (selectedChatId.value !== chatId) {
      return;
    }
    const isBuildingGroup = activeConversation.value?.type === buildingGroupChatType;
    let afterMessageID = lastMessage?.id;
    try {
      while (selectedChatId.value === chatId) {
        const { data } = isBuildingGroup
          ? await fetchBuildingChatMessages(chatId, afterMessageID ? { after_message_id: afterMessageID, page_size: 100 } : { latest: true, page_size: 100 })
          : await fetchMessages(chatId, afterMessageID ? { after_message_id: afterMessageID, page_size: 100 } : { latest: true, page_size: 100 });
        if (selectedChatId.value !== chatId) {
          return;
        }
        const items = data.data.items;
        const knownIds = new Set(activeMessages.value.map((item) => item.id));
        const missing = items
          .filter((item) => !knownIds.has(item.message_id))
          .map(mapMessage);
        if (missing.length > 0) {
          activeMessages.value = [...activeMessages.value, ...missing];
          const newestMessage = missing[missing.length - 1];
          conversations.value = conversations.value.map((conversation) =>
            conversation.id === chatId
              ? {
                  ...conversation,
                  last_message: newestMessage.body || t('chat.attachmentMessagePreview'),
                  last_message_at: newestMessage.sent_at,
                  unread_count: 0,
                }
              : conversation,
          );
        }
        if (!afterMessageID || items.length < 100) {
          return;
        }
        afterMessageID = items[items.length - 1]?.message_id;
        if (!afterMessageID) {
          return;
        }
      }
    } catch {
      // 歷史接口仍可在下次連線恢復時補償，不阻斷目前的實時連線。
    }
  };

  // 4.14 建立目前會話的 WebSocket 連線，失敗時使用退避重連
  const connectRealtime = async (): Promise<void> => {
    const chatId = selectedChatId.value;
    if (!chatId || typeof window === 'undefined' || typeof WebSocket === 'undefined') {
      return;
    }
    try {
      const { data } = await fetchRealtimeTicket();
      if (selectedChatId.value !== chatId) {
        return;
      }
      const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const configuredBase = String(import.meta.env.VITE_WS_BASE_URL || '').trim().replace(/\/$/, '');
      const base = configuredBase || `${wsProtocol}//${window.location.host}/api/v1`;
      const socket = new WebSocket(
        `${base}/realtime/ws?ticket=${encodeURIComponent(data.data.ticket)}&chat_id=${encodeURIComponent(chatId)}`,
      );
      realtimeSocket.value = socket;
      socket.onopen = () => {
        realtimeReconnectAttempt = 0;
        void syncRealtimeMessages(chatId);
      };
      socket.onmessage = handleRealtimeMessage;
      socket.onerror = () => socket.close();
      socket.onclose = () => {
        if (realtimeSocket.value !== socket || selectedChatId.value !== chatId) {
          return;
        }
        realtimeSocket.value = null;
        const delay = Math.min(10000, 1000 * 2 ** realtimeReconnectAttempt);
        realtimeReconnectAttempt += 1;
        realtimeReconnectTimer = window.setTimeout(() => {
          realtimeReconnectTimer = null;
          void connectRealtime();
        }, delay);
      };
    } catch {
      if (selectedChatId.value !== chatId) {
        return;
      }
      const delay = Math.min(10000, 1000 * 2 ** realtimeReconnectAttempt);
      realtimeReconnectAttempt += 1;
      realtimeReconnectTimer = window.setTimeout(() => {
        realtimeReconnectTimer = null;
        void connectRealtime();
      }, delay);
    }
  };

  // 4.10 開啟大廈群聊成員名單
  const handleOpenGroupMembers = async (): Promise<void> => {
    if (!selectedChatId.value || activeConversation.value?.type !== buildingGroupChatType) {
      return;
    }

    groupMembersOpen.value = true;
    loadingGroupMembers.value = true;

    try {
      const { data } = await fetchBuildingChatMembers(selectedChatId.value);
      groupMembers.value = data.data.items.map((member) => ({
        user_id: member.user_id,
        display_name: member.display_name,
        role_in_chat: member.role_in_chat,
        membership_status: member.membership_status,
        muted_until: member.muted_until,
        banned_until: member.banned_until,
      }));
    } catch (error) {
      feedbackStore.pushToast(
        axios.isAxiosError(error)
          ? error.response?.data?.message ?? t('chat.loadMembersError')
          : t('chat.loadMembersError'),
        'error',
      );
      groupMembers.value = [];
    } finally {
      loadingGroupMembers.value = false;
    }
  };

  // 4.15 關閉大廈群聊成員名單
  const handleCloseGroupMembers = (): void => {
    groupMembersOpen.value = false;
    groupMembers.value = [];
  };

  // 4.16 確認離開大廈群聊
  const handleLeaveBuildingChat = async (): Promise<void> => {
    if (!selectedChatId.value || activeConversation.value?.type !== buildingGroupChatType || leavingGroup.value) {
      return;
    }
    leavingGroup.value = true;

    try {
      await leaveBuildingChat(selectedChatId.value);
      leaveConfirmOpen.value = false;
      await handleBackToChats();
      await loadChats();
    } catch (error) {
      feedbackStore.pushToast(
        axios.isAxiosError(error)
          ? error.response?.data?.message ?? t('chat.leaveError')
          : t('chat.leaveError'),
        'error',
      );
    } finally {
      leavingGroup.value = false;
    }
  };

  // 4.13 綁定訊息滾動容器
  const setMessageContainerRef = (element: Element | ComponentPublicInstance | null): void => {
    messageContainerRef.value = element instanceof HTMLDivElement ? element : null;
  };

  watch(
    () => selectedChatId.value,
    () => {
      const chatId = selectedChatId.value;
      closeRealtime();
      draftMessage.value = '';
      groupMembersOpen.value = false;
      leaveConfirmOpen.value = false;
      void (async () => {
        await loadActiveChat();
        if (selectedChatId.value === chatId) {
          await connectRealtime();
        }
      })();
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

  onMounted(() => {
    void loadChats();
  });

  onBeforeUnmount(() => {
    closeRealtime();
  });

  return {
    activeConversation,
    activeReferencePrice,
    activeMessages,
    canSendMessage,
    conversations,
    draftMessage,
    formatPrice,
    groupMembers,
    groupMembersOpen,
    handleBackToChats,
    handleCloseGroupMembers,
    handleLeaveBuildingChat,
    handleOpenGroupMembers,
    handleSelectChat,
    handleSendByEnter,
    handleSendMessage,
    leaveConfirmOpen,
    leavingGroup,
    loadingConversations,
    loadingGroupMembers,
    loadingMessages,
    messageMaxLength,
    preferenceStore,
    selectedChatId,
    setMessageContainerRef,
    sendingMessage,
    uploadingAttachment,
    handleSelectAttachment,
    sessionStore,
    t,
    totalUnreadCount,
  };
};
