<!--
 * 登入頁表單面板。
 * 1. 展示手提電話密碼、用戶名稱 / 電郵 / iSmart 密碼與住戶註冊表單。
 * 2. 將表單輸入與操作事件回傳給頁面入口。
-->
<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';

import AppIcon from '@/shared/components/base/AppIcon.vue';

import type { LoginAuthMode, LoginEmailAction, LoginSelectOption } from '../login';

interface LoginFormPanelProps {
  authMode: LoginAuthMode;
  emailAction: LoginEmailAction;
  emailActionSwitchLabel: string;
  buildingOptions: LoginSelectOption[];
  buildingsLoading: boolean;
  engName: string;
  email: string;
  password: string;
  phone: string;
  phoneCountryCode: string;
  publisherIdentityType: string;
  ismartAccount: string;
  primaryCommunityId: string;
  residenceFloor: string;
  residenceUnit: string;
  rememberMe: boolean;
  residenceFloorOptions: LoginSelectOption[];
  residenceUnitOptions: LoginSelectOption[];
  submitting: boolean;
  unitsLoading: boolean;
  emailPlaceholder: string;
  footerPrompt: string;
  isAuthenticated: boolean;
  publisherIdentityOptions: LoginSelectOption[];
  submitLabel: string;
}

const props = defineProps<LoginFormPanelProps>();

const emit = defineEmits<{
  'update:engName': [value: string];
  'update:email': [value: string];
  'update:password': [value: string];
  'update:phone': [value: string];
  'update:phoneCountryCode': [value: string];
  'update:publisherIdentityType': [value: string];
  'update:ismartAccount': [value: string];
  'update:primaryCommunityId': [value: string];
  'update:residenceFloor': [value: string];
  'update:residenceUnit': [value: string];
  'update:rememberMe': [value: boolean];
  'set-auth-mode': [value: LoginAuthMode];
  'sign-out': [];
  'submit-login': [];
  'forgot-password': [];
  'toggle-email-action': [];
}>();

const { t } = useI18n();
const showPassword = ref(false);

// 1. 讀取文字輸入值
const readInputValue = (event: Event): string => (event.target as HTMLInputElement).value;

// 2. 讀取勾選輸入值
const readCheckboxValue = (event: Event): boolean => (event.target as HTMLInputElement).checked;

// 3. 選擇手提電話密碼登入
const selectPhonePasswordLogin = (): void => {
  emit('set-auth-mode', 'phone');
};

// 4. 切換密碼顯示狀態
const togglePasswordVisibility = (): void => {
  showPassword.value = !showPassword.value;
};
</script>

