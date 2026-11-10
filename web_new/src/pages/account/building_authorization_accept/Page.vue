<!--
 * 大廈住戶授權啟用頁。
 * 1. 讀取電郵連結中的一次性 token。
 * 2. 由受邀住戶設定用戶 ID 及密碼。
-->
<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';

import { acceptBuildingAuthorization } from '@/httpapis/building';

const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const token = computed(() => String(route.query.token ?? '').trim());
const username = ref('');
const password = ref('');
const confirmingPassword = ref('');
const loading = ref(false);
const error = ref('');
const success = ref(false);

// 1. 提交啟用資料
const submit = async (): Promise<void> => {
  error.value = '';
  if (!token.value || !username.value.trim() || password.value.length < 8 || password.value !== confirmingPassword.value) {
    error.value = t('building.authorizationAccept.validation');
    return;
  }
  loading.value = true;
  try {
    await acceptBuildingAuthorization({ token: token.value, username: username.value.trim(), password: password.value });
    success.value = true;
  } catch {
    error.value = t('building.authorizationAccept.error');
  } finally {
    loading.value = false;
  }
};
</script>

<template>
  <main class="authorization-accept-page">
    <section class="authorization-accept-card">
      <h1>{{ t('building.authorizationAccept.title') }}</h1>
      <p>{{ t('building.authorizationAccept.description') }}</p>
      <div v-if="success" class="authorization-accept-success">
        <p>{{ t('building.authorizationAccept.success') }}</p>
        <button type="button" class="work-action" @click="router.push('/login')">{{ t('building.authorizationAccept.login') }}</button>
      </div>
      <form v-else @submit.prevent="submit">
        <label><span>{{ t('building.authorizationAccept.username') }}</span><input v-model="username" autocomplete="username"></label>
        <label><span>{{ t('building.authorizationAccept.password') }}</span><input v-model="password" type="password" autocomplete="new-password"></label>
        <label><span>{{ t('building.authorizationAccept.confirmPassword') }}</span><input v-model="confirmingPassword" type="password" autocomplete="new-password"></label>
        <p v-if="error" class="authorization-accept-error">{{ error }}</p>
        <button type="submit" class="work-action" :disabled="loading">{{ loading ? t('building.authorizationAccept.submitting') : t('building.authorizationAccept.submit') }}</button>
      </form>
    </section>
  </main>
</template>

<style scoped>
.authorization-accept-page { min-height: calc(100svh - var(--nav-h, 52px)); display: grid; place-items: center; padding: 24px; background: var(--sur); }
.authorization-accept-card { width: min(100%, 520px); padding: 28px; border: 1px solid var(--bdr); border-radius: 8px; background: #fff; color: var(--ink); }
.authorization-accept-card form { display: grid; gap: 16px; }
.authorization-accept-card label { display: grid; gap: 6px; }
.authorization-accept-card input { min-height: 42px; border: 1px solid var(--bdr); border-radius: 4px; padding: 8px 10px; }
.authorization-accept-error { color: var(--danger); }
.authorization-accept-success { display: grid; gap: 16px; }
</style>
