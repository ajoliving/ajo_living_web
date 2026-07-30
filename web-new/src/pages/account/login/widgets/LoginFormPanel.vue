<!--
 * 登入頁表單面板。
 * 1. 展示統一帳戶密碼登入與用戶註冊表單。
 * 2. 將表單輸入與操作事件回傳給頁面入口。
-->
<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';

import AppIcon from '@/shared/components/base/AppIcon.vue';

import type { LoginEmailAction, LoginSelectOption, LoginValidationErrors, RegistrationStep } from '../login';

interface LoginFormPanelProps {
  emailAction: LoginEmailAction;
  emailActionSwitchLabel: string;
  buildingOptions: LoginSelectOption[];
  buildingsLoading: boolean;
  engName: string;
  chiName: string;
  email: string;
  password: string;
  confirmPassword: string;
  username: string;
  phone: string;
  phoneCountryCode: string;
  publisherIdentityType: string;
  ismartAccount: string;
  primaryCommunityId: string;
  residenceFloor: string;
  residenceUnit: string;
  idCard: string;
  agencyLicenseNumber: string;
  agencyLicenseFile: File | null;
  agencyContactName: string;
  shouldBindResidence: boolean;
  rememberMe: boolean;
  residenceFloorOptions: LoginSelectOption[];
  residenceUnitOptions: LoginSelectOption[];
  submitting: boolean;
  unitsLoading: boolean;
  emailPlaceholder: string;
  footerPrompt: string;
  isAuthenticated: boolean;
  publisherIdentityOptions: LoginSelectOption[];
  registrationStep: RegistrationStep;
  submitLabel: string;
  validationErrors: LoginValidationErrors;
}

const props = defineProps<LoginFormPanelProps>();

const emit = defineEmits<{
  'update:engName': [value: string];
  'update:chiName': [value: string];
  'update:email': [value: string];
  'update:password': [value: string];
  'update:confirmPassword': [value: string];
  'update:username': [value: string];
  'update:phone': [value: string];
  'update:phoneCountryCode': [value: string];
  'update:publisherIdentityType': [value: string];
  'update:ismartAccount': [value: string];
  'update:primaryCommunityId': [value: string];
  'update:residenceFloor': [value: string];
  'update:residenceUnit': [value: string];
  'update:idCard': [value: string];
  'update:agencyLicenseNumber': [value: string];
  'update:agencyLicenseFile': [value: File | null];
  'update:agencyContactName': [value: string];
  'update:shouldBindResidence': [value: boolean];
  'update:rememberMe': [value: boolean];
  'sign-out': [];
  'submit-login': [];
  'forgot-password': [];
  'registration-back': [];
  'toggle-email-action': [];
}>();

const { t } = useI18n();
const showPassword = ref(false);
const showConfirmPassword = ref(false);

// 1. 讀取文字輸入值
const readInputValue = (event: Event): string => (event.target as HTMLInputElement).value;

// 2. 讀取勾選輸入值
const readCheckboxValue = (event: Event): boolean => (event.target as HTMLInputElement).checked;

// 3. 讀取代理牌照證明檔案
const readAgencyLicenseFile = (event: Event): void => {
  const input = event.target as HTMLInputElement;
  emit('update:agencyLicenseFile', input.files?.[0] ?? null);
};

// 4. 切換密碼顯示狀態
const togglePasswordVisibility = (): void => {
  showPassword.value = !showPassword.value;
};

// 5. 切換確認密碼顯示狀態
const toggleConfirmPasswordVisibility = (): void => {
  showConfirmPassword.value = !showConfirmPassword.value;
};
</script>

