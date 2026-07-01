<!--
 * 忘記密碼頁。
 * 1. 使用電郵驗證碼完成密碼重設。
 * 2. 保持手機重設入口暫不開放，避免未完成短信流程曝光。
 * 3. 重設成功後返回登入頁。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { RouterLink, useRouter } from 'vue-router';

import AppIcon from '@/shared/components/base/AppIcon.vue';
import { useFeedbackStore } from '@/stores/feedback';
import { useSessionStore } from '@/stores/session';

interface ForgotPasswordFormState {
  email: string;
  code: string;
  password: string;
  confirmPassword: string;
}

const router = useRouter();
const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const sessionStore = useSessionStore();
const EMAIL_PLACEHOLDER = 'resident@example.com';

const formState = reactive<ForgotPasswordFormState>({
  email: '',
  code: '',
  password: '',
  confirmPassword: '',
});
const requestingCode = ref(false);
const resettingPassword = ref(false);
const codeSent = ref(false);
const showPassword = ref(false);
const showConfirmPassword = ref(false);

const canSubmit = computed(() =>
  formState.email.trim().length > 0 &&
  formState.code.trim().length > 0 &&
  formState.password.trim().length >= 8 &&
  formState.confirmPassword.trim().length >= 8,
);

// 1. 讀取輸入值
const readInputValue = (event: Event): string => (event.target as HTMLInputElement).value;

// 2. 驗證電郵格式
const isValidEmailInput = (email: string): boolean => {
  const value = email.trim();
  return value.includes('@') && value.includes('.') && value.length <= 255;
};

// 3. 解析接口錯誤訊息
const readErrorMessage = (error: unknown): string =>
  axios.isAxiosError(error)
    ? error.response?.data?.message ?? error.message
    : error instanceof Error ? error.message : t('auth.passwordResetFailed');

// 4. 請求電郵驗證碼
const handleRequestCode = async (): Promise<void> => {
  if (!isValidEmailInput(formState.email)) {
    feedbackStore.pushToast(t('auth.invalidEmail'), 'error');
    return;
  }

  requestingCode.value = true;
  try {
    const result = await sessionStore.sendPasswordResetEmail(formState.email.trim());
    codeSent.value = true;
    feedbackStore.pushToast(t('auth.passwordResetCodeSent'), 'success');
    if (result.mock_code) {
      feedbackStore.pushToast(t('auth.otpPreview', { code: result.mock_code }), 'info');
    }
  } catch (error) {
    feedbackStore.pushToast(readErrorMessage(error), 'error');
  } finally {
    requestingCode.value = false;
  }
};

// 5. 提交新密碼
const handleResetPassword = async (): Promise<void> => {
  if (!isValidEmailInput(formState.email)) {
    feedbackStore.pushToast(t('auth.invalidEmail'), 'error');
    return;
  }
  if (formState.code.trim().length === 0) {
    feedbackStore.pushToast(t('auth.passwordResetCodeRequired'), 'error');
    return;
  }
  if (formState.password.trim().length < 8) {
    feedbackStore.pushToast(t('auth.passwordResetPasswordRequired'), 'error');
    return;
  }
  if (formState.password !== formState.confirmPassword) {
    feedbackStore.pushToast(t('auth.passwordResetConfirmMismatch'), 'error');
    return;
  }

  resettingPassword.value = true;
  try {
    await sessionStore.resetPasswordByEmail(
      formState.email.trim(),
      formState.code.trim(),
      formState.password,
    );
    feedbackStore.pushToast(t('auth.passwordResetSuccess'), 'success');
    await router.push('/login');
  } catch (error) {
    feedbackStore.pushToast(readErrorMessage(error), 'error');
  } finally {
    resettingPassword.value = false;
  }
};
</script>

