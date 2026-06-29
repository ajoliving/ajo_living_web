<!--
 * AJO Pay 首頁。
 * 1. 高保真還原桌面 HTML 原型的首頁層級與視覺密度。
 * 2. 讀取會員已綁定單位的待繳賬單，展示管理費、其他費用與回贈估算。
 * 3. 僅提供進入付款頁與會員中心綁定單位兩個正式入口。
-->
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { RouterLink, useRouter } from 'vue-router';

import { fetchPOSPaymentBills } from '@/httpapis/payments';
import type { POSPaymentContext, POSPaymentRow } from '@/model/payments';
import { useFeedbackStore } from '@/stores/feedback';

const router = useRouter();
const feedbackStore = useFeedbackStore();
const rows = ref<POSPaymentRow[]>([]);
const context = ref<POSPaymentContext | undefined>();
const isLoading = ref(false);
const needsProfile = ref(false);
const needsIsmart = ref(false);

// 1. 讀取字串值
const readPaymentText = (row: POSPaymentRow, keys: string[]): string => {
  for (const key of keys) {
    const value = row[key];
    if (value !== undefined && value !== null && String(value).trim() !== '') {
      return String(value).trim();
    }
  }
  return '';
};

// 2. 讀取金額值
const readPaymentAmountValue = (row: POSPaymentRow): number => {
  const raw = readPaymentText(row, [
    'net_amount',
    'amount',
    'final_amount',
    'paid_amount',
    'trs_val',
    'total',
    'total_amount',
    'payable',
  ]);
  const numeric = Number(raw);
  return Number.isFinite(numeric) ? numeric : 0;
};

// 3. 格式化港幣金額
const formatHKD = (value: number): string =>
  `HK$${value.toLocaleString('en-HK', { minimumFractionDigits: 0, maximumFractionDigits: 2 })}`;

// 4. 讀取賬單名稱
const readBillName = (row: POSPaymentRow): string =>
  readPaymentText(row, ['item_name', 'item_id', 'name', 'fee_name', 'item']) || '物業費';

// 5. 判斷是否屬於管理費
const isManagementBill = (row: POSPaymentRow): boolean => {
  const label = readBillName(row).toLowerCase();
  return label.includes('管理') || label.includes('management');
};

const payableRows = computed(() =>
  rows.value.filter((row) => readPaymentAmountValue(row) > 0),
);
const managementAmount = computed(() =>
  payableRows.value.filter(isManagementBill).reduce((sum, row) => sum + readPaymentAmountValue(row), 0),
);
const otherAmount = computed(() =>
  payableRows.value.filter((row) => !isManagementBill(row)).reduce((sum, row) => sum + readPaymentAmountValue(row), 0),
);
const totalAmount = computed(() => managementAmount.value + otherAmount.value);
const rewardCoins = computed(() => Math.floor(totalAmount.value / 100));
const contextLabel = computed(() => {
  if (context.value) {
    return [
      context.value.building_name || context.value.building_id,
      context.value.unit_label || [context.value.floor, context.value.unit].filter(Boolean).join(' / '),
    ].filter(Boolean).join(' · ');
  }
  if (needsProfile.value) {
    return '請先於會員中心綁定繳費單位';
  }
  if (needsIsmart.value) {
    return '請先於會員中心綁定或更新 iSmart 帳戶';
  }
  return '目前未有待繳賬單';
});
const primaryActionLabel = computed(() => {
  if (needsProfile.value) {
    return '前往綁定單位';
  }
  if (needsIsmart.value) {
    return '前往會員中心';
  }
  return '立即繳費';
});
const featureCards = computed(() => [
  {
    title: '繳交大廈管理費',
    text: '以目前會員綁定單位讀取賬單，支援按項目核對後付款。',
    tone: 'blue',
  },
  {
    title: '繳交其他費用',
    text: '將單位相關待繳費用集中在付款頁第一步，避免分散查找。',
    tone: 'orange',
  },
  {
    title: 'AJO Coin 回贈',
    text: '付款完成後按已入賬金額估算 AJO Coin 回贈。',
    tone: 'highlight',
  },
]);

// 6. 載入首頁賬單資料
const loadBills = async (): Promise<void> => {
  isLoading.value = true;
  needsProfile.value = false;
  needsIsmart.value = false;
  try {
    const response = await fetchPOSPaymentBills();
    rows.value = response.data.data.items ?? [];
    context.value = response.data.data.context;
  } catch (error: unknown) {
    rows.value = [];
    context.value = undefined;
    const response = (error as { response?: { status?: number; data?: { code?: string; message?: string } } }).response;
    const message = String(response?.data?.message ?? '').toLowerCase();
    needsProfile.value = response?.status === 400 || message.includes('profile') || message.includes('residence');
    needsIsmart.value = response?.status === 401 || message.includes('ismart');
    if (!needsProfile.value && !needsIsmart.value) {
      feedbackStore.pushToast('AJO Pay 賬單載入失敗。', 'error');
    }
  } finally {
    isLoading.value = false;
  }
};

