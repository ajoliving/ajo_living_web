<!--
 * 全域提示視窗。
 * 1. 顯示來自回饋狀態的成功或錯誤提示。
 * 2. 以輕量浮層形式提供操作結果反饋。
-->
<script setup lang="ts">
import { storeToRefs } from 'pinia';

import { useFeedbackStore } from '@/stores/feedback';

import AppIcon from '@/shared/components/base/AppIcon.vue';
import BaseButton from '@/shared/components/base/BaseButton.vue';

const feedbackStore = useFeedbackStore();
const { toasts } = storeToRefs(feedbackStore);

// 1. 取得提示顏色樣式
const getToneClassName = (tone: 'success' | 'error' | 'info') => {
  if (tone === 'success') {
    return 'border-success/20 bg-success/10 text-success';
  }

  if (tone === 'error') {
    return 'border-danger/20 bg-danger/10 text-danger';
  }

  return 'border-primary/20 bg-primary/10 text-primary';
};
</script>

<template>
  <Teleport to="body">
    <div class="pointer-events-none fixed inset-x-0 top-5 z-50 flex flex-col items-center gap-3 px-4">
      <transition-group
        name="toast-fade"
        tag="div"
        class="flex w-full max-w-md flex-col gap-3"
      >
        <div
          v-for="toast in toasts"
          :key="toast.id"
          class="app-toast pointer-events-auto"
        >
          <div
            class="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-full border"
            :class="getToneClassName(toast.tone)"
          >
            <AppIcon
              :name="toast.tone === 'success' ? 'check-circle' : 'shield'"
              :size="16"
            />
          </div>
          <p class="flex-1 text-sm leading-6 text-text">
            {{ toast.message }}
          </p>
          <BaseButton
            variant="ghost"
            size="sm"
            @click="feedbackStore.dismissToast(toast.id)"
          >
            <template #leading>
              <AppIcon
                name="close"
                :size="16"
              />
            </template>
          </BaseButton>
        </div>
      </transition-group>
    </div>
  </Teleport>
</template>

<style scoped>
.toast-fade-enter-active,
.toast-fade-leave-active {
  transition:
    opacity 0.22s ease,
    transform 0.22s ease;
}

.toast-fade-enter-from,
.toast-fade-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
