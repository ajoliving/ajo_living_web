<!--
 * 登入頁表單面板。
 * 1. 展示郵箱密碼、手機密碼、郵箱驗證碼、用戶名密碼與住戶註冊表單。
 * 2. 將表單輸入與操作事件回傳給頁面入口。
-->
<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';

import AppIcon from '@/shared/components/base/AppIcon.vue';

import type { LoginAccountSource, LoginAuthMode, LoginEmailAction, LoginEmailMethod, LoginSelectOption } from '../login';

interface LoginFormPanelProps {
  accountSource: LoginAccountSource;
  authMode: LoginAuthMode;
  emailAction: LoginEmailAction;
  emailActionSwitchLabel: string;
  emailLoginMethod: LoginEmailMethod;
  buildingOptions: LoginSelectOption[];
  buildingsLoading: boolean;
  displayName: string;
  email: string;
  password: string;
  phone: string;
  ismartAccount: string;
  otp: string;
  primaryCommunityId: string;
  residenceFloor: string;
  residenceUnit: string;
  rememberMe: boolean;
  requestingOtp: boolean;
  residenceFloorOptions: LoginSelectOption[];
  residenceUnitOptions: LoginSelectOption[];
  submitting: boolean;
  unitsLoading: boolean;
  emailPlaceholder: string;
  footerPrompt: string;
  isAuthenticated: boolean;
  otpRequestLabel: string;
  submitLabel: string;
}

const props = defineProps<LoginFormPanelProps>();

const emit = defineEmits<{
  'update:displayName': [value: string];
  'update:email': [value: string];
  'update:password': [value: string];
  'update:phone': [value: string];
  'update:ismartAccount': [value: string];
  'update:otp': [value: string];
  'update:primaryCommunityId': [value: string];
  'update:residenceFloor': [value: string];
  'update:residenceUnit': [value: string];
  'update:rememberMe': [value: boolean];
  'request-otp': [];
  'set-account-source': [value: LoginAccountSource];
  'set-auth-mode': [value: LoginAuthMode];
  'set-email-login-method': [value: LoginEmailMethod];
  'sign-out': [];
  'submit-login': [];
  'toggle-email-action': [];
}>();

const { t } = useI18n();
const showPassword = ref(false);

// 1. 讀取文字輸入值
const readInputValue = (event: Event): string => (event.target as HTMLInputElement).value;

// 2. 讀取勾選輸入值
const readCheckboxValue = (event: Event): boolean => (event.target as HTMLInputElement).checked;

// 3. 選擇郵箱登入方式
const selectEmailLoginMethod = (method: LoginEmailMethod): void => {
  emit('set-auth-mode', 'email');
  emit('set-email-login-method', method);
};

// 4. 選擇手機密碼登入
const selectPhonePasswordLogin = (): void => {
  emit('set-auth-mode', 'phone');
};

// 5. 切換密碼顯示狀態
const togglePasswordVisibility = (): void => {
  showPassword.value = !showPassword.value;
};

// 6. 切換郵箱登入方式
const toggleEmailLoginMethod = (): void => {
  selectEmailLoginMethod(props.emailLoginMethod === 'password' ? 'code' : 'password');
};
</script>

