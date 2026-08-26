<!--
 * 管理 - 大廈群聊管理頁。
 * 1. 查看所有大廈群聊列表。
 * 2. 管理群聊成員：邀請用戶直接加入、移除成員、禁言等。
 * 3. 查看群聊詳情與成員列表。
-->
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import {
  fetchStaffBuildingChats,
  fetchBuildingChatMembers,
  moderateBuildingChatMember,
  type BuildingChatSummaryResponse,
  type BuildingChatMemberResponse,
} from '@/httpapis/chats';
import { fetchStaffUsers } from '@/httpapis/staff';
import type { StaffUserSummary } from '@/model/user';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import { useFeedbackStore } from '@/stores/feedback';
import { formatDate } from '@/utils/format';

import '../styles.scss';

const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const loading = ref(false);
const loadingMembers = ref(false);
const loadingUsers = ref(false);
const chats = ref<BuildingChatSummaryResponse[]>([]);
const selectedChat = ref<BuildingChatSummaryResponse | null>(null);
const members = ref<BuildingChatMemberResponse[]>([]);
const availableUsers = ref<StaffUserSummary[]>([]);
const isInviteDialogOpen = ref(false);
const inviting = ref(false);
const selectedUserId = ref('');
const searchKeyword = ref('');

// 1. 讀取大廈群聊列表
const loadChats = async (): Promise<void> => {
  loading.value = true;

  try {
    const { data } = await fetchStaffBuildingChats();
    chats.value = data.data.items;
  } catch (error: unknown) {
    console.error('Load building chats error:', error);
    feedbackStore.pushToast(t('marketplace.management.buildingChats.loadError'), 'error');
    chats.value = [];
  } finally {
    loading.value = false;
  }
};

// 2. 格式化大廈名稱
const formatBuildingName = (buildingId: string): string => {
  return buildingId || t('marketplace.management.buildingChats.unknownBuilding');
};

// 3. 格式化最後訊息時間
const formatLastMessageTime = (time: string | null | undefined): string => {
  if (!time) {
    return '-';
  }
  return formatDate(time, 'yyyy/MM/dd HH:mm');
};

// 4. 選擇群聊並載入成員
const selectChat = async (chat: BuildingChatSummaryResponse): Promise<void> => {
  selectedChat.value = chat;
  await loadMembers(chat.chat_id);
};

// 5. 讀取群聊成員
const loadMembers = async (chatId: string): Promise<void> => {
  loadingMembers.value = true;

  try {
    const { data } = await fetchBuildingChatMembers(chatId);
    members.value = data.data.items;
  } catch (error: unknown) {
    console.error('Load members error:', error);
    feedbackStore.pushToast(t('marketplace.management.buildingChats.loadMembersError'), 'error');
    members.value = [];
  } finally {
    loadingMembers.value = false;
  }
};

// 6. 開啟邀請彈窗
const openInviteDialog = async (): Promise<void> => {
  selectedUserId.value = '';
  searchKeyword.value = '';
  isInviteDialogOpen.value = true;
  await loadAvailableUsers();
};

// 7. 關閉邀請彈窗
const closeInviteDialog = (): void => {
  if (!inviting.value) {
    isInviteDialogOpen.value = false;
  }
};

// 8. 讀取可邀請用戶列表
const loadAvailableUsers = async (): Promise<void> => {
  loadingUsers.value = true;

  try {
    const { data } = await fetchStaffUsers({
      page: 1,
      page_size: 100,
      keyword: searchKeyword.value.trim() || undefined,
    });
    availableUsers.value = data.data.items;
  } catch {
    feedbackStore.pushToast(t('marketplace.management.buildingChats.loadUsersError'), 'error');
  } finally {
    loadingUsers.value = false;
  }
};

// 9. 搜尋用戶
const searchUsers = async (): Promise<void> => {
  await loadAvailableUsers();
};

// 10. 格式化用戶名稱
const formatUserName = (user: StaffUserSummary): string => {
  return user.display_name?.trim() || user.email || user.phone_number || user.public_id;
};

// 11. 格式化用戶標識
const formatUserIdentifier = (user: StaffUserSummary): string => {
  const parts: string[] = [];
  if (user.email) {
    parts.push(user.email);
  }
  if (user.phone_number) {
    parts.push(`${user.phone_country_code} ${user.phone_number}`.trim());
  }
  return parts.join(' · ') || user.public_id;
};

// 12. 可邀請的用戶選項
const userOptions = computed(() =>
  availableUsers.value
    .filter((user) => !members.value.some((member) => member.public_id === user.public_id))
    .map((user) => ({
      value: user.public_id,
      label: formatUserName(user),
      subtitle: formatUserIdentifier(user),
    })),
);

