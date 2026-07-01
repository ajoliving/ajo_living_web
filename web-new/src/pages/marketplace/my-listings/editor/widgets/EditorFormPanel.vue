<!--
 * 發布頁表單區。
 * 1. 管理帖子資料、詳情、媒體與可見範圍欄位。
 * 2. 提供大面積圖片點選與拖拽上傳入口。
-->
<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import AppGlassSelect from '@/shared/components/base/AppGlassSelect.vue';
import AppIcon from '@/shared/components/base/AppIcon.vue';

import type {
  EditorImageSlot,
  EditorOption,
  ListingEditorBusinessStatus,
  ListingEditorFormState,
  ListingEditorVisibility,
} from '../editor';

interface EditorFormPanelProps {
  areaOptions: EditorOption[];
  businessStatusOptions: EditorOption<ListingEditorBusinessStatus>[];
  categoryOptions: EditorOption[];
  conditionOptions: EditorOption[];
  formState: ListingEditorFormState;
  imageSlots: EditorImageSlot[];
  priceModeOptions: EditorOption[];
  visibilityOptions: EditorOption<ListingEditorVisibility>[];
}

const props = defineProps<EditorFormPanelProps>();

const emit = defineEmits<{
  imageChange: [event: Event, slotId: string];
  imagesChange: [event: Event];
  imagesDrop: [files: File[]];
  moveImage: [slotId: string, offset: -1 | 1];
  removeImage: [slotId: string];
  selectCover: [slotId: string];
}>();

const { t } = useI18n();
const isDropActive = ref(false);
const isDeliveryMenuOpen = ref(false);
const deliveryMenuRef = ref<HTMLElement | null>(null);
const uploadedImageSlots = computed(() =>
  props.imageSlots.filter((slot) => Boolean(slot.url) || slot.uploading),
);
const deliveryTagOptions = computed(() => [
  { value: 'self_pickup', label: t('marketplace.editor.deliveryTagSelfPickup') },
  { value: 'door_delivery', label: t('marketplace.editor.deliveryTagDoorDelivery') },
  { value: 'free_post', label: t('marketplace.editor.deliveryTagFreePost') },
  { value: 'paid_post', label: t('marketplace.editor.deliveryTagPaidPost') },
  { value: 'face_check', label: t('marketplace.editor.deliveryTagFaceCheck') },
]);
const legacyDeliveryTagMap: Record<string, string> = {
  '自取': 'self_pickup',
  '送貨上門': 'door_delivery',
  '免費郵寄': 'free_post',
  '付費郵寄': 'paid_post',
  '面對面驗貨': 'face_check',
};
const normalizeDeliveryTagValue = (value: string): string => legacyDeliveryTagMap[value] ?? value;
const selectedDeliveryTagLabel = computed(() => {
  if (props.formState.deliveryTags.length === 0) {
    return t('marketplace.editor.deliveryTagsPlaceholder');
  }

  return props.formState.deliveryTags
    .map((value) => deliveryTagOptions.value.find((option) => option.value === normalizeDeliveryTagValue(value))?.label ?? value)
    .join(' / ');
});

// 1. 更新封面圖片
const selectCover = (slotId: string): void => {
  emit('selectCover', slotId);
};

// 2. 轉交批量圖片 input 事件
const handleImagesChange = (event: Event): void => {
  emit('imagesChange', event);
};

// 3. 處理拖拽圖片檔案
const handleImagesDrop = (event: DragEvent): void => {
  isDropActive.value = false;
  const files = Array.from(event.dataTransfer?.files ?? []);

  if (files.length > 0) {
    emit('imagesDrop', files);
  }
};

// 4. 更新可見範圍開關
const updateBuildingOnly = (event: Event): void => {
  const input = event.target as HTMLInputElement;
  props.formState.visibility = input.checked ? 'building_only' : 'public';
};

