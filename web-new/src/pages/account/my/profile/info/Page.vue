<!--
 * 會員中心個人資料頁。
 * 1. 以卡片式網格布局展示會員資料。
 * 2. 集中顯示帳戶、身份狀態、角色與權限資訊。
 * 3. 透過編輯彈窗修改個人資料、綁定單位與 iSmart 資料。
-->
<script setup lang="ts">
import AppIcon from '@/shared/components/base/AppIcon.vue';
import AppGlassSelect from '@/shared/components/base/AppGlassSelect.vue';
import BaseAvatar from '@/shared/components/base/BaseAvatar.vue';
import BaseButton from '@/shared/components/base/BaseButton.vue';
import BaseInput from '@/shared/components/base/BaseInput.vue';

import { useAccountProfilePage } from '../profile';

const {
  accountRows,
  buildingOptions,
  canSaveUnit,
  closeEditModal,
  closeIsmartModal,
  closeUnitModal,
  floorOptions,
  formState,
  handleAvatarFileChange,
  handleBindIsmart,
  handleSaveUnit,
  handleSaveProfile,
  handleSignOut,
  handleUnitBuildingChange,
  handleUnitFloorChange,
  isBuildingsLoading,
  isBindingIsmart,
  isEditModalOpen,
  isIsmartModalOpen,
  isLoading,
  isSaving,
  isSavingUnit,
  isSigningOut,
  isUnitModalOpen,
  isUnitsLoading,
  isUploadingAvatar,
  ismartFormState,
  openEditModal,
  openIsmartModal,
  openUnitModal,
  permissionChips,
  profileRows,
  roleChips,
  selectedBuildingID,
  selectedFloor,
  selectedUnitDisplay,
  selectedUnitID,
  sessionStore,
  t,
  unitOptions,
} = useAccountProfilePage();
</script>

