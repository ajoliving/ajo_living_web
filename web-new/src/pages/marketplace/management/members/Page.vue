<!--
 * 管理 - 會員列表頁。
 * 1. 分頁查詢 Staff 可見會員。
 * 2. 支援新增 Staff 帳戶與對會員發放積分。
-->
<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { createStaffUser, fetchStaffRoles, fetchStaffUsers } from '@/httpapis/staff';
import { grantStaffWalletPoints } from '@/httpapis/wallet';
import type { RoleCatalogItem, StaffUserSummary } from '@/model/user';
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
const roleOptions = ref<RoleCatalogItem[]>([]);
const isCreateDialogOpen = ref(false);
const creating = ref(false);
const grantTarget = ref<StaffUserSummary | null>(null);
const granting = ref(false);
const createForm = ref({
  email: '',
  password: '',
  displayName: '',
  phoneCountryCode: '+852',
  phoneNumber: '',
  memberType: 'user',
  roleCode: 'staff',
});
const grantForm = ref({
  amount: 100,
  note: '',
});
const hasPrevious = computed(() => page.value > 1);
const hasNext = computed(() => page.value * pageSize < total.value);
const staffRoleOptions = computed(() =>
  roleOptions.value.filter((role) => role.scope === 'staff'),
);
const canCreateMember = computed(() =>
  createForm.value.email.trim().length > 0 &&
  createForm.value.password.trim().length >= 8 &&
  createForm.value.phoneCountryCode.trim().length > 0 &&
  createForm.value.phoneNumber.trim().length > 0 &&
  createForm.value.roleCode.trim().length > 0,
);
const canGrantPoints = computed(() =>
  grantTarget.value !== null &&
  Number.isFinite(Number(grantForm.value.amount)) &&
  Number(grantForm.value.amount) > 0 &&
  grantForm.value.note.trim().length > 0,
);

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

// 2. 讀取角色列表
const loadRoles = async (): Promise<void> => {
  try {
    const { data } = await fetchStaffRoles();
    roleOptions.value = data.data.items;
  } catch {
    feedbackStore.pushToast(t('marketplace.management.loadRolesError'), 'error');
  }
};

// 3. 搜尋會員
const search = async (): Promise<void> => {
  await loadMembers(1);
};

// 4. 格式化會員名稱
const formatMemberName = (member: StaffUserSummary): string =>
  member.display_name?.trim() || member.email || member.phone_number || '-';

// 5. 格式化會員電話
const formatPhone = (member: StaffUserSummary): string =>
  member.phone_number ? `${member.phone_country_code} ${member.phone_number}`.trim() : '-';

// 6. 格式化會員電郵
const formatEmail = (member: StaffUserSummary): string => member.email?.trim() || '-';

// 7. 格式化會員角色
const formatRoles = (member: StaffUserSummary): string =>
  member.roles.filter((role) => role !== 'member').join(', ') || '-';

// 8. 開啟新增帳戶彈窗
const openCreateDialog = (): void => {
  createForm.value = {
    email: '',
    password: '',
    displayName: '',
    phoneCountryCode: '+852',
    phoneNumber: '',
    memberType: 'user',
    roleCode: 'staff',
  };
  isCreateDialogOpen.value = true;
};

// 9. 關閉新增帳戶彈窗
const closeCreateDialog = (): void => {
  if (!creating.value) {
    isCreateDialogOpen.value = false;
  }
};

// 10. 新增 Staff 帳戶
const submitCreateMember = async (): Promise<void> => {
  if (!canCreateMember.value) {
    feedbackStore.pushToast(t('marketplace.management.createMemberRequired'), 'error');
    return;
  }

  creating.value = true;
  try {
    await createStaffUser({
      email: createForm.value.email.trim(),
      password: createForm.value.password,
      display_name: createForm.value.displayName.trim(),
      phone_country_code: createForm.value.phoneCountryCode.trim(),
      phone_number: createForm.value.phoneNumber.trim(),
      member_type: createForm.value.memberType,
      role_codes: [createForm.value.roleCode],
    });
    feedbackStore.pushToast(t('marketplace.management.createMemberSuccess'), 'success');
    isCreateDialogOpen.value = false;
    await loadMembers(1);
  } catch {
    feedbackStore.pushToast(t('marketplace.management.createMemberError'), 'error');
  } finally {
    creating.value = false;
  }
};

