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

import { useMarketplaceChatPage } from './chat';
import MessageBubble from './widgets/MessageBubble.vue';

const {
  activeConversation,
  activeMessages,
  conversations,
  draftMessage,
  formatPrice,
  handleSelectChat,
  handleSendMessage,
  isSystemNoticeConversation,
  loadingConversations,
  loadingMessages,
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
    <section>
      <div
        v-if="sessionStore.isAuthenticated"
        class="grid gap-6 xl:grid-cols-[340px_minmax(0,1fr)]"
      >
        <div class="panel-surface overflow-hidden">
          <div class="border-b border-border/80 px-5 py-4">
            <p class="text-sm font-semibold uppercase tracking-[0.18em] text-text-muted">
              {{ t('chat.conversations') }}
            </p>
          </div>
          <div
            v-if="loadingConversations"
            class="p-5 text-sm text-text-muted"
          >
            {{ t('chat.loadingConversations') }}
          </div>
          <div
            v-else
            class="flex max-h-[42rem] flex-col overflow-y-auto p-3"
          >
            <button
              v-for="conversation in conversations"
              :key="conversation.id"
              type="button"
              class="rounded-3xl border p-4 text-left transition"
              :class="
                selectedChatId === conversation.id
                  ? 'border-primary/30 bg-primary/10'
                  : 'border-transparent hover:border-border/80 hover:bg-surface-raised'
              "
              @click="handleSelectChat(conversation.id)"
            >
              <div class="flex items-start gap-3">
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
                      class="inline-flex min-w-6 items-center justify-center rounded-full bg-primary px-2 py-1 text-[11px] font-semibold text-white"
                    >
                      {{ conversation.unread_count }}
                    </span>
                  </div>
                  <p class="mt-3 line-clamp-2 text-sm leading-6 text-text-muted">
                    {{ conversation.last_message }}
                  </p>
                </div>
              </div>
            </button>
          </div>
        </div>

        <div class="panel-surface flex min-h-[42rem] flex-col overflow-hidden">
          <template v-if="activeConversation">
            <div class="border-b border-border/80 px-6 py-5">
              <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
                <div>
                  <p class="text-lg font-semibold text-text">
                    {{ activeConversation.peer.display_name }}
                  </p>
                  <p class="mt-1 text-sm text-text-muted">
                    {{ isSystemNoticeConversation ? t('chat.systemNoticeSubtitle') : activeConversation.listing.title }}
                  </p>
                </div>
                <div
                  v-if="!isSystemNoticeConversation"
                  class="rounded-2xl bg-surface-raised px-4 py-3 text-right"
                >
                  <p class="text-xs uppercase tracking-[0.16em] text-text-muted">
                    {{ t('chat.listingPrice') }}
                  </p>
                  <p class="mt-2 font-display text-2xl text-text">
                    {{ formatPrice(activeConversation.listing.price_hkd, preferenceStore.locale) }}
                  </p>
                </div>
              </div>
            </div>

            <div
              :ref="setMessageContainerRef"
              class="flex-1 space-y-4 overflow-y-auto px-5 py-5 sm:px-6"
            >
              <div
                v-if="loadingMessages"
                class="text-sm text-text-muted"
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
              class="border-t border-border/80 px-5 py-5 sm:px-6"
            >
              <div class="flex flex-col gap-3 sm:flex-row">
                <BaseTextarea
                  :model-value="draftMessage"
                  :rows="2"
                  :placeholder="t('chat.composePlaceholder')"
                  @update:model-value="draftMessage = $event"
                />
                <BaseButton
                  variant="primary"
                  class="self-end sm:self-auto"
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
</style>
