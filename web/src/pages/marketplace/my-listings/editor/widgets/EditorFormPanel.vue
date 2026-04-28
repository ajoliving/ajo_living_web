<!--
 * 發布頁表單區。
 * 1. 管理帖子資料、詳情、媒體與可見範圍欄位。
 * 2. 提供大面積圖片點選與拖拽上傳入口。
-->
<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

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

type SingleSelectKey = 'categoryCode' | 'priceMode' | 'condition' | 'districtCode' | 'businessStatus';

const { t } = useI18n();
const isDropActive = ref(false);
const isDeliveryMenuOpen = ref(false);
const openSingleSelect = ref<SingleSelectKey | ''>('');
const deliveryMenuRef = ref<HTMLElement | null>(null);
const uploadedImageSlots = computed(() =>
  props.imageSlots.filter((slot) => Boolean(slot.url) || slot.uploading),
);
const deliveryTagOptions = computed(() => [
  { value: '自取', label: t('marketplace.editor.deliveryTagSelfPickup') },
  { value: '送貨上門', label: t('marketplace.editor.deliveryTagDoorDelivery') },
  { value: '免費郵寄', label: t('marketplace.editor.deliveryTagFreePost') },
  { value: '付費郵寄', label: t('marketplace.editor.deliveryTagPaidPost') },
  { value: '面對面驗貨', label: t('marketplace.editor.deliveryTagFaceCheck') },
]);
const selectedDeliveryTagLabel = computed(() => {
  if (props.formState.deliveryTags.length === 0) {
    return t('marketplace.editor.deliveryTagsPlaceholder');
  }

  return props.formState.deliveryTags
    .map((value) => deliveryTagOptions.value.find((option) => option.value === value)?.label ?? value)
    .join(' / ');
});

// 1. 讀取單選下拉顯示文字
const resolveSingleSelectLabel = (options: EditorOption[], value: string): string =>
  options.find((option) => option.value === value)?.label ?? value;

// 2. 更新封面圖片
const selectCover = (slotId: string): void => {
  emit('selectCover', slotId);
};

// 3. 轉交批量圖片 input 事件
const handleImagesChange = (event: Event): void => {
  emit('imagesChange', event);
};

// 4. 處理拖拽圖片檔案
const handleImagesDrop = (event: DragEvent): void => {
  isDropActive.value = false;
  const files = Array.from(event.dataTransfer?.files ?? []);

  if (files.length > 0) {
    emit('imagesDrop', files);
  }
};

// 5. 更新可見範圍開關
const updateBuildingOnly = (event: Event): void => {
  const input = event.target as HTMLInputElement;
  props.formState.visibility = input.checked ? 'building_only' : 'public';
};

// 6. 切換交收標籤
const toggleDeliveryTag = (value: string): void => {
  props.formState.deliveryTags = props.formState.deliveryTags.includes(value)
    ? props.formState.deliveryTags.filter((item) => item !== value)
    : [...props.formState.deliveryTags, value];
};

// 7. 切換單選下拉
const toggleSingleSelect = (key: SingleSelectKey): void => {
  openSingleSelect.value = openSingleSelect.value === key ? '' : key;
  isDeliveryMenuOpen.value = false;
};

// 8. 更新單選下拉值
const updateSingleSelect = (key: SingleSelectKey, value: string): void => {
  switch (key) {
    case 'categoryCode':
      props.formState.categoryCode = value as ListingEditorFormState['categoryCode'];
      break;
    case 'priceMode':
      props.formState.priceMode = value as ListingEditorFormState['priceMode'];
      break;
    case 'condition':
      props.formState.condition = value as ListingEditorFormState['condition'];
      break;
    case 'districtCode':
      props.formState.districtCode = value as ListingEditorFormState['districtCode'];
      break;
    case 'businessStatus':
      props.formState.businessStatus = value as ListingEditorFormState['businessStatus'];
      break;
    default:
      break;
  }

  openSingleSelect.value = '';
};

