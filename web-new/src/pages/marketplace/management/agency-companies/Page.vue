<!--
 * 管理端 - 代理資料審核。
 * 1. 分頁查詢個人代理及代理公司資料，支援類型、關鍵字及狀態篩選。
 * 2. 開啟資料詳情、檢視 OSS 證明文件及提交通過或拒絕決定。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import {
  fetchStaffAgencyProfiles,
  fetchStaffAgencyProfileDetail,
  reviewStaffAgencyProfile,
} from '@/httpapis/agency-profiles';
import type { AgencyProfileStatus, AgencyProfileType, StaffAgencyProfile } from '@/model/agency-profile';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';
import { formatDate } from '@/utils/format';

import ManagementPagination from '../widgets/ManagementPagination.vue';
import '../styles.scss';
import AgencyCompanyReviewDialog from './widgets/AgencyCompanyReviewDialog.vue';

const PAGE_SIZE = 20;

const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const preferenceStore = usePreferenceStore();
const keyword = ref('');
const status = ref<AgencyProfileStatus | ''>('pending');
const profileType = ref<AgencyProfileType | ''>('');
const page = ref(1);
const total = ref(0);
const items = ref<StaffAgencyProfile[]>([]);
const loading = ref(false);
const detailLoadingProfileId = ref('');
const reviewing = ref(false);
const selectedCompany = ref<StaffAgencyProfile | null>(null);
const reviewNote = ref('');

// 1. 正規化 API 錯誤文案
const readErrorMessage = (error: unknown, fallback: string): string =>
  axios.isAxiosError<{ message?: string }>(error)
    ? error.response?.data?.message ?? fallback
    : fallback;

// 2. 計算分頁操作狀態
const hasPrevious = computed(() => page.value > 1);
const hasNext = computed(() => page.value * PAGE_SIZE < total.value);

// 3. 讀取代理公司審核列表
const loadCompanies = async (targetPage = page.value): Promise<void> => {
  loading.value = true;

  try {
    const { data } = await fetchStaffAgencyProfiles({
      page: targetPage,
      page_size: PAGE_SIZE,
      keyword: keyword.value.trim() || undefined,
      status: status.value || 'all',
      profile_type: profileType.value || undefined,
    });
    items.value = data.data.items;
    page.value = data.data.pagination.page;
    total.value = data.data.pagination.total;
  } catch (error) {
    feedbackStore.pushToast(readErrorMessage(error, t('marketplace.management.agencyCompany.loadError')), 'error');
  } finally {
    loading.value = false;
  }
};

// 4. 搜尋代理公司審核資料
const search = async (): Promise<void> => {
  await loadCompanies(1);
};

// 5. 開啟代理公司資料詳情
const openDetail = async (profileId: string): Promise<void> => {
  detailLoadingProfileId.value = profileId;

  try {
    const { data } = await fetchStaffAgencyProfileDetail(profileId);
    selectedCompany.value = data.data;
    reviewNote.value = '';
  } catch (error) {
    feedbackStore.pushToast(readErrorMessage(error, t('marketplace.management.agencyCompany.detailError')), 'error');
  } finally {
    detailLoadingProfileId.value = '';
  }
};

// 6. 關閉代理公司資料詳情
const closeDetail = (): void => {
  if (!reviewing.value) {
    selectedCompany.value = null;
    reviewNote.value = '';
  }
};

// 7. 提交通過或拒絕審核
const submitReview = async (approved: boolean): Promise<void> => {
  if (!selectedCompany.value) {
    return;
  }
  if (!approved && !reviewNote.value.trim()) {
    feedbackStore.pushToast(t('marketplace.management.agencyCompany.rejectReasonRequired'), 'error');
    return;
  }

  reviewing.value = true;
  try {
    await reviewStaffAgencyProfile(selectedCompany.value.profile_id, {
      approved,
      review_note: reviewNote.value.trim(),
    });
    feedbackStore.pushToast(
      approved
        ? t('marketplace.management.agencyCompany.approveSuccess')
        : t('marketplace.management.agencyCompany.rejectSuccess'),
      'success',
    );
    selectedCompany.value = null;
    reviewNote.value = '';
    await loadCompanies(page.value);
  } catch (error) {
    feedbackStore.pushToast(readErrorMessage(error, t('marketplace.management.agencyCompany.reviewError')), 'error');
  } finally {
    reviewing.value = false;
  }
};

// 8. 顯示列表審核狀態
const statusLabel = (companyStatus: AgencyProfileStatus): string =>
  t(`marketplace.management.agencyCompany.status.${companyStatus}`);

// 9. 顯示列表會員名稱
const ownerName = (company: StaffAgencyProfile): string =>
  company.owner.display_name || company.owner.email || company.owner.public_id || '-';

// 10. 顯示代理資料名稱
const profileName = (profile: StaffAgencyProfile): string => profile.name_zh || profile.name_en || '-';

watch(status, () => {
  void loadCompanies(1);
});

watch(profileType, () => {
  void loadCompanies(1);
});

onMounted(() => {
  void loadCompanies();
});
</script>

