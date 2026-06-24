<!--
 * 二手交易個人資料頁。
 * 1. 直接使用帳戶會員資料，保持二手交易與會員中心一致。
 * 2. 預設只讀展示身份、聯絡、屋苑與權限資訊。
 * 3. 透過彈窗編輯可同步回寫 `/me/profile` 的個人資料欄位。
-->
<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import { updateMe } from '@/httpapis/me';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import BaseAvatar from '@/shared/components/base/BaseAvatar.vue';
import BaseButton from '@/shared/components/base/BaseButton.vue';
import BaseInput from '@/shared/components/base/BaseInput.vue';
import { useFeedbackStore } from '@/stores/feedback';
import { useSessionStore } from '@/stores/session';

interface ProfileInfoRow {
  key: string;
  label: string;
  value: string;
}

const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const sessionStore = useSessionStore();
const isLoading = ref(false);
const isSaving = ref(false);
const isEditModalOpen = ref(false);

const formState = reactive({
  display_name: '',
  publisher_identity_type: '',
  district_code: '',
});

// 1. 輸出缺省顯示文案
const fallbackValue = computed(() => t('marketplace.myProfile.emptyValue'));

// 2. 組合會員電話顯示
const phoneDisplay = computed(() => {
  if (sessionStore.me?.phone_country_code === 'email') {
    return t('account.profile.phoneUnavailable');
  }

  return `${sessionStore.me?.phone_country_code ?? '+852'} ${sessionStore.me?.phone_number ?? ''}`.trim() || fallbackValue.value;
});

// 3. 組合主要屋苑顯示
const communityDisplay = computed(() => {
  const community = sessionStore.me?.primary_community;

  return (
    community?.name_zh?.trim() ||
    community?.name_en?.trim() ||
    community?.address_text?.trim() ||
    t('marketplace.myProfile.noCommunity')
  );
});

// 4. 組合基本會員資料行
const profileRows = computed<ProfileInfoRow[]>(() => [
  {
    key: 'display_name',
    label: t('account.profile.displayName'),
    value: sessionStore.me?.display_name?.trim() || sessionStore.currentUser.display_name || fallbackValue.value,
  },
  {
    key: 'email',
    label: t('account.profile.email'),
    value: sessionStore.me?.email?.trim() || t('account.profile.emailUnavailable'),
  },
  {
    key: 'phone',
    label: t('account.profile.phone'),
    value: phoneDisplay.value,
  },
  {
    key: 'publisher_identity_type',
    label: t('account.profile.publisherIdentity'),
    value: sessionStore.me?.publisher_identity_type?.trim() || fallbackValue.value,
  },
  {
    key: 'district_code',
    label: t('account.profile.districtCode'),
    value: sessionStore.me?.district_code?.trim() || fallbackValue.value,
  },
  {
    key: 'primary_community',
    label: t('common.label.community'),
    value: communityDisplay.value,
  },
]);

// 5. 組合身份與狀態資料行
const accountRows = computed<ProfileInfoRow[]>(() => [
  {
    key: 'public_id',
    label: t('marketplace.myProfile.memberId'),
    value: sessionStore.me?.public_id?.trim() || fallbackValue.value,
  },
  {
    key: 'member_status',
    label: t('marketplace.myProfile.memberStatus'),
    value: sessionStore.me?.member_status?.trim() || fallbackValue.value,
  },
  {
    key: 'member_type',
    label: t('marketplace.myProfile.memberType'),
    value: sessionStore.me?.member_type?.trim() || fallbackValue.value,
  },
  {
    key: 'role',
    label: t('marketplace.myProfile.primaryRole'),
    value: sessionStore.me?.role?.trim() || fallbackValue.value,
  },
  {
    key: 'profile_completed',
    label: t('marketplace.myProfile.profileCompleted'),
    value: sessionStore.me?.profile_completed ? t('marketplace.myProfile.completed') : t('marketplace.myProfile.incomplete'),
  },
  {
    key: 'is_staff',
    label: t('marketplace.myProfile.staffAccess'),
    value: sessionStore.me?.is_staff ? t('marketplace.myProfile.enabled') : t('marketplace.myProfile.disabled'),
  },
]);