// 13. 檢查是否可以邀請
const canInvite = computed(() => selectedUserId.value.trim().length > 0);

// 14. 邀請用戶加入群聊（直接加入，不需要審批）
const submitInvite = async (): Promise<void> => {
  if (!selectedChat.value || !canInvite.value) {
    feedbackStore.pushToast(t('marketplace.management.buildingChats.inviteRequired'), 'error');
    return;
  }

  inviting.value = true;
  try {
    await moderateBuildingChatMember(selectedChat.value.chat_id, {
      user_id: Number(selectedUserId.value),
      action: 'add',
    });
    feedbackStore.pushToast(t('marketplace.management.buildingChats.inviteSuccess'), 'success');
    isInviteDialogOpen.value = false;
    await loadMembers(selectedChat.value.chat_id);
  } catch {
    feedbackStore.pushToast(t('marketplace.management.buildingChats.inviteError'), 'error');
  } finally {
    inviting.value = false;
  }
};

// 15. 禁言成員
const muteMember = async (member: BuildingChatMemberResponse): Promise<void> => {
  if (!selectedChat.value) {
    return;
  }

  try {
    await moderateBuildingChatMember(selectedChat.value.chat_id, {
      user_id: Number(member.user_id),
      action: 'mute',
      duration_minutes: 60 * 24,
    });
    feedbackStore.pushToast(t('marketplace.management.buildingChats.muteSuccess'), 'success');
    await loadMembers(selectedChat.value.chat_id);
  } catch {
    feedbackStore.pushToast(t('marketplace.management.buildingChats.muteError'), 'error');
  }
};

// 16. 解除禁言
const unmuteMember = async (member: BuildingChatMemberResponse): Promise<void> => {
  if (!selectedChat.value) {
    return;
  }

  try {
    await moderateBuildingChatMember(selectedChat.value.chat_id, {
      user_id: Number(member.user_id),
      action: 'unmute',
    });
    feedbackStore.pushToast(t('marketplace.management.buildingChats.unmuteSuccess'), 'success');
    await loadMembers(selectedChat.value.chat_id);
  } catch {
    feedbackStore.pushToast(t('marketplace.management.buildingChats.unmuteError'), 'error');
  }
};

// 17. 移除成員
const kickMember = async (member: BuildingChatMemberResponse): Promise<void> => {
  if (!selectedChat.value) {
    return;
  }

  if (!confirm(t('marketplace.management.buildingChats.kickConfirm', { name: member.display_name }))) {
    return;
  }

  try {
    await moderateBuildingChatMember(selectedChat.value.chat_id, {
      user_id: Number(member.user_id),
      action: 'kick',
    });
    feedbackStore.pushToast(t('marketplace.management.buildingChats.kickSuccess'), 'success');
    await loadMembers(selectedChat.value.chat_id);
  } catch {
    feedbackStore.pushToast(t('marketplace.management.buildingChats.kickError'), 'error');
  }
};

// 18. 封禁成員
const banMember = async (member: BuildingChatMemberResponse): Promise<void> => {
  if (!selectedChat.value) {
    return;
  }

  if (!confirm(t('marketplace.management.buildingChats.banConfirm', { name: member.display_name }))) {
    return;
  }

  try {
    await moderateBuildingChatMember(selectedChat.value.chat_id, {
      user_id: Number(member.user_id),
      action: 'ban',
    });
    feedbackStore.pushToast(t('marketplace.management.buildingChats.banSuccess'), 'success');
    await loadMembers(selectedChat.value.chat_id);
  } catch {
    feedbackStore.pushToast(t('marketplace.management.buildingChats.banError'), 'error');
  }
};

// 19. 解除封禁
const unbanMember = async (member: BuildingChatMemberResponse): Promise<void> => {
  if (!selectedChat.value) {
    return;
  }

  try {
    await moderateBuildingChatMember(selectedChat.value.chat_id, {
      user_id: Number(member.user_id),
      action: 'unban',
    });
    feedbackStore.pushToast(t('marketplace.management.buildingChats.unbanSuccess'), 'success');
    await loadMembers(selectedChat.value.chat_id);
  } catch {
    feedbackStore.pushToast(t('marketplace.management.buildingChats.unbanError'), 'error');
  }
};

onMounted(() => {
  void loadChats();
});
</script>

