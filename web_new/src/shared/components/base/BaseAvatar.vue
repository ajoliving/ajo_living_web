<!--
 * 自定義頭像元件。
 * 1. 優先顯示預設頭像 key，其次顯示上傳圖片，最後以識別碼穩定映射。
 * 2. 統一列表、導航與聊天頭像顯示。
-->
<script setup lang="ts">
import { computed, ref, watch } from 'vue';

import { resolveAvatarSrc } from '@/constants/presetAvatars';

interface BaseAvatarProps {
  src?: string;
  avatarKey?: string;
  seed?: string;
  alt?: string;
  name?: string;
  size?: number;
}

const props = withDefaults(defineProps<BaseAvatarProps>(), {
  src: '',
  avatarKey: '',
  seed: '',
  alt: '',
  name: '',
  size: 40,
});

const imageLoadFailed = ref(false);

// 1. 解析頭像來源
const imageSrc = computed(() => {
  if (!props.avatarKey && !props.src && !props.seed) {
    return '';
  }

  return resolveAvatarSrc({
    avatarKey: props.avatarKey,
    avatarUrl: props.src,
    seed: props.seed,
  });
});

// 2. 產生頭像回退文字
const fallbackLabel = computed(() =>
  props.name
    .split(' ')
    .filter(Boolean)
    .slice(0, 2)
    .map((item) => item[0]?.toUpperCase() ?? '')
    .join(''),
);

// 3. 頭像地址變更時重置圖片錯誤狀態
watch(
  imageSrc,
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
      v-if="imageSrc && !imageLoadFailed"
      :src="imageSrc"
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
