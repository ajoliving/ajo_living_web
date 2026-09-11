<!--
 * 管理 - 大廈群聊管理頁。
 * 1. 查看所有大廈群聊列表。
 * 2. 審核待處理加入申請，並處置群聊成員。
 * 3. 以統一管理列表呈現群聊詳情、申請與成員資料。
-->
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import {
  fetchStaffBuildingChats,
  fetchBuildingChatMembers,
  fetchBuildingChatJoinRequests,
  moderateBuildingChatMember,
  reviewBuildingChatJoin,
  type BuildingChatSummaryResponse,
  type BuildingChatMemberResponse,
  type BuildingChatJoinRequestResponse,
} from '@/httpapis/chats';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import { useFeedbackStore } from '@/stores/feedback';

import '../styles.scss';

const { t, locale } = useI18n();
const feedbackStore = useFeedbackStore();
const loading = ref(false);
const loadingMembers = ref(false);
const loadingRequests = ref(false);
const chats = ref<BuildingChatSummaryResponse[]>([]);
const selectedChat = ref<BuildingChatSummaryResponse | null>(null);
const members = ref<BuildingChatMemberResponse[]>([]);
const joinRequests = ref<BuildingChatJoinRequestResponse[]>([]);
const processingMemberId = ref('');
const reviewingRequestId = ref('');

type MemberAction = 'mute' | 'unmute' | 'kick' | 'ban' | 'unban';

const pendingJoinRequests = computed(() =>
  joinRequests.value.filter((request) => request.status === 'pending'),
);

// 1. 讀取大廈群聊列表
const loadChats = async (): Promise<void> => {
  loading.value = true;

  try {
    const { data } = await fetchStaffBuildingChats();
    chats.value = Array.isArray(data.data?.items) ? data.data.items : [];
  } catch {
    feedbackStore.pushToast(t('marketplace.management.buildingChats.loadError'), 'error');
    chats.value = [];
  } finally {
    loading.value = false;
  }
};

// 2. 同步目前選中群聊摘要
const refreshSelectedChatSummary = async (): Promise<void> => {
  if (!selectedChat.value) {
    return;
  }

  try {
    const { data } = await fetchStaffBuildingChats();
    chats.value = Array.isArray(data.data?.items) ? data.data.items : [];
    selectedChat.value = chats.value.find((chat) => chat.chat_id === selectedChat.value?.chat_id) ?? null;
  } catch {
    feedbackStore.pushToast(t('marketplace.management.buildingChats.loadError'), 'error');
  }
};

// 3. 格式化大廈名稱
const formatBuildingName = (buildingId: string): string => {
  return buildingId || t('marketplace.management.buildingChats.unknownBuilding');
};

// 4. 格式化管理列表的完整日期時間
const formatDateTime = (value: string): string => {
  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return '-';
  }

  return new Intl.DateTimeFormat(locale.value, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(date);
};

// 5. 格式化最後訊息時間
const formatLastMessageTime = (time: string | null | undefined): string => {
  if (!time) {
    return '-';
  }
  return formatDateTime(time);
};

// 6. 選擇群聊並載入管理資料
const selectChat = async (chat: BuildingChatSummaryResponse): Promise<void> => {
  selectedChat.value = chat;
  await Promise.all([loadMembers(chat.chat_id), loadJoinRequests(chat.chat_id)]);
};

// 7. 讀取群聊成員
const loadMembers = async (chatId: string): Promise<void> => {
  loadingMembers.value = true;

  try {
    const { data } = await fetchBuildingChatMembers(chatId);
    members.value = Array.isArray(data.data?.items) ? data.data.items : [];
  } catch {
    feedbackStore.pushToast(t('marketplace.management.buildingChats.loadMembersError'), 'error');
    members.value = [];
  } finally {
    loadingMembers.value = false;
  }
};

// 8. 讀取加入申請
const loadJoinRequests = async (chatId: string): Promise<void> => {
  loadingRequests.value = true;

  try {
    const { data } = await fetchBuildingChatJoinRequests(chatId);
    joinRequests.value = Array.isArray(data.data?.items) ? data.data.items : [];
  } catch {
    feedbackStore.pushToast(t('marketplace.management.buildingChats.loadRequestsError'), 'error');
    joinRequests.value = [];
  } finally {
    loadingRequests.value = false;
  }
};

