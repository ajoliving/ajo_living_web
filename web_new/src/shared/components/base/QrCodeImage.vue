<!--
 * 二維碼圖片元件。
 * 1. 將付款鏈接渲染成本地二維碼圖片。
 * 2. 在生成失敗時顯示穩定 fallback。
-->
<script setup lang="ts">
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import * as QRCode from 'qrcode';

const props = withDefaults(defineProps<{
  text: string;
  alt?: string;
  size?: number;
}>(), {
  alt: 'QR code',
  size: 260,
});

const { t } = useI18n();
const dataUrl = ref('');
const hasError = ref(false);

// 1. 生成二維碼圖片
const renderQrCode = async (value: string): Promise<void> => {
  const text = String(value || '').trim();
  if (!text) {
    dataUrl.value = '';
    hasError.value = false;
    return;
  }

  try {
    dataUrl.value = await QRCode.toDataURL(text, {
      width: props.size,
      margin: 1,
      errorCorrectionLevel: 'M',
    });
    hasError.value = false;
  } catch {
    dataUrl.value = '';
    hasError.value = true;
  }
};

watch(
  () => [props.text, props.size] as const,
  ([text]) => {
    void renderQrCode(text);
  },
  { immediate: true },
);
</script>

<template>
  <img
    v-if="dataUrl"
    :src="dataUrl"
    :alt="props.alt"
  />
  <div
    v-else-if="hasError"
    class="qr-code-fallback"
  >
    {{ t('account.wallet.qrRenderError') }}
  </div>
</template>

<style scoped>
.qr-code-fallback {
  display: grid;
  min-height: 13rem;
  place-items: center;
  border-radius: 8px;
  background: rgb(var(--color-surface-raised));
  color: rgb(var(--color-text-muted));
  font-size: 0.875rem;
  font-weight: 800;
}
</style>
