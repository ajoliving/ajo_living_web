<!--
 * 自定義頭像元件。
 * 1. 提供圖片頭像與文字回退樣式。
 * 2. 統一列表、導航與聊天頭像顯示。
-->
<script setup lang="ts">
import { computed, ref, watch } from 'vue';

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

const imageLoadFailed = ref(false);

// 1. 產生頭像回退文字
const fallbackLabel = computed(() =>
  props.name
    .split(' ')
    .filter(Boolean)
    .slice(0, 2)
    .map((item) => item[0]?.toUpperCase() ?? '')
    .join(''),
);

// 2. 頭像地址變更時重置圖片錯誤狀態
watch(
  () => props.src,
  () => {
    imageLoadFailed.value = false;
  },
);
</script>

<template>
  <div
    class="app-avatar shrink-0"
    :style="{ width: `${props.size}px`, height: `${props.size}px` }"
  >
    <img
      v-if="props.src && !imageLoadFailed"
      :src="props.src"
      :alt="props.alt || props.name"
      class="h-full w-full object-cover"
      @error="imageLoadFailed = true"
    />
    <div
      v-else
      class="app-avatar-fallback"
    >
      {{ fallbackLabel }}
    </div>
  </div>
</template>