// 5. 切換交收標籤
const toggleDeliveryTag = (value: string): void => {
  const selectedValues = props.formState.deliveryTags.map(normalizeDeliveryTagValue);
  props.formState.deliveryTags = selectedValues.includes(value)
    ? props.formState.deliveryTags.filter((item) => normalizeDeliveryTagValue(item) !== value)
    : [...props.formState.deliveryTags, value];
};

// 6. 點擊外部時關閉下拉選單
const closeDeliveryMenuOnOutsideClick = (event: PointerEvent): void => {
  const target = event.target as HTMLElement;

  if (!deliveryMenuRef.value?.contains(target)) {
    isDeliveryMenuOpen.value = false;
  }
};

// 7. 掛載全域點擊監聽
onMounted(() => {
  document.addEventListener('pointerdown', closeDeliveryMenuOnOutsideClick);
});

// 8. 移除全域點擊監聽
onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', closeDeliveryMenuOnOutsideClick);
});
</script>

<template>
  <form
    class="editor-form-panel"
    @submit.prevent
  >
    <section class="editor-section">
      <header class="editor-section-heading">
        <span class="editor-step-number">1</span>
        <div>
          <p class="editor-kicker">
            {{ t('marketplace.editor.basicInfo') }}
          </p>
          <h2>{{ t('marketplace.editor.basicStepTitle') }}</h2>
        </div>
      </header>

      <div class="editor-field-grid">
        <label class="editor-field editor-field--wide">
          <span>{{ t('marketplace.editor.titleField') }}</span>
          <input
            v-model="props.formState.title"
            type="text"
            :placeholder="t('marketplace.editor.titlePlaceholder')"
          />
        </label>

        <div class="editor-field">
          <span>{{ t('marketplace.editor.categoryField') }}</span>
          <AppGlassSelect
            v-model="props.formState.categoryCode"
            :options="props.categoryOptions"
          />
        </div>

        <div class="editor-field">
          <span>{{ t('marketplace.editor.priceMode') }}</span>
          <AppGlassSelect
            v-model="props.formState.priceMode"
            :options="props.priceModeOptions"
          />
        </div>

        <label class="editor-field">
          <span>{{ t('marketplace.editor.priceField') }}</span>
          <input
            v-model.number="props.formState.price"
            type="number"
            min="0"
            placeholder="0.00"
          />
        </label>

        <div class="editor-field">
          <span>{{ t('common.label.condition') }}</span>
          <AppGlassSelect
            v-model="props.formState.condition"
            :options="props.conditionOptions"
          />
        </div>

        <div class="editor-field">
          <span>{{ t('marketplace.filter.area') }}</span>
          <AppGlassSelect
            v-model="props.formState.districtCode"
            :options="props.areaOptions"
          />
        </div>

        <label class="editor-donation-field">
          <input
            v-model="props.formState.isDonation"
            type="checkbox"
          />
          <span>{{ t('marketplace.editor.donationAvailable') }}</span>
        </label>

        <div class="editor-field">
          <span>{{ t('marketplace.editor.businessStatus') }}</span>
          <AppGlassSelect
            v-model="props.formState.businessStatus"
            :options="props.businessStatusOptions"
          />
        </div>
      </div>
    </section>

    <section class="editor-section">
      <header class="editor-section-heading">
        <span class="editor-step-number">2</span>
        <div>
          <p class="editor-kicker">
            {{ t('marketplace.editor.detailsStepKicker') }}
          </p>
          <h2>{{ t('marketplace.editor.detailsStepTitle') }}</h2>
        </div>
      </header>

      <div class="editor-field-grid">
        <label class="editor-field editor-field--wide">
          <span>{{ t('marketplace.editor.summaryField') }}</span>
          <input
            v-model="props.formState.summary"
            type="text"
            :placeholder="t('marketplace.editor.summaryPlaceholder')"
          />
        </label>

        <label class="editor-field editor-field--wide">
          <span>{{ t('marketplace.editor.descriptionField') }}</span>
          <textarea
            v-model="props.formState.description"
            rows="5"
            :placeholder="t('marketplace.editor.descriptionPlaceholder')"
          />
        </label>

        <div class="editor-field editor-field--wide">
          <span>{{ t('marketplace.editor.dimensionField') }}</span>
          <div class="editor-measurement-grid">
            <label class="editor-field">
              <span>{{ t('marketplace.editor.dimensionLength') }}</span>
              <input
                v-model="props.formState.dimensionLength"
                type="number"
                min="0"
                step="0.1"
                inputmode="decimal"
                :placeholder="t('marketplace.editor.dimensionLengthPlaceholder')"
              />
            </label>
            <label class="editor-field">
              <span>{{ t('marketplace.editor.dimensionWidth') }}</span>
              <input
                v-model="props.formState.dimensionWidth"
                type="number"
                min="0"
                step="0.1"
                inputmode="decimal"
                :placeholder="t('marketplace.editor.dimensionWidthPlaceholder')"
              />
            </label>
            <label class="editor-field">
              <span>{{ t('marketplace.editor.dimensionHeight') }}</span>
              <input
                v-model="props.formState.dimensionHeight"
                type="number"
                min="0"
                step="0.1"
                inputmode="decimal"
                :placeholder="t('marketplace.editor.dimensionHeightPlaceholder')"
              />
            </label>
            <label class="editor-field">
              <span>{{ t('marketplace.editor.dimensionWeight') }}</span>
              <input
                v-model="props.formState.dimensionWeight"
                type="number"
                min="0"
                step="0.1"
                inputmode="decimal"
                :placeholder="t('marketplace.editor.dimensionWeightPlaceholder')"
              />
            </label>
          </div>
        </div>
      </div>
    </section>

    <section class="editor-section">
      <header class="editor-section-heading">
        <span class="editor-step-number">3</span>
        <div>
          <p class="editor-kicker">
            {{ t('marketplace.editor.mediaStepKicker') }}
          </p>
          <h2>{{ t('marketplace.editor.mediaStepTitle') }}</h2>
        </div>
      </header>

      <input
        id="editor-media-upload"
        type="file"
        accept="image/*"
        multiple
        class="sr-only"
        @change="handleImagesChange"
      />
      <label
        for="editor-media-upload"
        class="editor-dropzone"
        :class="isDropActive ? 'editor-dropzone--active' : ''"
        @dragenter.prevent="isDropActive = true"
        @dragover.prevent="isDropActive = true"
        @dragleave.prevent="isDropActive = false"
        @drop.prevent="handleImagesDrop"
      >
        <AppIcon
          name="cloud-upload"
          class="editor-dropzone__icon"
          :size="42"
        />
        <strong>{{ t('marketplace.editor.dragUploadTitle') }}</strong>
        <span>{{ t('marketplace.editor.dragUploadHint') }}</span>
      </label>

      <div
        v-if="uploadedImageSlots.length > 0"
        class="editor-upload-grid"
      >
        <article
          v-for="slot in uploadedImageSlots"
          :key="slot.id"
          class="editor-upload-slot"
          :class="slot.isCover ? 'editor-upload-slot--cover' : ''"
        >
          <div class="editor-upload-slot__media">
            <img
              v-if="slot.url"
              :src="slot.url"
              :alt="slot.label"
            />
            <span
              v-else
              class="editor-upload-slot__empty"
            >
              {{ t('common.status.loading') }}
            </span>
          </div>

          <span
            v-if="slot.isCover && slot.url"
            class="editor-cover-badge"
          >
            {{ t('marketplace.editor.coverImage') }}
          </span>

          <button
            v-if="slot.url"
            type="button"
            class="editor-image-remove-button"
            :aria-label="t('marketplace.editor.removeImage')"
            @click="emit('removeImage', slot.id)"
          >
            ×
          </button>

          <div
            v-if="slot.url && !slot.isCover"
            class="editor-image-actions"
          >
            <button
              type="button"
              class="editor-image-action"
              @click="selectCover(slot.id)"
            >
              {{ t('marketplace.editor.setCover') }}
            </button>
          </div>
        </article>
      </div>
    </section>

    <section class="editor-section">
      <header class="editor-section-heading">
        <span class="editor-step-number">4</span>
        <div>
          <p class="editor-kicker">
            {{ t('marketplace.editor.visibilityTitle') }}
          </p>
          <h2>{{ t('marketplace.editor.visibilityStepTitle') }}</h2>
        </div>
      </header>

      <div class="editor-visibility-card">
        <div>
          <p>{{ t('marketplace.editor.communityExclusive') }}</p>
          <span>{{ t('marketplace.editor.communityExclusiveHint') }}</span>
        </div>
        <label class="editor-toggle">
          <input
            type="checkbox"
            class="peer sr-only"
            :checked="props.formState.visibility === 'building_only'"
            @change="updateBuildingOnly"
          />
          <span class="editor-toggle__rail" />
          <span class="editor-toggle__thumb" />
        </label>
      </div>

      <div class="editor-visibility-options">
        <button
          v-for="option in props.visibilityOptions"
          :key="option.value"
          type="button"
          class="editor-visibility-option"
          :class="props.formState.visibility === option.value ? 'editor-visibility-option--active' : ''"
          @click="props.formState.visibility = option.value"
        >
          {{ option.label }}
        </button>
      </div>

      <div class="editor-field-grid">
        <label class="editor-field">
          <span class="editor-field-label-row">
            <span>{{ t('marketplace.editor.phoneField') }}</span>
            <span class="editor-phone-prefix">{{ t('marketplace.editor.phoneCountryCode') }}</span>
          </span>
          <input
            v-model="props.formState.phone"
            type="tel"
            :placeholder="t('marketplace.editor.phonePlaceholder')"
          />
        </label>

        <label class="editor-field">
          <span class="editor-field-label-row">
            <span>{{ t('marketplace.editor.whatsAppField') }}</span>
            <span class="editor-phone-prefix">{{ t('marketplace.editor.phoneCountryCode') }}</span>
          </span>
          <input
            v-model="props.formState.whatsapp"
            type="tel"
            :placeholder="t('marketplace.editor.whatsAppPlaceholder')"
          />
          <small class="editor-field-hint">{{ t('marketplace.editor.contactRetainHint') }}</small>
        </label>

        <label class="editor-field">
          <span>{{ t('marketplace.editor.tradeNote') }}</span>
          <input
            v-model="props.formState.tradeNote"
            type="text"
            :placeholder="t('marketplace.editor.tradeNotePlaceholder')"
          />
        </label>

        <div class="editor-field">
          <span>{{ t('marketplace.editor.deliveryTags') }}</span>
          <div
            ref="deliveryMenuRef"
            class="editor-multi-select"
          >
            <button
              type="button"
              class="editor-multi-select__trigger"
              :aria-expanded="isDeliveryMenuOpen"
              @click="isDeliveryMenuOpen = !isDeliveryMenuOpen"
            >
              <span>{{ selectedDeliveryTagLabel }}</span>
              <AppIcon
                name="chevron-down"
                :size="16"
              />
            </button>
            <div
              v-if="isDeliveryMenuOpen"
              class="editor-multi-select__menu"
            >
              <label
                v-for="option in deliveryTagOptions"
                :key="option.value"
                class="editor-multi-select__option"
              >
                <input
                  type="checkbox"
                  :checked="props.formState.deliveryTags.map(normalizeDeliveryTagValue).includes(option.value)"
                  @change="toggleDeliveryTag(option.value)"
                />
                <span>{{ option.label }}</span>
              </label>
            </div>
          </div>
        </div>

        <div class="editor-field">
          <span>{{ t('marketplace.editor.contactPermission') }}</span>
          <label class="editor-chat-toggle">
            <span>{{ t('marketplace.editor.allowChat') }}</span>
            <span class="editor-toggle">
              <input
                v-model="props.formState.allowChat"
                type="checkbox"
                class="peer sr-only"
              />
              <span class="editor-toggle__rail" />
              <span class="editor-toggle__thumb" />
            </span>
          </label>
        </div>
      </div>
    </section>
  </form>