<template>
  <main class="forgot-password-page">
    <section class="forgot-password-shell">
      <div class="forgot-password-brand">
        <p>AJO Living</p>
        <span />
      </div>

      <form
        class="forgot-password-panel"
        @submit.prevent="handleResetPassword"
      >
        <div class="forgot-password-heading">
          <p>{{ t('auth.passwordResetEyebrow') }}</p>
          <h1>{{ t('auth.passwordResetTitle') }}</h1>
          <span>{{ t('auth.passwordResetDescription') }}</span>
        </div>

        <label class="forgot-password-field">
          <span>{{ t('auth.email') }}</span>
          <div class="forgot-password-control">
            <input
              :value="formState.email"
              autocapitalize="none"
              autocomplete="email"
              inputmode="email"
              spellcheck="false"
              type="email"
              :placeholder="EMAIL_PLACEHOLDER"
              @input="formState.email = readInputValue($event)"
            >
            <AppIcon
              name="user"
              :size="20"
              class="forgot-password-icon"
            />
          </div>
        </label>

        <button
          class="forgot-password-secondary-button"
          type="button"
          :disabled="requestingCode"
          @click="handleRequestCode"
        >
          {{ requestingCode ? t('auth.sending') : codeSent ? t('auth.passwordResetResendCode') : t('auth.passwordResetSendCode') }}
        </button>

        <label class="forgot-password-field">
          <span>{{ t('auth.emailOtp') }}</span>
          <div class="forgot-password-control">
            <input
              :value="formState.code"
              autocomplete="one-time-code"
              inputmode="numeric"
              type="text"
              :placeholder="t('auth.emailOtpPlaceholder')"
              @input="formState.code = readInputValue($event)"
            >
            <AppIcon
              name="shield"
              :size="20"
              class="forgot-password-icon"
            />
          </div>
        </label>

        <label class="forgot-password-field">
          <span>{{ t('auth.passwordResetNewPassword') }}</span>
          <div class="forgot-password-control">
            <input
              :value="formState.password"
              autocomplete="new-password"
              :type="showPassword ? 'text' : 'password'"
              :placeholder="t('auth.passwordPlaceholder')"
              @input="formState.password = readInputValue($event)"
            >
            <button
              type="button"
              class="forgot-password-icon-button"
              :aria-label="showPassword ? t('auth.hidePassword') : t('auth.showPassword')"
              @click="showPassword = !showPassword"
            >
              <AppIcon
                :name="showPassword ? 'view-off' : 'view'"
                :size="18"
              />
            </button>
          </div>
        </label>

        <label class="forgot-password-field">
          <span>{{ t('auth.passwordResetConfirmPassword') }}</span>
          <div class="forgot-password-control">
            <input
              :value="formState.confirmPassword"
              autocomplete="new-password"
              :type="showConfirmPassword ? 'text' : 'password'"
              :placeholder="t('auth.passwordResetConfirmPlaceholder')"
              @input="formState.confirmPassword = readInputValue($event)"
            >
            <button
              type="button"
              class="forgot-password-icon-button"
              :aria-label="showConfirmPassword ? t('auth.hidePassword') : t('auth.showPassword')"
              @click="showConfirmPassword = !showConfirmPassword"
            >
              <AppIcon
                :name="showConfirmPassword ? 'view-off' : 'view'"
                :size="18"
              />
            </button>
          </div>
        </label>

        <button
          class="forgot-password-submit"
          type="submit"
          :disabled="resettingPassword || !canSubmit"
        >
          {{ resettingPassword ? t('auth.loading') : t('auth.passwordResetSubmit') }}
        </button>

        <RouterLink
          class="forgot-password-back"
          to="/login"
        >
          {{ t('auth.passwordResetBackToLogin') }}
        </RouterLink>
      </form>
    </section>
  </main>
</template>

<style scoped>
.forgot-password-page {
  min-height: calc(100vh - 48px);
  background: rgb(var(--color-surface));
}

.forgot-password-shell {
  display: grid;
  min-height: calc(100vh - 48px);
  min-width: 0;
  place-items: center;
  padding: clamp(2rem, 8vw, 4.5rem) clamp(1.5rem, 6vw, 3rem);
}

