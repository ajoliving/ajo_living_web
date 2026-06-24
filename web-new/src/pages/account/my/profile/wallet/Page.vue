<!--
 * 會員錢包頁。
 * 1. 顯示 AJO Point 餘額、廣告獎勵與最近流水。
 * 2. 提供後台廣告任務彈窗觀看、倒數與積分自動領取流程。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import {
  claimRewardAdTask,
  fetchRewardAdTasks,
  fetchWalletOverview,
  startRewardAdTask,
  trackRewardAdClick,
} from '@/httpapis/wallet';
import type {
  RewardAdSessionResponse,
  RewardAdTaskResponse,
  WalletOverviewResponse,
} from '@/model/wallet';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';
import { useSessionStore } from '@/stores/session';
import { formatDate } from '@/utils/format';
import { formatAjoPoints } from '@/utils/wallet';

interface RunningAdSession {
  taskId: string;
  claimId: string;
  remainingSeconds: number;
}

const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const preferenceStore = usePreferenceStore();
const sessionStore = useSessionStore();

const loading = ref(false);
const claiming = ref(false);
const overview = ref<WalletOverviewResponse | null>(null);
const adTasks = ref<RewardAdTaskResponse[]>([]);
const runningSession = ref<RunningAdSession | null>(null);
const selectedAdTask = ref<RewardAdTaskResponse | null>(null);
const showingExitConfirm = ref(false);
const adClaimSucceeded = ref(false);
const startingAd = ref(false);
let countdownTimer: number | undefined;

const account = computed(() => overview.value?.account);
const transactions = computed(() => overview.value?.recent_transactions ?? []);
const dailyProgress = computed(() => {
  if (!account.value || account.value.daily_ad_reward_limit <= 0) {
    return 0;
  }
  return Math.min(100, Math.round((account.value.today_ad_reward_points / account.value.daily_ad_reward_limit) * 100));
});
const pointFormatter = new Intl.NumberFormat('zh-HK', { maximumFractionDigits: 0 });
const selectedAdRemainingSeconds = computed(() => {
  const session = runningSession.value;
  if (!session || session.taskId !== selectedAdTask.value?.task_id) {
    return 0;
  }
  return session.remainingSeconds;
});
const selectedAdIsRunning = computed(() => Boolean(
  runningSession.value && runningSession.value.taskId === selectedAdTask.value?.task_id,
));

// 1. 格式化錢包主數字
const formatPointNumber = (value: number): string => pointFormatter.format(value);
const formatPoints = (value: number): string =>
  formatAjoPoints(value, t('common.brand.pointsName'), preferenceStore.locale);

// 2. 讀取錯誤訊息
const readErrorMessage = (error: unknown, fallback: string): string =>
  axios.isAxiosError<{ message?: string }>(error)
    ? error.response?.data?.message ?? fallback
    : fallback;

// 3. 讀取錢包和廣告任務
const loadWallet = async (): Promise<void> => {
  loading.value = true;
  try {
    const [walletResponse, adResponse] = await Promise.all([
      fetchWalletOverview(),
      fetchRewardAdTasks(),
    ]);
    overview.value = walletResponse.data.data;
    adTasks.value = adResponse.data.data.items;
    await sessionStore.loadCurrentUser();
  } catch (error: unknown) {
    feedbackStore.pushToast(readErrorMessage(error, t('account.wallet.loadError')), 'error');
  } finally {
    loading.value = false;
  }
};

// 4. 停止倒數計時
const stopCountdown = (): void => {
  if (countdownTimer) {
    window.clearInterval(countdownTimer);
    countdownTimer = undefined;
  }
};

// 5. 開始倒數計時
const startCountdown = (session: RewardAdSessionResponse): void => {
  stopCountdown();
  const availableAtTime = Date.parse(session.available_at);
  const fallbackAvailableAtTime = Date.now() + session.watch_seconds * 1000;
  const targetAvailableAtTime = Number.isFinite(availableAtTime) ? availableAtTime : fallbackAvailableAtTime;
  const calculateRemainingSeconds = (): number =>
    Math.max(0, Math.ceil((targetAvailableAtTime - Date.now()) / 1000));

  runningSession.value = {
    taskId: session.task_id,
    claimId: session.claim_id,
    remainingSeconds: calculateRemainingSeconds(),
  };
  countdownTimer = window.setInterval(() => {
    if (!runningSession.value) {
      stopCountdown();
      return;
    }
    runningSession.value.remainingSeconds = calculateRemainingSeconds();
    if (runningSession.value.remainingSeconds <= 0) {
      stopCountdown();
      void handleClaimSelectedAd();
    }
  }, 500);

  if (runningSession.value.remainingSeconds <= 0) {
    stopCountdown();
    void handleClaimSelectedAd();
  }
};