</template>

<style scoped>
.editor-form-panel {
  display: grid;
  gap: 1rem;
  min-width: 0;
}

.editor-section {
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  padding: 1rem;
  box-shadow: none;
}

.editor-section-heading {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  margin-bottom: 1rem;
}

.editor-step-number {
  display: inline-flex;
  width: 1.75rem;
  height: 1.75rem;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border-radius: 2px;
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
  font-size: 0.8125rem;
  font-weight: 800;
  line-height: 1;
}

.editor-kicker,
.editor-field span {
  color: rgb(var(--color-text-muted));
  font-size: 0.6875rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  line-height: 1;
  text-transform: uppercase;
}

.editor-section-heading h2 {
  margin: 0.25rem 0 0;
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 1.125rem;
  font-weight: 500;
  line-height: 1.3;
}

.editor-field-grid {
  display: grid;
  gap: 1rem;
}

.editor-field {
  display: grid;
  gap: 0.45rem;
  min-width: 0;
}

.editor-field-label-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.editor-field .editor-phone-prefix {
  display: inline-flex;
  align-items: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  padding: 0.25rem 0.45rem;
  color: rgb(var(--color-primary));
  font-size: 0.6875rem;
  font-weight: 800;
  letter-spacing: 0.04em;
  line-height: 1;
  text-transform: none;
}