<template>
  <div class="account-profile-page">
    <section class="account-profile-hero">
      <div class="account-profile-hero__identity">
        <BaseAvatar
          :src="sessionStore.currentUser.avatar_url"
          :name="sessionStore.currentUser.display_name"
          :size="64"
        />
        <div class="account-profile-hero__text">
          <h1>{{ sessionStore.currentUser.display_name }}</h1>
          <p>{{ t('account.profile.title') }}</p>
        </div>
      </div>

      <div class="account-profile-hero__actions">
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
    </section>

    <div
      v-if="isLoading"
      class="account-profile-loading"
    >
      {{ t('common.status.loading') }}
    </div>

    <div class="account-profile-grid">
      <article class="account-profile-card">
        <header class="account-profile-card__header">
          <h2>{{ t('account.profile.accountInfo') }}</h2>
        </header>
        <dl class="account-profile-card__list">
          <div
            v-for="row in profileRows"
            :key="row.key"
            class="account-profile-card__row"
          >
            <dt>{{ row.label }}</dt>
            <dd>{{ row.value }}</dd>
          </div>
        </dl>
      </article>

      <article class="account-profile-card">
        <header class="account-profile-card__header">
          <h2>{{ t('account.profile.identityInfo') }}</h2>
        </header>
        <dl class="account-profile-card__list">
          <div
            v-for="row in accountRows"
            :key="row.key"
            class="account-profile-card__row"
          >
            <dt>{{ row.label }}</dt>
            <dd>{{ row.value }}</dd>
          </div>
        </dl>
      </article>

      <article class="account-profile-card">
        <header class="account-profile-card__header">
          <h2>{{ t('account.profile.permissions') }}</h2>
        </header>
        <div class="account-profile-card__content">
          <div class="account-profile-chip-group">
            <span
              v-for="role in roleChips"
              :key="role"
              class="account-profile-chip account-profile-chip--role"
            >
              {{ role }}
            </span>
          </div>
          <div class="account-profile-chip-group">
            <span
              v-for="permission in permissionChips"
              :key="permission"
              class="account-profile-chip"
            >
              {{ permission }}
            </span>
          </div>
        </div>
      </article>

      <article class="account-profile-card">
        <header class="account-profile-card__header">
          <h2>{{ t('account.profile.bindUnit') }}</h2>
          <BaseButton
            variant="secondary"
            size="sm"
            @click="openUnitModal"
          >
            {{ t('account.center.account.editAction') }}
          </BaseButton>
        </header>
        <div class="account-profile-card__content">
          <p class="account-profile-unit-display">
            {{ selectedUnitDisplay }}
          </p>
        </div>
      </article>

      <article
        v-if="sessionStore.me?.ismart_linked"
        class="account-profile-card account-profile-card--wide"
      >
        <header class="account-profile-card__header">
          <h2>{{ t('account.center.account.ismartData') }}</h2>
          <BaseButton
            variant="secondary"
            size="sm"
            :disabled="isLoading || isBindingIsmart"
            @click="openIsmartModal"
          >
            {{ t('account.profile.updateIsmart') }}
          </BaseButton>
        </header>
        <div
          v-if="sessionStore.me?.ismart_account_profile"
          class="account-profile-ismart-grid"
        >
          <div class="account-profile-ismart-section">
            <h3>{{ t('account.center.account.ismartAccountData') }}</h3>
            <dl class="account-profile-card__list account-profile-card__list--compact">
              <div class="account-profile-card__row">
                <dt>{{ t('account.center.account.accountCode') }}</dt>
                <dd>{{ sessionStore.me.ismart_account_profile.account_code || '-' }}</dd>
              </div>
              <div class="account-profile-card__row">
                <dt>{{ t('account.center.account.accountPhone') }}</dt>
                <dd>{{ sessionStore.me.ismart_account_profile.account_phone || '-' }}</dd>
              </div>
              <div class="account-profile-card__row">
                <dt>{{ t('account.center.account.accountEmail') }}</dt>
                <dd>{{ sessionStore.me.ismart_account_profile.account_email || '-' }}</dd>
              </div>
            </dl>
          </div>

          <div class="account-profile-ismart-section">
            <h3>{{ t('account.center.account.ownerData') }}</h3>
            <dl class="account-profile-card__list account-profile-card__list--compact">
              <div class="account-profile-card__row">
                <dt>{{ t('account.center.account.ownerNameEn') }}</dt>
                <dd>{{ sessionStore.me.ismart_account_profile.owner_name_en || '-' }}</dd>
              </div>
              <div class="account-profile-card__row">
                <dt>{{ t('account.center.account.ownerNameZh') }}</dt>
                <dd>{{ sessionStore.me.ismart_account_profile.owner_name_zh || '-' }}</dd>
              </div>
              <div class="account-profile-card__row">
                <dt>{{ t('account.center.account.gender') }}</dt>
                <dd>{{ sessionStore.me.ismart_account_profile.gender || '-' }}</dd>
              </div>
              <div class="account-profile-card__row">
                <dt>{{ t('account.center.account.birthDate') }}</dt>
                <dd>{{ sessionStore.me.ismart_account_profile.birth_date || '-' }}</dd>
              </div>
            </dl>
          </div>

          <div class="account-profile-ismart-section">
            <h3>{{ t('account.center.account.contactData') }}</h3>
            <dl class="account-profile-card__list account-profile-card__list--compact">
              <div class="account-profile-card__row">
                <dt>{{ t('account.center.account.contactName') }}</dt>
                <dd>{{ sessionStore.me.ismart_account_profile.contact_name || '-' }}</dd>
              </div>
              <div class="account-profile-card__row">
                <dt>{{ t('account.center.account.contactPhone') }}</dt>
                <dd>{{ sessionStore.me.ismart_account_profile.contact_phone || '-' }}</dd>
              </div>
              <div class="account-profile-card__row">
                <dt>{{ t('account.center.account.emergencyContactName') }}</dt>
                <dd>{{ sessionStore.me.ismart_account_profile.emergency_contact_name || '-' }}</dd>
              </div>
              <div class="account-profile-card__row">
                <dt>{{ t('account.center.account.emergencyContactPhone') }}</dt>
                <dd>{{ sessionStore.me.ismart_account_profile.emergency_contact_phone || '-' }}</dd>
              </div>
            </dl>
          </div>

          <div class="account-profile-ismart-section">
            <h3>{{ t('account.center.account.billingData') }}</h3>
            <dl class="account-profile-card__list account-profile-card__list--compact">
              <div class="account-profile-card__row">
                <dt>{{ t('account.center.account.billingPhone') }}</dt>
                <dd>{{ sessionStore.me.ismart_account_profile.billing_phone || '-' }}</dd>
              </div>
              <div class="account-profile-card__row">
                <dt>{{ t('account.center.account.billingEmail') }}</dt>
                <dd>{{ sessionStore.me.ismart_account_profile.billing_email || '-' }}</dd>
              </div>
              <div class="account-profile-card__row">
                <dt>{{ t('account.center.account.billingAddressEn') }}</dt>
                <dd>{{ sessionStore.me.ismart_account_profile.billing_address_en || '-' }}</dd>
              </div>
              <div class="account-profile-card__row">
                <dt>{{ t('account.center.account.billingAddressZh') }}</dt>
                <dd>{{ sessionStore.me.ismart_account_profile.billing_address_zh || '-' }}</dd>
              </div>
            </dl>
          </div>
        </div>
      </article>

      <article
        v-else
        class="account-profile-card"
      >
        <header class="account-profile-card__header">
          <h2>{{ t('account.profile.ismartAccount') }}</h2>
        </header>
        <div class="account-profile-card__content">
          <p class="account-profile-empty-state">
            {{ t('account.profile.ismartUnlinked') }}
          </p>
          <BaseButton
            variant="primary"
            size="md"
            :disabled="isLoading || isBindingIsmart"
            @click="openIsmartModal"
          >
            {{ t('account.profile.bindIsmart') }}
          </BaseButton>
        </div>
      </article>
    </div>

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

    <div
      v-if="isUnitModalOpen"
      class="account-profile-modal"
      role="presentation"
      @click.self="closeUnitModal"
    >
      <form
        class="account-profile-dialog"
        role="dialog"
        aria-modal="true"
        :aria-label="t('account.profile.bindUnit')"
        @submit.prevent="handleSaveUnit"
      >
        <header class="account-profile-dialog__header">
          <div>
            <p class="account-profile-kicker">
              {{ t('account.center.account.propertyManagement') }}
            </p>
            <h2>{{ t('account.profile.bindUnit') }}</h2>
          </div>
          <button
            type="button"
            class="account-profile-dialog__close"
            :aria-label="t('marketplace.myProfile.closeEdit')"
            @click="closeUnitModal"
          >
            <AppIcon
              name="close"
              :size="18"
            />
          </button>
        </header>

        <div class="account-profile-dialog__fields">
          <label class="account-profile-field">
            <span>{{ t('account.profile.building') }}</span>
            <AppGlassSelect
              :model-value="selectedBuildingID"
              :options="buildingOptions"
              :disabled="isBuildingsLoading || buildingOptions.length <= 1"
              @update:model-value="handleUnitBuildingChange"
            />
          </label>

          <label class="account-profile-field">
            <span>{{ t('account.profile.floor') }}</span>
            <AppGlassSelect
              :model-value="selectedFloor"
              :options="floorOptions"
              :disabled="!selectedBuildingID || isUnitsLoading || floorOptions.length <= 1"
              @update:model-value="handleUnitFloorChange"
            />
          </label>

          <label class="account-profile-field">
            <span>{{ t('account.profile.unit') }}</span>
            <AppGlassSelect
              v-model="selectedUnitID"
              :options="unitOptions"
              :disabled="!selectedFloor || isUnitsLoading || unitOptions.length <= 1"
            />
          </label>
        </div>

        <footer class="account-profile-dialog__actions">
          <BaseButton
            variant="secondary"
            size="md"
            type="button"
            :disabled="isSavingUnit"
            @click="closeUnitModal"
          >
            {{ t('marketplace.myProfile.cancelAction') }}
          </BaseButton>
          <BaseButton
            variant="primary"
            size="md"
            type="submit"
            :disabled="!canSaveUnit || isSavingUnit"
          >
            {{ isSavingUnit ? t('account.profile.savingUnit') : t('account.profile.saveUnit') }}
          </BaseButton>
        </footer>
      </form>
    </div>

    <div
      v-if="isIsmartModalOpen"
      class="account-profile-modal"
      role="presentation"
      @click.self="closeIsmartModal"
    >
      <form
        class="account-profile-dialog account-profile-dialog--compact"
        role="dialog"
        aria-modal="true"
        :aria-label="t('account.profile.ismartDialogTitle')"
        @submit.prevent="handleBindIsmart"
      >
        <header class="account-profile-dialog__header">
          <div>
            <p class="account-profile-kicker">
              {{ t('account.profile.ismartAccount') }}
            </p>
            <h2>{{ t('account.profile.ismartDialogTitle') }}</h2>
          </div>
          <button
            type="button"
            class="account-profile-dialog__close"
            :aria-label="t('marketplace.myProfile.closeEdit')"
            @click="closeIsmartModal"
          >
            <AppIcon
              name="close"
              :size="18"
            />
          </button>
        </header>

        <div class="account-profile-dialog__fields">
          <label class="account-profile-field">
            <span>{{ t('account.profile.ismartUsername') }}</span>
            <BaseInput
              :model-value="ismartFormState.account"
              autocomplete="username"
              @update:model-value="ismartFormState.account = $event"
            />
          </label>

          <label class="account-profile-field">
            <span>{{ t('account.profile.ismartPassword') }}</span>
            <BaseInput
              :model-value="ismartFormState.password"
              type="password"
              autocomplete="current-password"
              @update:model-value="ismartFormState.password = $event"
            />
          </label>
        </div>

        <footer class="account-profile-dialog__actions">
          <BaseButton
            variant="secondary"
            size="md"
            type="button"
            :disabled="isBindingIsmart"
            @click="closeIsmartModal"
          >
            {{ t('marketplace.myProfile.cancelAction') }}
          </BaseButton>
          <BaseButton
            variant="primary"
            size="md"
            type="submit"
            :disabled="isBindingIsmart"
          >
            {{
              isBindingIsmart
                ? t('common.status.loading')
                : sessionStore.me?.ismart_linked
                  ? t('account.profile.updateIsmart')
                  : t('account.profile.bindIsmart')
            }}
          </BaseButton>
        </footer>
      </form>
    </div>
  </div>
