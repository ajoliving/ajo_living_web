<!--
 * 管理 - 帖子列表頁。
 * 1. 分頁查詢 Staff 可見二手帖子。
 * 2. 提供上架、下架與續期狀態操作。
-->
<script setup lang="ts">
import { computed, ref } from 'vue';

import {
  deactivateStaffSecondhandListing,
  fetchStaffSecondhandListings,
  publishStaffSecondhandListing,
  renewStaffSecondhandListing,
} from '@/httpapis/staff';
import {
  getMarketplaceCategoryLabel,
  getMarketplaceDistrictLabel,
  getMarketplaceOptionLabel,
  marketplaceCategories,
} from '@/constants/marketplace';
import type { SecondhandListingSummaryResponse } from '@/model/marketplace';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import { usePreferenceStore } from '@/stores/preferences';
import { formatDate, formatPrice } from '@/utils/format';
import { resolveListingOwnerName, resolveListingStatus } from '@/utils/marketplace';

import { useStaffManagementList } from '../composables/useStaffManagementList';
import ManagementPagination from '../widgets/ManagementPagination.vue';
import '../styles.scss';

const {
  categoryCode,
  hasNext,
  hasPrevious,
  items,
  keyword,
  loading,
  next,
  page,
  pageSize,
  previous,
  runAction,
  search,
  status,
  t,
  total,
} = useStaffManagementList<SecondhandListingSummaryResponse>({
  fetcher: fetchStaffSecondhandListings,
  loadErrorKey: 'marketplace.management.loadListingsError',
});
const renewalDialogListing = ref<SecondhandListingSummaryResponse | null>(null);
const renewalDays = ref(14);
const renewing = ref(false);
const renewalDayOptions = [7, 14, 30, 60, 90];
const preferenceStore = usePreferenceStore();
const categoryOptions = computed(() =>
  marketplaceCategories.map((category) => ({
    label: getMarketplaceOptionLabel(category, preferenceStore.locale),
    value: category.value,
  })),
);
const canSubmitRenewal = computed(() =>
  Number.isInteger(Number(renewalDays.value)) &&
  Number(renewalDays.value) > 0 &&
  Number(renewalDays.value) <= 365,
);

// 1. 格式化狀態標籤
const formatStatus = (listing: SecondhandListingSummaryResponse): string =>
  t(`common.state.${resolveListingStatus(listing)}`);

// 2. 格式化分類、價格與地區
const formatCategory = (listing: SecondhandListingSummaryResponse): string =>
  getMarketplaceCategoryLabel(listing.category_code, preferenceStore.locale);

const formatDistrict = (listing: SecondhandListingSummaryResponse): string =>
  getMarketplaceDistrictLabel(listing.district_code, preferenceStore.locale);

const formatListingPrice = (listing: SecondhandListingSummaryResponse): string =>
  listing.price_mode === 'free'
    ? t('common.price.free')
    : formatPrice(listing.price_hkd ?? 0, preferenceStore.locale);

// 3. 開啟續期彈窗
const openRenewalDialog = (listing: SecondhandListingSummaryResponse): void => {
  renewalDialogListing.value = listing;
  renewalDays.value = 14;
};

// 4. 關閉續期彈窗
const closeRenewalDialog = (): void => {
  renewalDialogListing.value = null;
  renewalDays.value = 14;
};

// 5. 提交續期設定
const submitRenewal = async (): Promise<void> => {
  if (!renewalDialogListing.value || !canSubmitRenewal.value) {
    return;
  }

  const listingId = renewalDialogListing.value.listing_id;
  const targetRenewalDays = Number(renewalDays.value);
  renewing.value = true;
  const renewed = await runAction(() => renewStaffSecondhandListing(listingId, {
    renewal_days: targetRenewalDays,
  }));
  renewing.value = false;
  if (renewed) {
    closeRenewalDialog();
  }
};
</script>

