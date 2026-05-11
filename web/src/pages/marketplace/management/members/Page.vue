<!--
 * 管理 - 會員列表頁。
 * 1. 分頁查詢 Staff 可見會員。
 * 2. 支援關鍵字與會員狀態篩選。
-->
<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { fetchStaffUsers } from '@/httpapis/staff';
import type { StaffUserSummary } from '@/model/user';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import { useFeedbackStore } from '@/stores/feedback';
import { formatDate } from '@/utils/format';

import ManagementPagination from '../widgets/ManagementPagination.vue';
import '../styles.scss';

const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const loading = ref(false);
const keyword = ref('');
const status = ref('');
const page = ref(1);
const pageSize = 20;
const total = ref(0);
const items = ref<StaffUserSummary[]>([]);
const hasPrevious = computed(() => page.value > 1);
const hasNext = computed(() => page.value * pageSize < total.value);

// 1. 讀取會員列表
const loadMembers = async (targetPage = page.value): Promise<void> => {
  loading.value = true;

  try {
    const { data } = await fetchStaffUsers({
      page: targetPage,
      page_size: pageSize,
      keyword: keyword.value.trim() || undefined,
      status: status.value || undefined,
    });
    items.value = data.data.items;
    page.value = data.data.pagination.page;
    total.value = data.data.pagination.total;
  } catch {
    feedbackStore.pushToast(t('marketplace.management.loadMembersError'), 'error');
  } finally {
    loading.value = false;
  }
};

// 2. 搜尋會員
const search = async (): Promise<void> => {
  await loadMembers(1);
};

// 3. 格式化會員名稱
const formatMemberName = (member: StaffUserSummary): string =>
  member.display_name?.trim() || member.phone_number || member.public_id;

// 4. 格式化會員電話
const formatPhone = (member: StaffUserSummary): string =>
  member.phone_number ? `${member.phone_country_code} ${member.phone_number}`.trim() : '-';

watch(status, () => {
  void loadMembers(1);
});

onMounted(() => {
  void loadMembers(1);
});
</script>

<template>
  <section class="management-list-page">
    <header class="management-list-header">
      <div>
        <p class="management-list-kicker">
          Staff
        </p>
        <h1>{{ t('marketplace.management.members') }}</h1>
        <p>{{ t('marketplace.management.membersDescription') }}</p>
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
            :placeholder="t('marketplace.management.memberSearchPlaceholder')"
            @keyup.enter="search"
          />
        </label>
        <select
          v-model="status"
          class="management-list-select"
        >
          <option value="">
            {{ t('marketplace.management.allMemberStatuses') }}
          </option>
          <option value="active">
            {{ t('marketplace.management.memberActive') }}
          </option>
          <option value="inactive">
            {{ t('marketplace.management.memberInactive') }}
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
        {{ t('marketplace.management.emptyMembers') }}
      </div>

      <div
        v-else
        class="management-table-wrap"
      >
        <table class="management-table">
          <thead>
            <tr>
              <th>{{ t('marketplace.management.columnMember') }}</th>
              <th>{{ t('marketplace.management.columnPhone') }}</th>
              <th>{{ t('marketplace.management.columnMemberType') }}</th>
              <th>{{ t('common.label.status') }}</th>
              <th>{{ t('marketplace.management.columnCommunity') }}</th>
              <th>{{ t('marketplace.management.columnUpdatedAt') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="member in items"
              :key="member.public_id"
            >
              <td>
                <strong>{{ formatMemberName(member) }}</strong>
                <small>{{ member.public_id }}</small>
              </td>
              <td>{{ formatPhone(member) }}</td>
              <td>{{ member.member_type || '-' }}</td>
              <td>
                <span class="management-status-pill">{{ member.member_status || '-' }}</span>
              </td>
              <td>
                {{ member.primary_community?.name_zh || member.primary_community?.name_en || '-' }}
              </td>
              <td>{{ formatDate(member.updated_at) }}</td>
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
        @next="loadMembers(page + 1)"
        @previous="loadMembers(page - 1)"
      />
    </article>
  </section>
</template>
