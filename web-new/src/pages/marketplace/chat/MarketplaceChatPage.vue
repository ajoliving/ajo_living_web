<!--
 * 聊天頁。
 * 1. 串接真實會話列表、詳情、訊息與已讀流程。
 * 2. 左側會話列表採用與通知導航一致的按鈕樣式。
 * 3. 頂部提供群聊/私聊篩選器。
 * 4. 默認自動打開第一個對話。
-->
<script setup lang="ts">
import AppIcon from '@/shared/components/base/AppIcon.vue';
import BaseAvatar from '@/shared/components/base/BaseAvatar.vue';
import BaseButton from '@/shared/components/base/BaseButton.vue';
import BaseTextarea from '@/shared/components/base/BaseTextarea.vue';
import BaseEmpty from '@/shared/components/feedback/BaseEmpty.vue';

import { useMarketplaceChatPage } from './chat';
import MessageBubble from './widgets/MessageBubble.vue';

const {
  activeConversation,
  activeMembers,
  activeJoinRequests,
  activeReferencePrice,
  activeMessages,
  canSendMessage,
  canManageGroup,
  conversationFilter,
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
  sendingMessage,
  sessionStore,
  t,
} = useMarketplaceChatPage();
</script>

<template>
  <div class="marketplace-chat-page">
    <section class="chat-shell">
      <div
        v-if="sessionStore.isAuthenticated"
        class="chat-layout"
      >
        <div class="chat-sidebar">
          <header class="chat-sidebar__header">
            <h2>{{ t('chat.title') }}</h2>
            <div class="chat-filter-tabs">
              <button
                type="button"
                class="chat-filter-tab"
                :class="conversationFilter === 'all' ? 'active' : ''"
                @click="setConversationFilter('all')"
              >
                {{ t('chat.filterAll') }}
              </button>
              <button
                type="button"
                class="chat-filter-tab"
                :class="conversationFilter === 'building_group' ? 'active' : ''"
                @click="setConversationFilter('building_group')"
              >
                {{ t('chat.filterBuilding') }}
              </button>
            </div>
          </header>

          <div
            v-if="loadingConversations"
            class="chat-sidebar__loading"
          >
            {{ t('chat.loadingConversations') }}
          </div>

          <div
            v-else-if="filteredConversations.length === 0"
            class="chat-sidebar__empty"
          >
            <p>{{ t('chat.noConversationsTitle') }}</p>
            <span>{{ t('chat.noConversationsDescription') }}</span>
          </div>

          <nav
            v-else
            class="chat-conversation-list"
          >
            <button
              v-for="conversation in filteredConversations"
              :key="conversation.id"
              type="button"
              class="chat-nav-item"
              :class="selectedChatId === conversation.id ? 'active' : ''"
              @click="handleSelectChat(conversation.id)"
            >
              <div class="chat-nav-item__avatar">
                <BaseAvatar
                  :src="conversation.peer.avatar_url"
                  :name="conversation.peer.display_name"
                  :size="40"
                />
              </div>
              <div class="chat-nav-item__content">
                <div class="chat-nav-item__header">
                  <h4 class="chat-nav-item__title">{{ conversation.title }}</h4>
                  <span
                    v-if="conversation.unread_count > 0"
                    class="chat-nav-item__badge"
                  >
                    {{ conversation.unread_count }}
                  </span>
                </div>
                <p
                  v-if="conversation.peer.role_in_chat"
                  class="chat-nav-item__role"
                >
                  {{ conversation.peer.role_in_chat }}
                </p>
                <p class="chat-nav-item__subtitle">
                  {{ conversation.subtitle }}
                </p>
                <p class="chat-nav-item__message">
                  {{ conversation.last_message }}
                </p>
              </div>
            </button>
          </nav>
        </div>

        <div
          v-if="activeConversation"
          class="chat-main"
        >
          <header class="chat-main__header">
            <div class="chat-thread-identity">
              <button
                type="button"
                class="chat-back-button"
                :aria-label="t('chat.backToConversations')"
                :title="t('chat.backToConversations')"
                @click="handleBackToChats"
              >
                <AppIcon
                  name="arrow-left"
                  :size="18"
                />
              </button>
              <div class="chat-thread-info">
                <h3 class="chat-thread-title">{{ activeConversation.title }}</h3>
                <p class="chat-thread-subtitle">{{ activeConversation.subtitle }}</p>
              </div>
            </div>
            <div
              v-if="activeReferencePrice > 0"
              class="chat-listing-price"
            >
              <span class="chat-listing-price__label">{{ t('chat.listingPrice') }}</span>
              <strong class="chat-listing-price__value">
                {{ formatPrice(activeReferencePrice, preferenceStore.locale) }}
              </strong>
            </div>
          </header>

          <div
            v-if="activeConversation.type === 'building_group'"
            class="chat-group-panel"
          >
            <div class="chat-group-panel__header">
              <span>{{ t('chat.groupMembers') }} {{ activeMembers.length }}</span>
              <BaseButton
                variant="ghost"
                size="sm"
                @click="handleLeaveBuildingChat"
              >
                {{ t('chat.leaveGroup') }}
              </BaseButton>
            </div>
            <div class="chat-group-members">
              <div
                v-for="member in activeMembers"
                :key="member.user_id"
                class="chat-group-member"
              >
                <div class="chat-group-member__info">
                  <strong>{{ member.display_name }}</strong>
                  <span>{{ member.role_in_chat }} · {{ member.membership_status }}</span>
                </div>
                <div
                  v-if="canManageGroup"
                  class="chat-group-member__actions"
                >
                  <button
                    v-if="member.membership_status === 'active' && !member.muted_until"
                    type="button"
                    @click="handleModerateBuildingMember(member.user_id, 'mute')"
                  >
                    {{ t('chat.mute') }}
                  </button>
                  <button
                    v-if="member.membership_status === 'active' && member.muted_until"
                    type="button"
                    @click="handleModerateBuildingMember(member.user_id, 'unmute')"
                  >
                    {{ t('chat.unmute') }}
                  </button>
                  <button
                    v-if="member.membership_status === 'active'"
                    type="button"
                    @click="handleModerateBuildingMember(member.user_id, 'kick')"
                  >
                    {{ t('chat.kick') }}
                  </button>
                  <button
                    v-if="member.membership_status === 'active'"
                    type="button"
                    @click="handleModerateBuildingMember(member.user_id, 'ban')"
                  >
                    {{ t('chat.ban') }}
                  </button>
                  <button
                    v-if="member.membership_status === 'banned'"
                    type="button"
                    @click="handleModerateBuildingMember(member.user_id, 'unban')"
                  >
                    {{ t('chat.unban') }}
                  </button>
                </div>
              </div>
            </div>
            <div
              v-if="canManageGroup && activeJoinRequests.length"
              class="chat-group-requests"
            >
              <strong>{{ t('chat.joinRequests') }}</strong>
              <div
                v-for="request in activeJoinRequests"
                :key="request.request_id"
                class="chat-group-request"
              >
                <span>{{ request.user_id }}<template v-if="request.reason"> · {{ request.reason }}</template></span>
                <div class="chat-group-member__actions">
                  <button type="button" @click="handleReviewBuildingChatJoin(request.request_id, 'approved')">
                    {{ t('chat.approve') }}
                  </button>
                  <button type="button" @click="handleReviewBuildingChatJoin(request.request_id, 'rejected')">
                    {{ t('chat.reject') }}
                  </button>
                </div>
              </div>
            </div>
          </div>

          <div
            :ref="setMessageContainerRef"
            class="chat-message-scroll"
          >
            <div
              v-if="loadingMessages"
              class="chat-message-loading"
            >
              {{ t('chat.loadingMessages') }}
            </div>
            <MessageBubble
              v-for="messageItem in activeMessages"
              :key="messageItem.id"
              :message="messageItem"
            />
          </div>

          <div class="chat-composer">
            <div class="chat-composer__meta">
              <span>{{ t('chat.replying') }}</span>
              <span>{{ draftMessage.length }}/{{ messageMaxLength }}</span>
            </div>
            <div class="chat-composer__row">
              <BaseTextarea
                :model-value="draftMessage"
                :rows="2"
                :maxlength="messageMaxLength"
                class="chat-composer__textarea"
                :placeholder="t('chat.composePlaceholder')"
                @update:model-value="draftMessage = $event"
                @keydown.enter="handleSendByEnter"
              />
              <div class="chat-composer__actions">
                <BaseButton
                  variant="primary"
                  size="sm"
                  class="chat-send-button"
                  :disabled="!canSendMessage"
                  @click="handleSendMessage"
                >
                  <template #leading>
                    <AppIcon
                      name="send"
                      :size="16"
                    />
                  </template>
                  {{ sendingMessage ? t('chat.sending') : t('common.action.send') }}
                </BaseButton>
              </div>
            </div>
          </div>
        </div>

        <div
          v-else
          class="chat-main chat-main--empty"
        >
          <BaseEmpty
            :title="t('chat.emptyThreadTitle')"
            :description="t('chat.emptyThreadDescription')"
          />
        </div>
      </div>

      <BaseEmpty
        v-else
        :title="t('chat.memberRequiredTitle')"
        :description="t('chat.memberRequiredDescription')"
      />
    </section>
  </div>
