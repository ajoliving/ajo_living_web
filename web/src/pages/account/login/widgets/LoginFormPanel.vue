<!--
 * 登入頁表單面板。
 * 1. 展示郵箱密碼登入、手機密碼登入、郵箱驗證碼登入與住戶註冊表單。
 * 2. 將表單輸入與操作事件回傳給頁面入口。
-->
<script setup lang="ts">
import { useI18n } from 'vue-i18n';

import AppIcon from '@/shared/components/base/AppIcon.vue';

import type { LoginAuthMode, LoginEmailAction, LoginEmailMethod } from '../login';

interface LoginFormPanelProps {
  authMode: LoginAuthMode;
  emailAction: LoginEmailAction;
  emailActionSwitchLabel: string;
  emailLoginMethod: LoginEmailMethod;
  displayName: string;
  email: string;
  password: string;
  phone: string;
  otp: string;
  rememberMe: boolean;
  requestingOtp: boolean;
  submitting: boolean;
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
  'update:otp': [value: string];
  'update:rememberMe': [value: boolean];
  'request-otp': [];
  'set-auth-mode': [value: LoginAuthMode];
  'set-email-login-method': [value: LoginEmailMethod];
  'sign-out': [];
  'submit-login': [];
  'toggle-email-action': [];
}>();

const { t } = useI18n();

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
</script>

<template>
  <div class="login-form-panel z-20 flex h-full min-h-0 flex-col justify-between overflow-hidden bg-surface shadow-2xl">
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
          v-show="props.emailAction === 'login'"
          class="login-method-tabs grid grid-cols-3 rounded-lg bg-surface-raised p-1"
        >
          <button
            type="button"
            class="rounded-md px-2 py-2 text-xs font-bold transition sm:text-sm"
            :class="props.authMode === 'email' && props.emailLoginMethod === 'password' ? 'bg-surface text-text shadow-sm' : 'text-text-muted hover:text-text'"
            @click="selectEmailLoginMethod('password')"
          >
            {{ t('auth.emailPasswordLogin') }}
          </button>
          <button
            type="button"
            class="rounded-md px-2 py-2 text-xs font-bold transition sm:text-sm"
            :class="props.authMode === 'phone' ? 'bg-surface text-text shadow-sm' : 'text-text-muted hover:text-text'"
            @click="selectPhonePasswordLogin"
          >
            {{ t('auth.phonePasswordLogin') }}
          </button>
          <button
            type="button"
            class="rounded-md px-2 py-2 text-xs font-bold transition sm:text-sm"
            :class="props.authMode === 'email' && props.emailLoginMethod === 'code' ? 'bg-surface text-text shadow-sm' : 'text-text-muted hover:text-text'"
            @click="selectEmailLoginMethod('code')"
          >
            {{ t('auth.emailCodeLogin') }}
          </button>
        </div>

        <template v-if="props.emailAction === 'register'">
          <label class="block">
            <span class="mb-1.5 block text-sm font-bold uppercase tracking-[0.16em] text-text">{{ t('auth.displayName') }}</span>
            <div class="group relative">
              <input
                :value="props.displayName"
                class="login-form-input w-full rounded-lg border-none bg-surface-raised px-4 pr-12 text-text outline-none transition placeholder:text-text-muted/50 focus:ring-1 focus:ring-primary"
                :placeholder="t('auth.displayNamePlaceholder')"
                autocomplete="name"
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
                type="password"
                @input="emit('update:password', readInputValue($event))"
              >
              <AppIcon
                name="lock"
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
                type="password"
                @input="emit('update:password', readInputValue($event))"
              >
              <AppIcon
                name="lock"
                :size="20"
                class="absolute right-4 top-1/2 -translate-y-1/2 text-text-muted transition group-focus-within:text-primary"
              />
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
                type="password"
                @input="emit('update:password', readInputValue($event))"
              >
              <AppIcon
                name="lock"
                :size="20"
                class="absolute right-4 top-1/2 -translate-y-1/2 text-text-muted transition group-focus-within:text-primary"
              />
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

    <div class="login-form-footer border-t border-border/70 text-center text-sm leading-6 text-text-muted">
      <span>{{ props.footerPrompt }}</span>
      <button
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
}

.login-form-heading {
  margin-bottom: clamp(0.9rem, 2svh, 1.5rem);
}

.login-form-heading p {
  margin-top: clamp(0.55rem, 1.4svh, 0.75rem);
}

.login-form-fields {
  display: grid;
  gap: clamp(0.65rem, 1.55svh, 1rem);
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