<template>
  <section class="management-list-page furniture-list-page">
    <header class="management-list-header furniture-list-header">
      <div class="furniture-list-title-block">
        <p class="management-list-kicker">
          {{ t('marketplace.management.staffKicker') }}
        </p>
        <h1>{{ t('marketplace.management.secondhandListings') }}</h1>
        <p>{{ t('marketplace.management.secondhandListingsDescription') }}</p>
      </div>
    </header>

    <article class="management-list-panel">
      <div class="management-list-toolbar furniture-list-toolbar">
        <label class="management-list-search furniture-list-search">
          <AppIcon
            name="search"
            :size="16"
          />
          <input
            v-model="keyword"
            type="search"
            :placeholder="t('marketplace.management.listSearchPlaceholder')"
            @keyup.enter="search"
          />
        </label>
        <div class="furniture-list-filter-actions">
          <label class="furniture-list-filter-field">
            <span>{{ t('marketplace.list.categoryTitle') }}</span>
            <select
              v-model="categoryCode"
              class="management-list-select"
            >
              <option value="">
                {{ t('marketplace.list.categoryTitle') }}
              </option>
              <option
                v-for="option in categoryOptions"
                :key="option.value"
                :value="option.value"
              >
                {{ option.label }}
              </option>
            </select>
          </label>
          <label class="furniture-list-filter-field">
            <span>{{ t('common.label.status') }}</span>
            <select
              v-model="status"
              class="management-list-select"
            >
              <option value="">
                {{ t('marketplace.management.allStatuses') }}
              </option>
              <option value="draft">
                {{ t('common.state.draft') }}
              </option>
              <option value="active">
                {{ t('common.state.active') }}
              </option>
              <option value="hidden">
                {{ t('common.state.hidden') }}
              </option>
              <option value="expired">
                {{ t('common.state.expired') }}
              </option>
              <option value="sold">
                {{ t('common.state.sold') }}
              </option>
            </select>
          </label>
          <button
            type="button"
            class="management-list-button furniture-list-search-button"
            @click="search"
          >
            {{ t('marketplace.list.searchAction') }}
          </button>
        </div>
      </div>

      <div
        v-if="loading"
        class="management-list-loading"
      >
        {{ t('common.status.loading') }}
      </div>

      <div
        v-else-if="items.length === 0"
        class="management-list-empty"
      >
        {{ t('marketplace.management.emptyListings') }}
      </div>

      <div
        v-else
        class="management-table-wrap"
      >
        <table class="management-table">
          <thead>
            <tr>
              <th>{{ t('marketplace.management.columnTitle') }}</th>
              <th>{{ t('marketplace.management.columnOwner') }}</th>
              <th>{{ t('marketplace.mine.category') }}</th>
              <th>{{ t('marketplace.mine.price') }}</th>
              <th>{{ t('marketplace.detail.pickupArea') }}</th>
              <th>{{ t('common.label.status') }}</th>
              <th>{{ t('marketplace.management.columnUpdatedAt') }}</th>
              <th class="management-table-actions">{{ t('marketplace.management.columnActions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="listing in items"
              :key="listing.listing_id"
            >
              <td>
                <strong>{{ listing.title }}</strong>
                <small>{{ listing.listing_id }}</small>
              </td>
              <td>{{ resolveListingOwnerName(listing) }}</td>
              <td>{{ formatCategory(listing) }}</td>
              <td>{{ formatListingPrice(listing) }}</td>
              <td>{{ formatDistrict(listing) }}</td>
              <td>
                <span class="management-status-pill">{{ formatStatus(listing) }}</span>
              </td>
              <td>{{ formatDate(listing.updated_at, preferenceStore.locale) }}</td>
              <td
                class="management-table-actions"
              >
                <div class="management-action-group">
                  <button
                    v-if="listing.publication_status !== 'active'"
                    type="button"
                    class="management-list-action"
                    @click="runAction(() => publishStaffSecondhandListing(listing.listing_id))"
                  >
                    {{ t('marketplace.management.publishAction') }}
                  </button>
                  <button
                    v-if="listing.publication_status === 'active'"
                    type="button"
                    class="management-list-action"
                    @click="runAction(() => deactivateStaffSecondhandListing(listing.listing_id))"
                  >
                    {{ t('marketplace.management.deactivateAction') }}
                  </button>
                  <button
                    type="button"
                    class="management-list-action"
                    @click="openRenewalDialog(listing)"
                  >
                    {{ t('marketplace.management.renewAction') }}
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <ManagementPagination
        :has-next="hasNext"
        :has-previous="hasPrevious"
        :loading="loading"
        :page="page"
        :page-size="pageSize"
        :t="t"
        :total="total"
        @next="next"
        @previous="previous"
      />
    </article>

    <Teleport to="body">
      <Transition name="management-dialog">
        <div
          v-if="renewalDialogListing"
          class="management-dialog"
          role="dialog"
          aria-modal="true"
          :aria-label="t('marketplace.management.renewDialogTitle')"
          @click.self="closeRenewalDialog"
        >
          <section class="management-dialog-panel">
            <header class="management-dialog-header">
              <div>
                <p>{{ t('marketplace.management.renewDialogKicker') }}</p>
                <h2>{{ t('marketplace.management.renewDialogTitle') }}</h2>
              </div>
              <button
                type="button"
                class="management-dialog-close"
                :aria-label="t('marketplace.management.renewDialogClose')"
                @click="closeRenewalDialog"
              >
                <AppIcon
                  name="close"
                  :size="18"
                />
              </button>
            </header>

            <div class="management-dialog-body">
              <div class="management-renew-listing">
                <span>{{ t('marketplace.management.renewDialogListing') }}</span>
                <strong>{{ renewalDialogListing.title }}</strong>
                <small>{{ renewalDialogListing.listing_id }}</small>
              </div>
              <label class="management-dialog-field">
                <span>{{ t('marketplace.management.renewDaysField') }}</span>
                <input
                  v-model.number="renewalDays"
                  type="number"
                  min="1"
                  max="365"
                  step="1"
                  inputmode="numeric"
                />
              </label>
              <div class="management-renew-options">
                <button
                  v-for="dayOption in renewalDayOptions"
                  :key="dayOption"
                  type="button"
                  :class="{ 'is-active': renewalDays === dayOption }"
                  @click="renewalDays = dayOption"
                >
                  {{ t('marketplace.management.renewDaysOption', { days: dayOption }) }}
                </button>
              </div>
              <p
                v-if="!canSubmitRenewal"
                class="management-dialog-error"
              >
                {{ t('marketplace.management.renewDaysInvalid') }}
              </p>
              <p class="management-dialog-hint">
                {{ t('marketplace.management.renewDialogHint') }}
              </p>
            </div>

            <footer class="management-dialog-actions">
              <button
                type="button"
                class="management-dialog-button management-dialog-button--secondary"
                @click="closeRenewalDialog"
              >
                {{ t('marketplace.management.cancelAction') }}
              </button>
              <button
                type="button"
                class="management-dialog-button management-dialog-button--primary"
                :disabled="!canSubmitRenewal || renewing"
                @click="submitRenewal"
              >
                {{ renewing ? t('marketplace.management.saving') : t('marketplace.management.renewConfirmAction') }}
              </button>
            </footer>
          </section>
        </div>
      </Transition>
    </Teleport>
  </section>
</template>

<style scoped lang="scss">
.furniture-list-header {
  align-items: stretch;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface-raised));
  padding: 1rem;
}

.furniture-list-title-block {
  min-width: 0;
  max-width: 42rem;
}

.furniture-list-toolbar {
  grid-template-columns: minmax(18rem, 1fr) auto;
  align-items: end;
  background: rgb(var(--color-surface));
}

.furniture-list-filter-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(9rem, 12rem)) auto;
  align-items: end;
  gap: 0.75rem;
}

.furniture-list-filter-field {
  display: grid;
  min-width: 0;
  gap: 0.35rem;
}

.furniture-list-filter-field span {
  color: rgb(var(--color-text-muted));
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.06em;
  line-height: 1;
  text-transform: uppercase;
}

.furniture-list-filter-field .management-list-select {
  width: 100%;
}

.furniture-list-search,
.furniture-list-search-button {
  align-self: end;
}

@media (max-width: 1023px) {
  .furniture-list-toolbar {
    grid-template-columns: 1fr;
  }

  .furniture-list-filter-actions {
    grid-template-columns: repeat(2, minmax(0, 1fr)) auto;
  }
}

@media (max-width: 767px) {
  .furniture-list-header {
    align-items: stretch;
    flex-direction: column;
  }

  .furniture-list-search-button {
    width: 100%;
  }

  .furniture-list-filter-actions {
    grid-template-columns: 1fr;
  }
}
</style>
