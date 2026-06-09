<!--
 * 聊天頁。
 * 1. 串接真實會話列表、詳情、訊息與已讀流程。
 * 2. 保持左欄會話與右欄商品上下文的高資訊密度布局。
-->
<script setup lang="ts">
import AppIcon from '@/shared/components/base/AppIcon.vue';
import BaseAvatar from '@/shared/components/base/BaseAvatar.vue';
import BaseButton from '@/shared/components/base/BaseButton.vue';
import BaseTextarea from '@/shared/components/base/BaseTextarea.vue';
import BaseEmpty from '@/shared/components/feedback/BaseEmpty.vue';

import { useFurnitureChatPage } from './chat';
import EmojiPicker from './widgets/EmojiPicker.vue';
import MessageBubble from './widgets/MessageBubble.vue';

const {
  activeConversation,
  activeMessages,
  chatEmojiOptions,
  conversations,
  draftMessage,
  emojiPickerOpen,
  formatPrice,
  handleSelectChat,
  handleSendByEnter,
  handleSendMessage,
  insertEmoji,
  isSystemNoticeConversation,
  loadingConversations,
  loadingMessages,
  preferenceStore,
  selectedChatId,
  setMessageContainerRef,
  sendingMessage,
  sessionStore,
  t,
  toggleEmojiPicker,
} = useFurnitureChatPage();
</script>

<template>
  <div class="marketplace-chat-page">
    <section class="chat-shell">
      <div
        v-if="sessionStore.isAuthenticated"
        class="chat-layout"
      >
        <div class="panel-surface chat-panel chat-panel--conversations">
          <div class="chat-panel__header chat-panel__header--compact">
            <p class="chat-panel__eyebrow">
              {{ t('chat.conversations') }}
            </p>
          </div>
          <div
            v-if="loadingConversations"
            class="chat-loading"
          >
            {{ t('chat.loadingConversations') }}
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
                        {{ conversation.type === 'system_notice' ? t('chat.systemNoticeSubtitle') : conversation.listing.title }}
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

        <div class="panel-surface chat-panel chat-thread">
          <template v-if="activeConversation">
            <div class="chat-panel__header">
              <div class="chat-thread-header">
                <div class="chat-thread-header__main">
                  <p class="chat-thread-header__title">
                    {{ activeConversation.peer.display_name }}
                  </p>
                  <p class="chat-thread-header__subtitle">
                    {{ isSystemNoticeConversation ? t('chat.systemNoticeSubtitle') : activeConversation.listing.title }}
                  </p>
                </div>
                <div
                  v-if="!isSystemNoticeConversation"
                  class="chat-listing-summary"
                >
                  <p class="chat-listing-summary__label">
                    {{ t('chat.listingPrice') }}
                  </p>
                  <p class="chat-listing-summary__price">
                    {{ formatPrice(activeConversation.listing.price_hkd, preferenceStore.locale) }}
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

            <div
              v-if="!isSystemNoticeConversation"
              class="chat-composer"
            >
              <div class="chat-composer__row">
                <BaseTextarea
                  :model-value="draftMessage"
                  :rows="2"
                  class="chat-composer__textarea"
                  :placeholder="t('chat.composePlaceholder')"
                  @update:model-value="draftMessage = $event"
                  @keydown.enter="handleSendByEnter"
                />
                <div class="chat-composer__actions">
                  <EmojiPicker
                    :emojis="chatEmojiOptions"
                    :open="emojiPickerOpen"
                    :t="t"
                    @select="insertEmoji"
                    @toggle="toggleEmojiPicker"
                  />
                  <BaseButton
                    variant="primary"
                    size="sm"
                    class="chat-send-button"
                    :disabled="sendingMessage"
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
  gap: 1.25rem;
  align-items: stretch;
}

.chat-panel {
  min-height: clamp(34rem, calc(100vh - 12rem), 46rem);
  overflow: hidden;
}

.chat-panel--conversations {
  display: flex;
  flex-direction: column;
}

.chat-panel__header {
  border-bottom: 1px solid rgb(var(--color-border) / 0.82);
  padding: 1.15rem 1.25rem;
}

.chat-panel__header--compact {
  padding-block: 1rem;
}

.chat-panel__eyebrow {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  line-height: 1.3;
  text-transform: uppercase;
}

.chat-loading {
  color: rgb(var(--color-text-muted));
  font-size: 0.88rem;
  line-height: 1.6;
  padding: 1.25rem;
}

.chat-conversation-list {
  display: flex;
  flex: 1;
  min-height: 0;
  flex-direction: column;
  gap: 0.5rem;
  overflow-y: auto;
  padding: 0.75rem;
}

.chat-conversation-card {
  width: 100%;
  border: 1px solid transparent;
  border-radius: 0.75rem;
  padding: 0.85rem;
  text-align: left;
  transition:
    border-color 0.2s ease,
    background 0.2s ease,
    box-shadow 0.2s ease;
}

.chat-conversation-card--active {
  border-color: rgb(var(--color-primary) / 0.34);
  background: rgb(var(--color-primary) / 0.1);
  box-shadow: inset 0 0 0 1px rgb(var(--color-primary) / 0.08);
}

.chat-conversation-card--idle:hover {
  border-color: rgb(var(--color-border) / 0.86);
  background: rgb(var(--color-surface-raised));
}

.chat-conversation-card__content {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
}

.chat-conversation-card__message {
  display: -webkit-box;
  margin: 0.7rem 0 0;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  color: rgb(var(--color-text-muted));
  font-size: 0.88rem;
  line-height: 1.55;
}

.chat-unread-badge {
  display: inline-flex;
  min-width: 1.45rem;
  height: 1.45rem;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
  font-size: 0.68rem;
  font-weight: 700;
  line-height: 1;
  padding-inline: 0.38rem;
}

.chat-thread {
  display: flex;
  flex-direction: column;
}

.chat-thread-header {
  display: grid;
  gap: 0.9rem;
}

.chat-thread-header__main {
  min-width: 0;
}

.chat-thread-header__title {
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 1.05rem;
  font-weight: 800;
  line-height: 1.35;
}

.chat-thread-header__subtitle {
  margin: 0.25rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.88rem;
  line-height: 1.5;
}

.chat-listing-summary {
  min-width: 11rem;
  border: 1px solid rgb(var(--color-border) / 0.62);
  border-radius: 0.75rem;
  background: rgb(var(--color-surface-raised));
  padding: 0.8rem 0.95rem;
  text-align: left;
}

.chat-listing-summary__label {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  line-height: 1.3;
  text-transform: uppercase;
}

.chat-listing-summary__price {
  margin: 0.35rem 0 0;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: 1.35rem;
  font-weight: 800;
  line-height: 1.15;
}

.chat-message-scroll {
  display: flex;
  flex: 1;
  min-height: 0;
  flex-direction: column;
  gap: 0.8rem;
  overflow-y: auto;
  padding: 1.1rem;
}

.chat-composer {
  border-top: 1px solid rgb(var(--color-border) / 0.82);
  padding: 1rem 1.1rem;
}

.chat-composer__row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 0.75rem;
  align-items: end;
}

.chat-composer__textarea {
  min-height: 4.35rem;
  resize: none;
}

.chat-composer__actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.6rem;
}

.chat-send-button {
  min-height: 2.75rem;
  padding-inline: 0.95rem;
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
