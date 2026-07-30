<!--
 * 樓盤搜尋與發布預覽共用卡片。
 * 1. 統一公開搜尋結果與發布中的樓盤資訊層級。
 * 2. 預覽模式隱藏封面、發布身份與訪客操作。
 -->
<script setup lang="ts">
import { useI18n } from 'vue-i18n';

import type { PropertyListingCardViewModel } from '@/model/property';
import AppIcon from '@/shared/components/base/AppIcon.vue';

interface PropertyListingCardProps {
  card: PropertyListingCardViewModel;
  preview?: boolean;
}

const props = withDefaults(defineProps<PropertyListingCardProps>(), {
  preview: false,
});
const emit = defineEmits<{
  open: [];
  favorite: [];
}>();
const { t } = useI18n();

// 1. 開啟公開樓盤詳情
const handleOpen = (): void => {
  if (!props.preview) {
    emit('open');
  }
};

// 2. 更新收藏狀態
const handleFavorite = (): void => {
  emit('favorite');
};
</script>

<template>
  <article
    class="property-listing-card"
    :class="{ 'property-listing-card--preview': preview }"
    :tabindex="preview ? undefined : 0"
    :role="preview ? undefined : 'link'"
    @click="handleOpen"
    @keydown.enter="handleOpen"
  >
    <div
      v-if="!preview"
      class="property-listing-card__type-stack"
    >
      <span>{{ card.propertyType }}</span>
      <span>{{ card.publisherLabel }}</span>
    </div>
    <div
      v-if="!preview"
      class="property-listing-card__media"
    >
      <img
        v-if="card.imageUrl"
        :src="card.imageUrl"
        :alt="card.title"
      >
      <AppIcon
        v-else
        name="picture"
        :size="44"
      />
    </div>
    <div class="property-listing-card__body">
      <div class="property-listing-card__tags">
        <span
          v-for="tag in card.tags"
          :key="`${tag.label}-${tag.dark ? 'dark' : 'light'}`"
          :class="{ 'property-listing-card__tag--dark': tag.dark }"
          class="property-listing-card__tag"
        >{{ tag.label }}</span>
      </div>
      <h3>{{ card.title }}</h3>
      <p class="property-listing-card__location">{{ card.location }}</p>
      <p
        v-if="card.facts.length > 0"
        class="property-listing-card__facts"
      >{{ card.facts.join(' · ') }}</p>
      <div class="property-listing-card__price">
        <span :class="card.priceKind">{{ card.priceKind === 'sale' ? t('property.publicList.sale') : t('property.publicList.rent') }}</span>
        {{ card.price }}<small v-if="card.priceUnit">{{ card.priceUnit }}</small>
      </div>
      <p
        v-if="card.area"
        class="property-listing-card__area"
      >
        {{ card.area }}
        <span v-if="card.areaPrice">{{ card.areaPrice }}</span>
      </p>
      <div
        v-if="card.pills.length > 0"
        class="property-listing-card__pills"
      >
        <span
          v-for="pill in card.pills"
          :key="pill"
        >{{ pill }}</span>
      </div>
    </div>
    <div
      v-if="!preview"
      class="property-listing-card__actions"
    >
      <button
        type="button"
        @click.stop="handleFavorite"
      >{{ card.favorite ? t('property.publicList.favorited') : t('property.publicList.favorite') }}</button>
      <button
        type="button"
        class="property-listing-card__view"
        @click.stop="handleOpen"
      >{{ t('property.publicList.view') }}</button>
    </div>
  </article>
</template>

<style scoped>
.property-listing-card {
  position: relative;
  display: grid;
  grid-template-columns: minmax(300px, 42%) minmax(0, 1fr);
  min-height: 246px;
  overflow: hidden;
  border: 1px solid var(--bdr);
  border-radius: var(--r-md);
  background: var(--sur);
  cursor: pointer;
  transition: box-shadow 0.15s, border-color 0.15s;
}

.property-listing-card:hover,
.property-listing-card:focus-visible {
  border-color: var(--brand-mid);
  box-shadow: var(--shadow-md);
  outline: none;
}

.property-listing-card--preview {
  grid-template-columns: minmax(0, 1fr);
  min-height: 0;
  cursor: default;
}

.property-listing-card--preview:hover {
  border-color: var(--bdr);
  box-shadow: none;
}

.property-listing-card__media {
  display: flex;
  grid-column: 1;
  grid-row: 1 / span 2;
  align-items: center;
  justify-content: center;
  min-height: 0;
  background: var(--sur-2);
  color: var(--ink-3);
}

.property-listing-card__media img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.property-listing-card__body {
  display: flex;
  grid-column: 2;
  flex-direction: column;
  min-width: 0;
  padding: 28px 18px 10px;
}