// 9. 點擊外部時關閉下拉選單
const closeDeliveryMenuOnOutsideClick = (event: PointerEvent): void => {
  const target = event.target as HTMLElement;

  if (!deliveryMenuRef.value?.contains(target)) {
    isDeliveryMenuOpen.value = false;
  }
  if (!target.closest('.editor-select')) {
    openSingleSelect.value = '';
  }
};

// 10. 掛載全域點擊監聽
onMounted(() => {
  document.addEventListener('pointerdown', closeDeliveryMenuOnOutsideClick);
});

// 11. 移除全域點擊監聽
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
          <div class="editor-select">
            <button
              type="button"
              class="editor-select__trigger"
              :aria-expanded="openSingleSelect === 'categoryCode'"
              @click="toggleSingleSelect('categoryCode')"
            >
              <span>{{ resolveSingleSelectLabel(props.categoryOptions, props.formState.categoryCode) }}</span>
              <AppIcon
                name="chevron-down"
                :size="16"
              />
            </button>
            <div
              v-if="openSingleSelect === 'categoryCode'"
              class="editor-select__menu"
            >
              <button
                v-for="option in props.categoryOptions"
                :key="option.value"
                type="button"
                class="editor-select__option"
                :class="props.formState.categoryCode === option.value ? 'editor-select__option--active' : ''"
                @click="updateSingleSelect('categoryCode', option.value)"
              >
                {{ option.label }}
              </button>
            </div>
          </div>
        </div>

        <div class="editor-field">
          <span>{{ t('marketplace.editor.priceMode') }}</span>
          <div class="editor-select">
            <button
              type="button"
              class="editor-select__trigger"
              :aria-expanded="openSingleSelect === 'priceMode'"
              @click="toggleSingleSelect('priceMode')"
            >
              <span>{{ resolveSingleSelectLabel(props.priceModeOptions, props.formState.priceMode) }}</span>
              <AppIcon
                name="chevron-down"
                :size="16"
              />
            </button>
            <div
              v-if="openSingleSelect === 'priceMode'"
              class="editor-select__menu"
            >
              <button
                v-for="option in props.priceModeOptions"
                :key="option.value"
                type="button"
                class="editor-select__option"
                :class="props.formState.priceMode === option.value ? 'editor-select__option--active' : ''"
                @click="updateSingleSelect('priceMode', option.value)"
              >
                {{ option.label }}
              </button>
            </div>
          </div>
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
          <div class="editor-select">
            <button
              type="button"
              class="editor-select__trigger"
              :aria-expanded="openSingleSelect === 'condition'"
              @click="toggleSingleSelect('condition')"
            >
              <span>{{ resolveSingleSelectLabel(props.conditionOptions, props.formState.condition) }}</span>
              <AppIcon
                name="chevron-down"
                :size="16"
              />
            </button>
            <div
              v-if="openSingleSelect === 'condition'"
              class="editor-select__menu"
            >
              <button
                v-for="option in props.conditionOptions"
                :key="option.value"
                type="button"
                class="editor-select__option"
                :class="props.formState.condition === option.value ? 'editor-select__option--active' : ''"
                @click="updateSingleSelect('condition', option.value)"
              >
                {{ option.label }}
              </button>
            </div>
          </div>
        </div>

        <div class="editor-field">
          <span>{{ t('marketplace.filter.area') }}</span>
          <div class="editor-select">
            <button
              type="button"
              class="editor-select__trigger"
              :aria-expanded="openSingleSelect === 'districtCode'"
              @click="toggleSingleSelect('districtCode')"
            >
              <span>{{ resolveSingleSelectLabel(props.areaOptions, props.formState.districtCode) }}</span>
              <AppIcon
                name="chevron-down"
                :size="16"
              />
            </button>
            <div
              v-if="openSingleSelect === 'districtCode'"
              class="editor-select__menu"
            >
              <button
                v-for="option in props.areaOptions"
                :key="option.value"
                type="button"
                class="editor-select__option"
                :class="props.formState.districtCode === option.value ? 'editor-select__option--active' : ''"
                @click="updateSingleSelect('districtCode', option.value)"
              >
                {{ option.label }}
              </button>
            </div>
          </div>
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
          <div class="editor-select">
            <button
              type="button"
              class="editor-select__trigger"
              :aria-expanded="openSingleSelect === 'businessStatus'"
              @click="toggleSingleSelect('businessStatus')"
            >
              <span>{{ resolveSingleSelectLabel(props.businessStatusOptions, props.formState.businessStatus) }}</span>
              <AppIcon
                name="chevron-down"
                :size="16"
              />
            </button>
            <div
              v-if="openSingleSelect === 'businessStatus'"
              class="editor-select__menu"
            >
              <button
                v-for="option in props.businessStatusOptions"
                :key="option.value"
                type="button"
                class="editor-select__option"
                :class="props.formState.businessStatus === option.value ? 'editor-select__option--active' : ''"
                @click="updateSingleSelect('businessStatus', option.value)"
              >
                {{ option.label }}
              </button>
            </div>
          </div>
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
                  :checked="props.formState.deliveryTags.includes(option.value)"
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
  gap: 2rem;
  min-width: 0;
}

