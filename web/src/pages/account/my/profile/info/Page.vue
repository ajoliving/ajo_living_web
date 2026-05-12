<!--
 * 會員中心個人資料頁。
 * 1. 以只讀面板展示與二手交易一致的會員資料。
 * 2. 集中顯示帳戶、身份狀態、角色與權限資訊。
 * 3. 透過編輯彈窗修改個人資料與上傳頭像。
-->
<script setup lang="ts">
import AppIcon from '@/shared/components/base/AppIcon.vue';
import BaseAvatar from '@/shared/components/base/BaseAvatar.vue';
import BaseButton from '@/shared/components/base/BaseButton.vue';
import BaseInput from '@/shared/components/base/BaseInput.vue';

import { useAccountProfilePage } from '../profile';

const {
  accountRows,
  closeEditModal,
  communityDisplay,
  formState,
  handleAvatarFileChange,
  handleSaveProfile,
  handleSignOut,
  isEditModalOpen,
  isLoading,
  isSaving,
  isSigningOut,
  isUploadingAvatar,
  openEditModal,
  permissionChips,
  profileRows,
  roleChips,
  sessionStore,
  t,
} = useAccountProfilePage();
</script>

<template>
  <div class="account-profile-page">
    <div class="account-profile-header">
      <div class="account-profile-header__identity">
        <BaseAvatar
          :src="sessionStore.currentUser.avatar_url"
          :name="sessionStore.currentUser.display_name"
          :size="56"
        />
        <div class="account-profile-header__text">
          <p>{{ sessionStore.currentUser.display_name }}</p>
          <span>{{ communityDisplay }}</span>
        </div>
        <span
          class="account-profile-status"
          :class="sessionStore.me?.profile_completed ? 'account-profile-status--ready' : ''"
        >
          {{ sessionStore.me?.profile_completed ? t('marketplace.myProfile.completed') : t('marketplace.myProfile.incomplete') }}
        </span>
      </div>

      <div class="account-profile-header__actions">
        <BaseButton
          variant="secondary"
          size="md"
          :disabled="isSigningOut"
          @click="handleSignOut"
        >
          {{ isSigningOut ? t('common.status.loading') : t('account.actions.signOut') }}
        </BaseButton>
        <BaseButton
          variant="primary"
          size="md"
          :disabled="isLoading"
          @click="openEditModal"
        >
          {{ t('marketplace.myProfile.editAction') }}
        </BaseButton>
      </div>
    </div>

    <div
      v-if="isLoading"
      class="account-profile-loading"
    >
      {{ t('common.status.loading') }}
    </div>

    <article class="account-profile-panel">
      <div class="account-profile-grid">
        <section class="account-profile-section">
          <h2>{{ t('marketplace.myProfile.accountInfo') }}</h2>
          <dl class="account-profile-list">
            <div
              v-for="row in profileRows"
              :key="row.key"
              class="account-profile-row"
            >
              <dt>{{ row.label }}</dt>
              <dd>{{ row.value }}</dd>
            </div>
          </dl>
        </section>

        <section class="account-profile-section">
          <h2>{{ t('marketplace.myProfile.identityInfo') }}</h2>
          <dl class="account-profile-list">
            <div
              v-for="row in accountRows"
              :key="row.key"
              class="account-profile-row"
            >
              <dt>{{ row.label }}</dt>
              <dd>{{ row.value }}</dd>
            </div>
          </dl>
        </section>
      </div>

      <section class="account-profile-permissions">
        <div class="account-profile-permissions__title">
          <span>
            <AppIcon
              name="shield"
              :size="18"
            />
          </span>
          <h2>{{ t('marketplace.myProfile.permissions') }}</h2>
        </div>

        <div class="account-profile-chip-group">
          <span
            v-for="role in roleChips"
            :key="role"
            class="account-profile-chip account-profile-chip--role"
          >
            {{ role }}
          </span>
          <span
            v-for="permission in permissionChips"
            :key="permission"
            class="account-profile-chip"
          >
            {{ permission }}
          </span>
        </div>
      </section>
    </article>

    <div
      v-if="isEditModalOpen"
      class="account-profile-modal"
      role="presentation"
      @click.self="closeEditModal"
    >
      <form
        class="account-profile-dialog"
        role="dialog"
        aria-modal="true"
        :aria-label="t('marketplace.myProfile.editTitle')"
        @submit.prevent="handleSaveProfile"
      >
        <header class="account-profile-dialog__header">
          <div>
            <p class="account-profile-kicker">
              {{ t('marketplace.myProfile.editKicker') }}
            </p>
            <h2>{{ t('marketplace.myProfile.editTitle') }}</h2>
          </div>
          <button
            type="button"
            class="account-profile-dialog__close"
            :aria-label="t('marketplace.myProfile.closeEdit')"
            @click="closeEditModal"
          >
            <AppIcon
              name="close"
              :size="18"
            />
          </button>
        </header>

        <div class="account-avatar-editor">
          <BaseAvatar
            :src="sessionStore.currentUser.avatar_url"
            :name="sessionStore.currentUser.display_name"
            :size="56"
          />
          <div class="account-avatar-editor__copy">
            <span>{{ t('account.profile.avatar') }}</span>
            <p>{{ t('account.profile.avatarHint') }}</p>
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

        <div class="account-profile-dialog__fields">
          <label class="account-profile-field">
            <span>{{ t('account.profile.displayName') }}</span>
            <BaseInput
              :model-value="formState.display_name"
              autocomplete="name"
              @update:model-value="formState.display_name = $event"
            />
          </label>

          <div class="account-profile-phone-grid">
            <label class="account-profile-field">
              <span>{{ t('account.profile.phoneCountryCode') }}</span>
              <BaseInput
                :model-value="formState.phone_country_code"
                type="tel"
                inputmode="tel"
                autocomplete="tel-country-code"
                @update:model-value="formState.phone_country_code = $event"
              />
            </label>

            <label class="account-profile-field">
              <span>{{ t('account.profile.phoneNumber') }}</span>
              <BaseInput
                :model-value="formState.phone_number"
                type="tel"
                inputmode="tel"
                autocomplete="tel-national"
                @update:model-value="formState.phone_number = $event"
              />
            </label>
          </div>

          <label class="account-profile-field">
            <span>{{ t('account.profile.publisherIdentity') }}</span>
            <BaseInput
              :model-value="formState.publisher_identity_type"
              @update:model-value="formState.publisher_identity_type = $event"
            />
          </label>

          <label class="account-profile-field">
            <span>{{ t('account.profile.districtCode') }}</span>
            <BaseInput
              :model-value="formState.district_code"
              @update:model-value="formState.district_code = $event"
            />
          </label>
        </div>

        <div class="account-profile-dialog__readonly">
          <div>
            <span>{{ t('account.profile.email') }}</span>
            <strong>{{ formState.email || t('account.profile.emailUnavailable') }}</strong>
          </div>
        </div>

        <footer class="account-profile-dialog__actions">
          <BaseButton
            variant="secondary"
            size="md"
            type="button"
            :disabled="isSaving || isUploadingAvatar"
            @click="closeEditModal"
          >
            {{ t('marketplace.myProfile.cancelAction') }}
          </BaseButton>
          <BaseButton
            variant="primary"
            size="md"
            type="submit"
            :disabled="isSaving || isUploadingAvatar"
          >
            {{ isSaving ? t('common.status.loading') : t('account.actions.save') }}
          </BaseButton>
        </footer>
      </form>
    </div>
  </div>
