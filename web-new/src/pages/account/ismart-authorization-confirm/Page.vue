<!--
 * iSmart 副戶授權確認頁。
 * 1. 電郵連結入口：以 token 讀取授權內容後確認。
 * 2. 站內通知入口：以 authorization_id 登入態確認。
 * 3. 確認成功後由後端呼叫 iSmart 授予副戶權限。
-->
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';

import AppIcon from '@/shared/components/base/AppIcon.vue';
import { useSessionStore } from '@/stores/session';
import {
  resolveIsmartSubaccountAuthorizationToken,
  confirmIsmartSubaccountAuthorizationByToken,
  confirmMemberIsmartSubaccountAuthorization,
  type IsmartSubaccountAuthorizationRow,
} from '@/httpapis/building';

const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const sessionStore = useSessionStore();

const token = computed(() => (typeof route.query.token === 'string' ? route.query.token.trim() : ''));
const authorizationID = computed(() => (
  typeof route.query.authorization_id === 'string' ? route.query.authorization_id.trim() : ''
));
const mode = computed<'email' | 'inapp' | 'invalid'>(() => {
  if (token.value) return 'email';
  if (authorizationID.value) return 'inapp';
  return 'invalid';
});

const loading = ref(true);
const submitting = ref(false);
const granted = ref(false);
const errorKey = ref('');
const detail = ref<IsmartSubaccountAuthorizationRow | null>(null);

// 1. 讀取授權內容；站內入口只確認登入狀態
const initialize = async (): Promise<void> => {
  loading.value = true;
  errorKey.value = '';
  if (!sessionStore.isLoaded) {
    await sessionStore.hydrateSession();
  }
  if (mode.value === 'invalid') {
    errorKey.value = 'building.ismartAuthorizationConfirm.invalid';
    loading.value = false;
    return;
  }
  if (mode.value === 'inapp' && !sessionStore.isAuthenticated) {
    await router.replace({ path: '/login', query: { redirect: route.fullPath } });
    return;
  }
  if (mode.value !== 'email') {
    loading.value = false;
    return;
  }
  try {
    const result = await resolveIsmartSubaccountAuthorizationToken(token.value);
    detail.value = result;
    if (result.status === 'active') {
      granted.value = true;
    } else if (result.status === 'expired' || result.status === 'revoked') {
      errorKey.value = 'building.ismartAuthorizationConfirm.expired';
    }
  } catch (error) {
    console.error(error);
    errorKey.value = 'building.ismartAuthorizationConfirm.invalid';
  } finally {
    loading.value = false;
  }
};

// 2. 確認接受授權
const confirm = async (): Promise<void> => {
  if (submitting.value || errorKey.value) return;
  submitting.value = true;
  try {
    const result = mode.value === 'email'
      ? await confirmIsmartSubaccountAuthorizationByToken(token.value)
      : await confirmMemberIsmartSubaccountAuthorization(authorizationID.value);
    granted.value = result.granted !== false;
  } catch (error) {
    console.error(error);
    errorKey.value = 'building.ismartAuthorizationConfirm.confirmFailed';
  } finally {
    submitting.value = false;
  }
};

const unitLabel = computed(() => (
  detail.value?.unit_display
  || detail.value?.building_name
  || detail.value?.unit_id
  || ''
));

const contactLabel = computed(() => (
  [detail.value?.phone, detail.value?.email].filter(Boolean).join(' · ')
));

const expiresLabel = computed(() => {
  const value = detail.value?.confirm_expires_at || detail.value?.invitation_expires_at;
  if (!value) return '';
  const parsed = new Date(value);
  return Number.isNaN(parsed.getTime()) ? '' : parsed.toLocaleString();
});

onMounted(() => {
  void initialize();
});
</script>