// 11. 開啟積分發放彈窗
const openGrantDialog = (member: StaffUserSummary): void => {
  grantTarget.value = member;
  grantForm.value = {
    amount: 100,
    note: '',
  };
};

// 12. 關閉積分發放彈窗
const closeGrantDialog = (): void => {
  if (!granting.value) {
    grantTarget.value = null;
  }
};

// 13. 發放積分
const submitGrantPoints = async (): Promise<void> => {
  if (!grantTarget.value || !canGrantPoints.value) {
    feedbackStore.pushToast(t('marketplace.management.walletGrantRequired'), 'error');
    return;
  }

  granting.value = true;
  try {
    await grantStaffWalletPoints({
      user_id: grantTarget.value.public_id,
      amount: Number(grantForm.value.amount),
      note: grantForm.value.note.trim(),
    });
    feedbackStore.pushToast(t('marketplace.management.walletGrantSuccess'), 'success');
    grantTarget.value = null;
  } catch {
    feedbackStore.pushToast(t('marketplace.management.walletGrantError'), 'error');
  } finally {
    granting.value = false;
  }
};

watch(status, () => {
  void loadMembers(1);
});

onMounted(() => {
  void loadMembers(1);
  void loadRoles();
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
      <button
        type="button"
        class="management-list-button"
        @click="openCreateDialog"
      >
        <AppIcon
          name="plus-square"
          :size="16"
        />
        {{ t('marketplace.management.createMemberAction') }}
      </button>
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
              <th>{{ t('marketplace.management.columnContact') }}</th>
              <th>{{ t('marketplace.management.columnMemberType') }}</th>
              <th>{{ t('marketplace.management.columnRoles') }}</th>
              <th>{{ t('common.label.status') }}</th>
              <th>{{ t('marketplace.management.columnCommunity') }}</th>
              <th>{{ t('marketplace.management.columnUpdatedAt') }}</th>
              <th>{{ t('marketplace.management.columnActions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="member in items"
              :key="member.public_id"
            >
              <td>
                <strong>{{ formatMemberName(member) }}</strong>
              </td>
              <td>
                <div class="management-contact-cell">
                  <span>{{ formatPhone(member) }}</span>
                  <small>{{ formatEmail(member) }}</small>
                </div>
              </td>
              <td>{{ member.member_type || '-' }}</td>
              <td>{{ formatRoles(member) }}</td>
              <td>
                <span class="management-status-pill">{{ member.member_status || '-' }}</span>
              </td>
              <td>
                {{ member.primary_community?.name_zh || member.primary_community?.name_en || '-' }}
              </td>
              <td class="management-table-nowrap">{{ formatDate(member.updated_at) }}</td>
              <td class="management-table-nowrap management-table-actions">
                <button
                  type="button"
                  class="management-list-action"
                  @click="openGrantDialog(member)"
                >
                  {{ t('marketplace.management.walletGrantAction') }}
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
        :page-size="pageSize"
        :t="t"
        :total="total"
        @next="loadMembers(page + 1)"
        @previous="loadMembers(page - 1)"
      />
    </article>

    <Teleport to="body">
      <Transition name="management-dialog">
        <div
          v-if="isCreateDialogOpen"
          class="management-dialog"
          role="dialog"
          aria-modal="true"
          :aria-label="t('marketplace.management.createMemberTitle')"
          @click.self="closeCreateDialog"
        >
          <section class="management-dialog-panel management-dialog-panel--wide">
            <header class="management-dialog-header">
              <div>
                <p>{{ t('marketplace.management.createMemberKicker') }}</p>
                <h2>{{ t('marketplace.management.createMemberTitle') }}</h2>
              </div>
              <button
                type="button"
                class="management-dialog-close"
                :aria-label="t('marketplace.management.closeDialog')"
                @click="closeCreateDialog"
              >
                <AppIcon
                  name="close"
                  :size="18"
                />
              </button>
            </header>

            <div class="management-dialog-body management-dialog-grid">
              <label class="management-dialog-field">
                <span>{{ t('marketplace.management.createEmailField') }}</span>
                <input
                  v-model="createForm.email"
                  type="email"
                  autocomplete="off"
                />
              </label>
              <label class="management-dialog-field">
                <span>{{ t('marketplace.management.createPasswordField') }}</span>
                <input
                  v-model="createForm.password"
                  type="password"
                  autocomplete="new-password"
                  minlength="8"
                />
              </label>
              <label class="management-dialog-field">
                <span>{{ t('marketplace.management.createDisplayNameField') }}</span>
                <input
                  v-model="createForm.displayName"
                  type="text"
                  autocomplete="off"
                />
              </label>
              <div class="management-dialog-inline">
                <label class="management-dialog-field">
                  <span>{{ t('marketplace.management.createPhoneCountryField') }}</span>
                  <input
                    v-model="createForm.phoneCountryCode"
                    type="text"
                    autocomplete="off"
                  />
                </label>
                <label class="management-dialog-field">
                  <span>{{ t('marketplace.management.createPhoneField') }}</span>
                  <input
                    v-model="createForm.phoneNumber"
                    type="tel"
                    autocomplete="off"
                  />
                </label>
              </div>
              <label class="management-dialog-field">
                <span>{{ t('marketplace.management.createMemberTypeField') }}</span>
                <select v-model="createForm.memberType">
                  <option value="user">
                    user
                  </option>
                  <option value="pro_user">
                    pro_user
                  </option>
                </select>
              </label>
              <label class="management-dialog-field">
                <span>{{ t('marketplace.management.createRoleField') }}</span>
                <select v-model="createForm.roleCode">
                  <option
                    v-for="role in staffRoleOptions"
                    :key="role.code"
                    :value="role.code"
                  >
                    {{ role.name }}
                  </option>
                </select>
              </label>
              <p class="management-dialog-hint">
                {{ t('marketplace.management.createMemberHint') }}
              </p>
            </div>

            <footer class="management-dialog-actions">
              <button
                type="button"
                class="management-dialog-button management-dialog-button--secondary"
                @click="closeCreateDialog"
              >
                {{ t('marketplace.management.cancelAction') }}
              </button>
              <button
                type="button"
                class="management-dialog-button management-dialog-button--primary"
                :disabled="!canCreateMember || creating"
                @click="submitCreateMember"
              >
                {{ creating ? t('marketplace.management.saving') : t('marketplace.management.createMemberSubmit') }}
              </button>
            </footer>
          </section>
        </div>
      </Transition>
    </Teleport>

    <Teleport to="body">
      <Transition name="management-dialog">
        <div
          v-if="grantTarget"
          class="management-dialog"
          role="dialog"
          aria-modal="true"
          :aria-label="t('marketplace.management.grantDialogTitle')"
          @click.self="closeGrantDialog"
        >
          <section class="management-dialog-panel">
            <header class="management-dialog-header">
              <div>
                <p>{{ t('common.brand.pointsName') }}</p>
                <h2>{{ t('marketplace.management.grantDialogTitle') }}</h2>
              </div>
              <button
                type="button"
                class="management-dialog-close"
                :aria-label="t('marketplace.management.closeDialog')"
                @click="closeGrantDialog"
              >
                <AppIcon
                  name="close"
                  :size="18"
                />
              </button>
            </header>

            <div class="management-dialog-body">
              <div class="management-renew-listing">
                <span>{{ t('marketplace.management.columnMember') }}</span>
                <strong>{{ formatMemberName(grantTarget) }}</strong>
                <small>{{ formatPhone(grantTarget) }}</small>
                <small>{{ formatEmail(grantTarget) }}</small>
              </div>
              <label class="management-dialog-field">
                <span>{{ t('marketplace.management.walletAmountField') }}</span>
                <input
                  v-model.number="grantForm.amount"
                  type="number"
                  min="1"
                  max="1000000"
                  step="1"
                />
              </label>
              <label class="management-dialog-field">
                <span>{{ t('marketplace.management.walletNoteField') }}</span>
                <input
                  v-model="grantForm.note"
                  type="text"
                  maxlength="500"
                  :placeholder="t('marketplace.management.walletNotePlaceholder')"
                />
              </label>
            </div>

            <footer class="management-dialog-actions">
              <button
                type="button"
                class="management-dialog-button management-dialog-button--secondary"
                @click="closeGrantDialog"
              >
                {{ t('marketplace.management.cancelAction') }}
              </button>
              <button
                type="button"
                class="management-dialog-button management-dialog-button--primary"
                :disabled="!canGrantPoints || granting"
                @click="submitGrantPoints"
              >
                {{ granting ? t('marketplace.management.walletGranting') : t('marketplace.management.walletGrantAction') }}
              </button>
            </footer>
          </section>
        </div>
      </Transition>
    </Teleport>
  </section>
</template>