// 6. 組合角色與權限
const roleChips = computed(() => (sessionStore.me?.roles?.length ? sessionStore.me.roles : [t('marketplace.myProfile.noRoles')]));
const permissionChips = computed(() =>
  sessionStore.me?.permissions?.length ? sessionStore.me.permissions : [t('marketplace.myProfile.noPermissions')],
);

// 7. 同步編輯表單
const syncFormState = (): void => {
  formState.display_name = sessionStore.me?.display_name ?? sessionStore.currentUser.display_name;
  formState.publisher_identity_type = sessionStore.me?.publisher_identity_type ?? '';
  formState.district_code = sessionStore.me?.district_code ?? '';
};

// 8. 載入會員資料
const loadProfile = async (): Promise<void> => {
  if (sessionStore.me) {
    syncFormState();
    return;
  }

  isLoading.value = true;

  try {
    await sessionStore.loadCurrentUser();
    syncFormState();
  } catch (error) {
    console.error(error);
    feedbackStore.pushToast(t('marketplace.myProfile.loadError'), 'error');
  } finally {
    isLoading.value = false;
  }
};

// 9. 開啟編輯彈窗
const openEditModal = (): void => {
  syncFormState();
  isEditModalOpen.value = true;
};

// 10. 關閉編輯彈窗
const closeEditModal = (): void => {
  if (!isSaving.value) {
    isEditModalOpen.value = false;
  }
};

// 11. 儲存會員資料
const handleSaveProfile = async (): Promise<void> => {
  isSaving.value = true;

  try {
    const { data } = await updateMe({
      display_name: formState.display_name.trim(),
      publisher_identity_type: formState.publisher_identity_type.trim(),
      district_code: formState.district_code.trim(),
    });

    sessionStore.me = data.data;
    syncFormState();
    isEditModalOpen.value = false;
    feedbackStore.pushToast(t('account.profile.updateSuccess'), 'success');
  } catch (error) {
    console.error(error);
    feedbackStore.pushToast(t('account.profile.updateError'), 'error');
  } finally {
    isSaving.value = false;
  }
};

// 12. 初始化頁面資料
onMounted(() => {
  void loadProfile();
});
</script>

