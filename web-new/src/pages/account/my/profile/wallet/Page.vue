<!--
 * 會員錢包頁。
 * 1. 顯示 AJO Point 餘額、廣告獎勵與最近流水。
 * 2. 提供線上充值、支付狀態查詢與廣告積分領取流程。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import {
  claimRewardAdTask,
  createWalletRechargeOrder,
  fetchWalletRechargeOrder,
  fetchRewardAdTasks,
  fetchWalletOverview,
  startRewardAdTask,
  trackRewardAdClick,
} from '@/httpapis/wallet';
import type {
  RewardAdSessionResponse,
  RewardAdTaskResponse,
  WalletOverviewResponse,
  WalletRechargeDeviceMode,
  WalletRechargeOrderResponse,
  WalletRechargePayMethod,
  WalletRechargePayRegion,
} from '@/model/wallet';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import QrCodeImage from '@/shared/components/base/QrCodeImage.vue';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';
import { useSessionStore } from '@/stores/session';
import { formatDate } from '@/utils/format';
import { formatAjoPoints } from '@/utils/wallet';

interface RunningAdSession {
  taskId: string;
  claimId: string;
  remainingSeconds: number;
  availableAtTime: number;
}

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const feedbackStore = useFeedbackStore();
const preferenceStore = usePreferenceStore();
const sessionStore = useSessionStore();

const loading = ref(false);
const claiming = ref(false);
const rechargeSubmitting = ref(false);
const rechargeRefreshing = ref(false);
const overview = ref<WalletOverviewResponse | null>(null);
const adTasks = ref<RewardAdTaskResponse[]>([]);
const runningSession = ref<RunningAdSession | null>(null);
const selectedAdTask = ref<RewardAdTaskResponse | null>(null);
const showingExitConfirm = ref(false);
const adClaimSucceeded = ref(false);
const startingAd = ref(false);
const showingRechargeDialog = ref(false);
const rechargeAmount = ref<number>(100);
const rechargePayMethod = ref<WalletRechargePayMethod>('wechat');
const rechargePayRegion = ref<WalletRechargePayRegion>('HK');
const activeRechargeOrder = ref<WalletRechargeOrderResponse | null>(null);
let countdownTimer: number | undefined;
let rechargePollTimer: number | undefined;

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
const rechargeRate = computed(() => overview.value?.recharge_rate ?? 100);
const rechargeAmountCents = computed(() => Math.round((Number(rechargeAmount.value) || 0) * 100));
const rechargePoints = computed(() =>
  Math.max(0, Math.floor(rechargeAmountCents.value * rechargeRate.value / 100)));
const rechargeAmountValue = computed(() => rechargeAmountCents.value / 100);
const rechargeAmountIsValid = computed(() =>
  rechargeAmountCents.value >= 1 && rechargeAmountCents.value <= 5000000);