// 9. 執行成員管理操作
const handleMemberAction = async (member: BuildingChatMemberResponse, action: MemberAction): Promise<void> => {
  if (!selectedChat.value || processingMemberId.value) {
    return;
  }

  if (action === 'kick' && !confirm(t('marketplace.management.buildingChats.kickConfirm', { name: member.display_name }))) {
    return;
  }
  if (action === 'ban' && !confirm(t('marketplace.management.buildingChats.banConfirm', { name: member.display_name }))) {
    return;
  }

  processingMemberId.value = member.user_id;
  try {
    await moderateBuildingChatMember(selectedChat.value.chat_id, {
      user_id: Number(member.user_id),
      action,
      duration_minutes: action === 'mute' ? 60 * 24 : undefined,
    });
    feedbackStore.pushToast(t(`marketplace.management.buildingChats.${action}Success`), 'success');
    await Promise.all([loadMembers(selectedChat.value.chat_id), refreshSelectedChatSummary()]);
  } catch {
    feedbackStore.pushToast(t(`marketplace.management.buildingChats.${action}Error`), 'error');
  } finally {
    processingMemberId.value = '';
  }
};

// 10. 審核加入申請
const handleReviewJoinRequest = async (
  request: BuildingChatJoinRequestResponse,
  status: 'approved' | 'rejected',
): Promise<void> => {
  if (!selectedChat.value || reviewingRequestId.value) {
    return;
  }

  reviewingRequestId.value = request.request_id;
  try {
    await reviewBuildingChatJoin(request.request_id, status);
    feedbackStore.pushToast(
      t(`marketplace.management.buildingChats.${status === 'approved' ? 'approveSuccess' : 'rejectSuccess'}`),
      'success',
    );
    await Promise.all([
      loadMembers(selectedChat.value.chat_id),
      loadJoinRequests(selectedChat.value.chat_id),
      refreshSelectedChatSummary(),
    ]);
  } catch {
    feedbackStore.pushToast(t('marketplace.management.buildingChats.reviewError'), 'error');
  } finally {
    reviewingRequestId.value = '';
  }
};

// 11. 格式化成員狀態
const formatMemberStatus = (status: string): string => {
  const knownStatuses = ['active', 'banned', 'kicked', 'left'];

  return knownStatuses.includes(status)
    ? t(`marketplace.management.buildingChats.memberStatuses.${status}`)
    : status;
};

onMounted(() => {
  void loadChats();
});
</script>