</template>

<style scoped>
.account-profile-page {
  display: grid;
  gap: 18px;
  padding-bottom: 48px;
  color: rgb(var(--color-text));
}

.account-profile-hero {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 20px 24px;
  box-shadow: 0 1px 3px rgb(0 0 0 / 0.05);
}

.account-profile-hero__identity {
  display: flex;
  align-items: center;
  gap: 16px;
  min-width: 0;
}

.account-profile-hero__text {
  min-width: 0;
}

.account-profile-hero__text h1 {
  margin: 0;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: 22px;
  font-weight: 600;
  line-height: 1.3;
  overflow-wrap: anywhere;
}

.account-profile-hero__text p {
  margin: 4px 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 13px;
  font-weight: 600;
  line-height: 1.4;
}

.account-profile-hero__actions {
  display: flex;
  gap: 10px;
  flex-shrink: 0;
}

.account-profile-loading {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 16px 20px;
  color: rgb(var(--color-text-muted));
  font-size: 14px;
  font-weight: 600;
  text-align: center;
}

.account-profile-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 18px;
}

.account-profile-card {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  box-shadow: 0 1px 3px rgb(0 0 0 / 0.05);
  overflow: hidden;
}

.account-profile-card--wide {
  grid-column: 1 / -1;
}

.account-profile-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border-bottom: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface-raised));
  padding: 14px 18px;
}