// 7. 處理主要行動
const handlePrimaryAction = async (): Promise<void> => {
  if (needsProfile.value || needsIsmart.value) {
    await router.push('/account/profile');
    return;
  }
  await router.push('/payments/pay');
};

onMounted(() => {
  void loadBills();
});
</script>

<template>
  <main class="ajo-pay-home">
    <section class="ajo-pay-home__hero">
      <div class="ajo-pay-home__hero-copy">
        <div class="ajo-pay-home__badge">香港智慧物業支付平台</div>
        <h1>繳費，從此<br>變成一種投資。</h1>
        <p>
          用 AJO PAY 繳交大廈管理費及其他費用，日常生活每筆開支，筆筆賺取 AJO Coin 回贈。
        </p>
        <button
          type="button"
          class="ajo-pay-home__primary"
          :disabled="isLoading"
          @click="handlePrimaryAction"
        >
          {{ isLoading ? '載入中' : primaryActionLabel }}
          <span aria-hidden="true">→</span>
        </button>
      </div>

      <div class="ajo-pay-home__summary">
        <article class="ajo-pay-home__stat-card">
          <span>大廈管理費</span>
          <strong>{{ formatHKD(managementAmount) }}</strong>
          <small>{{ contextLabel }}</small>
        </article>
        <article class="ajo-pay-home__stat-card">
          <span>其他費用</span>
          <strong>{{ formatHKD(otherAmount) }}</strong>
          <small>{{ payableRows.length }} 筆待繳項目</small>
        </article>
        <article class="ajo-pay-home__stat-card ajo-pay-home__stat-card--accent">
          <div>
            <span>AJO COIN 回贈</span>
            <strong>+{{ rewardCoins }} Coins</strong>
            <small>本次預計賺取</small>
          </div>
          <b>A</b>
        </article>
      </div>
    </section>

    <section class="ajo-pay-home__metrics">
      <article>
        <strong>{{ payableRows.length.toLocaleString('en-HK') }}</strong>
        <span>目前待繳項目</span>
      </article>
      <article>
        <strong>1%</strong>
        <span>每次付款回贈率</span>
      </article>
      <article>
        <strong>{{ formatHKD(totalAmount) }}</strong>
        <span>目前待繳總額</span>
      </article>
    </section>

    <section class="ajo-pay-home__features">
      <article
        v-for="item in featureCards"
        :key="item.title"
        class="ajo-pay-home__feature"
        :class="`ajo-pay-home__feature--${item.tone}`"
      >
        <span aria-hidden="true"></span>
        <strong>{{ item.title }}</strong>
        <p>{{ item.text }}</p>
      </article>
    </section>

    <footer class="ajo-pay-home__footer">
      <div>
        <span>AJO</span>
        <b>PAY</b>
      </div>
      <nav aria-label="AJO Pay footer">
        <RouterLink to="/account/profile">會員中心</RouterLink>
        <RouterLink to="/payments/pay">付款頁面</RouterLink>
      </nav>
    </footer>
  </main>
</template>

<style scoped>
.ajo-pay-home {
  --pay-brand: #f26419;
  --pay-brand-dark: #c44f0e;
  --pay-brand-light: #fff3eb;
  --pay-brand-mid: #fda96a;
  --pay-ink: #12122a;
  --pay-ink-2: #454566;
  --pay-ink-3: #8888aa;
  --pay-surface: #ffffff;
  --pay-surface-2: #f6f6fb;
  --pay-border: #e2e2ee;
  --pay-border-2: #cacade;
  display: grid;
  min-height: calc(100vh - var(--nav-h, 52px));
  background: var(--pay-surface-2);
  color: var(--pay-ink);
}

.ajo-pay-home__hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 420px;
  align-items: center;
  gap: 60px;
  background: linear-gradient(140deg, #ebf4ff, var(--pay-surface-2));
  padding: 48px 40px;
}

.ajo-pay-home__badge {
  display: inline-flex;
  align-items: center;
  border-radius: 100px;
  background: var(--pay-brand-light);
  color: var(--pay-brand);
  padding: 5px 12px;
  font-size: 12px;
  font-weight: 600;
  line-height: 1.3;
}

.ajo-pay-home__hero-copy h1 {
  margin: 20px 0 0;
  color: var(--pay-ink);
  font-family: 'Noto Sans TC', var(--font-sans);
  font-size: 44px;
  font-weight: 700;
  line-height: 1.15;
}

.ajo-pay-home__hero-copy p {
  max-width: 480px;
  margin: 16px 0 0;
  color: var(--pay-ink-2);
  font-size: 16px;
  line-height: 1.7;
}

.ajo-pay-home__primary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 46px;
  border: 0;
  border-radius: 9px;
  background: var(--pay-brand);
  color: #ffffff;
  cursor: pointer;
  font-family: inherit;
  font-size: 14px;
  font-weight: 600;
  line-height: 1;
  margin-top: 28px;
  padding: 13px 28px;
  transition: background 0.15s ease;
}

.ajo-pay-home__primary:hover {
  background: var(--pay-brand-dark);
}

.ajo-pay-home__primary:disabled {
  cursor: not-allowed;
  opacity: 0.65;
}