<template>
  <section class="management-list-page building-chats-page">
    <header class="management-list-header building-chats-header">
      <div>
        <p class="management-list-kicker">{{ t('marketplace.management.buildingChats.kicker') }}</p>
        <h1>{{ t('marketplace.management.buildingChats.title') }}</h1>
        <p>{{ t('marketplace.management.buildingChats.description') }}</p>
      </div>
      <button
        type="button"
        class="management-list-button building-chats-refresh-button"
        :disabled="loading"
        @click="loadChats"
      >
        <AppIcon
          name="reload"
          :size="16"
        />
        {{ t('marketplace.management.buildingChats.refresh') }}
      </button>
    </header>

    <section
      v-if="loading"
      class="management-list-panel building-chats-state-panel"
    >
      <div class="management-list-loading">
        {{ t('marketplace.management.buildingChats.loading') }}
      </div>
    </section>

    <section
      v-else-if="chats.length === 0"
      class="management-list-panel building-chats-state-panel"
    >
      <div class="building-chats-empty">
        <AppIcon
          name="message"
          :size="48"
        />
        <strong>{{ t('marketplace.management.buildingChats.emptyTitle') }}</strong>
        <span>{{ t('marketplace.management.buildingChats.emptyDescription') }}</span>
      </div>
    </section>

    <template v-else>
      <article class="management-list-panel building-chat-table-panel">
        <div class="management-list-toolbar building-chat-table-toolbar">
          <span class="building-chat-table-toolbar__count">
            {{ chats.length }} {{ t('marketplace.management.buildingChats.navTitle') }}
          </span>
        </div>
        <div class="management-table-wrap">
          <table class="management-table building-chat-table">
            <thead>
              <tr>
                <th scope="col">{{ t('marketplace.management.buildingChats.buildingId') }}</th>
                <th scope="col">{{ t('marketplace.management.buildingChats.memberCount') }}</th>
                <th scope="col">{{ t('marketplace.management.buildingChats.lastMessage') }}</th>
                <th scope="col">{{ t('marketplace.management.buildingChats.lastMessageTime') }}</th>
                <th scope="col">{{ t('marketplace.management.buildingChats.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="chat in chats"
                :key="chat.chat_id"
                :class="selectedChat?.chat_id === chat.chat_id ? 'selected' : ''"
              >
                <td>
                  <strong>{{ formatBuildingName(chat.building_id) }}</strong>
                </td>
                <td>{{ chat.member_count }}</td>
                <td class="management-table-text">{{ chat.last_message_preview || '-' }}</td>
                <td>{{ formatLastMessageTime(chat.last_message_at) }}</td>
                <td class="management-table-actions">
                  <button
                    type="button"
                    class="management-list-action building-chat-select-action"
                    :class="{ 'is-selected': selectedChat?.chat_id === chat.chat_id }"
                    :aria-pressed="selectedChat?.chat_id === chat.chat_id"
                    @click="selectChat(chat)"
                  >
                    {{ t('marketplace.management.buildingChats.viewMembers') }}
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </article>

      <article
        v-if="selectedChat"
        class="management-list-panel building-chat-detail-panel"
      >
        <header class="building-chat-detail-header">
          <div>
            <h2 class="building-chat-detail-title">
              {{ t('marketplace.management.buildingChats.membersTitle') }}
            </h2>
            <p class="building-chat-detail-subtitle">
              {{ formatBuildingName(selectedChat.building_id) }} · {{ members.length }} {{ t('marketplace.management.buildingChats.membersUnit') }}
            </p>
          </div>
          <span
            v-if="pendingJoinRequests.length"
            class="management-badge management-badge--pending"
          >{{ pendingJoinRequests.length }}</span>
        </header>

        <section class="building-chat-detail-section">
          <header class="building-chat-detail-section__header">
            <h3>{{ t('marketplace.management.buildingChats.requestsTitle') }}</h3>
          </header>

          <div
            v-if="loadingRequests"
            class="building-chat-inline-state"
          >{{ t('common.status.loading') }}</div>

          <div
            v-else-if="pendingJoinRequests.length === 0"
            class="building-chat-inline-state"
          >{{ t('marketplace.management.buildingChats.noRequests') }}</div>

          <div
            v-else
            class="management-table-wrap"
          >
            <table class="management-table building-chat-table">
              <thead>
                <tr>
                  <th scope="col">{{ t('marketplace.management.buildingChats.requestMember') }}</th>
                  <th scope="col">{{ t('marketplace.management.buildingChats.requestReason') }}</th>
                  <th scope="col">{{ t('marketplace.management.buildingChats.requestedAt') }}</th>
                  <th scope="col">{{ t('marketplace.management.buildingChats.actions') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="request in pendingJoinRequests"
                  :key="request.request_id"
                >
                  <td>
                    <div class="management-user-cell">
                      <strong>{{ request.display_name || request.user_public_id || request.user_id }}</strong>
                      <span>{{ request.user_public_id || request.user_id }}</span>
                    </div>
                  </td>
                  <td class="management-table-text">{{ request.reason || '-' }}</td>
                  <td>{{ formatDateTime(request.created_at) }}</td>
                  <td class="management-table-actions">
                    <div class="management-action-group">
                      <button
                        type="button"
                        class="management-list-action"
                        :disabled="reviewingRequestId === request.request_id"
                        @click="handleReviewJoinRequest(request, 'approved')"
                      >{{ t('marketplace.management.buildingChats.approve') }}</button>
                      <button
                        type="button"
                        class="management-list-action management-list-action--danger"
                        :disabled="reviewingRequestId === request.request_id"
                        @click="handleReviewJoinRequest(request, 'rejected')"
                      >{{ t('marketplace.management.buildingChats.reject') }}</button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <section class="building-chat-detail-section">
          <header class="building-chat-detail-section__header">
            <h3>{{ t('marketplace.management.buildingChats.membersTitle') }}</h3>
          </header>

          <div
            v-if="loadingMembers"
            class="building-chat-inline-state"
          >
            {{ t('marketplace.management.buildingChats.loadingMembers') }}
          </div>

          <div
            v-else-if="members.length === 0"
            class="building-chat-inline-state"
          >
            {{ t('marketplace.management.buildingChats.noMembers') }}
          </div>

          <div
            v-else
            class="management-table-wrap"
          >
            <table class="management-table building-chat-table">
              <thead>
                <tr>
                  <th scope="col">{{ t('marketplace.management.buildingChats.memberName') }}</th>
                  <th scope="col">{{ t('marketplace.management.buildingChats.memberRole') }}</th>
                  <th scope="col">{{ t('marketplace.management.buildingChats.memberStatus') }}</th>
                  <th scope="col">{{ t('marketplace.management.buildingChats.joinedAt') }}</th>
                  <th scope="col" class="management-table-actions">{{ t('marketplace.management.buildingChats.actions') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="member in members"
                  :key="member.user_id"
                >
                  <td>
                    <div class="management-user-cell">
                      <strong>{{ member.display_name }}</strong>
                      <span>{{ member.public_id }}</span>
                    </div>
                  </td>
                  <td>{{ member.role_in_chat }}</td>
                  <td>
                    <span
                      class="management-status-pill building-chat-status-pill"
                      :class="member.membership_status"
                    >
                      {{ formatMemberStatus(member.membership_status) }}
                    </span>
                    <span
                      v-if="member.muted_until"
                      class="management-status-pill building-chat-status-pill muted"
                    >
                      {{ t('marketplace.management.buildingChats.muted') }}
                    </span>
                  </td>
                  <td>{{ formatDateTime(member.joined_at) }}</td>
                  <td class="management-table-actions">
                    <div class="management-action-group">
                      <button
                        v-if="member.membership_status === 'active' && !member.muted_until"
                        type="button"
                        class="management-list-action"
                        :disabled="processingMemberId === member.user_id"
                        @click="handleMemberAction(member, 'mute')"
                      >
                        {{ t('marketplace.management.buildingChats.mute') }}
                      </button>
                      <button
                        v-if="member.membership_status === 'active' && member.muted_until"
                        type="button"
                        class="management-list-action"
                        :disabled="processingMemberId === member.user_id"
                        @click="handleMemberAction(member, 'unmute')"
                      >
                        {{ t('marketplace.management.buildingChats.unmute') }}
                      </button>
                      <button
                        v-if="member.membership_status === 'active'"
                        type="button"
                        class="management-list-action management-list-action--danger"
                        :disabled="processingMemberId === member.user_id"
                        @click="handleMemberAction(member, 'kick')"
                      >
                        {{ t('marketplace.management.buildingChats.kick') }}
                      </button>
                      <button
                        v-if="member.membership_status === 'active'"
                        type="button"
                        class="management-list-action management-list-action--danger"
                        :disabled="processingMemberId === member.user_id"
                        @click="handleMemberAction(member, 'ban')"
                      >
                        {{ t('marketplace.management.buildingChats.ban') }}
                      </button>
                      <button
                        v-if="member.membership_status === 'banned'"
                        type="button"
                        class="management-list-action"
                        :disabled="processingMemberId === member.user_id"
                        @click="handleMemberAction(member, 'unban')"
                      >
                        {{ t('marketplace.management.buildingChats.unban') }}
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </article>
    </template>
  </section>
</template>

<style scoped>
/* 1. 群聊列表頁主容器 */
.building-chats-page {
  gap: 0.9rem;
  background: rgb(var(--color-page-tint));
  padding: 0.9rem;
}

.building-chats-header {
  align-items: flex-end;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface-raised));
  padding: 0.9rem;
}

.building-chats-header > div {
  min-width: 0;
  max-width: 42rem;
}

.building-chats-refresh-button {
  flex: 0 0 auto;
}

.building-chats-state-panel {
  border-radius: 2px;
}

.building-chats-empty {
  display: flex;
  min-height: 12.5rem;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.7rem;
  padding: 2rem 1rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.8125rem;
  line-height: 1.5;
  text-align: center;
}

.building-chats-empty strong {
  color: rgb(var(--color-text));
  font-size: 0.9375rem;
  font-weight: 750;
}

/* 2. 表格工具列與清單 */
.building-chat-table-panel,
.building-chat-detail-panel {
  border-radius: 2px;
}

.building-chat-table-toolbar {
  display: flex;
  min-height: 3.25rem;
  align-items: center;
  border-bottom: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface));
  padding: 0.8rem 0.9rem;
}

.building-chat-table-toolbar__count {
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0.06em;
  line-height: 1.2;
  text-transform: uppercase;
}

.building-chats-page .management-table {
  min-width: 50rem;
}

.building-chats-page .management-table th,
.building-chats-page .management-table td {
  padding: 0.72rem 0.9rem;
}

.building-chats-page .management-table tbody tr {
  transition: border-color 0.16s ease, box-shadow 0.16s ease;
}

.building-chats-page .management-table tbody tr:hover {
  background: rgb(var(--color-surface));
}

.building-chats-page .management-table tbody tr.selected {
  background: rgb(var(--color-surface-raised));
  box-shadow: inset 3px 0 0 rgb(var(--color-primary));
}

.building-chats-page .management-table tbody tr.selected td {
  border-bottom-color: rgb(var(--color-primary) / 0.24);
}

.building-chats-page .management-table tbody tr.selected td:first-child {
  box-shadow: none;
}

.building-chat-table th:first-child {
  width: 20%;
}

.building-chat-table th:last-child {
  width: 16rem;
}

.building-chats-page .management-table-actions {
  width: 16rem;
  min-width: 16rem;
}

.building-chats-page .management-table-text {
  max-width: 18rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 3. 已選群組詳情 */
.building-chat-detail-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface));
  padding: 0.9rem;
}

.building-chat-detail-title {
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 1rem;
  font-weight: 750;
  line-height: 1.35;
}

.building-chat-detail-subtitle {
  margin: 0.3rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  line-height: 1.5;
}

.building-chats-page .management-badge {
  display: inline-flex;
  min-width: 1.65rem;
  min-height: 1.65rem;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: rgb(var(--color-primary));
  padding: 0 0.55rem;
  color: rgb(var(--color-primary-contrast));
  font-size: 0.7rem;
  font-weight: 800;
  line-height: 1;
}

.building-chats-page .management-badge--pending {
  background: rgb(var(--color-primary-soft));
  color: rgb(var(--color-primary));
}

.building-chat-detail-section + .building-chat-detail-section {
  border-top: 1px solid rgb(var(--color-border));
}

.building-chat-detail-section__header {
  display: flex;
  align-items: center;
  min-height: 3.1rem;
  padding: 0.72rem 0.9rem;
}

.building-chat-detail-section__header h3 {
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 0.8125rem;
  font-weight: 750;
  line-height: 1.35;
}

.building-chat-inline-state {
  min-height: 4.75rem;
  padding: 1rem 0.9rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.8125rem;
  font-weight: 650;
  line-height: 1.5;
}

.building-chats-page .management-user-cell {
  display: grid;
  min-width: 0;
  gap: 0.18rem;
}

.building-chats-page .management-user-cell strong {
  font-size: 0.875rem;
  font-weight: 750;
}

.building-chats-page .management-user-cell span {
  font-size: 0.7rem;
}

.building-chat-status-pill {
  margin-right: 0.3rem;
}

.building-chat-status-pill.active {
  background: rgb(34 197 94 / 0.12);
  color: rgb(21 128 61);
}

.building-chat-status-pill.banned {
  background: rgb(239 68 68 / 0.12);
  color: rgb(185 28 28);
}

.building-chat-status-pill.kicked,
.building-chat-status-pill.left {
  background: rgb(var(--color-surface-muted));
  color: rgb(var(--color-text-muted));
}

.building-chat-status-pill.muted {
  background: rgb(251 146 60 / 0.13);
  color: rgb(194 65 12);
}

/* 4. 操作與焦點態 */
.building-chats-page .management-action-group {
  flex-wrap: nowrap;
  gap: 0.4rem;
}

.building-chats-page .management-list-action {
  position: relative;
  min-height: 2.1rem;
  border-radius: 2px;
  background: rgb(var(--color-surface));
  font-size: 0.75rem;
}

.building-chats-page .management-list-action:hover:not(:disabled) {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-surface));
  color: rgb(var(--color-primary));
}

