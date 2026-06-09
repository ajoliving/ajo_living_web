<!--
 * 展示廣告設定頁。
 * 1. 左側提供可用圖片廣告素材列表。
 * 2. 右側固定配置 10 個列表右側廣告位，每個位置可綁定多個廣告。
 * 3. 前台讀取時由後端在每個位置內隨機挑選一個廣告展示。
-->
<script setup lang="ts">
import { ref } from 'vue';
import { onBeforeRouteLeave, onBeforeRouteUpdate, useRouter, type RouteLocationNormalized } from 'vue-router';

import AppIcon from '@/shared/components/base/AppIcon.vue';
import AppUnsavedChangesDialog from '@/shared/components/base/AppUnsavedChangesDialog.vue';

import { useDisplayAdsSettingsPage } from './display-ads';

const props = defineProps<{
  channel: 'property_sale' | 'furniture' | 'serviced_apartment';
  titleKey: string;
}>();

const {
  assignedAdIds,
  canSave,
  clearSlot,
  filteredAds,
  formatDate,
  hasUnsavedChanges,
  keyword,
  layoutFilter,
  loadingAds,
  loadingSlots,
  saveSlots,
  saving,
  selectedAd,
  selectedAdId,
  selectAd,
  assignSelectedAd,
  removeAdFromSlot,
  updateSlotAdDisplayTitle,
  updateSlotAdDisplayText,
  updateSlotAdTargetURL,
  slots,
  t,
} = useDisplayAdsSettingsPage({ channel: props.channel });

const router = useRouter();
const leaveConfirmOpen = ref(false);
const pendingRoutePath = ref('');
const bypassLeaveConfirm = ref(false);

// 1. 攔截未儲存狀態下的路由切換
const requestRouteChange = (to: RouteLocationNormalized): boolean => {
  if (bypassLeaveConfirm.value || !hasUnsavedChanges.value) {
    return true;
  }

  pendingRoutePath.value = to.fullPath;
  leaveConfirmOpen.value = true;
  return false;
};

// 2. 繼續前往已暫存的目標路由
const continuePendingRoute = async (): Promise<void> => {
  const targetPath = pendingRoutePath.value;
  pendingRoutePath.value = '';
  leaveConfirmOpen.value = false;

  if (!targetPath) {
    return;
  }

  bypassLeaveConfirm.value = true;
  try {
    await router.push(targetPath);
  } finally {
    bypassLeaveConfirm.value = false;
  }
};

// 3. 保存後離開目前設定頁
const saveAndContinueRoute = async (): Promise<void> => {
  const saved = await saveSlots();
  if (saved) {
    await continuePendingRoute();
  }
};

// 4. 放棄修改並離開目前設定頁
const discardAndContinueRoute = async (): Promise<void> => {
  await continuePendingRoute();
};

// 5. 留在目前設定頁
const stayCurrentRoute = (): void => {
  pendingRoutePath.value = '';
  leaveConfirmOpen.value = false;
};

onBeforeRouteLeave((to) => requestRouteChange(to));
onBeforeRouteUpdate((to) => requestRouteChange(to));
</script>

