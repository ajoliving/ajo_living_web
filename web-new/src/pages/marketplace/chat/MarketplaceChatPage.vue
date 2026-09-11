<!--
 * 聊天頁。
 * 1. 串接真實會話列表、詳情、訊息與已讀流程，一對一與大廈群聊共用同一佈局。
 * 2. 中間會話列表與通信中心導航共用一致的選中與聚焦樣式，列表項之間以留白分隔線區分。
 * 3. 大廈群聊於標題欄提供「更多」選單，收納成員名單與離開群組操作。
-->
<script setup lang="ts">
import { ref, watch } from 'vue';

import AppIcon from '@/shared/components/base/AppIcon.vue';
import AppActionConfirmDialog from '@/shared/components/base/AppActionConfirmDialog.vue';
import BaseAvatar from '@/shared/components/base/BaseAvatar.vue';
import BaseTextarea from '@/shared/components/base/BaseTextarea.vue';
import BaseEmpty from '@/shared/components/feedback/BaseEmpty.vue';

import { useMarketplaceChatPage } from './chat';
import MessageBubble from './widgets/MessageBubble.vue';
import GroupMembersDialog from './widgets/GroupMembersDialog.vue';

const emit = defineEmits<{
  (event: 'unread-count-change', count: number): void;
}>();

const {
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
  handleSelectAttachment,
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
  sessionStore,
  t,
  totalUnreadCount,
} = useMarketplaceChatPage();

const moreMenuOpen = ref(false);

// 1. 關閉群聊更多選單
const closeMoreMenu = (): void => {
  moreMenuOpen.value = false;
};

// 2. 從更多選單開啟成員名單
const openGroupMembers = async (): Promise<void> => {
  closeMoreMenu();
  await handleOpenGroupMembers();
};

// 3. 從更多選單開啟離開群組確認
const openLeaveConfirm = (): void => {
  closeMoreMenu();
  leaveConfirmOpen.value = true;
};

// 4. 上報會話未讀總數給通信中心導航
watch(
  totalUnreadCount,
  (count) => {
    emit('unread-count-change', count);
  },
  { immediate: true },
);
</script>