</template>

<style scoped>
.account-profile-page {
  display: grid;
  min-height: 100%;
  gap: 1.25rem;
  color: rgb(var(--color-text));
}

.account-profile-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid rgb(var(--color-border));
  padding-bottom: 1rem;
}

.account-profile-header__identity {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  gap: 0.9rem;
  align-items: center;
  min-width: 0;
}

.account-profile-header__text {
  min-width: 0;
}

.account-profile-header__text p {
  margin: 0;
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 1.25rem;
  font-weight: 500;
  line-height: 1.3;
  overflow-wrap: anywhere;
}

.account-profile-header__text span {
  display: block;
  margin-top: 0.2rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.875rem;
  font-weight: 700;
  line-height: 1.45;
  overflow-wrap: anywhere;
}

.account-profile-header__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.account-profile-loading {
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: rgb(var(--color-surface));
  padding: 0.9rem 1rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.875rem;
  font-weight: 700;
}

.account-profile-panel {
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: rgb(var(--color-surface));
  box-shadow: 0 10px 30px -5px rgb(0 39 39 / 0.05);
}

.account-profile-status {
  display: inline-flex;
  min-height: 2rem;
  align-items: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 9999px;
  background: rgb(var(--color-surface-muted));
  padding: 0.45rem 0.8rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 800;
  line-height: 1;
}

.account-profile-status--ready {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.account-profile-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
}

.account-profile-section {
  min-width: 0;
  padding: 1.25rem;
}

.account-profile-section + .account-profile-section {
  border-left: 1px solid rgb(var(--color-border));
}