<template>
  <main class="ismart-confirm-page">
    <section class="ismart-confirm-card">
      <span class="ismart-confirm-badge"><AppIcon name="shield" :size="20" /></span>
      <h1>{{ t('building.ismartAuthorizationConfirm.title') }}</h1>

      <p v-if="loading" class="ismart-confirm-loading">{{ t('building.common.loading') }}</p>

      <template v-else-if="granted">
        <p class="ismart-confirm-success">
          {{ t('building.ismartAuthorizationConfirm.granted') }}
        </p>
        <div class="ismart-confirm-actions">
          <button
            v-if="sessionStore.isAuthenticated"
            type="button"
            class="work-action"
            @click="router.push('/building')"
          >
            <AppIcon name="building" :size="15" />
            <span>{{ t('building.ismartAuthorizationConfirm.goBuilding') }}</span>
          </button>
          <button
            v-else
            type="button"
            class="work-action"
            @click="router.push({ path: '/login', query: { redirect: '/building' } })"
          >
            <AppIcon name="user" :size="15" />
            <span>{{ t('building.ismartAuthorizationConfirm.goLogin') }}</span>
          </button>
        </div>
      </template>

      <template v-else-if="errorKey">
        <p class="ismart-confirm-error">{{ t(errorKey) }}</p>
        <div v-if="!sessionStore.isAuthenticated" class="ismart-confirm-actions">
          <button
            type="button"
            class="work-action"
            @click="router.push({ path: '/login', query: { redirect: route.fullPath } })"
          >
            <AppIcon name="user" :size="15" />
            <span>{{ t('building.ismartAuthorizationConfirm.goLogin') }}</span>
          </button>
        </div>
      </template>

      <template v-else>
        <p class="ismart-confirm-desc">
          {{ t(mode === 'email'
            ? 'building.ismartAuthorizationConfirm.descriptionEmail'
            : 'building.ismartAuthorizationConfirm.descriptionInApp') }}
        </p>
        <dl v-if="mode === 'email'" class="ismart-confirm-detail">
          <div v-if="unitLabel">
            <dt>{{ t('building.ismartAuthorizationConfirm.unitLabel') }}</dt>
            <dd>{{ unitLabel }}</dd>
          </div>
          <div v-if="contactLabel">
            <dt>{{ t('building.ismartAuthorizationConfirm.contactLabel') }}</dt>
            <dd>{{ contactLabel }}</dd>
          </div>
          <div v-if="expiresLabel">
            <dt>{{ t('building.ismartAuthorizationConfirm.expiresLabel') }}</dt>
            <dd>{{ expiresLabel }}</dd>
          </div>
        </dl>
        <div class="ismart-confirm-actions">
          <button type="button" class="work-action" :disabled="submitting" @click="confirm">
            <AppIcon name="shield" :size="15" />
            <span>{{ submitting ? t('building.ismartAuthorizationConfirm.confirming') : t('building.ismartAuthorizationConfirm.confirm') }}</span>
          </button>
        </div>
      </template>
    </section>
  </main>
</template>

<style scoped>
.ismart-confirm-page {
  display: grid;
  place-items: center;
  min-height: calc(100svh - var(--app-header-offset, 48px));
  padding: 24px;
  background: var(--sur);
}

.ismart-confirm-card {
  width: min(100%, 520px);
  border: 1px solid var(--bdr);
  border-radius: 10px;
  padding: 28px;
  background: var(--card, #fff);
  color: var(--ink);
}

.ismart-confirm-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: var(--sur-2);
  color: var(--brand);
}

.ismart-confirm-card h1 {
  margin: 14px 0 0;
  font-size: 20px;
  font-weight: 700;
}

.ismart-confirm-desc,
.ismart-confirm-loading {
  margin: 10px 0 0;
  color: var(--ink-2);
  font-size: 13px;
  line-height: 1.6;
}

.ismart-confirm-detail {
  display: grid;
  gap: 10px;
  margin: 18px 0 0;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  padding: 14px;
}

.ismart-confirm-detail div {
  display: flex;
  gap: 12px;
  font-size: 13px;
}

.ismart-confirm-detail dt {
  flex: 0 0 96px;
  color: var(--ink-3);
}

.ismart-confirm-detail dd {
  margin: 0;
  color: var(--ink);
  font-weight: 600;
  word-break: break-all;
}

.ismart-confirm-actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-top: 24px;
}

.ismart-confirm-actions .work-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 100%;
  border: 0;
  border-radius: 10px;
  background: var(--brand);
  color: #fff;
  cursor: pointer;
  font-family: inherit;
  font-size: 15px;
  font-weight: 700;
  padding: 14px 24px;
  white-space: nowrap;
  text-decoration: none;
  box-shadow: 0 2px 10px rgb(0 0 0 / 14%);
  transition: background-color 0.15s ease, transform 0.1s ease;
}

.ismart-confirm-actions .work-action:hover:not(:disabled) {
  background: var(--brand-dark);
}

.ismart-confirm-actions .work-action:active:not(:disabled) {
  transform: scale(0.99);
}

.ismart-confirm-actions .work-action:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.ismart-confirm-success {
  margin: 12px 0 0;
  color: var(--success);
  font-size: 13px;
  font-weight: 600;
}

.ismart-confirm-error {
  margin: 12px 0 0;
  border-left: 3px solid var(--danger);
  padding-left: 10px;
  color: var(--danger);
  font-size: 13px;
  line-height: 1.6;
}
</style>