const rechargeMethods = computed<Array<{ key: WalletRechargePayMethod; label: string; description: string }>>(() => [
  {
    key: 'wechat',
    label: t('account.wallet.payWechat'),
    description: t('account.wallet.payWechatHint'),
  },
  {
    key: 'alipay',
    label: t('account.wallet.payAlipay'),
    description: t('account.wallet.payAlipayHint'),
  },
  {
    key: 'unionpay',
    label: t('account.wallet.payUnionpay'),
    description: t('account.wallet.payUnionpayHint'),
  },
]);
const rechargeRegionOptions = computed<Array<{ key: WalletRechargePayRegion; label: string }>>(() => [
  { key: 'HK', label: t('account.wallet.regionHK') },
  { key: 'CN', label: t('account.wallet.regionCN') },
]);
const shouldShowRechargeRegion = computed(() => rechargePayMethod.value === 'alipay');
const activeRechargeIsPending = computed(() => activeRechargeOrder.value?.state === 'PAYING');
const activeRechargeMethodLabel = computed(() => {
  const order = activeRechargeOrder.value;
  if (!order) {
    return '';
  }
  if (order.pay_channel === 'YSF_QR') {
    return t('account.wallet.payUnionpay');
  }
  if (order.pay_channel.startsWith('WX_')) {
    return t('account.wallet.payWechat');
  }
  if (order.pay_region === 'CN') {
    return t('account.wallet.payAlipayCN');
  }
  return t('account.wallet.payAlipayHK');
});
const activeRechargeStateClass = computed(() => {
  const state = activeRechargeOrder.value?.state ?? '';
  if (state === 'SUCCESS') {
    return 'wallet-recharge-status--success';
  }
  if (state === 'FAILED' || state === 'CLOSED' || state === 'EXPIRED' || state === 'REVOKED') {
    return 'wallet-recharge-status--failed';
  }
  return 'wallet-recharge-status--pending';
});
const activeRechargeQrImage = computed(() => {
  const order = activeRechargeOrder.value;
  if (!order || order.state !== 'PAYING') {
    return '';
  }

  return normalizePayDataType(order.pay_data_type) === 'codeimgurl' ? order.pay_data : '';
});
const activeRechargeQrLink = computed(() => {
  const order = activeRechargeOrder.value;
  if (!order || order.state !== 'PAYING') {
    return '';
  }
  if (normalizePayDataType(order.pay_data_type) === 'codeurl') {
    return order.pay_data;
  }
  if (normalizePayDataType(order.pay_data_type) === 'payurl' && isDesktopBrowser()) {
    return order.pay_data;
  }

  return '';
});
const activeRechargeCanRedirect = computed(() => {
  const order = activeRechargeOrder.value;
  return Boolean(order && order.state === 'PAYING' && normalizePayDataType(order.pay_data_type) === 'payurl' && order.pay_data);
});
const activeRechargeHasPaymentEntry = computed(() =>
  Boolean(activeRechargeQrImage.value || activeRechargeQrLink.value || activeRechargeCanRedirect.value));

// 1. 格式化錢包主數字
const formatPointNumber = (value: number): string => pointFormatter.format(value);
const formatPoints = (value: number): string =>
  formatAjoPoints(value, t('common.brand.pointsName'), preferenceStore.locale);