<template>
  <section class="marketplace-profile-page">
    <div class="marketplace-profile-heading">
      <div>
        <p class="marketplace-profile-kicker">
          {{ t('marketplace.myProfile.kicker') }}
        </p>
        <h2>{{ t('marketplace.myHub.profileTitle') }}</h2>
      </div>
      <BaseButton
        variant="primary"
        size="md"
        :disabled="isLoading"
        @click="openEditModal"
      >
        {{ t('marketplace.myProfile.editAction') }}
      </BaseButton>
    </div>

    <div
      v-if="isLoading"
      class="marketplace-profile-loading"
    >
      {{ t('common.status.loading') }}
    </div>

    <article class="marketplace-profile-panel">
      <header class="marketplace-profile-summary">
        <BaseAvatar
          :src="sessionStore.currentUser.avatar_url"
          :name="sessionStore.currentUser.display_name"
          :size="64"
        />
        <div class="marketplace-profile-summary__text">
          <p>{{ sessionStore.currentUser.display_name }}</p>
          <span>{{ communityDisplay }}</span>
        </div>
        <span
          class="marketplace-profile-status"
          :class="sessionStore.me?.profile_completed ? 'marketplace-profile-status--ready' : ''"
        >
          {{ sessionStore.me?.profile_completed ? t('marketplace.myProfile.completed') : t('marketplace.myProfile.incomplete') }}
        </span>
      </header>

      <section class="marketplace-profile-metrics">
        <article class="marketplace-profile-metric">
          <strong>{{ roleChips.length }}</strong>
          <span>{{ t('marketplace.myProfile.permissions') }}</span>
        </article>
        <article class="marketplace-profile-metric">
          <strong>{{ permissionChips.length }}</strong>
          <span>Access</span>
        </article>
        <article class="marketplace-profile-metric">
          <strong>{{ sessionStore.me?.profile_completed ? '100%' : '60%' }}</strong>
          <span>{{ t('marketplace.myProfile.profileCompleted') }}</span>
        </article>
      </section>

      <div class="marketplace-profile-grid">
        <section class="marketplace-profile-section">
          <h3>{{ t('marketplace.myProfile.accountInfo') }}</h3>
          <dl class="marketplace-profile-list">
            <div
              v-for="row in profileRows"
              :key="row.key"
              class="marketplace-profile-row"
            >
              <dt>{{ row.label }}</dt>
              <dd>{{ row.value }}</dd>
            </div>
          </dl>
        </section>

        <section class="marketplace-profile-section">
          <h3>{{ t('marketplace.myProfile.identityInfo') }}</h3>
          <dl class="marketplace-profile-list">
            <div
              v-for="row in accountRows"
              :key="row.key"
              class="marketplace-profile-row"
            >
              <dt>{{ row.label }}</dt>
              <dd>{{ row.value }}</dd>
            </div>
          </dl>
        </section>
      </div>

      <section class="marketplace-profile-permissions">
        <div class="marketplace-profile-permissions__title">
          <span>
            <AppIcon
              name="shield"
              :size="18"
            />
          </span>
          <h3>{{ t('marketplace.myProfile.permissions') }}</h3>
        </div>

        <div class="marketplace-profile-chip-group">
          <span
            v-for="role in roleChips"
            :key="role"
            class="marketplace-profile-chip marketplace-profile-chip--role"
          >
            {{ role }}
          </span>
          <span
            v-for="permission in permissionChips"
            :key="permission"
            class="marketplace-profile-chip"
          >
            {{ permission }}
          </span>
        </div>
      </section>
    </article>

    <div
      v-if="isEditModalOpen"
      class="marketplace-profile-modal"
      role="presentation"
      @click.self="closeEditModal"
    >
      <form
        class="marketplace-profile-dialog"
        role="dialog"
        aria-modal="true"
        :aria-label="t('marketplace.myProfile.editTitle')"
        @submit.prevent="handleSaveProfile"
      >
        <header class="marketplace-profile-dialog__header">
          <div>
            <p class="marketplace-profile-kicker">
              {{ t('marketplace.myProfile.editKicker') }}
            </p>
            <h3>{{ t('marketplace.myProfile.editTitle') }}</h3>
          </div>
          <button
            type="button"
            class="marketplace-profile-dialog__close"
            :aria-label="t('marketplace.myProfile.closeEdit')"
            @click="closeEditModal"
          >
            <AppIcon
              name="close"
              :size="18"
            />
          </button>
        </header>

        <div class="marketplace-profile-dialog__fields">
          <label class="marketplace-profile-field">
            <span>{{ t('account.profile.displayName') }}</span>
            <BaseInput
              :model-value="formState.display_name"
              autocomplete="name"
              @update:model-value="formState.display_name = $event"
            />
          </label>

          <label class="marketplace-profile-field">
            <span>{{ t('account.profile.publisherIdentity') }}</span>
            <BaseInput
              :model-value="formState.publisher_identity_type"
              @update:model-value="formState.publisher_identity_type = $event"
            />
          </label>

          <label class="marketplace-profile-field">
            <span>{{ t('account.profile.districtCode') }}</span>
            <BaseInput
              :model-value="formState.district_code"
              @update:model-value="formState.district_code = $event"
            />
          </label>
        </div>

        <div class="marketplace-profile-dialog__readonly">
          <div>
            <span>{{ t('account.profile.email') }}</span>
            <strong>{{ sessionStore.me?.email || t('account.profile.emailUnavailable') }}</strong>
          </div>
          <div>
            <span>{{ t('account.profile.phone') }}</span>
            <strong>{{ phoneDisplay }}</strong>
          </div>
        </div>

        <footer class="marketplace-profile-dialog__actions">
          <BaseButton
            variant="secondary"
            size="md"
            type="button"
            :disabled="isSaving"
            @click="closeEditModal"
          >
            {{ t('marketplace.myProfile.cancelAction') }}
          </BaseButton>
          <BaseButton
            variant="primary"
            size="md"
            type="submit"
            :disabled="isSaving"
          >
            {{ isSaving ? t('common.status.loading') : t('account.actions.save') }}
          </BaseButton>
        </footer>
      </form>
    </div>
  </section>