<template>
  <section class="display-ads-page">
    <header class="display-ads-header">
      <div>
        <p class="display-ads-kicker">
          {{ t('marketplace.settings.displayAdsSection') }}
        </p>
        <h2>{{ t(titleKey) }}</h2>
      </div>
      <button
        type="button"
        class="display-ads-button display-ads-button--primary"
        :disabled="!canSave"
        @click="saveSlots"
      >
        <AppIcon
          name="check-circle"
          :size="16"
        />
        <span>{{ saving ? t('marketplace.settings.saving') : t('marketplace.settings.displayAdsSave') }}</span>
      </button>
    </header>

    <div class="display-ads-layout">
      <aside class="display-ads-library">
        <div class="display-ads-library__header">
          <div>
            <h3>{{ t('marketplace.settings.displayAdsLibraryTitle') }}</h3>
          </div>
          <span>{{ filteredAds.length }}</span>
        </div>

        <label class="display-ads-search">
          <AppIcon
            name="search"
            :size="16"
          />
          <input
            v-model="keyword"
            type="search"
            :placeholder="t('marketplace.settings.displayAdsSearchPlaceholder')"
          />
        </label>

        <div class="display-ads-layout-filter">
          <button
            type="button"
            :class="{ 'is-active': layoutFilter === '' }"
            @click="layoutFilter = ''"
          >
            {{ t('marketplace.settings.displayAdsAllLayouts') }}
          </button>
          <button
            type="button"
            :class="{ 'is-active': layoutFilter === 'text_compact' }"
            @click="layoutFilter = 'text_compact'"
          >
            {{ t('marketplace.settings.displayAdsLayout.text_compact') }}
          </button>
          <button
            type="button"
            :class="{ 'is-active': layoutFilter === 'image_text' }"
            @click="layoutFilter = 'image_text'"
          >
            {{ t('marketplace.settings.displayAdsLayout.image_text') }}
          </button>
          <button
            type="button"
            :class="{ 'is-active': layoutFilter === 'image_full' }"
            @click="layoutFilter = 'image_full'"
          >
            {{ t('marketplace.settings.displayAdsLayout.image_full') }}
          </button>
        </div>

        <div
          v-if="loadingAds"
          class="display-ads-empty"
        >
          {{ t('common.status.loading') }}
        </div>
        <div
          v-else-if="filteredAds.length === 0"
          class="display-ads-empty"
        >
          {{ t('marketplace.settings.displayAdsEmptyLibrary') }}
        </div>
        <div
          v-else
          class="display-ads-list"
        >
          <article
            v-for="ad in filteredAds"
            :key="ad.task_id"
            class="display-ads-list-item"
            :class="{
              'is-assigned': assignedAdIds.has(ad.task_id),
              'is-selected': selectedAdId === ad.task_id,
            }"
            role="button"
            tabindex="0"
            @click="selectAd(ad)"
            @keydown.enter.prevent="selectAd(ad)"
            @keydown.space.prevent="selectAd(ad)"
          >
            <strong>{{ ad.title }}</strong>
            <span>
              {{ t(`marketplace.settings.displayAdsLayout.${ad.display_layout}`) }}
              ·
              {{ t('marketplace.management.walletAdWatchCountValue', { count: ad.watch_count }) }}
              ·
              {{ ad.is_active ? t('common.state.active') : t('common.state.hidden') }}
            </span>
            <small>{{ ad.ends_at ? formatDate(ad.ends_at) : '-' }}</small>
          </article>
        </div>
      </aside>

      <section class="display-ads-slots">
        <div
          v-if="loadingSlots"
          class="display-ads-empty"
        >
          {{ t('common.status.loading') }}
        </div>

        <article
          v-for="slot in slots"
          v-else
          :key="slot.slotIndex"
          class="display-ads-slot"
          :class="`display-ads-slot--${slot.layout}`"
        >
          <div class="display-ads-slot__header">
            <div>
              <p>{{ t('marketplace.settings.displayAdsSlotLabel', { index: slot.slotIndex }) }}</p>
              <h3>{{ t(`marketplace.settings.displayAdsLayout.${slot.layout}`) }}</h3>
            </div>
            <div class="display-ads-slot__actions">
              <button
                type="button"
                class="display-ads-button display-ads-button--compact"
                :disabled="!selectedAd || selectedAd.display_layout !== slot.layout"
                @click="assignSelectedAd(slot)"
              >
                {{ t('marketplace.settings.displayAdsAssignAction') }}
              </button>
              <button
                type="button"
                class="display-ads-link-button"
                @click="clearSlot(slot)"
              >
                {{ t('marketplace.settings.displayAdsClearSlot') }}
              </button>
            </div>
          </div>

          <div
            v-if="slot.ads.length === 0"
            class="display-ads-slot__empty"
          >
            {{ t('marketplace.settings.displayAdsEmptySlot') }}
          </div>

          <div
            v-else
            class="display-ads-slot__chips"
          >
            <div
              v-for="ad in slot.ads"
              :key="ad.task_id"
              class="display-ads-slot-ad"
            >
              <div class="display-ads-slot-ad__header">
                <strong>{{ ad.title }}</strong>
                <button
                  type="button"
                  :aria-label="t('marketplace.settings.displayAdsRemoveAd')"
                  @click="removeAdFromSlot(slot, ad)"
                >
                  <AppIcon
                    name="close"
                    :size="12"
                  />
                </button>
              </div>
              <template v-if="slot.layout === 'image_text'">
                <input
                  :value="ad.slot_display_title"
                  type="text"
                  maxlength="160"
                  :placeholder="t('marketplace.settings.displayAdsDisplayTitlePlaceholder')"
                  @input="updateSlotAdDisplayTitle(ad, ($event.target as HTMLInputElement).value)"
                />
                <input
                  :value="ad.display_text"
                  type="text"
                  maxlength="255"
                  :placeholder="t('marketplace.settings.displayAdsDisplayTextPlaceholder')"
                  @input="updateSlotAdDisplayText(ad, ($event.target as HTMLInputElement).value)"
                />
              </template>
              <input
                :value="ad.slot_target_url"
                type="url"
                maxlength="1024"
                :placeholder="t('marketplace.settings.displayAdsTargetUrlPlaceholder')"
                @input="updateSlotAdTargetURL(ad, ($event.target as HTMLInputElement).value)"
              />
            </div>
          </div>
        </article>
      </section>
    </div>

    <AppUnsavedChangesDialog
      :open="leaveConfirmOpen"
      :title="t('marketplace.settings.displayAdsLeaveConfirmTitle')"
      :description="t('marketplace.settings.displayAdsLeaveConfirmDescription')"
      :save-label="saving ? t('marketplace.settings.saving') : t('marketplace.settings.displayAdsLeaveConfirmSave')"
      :discard-label="t('marketplace.settings.displayAdsLeaveConfirmDiscard')"
      :stay-label="t('marketplace.settings.displayAdsLeaveConfirmStay')"
      :saving="saving"
      @save="saveAndContinueRoute"
      @discard="discardAndContinueRoute"
      @stay="stayCurrentRoute"
    />
  </section>