<template>
  <div class="marketplace-chat-page">
    <section class="chat-shell">
      <div
        v-if="sessionStore.isAuthenticated"
        class="chat-layout"
        :class="activeConversation ? 'chat-layout--thread-open' : ''"
      >
        <div class="chat-sidebar">
          <div
            v-if="loadingConversations"
            class="chat-sidebar__loading"
          >
            {{ t('chat.loadingConversations') }}
          </div>

          <div
            v-else-if="conversations.length === 0"
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
              v-for="conversation in conversations"
              :key="conversation.id"
              type="button"
              class="chat-nav-item"
              :class="selectedChatId === conversation.id ? 'active' : ''"
              @click="handleSelectChat(conversation.id)"
            >
              <div class="chat-nav-item__avatar">
                <BaseAvatar
                  v-if="conversation.type === 'direct_listing_chat'"
                  :src="conversation.peer.avatar_url"
                  :name="conversation.peer.display_name"
                  :size="40"
                />
                <div
                  v-else
                  class="chat-nav-item__building-icon"
                >
                  <AppIcon
                    name="building"
                    :size="20"
                  />
                </div>
              </div>
              <div class="chat-nav-item__content">
                <div class="chat-nav-item__header">
                  <h4 class="chat-nav-item__title">
                    <span
                      v-if="conversation.type === 'building_group'"
                      class="chat-nav-item__group-badge"
                    >{{ t('chat.buildingGroupBadge') }}</span>
                    <span class="chat-nav-item__title-text">{{ conversation.title }}</span>
                  </h4>
                  <span
                    v-if="conversation.unread_count > 0"
                    class="chat-nav-item__badge"
                  >
                    {{ conversation.unread_count }}
                  </span>
                </div>
                <p
                  v-if="conversation.type === 'building_group'"
                  class="chat-nav-item__subtitle chat-nav-item__subtitle--group"
                >
                  <AppIcon
                    name="user"
                    :size="12"
                  />
                  {{ conversation.member_count }} {{ t('chat.members') }}
                </p>
                <p
                  v-else-if="conversation.peer.role_in_chat"
                  class="chat-nav-item__role"
                >
                  {{ conversation.peer.role_in_chat }}
                </p>
                <p
                  v-if="conversation.type === 'direct_listing_chat'"
                  class="chat-nav-item__subtitle"
                >
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
                <p
                  v-if="activeConversation.type === 'building_group'"
                  class="chat-thread-subtitle"
                >
                  <AppIcon
                    name="user"
                    :size="12"
                  />
                  {{ activeConversation.member_count }} {{ t('chat.members') }}
                </p>
                <p
                  v-else
                  class="chat-thread-subtitle"
                >
                  {{ activeConversation.subtitle }}
                </p>
              </div>
            </div>
            <div class="chat-header-actions">
              <div
                v-if="activeConversation.type === 'direct_listing_chat' && activeReferencePrice > 0"
                class="chat-listing-price"
              >
                <span class="chat-listing-price__label">{{ t('chat.listingPrice') }}</span>
                <strong class="chat-listing-price__value">
                  {{ formatPrice(activeReferencePrice, preferenceStore.locale) }}
                </strong>
              </div>

              <div
                v-if="activeConversation.type === 'building_group'"
                class="chat-more"
              >
                <button
                  type="button"
                  class="chat-more__button"
                  :class="moreMenuOpen ? 'open' : ''"
                  :aria-label="t('chat.moreActions')"
                  :aria-expanded="moreMenuOpen"
                  @click="moreMenuOpen = !moreMenuOpen"
                >
                  <AppIcon
                    name="more"
                    :size="20"
                  />
                </button>
                <Teleport to="body">
                  <div
                    v-if="moreMenuOpen"
                    class="chat-more__backdrop"
                    @click="closeMoreMenu"
                  />
                </Teleport>
                <div
                  v-if="moreMenuOpen"
                  class="chat-more__menu"
                >
                  <button
                    type="button"
                    class="chat-more__menu-item"
                    @click="openGroupMembers"
                  >
                    <AppIcon
                      name="user"
                      :size="16"
                    />
                    {{ t('chat.viewMembers') }}
                  </button>
                  <button
                    type="button"
                    class="chat-more__menu-item chat-more__menu-item--danger"
                    @click="openLeaveConfirm"
                  >
                    <AppIcon
                      name="logout"
                      :size="16"
                    />
                    {{ t('chat.leaveGroup') }}
                  </button>
                </div>
              </div>
            </div>
          </header>

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
              :show-sender-name="activeConversation.type === 'building_group'"
            />
          </div>

          <div class="chat-composer">
            <div class="chat-composer__surface">
              <BaseTextarea
                :model-value="draftMessage"
                :rows="3"
                :maxlength="messageMaxLength"
                class="chat-composer__textarea"
                :placeholder="t('chat.composePlaceholder')"
                @update:model-value="draftMessage = $event"
                @keydown.enter="handleSendByEnter"
              />
              <div class="chat-composer__toolbar">
                <label
                  class="chat-attachment-button"
                  :class="uploadingAttachment ? 'chat-attachment-button--disabled' : ''"
                  :title="t('chat.attachments')"
                >
                  <AppIcon name="plus" :size="22" />
                  <input
                    type="file"
                    accept="image/*,video/*,application/pdf,text/plain"
                    multiple
                    :disabled="uploadingAttachment"
                    @change="handleSelectAttachment"
                  >
                </label>
                <button
                  type="button"
                  class="chat-send-button"
                  :disabled="!canSendMessage || uploadingAttachment"
                  :aria-label="sendingMessage ? t('chat.sending') : t('common.action.send')"
                  :title="sendingMessage ? t('chat.sending') : t('common.action.send')"
                  @click="handleSendMessage"
                >
                  <AppIcon name="send" :size="18" />
                </button>
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

        <GroupMembersDialog
          :open="groupMembersOpen"
          :loading="loadingGroupMembers"
          :members="groupMembers"
          :title="activeConversation?.title || t('chat.groupMembers')"
          :member-count-label="`${groupMembers.length} ${t('chat.members')}`"
          @close="handleCloseGroupMembers"
        />

        <AppActionConfirmDialog
          :open="leaveConfirmOpen"
          :title="t('chat.leaveGroup')"
          :description="t('chat.leaveGroupConfirmDescription')"
          :cancel-label="t('common.action.cancel')"
          :confirm-label="t('chat.leaveGroup')"
          :confirming="leavingGroup"
          @cancel="leaveConfirmOpen = false"
          @confirm="handleLeaveBuildingChat"
        />
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
  height: calc(100svh - 116px);
  min-height: 480px;
  align-items: stretch;
}