// 6. 開始觀看廣告
const handleStartSelectedAd = async (): Promise<void> => {
  const task = selectedAdTask.value;
  if (!task || startingAd.value || claiming.value) {
    return;
  }

  startingAd.value = true;
  try {
    const response = await startRewardAdTask(task.task_id);
    startCountdown(response.data.data);
  } catch (error: unknown) {
    feedbackStore.pushToast(readErrorMessage(error, t('account.wallet.adStartError')), 'error');
  } finally {
    startingAd.value = false;
  }
};

// 7. 打開並立即開始廣告
const openAndStartAd = async (task: RewardAdTaskResponse): Promise<void> => {
  selectedAdTask.value = task;
  showingExitConfirm.value = false;
  adClaimSucceeded.value = false;
  runningSession.value = null;
  await handleStartSelectedAd();
};

// 8. 領取廣告積分
const handleClaimSelectedAd = async (): Promise<void> => {
  const task = selectedAdTask.value;
  if (!task) {
    return;
  }
  if (!runningSession.value || runningSession.value.taskId !== task.task_id || runningSession.value.remainingSeconds > 0) {
    return;
  }
  if (claiming.value) {
    return;
  }

  claiming.value = true;
  try {
    await claimRewardAdTask(task.task_id, runningSession.value.claimId);
    feedbackStore.pushToast(t('account.wallet.adClaimSuccess'), 'success');
    runningSession.value = null;
    adClaimSucceeded.value = true;
    await loadWallet();
  } catch (error: unknown) {
    feedbackStore.pushToast(readErrorMessage(error, t('account.wallet.adClaimError')), 'error');
  } finally {
    claiming.value = false;
  }
};

// 9. 記錄並打開廣告連結
const openSelectedAdLink = (): void => {
  const task = selectedAdTask.value;
  if (!task?.target_url) {
    return;
  }

  window.open(task.target_url, '_blank', 'noopener,noreferrer');
  void trackRewardAdClick(task.task_id)
    .then(() => loadWallet())
    .catch((error: unknown) => {
      feedbackStore.pushToast(readErrorMessage(error, t('account.wallet.adClickTrackError')), 'error');
    });
};

// 10. 輸出廣告按鈕文案
const resolveAdActionLabel = (task: RewardAdTaskResponse): string => {
  if (task.claimed_today) {
    return t('account.wallet.claimedToday');
  }
  if (!task.can_claim_today) {
    return t('account.wallet.limitReached');
  }
  return t('account.wallet.startAd');
};

// 11. 判斷廣告按鈕是否停用
const isAdActionDisabled = (task: RewardAdTaskResponse): boolean =>
  claiming.value ||
  startingAd.value ||
  task.claimed_today ||
  !task.can_claim_today ||
  runningSession.value !== null;

// 12. 處理廣告主動作
const runAdAction = async (task: RewardAdTaskResponse): Promise<void> => {
  if (isAdActionDisabled(task)) {
    return;
  }
  await openAndStartAd(task);
};

// 13. 要求關閉廣告彈窗
const requestCloseAdDialog = (): void => {
  if (claiming.value || startingAd.value) {
    return;
  }
  if (runningSession.value && runningSession.value.remainingSeconds > 0 && !adClaimSucceeded.value) {
    showingExitConfirm.value = true;
    return;
  }
  closeAdDialog();
};

// 14. 關閉廣告彈窗
const closeAdDialog = (): void => {
  stopCountdown();
  runningSession.value = null;
  selectedAdTask.value = null;
  showingExitConfirm.value = false;
  adClaimSucceeded.value = false;
};

// 15. 確認未完成觀看時離開
const confirmExitAdDialog = (): void => {
  closeAdDialog();
};

onMounted(() => {
  void loadWallet();
});

onBeforeUnmount(() => {
  stopCountdown();
});
</script>