<template>
  <div class="login-form-panel z-20 flex h-full min-h-0 flex-col justify-between overflow-hidden bg-surface shadow-soft">
    <div class="login-form-brand flex flex-col items-center text-center lg:items-start lg:text-left">
      <p class="font-display text-2xl uppercase tracking-[0.26em] text-text">
        AJO Living
      </p>
      <div class="h-1 w-12 rounded-full bg-primary" />
    </div>

    <div class="login-form-main mx-auto flex min-h-0 w-full max-w-md flex-1 flex-col justify-center">
      <div class="login-form-heading text-center lg:text-left">
        <h2 class="font-display text-3xl leading-tight text-text">
          {{ props.emailAction === 'register' ? t('auth.residentRegister') : t('auth.residentLogin') }}
        </h2>
        <p class="mt-3 text-base leading-7 text-text-muted">
          {{ props.emailAction === 'register' ? t('auth.registerNote') : t('auth.note') }}
        </p>
      </div>

      <form
        class="login-form-fields"
        @submit.prevent="emit('submit-login')"
      >
        <div
          v-show="props.emailAction === 'register'"
          class="login-mode-tabs login-mode-tabs--register"
        >
          <button
            type="button"
            class="login-mode-tab"
            :class="props.accountSource === 'local' ? 'login-mode-tab--active' : ''"
            @click="emit('set-account-source', 'local')"
          >
            {{ t('auth.newUserAccount') }}
          </button>
          <button
            type="button"
            class="login-mode-tab"
            :class="props.accountSource === 'ismart' ? 'login-mode-tab--active' : ''"
            @click="emit('set-account-source', 'ismart')"
          >
            {{ t('auth.existingIsmartAccount') }}
          </button>
        </div>

        <div
          v-show="props.emailAction === 'login'"
          class="login-mode-tabs"
        >
          <button
            type="button"
            class="login-mode-tab"
            :class="props.authMode === 'phone' ? 'login-mode-tab--active' : ''"
            @click="selectPhonePasswordLogin"
          >
            {{ t('auth.phoneMode') }}
          </button>
          <button
            type="button"
            class="login-mode-tab"
            :class="props.authMode === 'username' ? 'login-mode-tab--active' : ''"
            @click="emit('set-auth-mode', 'username')"
          >
            {{ t('auth.usernameMode') }}
          </button>
          <button
            type="button"
            class="login-mode-tab"
            :class="props.authMode === 'email' ? 'login-mode-tab--active' : ''"
            @click="selectEmailLoginMethod(props.emailLoginMethod)"
          >
            {{ t('auth.emailMode') }}
          </button>
        </div>

        <template v-if="props.emailAction === 'register' && props.accountSource === 'ismart'">
          <label class="block">
            <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.ismartAccount') }}</span>
            <div class="group relative">
              <input
                :value="props.ismartAccount"
                class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                :placeholder="t('auth.ismartAccountPlaceholder')"
                autocomplete="username"
                type="text"
                @input="emit('update:ismartAccount', readInputValue($event))"
              >
              <AppIcon
                name="user"
                :size="20"
                class="absolute right-4 top-1/2 -translate-y-1/2 text-text-muted transition group-focus-within:text-primary"
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

          <label class="block">
            <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.bindEmailOptional') }}</span>
            <div class="group relative">
              <input
                :value="props.email"
                class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                :placeholder="props.emailPlaceholder"
                autocomplete="email"
                type="email"
                @input="emit('update:email', readInputValue($event))"
              >
              <AppIcon
                name="user"
                :size="20"
                class="absolute right-4 top-1/2 -translate-y-1/2 text-text-muted transition group-focus-within:text-primary"
              />
            </div>
          </label>
        </template>

        <template v-else-if="props.emailAction === 'register'">
          <label class="block">
            <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.username') }}</span>
            <div class="group relative">
              <input
                :value="props.displayName"
                class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                :placeholder="t('auth.usernamePlaceholder')"
                autocomplete="username"
                type="text"
                @input="emit('update:displayName', readInputValue($event))"
              >
              <AppIcon
                name="user"
                :size="20"
                class="absolute right-4 top-1/2 -translate-y-1/2 text-text-muted transition group-focus-within:text-primary"
              />
            </div>
          </label>

          <label class="block">
            <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.phone') }}</span>
            <div class="group relative">
              <input
                :value="props.phone"
                class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                :placeholder="t('auth.phonePlaceholder')"
                autocomplete="tel"
                type="tel"
                @input="emit('update:phone', readInputValue($event))"
              >
              <AppIcon
                name="phone"
                :size="20"
                class="absolute right-4 top-1/2 -translate-y-1/2 text-text-muted transition group-focus-within:text-primary"
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
            <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.residenceBuilding') }}</span>
            <div class="group relative">
              <select
                :value="props.primaryCommunityId"
                class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition focus:ring-1 focus:ring-primary"
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
                name="building"
                :size="20"
                class="absolute right-4 top-1/2 -translate-y-1/2 text-text-muted transition group-focus-within:text-primary"
              />
            </div>
          </label>

          <div class="grid gap-3 sm:grid-cols-2">
            <label class="block">
              <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.residenceFloor') }}</span>
              <div class="group relative">
                <select
                  :value="props.residenceFloor"
                  class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition focus:ring-1 focus:ring-primary"
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
                  name="home"
                  :size="20"
                  class="absolute right-4 top-1/2 -translate-y-1/2 text-text-muted transition group-focus-within:text-primary"
                />
              </div>
            </label>

            <label class="block">
              <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.residenceUnit') }}</span>
              <div class="group relative">
                <select
                  :value="props.residenceUnit"
                  class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition focus:ring-1 focus:ring-primary"
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
                  name="browse"
                  :size="20"
                  class="absolute right-4 top-1/2 -translate-y-1/2 text-text-muted transition group-focus-within:text-primary"
                />
              </div>
            </label>
          </div>

          <label class="block">
            <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.email') }}</span>
            <div class="group relative">
              <input
                :value="props.email"
                class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                :placeholder="props.emailPlaceholder"
                autocomplete="email"
                type="email"
                @input="emit('update:email', readInputValue($event))"
              >
              <AppIcon
                name="user"
                :size="20"
                class="absolute right-4 top-1/2 -translate-y-1/2 text-text-muted transition group-focus-within:text-primary"
              />
            </div>
          </label>
        </template>

        <template v-else-if="props.authMode === 'phone'">
          <label class="block">
            <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.phone') }}</span>
            <div class="group relative">
              <input
                :value="props.phone"
                class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                :placeholder="t('auth.phonePlaceholder')"
                autocomplete="tel"
                type="tel"
                @input="emit('update:phone', readInputValue($event))"
              >
              <AppIcon
                name="phone"
                :size="20"
                class="absolute right-4 top-1/2 -translate-y-1/2 text-text-muted transition group-focus-within:text-primary"
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

        <template v-else-if="props.authMode === 'username'">
          <label class="block">
            <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.username') }}</span>
            <div class="group relative">
              <input
                :value="props.ismartAccount"
                class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                :placeholder="t('auth.usernamePlaceholder')"
                autocomplete="username"
                type="text"
                @input="emit('update:ismartAccount', readInputValue($event))"
              >
              <AppIcon
                name="user"
                :size="20"
                class="absolute right-4 top-1/2 -translate-y-1/2 text-text-muted transition group-focus-within:text-primary"
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
            <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.email') }}</span>
            <div class="group relative">
              <input
                :value="props.email"
                class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                :placeholder="props.emailPlaceholder"
                autocomplete="email"
                type="email"
                @input="emit('update:email', readInputValue($event))"
              >
              <AppIcon
                name="user"
                :size="20"
                class="absolute right-4 top-1/2 -translate-y-1/2 text-text-muted transition group-focus-within:text-primary"
              />
            </div>
          </label>

          <label
            v-show="props.emailLoginMethod === 'password'"
            class="block"
          >
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

          <label
            v-show="props.emailLoginMethod === 'code'"
            class="block"
          >
            <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.emailOtp') }}</span>
            <div class="grid gap-3 sm:grid-cols-[1fr_auto]">
              <input
                :value="props.otp"
                class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                :placeholder="t('auth.emailOtpPlaceholder')"
                autocomplete="one-time-code"
                inputmode="numeric"
                type="text"
                @input="emit('update:otp', readInputValue($event))"
              >
              <button
                type="button"
                class="rounded-lg border border-border px-4 py-3 text-sm font-bold text-text-muted transition hover:bg-surface-raised hover:text-text disabled:cursor-not-allowed disabled:opacity-50"
                :disabled="props.requestingOtp"
                @click="emit('request-otp')"
              >
                {{ props.otpRequestLabel }}
              </button>
            </div>
          </label>
        </template>

        <div
          v-show="props.emailAction === 'login'"
          class="flex items-center justify-between gap-4"
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
            v-show="props.authMode === 'email'"
            type="button"
            class="ml-auto rounded-md px-1 py-0.5 text-sm font-bold text-primary transition hover:text-text hover:underline"
            @click="toggleEmailLoginMethod"
          >
            {{ props.emailLoginMethod === 'password' ? t('auth.useCodeLogin') : t('auth.usePasswordLogin') }}
          </button>
        </div>

        <button
          class="login-submit-button w-full rounded-lg bg-primary text-sm font-bold uppercase tracking-[0.24em] text-primary-contrast shadow-[0_10px_30px_rgba(0,0,0,0.04)] transition hover:bg-text active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-60"
          type="submit"
          :disabled="props.submitting"
        >
          {{ props.submitLabel }}
        </button>
      </form>
    </div>

    <div
      v-if="props.isAuthenticated || props.emailAction === 'login' || props.accountSource === 'local'"
      class="login-form-footer border-t border-border/70 text-center text-sm leading-6 text-text-muted"
    >
      <span v-if="props.emailAction === 'login' || props.accountSource === 'local'">{{ props.footerPrompt }}</span>
      <button
        v-if="props.emailAction === 'login' || props.accountSource === 'local'"
        type="button"
        class="ml-1 font-bold text-primary transition hover:underline"
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
  </div>
