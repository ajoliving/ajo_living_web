<!--
 * 會員中心區塊標題。
 * 1. 統一會員中心子頁的標題、眉標與說明文。
 * 2. 支援右側操作插槽。
-->
<script setup lang="ts">
interface AccountSectionHeaderProps {
  eyebrow?: string;
  title: string;
  description: string;
}

const props = withDefaults(defineProps<AccountSectionHeaderProps>(), {
  eyebrow: '',
});
</script>

<template>
  <div class="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
    <div class="space-y-3">
      <p
        v-if="props.eyebrow"
        class="text-kicker"
      >
        {{ props.eyebrow }}
      </p>
      <div class="space-y-2">
        <h1 class="display-title text-3xl text-text sm:text-4xl">
          {{ props.title }}
        </h1>
        <p class="max-w-3xl text-sm leading-7 text-text-muted sm:text-base">
          {{ props.description }}
        </p>
      </div>
    </div>

    <slot name="actions" />
  </div>
</template>
