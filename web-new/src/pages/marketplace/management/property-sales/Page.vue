<!--
 * 管理 - 樓盤列表頁。
 * 1. 分頁查詢 Staff 可見樓盤放售。
 * 2. 提供上架、下架與續期狀態操作。
-->
<script setup lang="ts">
import { computed, ref } from 'vue';

import {
  deactivateStaffPropertySale,
  fetchStaffPropertySales,
  publishStaffPropertySale,
  renewStaffPropertySale,
} from '@/httpapis/staff';
import type { PropertyListingSummaryResponse } from '@/model/property';
import PropertyEditorPage from '@/pages/property/editor/PropertyEditorPage.vue';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import { formatDate } from '@/utils/format';
import { resolvePropertyPriceText, resolvePropertyPublisherRole, resolvePropertyStatus } from '@/utils/property';

import { useStaffManagementList } from '../composables/useStaffManagementList';
import ManagementPagination from '../widgets/ManagementPagination.vue';
import '../styles.scss';

type PropertyEditorDialogInstance = InstanceType<typeof PropertyEditorPage> & {
  requestCloseEditor: () => Promise<void>;
};

const {
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
} = useStaffManagementList<PropertyListingSummaryResponse>({
  fetcher: fetchStaffPropertySales,
  loadErrorKey: 'marketplace.management.loadPropertySalesError',
});
const editorOpen = ref(false);
const editorListingId = ref('');
const propertyEditorDialog = ref<PropertyEditorDialogInstance | null>(null);
const editorKey = computed(() => `staff-sale-${editorListingId.value}-${editorOpen.value ? 'open' : 'closed'}`);

// 1. 格式化狀態標籤
const formatStatus = (listing: PropertyListingSummaryResponse): string =>
  t(`common.state.${resolvePropertyStatus(listing)}`);

// 2. 格式化發布者
const formatOwner = (listing: PropertyListingSummaryResponse): string =>
  listing.owner?.display_name?.trim() || resolvePropertyPublisherRole(listing);

// 3. 格式化價格
const formatPriceText = (listing: PropertyListingSummaryResponse): string =>
  resolvePropertyPriceText(listing, 'zh-HK');

// 4. 開啟管理編輯器
const openEditor = (listingId: string): void => {
  editorListingId.value = listingId;
  editorOpen.value = true;
};

// 5. 關閉管理編輯器
const closeEditor = (): void => {
  editorOpen.value = false;
  editorListingId.value = '';
};

// 6. 請求關閉管理編輯器
const requestCloseEditor = (): void => {
  void propertyEditorDialog.value?.requestCloseEditor();
};

// 7. 完成管理編輯後刷新列表
const handleEditorDone = async (): Promise<void> => {
  closeEditor();
  await search();
};
</script>

<template>
  <section class="management-list-page">
    <header class="management-list-header">
      <div>
        <p class="management-list-kicker">
          Staff
        </p>
        <h1>{{ t('marketplace.management.propertySales') }}</h1>
        <p>{{ t('marketplace.management.propertySalesDescription') }}</p>
      </div>
    </header>

    <article class="management-list-panel">
      <div class="management-list-toolbar">
        <label class="management-list-search">
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
        <button
          type="button"
          class="management-list-button"
          @click="search"
        >
          {{ t('marketplace.list.searchAction') }}
        </button>
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
              <th>{{ t('marketplace.management.columnPrice') }}</th>
              <th>{{ t('common.label.status') }}</th>
              <th>{{ t('marketplace.management.columnUpdatedAt') }}</th>
              <th>{{ t('marketplace.management.columnActions') }}</th>
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
              <td>{{ formatOwner(listing) }}</td>
              <td>{{ formatPriceText(listing) }}</td>
              <td>
                <span class="management-status-pill">{{ formatStatus(listing) }}</span>
              </td>
              <td>{{ formatDate(listing.updated_at) }}</td>
              <td
                class="management-table-actions"
              >
                <div class="management-action-group">
                  <button
                    type="button"
                    class="management-list-action"
                    @click="openEditor(listing.listing_id)"
                  >
                    {{ t('property.mine.edit') }}
                  </button>
                  <button
                    v-if="listing.publication_status !== 'active'"
                    type="button"
                    class="management-list-action"
                    @click="runAction(() => publishStaffPropertySale(listing.listing_id))"
                  >
                    {{ t('marketplace.management.publishAction') }}
                  </button>
                  <button
                    v-if="listing.publication_status === 'active'"
                    type="button"
                    class="management-list-action"
                    @click="runAction(() => deactivateStaffPropertySale(listing.listing_id))"
                  >
                    {{ t('marketplace.management.deactivateAction') }}
                  </button>
                  <button
                    type="button"
                    class="management-list-action"
                    @click="runAction(() => renewStaffPropertySale(listing.listing_id))"
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
          v-if="editorOpen"
          class="management-dialog"
          role="dialog"
          aria-modal="true"
          :aria-label="t('property.editor.editMode')"
          @click.self="requestCloseEditor"
        >
          <section class="management-dialog-panel management-dialog-panel--editor">
            <header class="management-dialog-header">
              <div>
                <p>Staff</p>
                <h2>{{ t('property.editor.editMode') }}</h2>
              </div>
              <button
                type="button"
                class="management-dialog-close"
                :aria-label="t('marketplace.management.closeDialog')"
                @click="requestCloseEditor"
              >
                <AppIcon
                  name="close"
                  :size="18"
                />
              </button>
            </header>

            <div class="management-dialog-body management-dialog-body--editor">
              <PropertyEditorPage
                ref="propertyEditorDialog"
                :key="editorKey"
                channel="sale"
                :listing-id="editorListingId"
                embedded
                staff-mode
                hide-header
                @cancel="closeEditor"
                @published="handleEditorDone"
                @saved="handleEditorDone"
              />
            </div>
          </section>
        </div>
      </Transition>
    </Teleport>
  </section>
</template>