.chat-sidebar {
  display: flex;
  flex-direction: column;
  align-self: start;
  height: 50svh;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  box-shadow: 0 1px 3px rgb(0 0 0 / 0.05);
  overflow: hidden;
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
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
  padding: 8px 8px 10px;
  overflow-y: auto;
}

.chat-nav-item {
  position: relative;
  display: grid;
  grid-template-columns: 40px minmax(0, 1fr);
  gap: 12px;
  align-items: center;
  box-sizing: border-box;
  height: 96px;
  min-height: 96px;
  max-height: 96px;
  flex: 0 0 96px;
  border: 0;
  border-radius: 0;
  background: transparent;
  padding: 10px 4px;
  color: rgb(var(--color-text-muted));
  cursor: pointer;
  font-family: inherit;
  text-align: left;
  overflow: hidden;
  transition:
    background-color 0.2s ease,
    transform 0.1s ease;
}

.chat-nav-item::before {
  content: '';
  position: absolute;
  right: 12px;
  bottom: 0;
  left: 12px;
  height: 1px;
  background: rgb(var(--color-border) / 0.55);
}

.chat-nav-item:last-child::before {
  display: none;
}

.chat-nav-item::after {
  content: '';
  position: absolute;
  right: 12px;
  bottom: 0;
  left: 12px;
  height: 2px;
  background: rgb(var(--color-primary));
  transform: scaleX(0);
  transform-origin: left center;
  transition: transform 0.24s ease;
}

.chat-nav-item:hover,
.chat-nav-item:focus-visible {
  background: rgb(var(--color-primary) / 0.08);
  color: rgb(var(--color-text));
  outline: none;
}

.chat-nav-item.active {
  background: transparent;
  color: rgb(var(--color-primary));
  font-weight: 700;
}

.chat-nav-item.active:hover {
  background: rgb(var(--color-primary) / 0.08);
}

.chat-nav-item.active::after,
.chat-nav-item:focus-visible::after {
  transform: scaleX(1);
}

.chat-nav-item__content {
  min-width: 0;
  overflow: hidden;
}

.chat-nav-item__avatar {
  flex-shrink: 0;
}

.chat-nav-item__building-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  flex-shrink: 0;
  border-radius: 50%;
  background: rgb(var(--color-primary) / 0.1);
  color: rgb(var(--color-primary));
}

.chat-nav-item__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 4px;
}

.chat-nav-item__title {
  display: flex;
  align-items: center;
  gap: 6px;
  margin: 0;
  min-width: 0;
  overflow: hidden;
  color: inherit;
  font-size: 13px;
  font-weight: 700;
  line-height: 1.4;
  white-space: nowrap;
}

.chat-nav-item__title-text {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
}

.chat-nav-item__group-badge {
  display: inline-flex;
  align-items: center;
  flex-shrink: 0;
  border-radius: 4px;
  background: rgb(var(--color-primary) / 0.12);
  color: rgb(var(--color-primary));
  font-size: 10px;
  font-weight: 700;
  line-height: 1;
  padding: 3px 5px;
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
  background: rgb(var(--color-primary) / 0.15);
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

.chat-nav-item__subtitle--group {
  display: flex;
  align-items: center;
  gap: 4px;
  color: rgb(var(--color-primary));
  font-weight: 600;
  opacity: 0.9;
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
  grid-template-rows: auto minmax(0, 1fr) auto;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  box-shadow: 0 1px 3px rgb(0 0 0 / 0.05);
  min-height: 0;
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
  padding: 14px 20px;
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
  display: flex;
  align-items: center;
  gap: 4px;
  margin: 4px 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chat-header-actions {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-shrink: 0;
}

.chat-listing-price {
  display: flex;
  flex-direction: column;
  gap: 4px;
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

.chat-more {
  position: relative;
}

.chat-more__button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  cursor: pointer;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease;
}

.chat-more__button:hover,
.chat-more__button.open,
.chat-more__button:focus-visible {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary) / 0.05);
  outline: none;
}