const formatHKD = (value: number | string): string =>
  new Intl.NumberFormat(preferenceStore.locale, {
    style: 'currency',
    currency: 'HKD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(Number(value) || 0);

// 2. 讀取錯誤訊息
const readErrorMessage = (error: unknown, fallback: string): string =>
  axios.isAxiosError<{ message?: string }>(error)
    ? error.response?.data?.message ?? fallback
    : fallback;

// 3.1 判斷桌面瀏覽器
const isDesktopBrowser = (): boolean => {
  if (typeof navigator === 'undefined') {
    return true;
  }
  const userAgent = navigator.userAgent.toLowerCase();
  const isTablet =
    userAgent.includes('ipad') ||
    (navigator.platform === 'MacIntel' && navigator.maxTouchPoints > 1) ||
    (userAgent.includes('android') && !userAgent.includes('mobile'));
  const isMobile = /(iphone|ipod|windows phone|iemobile|blackberry|bb10|opera mini|mobile)/i.test(userAgent);
  return !isMobile && !isTablet;
};

// 3.2 輸出設備支付模式
const resolveDeviceMode = (): WalletRechargeDeviceMode =>
  isDesktopBrowser() ? 'desktop' : 'mobile';

// 3.3 標準化網關付款資料類型
const normalizePayDataType = (value: string): string =>
  String(value || '').trim().toLowerCase();

// 3.4 建立 H5 回跳地址
const buildRechargeReturnPath = (): string => {
  if (typeof window === 'undefined') {
    return '/account/profile/wallet';
  }

  return `${window.location.origin}/account/profile/wallet`;
};

// 3.5 讀取錢包和廣告任務
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

// 3.6 發起充值
const submitRecharge = async (): Promise<void> => {
  if (rechargeSubmitting.value) {
    return;
  }
  if (!rechargeAmountIsValid.value) {
    feedbackStore.pushToast(t('account.wallet.rechargeAmountInvalid'), 'error');
    return;
  }

  rechargeSubmitting.value = true;
  try {
    const response = await createWalletRechargeOrder({
      amount_hkd: rechargeAmountValue.value,
      pay_method: rechargePayMethod.value,
      pay_region: rechargePayRegion.value,
      device_mode: resolveDeviceMode(),
      return_path: buildRechargeReturnPath(),
    });
    activeRechargeOrder.value = response.data.data;
    if (activeRechargeCanRedirect.value && !isDesktopBrowser()) {
      window.location.assign(response.data.data.pay_data);
      return;
    }
    feedbackStore.pushToast(t('account.wallet.rechargeOrderCreated'), 'success');
  } catch (error: unknown) {
    feedbackStore.pushToast(readErrorMessage(error, t('account.wallet.rechargeCreateError')), 'error');
  } finally {
    rechargeSubmitting.value = false;
  }
};

// 3.7 刷新充值訂單
const refreshRechargeOrder = async (silent = false): Promise<void> => {
  const order = activeRechargeOrder.value;
  if (!order || rechargeRefreshing.value) {
    return;
  }

  rechargeRefreshing.value = true;
  try {
    const response = await fetchWalletRechargeOrder(order.order_id, { refresh: true });
    activeRechargeOrder.value = response.data.data;
    if (response.data.data.state === 'SUCCESS') {
      feedbackStore.pushToast(t('account.wallet.rechargeSuccess'), 'success');
      await loadWallet();
      stopRechargePolling();
      activeRechargeOrder.value = null;
      showingRechargeDialog.value = false;
    } else if (!silent) {
      feedbackStore.pushToast(t('account.wallet.rechargeStatusUpdated'), 'info');
    }
  } catch (error: unknown) {
    feedbackStore.pushToast(readErrorMessage(error, t('account.wallet.rechargeRefreshError')), 'error');
  } finally {
    rechargeRefreshing.value = false;
  }
};

// 3.8 從支付回跳參數恢復充值訂單
const hydrateRechargeReturn = async (): Promise<void> => {
  const queryOrderNo = String(route.query.mch_order_no ?? '').trim();
  const isWalletRechargeReturn = String(route.query.wallet_recharge ?? '') === '1';
  if (!queryOrderNo || !isWalletRechargeReturn) {
    return;
  }

  rechargeRefreshing.value = true;
  try {
    const response = await fetchWalletRechargeOrder(queryOrderNo, { refresh: true });
    activeRechargeOrder.value = response.data.data;
    await loadWallet();
    if (response.data.data.state === 'SUCCESS') {
      feedbackStore.pushToast(t('account.wallet.rechargeSuccess'), 'success');
      activeRechargeOrder.value = null;
      showingRechargeDialog.value = false;
    }
  } catch (error: unknown) {
    feedbackStore.pushToast(readErrorMessage(error, t('account.wallet.rechargeRefreshError')), 'error');
  } finally {
    rechargeRefreshing.value = false;
    await router.replace({ path: route.path, query: {} });
  }
};

// 3.9 打開當前充值支付鏈接
const openRechargePayData = (): void => {
  const order = activeRechargeOrder.value;
  if (!order?.pay_data) {
    return;
  }

  window.location.assign(order.pay_data);
};

// 3.10 切換充值支付方式
const selectRechargeMethod = (method: WalletRechargePayMethod): void => {
  rechargePayMethod.value = method;
  if (method === 'wechat' || method === 'unionpay') {
    rechargePayRegion.value = 'HK';
  }
};

// 3.11 打開充值彈窗
const openRechargeDialog = (): void => {
  showingRechargeDialog.value = true;
};

// 3.12 關閉充值彈窗
const closeRechargeDialog = (): void => {
  if (rechargeSubmitting.value || rechargeRefreshing.value) {
    return;
  }
  showingRechargeDialog.value = false;
};

// 3.13 停止充值查單輪詢
const stopRechargePolling = (): void => {
  if (rechargePollTimer) {
    window.clearInterval(rechargePollTimer);
    rechargePollTimer = undefined;
  }
};

// 3.14 根據彈窗狀態啟停充值查單輪詢
const syncRechargePolling = (): void => {
  stopRechargePolling();
  if (!showingRechargeDialog.value || activeRechargeOrder.value?.state !== 'PAYING') {
    return;
  }

  rechargePollTimer = window.setInterval(() => {
    void refreshRechargeOrder(true);
  }, 3000);
};

// 4. 停止倒數計時
const stopCountdown = (): void => {
  if (countdownTimer) {
    window.clearInterval(countdownTimer);
    countdownTimer = undefined;
  }
};

// 4.1 計算最新剩餘觀看秒數
const getRunningSessionRemainingSeconds = (session: RunningAdSession): number =>
  Math.max(0, Math.ceil((session.availableAtTime - Date.now()) / 1000));

// 4.2 判斷目前廣告是否已完整觀看
const isSelectedAdReadyToClaim = (): boolean => {
  const task = selectedAdTask.value;
  const session = runningSession.value;
  if (!task || !session || session.taskId !== task.task_id) {
    return false;
  }

  session.remainingSeconds = getRunningSessionRemainingSeconds(session);
  return session.remainingSeconds <= 0;
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
    availableAtTime: targetAvailableAtTime,
    remainingSeconds: calculateRemainingSeconds(),
  };
  countdownTimer = window.setInterval(() => {
    if (!runningSession.value) {
      stopCountdown();
      return;
    }
    runningSession.value.remainingSeconds = getRunningSessionRemainingSeconds(runningSession.value);
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
  const session = runningSession.value;
  if (!session || session.taskId !== task.task_id) {
    return;
  }
  session.remainingSeconds = getRunningSessionRemainingSeconds(session);
  if (session.remainingSeconds > 0) {
    return;
  }
  if (claiming.value) {
    return;
  }

  claiming.value = true;
  try {
    await claimRewardAdTask(task.task_id, session.claimId);
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
const requestCloseAdDialog = async (): Promise<void> => {
  if (claiming.value || startingAd.value) {
    return;
  }

  if (isSelectedAdReadyToClaim()) {
    await handleClaimSelectedAd();
    if (adClaimSucceeded.value) {
      closeAdDialog();
    }
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
const confirmExitAdDialog = async (): Promise<void> => {
  if (claiming.value || startingAd.value) {
    return;
  }
  if (isSelectedAdReadyToClaim()) {
    await handleClaimSelectedAd();
    if (adClaimSucceeded.value) {
      closeAdDialog();
    }
    return;
  }

  closeAdDialog();
};

onMounted(() => {
  void hydrateRechargeReturn().then(() => {
    if (!overview.value) {
      return loadWallet();
    }
    return undefined;
  });
});

onBeforeUnmount(() => {
  if (isSelectedAdReadyToClaim()) {
    void handleClaimSelectedAd();
  }
  stopCountdown();
  stopRechargePolling();
});

watch(
  [showingRechargeDialog, activeRechargeOrder],
  () => {
    syncRechargePolling();
  },
);
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
                <img
                  v-if="task.cover_url"
                  :src="task.cover_url"
                  :alt="task.title"
                />
                <video
                  v-else-if="task.media_type === 'video' && task.media_url"
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
                <div class="wallet-ad-meta">
                  <span>{{ t('account.wallet.rewardPoints', { points: formatPoints(task.reward_points) }) }}</span>
                  <span>{{ t('account.wallet.watchSecondsValue', { seconds: task.watch_seconds }) }}</span>
                  <span>{{ resolveAdActionLabel(task) }}</span>
                </div>
              </div>

              <button
                type="button"
                class="wallet-action-button wallet-action-button--compact"
                :disabled="isAdActionDisabled(task)"
                @click="runAdAction(task)"
              >
                {{ resolveAdActionLabel(task) }}
              </button>
            </article>
          </div>
        </article>

        <aside class="wallet-side-stack">
          <article class="wallet-panel wallet-recharge-panel">
            <div class="wallet-panel-title">
              <div>
                <h2>{{ t('account.wallet.rechargeTitle') }}</h2>
                <p>{{ t('account.wallet.rechargeDescription', { rate: rechargeRate }) }}</p>
              </div>
            </div>

            <div class="wallet-recharge-summary">
              <span>{{ t('account.wallet.rechargePointsPreview') }}</span>
              <strong>{{ formatPoints(rechargePoints) }}</strong>
            </div>

            <section
              v-if="activeRechargeIsPending"
              class="wallet-recharge-mini-order"
            >
              <div>
                <span>{{ activeRechargeMethodLabel }}</span>
                <strong>{{ activeRechargeOrder ? formatHKD(activeRechargeOrder.amount_hkd) : '' }}</strong>
              </div>
              <em
                class="wallet-recharge-status"
                :class="activeRechargeStateClass"
              >
                {{ activeRechargeOrder?.state_label }}
              </em>
            </section>

            <button
              type="button"
              class="wallet-action-button wallet-recharge-submit"
              @click="openRechargeDialog"
            >
              <AppIcon
                name="plus-square"
                :size="16"
              />
              <span>{{ t('account.wallet.rechargeOpenDialog') }}</span>
            </button>
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
          v-if="showingRechargeDialog"
          class="wallet-recharge-dialog"
          role="dialog"
          aria-modal="true"
          :aria-label="t('account.wallet.rechargeTitle')"
          @click.self="closeRechargeDialog"
        >
          <div class="wallet-recharge-dialog__panel">
            <header class="wallet-recharge-dialog__header">
              <div>
                <p class="wallet-kicker">{{ t('common.brand.pointsName') }}</p>
                <h2>{{ t('account.wallet.rechargeTitle') }}</h2>
                <p>{{ t('account.wallet.rechargeDescription', { rate: rechargeRate }) }}</p>
              </div>
              <button
                type="button"
                class="wallet-ad-dialog__close"
                :aria-label="t('account.wallet.rechargeClose')"
                @click="closeRechargeDialog"
              >
                <AppIcon
                  name="close"
                  :size="18"
                />
              </button>
            </header>

            <div class="wallet-recharge-dialog__body">
              <div class="wallet-recharge-form">
                <label class="wallet-recharge-field">
                  <span>{{ t('account.wallet.rechargeAmountLabel') }}</span>
                  <div class="wallet-recharge-input">
                    <span>HKD</span>
                    <input
                      v-model.number="rechargeAmount"
                      type="number"
                      min="0.01"
                      max="50000"
                      step="0.01"
                      inputmode="decimal"
                    />
                  </div>
                </label>

                <div class="wallet-recharge-points">
                  <span>{{ t('account.wallet.rechargePointsPreview') }}</span>
                  <strong>{{ formatPoints(rechargePoints) }}</strong>
                </div>

                <div
                  class="wallet-recharge-methods"
                  role="radiogroup"
                  :aria-label="t('account.wallet.rechargeMethodLabel')"
                >
                  <button
                    v-for="method in rechargeMethods"
                    :key="method.key"
                    type="button"
                    class="wallet-recharge-method"
                    :class="{ 'wallet-recharge-method--active': rechargePayMethod === method.key }"
                    @click="selectRechargeMethod(method.key)"
                  >
                    <strong>{{ method.label }}</strong>
                    <small>{{ method.description }}</small>
                  </button>
                </div>

                <div
                  v-if="shouldShowRechargeRegion"
                  class="wallet-recharge-regions"
                  role="radiogroup"
                  :aria-label="t('account.wallet.rechargeRegionLabel')"
                >
                  <button
                    v-for="region in rechargeRegionOptions"
                    :key="region.key"
                    type="button"
                    :class="{ 'wallet-recharge-region--active': rechargePayRegion === region.key }"
                    @click="rechargePayRegion = region.key"
                  >
                    {{ region.label }}
                  </button>
                </div>

                <button
                  type="button"
                  class="wallet-action-button wallet-recharge-submit"
                  :disabled="rechargeSubmitting || !rechargeAmountIsValid"
                  @click="submitRecharge"
                >
                  <AppIcon
                    name="send"
                    :size="16"
                  />
                  <span>{{ rechargeSubmitting ? t('account.wallet.rechargeCreating') : t('account.wallet.rechargeCreateAction') }}</span>
                </button>
              </div>

              <section
                v-if="activeRechargeOrder"
                class="wallet-recharge-order"
              >
                <div class="wallet-recharge-order-header">
                  <div>
                    <span>{{ activeRechargeMethodLabel }}</span>
                    <strong>{{ formatHKD(activeRechargeOrder.amount_hkd) }}</strong>
                  </div>
                  <em
                    class="wallet-recharge-status"
                    :class="activeRechargeStateClass"
                  >
                    {{ activeRechargeOrder.state_label }}
                  </em>
                </div>

                <div class="wallet-recharge-order-meta">
                  <span>{{ t('account.wallet.rechargeOrderNo') }}</span>
                  <strong>{{ activeRechargeOrder.mch_order_no }}</strong>
                  <span>{{ t('account.wallet.rechargePoints') }}</span>
                  <strong>{{ formatPoints(activeRechargeOrder.points_amount) }}</strong>
                </div>

                <div
                  v-if="activeRechargeIsPending && activeRechargeHasPaymentEntry"
                  class="wallet-recharge-pay-entry"
                >
                  <img
                    v-if="activeRechargeQrImage"
                    :src="activeRechargeQrImage"
                    :alt="t('account.wallet.rechargeQrAlt')"
                  />
                  <QrCodeImage
                    v-else-if="activeRechargeQrLink"
                    :text="activeRechargeQrLink"
                    :alt="t('account.wallet.rechargeQrAlt')"
                    :size="220"
                  />
                  <button
                    v-else-if="activeRechargeCanRedirect"
                    type="button"
                    class="wallet-action-button wallet-recharge-pay-button"
                    @click="openRechargePayData"
                  >
                    {{ t('account.wallet.rechargeContinuePay') }}
                  </button>
                </div>

                <div class="wallet-recharge-order-actions">
                  <button
                    type="button"
                    class="wallet-action-button wallet-action-button--secondary"
                    :disabled="rechargeRefreshing"
                    @click="refreshRechargeOrder()"
                  >
                    <AppIcon
                      name="reload"
                      :size="16"
                    />
                    <span>{{ rechargeRefreshing ? t('common.status.loading') : t('account.wallet.rechargeRefresh') }}</span>
                  </button>
                  <button
                    v-if="activeRechargeCanRedirect"
                    type="button"
                    class="wallet-action-button wallet-recharge-pay-button"
                    @click="openRechargePayData"
                  >
                    {{ t('account.wallet.rechargeContinuePay') }}
                  </button>
                </div>
              </section>
            </div>
          </div>
        </div>
      </Transition>

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
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
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

.wallet-action-button--compact {
  min-height: 2rem;
  border-radius: 0.35rem;
  padding: 0 0.65rem;
  font-size: 12px;
  white-space: nowrap;
}

.wallet-action-button:disabled {
  cursor: not-allowed;
  border-color: rgb(var(--color-border));
  background: rgb(var(--color-surface-raised));
  color: rgb(var(--color-text-muted));
}

.wallet-recharge-panel {
  gap: 0.85rem;
}

.wallet-recharge-summary {
  display: grid;
  gap: 0.25rem;
  border-radius: 0.65rem;
  background: rgb(var(--color-primary-soft));
  padding: 0.8rem;
}

.wallet-recharge-summary span,
.wallet-recharge-mini-order span {
  color: rgb(var(--color-text-muted));
  font-size: 0.76rem;
  font-weight: 900;
}

.wallet-recharge-summary strong {
  color: rgb(var(--color-primary));
  font-size: 1.25rem;
  line-height: 1.2;
}

.wallet-recharge-mini-order {
  display: flex;
  align-items: start;
  justify-content: space-between;
  gap: 0.75rem;
  border: 1px solid rgb(var(--color-border) / 0.72);
  border-radius: 0.65rem;
  background: rgb(var(--color-surface-raised) / 0.34);
  padding: 0.75rem;
}

.wallet-recharge-mini-order > div {
  display: grid;
  min-width: 0;
  gap: 0.2rem;
}

.wallet-recharge-mini-order strong {
  font-size: 0.95rem;
}

.wallet-recharge-dialog {
  position: fixed;
  z-index: 120;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgb(15 23 42 / 0.38);
  padding: 1rem;
}

.wallet-recharge-dialog__panel {
  display: grid;
  width: min(100%, 54rem);
  max-height: min(90vh, 44rem);
  overflow: hidden;
  border: 1px solid rgb(var(--color-border) / 0.74);
  border-radius: 8px;
  background: rgb(var(--color-surface));
  box-shadow: 0 28px 70px rgb(15 23 42 / 0.28);
}

.wallet-recharge-dialog__header {
  display: flex;
  align-items: start;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid rgb(var(--color-border) / 0.74);
  padding: 1rem;
}

.wallet-recharge-dialog__header h2 {
  margin: 0.25rem 0 0;
  color: rgb(var(--color-text));
  font-size: 1.18rem;
  font-weight: 900;
  line-height: 1.25;
}

.wallet-recharge-dialog__header p:not(.wallet-kicker) {
  margin: 0.35rem 0 0;
  color: rgb(var(--color-text-muted));
  line-height: 1.55;
}

.wallet-recharge-dialog__body {
  display: grid;
  grid-template-columns: minmax(0, 0.9fr) minmax(0, 1.1fr);
  gap: 1rem;
  overflow: auto;
  padding: 1rem;
}

.wallet-recharge-form {
  display: grid;
  align-content: start;
  gap: 0.85rem;
}

.wallet-recharge-field {
  display: grid;
  gap: 0.45rem;
}

.wallet-recharge-field > span,
.wallet-recharge-points > span {
  color: rgb(var(--color-text-muted));
  font-size: 0.76rem;
  font-weight: 900;
}

.wallet-recharge-input {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.55rem;
  background: rgb(var(--color-surface-raised) / 0.45);
}

.wallet-recharge-input span {
  display: inline-flex;
  align-items: center;
  border-right: 1px solid rgb(var(--color-border));
  padding: 0 0.7rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 900;
}

.wallet-recharge-input input {
  width: 100%;
  min-width: 0;
  border: 0;
  background: transparent;
  padding: 0.72rem 0.75rem;
  color: rgb(var(--color-text));
  font: inherit;
  font-weight: 900;
  outline: none;
}

.wallet-recharge-points {
  display: grid;
  gap: 0.25rem;
  border-radius: 0.6rem;
  background: rgb(var(--color-primary-soft));
  padding: 0.75rem;
}

.wallet-recharge-points strong {
  color: rgb(var(--color-primary));
  font-size: 1.18rem;
  line-height: 1.2;
}

.wallet-recharge-methods {
  display: grid;
  gap: 0.55rem;
}

.wallet-recharge-method {
  display: grid;
  gap: 0.25rem;
  width: 100%;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.55rem;
  background: rgb(var(--color-surface-raised) / 0.38);
  padding: 0.72rem;
  color: rgb(var(--color-text));
  text-align: left;
}

.wallet-recharge-method strong {
  font-size: 0.9rem;
}

.wallet-recharge-method small {
  color: rgb(var(--color-text-muted));
  font-size: 0.76rem;
  font-weight: 700;
  line-height: 1.45;
}

.wallet-recharge-method--active {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary-soft));
}

.wallet-recharge-regions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.45rem;
}

.wallet-recharge-regions button {
  min-height: 2.25rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.5rem;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text-muted));
  font-weight: 900;
}

.wallet-recharge-regions .wallet-recharge-region--active {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.wallet-recharge-submit,
.wallet-recharge-pay-button {
  width: 100%;
}

.wallet-recharge-order {
  display: grid;
  align-content: start;
  gap: 0.75rem;
  border: 1px solid rgb(var(--color-border) / 0.76);
  border-radius: 0.65rem;
  background: rgb(var(--color-surface-raised) / 0.24);
  padding: 0.9rem;
}

.wallet-recharge-order-header {
  display: flex;
  align-items: start;
  justify-content: space-between;
  gap: 0.75rem;
}

.wallet-recharge-order-header > div {
  display: grid;
  min-width: 0;
  gap: 0.2rem;
}

.wallet-recharge-order-header span {
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 800;
}

.wallet-recharge-order-header strong {
  font-size: 1.05rem;
}

.wallet-recharge-status {
  flex: 0 0 auto;
  border-radius: 999px;
  padding: 0.2rem 0.55rem;
  font-size: 0.72rem;
  font-style: normal;
  font-weight: 900;
  white-space: nowrap;
}

.wallet-recharge-status--pending {
  background: rgb(var(--color-primary-soft));
  color: rgb(var(--color-primary));
}

.wallet-recharge-status--success {
  background: rgb(var(--color-success) / 0.14);
  color: rgb(var(--color-success));
}

.wallet-recharge-status--failed {
  background: rgb(var(--color-danger) / 0.12);
  color: rgb(var(--color-danger));
}

.wallet-recharge-order-meta {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: 0.35rem 0.6rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.76rem;
}

.wallet-recharge-order-meta strong {
  min-width: 0;
  overflow-wrap: anywhere;
  color: rgb(var(--color-text));
  font-weight: 900;
}

.wallet-recharge-pay-entry {
  display: grid;
  place-items: center;
  overflow: hidden;
  border: 1px solid rgb(var(--color-border) / 0.76);
  border-radius: 0.65rem;
  background: rgb(var(--color-surface-raised) / 0.35);
  padding: 0.75rem;
}

.wallet-recharge-pay-entry img {
  display: block;
  width: min(100%, 13.75rem);
  height: auto;
  border-radius: 0.45rem;
}

.wallet-recharge-order-actions {
  display: grid;
  gap: 0.5rem;
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

  .wallet-recharge-dialog__body {
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
  .wallet-recharge-dialog__header,
  .wallet-ad-dialog__success-actions,
  .wallet-ad-exit-confirm__panel > div {
    align-items: stretch;
    flex-direction: column;
  }

  .wallet-heading,
  .wallet-balance-card,
  .wallet-stat-card,
  .wallet-panel,
  .wallet-recharge-dialog__panel {
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
.wallet-recharge-dialog__panel,
.wallet-recharge-mini-order,
.wallet-recharge-method,
.wallet-recharge-input,
.wallet-recharge-points,
.wallet-recharge-order,
.wallet-recharge-pay-entry,
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

.wallet-ad-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.wallet-ad-meta span {
  display: inline-flex;
  max-width: 100%;
  align-items: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 999px;
  background: rgb(var(--color-surface-raised) / 0.42);
  padding: 5px 8px;
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  font-weight: 700;
  line-height: 1;
}

.wallet-ad-meta span:first-child {
  border-color: rgb(var(--color-primary) / 0.24);
  background: rgb(var(--color-primary-soft));
  color: rgb(var(--color-primary));
}

.wallet-action-button,
.wallet-ad-countdown,
.wallet-ad-dialog__close {
  border-radius: 6px;
  font-size: 12px;
  font-weight: 700;
}

.wallet-recharge-dialog__header,
.wallet-recharge-dialog__body {
  padding: 16px;
}

.wallet-recharge-field > span,
.wallet-recharge-points > span,
.wallet-recharge-summary span,
.wallet-recharge-mini-order span,
.wallet-recharge-order-header span,
.wallet-recharge-order-meta {
  font-size: 12px;
}

.wallet-recharge-summary strong,
.wallet-recharge-points strong {
  color: rgb(var(--color-primary));
  font-size: 18px;
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