<template>
  <div class="login-form-panel z-20 flex h-full min-h-0 flex-col justify-between bg-surface shadow-soft">
    <div class="login-form-main mx-auto flex min-h-0 w-full max-w-md flex-1 flex-col justify-center">
      <div class="login-form-heading text-center lg:text-left">
        <h2 class="font-display text-3xl leading-tight text-text">
          {{ props.emailAction === 'register' ? t('auth.residentRegister') : t('auth.residentLogin') }}
        </h2>
        <p
          v-if="props.emailAction === 'login'"
          class="mt-3 text-base leading-7 text-text-muted"
        >
          {{ t('auth.note') }}
        </p>
      </div>

      <form
        class="login-form-fields"
        @submit.prevent="emit('submit-login')"
      >
        <div
          v-show="props.emailAction === 'login'"
          class="login-choice-tabs"
          role="tablist"
          aria-label="登入方式"
        >
          <button
            type="button"
            class="login-mode-tab"
            :class="props.authMode === 'phone' ? 'login-mode-tab--active' : ''"
            role="tab"
            :aria-selected="props.authMode === 'phone'"
            @click="selectPhonePasswordLogin"
          >
            {{ t('auth.phoneMode') }}
          </button>
          <button
            type="button"
            class="login-mode-tab"
            :class="props.authMode !== 'phone' ? 'login-mode-tab--active' : ''"
            role="tab"
            :aria-selected="props.authMode !== 'phone'"
            @click="emit('set-auth-mode', 'username')"
          >
            {{ t('auth.loginGroupAccount') }}
          </button>
        </div>

        <template v-if="props.emailAction === 'register'">
          <label class="block">
            <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.engName') }}</span>
            <div class="group relative">
              <input
                :value="props.engName"
                class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                :placeholder="t('auth.engNamePlaceholder')"
                autocomplete="name"
                spellcheck="false"
                type="text"
                @input="emit('update:engName', readInputValue($event))"
              >
              <AppIcon
                name="user"
                :size="20"
                class="pointer-events-none absolute right-4 top-1/2 -translate-y-1/2 text-text-muted transition group-focus-within:text-primary"
              />
            </div>
          </label>

          <label class="block">
            <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.password') }}</span>
            <div class="group relative">
              <input
                :value="props.password"
                class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                :placeholder="t('auth.passwordPlaceholder')"
                autocomplete="new-password"
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
          </label>

          <label class="block">
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
          </label>

          <label class="block">
            <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.email') }}</span>
            <div class="group relative">
              <input
                :value="props.email"
                class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                :placeholder="props.emailPlaceholder"
                autocapitalize="none"
                autocomplete="email"
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
          </label>

          <label class="block">
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

          <label class="block">
            <span class="login-building-label mb-1.5 text-sm font-bold uppercase tracking-[0.16em] text-text">
              {{ t('auth.residenceBuilding') }}
              <small class="login-field-hint">{{ t('auth.optionalFieldHint') }}</small>
              <small class="login-building-label__hint">
                {{ t('auth.bindAjoPlatformDescription') }}
              </small>
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
            <small>{{ t('auth.missingBuildingHint') }}</small>
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

        <template v-else-if="props.authMode === 'phone'">
          <label class="block">
            <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.phone') }}</span>
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
          </label>

          <label class="block">
            <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.password') }}</span>
            <div class="group relative">
              <input
                :value="props.password"
                class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                :placeholder="t('auth.passwordPlaceholder')"
                autocomplete="current-password"
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
          </label>
        </template>

        <template v-else>
          <label class="block">
            <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.loginGroupAccount') }}</span>
            <div class="group relative">
              <input
                :value="props.ismartAccount"
                class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                :placeholder="t('auth.accountIdentifierPlaceholder')"
                autocapitalize="none"
                autocomplete="username"
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
          </label>

          <label class="block">
            <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.password') }}</span>
            <div class="group relative">
              <input
                :value="props.password"
                class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                :placeholder="t('auth.passwordPlaceholder')"
                autocomplete="current-password"
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

        <button
          class="login-submit-button w-full rounded-lg bg-primary text-sm font-bold uppercase tracking-[0.24em] text-primary-contrast shadow-[0_10px_30px_rgba(0,0,0,0.04)] transition hover:bg-text active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-60"
          type="submit"
          :disabled="props.submitting"
        >
          {{ props.submitLabel }}
        </button>

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

.login-choice-tabs {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.6rem;
}

.login-mode-tab {
  position: relative;
  min-height: 38px;
  padding: 0.65rem 0.75rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  font-weight: 600;
  line-height: 1.2;
  text-align: center;
  transition:
    border-color 160ms ease,
    background-color 160ms ease,
    color 160ms ease;
}

.login-mode-tab:hover {
  border-color: rgb(var(--color-primary) / 0.45);
  color: rgb(var(--color-text));
}

.login-mode-tab--active {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary-soft));
  color: rgb(var(--color-text));
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

.login-building-label {
  display: block !important;
}

.login-building-label__hint {
  color: rgb(var(--color-text-muted));
  font-size: 10px;
  font-weight: 500;
  letter-spacing: 0;
  line-height: 1.45;
  margin-left: 0.45rem;
  text-transform: none;
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

  .login-choice-tabs {
    gap: 8px;
  }

  .login-mode-tab,
  .login-submit-button {
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