.chat-more__backdrop {
  position: fixed;
  z-index: 110;
  inset: 0;
}

.chat-more__menu {
  position: absolute;
  top: calc(100% + 6px);
  right: 0;
  z-index: 111;
  display: grid;
  min-width: 176px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  box-shadow: 0 12px 32px rgb(15 23 42 / 0.14);
  overflow: hidden;
}

.chat-more__menu-item {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 44px;
  border: 0;
  background: transparent;
  padding: 10px 14px;
  color: rgb(var(--color-text));
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  font-weight: 600;
  text-align: left;
  transition:
    background-color 0.15s ease,
    color 0.15s ease;
}

.chat-more__menu-item:hover,
.chat-more__menu-item:focus-visible {
  background: rgb(var(--color-primary) / 0.06);
  color: rgb(var(--color-primary));
  outline: none;
}

.chat-more__menu-item--danger {
  color: rgb(185 28 28);
}

.chat-more__menu-item--danger:hover,
.chat-more__menu-item--danger:focus-visible {
  background: rgb(239 68 68 / 0.08);
  color: rgb(185 28 28);
}

.chat-message-scroll {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 0;
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
  padding: 12px 16px;
}

.chat-composer__surface {
  display: grid;
  grid-template-rows: minmax(84px, 1fr) 44px;
  min-height: 132px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface-raised));
  overflow: hidden;
}

.chat-composer__textarea {
  min-width: 0;
}

.chat-composer__textarea :deep(.app-textarea) {
  width: 100%;
  min-height: 84px;
  border: 0;
  border-radius: 0;
  background: transparent;
  resize: none;
  padding: 14px 14px 6px;
  box-shadow: none;
}

.chat-composer__textarea :deep(.app-textarea:focus) {
  box-shadow: none;
}

.chat-composer__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 44px;
  padding: 4px 8px 8px;
}

.chat-attachment-button,
.chat-send-button {
  display: inline-flex;
  width: 36px;
  height: 36px;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 5px;
  background: transparent;
  color: rgb(var(--color-text-muted));
  cursor: pointer;
}

.chat-attachment-button:hover,
.chat-attachment-button:focus-within {
  background: rgb(var(--color-primary) / 0.08);
  color: rgb(var(--color-primary));
}

.chat-attachment-button input {
  display: none;
}

.chat-attachment-button--disabled,
.chat-send-button:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.chat-send-button {
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.chat-send-button:not(:disabled):hover,
.chat-send-button:not(:disabled):focus-visible {
  background: rgb(var(--color-primary) / 0.88);
  outline: none;
}


@media (max-width: 1023px) {
  .chat-layout {
    grid-template-columns: 1fr;
    gap: 0;
    height: auto;
    min-height: 0;
  }

  .chat-layout--thread-open .chat-sidebar {
    display: none;
  }

  .chat-layout:not(.chat-layout--thread-open) .chat-main {
    display: none;
  }

  .chat-sidebar {
    border-radius: 0;
    border-left: 0;
    border-right: 0;
    height: 50svh;
  }

  .chat-layout--thread-open {
    height: calc(100svh - var(--app-mobile-content-bottom) - 92px);
    min-height: 420px;
  }

  .chat-main {
    border-radius: 0;
    border-left: 0;
    border-right: 0;
    border-top: 0;
  }

  .chat-more__menu {
    right: 0;
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

  .chat-message-scroll {
    padding: 14px;
  }

  .chat-composer__textarea :deep(.app-textarea) {
    min-height: 84px;
  }
}

:deep(.app-button),
:deep(.app-input-shell) {
  border-radius: 6px;
}
</style>