</template>

<style scoped>
.display-ads-page {
  display: grid;
  gap: 1rem;
}

.display-ads-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.display-ads-kicker {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.display-ads-header h2 {
  margin: 0.35rem 0 0;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: clamp(1.5rem, 2vw, 2.25rem);
  font-weight: 650;
  line-height: 1.2;
}

.display-ads-header p:not(.display-ads-kicker),
.display-ads-library__header p {
  margin: 0.45rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.9rem;
  line-height: 1.65;
}

.display-ads-button {
  display: inline-flex;
  min-height: 2.5rem;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 0 0.85rem;
  color: rgb(var(--color-text));
  font-size: 0.85rem;
  font-weight: 750;
}

.display-ads-button--primary {
  border-color: rgb(var(--color-text));
  background: rgb(var(--color-text));
  color: rgb(255 255 255);
}

.display-ads-button--compact {
  min-height: 2.15rem;
  border-radius: 6px;
  padding: 0 0.7rem;
}

.display-ads-button:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.display-ads-layout {
  display: grid;
  grid-template-columns: minmax(16rem, 0.36fr) minmax(0, 0.64fr);
  gap: 1rem;
  align-items: start;
}

.display-ads-library,
.display-ads-slot {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
}

.display-ads-library {
  display: grid;
  gap: 0.85rem;
  padding: 1rem;
}

.display-ads-library__header {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
}

.display-ads-library__header h3,
.display-ads-slot__header h3 {
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 1rem;
  font-weight: 800;
}

.display-ads-library__header > span {
  display: grid;
  width: 2.4rem;
  height: 2.4rem;
  place-items: center;
  border-radius: 8px;
  background: rgb(var(--color-primary-soft));
  color: rgb(var(--color-primary));
  font-weight: 800;
}

.display-ads-search {
  display: grid;
  gap: 0.5rem;
}

.display-ads-search {
  position: relative;
}

.display-ads-search svg {
  position: absolute;
  top: 50%;
  left: 0.75rem;
  transform: translateY(-50%);
  color: rgb(var(--color-text-muted));
}

.display-ads-search input {
  min-height: 2.6rem;
  width: 100%;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface-raised));
  color: rgb(var(--color-text));
  font-size: 0.88rem;
  outline: 0;
}