<template>
  <div class="management-work-panel">
    <section class="work-hero">
      <div>
        <div class="work-kicker">{{ t('marketplace.management.buildingChats.kicker') }}</div>
        <h2 class="work-title">{{ t('marketplace.management.buildingChats.title') }}</h2>
        <p class="work-desc">{{ t('marketplace.management.buildingChats.description') }}</p>
      </div>
    </section>

    <section
      v-if="loading"
      class="work-card"
    >
      <div class="management-loading">
        {{ t('marketplace.management.buildingChats.loading') }}
      </div>
    </section>

    <section
      v-else-if="chats.length === 0"
      class="work-card"
    >
      <div class="management-empty">
        <AppIcon
          name="message"
          :size="48"
        />
        <strong>{{ t('marketplace.management.buildingChats.emptyTitle') }}</strong>
        <span>{{ t('marketplace.management.buildingChats.emptyDescription') }}</span>
      </div>
    </section>

    <template v-else>
      <section class="work-card">
        <div class="management-table-container">
          <table class="management-table">
            <thead>
              <tr>
                <th>{{ t('marketplace.management.buildingChats.buildingId') }}</th>
                <th>{{ t('marketplace.management.buildingChats.memberCount') }}</th>
                <th>{{ t('marketplace.management.buildingChats.unreadCount') }}</th>
                <th>{{ t('marketplace.management.buildingChats.lastMessage') }}</th>
                <th>{{ t('marketplace.management.buildingChats.lastMessageTime') }}</th>
                <th>{{ t('marketplace.management.buildingChats.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="chat in chats"
                :key="chat.chat_id"
                :class="selectedChat?.chat_id === chat.chat_id ? 'selected' : ''"
              >
                <td>{{ formatBuildingName(chat.building_id) }}</td>
                <td>{{ chat.member_count }}</td>
                <td>
                  <span
                    v-if="chat.unread_count > 0"
                    class="management-badge"
                  >
                    {{ chat.unread_count }}
                  </span>
                  <span v-else>-</span>
                </td>
                <td class="management-table-text">{{ chat.last_message_preview || '-' }}</td>
                <td>{{ formatLastMessageTime(chat.last_message_at) }}</td>
                <td>
                  <button
                    type="button"
                    class="management-action-button"
                    @click="selectChat(chat)"
                  >
                    {{ t('marketplace.management.buildingChats.viewMembers') }}
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section
        v-if="selectedChat"
        class="work-card"
      >
        <div class="management-section-header">
          <div>
            <h3 class="management-section-title">
              {{ t('marketplace.management.buildingChats.membersTitle') }}
            </h3>
            <p class="management-section-subtitle">
              {{ formatBuildingName(selectedChat.building_id) }} · {{ members.length }} {{ t('marketplace.management.buildingChats.membersUnit') }}
            </p>
          </div>
          <button
            type="button"
            class="management-primary-button"
            @click="openInviteDialog"
          >
            <AppIcon
              name="user"
              :size="16"
            />
            {{ t('marketplace.management.buildingChats.inviteUser') }}
          </button>
        </div>

        <div
          v-if="loadingMembers"
          class="management-loading"
        >
          {{ t('marketplace.management.buildingChats.loadingMembers') }}
        </div>

        <div
          v-else-if="members.length === 0"
          class="management-empty"
        >
          <strong>{{ t('marketplace.management.buildingChats.noMembers') }}</strong>
        </div>

        <div
          v-else
          class="management-table-container"
        >
          <table class="management-table">
            <thead>
              <tr>
                <th>{{ t('marketplace.management.buildingChats.memberName') }}</th>
                <th>{{ t('marketplace.management.buildingChats.memberRole') }}</th>
                <th>{{ t('marketplace.management.buildingChats.memberStatus') }}</th>
                <th>{{ t('marketplace.management.buildingChats.joinedAt') }}</th>
                <th>{{ t('marketplace.management.buildingChats.actions') }}</th>
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
                    class="management-status-badge"
                    :class="member.membership_status"
                  >
                    {{ member.membership_status }}
                  </span>
                  <span
                    v-if="member.muted_until"
                    class="management-status-badge muted"
                  >
                    {{ t('marketplace.management.buildingChats.muted') }}
                  </span>
                </td>
                <td>{{ formatDate(member.joined_at, 'yyyy/MM/dd HH:mm') }}</td>
                <td>
                  <div class="management-action-group">
                    <button
                      v-if="member.membership_status === 'active' && !member.muted_until"
                      type="button"
                      class="management-action-button"
                      @click="muteMember(member)"
                    >
                      {{ t('marketplace.management.buildingChats.mute') }}
                    </button>
                    <button
                      v-if="member.membership_status === 'active' && member.muted_until"
                      type="button"
                      class="management-action-button"
                      @click="unmuteMember(member)"
                    >
                      {{ t('marketplace.management.buildingChats.unmute') }}
                    </button>
                    <button
                      v-if="member.membership_status === 'active'"
                      type="button"
                      class="management-action-button management-action-button--danger"
                      @click="kickMember(member)"
                    >
                      {{ t('marketplace.management.buildingChats.kick') }}
                    </button>
                    <button
                      v-if="member.membership_status === 'active'"
                      type="button"
                      class="management-action-button management-action-button--danger"
                      @click="banMember(member)"
                    >
                      {{ t('marketplace.management.buildingChats.ban') }}
                    </button>
                    <button
                      v-if="member.membership_status === 'banned'"
                      type="button"
                      class="management-action-button"
                      @click="unbanMember(member)"
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
    </template>

    <div
      v-if="isInviteDialogOpen"
      class="management-modal"
      @click.self="closeInviteDialog"
    >
      <div class="management-dialog">
        <header class="management-dialog__header">
          <div>
            <h3>{{ t('marketplace.management.buildingChats.inviteDialogTitle') }}</h3>
            <p>{{ t('marketplace.management.buildingChats.inviteDialogDescription') }}</p>
          </div>
          <button
            type="button"
            class="management-dialog__close"
            :aria-label="t('common.action.close')"
            @click="closeInviteDialog"
          >
            <AppIcon
              name="close"
              :size="18"
            />
          </button>
        </header>

        <div class="management-dialog__content">
          <div class="management-form-field">
            <label>{{ t('marketplace.management.buildingChats.searchUser') }}</label>
            <div class="management-search-row">
              <input
                v-model="searchKeyword"
                type="text"
                class="management-input"
                :placeholder="t('marketplace.management.buildingChats.searchUserPlaceholder')"
              />
              <button
                type="button"
                class="management-secondary-button"
                :disabled="loadingUsers"
                @click="searchUsers"
              >
                {{ t('common.action.search') }}
              </button>
            </div>
          </div>

          <div class="management-form-field">
            <label>{{ t('marketplace.management.buildingChats.selectUser') }}</label>
            <select
              v-model="selectedUserId"
              class="management-select"
              :disabled="loadingUsers"
            >
              <option value="">
                {{ t('marketplace.management.buildingChats.selectUserPlaceholder') }}
              </option>
              <option
                v-for="option in userOptions"
                :key="option.value"
                :value="option.value"
              >
                {{ option.label }} - {{ option.subtitle }}
              </option>
            </select>
          </div>

          <p
            v-if="loadingUsers"
            class="management-dialog__loading"
          >
            {{ t('marketplace.management.buildingChats.loadingUsers') }}
          </p>
        </div>

        <footer class="management-dialog__actions">
          <button
            type="button"
            class="management-secondary-button"
            :disabled="inviting"
            @click="closeInviteDialog"
          >
            {{ t('common.action.cancel') }}
          </button>
          <button
            type="button"
            class="management-primary-button"
            :disabled="!canInvite || inviting"
            @click="submitInvite"
          >
            {{ inviting ? t('marketplace.management.buildingChats.inviting') : t('marketplace.management.buildingChats.inviteUser') }}
          </button>
        </footer>
      </div>
    </div>
  </div>
</template>

<style scoped>
.management-work-panel {
  display: grid;
  gap: 1rem;
  align-content: start;
}

.management-loading,
.management-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  min-height: 200px;
  padding: 40px 20px;
  color: rgb(var(--color-text-muted));
  text-align: center;
}

.management-empty strong {
  color: rgb(var(--color-text));
  font-size: 16px;
  font-weight: 700;
}

.management-empty span {
  font-size: 13px;
}

.management-section-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
  padding: 0 20px;
}

.management-section-title {
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 18px;
  font-weight: 700;
  line-height: 1.3;
}

.management-section-subtitle {
  margin: 6px 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 13px;
  line-height: 1.5;
}

.management-table-container {
  overflow-x: auto;
}

.management-table {
  width: 100%;
  border-collapse: collapse;
}

.management-table thead {
  background: rgb(var(--color-surface-raised));
  border-bottom: 2px solid rgb(var(--color-border));
}

.management-table th {
  padding: 12px 16px;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  font-weight: 700;
  line-height: 1.5;
  text-align: left;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.management-table td {
  padding: 14px 16px;
  border-bottom: 1px solid rgb(var(--color-border));
  color: rgb(var(--color-text));
  font-size: 13px;
  line-height: 1.5;
}

.management-table tbody tr:hover {
  background: rgb(var(--color-surface-raised));
}

.management-table tbody tr.selected {
  background: rgb(var(--color-primary) / 0.05);
}

.management-table-text {
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.management-user-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.management-user-cell strong {
  color: rgb(var(--color-text));
  font-weight: 700;
}

.management-user-cell span {
  color: rgb(var(--color-text-muted));
  font-size: 11px;
}

.management-badge {
  display: inline-flex;
  min-width: 24px;
  height: 24px;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
  font-size: 11px;
  font-weight: 700;
  padding: 0 8px;
}

.management-status-badge {
  display: inline-flex;
  min-height: 24px;
  align-items: center;
  border-radius: 12px;
  background: rgb(var(--color-surface-raised));
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  font-weight: 700;
  padding: 4px 10px;
  margin-right: 6px;
}

.management-status-badge.active {
  background: rgb(34 197 94 / 0.15);
  color: rgb(34 197 94);
}

.management-status-badge.banned {
  background: rgb(239 68 68 / 0.15);
  color: rgb(239 68 68);
}

.management-status-badge.muted {
  background: rgb(251 146 60 / 0.15);
  color: rgb(251 146 60);
}

.management-action-group {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.management-action-button {
  min-height: 32px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 4px;
  background: rgb(var(--color-surface));
  padding: 6px 12px;
  color: rgb(var(--color-text));
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 700;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease;
}

.management-action-button:hover {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary) / 0.05);
}

.management-action-button--danger {
  border-color: rgb(239 68 68 / 0.3);
  color: rgb(239 68 68);
}

.management-action-button--danger:hover {
  border-color: rgb(239 68 68);
  background: rgb(239 68 68 / 0.1);
}

.management-primary-button {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 38px;
  border: 1px solid rgb(var(--color-primary));
  border-radius: 6px;
  background: rgb(var(--color-primary));
  padding: 8px 16px;
  color: rgb(var(--color-primary-contrast));
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  font-weight: 700;
  transition:
    background-color 0.2s ease,
    transform 0.1s ease;
}

.management-primary-button:hover:not(:disabled) {
  background: rgb(var(--color-text));
  transform: translateY(-1px);
}

.management-primary-button:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.management-secondary-button {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 38px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: rgb(var(--color-surface));
  padding: 8px 16px;
  color: rgb(var(--color-text));
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  font-weight: 700;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease;
}

.management-secondary-button:hover:not(:disabled) {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary) / 0.05);
}

