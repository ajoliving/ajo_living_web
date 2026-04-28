<!--
 * 自定義頭像元件。
 * 1. 提供圖片頭像與文字回退樣式。
 * 2. 統一列表、導航與聊天頭像顯示。
-->
<script setup lang="ts">
import { computed } from 'vue';

interface BaseAvatarProps {
  src: string;
  alt?: string;
  name?: string;
  size?: number;
}

const props = withDefaults(defineProps<BaseAvatarProps>(), {
  alt: '',
  name: '',
  size: 40,
});

// 1. 產生頭像回退文字
const fallbackLabel = computed(() =>
  props.name
    .split(' ')
    .filter(Boolean)
    .slice(0, 2)
    .map((item) => item[0]?.toUpperCase() ?? '')
    .join(''),
);
</script>

<template>
  <div
    class="app-avatar shrink-0"
    :style="{ width: `${props.size}px`, height: `${props.size}px` }"
  >
    <img
      v-if="props.src"
      :src="props.src"
      :alt="props.alt || props.name"
      class="h-full w-full object-cover"
    />
    <div
      v-else
      class="app-avatar-fallback"
    >
      {{ fallbackLabel }}
    </div>
  </div>
</template>