</template>

<style scoped>
.marketplace-profile-page {
  display: grid;
  gap: 16px;
  color: rgb(var(--color-text));
}

.marketplace-profile-heading {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 16px;
  border-bottom: 1px solid rgb(var(--color-border));
  padding-bottom: 16px;
}

.marketplace-profile-kicker {
  margin: 0;
  color: rgb(var(--color-primary));
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.14em;
  line-height: 1;
  text-transform: uppercase;
}

.marketplace-profile-heading h2,
.marketplace-profile-dialog__header h3 {
  margin: 0;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-weight: 400;
  line-height: 1.25;
}

.marketplace-profile-heading h2 {
  margin-top: 6px;
  font-size: clamp(1.8rem, 3vw, 2.35rem);
}

.marketplace-profile-dialog__header h3 {
  font-size: 24px;
}

.marketplace-profile-loading {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 14px 16px;
  color: rgb(var(--color-text-muted));
  font-size: 13px;
  font-weight: 700;
}

.marketplace-profile-panel {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
}

.marketplace-profile-summary {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  gap: 14px;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid rgb(var(--color-border));
}

.marketplace-profile-summary__text {
  min-width: 0;
}

.marketplace-profile-summary__text p {
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 20px;
  font-weight: 700;
  line-height: 1.3;
  overflow-wrap: anywhere;
}

.marketplace-profile-summary__text span {
  display: block;
  margin-top: 4px;
  color: rgb(var(--color-text-muted));
  font-size: 13px;
  font-weight: 700;
  line-height: 1.45;
  overflow-wrap: anywhere;
}

.marketplace-profile-status {
  display: inline-flex;
  min-height: 30px;
  align-items: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 9999px;
  background: rgb(var(--color-surface-muted));
  padding: 0 12px;
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  font-weight: 700;
  line-height: 1;
}

.marketplace-profile-status--ready {
  border-color: rgb(var(--color-success) / 0.2);
  background: rgb(var(--color-success) / 0.12);
  color: rgb(var(--color-success));
}

.marketplace-profile-metrics {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  padding: 16px;
  border-bottom: 1px solid rgb(var(--color-border));
}

.marketplace-profile-metric {
  display: grid;
  gap: 8px;
  min-height: 112px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 16px;
}

.marketplace-profile-metric strong {
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 30px;
  font-weight: 400;
  line-height: 1;
}

.marketplace-profile-metric span {
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  font-weight: 700;
}

.marketplace-profile-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 0;
}

.marketplace-profile-section {
  min-width: 0;
  padding: 16px;
}

.marketplace-profile-section + .marketplace-profile-section {
  border-left: 1px solid rgb(var(--color-border));
}

.marketplace-profile-section h3,
.marketplace-profile-permissions h3 {
  margin: 0;
  color: rgb(var(--color-primary));
  font-size: 14px;
  font-weight: 700;
  letter-spacing: 0.1em;
  line-height: 1;
}

.marketplace-profile-list {
  display: grid;
  gap: 0;
  margin: 14px 0 0;
}