.property-listing-card--preview .property-listing-card__body {
  grid-column: 1;
  padding: 18px;
}

.property-listing-card__tags,
.property-listing-card__pills {
  display: flex;
  flex-wrap: wrap;
  gap: 5px;
}

.property-listing-card__tags {
  margin-bottom: 10px;
}

.property-listing-card__tag {
  border: 1px solid var(--bdr);
  border-radius: 1px;
  padding: 3px 7px;
  color: var(--ink-3);
  font-size: var(--text-sm);
  letter-spacing: 0;
}

.property-listing-card__tag--dark {
  border-color: var(--brand);
  background: var(--brand);
  color: #fff;
  font-weight: 500;
}

.property-listing-card h3 {
  margin: 0 0 3px;
  color: var(--ink);
  font-size: var(--text-md);
  font-weight: 600;
  line-height: 1.35;
}

.property-listing-card__location {
  margin: 2px 0 5px;
  color: var(--ink-2);
  font-size: var(--text-sm);
  font-weight: 700;
  line-height: 1.4;
}

.property-listing-card__facts {
  margin: 0 0 7px;
  color: var(--ink-3);
  font-size: var(--text-sm);
  line-height: 1.55;
}

.property-listing-card__price {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 7px;
  margin-top: 6px;
  color: var(--ink);
  font-size: 20px;
  font-weight: 300;
  line-height: 1.25;
}

.property-listing-card__price > span {
  display: inline-flex;
  width: 24px;
  height: 24px;
  align-items: center;
  justify-content: center;
  border-radius: 7px;
  background: var(--brand);
  color: #fff;
  font-size: var(--text-base);
  font-weight: 800;
  line-height: 1;
}

.property-listing-card__price > span.rent {
  background: #29b6e8;
}

.property-listing-card__price small {
  color: var(--ink-3);
  font-size: var(--text-sm);
  font-weight: 400;
}

.property-listing-card__area {
  margin: 6px 0 0;
  color: var(--ink-3);
  font-size: 10px;
  line-height: 1.4;
}

.property-listing-card__area span {
  margin-left: 8px;
  color: var(--accent);
  font-size: 11px;
  font-weight: 600;
  white-space: nowrap;
}

.property-listing-card__pills {
  margin-top: 9px;
}

.property-listing-card__pills span {
  border-radius: 2px;
  background: var(--sur-2);
  padding: 2px 6px;
  color: var(--ink-2);
  font-size: 9px;
}

.property-listing-card__type-stack {
  position: absolute;
  z-index: 2;
  top: 10px;
  left: 10px;
  display: grid;
  gap: 4px;
}

.property-listing-card__type-stack span {
  display: inline-flex;
  width: max-content;
  max-width: 90px;
  border-radius: 4px;
  background: rgb(26 26 26 / 0.82);
  padding: 5px 7px;
  color: #fff;
  font-size: 10px;
  font-weight: 700;
  line-height: 1;
}

.property-listing-card__type-stack span + span {
  background: var(--brand);
}

.property-listing-card__actions {
  display: grid;
  grid-column: 2;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
  align-self: end;
  max-height: 0;
  overflow: hidden;
  padding: 0 18px;
  opacity: 0;
  pointer-events: none;
  transform: translateY(-2px);
  transition: max-height 0.18s ease, opacity 0.15s ease, transform 0.15s ease, padding 0.18s ease;
}

.property-listing-card:hover .property-listing-card__actions,
.property-listing-card:focus-visible .property-listing-card__actions {
  max-height: 54px;
  padding: 0 18px 14px;
  opacity: 1;
  pointer-events: auto;
  transform: translateY(0);
}

.property-listing-card__actions button {
  min-height: 34px;
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: #fff;
  color: var(--ink-2);
  cursor: pointer;
  font: inherit;
  font-size: var(--text-sm);
  font-weight: 500;
}

.property-listing-card__actions .property-listing-card__view {
  border-color: var(--brand);
  background: var(--brand);
  color: #fff;
}

@media (max-width: 767px) {
  .property-listing-card {
    grid-template-columns: 1fr;
    min-height: 0;
  }

  .property-listing-card__media,
  .property-listing-card__body,
  .property-listing-card__actions {
    grid-column: 1;
    grid-row: auto;
  }

  .property-listing-card__media {
    min-height: 180px;
  }

  .property-listing-card__body {
    padding: 14px;
  }

  .property-listing-card__actions {
    max-height: none;
    padding: 0 14px 14px;
    opacity: 1;
    pointer-events: auto;
    transform: none;
  }
}
</style>