.account-profile-section h2,
.account-profile-permissions h2 {
  margin: 0;
  color: rgb(var(--color-primary));
  font-size: 0.8125rem;
  font-weight: 800;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.account-profile-list {
  display: grid;
  margin: 0.9rem 0 0;
}

.account-profile-row {
  display: grid;
  grid-template-columns: minmax(7.5rem, 0.42fr) minmax(0, 1fr);
  gap: 0.9rem;
  align-items: start;
  border-top: 1px solid rgb(var(--color-border) / 0.72);
  padding: 0.72rem 0;
}

.account-profile-row dt {
  color: rgb(var(--color-text-muted));
  font-size: 0.6875rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  line-height: 1.35;
  text-transform: uppercase;
}

.account-profile-row dd {
  min-width: 0;
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 0.875rem;
  font-weight: 700;
  line-height: 1.45;
  overflow-wrap: anywhere;
}

.account-profile-permissions {
  display: grid;
  gap: 0.9rem;
  border-top: 1px solid rgb(var(--color-border));
  padding: 1.25rem;
}

.account-profile-permissions__title {
  display: inline-flex;
  align-items: center;
  gap: 0.55rem;
}

.account-profile-permissions__title span {
  display: inline-flex;
  height: 2rem;
  width: 2rem;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: rgb(0 39 39 / 0.07);
  color: rgb(var(--color-primary));
}

.account-profile-chip-group {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.account-profile-chip {
  display: inline-flex;
  min-height: 2rem;
  align-items: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 9999px;
  background: rgb(var(--color-surface));
  padding: 0.45rem 0.8rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 800;
  line-height: 1;
}

.account-profile-chip--role {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.account-profile-modal {
  position: fixed;
  inset: 0;
  z-index: 60;
  display: grid;
  place-items: center;
  background: rgb(26 28 27 / 0.38);
  padding: 1.25rem;
}

.account-profile-dialog {
  width: min(100%, 36rem);
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: rgb(var(--color-surface));
  box-shadow: 0 24px 70px rgb(0 0 0 / 0.2);
}

.account-profile-dialog__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid rgb(var(--color-border));
  padding: 1.25rem;
}

.account-profile-kicker {
  margin: 0 0 0.45rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.account-profile-dialog__header h2 {
  margin: 0;
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 1.35rem;
  font-weight: 500;
  line-height: 1.25;
}

.account-profile-dialog__close {
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

.account-profile-dialog__close:hover {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-surface-muted));
}

.account-avatar-editor {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  gap: 1rem;
  align-items: center;
  border-bottom: 1px solid rgb(var(--color-border));
  padding: 1.25rem;
}

.account-avatar-editor__copy {
  min-width: 0;
}

.account-avatar-editor__copy span,
.account-profile-field span,
.account-profile-dialog__readonly span {
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.account-avatar-editor__copy p {
  margin: 0.45rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.8125rem;
  line-height: 1.55;
}

.account-avatar-upload-button {
  display: inline-flex;
  min-height: 2.375rem;
  flex-shrink: 0;
  cursor: pointer;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(var(--color-primary));
  border-radius: 9999px;
  background: rgb(var(--color-primary));
  padding: 0.625rem 1rem;
  color: rgb(var(--color-primary-contrast));
  font-size: 0.8125rem;
  font-weight: 800;
  line-height: 1;
  transition:
    background-color 0.2s ease,
    transform 0.2s ease;
}

.account-avatar-upload-button:hover {
  background: rgb(var(--color-text));
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

.account-profile-dialog__fields {
  display: grid;
  gap: 1rem;
  padding: 1.25rem;
}

.account-profile-phone-grid {
  display: grid;
  grid-template-columns: minmax(5.5rem, 0.45fr) minmax(0, 1fr);
  gap: 0.75rem;
}

.account-profile-field {
  display: grid;
  gap: 0.5rem;
  min-width: 0;
}

.account-profile-dialog__readonly {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  border-top: 1px solid rgb(var(--color-border));
  border-bottom: 1px solid rgb(var(--color-border));
}

.account-profile-dialog__readonly div {
  display: grid;
  gap: 0.45rem;
  min-width: 0;
  padding: 1rem 1.25rem;
}

.account-profile-dialog__readonly div + div {
  border-left: 1px solid rgb(var(--color-border));
}

.account-profile-dialog__readonly strong {
  min-width: 0;
  color: rgb(var(--color-text));
  font-size: 0.875rem;
  font-weight: 700;
  line-height: 1.45;
  overflow-wrap: anywhere;
}

.account-profile-dialog__actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1.25rem;
}

@media (max-width: 767px) {
  .account-profile-header {
    align-items: stretch;
    flex-direction: column;
  }

  .account-profile-header__actions,
  .account-profile-dialog__actions {
    flex-direction: column;
    align-items: stretch;
    width: 100%;
  }

  .account-profile-header__identity,
  .account-profile-grid,
  .account-profile-row,
  .account-avatar-editor,
  .account-profile-phone-grid,
  .account-profile-dialog__readonly {
    grid-template-columns: minmax(0, 1fr);
  }

  .account-profile-section + .account-profile-section,
  .account-profile-dialog__readonly div + div {
    border-left: 0;
    border-top: 1px solid rgb(var(--color-border));
  }

  .account-profile-status {
    justify-self: start;
  }

  .account-avatar-upload-button {
    width: 100%;
  }
}
</style>