.account-profile-card__header h2 {
  margin: 0;
  color: rgb(var(--color-primary));
  font-size: 15px;
  font-weight: 700;
  line-height: 1.3;
}

.account-profile-card__content {
  padding: 18px;
}

.account-profile-card__list {
  display: grid;
  margin: 0;
}

.account-profile-card__list--compact {
  gap: 4px;
}

.account-profile-card__row {
  display: grid;
  grid-template-columns: 140px minmax(0, 1fr);
  gap: 16px;
  align-items: start;
  padding: 12px 18px;
  border-bottom: 1px solid rgb(var(--color-border));
}

.account-profile-card__row:last-child {
  border-bottom: 0;
}

.account-profile-card__list--compact .account-profile-card__row {
  padding: 10px 18px;
}

.account-profile-card__row dt {
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  font-weight: 600;
  line-height: 1.5;
}

.account-profile-card__row dd {
  min-width: 0;
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 13px;
  font-weight: 600;
  line-height: 1.5;
  overflow-wrap: anywhere;
}

.account-profile-chip-group {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.account-profile-chip-group + .account-profile-chip-group {
  margin-top: 12px;
}

.account-profile-chip {
  display: inline-flex;
  min-height: 28px;
  align-items: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface-raised));
  padding: 6px 12px;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  font-weight: 700;
  line-height: 1;
}

.account-profile-chip--role {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.account-profile-unit-display {
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 14px;
  font-weight: 600;
  line-height: 1.5;
}

.account-profile-empty-state {
  margin: 0 0 16px;
  color: rgb(var(--color-text-muted));
  font-size: 13px;
  font-weight: 600;
  line-height: 1.6;
}

.account-profile-ismart-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 20px;
  padding: 18px;
}