</template>

<style scoped>
.marketplace-chat-page {
  width: 100%;
  min-width: 0;
}

.chat-shell {
  width: 100%;
  min-width: 0;
}

.chat-layout {
  display: grid;
  grid-template-columns: 260px minmax(0, 1fr);
  gap: 18px;
  width: 100%;
  max-width: none;
  align-items: stretch;
}

.chat-sidebar {
  display: flex;
  flex-direction: column;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  box-shadow: 0 1px 3px rgb(0 0 0 / 0.05);
  overflow: hidden;
}

.chat-sidebar__header {
  padding: 16px 18px;
  border-bottom: 1px solid rgb(var(--color-border));
}

.chat-sidebar__header h2 {
  margin: 0 0 12px;
  color: rgb(var(--color-text));
  font-size: 18px;
  font-weight: 700;
  line-height: 1.3;
}

.chat-filter-tabs {
  display: flex;
  gap: 6px;
  border-radius: 6px;
  background: rgb(var(--color-surface-raised));
  padding: 4px;
}

.chat-filter-tab {
  flex: 1;
  min-height: 32px;
  border: 0;
  border-radius: 4px;
  background: transparent;
  padding: 6px 10px;
  color: rgb(var(--color-text-muted));
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 700;
  transition:
    background-color 0.2s ease,
    color 0.2s ease;
}