.building-chats-page .management-list-action.is-selected {
  border-color: rgb(var(--color-primary));
  color: rgb(var(--color-primary));
}

.building-chats-page .management-list-action--danger {
  border-color: rgb(239 68 68 / 0.35);
  color: rgb(185 28 28);
}

.building-chats-page .management-list-action--danger:hover:not(:disabled) {
  border-color: rgb(220 38 38);
  background: rgb(var(--color-surface));
  color: rgb(185 28 28);
}

.building-chats-page :is(.management-list-button, .management-list-action):focus-visible {
  border-color: rgb(var(--color-primary));
  box-shadow:
    inset 0 -2px 0 rgb(var(--color-primary)),
    0 0 0 2px rgb(var(--color-primary) / 0.16);
  outline: 0;
}

.building-chats-page .management-list-action:disabled,
.building-chats-page .management-list-button:disabled {
  cursor: not-allowed;
  opacity: 0.48;
}

/* 5. 行動版 */
@media (max-width: 767px) {
  .building-chats-page {
    gap: 0.75rem;
    padding: 0.75rem;
  }

  .building-chats-header {
    align-items: stretch;
  }

  .building-chat-detail-header {
    flex-direction: column;
  }

  .building-chats-page .management-badge--pending {
    align-self: flex-start;
  }
}
</style>
