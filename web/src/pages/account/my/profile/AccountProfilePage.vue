<!--
 * 會員中心個人資料頁。
 * 1. 讀取並編輯真實會員資料。
 * 2. 串接 `/me/profile` 儲存並同步會員狀態。
-->
<script setup lang="ts">
import BaseAvatar from '@/shared/components/base/BaseAvatar.vue';
import BaseButton from '@/shared/components/base/BaseButton.vue';
import BaseInput from '@/shared/components/base/BaseInput.vue';

import { useAccountProfilePage } from './profile';
import AccountSectionHeader from './widgets/AccountSectionHeader.vue';

const {
  formState,
  handleAvatarFileChange,
  handleSaveProfile,
  handleSignOut,
  isLoading,
  isSaving,
  isSigningOut,
  isUploadingAvatar,
  phoneDisplay,
  sessionStore,
  t,
} = useAccountProfilePage();
</script>

<template>
  <div class="account-profile-page">
    <div class="account-profile-header">
      <AccountSectionHeader
        :eyebrow="t('account.sections.profile')"
        :title="t('account.profile.title')"
        :description="t('account.profile.description')"
      />
    </div>

    <div class="account-profile-card">
      <div class="account-profile-identity">
        <div class="account-avatar-uploader">
          <BaseAvatar
            :src="sessionStore.currentUser.avatar_url"
            :name="sessionStore.currentUser.display_name"
            :size="64"
          />
          <div class="min-w-0 flex-1">
            <p class="account-profile-label">
              {{ t('account.profile.avatar') }}
            </p>
            <p class="mt-1.5 text-[13px] leading-[1.55] text-[#414848]">
              {{ t('account.profile.avatarHint') }}
            </p>
          </div>
          <label class="account-avatar-upload-button">
            <input
              type="file"
              accept="image/*"
              class="sr-only"
              :disabled="isUploadingAvatar || isLoading"
              @change="handleAvatarFileChange"
            />
            <span>
              {{ isUploadingAvatar ? t('common.status.loading') : t('account.profile.uploadAvatar') }}
            </span>
          </label>
        </div>

        <div class="account-profile-info-stack">
          <div class="account-profile-info-row">
            <span class="account-profile-label">{{ t('account.profile.email') }}</span>
            <span class="account-profile-info-value">
              {{ formState.email || t('account.profile.emailUnavailable') }}
            </span>
          </div>
          <div class="account-profile-info-row">
            <span class="account-profile-label">{{ t('account.profile.phone') }}</span>
            <span class="account-profile-info-value">
              {{ phoneDisplay }}
            </span>
          </div>
          <div class="account-profile-info-row">
            <span class="account-profile-label">{{ t('account.profile.avatarUrl') }}</span>
            <span class="account-profile-info-value">
              {{ sessionStore.currentUser.avatar_url || t('account.profile.avatarUrlUnavailable') }}
            </span>
          </div>
        </div>
      </div>

      <p
        v-if="isLoading"
        class="mb-5 mt-6 text-sm text-[#717878]"
      >
        {{ t('common.status.loading') }}
      </p>

      <div class="account-profile-form-stack">
        <label class="account-profile-field">
          <span class="account-profile-label">{{ t('account.profile.displayName') }}</span>
          <BaseInput
            :model-value="formState.display_name"
            @update:model-value="formState.display_name = $event"
          />
        </label>

        <label class="account-profile-field">
          <span class="account-profile-label">{{ t('account.profile.publisherIdentity') }}</span>
          <BaseInput
            :model-value="formState.publisher_identity_type"
            @update:model-value="formState.publisher_identity_type = $event"
          />
        </label>

        <label class="account-profile-field">
          <span class="account-profile-label">{{ t('account.profile.districtCode') }}</span>
          <BaseInput
            :model-value="formState.district_code"
            @update:model-value="formState.district_code = $event"
          />
        </label>
      </div>

      <div class="account-profile-actions">
        <BaseButton
          variant="danger"
          size="md"
          :disabled="isSigningOut"
          @click="handleSignOut"
        >
          {{ isSigningOut ? t('common.status.loading') : t('account.actions.signOut') }}
        </BaseButton>
        <BaseButton
          variant="primary"
          size="md"
          :disabled="isSaving || isLoading"
          @click="handleSaveProfile"
        >
          {{ isSaving ? t('common.status.loading') : t('account.actions.save') }}
        </BaseButton>
      </div>
    </div>
  </div>
</template>

<style scoped>
.account-profile-page {
  display: flex;
  min-height: 100%;
  flex-direction: column;
  gap: 1.25rem;
}

.account-profile-header {
  border-bottom: 1px solid #e2e3e1;
  padding-bottom: 1rem;
}

.account-profile-card {
  border: 1px solid #e2e3e1;
  border-radius: 0.75rem;
  background: #f9f9f7;
  padding: 1.375rem;
  box-shadow: 0 4px 24px rgb(0 0 0 / 0.03);
}

.account-profile-identity {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  border-bottom: 1px solid #e2e3e1;
  padding-bottom: 1.25rem;
}

.account-avatar-uploader {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.account-avatar-upload-button {
  display: inline-flex;
  min-height: 2.375rem;
  flex-shrink: 0;
  cursor: pointer;
  align-items: center;
  justify-content: center;
  border: 1px solid #002727;
  border-radius: 0.75rem;
  background: #002727;
  padding: 0.625rem 1rem;
  color: #ffffff;
  font-size: 0.8125rem;
  font-weight: 700;
  letter-spacing: 0.02em;
  line-height: 1;
  transition:
    background-color 0.2s ease,
    transform 0.2s ease;
}

.account-avatar-upload-button:hover {
  background: #426464;
  transform: translateY(-1px);
}

.account-avatar-upload-button:focus-within {
  outline: 2px solid rgb(var(--color-primary));
  outline-offset: 3px;
}

.account-avatar-upload-button:has(input:disabled) {
  cursor: not-allowed;
  opacity: 0.62;
}

.account-profile-info-stack {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid #e2e3e1;
  border-radius: 0.625rem;
  background: #ffffff;
}

.account-profile-info-row {
  display: flex;
  min-width: 0;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.75rem 0.875rem;
}

.account-profile-info-row + .account-profile-info-row {
  border-top: 1px solid #edf0ed;
}

.account-profile-info-row .account-profile-label {
  margin-bottom: 0;
}

.account-profile-info-value {
  min-width: 0;
  color: #002727;
  font-size: 0.875rem;
  font-weight: 700;
  line-height: 1.45;
  overflow-wrap: anywhere;
  text-align: right;
}

.account-profile-form-stack {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  margin-top: 1.25rem;
}

.account-profile-field {
  display: block;
  min-width: 0;
}

.account-profile-label {
  display: block;
  margin-bottom: 0.375rem;
  color: #717878;
  font-size: 0.6875rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.account-profile-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  margin-top: 1.25rem;
}

@media (max-width: 767px) {
  .account-profile-card {
    padding: 1rem;
  }

  .account-avatar-uploader {
    align-items: flex-start;
    flex-direction: column;
  }

  .account-avatar-upload-button {
    width: 100%;
  }

  .account-profile-info-row,
  .account-profile-actions {
    flex-direction: column;
    align-items: stretch;
  }

  .account-profile-info-value {
    text-align: left;
  }
}
</style>
