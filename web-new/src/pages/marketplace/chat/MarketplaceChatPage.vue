<!--
 * 聊天頁。
 * 1. 串接真實會話列表、詳情、訊息與已讀流程。
 * 2. 保持左欄會話與右欄內容上下文的高資訊密度布局。
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
} = useMarketplaceChatPage();
</script>

<template>
  <div class="marketplace-chat-page">
    <section class="chat-shell">
      <div
        v-if="sessionStore.isAuthenticated"
        class="chat-layout"
      >
        <div class="chat-panel chat-panel--conversations">
          <div class="chat-panel__header chat-panel__header--compact">
            <div>
              <p class="chat-panel__eyebrow">{{ t('chat.kicker') }}</p>
              <h1 class="chat-panel__title">{{ t('chat.title') }}</h1>
            </div>
            <span class="chat-panel__count">{{ conversations.length }}</span>
          </div>
          <div
            v-if="loadingConversations"
            class="chat-loading"
          >
            {{ t('chat.loadingConversations') }}
          </div>
          <div
            v-else-if="conversations.length === 0"
            class="chat-conversation-empty"
          >
            <p>{{ t('chat.noConversationsTitle') }}</p>
            <span>{{ t('chat.noConversationsDescription') }}</span>
          </div>
          <div
            v-else
            class="chat-conversation-list"
          >
            <button
              v-for="conversation in conversations"
              :key="conversation.id"
              type="button"
              class="chat-conversation-card"
              :class="
                selectedChatId === conversation.id
                  ? 'chat-conversation-card--active'
                  : 'chat-conversation-card--idle'
              "
              @click="handleSelectChat(conversation.id)"
            >
              <div class="chat-conversation-card__content">
                <BaseAvatar
                  :src="conversation.peer.avatar_url"
                  :name="conversation.peer.display_name"
                  :size="48"
                />
                <div class="min-w-0 flex-1">
                  <div class="flex items-start justify-between gap-3">
                    <div class="min-w-0">
                      <p class="truncate text-sm font-semibold text-text">
                        {{ conversation.peer.display_name }}
                      </p>
                      <p class="truncate text-xs text-text-muted">
                        {{ conversation.listing.title }}
                      </p>
                    </div>
                    <span
                      v-if="conversation.unread_count > 0"
                      class="chat-unread-badge"
                    >
                      {{ conversation.unread_count }}
                    </span>
                  </div>
                  <p class="chat-conversation-card__message">
                    {{ conversation.last_message }}
                  </p>
                </div>
              </div>
            </button>
          </div>
        </div>

        <div class="chat-panel chat-thread">
          <template v-if="activeConversation">
            <div class="chat-panel__header">
              <div class="chat-thread-header">
                <div class="chat-thread-header__main">
                  <p class="chat-thread-header__title">
                    {{ activeConversation.peer.display_name }}
                  </p>
                  <p class="chat-thread-header__subtitle">
                    {{ activeConversation.listing.title }}
                  </p>
                </div>
                <div
                  v-if="activeReferencePrice > 0"
                  class="chat-listing-summary"
                >
                  <p class="chat-listing-summary__label">{{ t('chat.listingPrice') }}</p>
                  <p class="chat-listing-summary__price">
                    {{ formatPrice(activeReferencePrice, preferenceStore.locale) }}
                  </p>
                </div>
              </div>
            </div>

            <div
              :ref="setMessageContainerRef"
              class="chat-message-scroll"
            >
              <div
                v-if="loadingMessages"
                class="chat-loading"
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
          </template>

          <div
            v-else
            class="flex flex-1 items-center justify-center p-6"
          >
            <BaseEmpty
              :title="t('common.empty.chatsTitle')"
              :description="t('common.empty.chatsDescription')"
            />
          </div>
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
  gap: 1rem;
  align-items: stretch;
}

.chat-panel {
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  min-height: clamp(34rem, calc(100vh - 12rem), 46rem);
  overflow: hidden;
}

.chat-panel--conversations {
  display: flex;
  flex-direction: column;
}

.chat-panel__header {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 12px;
  border-bottom: 1px solid rgb(var(--color-border));
  padding: 0.9rem;
}

.chat-panel__header--compact {
  align-items: center;
}

.chat-panel__eyebrow {
  margin: 0;
  color: rgb(var(--color-primary));
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  line-height: 1.3;
  text-transform: uppercase;
}