.editor-section {
  border: 1px solid #e2e3e1;
  border-radius: 0.75rem;
  background: #ffffff;
  padding: 2rem;
  box-shadow: 0 10px 30px -5px rgb(0 39 39 / 0.05);
}

.editor-section-heading {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1.5rem;
}

.editor-step-number {
  display: inline-flex;
  width: 2rem;
  height: 2rem;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: #002727;
  color: #ffffff;
  font-size: 0.875rem;
  font-weight: 800;
  line-height: 1;
}

.editor-kicker,
.editor-field span {
  color: #717878;
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.editor-section-heading h2 {
  margin: 0.35rem 0 0;
  color: #002727;
  font-family: var(--font-display);
  font-size: 1.5rem;
  font-weight: 500;
  line-height: 1.4;
}

.editor-field-grid {
  display: grid;
  gap: 1.5rem;
}

.editor-field {
  display: grid;
  gap: 0.55rem;
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
  border-radius: 9999px;
  background: #ffffff;
  padding: 0.35rem 0.6rem;
  color: rgb(var(--color-primary));
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0.04em;
  line-height: 1;
  text-transform: none;
}

.editor-donation-field {
  display: flex;
  min-height: 3rem;
  align-items: center;
  gap: 0.65rem;
  align-self: end;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.5rem;
  background: #ffffff;
  padding: 0.75rem 1rem;
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
  gap: 0.85rem;
}

.editor-field input,
.editor-field select,
.editor-field textarea {
  width: 100%;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.5rem;
  background: #ffffff;
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
  height: 3rem;
  padding: 0 1rem;
}

.editor-field textarea {
  min-height: 10rem;
  padding: 0.85rem 1rem;
  resize: vertical;
}

.editor-field input::placeholder,
.editor-field textarea::placeholder {
  color: rgb(var(--color-text-muted) / 0.72);
}

.editor-field input:disabled,
.editor-field select:disabled,
.editor-field textarea:disabled {
  background: color-mix(in srgb, #ffffff 82%, rgb(var(--color-page-tint)) 18%);
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

.editor-select {
  position: relative;
  min-width: 0;
}

.editor-select__trigger,
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
  border-radius: 0.5rem;
  background: #ffffff;
  padding: 0 1rem;
  color: rgb(var(--color-text));
  font-size: 0.9375rem;
  line-height: 1.4;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease;
}

.editor-select__trigger:hover,
.editor-select__trigger[aria-expanded='true'],
.editor-multi-select__trigger:hover,
.editor-multi-select__trigger[aria-expanded='true'] {
  border-color: rgb(var(--color-primary));
  box-shadow: 0 0 0 3px rgb(var(--color-primary) / 0.16);
}

.editor-field .editor-select__trigger span,
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

.editor-select__trigger svg,
.editor-multi-select__trigger svg {
  flex-shrink: 0;
}

.editor-select__menu,
.editor-multi-select__menu {
  position: absolute;
  z-index: 10;
  top: calc(100% - 1px);
  left: 0;
  right: 0;
  display: grid;
  gap: 0.25rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0 0 0.625rem 0.625rem;
  background: #ffffff;
  max-height: 14rem;
  overflow-y: auto;
  padding: 0.5rem;
  box-shadow: 0 16px 36px rgb(0 0 0 / 0.12);
  scrollbar-color: rgb(var(--color-border)) transparent;
  scrollbar-width: thin;
}

.editor-select__menu::-webkit-scrollbar,
.editor-multi-select__menu::-webkit-scrollbar {
  width: 0.45rem;
}

.editor-select__menu::-webkit-scrollbar-thumb,
.editor-multi-select__menu::-webkit-scrollbar-thumb {
  border-radius: 9999px;
  background: rgb(var(--color-border));
}

.editor-select__option,
.editor-multi-select__option {
  display: flex;
  cursor: pointer;
  align-items: center;
  gap: 0.65rem;
  width: 100%;
  border: 0;
  border-radius: 0.5rem;
  background: #ffffff;
  padding: 0.65rem 0.7rem;
  color: rgb(var(--color-text));
  font-size: 0.875rem;
  font-weight: 700;
  line-height: 1.3;
  text-align: left;
  transition: background-color 0.2s ease;
}

.editor-select__option:hover,
.editor-select__option--active,
.editor-multi-select__option:hover {
  background: rgb(var(--color-surface));
}

.editor-select__option--active {
  color: rgb(var(--color-primary));
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
  min-height: 13rem;
  cursor: pointer;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.65rem;
  border: 2px dashed rgb(var(--color-border));
  border-radius: 0.75rem;
  background: #ffffff;
  padding: 3rem 2rem;
  text-align: center;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    color 0.2s ease;
}

.editor-dropzone:hover,
.editor-dropzone--active {
  border-color: rgb(var(--color-primary));
  background: #ffffff;
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
  gap: 1rem;
  margin-top: 1.25rem;
}

.editor-upload-slot {
  position: relative;
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: #ffffff;
}

.editor-upload-slot--cover {
  border-color: rgb(var(--color-primary));
}

.editor-upload-slot__media,
.editor-upload-slot img {
  height: 10rem;
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
  background: #ffffff;
  color: rgb(var(--color-text-muted));
  font-size: 0.8125rem;
  font-weight: 700;
}

.editor-cover-badge {
  position: absolute;
  left: 0.75rem;
  top: 0.75rem;
  border-radius: 9999px;
  background: rgb(var(--color-primary));
  padding: 0.45rem 0.75rem;
  color: #ffffff;
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
  border-radius: 9999px;
  background: rgb(255 255 255 / 0.92);
  color: #ba1a1a;
  cursor: pointer;
  font-size: 1.125rem;
  font-weight: 800;
  line-height: 1;
  box-shadow: 0 8px 18px rgb(0 0 0 / 0.12);
  transition:
    background-color 0.2s ease,
    border-color 0.2s ease,
    transform 0.2s ease;
}

.editor-image-remove-button:hover {
  border-color: #ba1a1a;
  background: #ffffff;
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
  border-radius: 9999px;
  padding: 0.4rem 0.7rem;
  background: #ffffff;
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
  color: #ffffff;
}

.editor-image-action--primary:hover {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary) / 0.9);
  color: #ffffff;
}

.editor-image-action--danger:hover {
  border-color: #ba1a1a;
  color: #ba1a1a;
}

.editor-visibility-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.5rem;
  background: #ffffff;
  padding: 1rem;
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
  margin: 1rem 0 1.5rem;
}

.editor-visibility-option {
  border: 1px solid rgb(var(--color-border));
  border-radius: 9999px;
  background: #ffffff;
  padding: 0.55rem 0.95rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.875rem;
  font-weight: 700;
  line-height: 1;
}

.editor-visibility-option--active {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: #ffffff;
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
  background: #ffffff;
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
  height: 3rem;
  min-height: 3rem;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.5rem;
  background: #ffffff;
  padding: 0 1rem;
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
