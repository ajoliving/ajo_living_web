<!--
 * 監控播放框。
 * 1. 先檢查播放回應，避免 JSON 錯誤觸發瀏覽器格式化工具列。
 * 2. 正常播放器維持原網址；錯誤以純文字顯示。
-->
<script setup lang="ts">
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps<{ src: string; title: string }>();
const { t } = useI18n();
const ready = ref(false);
const error = ref('');

// 1. 檢查目前播放網址，切換或收起鏡頭時取消舊請求。
watch(() => props.src, async (src, _, onCleanup) => {
  const controller = new AbortController();
  let current = true;
  const timeout = window.setTimeout(() => controller.abort(), 15000);
  onCleanup(() => {
    current = false;
    controller.abort();
    window.clearTimeout(timeout);
  });
  ready.value = false;
  error.value = '';
  try {
    const response = await fetch(src, { signal: controller.signal, credentials: 'omit' });
    const contentType = response.headers.get('content-type') ?? '';
    if (contentType.includes('json')) {
      const body: unknown = await response.json();
      if (controller.signal.aborted) return;
      error.value = body && typeof body === 'object' && 'detail' in body && typeof body.detail === 'string' && body.detail.trim()
        ? body.detail
        : t('building.icctv.playbackError');
    } else {
      await response.arrayBuffer();
      if (controller.signal.aborted) return;
      if (response.ok) ready.value = true;
      else error.value = t('building.icctv.playbackError');
    }
  } catch {
    if (current) error.value = t('building.icctv.playbackError');
  } finally {
    window.clearTimeout(timeout);
  }
}, { immediate: true });
</script>

<template>
  <iframe
    v-if="ready"
    :src="src"
    :title="title"
    allow="autoplay; fullscreen; encrypted-media; picture-in-picture"
  />
  <div v-else class="camera-frame-message" role="status">
    {{ error || t('building.common.loading') }}
  </div>
</template>

<style scoped>
.camera-frame-message {
  padding: 24px;
  color: rgb(var(--color-text));
  background: rgb(var(--color-surface));
  overflow-wrap: anywhere;
}
</style>