.chat-filter-tab:hover {
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
}

.chat-filter-tab.active {
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.chat-sidebar__loading,
.chat-sidebar__empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  min-height: 200px;
  padding: 20px;
  color: rgb(var(--color-text-muted));
  font-size: 13px;
  text-align: center;
}

.chat-sidebar__empty p {
  margin: 0;
  color: rgb(var(--color-text));
  font-weight: 700;
}

.chat-sidebar__empty span {
  font-size: 12px;
}

.chat-conversation-list {
  display: grid;
  gap: 2px;
  padding: 8px;
  overflow-y: auto;
}

.chat-nav-item {
  display: grid;
  grid-template-columns: 40px minmax(0, 1fr);
  gap: 12px;
  align-items: start;
  min-height: 68px;
  border: 0;
  border-radius: 6px;
  background: transparent;
  padding: 12px;
  color: rgb(var(--color-text));
  cursor: pointer;
  font-family: inherit;
  text-align: left;
  transition:
    background-color 0.2s ease,
    transform 0.1s ease;
}

.chat-nav-item:hover {
  background: rgb(var(--color-surface-raised));
}

.chat-nav-item.active {
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.chat-nav-item__content {
  min-width: 0;
}

.chat-nav-item__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 4px;
}

.chat-nav-item__title {
  margin: 0;
  min-width: 0;
  overflow: hidden;
  color: inherit;
  font-size: 13px;
  font-weight: 700;
  line-height: 1.4;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chat-nav-item__badge {
  display: inline-flex;
  min-width: 20px;
  height: 20px;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border-radius: 10px;
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
  font-size: 11px;
  font-weight: 700;
  padding: 0 6px;
}

.chat-nav-item.active .chat-nav-item__badge {
  background: rgb(var(--color-primary-contrast));
  color: rgb(var(--color-primary));
}

.chat-nav-item__role {
  margin: 0 0 2px;
  color: inherit;
  font-size: 11px;
  font-weight: 600;
  line-height: 1.3;
  opacity: 0.8;
}

.chat-nav-item__subtitle {
  margin: 0 0 4px;
  overflow: hidden;
  color: inherit;
  font-size: 11px;
  line-height: 1.4;
  opacity: 0.75;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chat-nav-item__message {
  margin: 0;
  overflow: hidden;
  color: inherit;
  font-size: 12px;
  line-height: 1.4;
  opacity: 0.7;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
}

.chat-main {
  display: grid;
  grid-template-rows: auto auto minmax(0, 1fr) auto;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  box-shadow: 0 1px 3px rgb(0 0 0 / 0.05);
  min-height: clamp(500px, calc(100svh - 200px), 700px);
  overflow: hidden;
}

.chat-main--empty {
  display: flex;
  align-items: center;
  justify-content: center;
}

.chat-main__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  border-bottom: 1px solid rgb(var(--color-border));
  padding: 16px 20px;
}

.chat-thread-identity {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.chat-back-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  flex-shrink: 0;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  cursor: pointer;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease;
}

.chat-back-button:hover {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary) / 0.05);
}