<template>
  <main class="wallet-page">
    <section class="wallet-heading">
      <div>
        <p class="wallet-kicker">{{ t('common.brand.pointsName') }}</p>
        <h1>{{ t('account.wallet.title') }}</h1>
        <p>{{ t('account.wallet.description') }}</p>
      </div>
    </section>

    <section
      v-if="loading && !overview"
      class="wallet-panel"
    >
      {{ t('common.status.loading') }}
    </section>

    <template v-else>
      <section class="wallet-summary-grid">
        <article class="wallet-balance-card wallet-balance-card--primary">
          <strong class="wallet-point-value wallet-point-value-main">
            <span>{{ formatPointNumber(account?.balance ?? 0) }}</span>
            <small>{{ t('common.brand.pointsName') }}</small>
          </strong>
          <span class="wallet-balance-label">{{ t('account.wallet.balance') }}</span>
        </article>
        <article class="wallet-stat-card">
          <span>{{ t('account.wallet.earned') }}</span>
          <strong class="wallet-point-value">
            <span>{{ formatPointNumber(account?.total_earned ?? 0) }}</span>
            <small>{{ t('common.brand.pointsName') }}</small>
          </strong>
        </article>
        <article class="wallet-stat-card">
          <span>{{ t('account.wallet.spent') }}</span>
          <strong class="wallet-point-value">
            <span>{{ formatPointNumber(account?.total_spent ?? 0) }}</span>
            <small>{{ t('common.brand.pointsName') }}</small>
          </strong>
        </article>
        <article class="wallet-stat-card">
          <span>{{ t('account.wallet.todayAd') }}</span>
          <strong class="wallet-point-value">
            <span>{{ formatPointNumber(account?.today_ad_reward_points ?? 0) }}</span>
            <small>{{ t('common.brand.pointsName') }}</small>
          </strong>
        </article>
      </section>

      <section class="wallet-grid">
        <article class="wallet-panel wallet-ad-panel">
          <div class="wallet-panel-title">
            <div>
              <h2>{{ t('account.wallet.rewardAds') }}</h2>
              <p>
                {{ t('account.wallet.todayAd') }}
                {{ formatPoints(account?.today_ad_reward_points ?? 0) }}
                / {{ t('account.wallet.dailyLimit') }}
                {{ formatPoints(account?.daily_ad_reward_limit ?? 0) }}
              </p>
            </div>
            <div class="wallet-panel-meter">
              <strong>{{ dailyProgress }}%</strong>
              <div class="wallet-progress">
                <span :style="{ width: `${dailyProgress}%` }" />
              </div>
            </div>
          </div>

          <div
            v-if="adTasks.length === 0"
            class="wallet-empty"
          >
            {{ t('account.wallet.noAds') }}
          </div>

          <div
            v-else
            class="wallet-ad-list"
          >
            <article
              v-for="task in adTasks"
              :key="task.task_id"
              class="wallet-ad-card"
            >
              <div class="wallet-ad-cover">
                <video
                  v-if="task.media_type === 'video' && task.media_url"
                  :src="task.media_url"
                  preload="metadata"
                />
                <img
                  v-else-if="task.media_url"
                  :src="task.media_url"
                  :alt="task.title"
                />
                <AppIcon
                  v-else
                  name="view"
                  :size="34"
                />
              </div>
              <div class="wallet-ad-body">
                <h3>{{ task.title }}</h3>
                <p>{{ task.summary }}</p>
                <span>{{ t('account.wallet.rewardPoints', { points: formatPoints(task.reward_points) }) }}</span>
              </div>
              <button
                type="button"
                class="wallet-action-button"
                :disabled="isAdActionDisabled(task)"
                @click="runAdAction(task)"
              >
                {{ resolveAdActionLabel(task) }}
              </button>
            </article>
          </div>
        </article>

        <aside class="wallet-side-stack">
          <article class="wallet-panel">
            <h2>{{ t('account.wallet.rechargeTitle') }}</h2>
            <p class="wallet-muted">{{ t('account.wallet.rechargeUnavailable') }}</p>
          </article>
        </aside>
      </section>

      <section class="wallet-panel">
        <div class="wallet-panel-title">
          <h2>{{ t('account.wallet.transactions') }}</h2>
        </div>
        <div
          v-if="transactions.length === 0"
          class="wallet-empty"
        >
          {{ t('account.wallet.noTransactions') }}
        </div>
        <div
          v-else
          class="wallet-transaction-table-wrap"
        >
          <table class="wallet-transaction-table">
            <thead>
              <tr>
                <th>{{ t('account.wallet.type') }}</th>
                <th>{{ t('account.wallet.source') }}</th>
                <th>{{ t('account.wallet.amount') }}</th>
                <th>{{ t('account.wallet.date') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="transaction in transactions"
                :key="transaction.transaction_id"
              >
                <td>{{ transaction.direction === 'credit' ? t('account.wallet.credit') : t('account.wallet.debit') }}</td>
                <td>{{ transaction.biz_module }} · {{ transaction.action_type }}</td>
                <td>{{ formatPoints(transaction.amount) }}</td>
                <td>{{ formatDate(transaction.created_at) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </template>

    <Teleport to="body">
      <Transition name="wallet-ad-dialog-fade">
        <div
          v-if="selectedAdTask"
          class="wallet-ad-dialog"
          role="dialog"
          aria-modal="true"
          :aria-label="selectedAdTask.title"
          @click.self="requestCloseAdDialog"
        >
          <div class="wallet-ad-dialog__panel">
            <header class="wallet-ad-dialog__header">
              <div>
                <p class="wallet-kicker">{{ t('account.wallet.rewardAds') }}</p>
                <h2>{{ selectedAdTask.title }}</h2>
                <p>{{ selectedAdTask.summary }}</p>
              </div>
              <div class="wallet-ad-dialog__header-actions">
                <span
                  v-if="selectedAdIsRunning"
                  class="wallet-ad-countdown"
                  :class="{ 'wallet-ad-countdown--done': selectedAdRemainingSeconds <= 0 }"
                >
                  {{ selectedAdRemainingSeconds > 0 ? t('account.wallet.watching', { seconds: selectedAdRemainingSeconds }) : t('account.wallet.claiming') }}
                </span>
                <button
                  type="button"
                  class="wallet-ad-dialog__close"
                  :aria-label="t('account.wallet.closeAd')"
                  @click="requestCloseAdDialog"
                >
                  <AppIcon
                    name="close"
                    :size="18"
                  />
                </button>
              </div>
            </header>

            <div class="wallet-ad-dialog__media">
              <video
                v-if="selectedAdTask.media_type === 'video'"
                :src="selectedAdTask.media_url"
                controls
                autoplay
                playsinline
              />
              <img
                v-else
                :src="selectedAdTask.media_url"
                :alt="selectedAdTask.title"
              />
            </div>

            <footer class="wallet-ad-dialog__footer">
              <div>
                <strong>{{ t('account.wallet.rewardPoints', { points: formatPoints(selectedAdTask.reward_points) }) }}</strong>
                <span>{{ t('account.wallet.watchSecondsHint', { seconds: selectedAdTask.watch_seconds }) }}</span>
              </div>

              <div
                v-if="adClaimSucceeded"
                class="wallet-ad-dialog__success-actions"
              >
                <a
                  v-if="selectedAdTask.target_url"
                  class="wallet-action-button wallet-action-button--secondary"
                  href="#"
                  role="button"
                  @click.prevent="openSelectedAdLink"
                >
                  {{ t('account.wallet.openAdLink') }}
                </a>
                <button
                  type="button"
                  class="wallet-action-button"
                  @click="closeAdDialog"
                >
                  {{ t('account.wallet.adClaimDone') }}
                </button>
              </div>
              <button
                v-else
                type="button"
                class="wallet-action-button"
                :disabled="startingAd || claiming || (Boolean(runningSession) && selectedAdRemainingSeconds > 0)"
                @click="runningSession && selectedAdRemainingSeconds <= 0 ? handleClaimSelectedAd() : handleStartSelectedAd()"
              >
                {{ claiming ? t('account.wallet.claiming') : startingAd ? t('common.status.loading') : runningSession && selectedAdRemainingSeconds <= 0 ? t('account.wallet.claimAd') : runningSession ? t('account.wallet.watching', { seconds: selectedAdRemainingSeconds }) : t('account.wallet.startAd') }}
              </button>
            </footer>
          </div>

          <Transition name="wallet-ad-dialog-fade">
            <div
              v-if="showingExitConfirm"
              class="wallet-ad-exit-confirm"
              role="alertdialog"
              aria-modal="true"
            >
              <div class="wallet-ad-exit-confirm__panel">
                <h3>{{ t('account.wallet.exitAdConfirmTitle') }}</h3>
                <p>{{ t('account.wallet.exitAdConfirmDescription') }}</p>
                <div>
                  <button
                    type="button"
                    class="wallet-action-button wallet-action-button--secondary"
                    @click="showingExitConfirm = false"
                  >
                    {{ t('account.wallet.continueWatching') }}
                  </button>
                  <button
                    type="button"
                    class="wallet-action-button"
                    @click="confirmExitAdDialog"
                  >
                    {{ t('account.wallet.exitAd') }}
                  </button>
                </div>
              </div>
            </div>
          </Transition>
        </div>
      </Transition>
    </Teleport>
  </main>
</template>

<style scoped>
.wallet-page {
  display: grid;
  gap: 1rem;
  color: rgb(var(--color-text));
}

.wallet-heading {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 1rem;
  border: 1px solid rgb(var(--color-border) / 0.74);
  border-radius: 0.75rem;
  background:
    linear-gradient(135deg, rgb(var(--color-surface)) 0%, rgb(var(--color-surface-raised)) 100%);
  padding: 1.15rem 1.2rem;
  box-shadow: 0 14px 34px rgb(15 23 42 / 0.055);
}

.wallet-kicker {
  margin: 0;
  color: rgb(var(--color-primary));
  font-size: 0.68rem;
  font-weight: 900;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.wallet-heading h1,
.wallet-panel h2,
.wallet-ad-card h3 {
  margin: 0;
}

.wallet-heading h1 {
  font-family: var(--font-display);
  font-size: clamp(1.55rem, 3vw, 2.35rem);
  line-height: 1.15;
}

.wallet-heading-meter {
  display: grid;
  gap: 0.3rem;
  min-width: 10rem;
  text-align: right;
}

.wallet-heading-meter > span {
  color: rgb(var(--color-text-muted));
  font-size: 0.76rem;
  font-weight: 800;
}

.wallet-heading p,
.wallet-muted,
.wallet-panel-title p,
.wallet-ad-body p,
.wallet-transaction-row span {
  margin: 0;
  color: rgb(var(--color-text-muted));
  line-height: 1.55;
}

.wallet-summary-grid,
.wallet-grid {
  display: grid;
  gap: 1rem;
}

.wallet-summary-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
  align-items: stretch;
}

.wallet-balance-card,
.wallet-stat-card,
.wallet-panel {
  border: 1px solid rgb(var(--color-border) / 0.82);
  border-radius: 0.75rem;
  background: rgb(var(--color-surface));
  box-shadow: 0 12px 28px rgb(15 23 42 / 0.055);
}

.wallet-balance-card,
.wallet-stat-card {
  display: grid;
  align-content: space-between;
  gap: 0.75rem;
  min-height: 9.25rem;
  padding: 1rem;
}

.wallet-balance-card {
  grid-column: span 1;
  background:
    linear-gradient(
      145deg,
      rgb(var(--color-primary)) 0%,
      color-mix(in srgb, rgb(var(--color-primary)) 76%, rgb(var(--color-text)) 24%) 100%
    );
  color: rgb(var(--color-primary-contrast));
}

.wallet-balance-card span,
.wallet-stat-card span {
  font-size: 0.8rem;
  font-weight: 800;
  opacity: 0.82;
}

.wallet-balance-card strong,
.wallet-stat-card strong {
  line-height: 1.1;
}

.wallet-card-label {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.wallet-card-label em {
  border: 1px solid rgb(255 255 255 / 0.28);
  border-radius: 999px;
  padding: 0.18rem 0.55rem;
  color: rgb(255 255 255 / 0.78);
  font-size: 0.66rem;
  font-style: normal;
  font-weight: 900;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.wallet-point-value {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: 0.32rem;
  color: inherit;
}

.wallet-point-value span {
  font-family: var(--font-display);
  font-size: clamp(1.45rem, 2.5vw, 2rem);
  font-weight: 900;
  letter-spacing: 0;
  opacity: 1;
}

.wallet-point-value small {
  color: currentColor;
  font-family: var(--font-sans);
  font-size: 0.68rem;
  font-weight: 900;
  letter-spacing: 0.06em;
  opacity: 0.68;
  text-transform: uppercase;
}

.wallet-point-value-main span {
  font-size: clamp(2rem, 4vw, 3rem);
}

.wallet-point-value-compact {
  justify-content: end;
  color: rgb(var(--color-text));
}

.wallet-point-value-compact span {
  font-size: clamp(1.15rem, 2vw, 1.45rem);
}

.wallet-card-progress {
  display: grid;
  gap: 0.45rem;
}

.wallet-card-progress > div:first-child {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
}

.wallet-card-progress span {
  color: rgb(255 255 255 / 0.76);
  font-size: 0.72rem;
  font-weight: 800;
}

.wallet-panel {
  display: grid;
  gap: 1rem;
  padding: 1rem;
}

.wallet-grid {
  grid-template-columns: minmax(0, 1fr) 20rem;
  align-items: start;
}

.wallet-panel-title {
  display: grid;
  gap: 0.35rem;
}

.wallet-panel-title h2,
.wallet-panel > h2 {
  font-size: 1.05rem;
  line-height: 1.25;
}

.wallet-panel-title p {
  font-size: 0.9rem;
}

.wallet-progress {
  overflow: hidden;
  height: 0.38rem;
  border-radius: 999px;
  background: rgb(255 255 255 / 0.24);
}

.wallet-progress span {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: rgb(255 255 255 / 0.84);
}

.wallet-ad-list,
.wallet-side-stack,
.wallet-transaction-list {
  display: grid;
  gap: 0.75rem;
}

.wallet-ad-card {
  display: grid;
  grid-template-columns: 5.25rem minmax(0, 1fr) auto;
  gap: 0.8rem;
  align-items: center;
  border: 1px solid rgb(var(--color-border) / 0.72);
  border-radius: 0.65rem;
  background: rgb(var(--color-surface-raised) / 0.36);
  padding: 0.7rem;
}

.wallet-ad-cover {
  display: grid;
  aspect-ratio: 4 / 3;
  place-items: center;
  overflow: hidden;
  border-radius: 0.5rem;
  background: rgb(var(--color-surface-raised));
  color: rgb(var(--color-text-muted));
}

.wallet-ad-cover img,
.wallet-ad-cover video {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.wallet-ad-body {
  min-width: 0;
  display: grid;
  gap: 0.35rem;
}

.wallet-ad-body h3 {
  overflow: hidden;
  font-size: 0.96rem;
  font-weight: 900;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.wallet-ad-body span {
  overflow: hidden;
  color: rgb(var(--color-primary));
  font-size: 0.83rem;
  font-weight: 900;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.wallet-ad-body p {
  display: -webkit-box;
  overflow: hidden;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.wallet-action-button {
  min-height: 2.4rem;
  border: 1px solid rgb(var(--color-primary));
  border-radius: 0.5rem;
  background: rgb(var(--color-primary));
  padding: 0 0.85rem;
  color: rgb(var(--color-primary-contrast));
  font-weight: 900;
}

.wallet-action-button--secondary {
  border-color: rgb(var(--color-border));
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
}

.wallet-action-button:disabled {
  cursor: not-allowed;
  border-color: rgb(var(--color-border));
  background: rgb(var(--color-surface-raised));
  color: rgb(var(--color-text-muted));
}

.wallet-transaction-row {
  display: grid;
  gap: 0.35rem;
  border-top: 1px solid rgb(var(--color-border) / 0.7);
  padding-top: 0.75rem;
}

.wallet-transaction-row {
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
}

.wallet-transaction-row div {
  display: grid;
  min-width: 0;
  gap: 0.25rem;
}

.wallet-transaction-row strong,
.wallet-transaction-row span,
.wallet-transaction-row time {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.wallet-transaction-row time {
  color: rgb(var(--color-text-muted));
  font-size: 0.85rem;
}

.wallet-empty {
  border: 1px dashed rgb(var(--color-border));
  border-radius: 0.65rem;
  padding: 1rem;
  color: rgb(var(--color-text-muted));
}

.wallet-ad-dialog,
.wallet-ad-exit-confirm {
  position: fixed;
  z-index: 120;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgb(15 23 42 / 0.38);
  padding: 1rem;
}

.wallet-ad-dialog__panel {
  display: grid;
  width: min(100%, 58rem);
  max-height: min(90vh, 54rem);
  overflow: hidden;
  border: 1px solid rgb(var(--color-border) / 0.74);
  border-radius: 3px;
  background: rgb(var(--color-surface));
  box-shadow: 0 28px 70px rgb(15 23 42 / 0.28);
}

.wallet-ad-dialog__header,
.wallet-ad-dialog__footer {
  display: flex;
  align-items: start;
  justify-content: space-between;
  gap: 1rem;
  padding: 1rem;
}

.wallet-ad-dialog__header {
  border-bottom: 1px solid rgb(var(--color-border) / 0.74);
}

.wallet-ad-dialog__header h2,
.wallet-ad-exit-confirm__panel h3 {
  margin: 0.25rem 0 0;
  color: rgb(var(--color-text));
  font-size: 1.18rem;
  font-weight: 900;
  line-height: 1.25;
}

.wallet-ad-dialog__header p:not(.wallet-kicker),
.wallet-ad-dialog__footer span,
.wallet-ad-exit-confirm__panel p {
  margin: 0.35rem 0 0;
  color: rgb(var(--color-text-muted));
  line-height: 1.55;
}

.wallet-ad-dialog__header-actions {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 0.65rem;
}

.wallet-ad-countdown {
  display: inline-flex;
  min-height: 2.25rem;
  align-items: center;
  border-radius: 999px;
  background: rgb(var(--color-primary-soft));
  padding: 0 0.85rem;
  color: rgb(var(--color-primary));
  font-size: 0.82rem;
  font-weight: 900;
  white-space: nowrap;
}

.wallet-ad-countdown--done {
  background: rgb(var(--color-success) / 0.16);
  color: rgb(var(--color-success));
}

.wallet-ad-dialog__close {
  display: inline-grid;
  width: 2.25rem;
  height: 2.25rem;
  place-items: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 999px;
  background: rgb(var(--color-surface-raised));
  color: rgb(var(--color-text));
}

.wallet-ad-dialog__media {
  display: grid;
  place-items: center;
  min-height: min(62vh, 28rem);
  max-height: min(68vh, 38rem);
  overflow: hidden;
  background: rgb(3 7 18 / 0.92);
}

.wallet-ad-dialog__media img,
.wallet-ad-dialog__media video {
  display: block;
  width: auto;
  height: auto;
  max-width: 100%;
  max-height: min(68vh, 38rem);
  object-fit: contain;
}

.wallet-ad-dialog__footer {
  align-items: center;
  border-top: 1px solid rgb(var(--color-border) / 0.74);
}

.wallet-ad-dialog__footer > div {
  display: grid;
  gap: 0.25rem;
}

.wallet-ad-dialog__success-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: end;
  gap: 0.65rem;
}

.wallet-ad-dialog__footer strong {
  color: rgb(var(--color-primary));
  font-size: 1rem;
}

.wallet-ad-exit-confirm {
  z-index: 130;
  background: rgb(15 23 42 / 0.44);
}

.wallet-ad-exit-confirm__panel {
  width: min(100%, 28rem);
  border: 1px solid rgb(var(--color-border) / 0.74);
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 1rem;
  box-shadow: 0 24px 60px rgb(15 23 42 / 0.24);
}

.wallet-ad-exit-confirm__panel > div {
  display: flex;
  justify-content: end;
  gap: 0.65rem;
  margin-top: 1rem;
}

.wallet-ad-dialog-fade-enter-active,
.wallet-ad-dialog-fade-leave-active {
  transition: opacity 0.2s ease;
}

.wallet-ad-dialog-fade-enter-from,
.wallet-ad-dialog-fade-leave-to {
  opacity: 0;
}

@media (max-width: 980px) {
  .wallet-heading {
    align-items: start;
    flex-direction: column;
  }

  .wallet-heading-meter {
    width: 100%;
    min-width: 0;
    text-align: left;
  }

  .wallet-point-value-compact {
    justify-content: start;
  }

  .wallet-summary-grid,
  .wallet-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .wallet-ad-card,
  .wallet-transaction-row {
    grid-template-columns: 1fr;
  }

  .wallet-ad-dialog__header,
  .wallet-ad-dialog__footer,
  .wallet-ad-dialog__success-actions,
  .wallet-ad-exit-confirm__panel > div {
    align-items: stretch;
    flex-direction: column;
  }

  .wallet-heading,
  .wallet-balance-card,
  .wallet-stat-card,
  .wallet-panel {
    border-radius: 0.65rem;
  }

  .wallet-card-progress > div:first-child,
  .wallet-transaction-row {
    align-items: start;
  }
}

.wallet-page {
  gap: 14px;
}

.wallet-heading,
.wallet-balance-card,
.wallet-stat-card,
.wallet-panel,
.wallet-ad-card,
.wallet-ad-dialog__panel,
.wallet-ad-exit-confirm__panel {
  border-radius: 3px;
  box-shadow: none;
}

.wallet-heading,
.wallet-balance-card,
.wallet-stat-card,
.wallet-panel {
  background: rgb(var(--color-surface));
  border-color: rgb(var(--color-border));
  padding: 14px;
}

.wallet-heading h1 {
  font-family: var(--font-display);
  font-size: 28px;
  font-weight: 400;
}

.wallet-kicker,
.wallet-heading-meter > span,
.wallet-card-label span,
.wallet-card-label em {
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.14em;
}

.wallet-heading p,
.wallet-muted,
.wallet-panel-title p,
.wallet-ad-body p,
.wallet-transaction-row span {
  font-size: 12px;
  line-height: 1.6;
}

.wallet-summary-grid,
.wallet-grid {
  gap: 10px;
}

.wallet-point-value span,
.wallet-point-value-main span {
  font-family: var(--font-sans);
  font-size: 24px;
  font-weight: 500;
}

.wallet-point-value-compact span {
  font-size: 18px;
}

.wallet-panel-title h2,
.wallet-panel > h2,
.wallet-ad-card h3 {
  font-size: 14px;
  font-weight: 600;
}

.wallet-ad-list,
.wallet-side-stack,
.wallet-transaction-list {
  gap: 8px;
}

.wallet-ad-card {
  border-color: rgb(var(--color-border));
}

.wallet-ad-cover {
  border-radius: 2px;
}

.wallet-action-button,
.wallet-ad-countdown,
.wallet-ad-dialog__close {
  border-radius: 2px;
  font-size: 12px;
  font-weight: 600;
}

.wallet-transaction-row {
  border-radius: 2px;
  padding: 10px 0;
}

.wallet-ad-dialog__header,
.wallet-ad-dialog__footer {
  padding: 14px;
}

.wallet-ad-dialog__media {
  background: rgb(var(--color-surface-muted));
}

.wallet-page {
  gap: 16px;
}

.wallet-heading,
.wallet-balance-card,
.wallet-stat-card,
.wallet-panel,
.wallet-ad-card,
.wallet-ad-dialog__panel,
.wallet-ad-exit-confirm__panel {
  border-radius: 8px;
  box-shadow: none;
}

.wallet-heading,
.wallet-balance-card,
.wallet-stat-card,
.wallet-panel {
  background: rgb(var(--color-surface));
  border-color: rgb(var(--color-border));
  padding: 16px;
}

.wallet-heading {
  display: grid;
  gap: 8px;
  border-radius: 0;
  border-width: 0 0 1px;
  padding: 0 0 16px;
}

.wallet-heading h1 {
  font-size: 32px;
}

.wallet-kicker,
.wallet-balance-label {
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.wallet-heading p,
.wallet-muted,
.wallet-panel-title p,
.wallet-ad-body p {
  font-size: 13px;
  line-height: 1.6;
}

.wallet-summary-grid {
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.wallet-grid,
.wallet-ad-list,
.wallet-side-stack {
  gap: 12px;
}

.wallet-balance-card,
.wallet-stat-card {
  min-height: 132px;
  align-content: space-between;
}

.wallet-balance-card--primary .wallet-point-value span,
.wallet-balance-card--primary .wallet-point-value small,
.wallet-balance-card--primary .wallet-balance-label {
  color: rgb(var(--color-primary));
}

.wallet-point-value {
  color: rgb(var(--color-text));
}

.wallet-point-value span,
.wallet-point-value-main span {
  font-family: var(--font-display);
  font-size: 30px;
  font-weight: 400;
}

.wallet-point-value small {
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  font-weight: 700;
}

.wallet-panel-title {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 12px;
  border-bottom: 1px solid rgb(var(--color-border));
  padding-bottom: 14px;
}

.wallet-panel-title h2,
.wallet-panel > h2,
.wallet-ad-card h3 {
  color: rgb(var(--color-primary));
  font-size: 14px;
  font-weight: 700;
}

.wallet-panel-meter {
  display: grid;
  gap: 6px;
  min-width: 120px;
}

.wallet-panel-meter strong {
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  font-weight: 700;
}

.wallet-progress {
  height: 6px;
  background: rgb(var(--color-primary) / 0.12);
}

.wallet-progress span {
  background: rgb(var(--color-primary));
}

.wallet-ad-card {
  grid-template-columns: 132px minmax(0, 1fr) auto;
  gap: 14px;
  border-color: rgb(var(--color-border));
  background: rgb(var(--color-surface));
  padding: 12px;
}

.wallet-ad-cover {
  border-radius: 6px;
}

.wallet-ad-body h3 {
  font-size: 15px;
  font-weight: 700;
}

.wallet-ad-body span {
  color: rgb(var(--color-primary));
  font-size: 12px;
  font-weight: 700;
}

.wallet-action-button,
.wallet-ad-countdown,
.wallet-ad-dialog__close {
  border-radius: 6px;
  font-size: 12px;
  font-weight: 700;
}

.wallet-transaction-table-wrap {
  overflow-x: auto;
}

.wallet-transaction-table {
  width: 100%;
  min-width: 640px;
  border-collapse: collapse;
  font-size: 13px;
}

.wallet-transaction-table th,
.wallet-transaction-table td {
  border-top: 1px solid rgb(var(--color-border));
  padding: 12px 10px;
  text-align: left;
  vertical-align: middle;
}

.wallet-transaction-table thead th {
  border-top: 0;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  font-weight: 700;
}

.wallet-transaction-table tbody td {
  color: rgb(var(--color-text));
}

@media (max-width: 980px) {
  .wallet-summary-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .wallet-summary-grid,
  .wallet-grid {
    grid-template-columns: 1fr;
  }

  .wallet-panel-meter {
    min-width: 0;
    width: 100%;
  }
}
</style>