<template>
  <section class="management-list-page agency-company-management-page">
    <header class="management-list-header">
      <div>
        <p class="management-list-kicker">
          {{ t('marketplace.management.agencyCompany.kicker') }}
        </p>
        <h1>{{ t('marketplace.management.agencyCompany.title') }}</h1>
        <p>{{ t('marketplace.management.agencyCompany.description') }}</p>
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
            :placeholder="t('marketplace.management.agencyCompany.searchPlaceholder')"
            @keyup.enter="search"
          >
        </label>
        <select
          v-model="profileType"
          class="management-list-select"
        >
          <option value="">{{ t('marketplace.management.agencyCompany.allTypes') }}</option>
          <option value="individual">{{ t('marketplace.management.agencyCompany.individualType') }}</option>
          <option value="company">{{ t('marketplace.management.agencyCompany.companyType') }}</option>
        </select>
        <select
          v-model="status"
          class="management-list-select"
        >
          <option value="">
            {{ t('marketplace.management.agencyCompany.allStatuses') }}
          </option>
          <option value="pending">
            {{ t('marketplace.management.agencyCompany.status.pending') }}
          </option>
          <option value="draft">
            {{ t('marketplace.management.agencyCompany.status.draft') }}
          </option>
          <option value="rejected">
            {{ t('marketplace.management.agencyCompany.status.rejected') }}
          </option>
          <option value="approved">
            {{ t('marketplace.management.agencyCompany.status.approved') }}
          </option>
        </select>
        <button
          type="button"
          class="management-list-button agency-search-button"
          :aria-label="t('marketplace.list.searchAction')"
          :title="t('marketplace.list.searchAction')"
          @click="search"
        >
          <AppIcon
            name="search"
            :size="17"
          />
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
        {{ t('marketplace.management.agencyCompany.empty') }}
      </div>

      <div
        v-else
        class="management-table-wrap"
      >
        <table class="management-table agency-company-management-table">
          <thead>
            <tr>
              <th>{{ t('marketplace.management.agencyCompany.profile') }}</th>
              <th>{{ t('marketplace.management.agencyCompany.licenseNumber') }}</th>
              <th>{{ t('marketplace.management.agencyCompany.owner') }}</th>
              <th>{{ t('common.label.status') }}</th>
              <th>{{ t('marketplace.management.agencyCompany.updatedAt') }}</th>
              <th>{{ t('marketplace.management.columnActions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="company in items"
              :key="company.profile_id"
            >
              <td>
                <strong>{{ profileName(company) }}</strong>
                <small>{{ t(`marketplace.management.agencyCompany.${company.profile_type}Type`) }}</small>
              </td>
              <td>{{ company.license_number || '-' }}</td>
              <td>
                <strong>{{ ownerName(company) }}</strong>
                <small>{{ company.owner.email || company.owner.public_id || '-' }}</small>
              </td>
              <td>
                <span
                  class="management-status-pill agency-status-pill"
                  :class="`agency-status-pill--${company.status}`"
                >
                  {{ statusLabel(company.status) }}
                </span>
              </td>
              <td class="management-table-nowrap">
                {{ formatDate(company.updated_at, preferenceStore.locale) }}
              </td>
              <td class="management-table-actions">
                <button
                  type="button"
                  class="management-list-action"
                  :disabled="detailLoadingProfileId !== ''"
                  @click="openDetail(company.profile_id)"
                >
                  <AppIcon
                    name="view"
                    :size="15"
                  />
                  {{ detailLoadingProfileId === company.profile_id ? t('common.status.loading') : t('marketplace.management.agencyCompany.reviewAction') }}
                </button>
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
        :page-size="PAGE_SIZE"
        :t="t"
        :total="total"
        @next="loadCompanies(page + 1)"
        @previous="loadCompanies(page - 1)"
      />
    </article>

    <AgencyCompanyReviewDialog
      v-if="selectedCompany"
      v-model:review-note="reviewNote"
      :company="selectedCompany"
      :is-reviewing="reviewing"
      @approve="submitReview(true)"
      @close="closeDetail"
      @reject="submitReview(false)"
    />
  </section>
</template>

<style scoped>
.agency-company-management-page :deep(.management-list-toolbar) {
  grid-template-columns: minmax(16rem, 1fr) minmax(9rem, 11rem) minmax(9rem, 11rem) 2.45rem;
  align-items: center;
}

.agency-search-button {
  width: 2.45rem;
  min-width: 2.45rem;
  padding: 0;
}

.agency-company-management-table {
  min-width: 58rem;
}

.agency-company-management-table th:first-child {
  width: 22%;
}

.agency-company-management-table th:nth-child(2) {
  width: 15%;
}

.agency-company-management-table th:nth-child(3) {
  width: 24%;
}

.agency-company-management-table th:nth-child(4) {
  width: 13%;
}

.agency-company-management-table th:nth-child(5) {
  width: 14%;
}

.agency-company-management-table th:last-child {
  width: 12%;
}

.agency-status-pill--pending {
  background: rgb(var(--color-warning-bg));
  color: rgb(var(--color-warning));
}

.agency-status-pill--approved {
  background: rgb(var(--color-success-bg));
  color: rgb(var(--color-success));
}

.agency-status-pill--rejected {
  background: rgb(var(--color-danger-bg));
  color: rgb(var(--color-danger));
}

.agency-status-pill--draft {
  background: rgb(var(--color-surface-muted));
  color: rgb(var(--color-text-muted));
}

@media (max-width: 767px) {
  .agency-company-management-page :deep(.management-list-toolbar) {
    grid-template-columns: repeat(2, minmax(0, 1fr)) 2.45rem;
  }

  .agency-company-management-page :deep(.management-list-search) {
    grid-column: 1 / -1;
  }
}
</style>