.management-secondary-button:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.management-modal {
  position: fixed;
  inset: 0;
  z-index: 60;
  display: grid;
  place-items: center;
  background: rgb(26 28 27 / 0.5);
  padding: 20px;
}

.management-dialog {
  width: min(100%, 560px);
  max-height: calc(100svh - 40px);
  overflow-y: auto;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  box-shadow: 0 20px 60px rgb(0 0 0 / 0.3);
}

.management-dialog__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  border-bottom: 1px solid rgb(var(--color-border));
  padding: 20px 24px;
}

.management-dialog__header h3 {
  margin: 0;
  color: rgb(var(--color-primary));
  font-size: 20px;
  font-weight: 700;
  line-height: 1.3;
}

.management-dialog__header p {
  margin: 6px 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 13px;
  line-height: 1.5;
}

.management-dialog__close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  border: 1px solid rgb(var(--color-border));
  border-radius: 4px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-primary));
  cursor: pointer;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease;
}

.management-dialog__close:hover {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-surface-muted));
}

.management-dialog__content {
  padding: 20px 24px;
}

.management-dialog__loading {
  margin: 16px 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 13px;
  text-align: center;
}

.management-form-field {
  margin-bottom: 16px;
}

.management-form-field:last-child {
  margin-bottom: 0;
}

.management-form-field label {
  display: block;
  margin-bottom: 8px;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.management-search-row {
  display: flex;
  gap: 10px;
}

.management-input,
.management-select {
  width: 100%;
  min-height: 42px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: rgb(var(--color-surface));
  padding: 10px 14px;
  color: rgb(var(--color-text));
  font-family: inherit;
  font-size: 14px;
  transition: border-color 0.2s ease;
}

.management-input:focus,
.management-select:focus {
  outline: none;
  border-color: rgb(var(--color-primary));
}

.management-dialog__actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  border-top: 1px solid rgb(var(--color-border));
  padding: 20px 24px;
}

@media (max-width: 767px) {
  .management-section-header {
    flex-direction: column;
  }

  .management-primary-button {
    width: 100%;
    justify-content: center;
  }

  .management-dialog__actions {
    flex-direction: column;
  }

  .management-secondary-button,
  .management-primary-button {
    width: 100%;
  }

  .management-search-row {
    flex-direction: column;
  }
}
</style>
