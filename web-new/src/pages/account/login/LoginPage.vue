<!--
 * 登入頁主入口。
 * 1. 組裝登入主視覺與住戶登入/註冊表單。
 * 2. 對接真實認證接口、住戶註冊大廈與單位選項。
 * 3. 維持單一路由 page 負責登入頁資料流。
-->
<script setup lang="ts">
import LoginHero from './widgets/LoginHero.vue';
import LoginFormPanel from './widgets/LoginFormPanel.vue';
import { useLoginPage } from './login';

const {
  authMode,
  buildingOptions,
  buildingsLoading,
  emailAction,
  emailActionSwitchLabel,
  emailPlaceholder,
  footerPrompt,
  formState,
  handleForgotPassword,
  handleSignOut,
  handleSubmit,
  isAuthenticated,
  rememberMe,
  publisherIdentityOptions,
  residenceFloorOptions,
  residenceUnitOptions,
  selectedHero,
  setAuthMode,
  submitting,
  submitLabel,
  toggleEmailAction,
  unitsLoading,
} = useLoginPage();
</script>

<template>
  <main class="login-page">
    <section class="login-shell">
      <LoginHero :hero="selectedHero" />
      <LoginFormPanel
        v-model:display-name="formState.displayName"
        v-model:email="formState.email"
        v-model:password="formState.password"
        v-model:phone="formState.phone"
        v-model:phone-country-code="formState.phoneCountryCode"
        v-model:publisher-identity-type="formState.publisherIdentityType"
        v-model:ismart-account="formState.ismartAccount"
        v-model:primary-community-id="formState.primaryCommunityID"
        v-model:residence-floor="formState.residenceFloor"
        v-model:residence-unit="formState.residenceUnit"
        v-model:remember-me="rememberMe"
        :auth-mode="authMode"
        :email-action="emailAction"
        :email-action-switch-label="emailActionSwitchLabel"
        :building-options="buildingOptions"
        :buildings-loading="buildingsLoading"
        :residence-floor-options="residenceFloorOptions"
        :residence-unit-options="residenceUnitOptions"
        :submitting="submitting"
        :units-loading="unitsLoading"
        :email-placeholder="emailPlaceholder"
        :footer-prompt="footerPrompt"
        :is-authenticated="isAuthenticated"
        :publisher-identity-options="publisherIdentityOptions"
        :submit-label="submitLabel"
        @forgot-password="handleForgotPassword"
        @set-auth-mode="setAuthMode"
        @sign-out="handleSignOut"
        @submit-login="handleSubmit"
        @toggle-email-action="toggleEmailAction"
      />
    </section>
  </main>
</template>

<style scoped>
.login-page {
  min-height: calc(100vh - 48px);
  background: rgb(var(--color-surface));
}

.login-shell {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 520px;
  align-items: stretch;
  min-width: 0;
  min-height: calc(100vh - 48px);
}

@media (max-width: 1023px) {
  .login-shell {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
