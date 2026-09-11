<!--
 * 大廈群聊成員名單彈窗。
 * 1. 只讀顯示目前群組成員、身份與禁言狀態。
 * 2. 提供載入中與空狀態，關閉由父層控制。
-->
<script setup lang="ts">
import { computed } from 'vue';

import AppIcon from '@/shared/components/base/AppIcon.vue';
import BaseAvatar from '@/shared/components/base/BaseAvatar.vue';
import { useDialogBackdropClose } from '@/shared/composables/useDialogBackdropClose';
import type { BuildingChatMemberView } from '@/model/chat';

interface GroupMembersDialogProps {
  open: boolean;
  loading: boolean;
  members: BuildingChatMemberView[];
  title: string;
  memberCountLabel: string;
}

const props = defineProps<GroupMembersDialogProps>();

const emit = defineEmits<{
  (event: 'close'): void;
}>();

// 1. 只在完整點擊遮罩時關閉彈窗
const {
  handleBackdropPointerCancel,
  handleBackdropPointerDown,
  handleBackdropPointerUp,
} = useDialogBackdropClose(() => emit('close'));

// 2. 顯示禁言到期時間的成員
const mutedMembers = computed(() =>
  props.members.filter((member) => member.muted_until),
);
</script>

<template>
  <Teleport to="body">
    <Transition name="group-members-dialog">
      <div
        v-if="props.open"
        class="group-members-dialog"
        role="dialog"
        aria-modal="true"
        :aria-label="props.title"
        @pointercancel="handleBackdropPointerCancel"
        @pointerdown="handleBackdropPointerDown"
        @pointerup="handleBackdropPointerUp"
      >
        <section class="group-members-dialog__panel">
          <header class="group-members-dialog__header">
            <div class="group-members-dialog__heading">
              <h2>{{ props.title }}</h2>
              <p>{{ props.memberCountLabel }}</p>
            </div>
            <button
              type="button"
              class="group-members-dialog__close"
              :aria-label="$t('chat.closeMembers')"
              @click="emit('close')"
            >
              <AppIcon
                name="close"
                :size="18"
              />
            </button>
          </header>

          <div
            v-if="props.loading"
            class="group-members-dialog__state"
          >
            {{ $t('chat.loadingMembers') }}
          </div>

          <div
            v-else-if="props.members.length === 0"
            class="group-members-dialog__state"
          >
            {{ $t('chat.noMembers') }}
          </div>

          <ul
            v-else
            class="group-members-dialog__list"
          >
            <li
              v-for="member in props.members"
              :key="member.user_id"
              class="group-members-dialog__item"
            >
              <BaseAvatar
                src=""
                :name="member.display_name"
                :size="36"
              />
              <div class="group-members-dialog__member">
                <span class="group-members-dialog__name">{{ member.display_name }}</span>
                <span class="group-members-dialog__role">
                  {{ member.role_in_chat === 'admin' ? $t('chat.groupAdmin') : $t('chat.groupMember') }}
                </span>
              </div>
              <span
                v-if="member.membership_status === 'banned'"
                class="group-members-dialog__pill group-members-dialog__pill--banned"
              >
                {{ $t('chat.bannedMember') }}
              </span>
              <span
                v-else-if="member.muted_until"
                class="group-members-dialog__pill group-members-dialog__pill--muted"
              >
                {{ $t('chat.mutedMember') }}
              </span>
            </li>
          </ul>

          <p
            v-if="!props.loading && mutedMembers.length > 0"
            class="group-members-dialog__hint"
          >
            {{ $t('chat.membersReadOnlyHint') }}
          </p>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.group-members-dialog {
  position: fixed;
  z-index: 120;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgb(15 23 42 / 0.32);
  padding: 1rem;
}

.group-members-dialog__panel {
  display: flex;
  flex-direction: column;
  width: min(100%, 26rem);
  max-height: min(80svh, 34rem);
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  box-shadow: 0 24px 60px rgb(15 23 42 / 0.2);
  overflow: hidden;
}

.group-members-dialog__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  border-bottom: 1px solid rgb(var(--color-border));
  padding: 16px 18px;
}

.group-members-dialog__heading {
  min-width: 0;
}

.group-members-dialog__heading h2 {
  margin: 0;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: 1.05rem;
  font-weight: 600;
  line-height: 1.35;
}

.group-members-dialog__heading p {
  margin: 4px 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.8rem;
  line-height: 1.4;
}

.group-members-dialog__close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
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

.group-members-dialog__close:hover {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary) / 0.05);
}

.group-members-dialog__state {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 10rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.85rem;
}

.group-members-dialog__list {
  display: grid;
  gap: 2px;
  margin: 0;
  padding: 8px 10px;
  overflow-y: auto;
  list-style: none;
}

.group-members-dialog__item {
  display: flex;
  align-items: center;
  gap: 12px;
  min-height: 52px;
  padding: 8px 8px;
}

.group-members-dialog__item + .group-members-dialog__item {
  border-top: 1px solid rgb(var(--color-border) / 0.5);
}

.group-members-dialog__member {
  display: grid;
  gap: 2px;
  min-width: 0;
  flex: 1;
}

.group-members-dialog__name {
  overflow: hidden;
  color: rgb(var(--color-text));
  font-size: 0.875rem;
  font-weight: 700;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.group-members-dialog__role {
  color: rgb(var(--color-text-muted));
  font-size: 0.72rem;
  line-height: 1.3;
}

.group-members-dialog__pill {
  display: inline-flex;
  min-height: 22px;
  flex-shrink: 0;
  align-items: center;
  border-radius: 999px;
  padding: 2px 9px;
  font-size: 0.7rem;
  font-weight: 700;
}

.group-members-dialog__pill--muted {
  background: rgb(251 146 60 / 0.13);
  color: rgb(194 65 12);
}

.group-members-dialog__pill--banned {
  background: rgb(239 68 68 / 0.12);
  color: rgb(185 28 28);
}

.group-members-dialog__hint {
  margin: 0;
  border-top: 1px solid rgb(var(--color-border));
  padding: 10px 18px;
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  line-height: 1.5;
}

.group-members-dialog-enter-active,
.group-members-dialog-leave-active {
  transition: opacity 0.18s ease;
}

.group-members-dialog-enter-active .group-members-dialog__panel,
.group-members-dialog-leave-active .group-members-dialog__panel {
  transition: opacity 0.18s ease, transform 0.18s ease;
}

.group-members-dialog-enter-from,
.group-members-dialog-leave-to {
  opacity: 0;
}

.group-members-dialog-enter-from .group-members-dialog__panel,
.group-members-dialog-leave-to .group-members-dialog__panel {
  opacity: 0;
  transform: translateY(0.5rem) scale(0.985);
}

@media (max-width: 640px) {
  .group-members-dialog {
    align-items: flex-end;
    padding: 0;
  }

  .group-members-dialog__panel {
    width: 100%;
    max-height: 82svh;
    border-radius: 12px 12px 0 0;
    padding-bottom: var(--app-safe-bottom);
  }
}
</style>
