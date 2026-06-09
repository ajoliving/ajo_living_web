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

// 6. 切換郵箱密碼與驗證碼登入
const toggleEmailLoginMethod = (): void => {
  selectEmailLoginMethod(props.emailLoginMethod === 'password' ? 'code' : 'password');
};
</script>

<template>
  <div class="login-form-panel z-20 flex h-full min-h-0 flex-col justify-between overflow-hidden bg-surface shadow-2xl">
    <div
      v-if="props.emailAction === 'login'"
      class="login-form-brand flex flex-col items-center text-center lg:items-start lg:text-left"
    >
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
          v-show="props.emailAction === 'register'"
          class="login-method-tabs grid grid-cols-2 rounded-lg bg-surface-raised p-1"
        >
          <button
            type="button"
            class="rounded-md px-2 py-2 text-xs font-bold transition sm:text-sm"
            :class="props.accountSource === 'local' ? 'bg-surface text-text shadow-sm' : 'text-text-muted hover:text-text'"
            @click="emit('set-account-source', 'local')"
          >
            {{ t('auth.newUserAccount') }}
          </button>
          <button
            type="button"
            class="rounded-md px-2 py-2 text-xs font-bold transition sm:text-sm"
            :class="props.accountSource === 'ismart' ? 'bg-surface text-text shadow-sm' : 'text-text-muted hover:text-text'"
            @click="emit('set-account-source', 'ismart')"
          >
            {{ t('auth.existingIsmartAccount') }}
          </button>
        </div>

        <div
          v-show="props.emailAction === 'login'"
          class="login-method-tabs grid grid-cols-3 rounded-lg bg-surface-raised p-1"
        >
          <button
            type="button"
            class="rounded-md px-2 py-2 text-xs font-bold transition sm:text-sm"
            :class="props.authMode === 'phone' ? 'bg-surface text-text shadow-sm' : 'text-text-muted hover:text-text'"
            @click="selectPhonePasswordLogin"
          >
            {{ t('auth.phoneMode') }}
          </button>
          <button
            type="button"
            class="rounded-md px-2 py-2 text-xs font-bold transition sm:text-sm"
            :class="props.authMode === 'username' ? 'bg-surface text-text shadow-sm' : 'text-text-muted hover:text-text'"
            @click="emit('set-auth-mode', 'username')"
          >
            {{ t('auth.usernameMode') }}
          </button>
          <button
            type="button"
            class="rounded-md px-2 py-2 text-xs font-bold transition sm:text-sm"
            :class="props.authMode === 'email' ? 'bg-surface text-text shadow-sm' : 'text-text-muted hover:text-text'"
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
    clamp(4.25rem, 8svh, 5.5rem)
    clamp(2rem, 4vw, 3.5rem)
    clamp(1rem, 3svh, 2rem);
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
  margin-bottom: clamp(0.9rem, 2svh, 1.5rem);
  flex-shrink: 0;
}

.login-form-heading p {
  margin-top: clamp(0.55rem, 1.4svh, 0.75rem);
}

.login-form-fields {
  display: grid;
  gap: clamp(0.65rem, 1.55svh, 1rem);
  padding-bottom: 0.25rem;
}

.login-form-input {
  height: clamp(2.85rem, 6svh, 3.5rem);
  padding-top: 0.5rem;
  padding-bottom: 0.5rem;
}

.login-submit-button {
  min-height: clamp(2.85rem, 6svh, 3.5rem);
  padding-block: 0.65rem;
}

.login-form-footer {
  flex-shrink: 0;
  padding-top: clamp(0.75rem, 2svh, 1.5rem);
  font-size: clamp(0.82rem, 1.6svh, 1rem);
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
