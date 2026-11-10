<!--
 * 共用空狀態元件。
 * 1. 統一空清單、未找到與待選狀態的視覺呈現。
 * 2. 支援額外操作按鈕插槽。
-->
<script setup lang="ts">
import AppIcon from '@/shared/components/base/AppIcon.vue';

interface EmptyStateProps {
  title: string;
  description: string;
}

const props = defineProps<EmptyStateProps>();
</script>

<template>
  <div class="panel-surface flex flex-col items-center justify-center gap-4 px-6 py-12 text-center">
    <div class="flex h-16 w-16 items-center justify-center rounded-full bg-primary/10 text-2xl text-primary">
      <AppIcon
        name="inbox"
        :size="28"
      />
    </div>
    <div class="max-w-md space-y-2">
      <h3 class="text-xl font-semibold text-text">
        {{ props.title }}
      </h3>
      <p class="text-sm leading-6 text-text-muted">
        {{ props.description }}
      </p>
    </div>
    <slot />
  </div>
</template>