.marketplace-profile-row {
  display: grid;
  grid-template-columns: minmax(7.5rem, 0.42fr) minmax(0, 1fr);
  gap: 12px;
  align-items: start;
  padding: 12px 0;
  border-top: 1px solid rgb(var(--color-border) / 0.72);
}

.marketplace-profile-row dt {
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  font-weight: 700;
  line-height: 1.35;
}

.marketplace-profile-row dd {
  min-width: 0;
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 13px;
  font-weight: 700;
  line-height: 1.45;
  overflow-wrap: anywhere;
}

.marketplace-profile-permissions {
  display: grid;
  gap: 14px;
  border-top: 1px solid rgb(var(--color-border));
  padding: 16px;
}

.marketplace-profile-permissions__title {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.marketplace-profile-permissions__title span {
  display: inline-flex;
  height: 32px;
  width: 32px;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: rgb(var(--color-primary) / 0.1);
  color: rgb(var(--color-primary));
}

.marketplace-profile-chip-group {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.marketplace-profile-chip {
  display: inline-flex;
  min-height: 32px;
  align-items: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 9999px;
  background: rgb(var(--color-surface));
  padding: 0 12px;
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  font-weight: 700;
  line-height: 1;
}

.marketplace-profile-chip--role {
  border-color: rgb(var(--color-primary) / 0.2);
  background: rgb(var(--color-primary) / 0.1);
  color: rgb(var(--color-primary));
}

.marketplace-profile-modal {
  position: fixed;
  inset: 0;
  z-index: 60;
  display: grid;
  place-items: center;
  background: rgb(26 28 27 / 0.38);
  padding: 1.25rem;
}

.marketplace-profile-dialog {
  width: min(100%, 34rem);
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  box-shadow: 0 24px 70px rgb(0 0 0 / 0.2);
}

.marketplace-profile-dialog__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  border-bottom: 1px solid rgb(var(--color-border));
  padding: 16px;
}

.marketplace-profile-dialog__close {
  display: inline-flex;
  height: 2.25rem;
  width: 2.25rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 9999px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-primary));
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease;
}

.marketplace-profile-dialog__close:hover {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-surface-muted));
}

.marketplace-profile-dialog__fields {
  display: grid;
  gap: 16px;
  padding: 16px;
}

.marketplace-profile-field {
  display: grid;
  gap: 0.5rem;
  min-width: 0;
}

.marketplace-profile-field span,
.marketplace-profile-dialog__readonly span {
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.marketplace-profile-dialog__readonly {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 0;
  border-top: 1px solid rgb(var(--color-border));
  border-bottom: 1px solid rgb(var(--color-border));
}

.marketplace-profile-dialog__readonly div {
  display: grid;
  gap: 6px;
  min-width: 0;
  padding: 14px 16px;
}

.marketplace-profile-dialog__readonly div + div {
  border-left: 1px solid rgb(var(--color-border));
}

.marketplace-profile-dialog__readonly strong {
  min-width: 0;
  color: rgb(var(--color-text));
  font-size: 13px;
  font-weight: 700;
  line-height: 1.45;
  overflow-wrap: anywhere;
}

.marketplace-profile-dialog__actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 16px;
}

@media (max-width: 767px) {
  .marketplace-profile-heading,
  .marketplace-profile-dialog__actions {
    align-items: stretch;
    flex-direction: column;
  }

  .marketplace-profile-summary,
  .marketplace-profile-metrics,
  .marketplace-profile-grid,
  .marketplace-profile-row,
  .marketplace-profile-dialog__readonly {
    grid-template-columns: minmax(0, 1fr);
  }

  .marketplace-profile-section + .marketplace-profile-section,
  .marketplace-profile-dialog__readonly div + div {
    border-left: 0;
    border-top: 1px solid rgb(var(--color-border));
  }

  .marketplace-profile-status {
    justify-self: start;
  }
}
</style>