.editor-field-hint {
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 600;
  line-height: 1.4;
}

.editor-donation-field {
  display: flex;
  min-height: 2.75rem;
  align-items: center;
  gap: 0.65rem;
  align-self: end;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  padding: 0.65rem 0.8rem;
  color: rgb(var(--color-text));
  font-size: 0.9375rem;
  font-weight: 700;
}

.editor-donation-field input {
  height: 1rem;
  width: 1rem;
  border-radius: 0.25rem;
  border-color: rgb(var(--color-border));
  color: rgb(var(--color-primary));
}

.editor-donation-field input:focus {
  --tw-ring-color: rgb(var(--color-primary));
  --tw-ring-offset-width: 0;
}

.editor-measurement-grid {
  display: grid;
  gap: 0.75rem;
}

.editor-field input,
.editor-field select,
.editor-field textarea {
  width: 100%;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  font-size: 0.9375rem;
  line-height: 1.6;
  outline: none;
  transition:
    background-color 0.2s ease,
    border-color 0.2s ease,
    box-shadow 0.2s ease;
}

.editor-field input,
.editor-field select {
  height: 2.75rem;
  padding: 0 0.85rem;
}

.editor-field textarea {
  min-height: 8.5rem;
  padding: 0.75rem 0.85rem;
  resize: vertical;
}