.forgot-password-brand {
  position: absolute;
  top: clamp(4.5rem, 8vw, 6rem);
  left: clamp(1.5rem, 6vw, 4rem);
  display: grid;
  gap: 0.45rem;
}

.forgot-password-brand p {
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: 18px;
  letter-spacing: 0.22em;
  text-transform: uppercase;
}

.forgot-password-brand span {
  width: 44px;
  height: 2px;
  background: rgb(var(--color-primary));
}

.forgot-password-panel {
  display: grid;
  width: min(100%, 440px);
  min-width: 0;
  gap: 0.85rem;
}

.forgot-password-heading {
  margin-bottom: 0.3rem;
  text-align: left;
}

.forgot-password-heading p {
  margin-bottom: 0.55rem;
  color: rgb(var(--color-text-muted));
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.forgot-password-heading h1 {
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: 30px;
  font-weight: 400;
  line-height: 1.2;
}

.forgot-password-heading span {
  display: block;
  margin-top: 0.7rem;
  color: rgb(var(--color-text-muted));
  font-size: 13px;
  line-height: 1.7;
}

.forgot-password-field {
  display: grid;
  min-width: 0;
  gap: 0.4rem;
}

.forgot-password-field > span {
  color: rgb(var(--color-text-muted));
  font-size: 10px;
  font-weight: 500;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.forgot-password-control {
  position: relative;
  min-width: 0;
}

.forgot-password-control input {
  width: 100%;
  height: 42px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  font-size: 13px;
  outline: none;
  padding: 8px 48px 8px 16px;
  transition: border-color 160ms ease, box-shadow 160ms ease;
}

.forgot-password-control input:focus {
  border-color: rgb(var(--color-primary));
  box-shadow: 0 0 0 1px rgb(var(--color-primary) / 0.16);
}

.forgot-password-icon,
.forgot-password-icon-button {
  position: absolute;
  top: 50%;
  right: 14px;
  transform: translateY(-50%);
}

.forgot-password-icon {
  pointer-events: none;
  color: rgb(var(--color-text-muted));
}

.forgot-password-icon-button {
  display: inline-flex;
  width: 32px;
  height: 32px;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  color: rgb(var(--color-text-muted));
  transition: background-color 160ms ease, color 160ms ease;
}

.forgot-password-icon-button:hover {
  background: rgb(var(--color-surface-muted));
  color: rgb(var(--color-primary));
}

.forgot-password-secondary-button,
.forgot-password-submit {
  min-height: 42px;
  border-radius: 2px;
  font-size: 12px;
  font-weight: 700;
  transition: background-color 160ms ease, border-color 160ms ease, color 160ms ease, opacity 160ms ease;
}

.forgot-password-secondary-button {
  border: 1px solid rgb(var(--color-border));
  color: rgb(var(--color-text));
}

.forgot-password-secondary-button:hover {
  border-color: rgb(var(--color-primary));
  color: rgb(var(--color-primary));
}

.forgot-password-submit {
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
  letter-spacing: 0.2em;
  text-transform: uppercase;
}

.forgot-password-submit:hover {
  background: rgb(var(--color-text));
}

.forgot-password-secondary-button:disabled,
.forgot-password-submit:disabled {
  cursor: not-allowed;
  opacity: 0.58;
}

.forgot-password-back {
  justify-self: center;
  color: rgb(var(--color-primary));
  font-size: 13px;
  font-weight: 700;
}

.forgot-password-back:hover {
  color: rgb(var(--color-text));
  text-decoration: underline;
}

@media (max-width: 640px) {
  .forgot-password-shell {
    align-items: start;
    padding-top: 7rem;
  }

  .forgot-password-brand {
    top: 4.25rem;
    left: 1.5rem;
  }

  .forgot-password-control input {
    font-size: 16px;
  }
}
</style>