.chat-thread-info {
  min-width: 0;
}

.chat-thread-title {
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 16px;
  font-weight: 700;
  line-height: 1.3;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chat-thread-subtitle {
  margin: 4px 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chat-listing-price {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex-shrink: 0;
  text-align: right;
}

.chat-listing-price__label {
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  font-weight: 600;
  line-height: 1;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.chat-listing-price__value {
  color: rgb(var(--color-primary));
  font-size: 16px;
  font-weight: 700;
  line-height: 1.2;
}

.chat-group-panel {
  border-bottom: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface-raised));
  padding: 12px 20px;
}

.chat-group-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  font-weight: 700;
}

.chat-group-members {
  display: grid;
  gap: 8px;
}

.chat-group-member {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border-radius: 4px;
  background: rgb(var(--color-surface));
  padding: 10px 12px;
}

.chat-group-member__info {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.chat-group-member__info strong {
  color: rgb(var(--color-text));
  font-size: 13px;
  font-weight: 700;
}

.chat-group-member__info span {
  color: rgb(var(--color-text-muted));
  font-size: 11px;
}

.chat-group-member__actions {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

.chat-group-member__actions button {
  min-height: 28px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 4px;
  background: rgb(var(--color-surface));
  padding: 4px 10px;
  color: rgb(var(--color-text));
  cursor: pointer;
  font-family: inherit;
  font-size: 11px;
  font-weight: 700;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease;
}

.chat-group-member__actions button:hover {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary) / 0.05);
}

.chat-group-requests {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid rgb(var(--color-border));
}

.chat-group-requests strong {
  display: block;
  margin-bottom: 8px;
  color: rgb(var(--color-text));
  font-size: 12px;
  font-weight: 700;
}

.chat-group-request {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border-radius: 4px;
  background: rgb(var(--color-surface));
  padding: 10px 12px;
  margin-bottom: 8px;
}

.chat-group-request span {
  color: rgb(var(--color-text-muted));
  font-size: 12px;
}

.chat-message-scroll {
  display: flex;
  flex-direction: column;
  gap: 12px;
  overflow-y: auto;
  padding: 20px;
  background: rgb(var(--color-surface-muted));
}

.chat-message-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 200px;
  color: rgb(var(--color-text-muted));
  font-size: 13px;
}

.chat-composer {
  border-top: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface));
  padding: 16px 20px;
}

.chat-composer__meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  font-weight: 600;
}

.chat-composer__row {
  display: flex;
  gap: 12px;
  align-items: flex-end;
}

.chat-composer__textarea {
  flex: 1;
  min-width: 0;
}

.chat-composer__actions {
  flex-shrink: 0;
}

.chat-send-button {
  min-height: 40px;
}

@media (max-width: 1023px) {
  .chat-layout {
    grid-template-columns: 1fr;
    gap: 0;
  }

  .chat-sidebar {
    border-radius: 0;
    border-left: 0;
    border-right: 0;
  }

  .chat-main {
    border-radius: 0;
    border-left: 0;
    border-right: 0;
    border-top: 0;
  }

  .chat-back-button {
    display: none;
  }
}

@media (max-width: 767px) {
  .chat-main__header {
    flex-direction: column;
    align-items: flex-start;
  }

  .chat-listing-price {
    text-align: left;
  }

  .chat-composer__row {
    flex-direction: column;
    align-items: stretch;
  }

  .chat-send-button {
    width: 100%;
  }
}

:deep(.app-button),
:deep(.app-input-shell),
:deep(.app-textarea-shell) {
  border-radius: 6px;
}
</style>