.editor-field input::placeholder,
.editor-field textarea::placeholder {
  color: rgb(var(--color-text-muted) / 0.72);
}

.editor-field input:disabled,
.editor-field select:disabled,
.editor-field textarea:disabled {
  background: color-mix(in srgb, rgb(var(--color-surface)) 82%, rgb(var(--color-page-tint)) 18%);
  color: rgb(var(--color-text-muted));
  cursor: not-allowed;
}

.editor-field input:focus,
.editor-field select:focus,
.editor-field textarea:focus {
  border-color: rgb(var(--color-primary));
  box-shadow: 0 0 0 3px rgb(var(--color-primary) / 0.16);
}

.editor-multi-select {
  position: relative;
  min-width: 0;
}

.editor-multi-select__trigger {
  display: flex;
  width: 100%;
  min-width: 0;
  min-height: 3rem;
  cursor: pointer;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  padding: 0 0.85rem;
  color: rgb(var(--color-text));
  font-size: 0.9375rem;
  line-height: 1.4;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease;
}

.editor-multi-select__trigger:hover,
.editor-multi-select__trigger[aria-expanded='true'] {
  border-color: rgb(var(--color-primary));
  box-shadow: 0 0 0 3px rgb(var(--color-primary) / 0.16);
}

.editor-field .editor-multi-select__trigger span {
  min-width: 0;
  max-width: 100%;
  overflow: hidden;
  color: inherit;
  font-size: inherit;
  font-weight: 600;
  letter-spacing: 0;
  line-height: inherit;
  text-overflow: ellipsis;
  text-transform: none;
  white-space: nowrap;
}