.chat-panel__title {
  margin: 6px 0 0;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: 1.55rem;
  font-weight: 500;
  line-height: 1.15;
}

.chat-panel__count {
  display: inline-flex;
  min-width: 32px;
  height: 32px;
  align-items: center;
  justify-content: center;
  border-radius: 2px;
  background: rgb(var(--color-primary) / 0.1);
  color: rgb(var(--color-primary));
  font-size: 12px;
  font-weight: 700;
  padding-inline: 10px;
}

.chat-loading {
  color: rgb(var(--color-text-muted));
  font-size: 0.8125rem;
  line-height: 1.6;
  padding: 1rem 0.9rem;
}

.chat-conversation-empty {
  display: grid;
  gap: 0.35rem;
  color: rgb(var(--color-text-muted));
  padding: 1rem 0.9rem;
}

.chat-conversation-empty p {
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 0.9rem;
  font-weight: 700;
}

.chat-conversation-empty span {
  font-size: 0.8rem;
  line-height: 1.6;
}

.chat-conversation-list {
  display: flex;
  flex: 1;
  min-height: 0;
  flex-direction: column;
  gap: 0.45rem;
  overflow-y: auto;
  padding: 0.7rem;
}

.chat-conversation-card {
  width: 100%;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  padding: 0.7rem;
  text-align: left;
  transition:
    border-color 0.2s ease,
    background 0.2s ease;
}

.chat-conversation-card--active {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary) / 0.1);
}

.chat-conversation-card--idle:hover {
  border-color: rgb(var(--color-primary) / 0.3);
  background: rgb(var(--color-surface-raised));
}

.chat-conversation-card__content {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
}

.chat-conversation-card__message {
  display: -webkit-box;
  margin: 8px 0 0;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  line-height: 1.55;
}

.chat-unread-badge {
  display: inline-flex;
  min-width: 22px;
  height: 22px;
  align-items: center;
  justify-content: center;
  border-radius: 2px;
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
  font-size: 11px;
  font-weight: 700;
  line-height: 1;
  padding-inline: 6px;
}

.chat-thread {
  display: flex;
  flex-direction: column;
}

.chat-thread-header {
  display: grid;
  gap: 12px;
}

.chat-thread-header__main {
  min-width: 0;
}

.chat-thread-header__title {
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 15px;
  font-weight: 700;
  line-height: 1.35;
}

.chat-thread-header__subtitle {
  margin: 4px 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  line-height: 1.5;
}

.chat-listing-summary {
  min-width: 132px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  padding: 0.65rem 0.75rem;
  text-align: left;
}

.chat-listing-summary__label {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.12em;
  line-height: 1.3;
  text-transform: uppercase;
}

.chat-listing-summary__price {
  margin: 6px 0 0;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: 1.25rem;
  font-weight: 500;
  line-height: 1.15;
}

.chat-message-scroll {
  display: flex;
  flex: 1;
  min-height: 0;
  flex-direction: column;
  gap: 0.6rem;
  overflow-y: auto;
  padding: 1rem;
  background: rgb(var(--color-surface-raised));
}

.chat-composer {
  border-top: 1px solid rgb(var(--color-border));
  padding: 0.75rem;
}

.chat-composer__meta {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  margin-bottom: 0.55rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.72rem;
  font-weight: 700;
}

.chat-composer__row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 10px;
  align-items: end;
}

.chat-composer__textarea {
  min-height: 72px;
  resize: none;
}

.chat-composer__actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
}

.chat-send-button {
  min-height: 40px;
  padding-inline: 14px;
  white-space: nowrap;
}

.chat-send-button :deep(span) {
  white-space: nowrap;
}

@media (min-width: 1024px) {
  .chat-thread-header {
    grid-template-columns: minmax(0, 1fr) auto;
    align-items: center;
  }
}

@media (min-width: 1280px) {
  .chat-layout {
    grid-template-columns: 20.5rem minmax(0, 1fr);
  }
}

@media (max-width: 640px) {
  .chat-panel {
    min-height: 32rem;
  }

  .chat-composer__row {
    grid-template-columns: 1fr;
  }

  .chat-composer__actions {
    width: 100%;
  }

  .chat-send-button {
    flex: 1;
  }
}
</style>