.account-profile-ismart-section h3 {
  margin: 0 0 12px;
  color: rgb(var(--color-text));
  font-size: 13px;
  font-weight: 700;
  line-height: 1.3;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.account-profile-modal {
  position: fixed;
  inset: 0;
  z-index: 60;
  display: grid;
  place-items: center;
  background: rgb(26 28 27 / 0.45);
  padding: 20px;
}

.account-profile-dialog {
  width: min(100%, 540px);
  max-height: calc(100svh - 40px);
  overflow-y: auto;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  box-shadow: 0 20px 60px rgb(0 0 0 / 0.3);
}

.account-profile-dialog--compact {
  width: min(100%, 440px);
}

.account-profile-dialog__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  border-bottom: 1px solid rgb(var(--color-border));
  padding: 20px 24px;
}

.account-profile-kicker {
  margin: 0 0 6px;
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.account-profile-dialog__header h2 {
  margin: 0;
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 20px;
  font-weight: 600;
  line-height: 1.3;
}

.account-profile-dialog__close {
  display: inline-flex;
  height: 32px;
  width: 32px;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
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
  gap: 16px;
  align-items: center;
  border-bottom: 1px solid rgb(var(--color-border));
  padding: 20px 24px;
}

.account-avatar-editor__copy {
  min-width: 0;
}

.account-avatar-editor__copy span {
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.account-avatar-editor__copy p {
  margin: 6px 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  line-height: 1.5;
}

.account-avatar-upload-button {
  display: inline-flex;
  min-height: 36px;
  flex-shrink: 0;
  cursor: pointer;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(var(--color-primary));
  border-radius: 2px;
  background: rgb(var(--color-primary));
  padding: 8px 16px;
  color: rgb(var(--color-primary-contrast));
  font-size: 13px;
  font-weight: 700;
  line-height: 1;
  transition:
    background-color 0.2s ease,
    transform 0.1s ease;
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
  opacity: 0.6;
}

.account-profile-dialog__fields {
  display: grid;
  gap: 16px;
  padding: 20px 24px;
}

.account-profile-phone-grid {
  display: grid;
  grid-template-columns: minmax(90px, 0.4fr) minmax(0, 1fr);
  gap: 12px;
}

.account-profile-field {
  display: grid;
  gap: 8px;
  min-width: 0;
}

.account-profile-field span {
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.account-profile-dialog__readonly {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  border-top: 1px solid rgb(var(--color-border));
  border-bottom: 1px solid rgb(var(--color-border));
}

.account-profile-dialog__readonly div {
  display: grid;
  gap: 8px;
  min-width: 0;
  padding: 16px 24px;
}

.account-profile-dialog__readonly span {
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.account-profile-dialog__readonly strong {
  min-width: 0;
  color: rgb(var(--color-text));
  font-size: 13px;
  font-weight: 600;
  line-height: 1.5;
  overflow-wrap: anywhere;
}

.account-profile-dialog__actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 20px 24px;
}

@media (min-width: 1024px) {
  .account-profile-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 1023px) {
  .account-profile-page {
    padding-bottom: calc(var(--app-mobile-content-bottom) + 16px);
  }

  .account-profile-modal {
    place-items: end center;
    padding:
      var(--app-safe-top)
      var(--layout-page-padding-inline)
      calc(var(--app-safe-bottom) + 10px);
  }

  .account-profile-dialog {
    max-height: calc(100svh - var(--app-safe-top) - var(--app-safe-bottom) - 20px);
  }
}

@media (max-width: 767px) {
  .account-profile-hero {
    flex-direction: column;
    align-items: stretch;
  }

  .account-profile-hero__actions {
    flex-direction: column;
    width: 100%;
  }

  .account-profile-grid {
    grid-template-columns: 1fr;
  }

  .account-profile-card__row {
    grid-template-columns: 1fr;
    gap: 6px;
  }

  .account-profile-phone-grid {
    grid-template-columns: 1fr;
  }

  .account-avatar-editor {
    grid-template-columns: 1fr;
    gap: 14px;
    text-align: center;
  }

  .account-avatar-upload-button {
    width: 100%;
  }

  .account-profile-dialog__actions {
    flex-direction: column;
  }

  .account-profile-ismart-grid {
    grid-template-columns: 1fr;
  }
}

:deep(.app-button),
:deep(.app-input-shell) {
  border-radius: 2px;
}
</style>
