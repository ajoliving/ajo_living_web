<!--
 * 管理 - 帖子列表頁。
 * 1. 分頁查詢 Staff 可見二手帖子。
 * 2. 提供上架、下架與續期狀態操作。
-->
<script setup lang="ts">
import {
  deactivateStaffSecondhandListing,
  fetchStaffSecondhandListings,
  publishStaffSecondhandListing,
  renewStaffSecondhandListing,
} from '@/httpapis/staff';
import type { SecondhandListingSummaryResponse } from '@/model/marketplace';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import { formatDate } from '@/utils/format';
import { resolveListingOwnerName, resolveListingStatus } from '@/utils/marketplace';

import { useStaffManagementList } from '../composables/useStaffManagementList';
import ManagementPagination from '../widgets/ManagementPagination.vue';
import '../styles.scss';

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
} = useStaffManagementList<SecondhandListingSummaryResponse>({
  fetcher: fetchStaffSecondhandListings,
  loadErrorKey: 'marketplace.management.loadListingsError',
});

// 1. 格式化狀態標籤
const formatStatus = (listing: SecondhandListingSummaryResponse): string =>
  t(`common.state.${resolveListingStatus(listing)}`);

// 2. 格式化可選日期
const formatOptionalDate = (value?: string | null): string =>
  value ? formatDate(value) : '-';
</script>

<template>
  <section class="management-list-page">
    <header class="management-list-header">
      <div>
        <p class="management-list-kicker">
          Staff
        </p>
        <h1>{{ t('marketplace.management.secondhandListings') }}</h1>
        <p>{{ t('marketplace.management.secondhandListingsDescription') }}</p>
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
              <th>{{ t('common.label.status') }}</th>
              <th>{{ t('common.label.publishedAt') }}</th>
              <th>{{ t('common.label.expiresAt') }}</th>
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
              <td>{{ resolveListingOwnerName(listing) }}</td>
              <td>
                <span class="management-status-pill">{{ formatStatus(listing) }}</span>
              </td>
              <td>{{ formatOptionalDate(listing.published_at) }}</td>
              <td>{{ formatOptionalDate(listing.expire_at) }}</td>
              <td>{{ formatDate(listing.updated_at) }}</td>
              <td>
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
                    @click="runAction(() => renewStaffSecondhandListing(listing.listing_id))"
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
  </section>
</template>