.editor-multi-select__trigger svg {
  flex-shrink: 0;
}

.editor-multi-select__menu {
  position: absolute;
  z-index: 40;
  top: calc(100% - 1px);
  left: 0;
  right: 0;
  display: grid;
  gap: 0.25rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0 0 2px 2px;
  background: rgb(var(--color-dropdown-surface));
  max-height: 14rem;
  overflow-y: auto;
  padding: 0.5rem;
  box-shadow: 0 8px 18px rgb(15 23 42 / 0.08);
  scrollbar-color: rgb(var(--color-border)) transparent;
  scrollbar-width: thin;
}

.editor-multi-select__menu::-webkit-scrollbar {
  width: 0.45rem;
}

.editor-multi-select__menu::-webkit-scrollbar-thumb {
  border-radius: 9999px;
  background: rgb(var(--color-border));
}

.editor-multi-select__option {
  display: flex;
  cursor: pointer;
  align-items: center;
  gap: 0.65rem;
  width: 100%;
  border: 0;
  border-radius: 2px;
  background: transparent;
  padding: 0.65rem 0.7rem;
  color: rgb(var(--color-text));
  font-size: 0.875rem;
  font-weight: 700;
  line-height: 1.3;
  text-align: left;
  transition: background-color 0.2s ease;
}

.editor-multi-select__option:hover {
  background: rgb(255 255 255 / 0.28);
}

.editor-multi-select__option input {
  width: 1rem;
  height: 1rem;
  accent-color: rgb(var(--color-primary));
}

.editor-field .editor-multi-select__option span {
  color: inherit;
  font-size: 0.875rem;
  font-weight: 700;
  letter-spacing: 0;
  line-height: 1.3;
  text-transform: none;
}

.editor-dropzone {
  display: flex;
  min-height: 9.5rem;
  cursor: pointer;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.65rem;
  border: 2px dashed rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  padding: 1.5rem;
  text-align: center;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    color 0.2s ease;
}

.editor-dropzone:hover,
.editor-dropzone--active {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-surface));
}

.editor-dropzone__icon {
  color: rgb(var(--color-text-muted));
}

.editor-dropzone:hover .editor-dropzone__icon,
.editor-dropzone--active .editor-dropzone__icon {
  color: rgb(var(--color-primary));
}

.editor-dropzone strong {
  color: rgb(var(--color-text));
  font-size: 1rem;
  font-weight: 600;
  line-height: 1.5;
}

.editor-dropzone span {
  color: rgb(var(--color-text-muted));
  font-size: 0.875rem;
  line-height: 1.6;
}

.editor-upload-grid {
  display: grid;
  gap: 0.75rem;
  margin-top: 1rem;
}

.editor-upload-slot {
  position: relative;
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
}

.editor-upload-slot--cover {
  border-color: rgb(var(--color-primary));
}

.editor-upload-slot__media,
.editor-upload-slot img {
  height: 8.5rem;
  width: 100%;
}

.editor-upload-slot__media {
  overflow: hidden;
}