</template>

<style scoped>
.login-form-panel {
  padding:
    clamp(2.75rem, 6svh, 4rem)
    clamp(1.5rem, 3vw, 2.25rem)
    clamp(1rem, 3svh, 1.5rem);
  box-shadow: none;
}

.login-form-brand {
  gap: clamp(0.45rem, 1svh, 0.75rem);
  flex-shrink: 0;
}

.login-form-main {
  padding-block: clamp(0.45rem, 1.8svh, 1rem);
  overflow-y: auto;
  overscroll-behavior: contain;
}

.login-form-heading {
  margin-bottom: clamp(0.8rem, 1.8svh, 1.25rem);
}

.login-form-heading h2 {
  font-family: var(--font-display);
  font-size: 28px;
  font-weight: 400;
}

.login-form-heading p {
  color: rgb(var(--color-text-muted));
  font-size: 13px;
  line-height: 1.7;
}

.login-form-heading p {
  margin-top: clamp(0.55rem, 1.4svh, 0.75rem);
}

.login-form-fields {
  display: grid;
  gap: clamp(0.55rem, 1.35svh, 0.85rem);
}

.login-form-input {
  height: 38px;
  border: 1px solid rgb(var(--color-border)) !important;
  border-radius: 2px !important;
  background: rgb(var(--color-surface)) !important;
  padding-top: 6px;
  padding-bottom: 6px;
  font-size: 12px;
}