.display-ads-search input {
  padding: 0 0.8rem 0 2.3rem;
}

.display-ads-layout-filter {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
}

.display-ads-layout-filter button {
  min-height: 2rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 999px;
  background: rgb(var(--color-surface-raised));
  padding: 0 0.65rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.76rem;
  font-weight: 750;
}

.display-ads-layout-filter button.is-active {
  border-color: rgb(var(--color-primary) / 0.45);
  background: rgb(var(--color-primary-soft));
  color: rgb(var(--color-primary));
}

.display-ads-list,
.display-ads-slots {
  display: grid;
  gap: 0.75rem;
}

.display-ads-list-item {
  display: grid;
  gap: 0.25rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface-raised));
  padding: 0.75rem;
  cursor: pointer;
}

.display-ads-list-item.is-assigned {
  border-color: rgb(var(--color-primary) / 0.28);
}

.display-ads-list-item.is-selected {
  border-color: rgb(var(--color-primary) / 0.5);
  background: rgb(var(--color-primary-soft));
}

.display-ads-list-item strong {
  overflow: hidden;
  color: rgb(var(--color-text));
  font-size: 0.9rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.display-ads-list-item span,
.display-ads-list-item small,
.display-ads-slot__header p,
.display-ads-slot__empty {
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  line-height: 1.45;
}

.display-ads-slot {
  display: grid;
  gap: 0.85rem;
  padding: 1rem;
}

.display-ads-slot--text_compact {
  min-height: 10rem;
}

.display-ads-slot--image_text {
  min-height: 11rem;
}

.display-ads-slot--image_full {
  min-height: 14rem;
}

.display-ads-slot__header {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
}

.display-ads-slot__actions {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  gap: 0.55rem;
}

.display-ads-slot__header p {
  margin: 0 0 0.25rem;
  font-weight: 800;
  text-transform: uppercase;
}

.display-ads-link-button {
  border: 0;
  background: transparent;
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 800;
}

.display-ads-slot__chips {
  display: grid;
  gap: 0.55rem;
}

.display-ads-slot-ad {
  display: grid;
  gap: 0.35rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface-raised));
  padding: 0.6rem;
}

.display-ads-slot-ad__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.55rem;
}

.display-ads-slot-ad strong {
  overflow: hidden;
  color: rgb(var(--color-text));
  font-size: 0.78rem;
  font-weight: 750;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.display-ads-slot-ad input {
  min-height: 2.15rem;
  width: 100%;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: rgb(var(--color-surface));
  padding: 0 0.65rem;
  color: rgb(var(--color-text));
  font-size: 0.78rem;
  outline: 0;
}

.display-ads-slot-ad input:focus {
  border-color: rgb(var(--color-primary) / 0.5);
}

.display-ads-slot-ad__header button {
  display: grid;
  flex-shrink: 0;
  width: 1.25rem;
  height: 1.25rem;
  place-items: center;
  border: 0;
  border-radius: 999px;
  background: rgb(var(--color-surface-muted));
  color: rgb(var(--color-text-muted));
}

.display-ads-empty {
  border: 1px dashed rgb(var(--color-border));
  border-radius: 8px;
  padding: 1rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.88rem;
  line-height: 1.6;
}

@media (max-width: 900px) {
  .display-ads-header,
  .display-ads-slot__header {
    flex-direction: column;
  }

  .display-ads-layout {
    grid-template-columns: 1fr;
  }

  .display-ads-slot__actions {
    width: 100%;
    justify-content: flex-start;
  }
}
</style>