.editor-upload-slot img {
  display: block;
  object-fit: cover;
}

.editor-upload-slot__empty {
  display: grid;
  height: 100%;
  place-items: center;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text-muted));
  font-size: 0.8125rem;
  font-weight: 700;
}

.editor-cover-badge {
  position: absolute;
  left: 0.75rem;
  top: 0.75rem;
  border-radius: 2px;
  background: rgb(var(--color-primary));
  padding: 0.4rem 0.6rem;
  color: rgb(var(--color-primary-contrast));
  font-size: 0.75rem;
  font-weight: 700;
  line-height: 1;
}

.editor-image-remove-button {
  position: absolute;
  right: 0.625rem;
  top: 0.625rem;
  display: inline-flex;
  width: 1.75rem;
  height: 1.75rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(255 255 255 / 0.92);
  color: rgb(var(--color-danger));
  cursor: pointer;
  font-size: 1.125rem;
  font-weight: 800;
  line-height: 1;
  box-shadow: 0 4px 10px rgb(0 0 0 / 0.1);
  transition:
    background-color 0.2s ease,
    border-color 0.2s ease,
    transform 0.2s ease;
}

.editor-image-remove-button:hover {
  border-color: rgb(var(--color-danger));
  background: rgb(var(--color-surface));
  transform: translateY(-1px);
}

.editor-image-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  padding: 0.75rem;
}

.editor-image-action {
  display: inline-flex;
  min-height: 1.95rem;
  cursor: pointer;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  padding: 0.4rem 0.7rem;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 700;
  line-height: 1;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    color 0.2s ease;
}

.editor-image-action:hover {
  border-color: rgb(var(--color-primary));
  color: rgb(var(--color-primary));
}

.editor-image-action--primary {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.editor-image-action--primary:hover {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary) / 0.9);
  color: rgb(var(--color-primary-contrast));
}

.editor-image-action--danger:hover {
  border-color: rgb(var(--color-danger));
  color: rgb(var(--color-danger));
}

.editor-visibility-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  padding: 0.85rem;
}

.editor-visibility-card p {
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 1rem;
  font-weight: 700;
  line-height: 1.5;
}

.editor-visibility-card span {
  display: block;
  margin-top: 0.25rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.875rem;
  line-height: 1.6;
}

.editor-visibility-options {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin: 0.85rem 0 1rem;
}

.editor-visibility-option {
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  padding: 0.55rem 0.95rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.875rem;
  font-weight: 700;
  line-height: 1;
}

.editor-visibility-option--active {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.editor-toggle {
  position: relative;
  display: inline-block;
  height: 1.5rem;
  width: 2.75rem;
  flex-shrink: 0;
}

.editor-toggle__rail {
  display: block;
  height: 1.5rem;
  width: 2.75rem;
  border-radius: 9999px;
  background: rgb(var(--color-border));
  transition: background-color 0.2s ease;
}

.editor-toggle__thumb {
  position: absolute;
  top: 0.1875rem;
  left: 0.1875rem;
  height: 1.125rem;
  width: 1.125rem;
  border-radius: 9999px;
  background: rgb(var(--color-surface));
  box-shadow: 0 2px 6px rgb(0 0 0 / 0.16);
  transition: transform 0.2s ease;
}

.peer:checked ~ .editor-toggle__rail {
  background: rgb(var(--color-primary));
}

.peer:checked ~ .editor-toggle__thumb {
  transform: translateX(1.25rem);
}

.editor-chat-toggle {
  display: flex;
  height: 2.75rem;
  min-height: 2.75rem;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  padding: 0 0.85rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.875rem;
  font-weight: 700;
  line-height: 1;
}

@media (min-width: 768px) {
  .editor-field-grid,
  .editor-upload-grid,
  .editor-measurement-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .editor-field--wide {
    grid-column: 1 / -1;
  }
}
</style>