.ajo-pay-home__summary {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.ajo-pay-home__stat-card {
  display: grid;
  gap: 8px;
  min-height: 132px;
  border: 0.5px solid var(--pay-border);
  border-radius: 14px;
  background: var(--pay-surface);
  padding: 20px;
}

.ajo-pay-home__stat-card span {
  color: var(--pay-ink-3);
  font-size: 11px;
  line-height: 1.3;
}

.ajo-pay-home__stat-card strong {
  color: var(--pay-ink);
  font-size: 22px;
  font-weight: 700;
  line-height: 1.15;
}

.ajo-pay-home__stat-card small {
  color: var(--pay-ink-3);
  font-size: 12px;
  line-height: 1.45;
}

.ajo-pay-home__stat-card--accent {
  grid-column: 1 / -1;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  background: linear-gradient(135deg, #1a1a2e, #2a1a10);
  border-color: transparent;
}

.ajo-pay-home__stat-card--accent span {
  color: var(--pay-brand-mid);
}

.ajo-pay-home__stat-card--accent strong {
  color: #ffffff;
}

.ajo-pay-home__stat-card--accent small {
  color: rgba(255, 255, 255, 0.48);
}

.ajo-pay-home__stat-card--accent b {
  color: var(--pay-brand);
  font-size: 44px;
  font-weight: 800;
}

.ajo-pay-home__metrics {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  border-top: 0.5px solid var(--pay-border);
  border-bottom: 0.5px solid var(--pay-border);
  background: var(--pay-surface);
}

.ajo-pay-home__metrics article {
  display: grid;
  justify-items: center;
  gap: 6px;
  border-right: 0.5px solid var(--pay-border);
  padding: 40px;
  text-align: center;
}

.ajo-pay-home__metrics article:last-child {
  border-right: 0;
}

.ajo-pay-home__metrics strong {
  color: var(--pay-brand);
  font-size: 36px;
  font-weight: 800;
  line-height: 1.1;
}

.ajo-pay-home__metrics span {
  color: var(--pay-ink-3);
  font-size: 13px;
  line-height: 1.4;
}

.ajo-pay-home__features {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 20px;
  padding: 60px 40px;
}

.ajo-pay-home__feature {
  display: grid;
  align-content: start;
  gap: 8px;
  border: 0.5px solid var(--pay-border);
  border-radius: 14px;
  background: var(--pay-surface);
  padding: 28px;
}

.ajo-pay-home__feature span {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  margin-bottom: 8px;
}

.ajo-pay-home__feature--blue span {
  background: #eef3ff;
}

.ajo-pay-home__feature--orange span {
  background: var(--pay-brand-light);
}

.ajo-pay-home__feature--highlight {
  border-color: var(--pay-brand-mid);
  background: linear-gradient(135deg, #fff8f2, var(--pay-brand-light));
}

.ajo-pay-home__feature--highlight span {
  background: rgba(242, 100, 25, 0.12);
}

.ajo-pay-home__feature strong {
  color: var(--pay-ink);
  font-family: 'Noto Sans TC', var(--font-sans);
  font-size: 16px;
  font-weight: 600;
  line-height: 1.35;
}

.ajo-pay-home__feature--highlight strong {
  color: var(--pay-brand);
}

.ajo-pay-home__feature p {
  margin: 0;
  color: var(--pay-ink-2);
  font-size: 13px;
  line-height: 1.7;
}

.ajo-pay-home__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  border-top: 0.5px solid var(--pay-border);
  background: var(--pay-surface);
  padding: 20px 40px;
}

.ajo-pay-home__footer div {
  display: inline-flex;
  align-items: baseline;
  gap: 4px;
}

.ajo-pay-home__footer span {
  color: var(--pay-ink);
  font-size: 14px;
  font-weight: 800;
}

.ajo-pay-home__footer b {
  color: var(--pay-brand);
  font-size: 14px;
  font-weight: 400;
  letter-spacing: 2px;
}

.ajo-pay-home__footer nav {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
}

.ajo-pay-home__footer a {
  color: var(--pay-ink-3);
  font-size: 12px;
  text-decoration: none;
}

.ajo-pay-home__footer a:hover {
  color: var(--pay-ink);
}

@media (max-width: 960px) {
  .ajo-pay-home__hero,
  .ajo-pay-home__features,
  .ajo-pay-home__metrics {
    grid-template-columns: 1fr;
  }

  .ajo-pay-home__hero {
    gap: 28px;
    padding: 32px 20px;
  }

  .ajo-pay-home__hero-copy h1 {
    font-size: 34px;
  }

  .ajo-pay-home__metrics article {
    border-right: 0;
    border-bottom: 0.5px solid var(--pay-border);
    padding: 28px 20px;
  }

  .ajo-pay-home__metrics article:last-child {
    border-bottom: 0;
  }

  .ajo-pay-home__features {
    padding: 32px 20px;
  }

  .ajo-pay-home__footer {
    align-items: flex-start;
    flex-direction: column;
    padding: 18px 20px 96px;
  }
}
</style>
