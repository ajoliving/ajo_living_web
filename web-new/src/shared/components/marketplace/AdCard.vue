<!--
 * 廣告卡片元件。
 * 1. 支援 banner（16:9）與 vertical（9:16）兩種版型。
 * 2. 提供 visual 屬性切換漸層背景（office/home/service/market/move）。
 * 3. 含標籤、標題、描述三層文案。
-->
<script setup lang="ts">
// 1. 定義 Props 介面
interface AdCardProps {
  variant: 'banner' | 'vertical';
  visual?: 'default' | 'office' | 'home' | 'service' | 'market' | 'move';
  label?: string;
  title: string;
  desc?: string;
}

// 2. 設定 visual 預設值為 'default'
withDefaults(defineProps<AdCardProps>(), {
  visual: 'default',
});
</script>

<template>
  <article class="ad-card" :class="variant">
    <div
      class="ad-visual"
      :class="visual !== 'default' ? `ad-visual--${visual}` : ''"
    ></div>
    <div class="ad-copy">
      <span v-if="label" class="ad-label">{{ label }}</span>
      <div class="ad-title">{{ title }}</div>
      <div v-if="desc" class="ad-desc">{{ desc }}</div>
    </div>
  </article>
</template>

<style scoped>
.ad-card {
  position: relative;
  display: flex;
  overflow: hidden;
  min-height: 0;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background: rgb(var(--color-surface));
  color: var(--color-text);
  cursor: pointer;
  box-shadow: var(--shadow-soft);
}

.ad-card:hover {
  border-color: var(--color-brand-mid);
  box-shadow: var(--shadow-raised);
}

.ad-card::after {
  content: '';
  position: absolute;
  inset: 0;
  background: repeating-linear-gradient(
    45deg,
    transparent,
    transparent 6px,
    rgba(0, 0, 0, 0.025) 6px,
    rgba(0, 0, 0, 0.025) 12px
  );
  pointer-events: none;
}

.ad-card.banner {
  aspect-ratio: 16 / 9;
}

.ad-card.vertical {
  aspect-ratio: 9 / 16;
}

.ad-visual {
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, #f7f3ee, #e7ded6);
}

.ad-visual--office {
  background: linear-gradient(135deg, #eef1f4, #d9e0e6);
}

.ad-visual--home {
  background: linear-gradient(135deg, #f3eee8, #dfd4c8);
}

.ad-visual--service {
  background: linear-gradient(135deg, #eef4ef, #d9e6dc);
}

.ad-visual--market {
  background: linear-gradient(135deg, #f3f0ea, #e2d8ca);
}

.ad-visual--move {
  background: linear-gradient(135deg, #edf2f3, #d7e1e2);
}

.ad-copy {
  position: relative;
  z-index: 1;
  display: flex;
  width: 100%;
  flex-direction: column;
  justify-content: flex-end;
  padding: 13px;
  background: linear-gradient(
    180deg,
    rgba(255, 255, 255, 0) 12%,
    rgba(255, 255, 255, 0.9) 76%,
    #fff 100%
  );
}

.ad-title {
  color: var(--color-text);
  font-size: 13px;
  font-weight: 700;
  line-height: 1.35;
}

.ad-desc {
  margin-top: 3px;
  color: var(--color-ink-3);
  font-size: 11px;
  line-height: 1.45;
}

.ad-card.vertical .ad-copy {
  padding: 12px;
  background: linear-gradient(
    180deg,
    rgba(255, 255, 255, 0) 6%,
    rgba(255, 255, 255, 0.94) 72%,
    #fff 100%
  );
}

.ad-card.vertical .ad-title {
  font-size: 12px;
}

.ad-card.vertical .ad-desc {
  font-size: 10px;
}

.ad-label {
  align-self: flex-start;
  margin-bottom: 6px;
  border-radius: 3px;
  background: var(--color-primary);
  color: #fff;
  font-size: 10px;
  font-weight: 700;
  padding: 3px 6px;
}
</style>