.login-submit-button {
  min-height: 38px;
  border-radius: 2px !important;
  box-shadow: none !important;
  padding-block: 8px;
  font-size: 11px;
}

.login-form-footer {
  flex-shrink: 0;
  padding-top: clamp(0.75rem, 2svh, 1.5rem);
  font-size: clamp(0.82rem, 1.6svh, 1rem);
}

.login-method-tabs {
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px !important;
  background: rgb(var(--color-surface-muted)) !important;
}

.login-method-tabs button {
  border-radius: 2px !important;
  font-size: 11px;
  font-weight: 600;
}

.login-mode-tabs {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  align-items: end;
  gap: 1.25rem;
  border-bottom: 1px solid rgb(var(--color-border) / 0.75);
  padding-bottom: 0;
}

.login-mode-tabs--register {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.login-mode-tab {
  position: relative;
  padding: 0 0 0.7rem;
  border: none;
  background: transparent;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  font-weight: 600;
  line-height: 1.2;
  text-align: center;
  transition: color 160ms ease;
}

.login-mode-tab::after {
  content: '';
  position: absolute;
  right: 0;
  bottom: -1px;
  left: 0;
  height: 2px;
  background: rgb(var(--color-primary));
  transform: scaleX(0);
  transform-origin: center;
  transition: transform 160ms ease;
}

.login-mode-tab:hover {
  color: rgb(var(--color-text));
}

.login-mode-tab--active {
  color: rgb(var(--color-text));
}

.login-mode-tab--active::after {
  transform: scaleX(1);
}

:deep(label > span) {
  color: rgb(var(--color-text-muted)) !important;
  font-size: 10px !important;
  font-weight: 600 !important;
  letter-spacing: 0.14em !important;
}

:deep(button.rounded-lg) {
  border-radius: 2px !important;
}

@media (max-height: 760px) {
  .login-form-panel {
    padding-top: clamp(3.75rem, 7svh, 4.6rem);
  }

  .login-form-heading h2 {
    font-size: 1.65rem;
  }

  .login-form-heading p {
    font-size: 0.92rem;
    line-height: 1.55;
  }

  .login-form-footer {
    line-height: 1.45;
  }
}

@media (max-width: 1023px) {
  .login-form-panel {
    padding-inline: clamp(2rem, 8vw, 3rem);
  }
}
</style>