<template>
  <div class="login-form-panel z-20 flex h-full min-h-0 flex-col justify-between bg-surface shadow-soft">
    <div class="login-form-main mx-auto flex min-h-0 w-full max-w-md flex-1 flex-col justify-center">
      <div class="login-form-heading text-center lg:text-left">
        <h2 class="font-display text-3xl leading-tight text-text">
          {{ props.emailAction === 'register' ? t('auth.residentRegister') : t('auth.residentLogin') }}
        </h2>
      </div>

      <form
        class="login-form-fields"
        @submit.prevent="emit('submit-login')"
      >
        <template v-if="props.emailAction === 'register'">
          <p class="login-registration-step">{{ t('auth.registrationStep', { step: props.registrationStep }) }}</p>
          <label v-if="props.registrationStep === 1" class="block form-field--required">
            <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.publisherIdentityType') }}</span>
            <div class="group relative">
              <select
                :value="props.publisherIdentityType"
                class="login-form-input login-form-select w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition focus:ring-1 focus:ring-primary"
                @change="emit('update:publisherIdentityType', readInputValue($event))"
              >
                <option
                  v-for="option in props.publisherIdentityOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ option.label }}
                </option>
              </select>
              <AppIcon
                name="chevron-down"
                :size="20"
                class="pointer-events-none absolute right-4 top-1/2 -translate-y-1/2 text-text-muted transition group-focus-within:text-primary"
              />
            </div>
          </label>

          <label v-if="props.registrationStep === 2 && props.publisherIdentityType === 'personal'" :class="['block form-field--required', { 'form-field--error': props.validationErrors.username }]">
            <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.username') }}</span>
            <div class="group relative">
              <input
                :value="props.username"
                class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                :placeholder="t('auth.registerUsernamePlaceholder')"
                autocapitalize="none"
                autocomplete="username"
                :aria-invalid="Boolean(props.validationErrors.username)"
                spellcheck="false"
                type="text"
                @input="emit('update:username', readInputValue($event))"
              >
              <AppIcon
                name="user"
                :size="20"
                class="pointer-events-none absolute right-4 top-1/2 -translate-y-1/2 text-text-muted transition group-focus-within:text-primary"
              />
            </div>
            <small v-if="props.validationErrors.username" class="form-field-error-message">{{ t(props.validationErrors.username) }}</small>
          </label>

          <template v-if="props.registrationStep === 2 && props.publisherIdentityType !== 'personal'">
            <label :class="['block form-field--required', { 'form-field--error': props.validationErrors.agencyLicenseNumber }]">
              <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.agencyLicenseNumber') }}</span>
              <div class="group relative">
                <input
                  :value="props.agencyLicenseNumber"
                  class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                  :placeholder="props.publisherIdentityType === 'individual_agent' ? t('auth.individualAgencyLicensePlaceholder') : t('auth.companyAgencyLicensePlaceholder')"
                  autocapitalize="characters"
                  :aria-invalid="Boolean(props.validationErrors.agencyLicenseNumber)"
                  spellcheck="false"
                  type="text"
                  @input="emit('update:agencyLicenseNumber', readInputValue($event))"
                >
                <AppIcon name="shield" :size="20" class="pointer-events-none absolute right-4 top-1/2 -translate-y-1/2 text-text-muted transition group-focus-within:text-primary" />
              </div>
              <small class="login-field-hint">{{ props.publisherIdentityType === 'individual_agent' ? t('auth.individualAgencyLicenseHint') : t('auth.companyAgencyLicenseHint') }}</small>
              <small v-if="props.validationErrors.agencyLicenseNumber" class="form-field-error-message">{{ t(props.validationErrors.agencyLicenseNumber) }}</small>
            </label>

            <label :class="['block form-field--required', { 'form-field--error': props.validationErrors.chiName }]">
              <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ props.publisherIdentityType === 'individual_agent' ? t('auth.chiName') : t('auth.companyNameZh') }}</span>
              <input
                :value="props.chiName"
                class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                :placeholder="props.publisherIdentityType === 'individual_agent' ? t('auth.chiNamePlaceholder') : t('auth.companyNameZhPlaceholder')"
                :aria-invalid="Boolean(props.validationErrors.chiName)"
                type="text"
                @input="emit('update:chiName', readInputValue($event))"
              >
              <small v-if="props.validationErrors.chiName" class="form-field-error-message">{{ t(props.validationErrors.chiName) }}</small>
            </label>

            <label :class="['block form-field--required', { 'form-field--error': props.validationErrors.engName }]">
              <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ props.publisherIdentityType === 'individual_agent' ? t('auth.engName') : t('auth.companyNameEn') }}</span>
              <input
                :value="props.engName"
                class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                :placeholder="props.publisherIdentityType === 'individual_agent' ? t('auth.engNamePlaceholder') : t('auth.companyNameEnPlaceholder')"
                :aria-invalid="Boolean(props.validationErrors.engName)"
                type="text"
                @input="emit('update:engName', readInputValue($event))"
              >
              <small v-if="props.validationErrors.engName" class="form-field-error-message">{{ t(props.validationErrors.engName) }}</small>
            </label>

            <label v-if="props.publisherIdentityType === 'agency_company'" :class="['block form-field--required', { 'form-field--error': props.validationErrors.agencyContactName }]">
              <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.agencyContactName') }}</span>
              <input
                :value="props.agencyContactName"
                class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                :placeholder="t('auth.agencyContactNamePlaceholder')"
                :aria-invalid="Boolean(props.validationErrors.agencyContactName)"
                autocomplete="name"
                type="text"
                @input="emit('update:agencyContactName', readInputValue($event))"
              >
              <small v-if="props.validationErrors.agencyContactName" class="form-field-error-message">{{ t(props.validationErrors.agencyContactName) }}</small>
            </label>
          </template>

          <label v-if="props.registrationStep === 1" :class="['block form-field--required', { 'form-field--error': props.validationErrors.phone }]">
            <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">
              {{ t('auth.phone') }}
              <small class="login-field-hint">{{ t('auth.phoneSmsHint') }}</small>
            </span>
            <div class="phone-input-row group">
              <select
                :value="props.phoneCountryCode"
                class="login-form-input phone-code-select border-none bg-surface-raised text-text outline-none transition focus:ring-1 focus:ring-primary"
                autocomplete="tel-country-code"
                @change="emit('update:phoneCountryCode', readInputValue($event))"
              >
                <option value="+852">+852</option>
                <option value="+86">+86</option>
              </select>
              <input
                :value="props.phone"
                class="login-form-input min-w-0 flex-1 rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                :placeholder="t('auth.phonePlaceholder')"
                autocomplete="tel"
                :aria-invalid="Boolean(props.validationErrors.phone)"
                inputmode="tel"
                type="tel"
                @input="emit('update:phone', readInputValue($event))"
              >
              <AppIcon
                name="phone"
                :size="20"
                class="pointer-events-none absolute right-4 top-1/2 -translate-y-1/2 text-text-muted transition group-focus-within:text-primary"
              />
            </div>
            <small v-if="props.validationErrors.phone" class="form-field-error-message">{{ t(props.validationErrors.phone) }}</small>
          </label>

          <label v-if="props.registrationStep === 1" :class="['block', { 'form-field--required': props.publisherIdentityType !== 'individual_agent', 'form-field--error': props.validationErrors.email }]">
            <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.email') }} <small v-if="props.publisherIdentityType === 'individual_agent'" class="login-field-hint">{{ t('auth.optionalFieldHint') }}</small></span>
            <div class="group relative">
              <input
                :value="props.email"
                class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                :placeholder="props.emailPlaceholder"
                autocapitalize="none"
                autocomplete="email"
                :aria-invalid="Boolean(props.validationErrors.email)"
                inputmode="email"
                spellcheck="false"
                type="email"
                @input="emit('update:email', readInputValue($event))"
              >
              <AppIcon
                name="user"
                :size="20"
                class="pointer-events-none absolute right-4 top-1/2 -translate-y-1/2 text-text-muted transition group-focus-within:text-primary"
              />
            </div>
            <small v-if="props.validationErrors.email" class="form-field-error-message">{{ t(props.validationErrors.email) }}</small>
          </label>

          <label v-if="props.registrationStep === 2" :class="['block form-field--required', { 'form-field--error': props.validationErrors.password }]">
            <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.password') }}</span>
            <div class="group relative">
              <input
                :value="props.password"
                class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                :placeholder="t('auth.passwordPlaceholder')"
                autocomplete="new-password"
                :aria-invalid="Boolean(props.validationErrors.password)"
                :type="showPassword ? 'text' : 'password'"
                @input="emit('update:password', readInputValue($event))"
              >
              <button
                type="button"
                class="absolute right-3 top-1/2 flex h-8 w-8 -translate-y-1/2 items-center justify-center rounded-full text-text-muted transition hover:bg-surface hover:text-primary"
                :aria-label="showPassword ? t('auth.hidePassword') : t('auth.showPassword')"
                @click="togglePasswordVisibility"
              >
                <AppIcon :name="showPassword ? 'view-off' : 'view'" :size="18" />
              </button>
            </div>
            <small v-if="props.validationErrors.password" class="form-field-error-message">{{ t(props.validationErrors.password) }}</small>
          </label>

          <label v-if="props.registrationStep === 2" :class="['block form-field--required', { 'form-field--error': props.validationErrors.confirmPassword }]">
            <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.confirmPassword') }}</span>
            <div class="group relative">
              <input
                :value="props.confirmPassword"
                class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                :placeholder="t('auth.confirmPasswordPlaceholder')"
                autocomplete="new-password"
                :aria-invalid="Boolean(props.validationErrors.confirmPassword)"
                :type="showConfirmPassword ? 'text' : 'password'"
                @input="emit('update:confirmPassword', readInputValue($event))"
              >
              <button
                type="button"
                class="absolute right-3 top-1/2 flex h-8 w-8 -translate-y-1/2 items-center justify-center rounded-full text-text-muted transition hover:bg-surface hover:text-primary"
                :aria-label="showConfirmPassword ? t('auth.hidePassword') : t('auth.showPassword')"
                @click="toggleConfirmPasswordVisibility"
              >
                <AppIcon :name="showConfirmPassword ? 'view-off' : 'view'" :size="18" />
              </button>
            </div>
            <small v-if="props.validationErrors.confirmPassword" class="form-field-error-message">{{ t(props.validationErrors.confirmPassword) }}</small>
          </label>

          <label v-if="props.registrationStep === 2 && props.publisherIdentityType !== 'personal'" :class="['block form-field--required', { 'form-field--error': props.validationErrors.agencyLicenseFile }]">
            <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ props.publisherIdentityType === 'individual_agent' ? t('auth.individualAgencyLicenseUpload') : t('auth.companyAgencyLicenseUpload') }}</span>
            <input
              class="login-file-input"
              accept="image/png,image/jpeg,image/webp"
              :aria-invalid="Boolean(props.validationErrors.agencyLicenseFile)"
              type="file"
              @change="readAgencyLicenseFile"
            >
            <small v-if="props.agencyLicenseFile" class="login-file-name">{{ props.agencyLicenseFile.name }}</small>
            <small v-if="props.validationErrors.agencyLicenseFile" class="form-field-error-message">{{ t(props.validationErrors.agencyLicenseFile) }}</small>
            <small class="login-field-hint">{{ t('auth.agencyReviewRequired') }}</small>
          </label>

          <template v-if="props.registrationStep === 2 && props.publisherIdentityType === 'personal'">
            <label class="login-residence-checkbox">
              <input
                :checked="props.shouldBindResidence"
                type="checkbox"
                @change="emit('update:shouldBindResidence', readCheckboxValue($event))"
              >
              <span>{{ t('auth.bindAjoBuilding') }}</span>
            </label>
            <p class="login-residence-description">{{ t('auth.bindAjoPlatformDescription') }}</p>
            <p class="login-residence-description">{{ t('auth.missingBuildingHint') }}</p>

            <template v-if="props.shouldBindResidence">
              <label :class="['block form-field--required', { 'form-field--error': props.validationErrors.engName }]">
                <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.engName') }}</span>
                <div class="group relative">
                  <input
                    :value="props.engName"
                    class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                    :placeholder="t('auth.engNamePlaceholder')"
                    autocomplete="name"
                    :aria-invalid="Boolean(props.validationErrors.engName)"
                    spellcheck="false"
                    type="text"
                    @input="emit('update:engName', readInputValue($event))"
                  >
                  <AppIcon name="user" :size="20" class="pointer-events-none absolute right-4 top-1/2 -translate-y-1/2 text-text-muted transition group-focus-within:text-primary" />
                </div>
                <small v-if="props.validationErrors.engName" class="form-field-error-message">{{ t(props.validationErrors.engName) }}</small>
              </label>

              <label class="block">
                <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">
                  {{ t('auth.chiName') }}
                  <small class="login-field-hint">{{ t('auth.optionalFieldHint') }}</small>
                </span>
                <div class="group relative">
                  <input
                    :value="props.chiName"
                    class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                    :placeholder="t('auth.chiNamePlaceholder')"
                    autocomplete="name"
                    type="text"
                    @input="emit('update:chiName', readInputValue($event))"
                  >
                  <AppIcon name="user" :size="20" class="pointer-events-none absolute right-4 top-1/2 -translate-y-1/2 text-text-muted transition group-focus-within:text-primary" />
                </div>
              </label>

              <label :class="['block form-field--required', { 'form-field--error': props.validationErrors.idCard }]">
                <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.idCard') }}</span>
                <input
                  :value="props.idCard"
                  class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                  :placeholder="t('auth.idCardPlaceholder')"
                  autocomplete="off"
                  :aria-invalid="Boolean(props.validationErrors.idCard)"
                  type="text"
                  @input="emit('update:idCard', readInputValue($event))"
                >
                <small v-if="props.validationErrors.idCard" class="form-field-error-message">{{ t(props.validationErrors.idCard) }}</small>
              </label>
          <label class="block form-field--required">
            <span class="login-building-label mb-1.5 text-sm font-bold uppercase tracking-[0.16em] text-text">
              {{ t('auth.residenceBuilding') }}
            </span>
            <div class="group relative">
              <select
                :value="props.primaryCommunityId"
                class="login-form-input login-form-select w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition focus:ring-1 focus:ring-primary"
                @change="emit('update:primaryCommunityId', readInputValue($event))"
              >
                <option
                  v-for="option in props.buildingOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ option.label }}
                </option>
              </select>
              <AppIcon
                name="chevron-down"
                :size="20"
                class="pointer-events-none absolute right-4 top-1/2 -translate-y-1/2 text-text-muted transition group-focus-within:text-primary"
              />
            </div>
          </label>

          <div class="login-residence-hint">
            <span>{{ t('auth.residenceFloorUnit') }}</span>
            <small v-if="props.validationErrors.residence" class="form-field-error-message">{{ t(props.validationErrors.residence) }}</small>
          </div>

          <div class="grid gap-3 sm:grid-cols-2">
            <label class="block">
              <div class="group relative">
                <select
                  :value="props.residenceFloor"
                  class="login-form-input login-form-select w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition focus:ring-1 focus:ring-primary"
                  :disabled="props.unitsLoading || !props.primaryCommunityId"
                  @change="emit('update:residenceFloor', readInputValue($event))"
                >
                  <option
                    v-for="option in props.residenceFloorOptions"
                    :key="option.value"
                    :value="option.value"
                  >
                    {{ option.label }}
                  </option>
                </select>
                <AppIcon
                  name="chevron-down"
                  :size="20"
                  class="pointer-events-none absolute right-4 top-1/2 -translate-y-1/2 text-text-muted transition group-focus-within:text-primary"
                />
              </div>
            </label>

            <label class="block">
              <div class="group relative">
                <select
                  :value="props.residenceUnit"
                  class="login-form-input login-form-select w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition focus:ring-1 focus:ring-primary"
                  :disabled="props.unitsLoading || !props.residenceFloor"
                  @change="emit('update:residenceUnit', readInputValue($event))"
                >
                  <option
                    v-for="option in props.residenceUnitOptions"
                    :key="option.value"
                    :value="option.value"
                  >
                    {{ option.label }}
                  </option>
                </select>
                <AppIcon
                  name="chevron-down"
                  :size="20"
                  class="pointer-events-none absolute right-4 top-1/2 -translate-y-1/2 text-text-muted transition group-focus-within:text-primary"
                />
              </div>
            </label>
          </div>
          </template>
          </template>
        </template>

        <template v-else>
          <label :class="['block form-field--required', { 'form-field--error': props.validationErrors.ismartAccount }]">
            <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.loginGroupAccount') }}</span>
            <div class="group relative">
              <input
                :value="props.ismartAccount"
                class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                :placeholder="t('auth.accountIdentifierPlaceholder')"
                autocapitalize="none"
                autocomplete="username"
                :aria-invalid="Boolean(props.validationErrors.ismartAccount)"
                spellcheck="false"
                type="text"
                @input="emit('update:ismartAccount', readInputValue($event))"
              >
              <AppIcon
                name="user"
                :size="20"
                class="pointer-events-none absolute right-4 top-1/2 -translate-y-1/2 text-text-muted transition group-focus-within:text-primary"
              />
            </div>
            <small v-if="props.validationErrors.ismartAccount" class="form-field-error-message">{{ t(props.validationErrors.ismartAccount) }}</small>
          </label>

          <label :class="['block form-field--required', { 'form-field--error': props.validationErrors.password }]">
            <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.password') }}</span>
            <div class="group relative">
              <input
                :value="props.password"
                class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                :placeholder="t('auth.passwordPlaceholder')"
                autocomplete="current-password"
                :aria-invalid="Boolean(props.validationErrors.password)"
                :type="showPassword ? 'text' : 'password'"
                @input="emit('update:password', readInputValue($event))"
              >
              <button
                type="button"
                class="absolute right-3 top-1/2 flex h-8 w-8 -translate-y-1/2 items-center justify-center rounded-full text-text-muted transition hover:bg-surface hover:text-primary"
                :aria-label="showPassword ? t('auth.hidePassword') : t('auth.showPassword')"
                @click="togglePasswordVisibility"
              >
                <AppIcon
                  :name="showPassword ? 'view-off' : 'view'"
                  :size="18"
                />
              </button>
            </div>
            <small v-if="props.validationErrors.password" class="form-field-error-message">{{ t(props.validationErrors.password) }}</small>
          </label>
        </template>

        <div
          v-show="props.emailAction === 'login'"
          class="login-help-row"
        >
          <label class="flex cursor-pointer items-center gap-2">
            <input
              :checked="props.rememberMe"
              class="h-4 w-4 rounded border-border text-primary focus:ring-primary/20"
              type="checkbox"
              @change="emit('update:rememberMe', readCheckboxValue($event))"
            >
            <span class="text-sm font-semibold text-text-muted">{{ t('auth.rememberMe') }}</span>
          </label>
          <button
            type="button"
            class="login-forgot-button"
            @click="emit('forgot-password')"
          >
            {{ t('auth.forgotPassword') }}
          </button>
        </div>

        <div
          :class="{ 'login-registration-actions': props.emailAction === 'register' && props.registrationStep === 2 }"
        >
          <button
            v-if="props.emailAction === 'register' && props.registrationStep === 2"
            class="login-back-button"
            type="button"
            :disabled="props.submitting"
            @click="emit('registration-back')"
          >
            {{ t('auth.registrationBack') }}
          </button>

          <button
            class="login-submit-button w-full rounded-lg bg-primary text-sm font-bold uppercase tracking-[0.24em] text-primary-contrast shadow-[0_10px_30px_rgba(0,0,0,0.04)] transition hover:bg-text active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-60"
            type="submit"
            :disabled="props.submitting"
          >
            {{ props.submitLabel }}
          </button>
        </div>

        <div
          v-if="props.isAuthenticated || props.emailAction === 'login' || props.emailAction === 'register'"
          class="login-inline-switch text-center text-sm leading-6 text-text-muted"
        >
          <span v-if="props.emailAction === 'login' || props.emailAction === 'register'">{{ props.footerPrompt }}</span>
          <button
            v-if="props.emailAction === 'login' || props.emailAction === 'register'"
            type="button"
            class="ml-1 inline-flex min-w-[36px] items-center justify-center font-bold text-primary transition hover:underline"
            @click="emit('toggle-email-action')"
          >
            {{ props.emailActionSwitchLabel }}
          </button>
          <button
            v-if="props.isAuthenticated"
            type="button"
            class="ml-4 font-bold text-primary transition hover:underline"
            @click="emit('sign-out')"
          >
            {{ t('common.action.signOut') }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
.login-form-panel {
  min-width: 0;
  max-width: 100%;
  border-left: 1px solid rgb(var(--color-border) / 0.7);
  padding:
    clamp(2.6rem, 5svh, 3.25rem)
    clamp(1.8rem, 3vw, 2.5rem)
    clamp(1rem, 3svh, 1.5rem);
  box-shadow: none;
}

.login-form-main {
  min-width: 0;
  max-width: min(100%, 28rem);
  padding-block: clamp(0.45rem, 1.8svh, 1rem);
  overflow: visible;
}

.login-form-heading {
  margin-bottom: clamp(0.9rem, 1.8svh, 1.25rem);
}

.login-form-heading h2 {
  font-family: var(--font-display);
  font-size: 30px;
  font-weight: 400;
  line-height: 1.2;
}

.login-form-heading p {
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  line-height: 1.7;
}

.login-form-heading p {
  margin-top: clamp(0.55rem, 1.4svh, 0.75rem);
}

.login-registration-step {
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  line-height: 1.5;
  margin: 0;
}

.login-form-fields {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: clamp(0.55rem, 1.35svh, 0.85rem);
  min-width: 0;
  max-width: 100%;
}

.login-form-fields > * {
  min-width: 0;
}

.login-form-input {
  height: 42px;
  border: 1px solid rgb(var(--color-border)) !important;
  border-radius: 2px !important;
  background: rgb(var(--color-surface)) !important;
  padding-top: 8px;
  padding-bottom: 8px;
  font-size: 13px;
}

.login-form-select {
  cursor: pointer;
  appearance: none;
  -webkit-appearance: none;
  -moz-appearance: none;
}

.login-form-select::-ms-expand {
  display: none;
}

.login-form-select:disabled {
  cursor: not-allowed;
  opacity: 0.62;
}

.login-submit-button {
  min-height: 42px;
  border-radius: 2px !important;
  box-shadow: none !important;
  padding-block: 8px;
  font-size: 12px;
}

.login-back-button {
  width: 100%;
  min-height: 42px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.24em;
  text-align: center;
}

.login-back-button:hover {
  border-color: rgb(var(--color-text));
}

.login-registration-actions {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.65rem;
}

.login-registration-actions .login-submit-button {
  width: 100%;
}

.phone-input-row {
  position: relative;
  display: flex;
  gap: 0.5rem;
}

.phone-code-select {
  width: 94px;
  flex: 0 0 94px;
  padding-inline: 0.75rem;
}

.login-field-hint {
  color: inherit;
  font-size: 10px;
  font-weight: 500;
  letter-spacing: 0;
  text-transform: none;
}

.login-residence-checkbox {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: rgb(var(--color-text));
  cursor: pointer;
  font-size: 13px;
  font-weight: 600;
}

.login-residence-checkbox input {
  width: 16px;
  height: 16px;
  accent-color: rgb(var(--color-primary));
}

.login-residence-description {
  margin: -0.25rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  line-height: 1.55;
}

.login-file-input {
  width: 100%;
  border: 1px dashed rgb(var(--color-border));
  border-radius: 2px;
  color: rgb(var(--color-text-muted));
  cursor: pointer;
  font-size: 12px;
  padding: 0.55rem 0.7rem;
}

.login-file-name {
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  overflow-wrap: anywhere;
}

.login-building-label {
  display: block !important;
}

.login-residence-hint {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.75rem;
  margin-top: -0.2rem;
  color: rgb(var(--color-text));
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0.14em;
  line-height: 1.5;
  text-transform: uppercase;
}

.login-residence-hint small {
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0;
  text-align: right;
  text-transform: none;
}

.login-help-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.login-help-row label {
  min-height: 40px;
}

.login-forgot-button {
  color: rgb(var(--color-primary));
  font-size: 13px;
  font-weight: 700;
  transition: color 160ms ease;
}

.login-forgot-button:hover {
  color: rgb(var(--color-text));
  text-decoration: underline;
}

.login-inline-switch {
  padding-top: 0.15rem;
}

:deep(label > span) {
  color: rgb(var(--color-text-muted)) !important;
  font-size: 10px !important;
  font-weight: 500 !important;
  letter-spacing: 0.14em !important;
}

:deep(button.rounded-lg) {
  border-radius: 2px !important;
}

@media (max-height: 760px) {
  .login-form-panel {
    padding-top: clamp(3.25rem, 6svh, 4.1rem);
  }

  .login-form-heading h2 {
    font-size: 1.65rem;
  }

  .login-form-heading p {
    font-size: 0.92rem;
    line-height: 1.55;
  }

}

@media (max-width: 1023px) {
  .login-form-panel {
    border-left: 0;
    padding-inline: clamp(2rem, 8vw, 3rem);
  }
}

@media (max-width: 640px) {
  .login-form-panel {
    padding: 1.1rem var(--layout-page-padding-inline) calc(var(--app-mobile-content-bottom) + 1rem);
  }

  .login-form-main {
    justify-content: start;
    padding-block: 0;
  }

  .login-form-heading {
    text-align: left;
  }

  .login-form-heading h2 {
    font-size: 26px;
  }

  .login-form-input {
    height: 44px;
    font-size: 16px;
  }

  .login-submit-button {
    min-height: 44px;
  }

  .login-back-button {
    min-height: 44px;
  }

  .login-help-row {
    align-items: flex-start;
    flex-direction: column;
    gap: 0.7rem;
  }

  .login-inline-switch {
    text-align: left;
  }
}
</style>
