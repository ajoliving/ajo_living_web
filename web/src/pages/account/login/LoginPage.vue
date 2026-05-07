<!--
 * 登入頁主入口。
 * 1. 組裝登入頁左右版面。
 * 2. 將登入資料流程交由 login.ts 管理。
-->
<script setup lang="ts">
import { useLoginPage } from './login';
import LoginFormPanel from './widgets/LoginFormPanel.vue';
import LoginHero from './widgets/LoginHero.vue';

const {
  authMode,
  emailAction,
  emailActionSwitchLabel,
  emailLoginMethod,
  emailPlaceholder,
  footerActionLabel,
  footerPrompt,
  formState,
  handleRequestOtp,
  handleSignOut,
  handleSubmit,
  isAuthenticated,
  otpRequestLabel,
  rememberMe,
  requestingOtp,
  selectedHero,
  submitting,
  submitLabel,
  authModeSwitchIcon,
  authModeSwitchLabel,
  setEmailLoginMethod,
  toggleAuthMode,
  toggleEmailAction,
} = useLoginPage();
</script>

<template>
  <div class="login-page bg-surface">
    <section class="login-page__grid grid bg-surface lg:grid-cols-[1.8fr_1fr] xl:grid-cols-[1.95fr_1fr]">
      <LoginHero :hero="selectedHero" />

      <LoginFormPanel
        v-model:display-name="formState.displayName"
        v-model:email="formState.email"
        v-model:otp="formState.otp"
        v-model:password="formState.password"
        v-model:phone="formState.phone"
        v-model:remember-me="rememberMe"
        :auth-mode="authMode"
        :auth-mode-switch-icon="authModeSwitchIcon"
        :auth-mode-switch-label="authModeSwitchLabel"
        :email-action="emailAction"
        :email-action-switch-label="emailActionSwitchLabel"
        :email-login-method="emailLoginMethod"
        :email-placeholder="emailPlaceholder"
        :footer-action-label="footerActionLabel"
        :footer-prompt="footerPrompt"
        :is-authenticated="isAuthenticated"
        :otp-request-label="otpRequestLabel"
        :requesting-otp="requestingOtp"
        :submitting="submitting"
        :submit-label="submitLabel"
        @request-otp="handleRequestOtp"
        @set-email-login-method="setEmailLoginMethod"
        @sign-out="handleSignOut"
        @submit-login="handleSubmit"
        @toggle-auth-mode="toggleAuthMode"
        @toggle-email-action="toggleEmailAction"
      />
    </section>
  </div>
</template>

<style scoped>
.login-page {
  width: 100%;
  height: calc(100svh - var(--app-header-offset, 0rem));
  min-height: 0;
  overflow: hidden;
}

.login-page__grid {
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
}
</style>
