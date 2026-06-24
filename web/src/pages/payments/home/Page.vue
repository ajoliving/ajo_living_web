<!--
 * AJO PAY 首頁。
 * 1. 展示目前綁定單位、待繳賬單、訂單與積分回贈摘要。
 * 2. 提供前往賬單、購物車、訂單、歷史與個人資料綁定入口。
-->
<script setup lang="ts">
import { computed } from 'vue';
import { RouterLink } from 'vue-router';

import { useSessionStore } from '@/app/stores/session';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import BaseButton from '@/shared/components/base/BaseButton.vue';

import { usePOSPaymentPage } from '../composables/usePOSPaymentPage';

const {
  contextLabel,
  isLoading,
  overview,
  posLoginRequired,
  profileRequired,
  reload,
} = usePOSPaymentPage('bills');

const sessionStore = useSessionStore();

const summary = computed(() => overview.value?.summary);
const pendingBillCount = computed(() => summary.value?.pending_bill_count ?? 0);
const pendingBillAmount = computed(() => summary.value?.pending_bill_amount ?? 0);
const pendingOrderCount = computed(() => summary.value?.pending_order_count ?? 0);
const abnormalOrderCount = computed(() => summary.value?.abnormal_order_count ?? 0);
const pointBalance = computed(() => sessionStore.me?.ajo_balance ?? 0);
const rewardPreview = computed(() => Math.floor(pendingBillAmount.value));
const needsSetup = computed(() => profileRequired.value || posLoginRequired.value);
</script>

<template>
  <main class="pay-home">
    <section class="pay-home__hero">
      <div class="pay-home__hero-copy">
        <p>AJO PAY</p>
        <h1>物業繳費</h1>
        <span>{{ needsSetup ? '完成帳戶與單位設定後即可查詢賬單。' : contextLabel }}</span>
        <div class="pay-home__actions">
          <RouterLink
            v-if="profileRequired"
            to="/account/profile/info"
          >
            <BaseButton
              variant="primary"
              size="md"
            >
              設定綁定單位
            </BaseButton>
          </RouterLink>
          <RouterLink
            v-else
            to="/payments/bills"
          >
            <BaseButton
              variant="primary"
              size="md"
            >
              查看賬單
            </BaseButton>
          </RouterLink>
          <BaseButton
            variant="secondary"
            size="md"
            :disabled="isLoading"
            @click="reload"
          >
            {{ isLoading ? '載入中' : '刷新' }}
          </BaseButton>
        </div>
      </div>

      <div class="pay-home__summary">
        <div>
          <span>待繳金額</span>
          <strong>HKD {{ pendingBillAmount.toFixed(2) }}</strong>
        </div>
        <div>
          <span>AJO Points</span>
          <strong>{{ pointBalance.toLocaleString() }}</strong>
        </div>
      </div>
    </section>

    <section class="pay-home__stats">
      <article>
        <span>待繳賬單</span>
        <strong>{{ pendingBillCount }}</strong>
      </article>
      <article>
        <span>待完成訂單</span>
        <strong>{{ pendingOrderCount }}</strong>
      </article>
      <article>
        <span>需跟進訂單</span>
        <strong>{{ abnormalOrderCount }}</strong>
      </article>
    </section>

    <section
      v-if="needsSetup"
      class="pay-home__notice"
    >
      <p>{{ profileRequired ? '請先到個人資料設定綁定單位。' : '請先綁定或更新 ismart 帳戶。' }}</p>
      <RouterLink to="/account/profile/info">
        <BaseButton
          variant="primary"
          size="sm"
        >
          前往個人資料
        </BaseButton>
      </RouterLink>
    </section>

    <section class="pay-home__grid">
      <RouterLink
        class="pay-home__entry"
        to="/payments/bills"
      >
        <span class="pay-home__entry-icon">
          <AppIcon
            name="wallet"
            :size="20"
          />
        </span>
        <strong>賬單</strong>
        <small>查看待繳項目並加入購物車</small>
      </RouterLink>
      <RouterLink
        class="pay-home__entry"
        to="/payments/cart"
      >
        <span class="pay-home__entry-icon">
          <AppIcon
            name="inbox"
            :size="20"
          />
        </span>
        <strong>購物車</strong>
        <small>確認賬單與支付方式</small>
      </RouterLink>
      <RouterLink
        class="pay-home__entry"
        to="/payments/orders"
      >
        <span class="pay-home__entry-icon">
          <AppIcon
            name="clock"
            :size="20"
          />
        </span>
        <strong>訂單</strong>
        <small>查看線上支付狀態</small>
      </RouterLink>
    </section>

    <section class="pay-home__points">
      <div>
        <span>AJO Coin 回贈</span>
        <strong>支付成功後按實付金額回贈 AJO Points</strong>
        <p>本次待繳金額預計可回贈 {{ rewardPreview.toLocaleString() }} AJO Points，實際以支付成功記錄為準。</p>
      </div>
      <RouterLink to="/account/profile/wallet">
        <BaseButton
          variant="secondary"
          size="md"
        >
          查看積分
        </BaseButton>
      </RouterLink>
    </section>
  </main>
</template>

<style scoped>
.pay-home {
  display: grid;
  gap: 1rem;
  color: rgb(var(--color-text));
}

.pay-home__hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 22rem;
  gap: 1.5rem;
  align-items: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: linear-gradient(135deg, rgb(var(--color-surface)), rgb(var(--color-primary-soft) / 0.22));
  padding: clamp(1.25rem, 4vw, 2.5rem);
}

.pay-home__hero-copy p,
.pay-home__points span {
  margin: 0;
  color: rgb(var(--color-primary));
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.pay-home__hero-copy h1 {
  margin: 0.75rem 0 0;
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: clamp(2.1rem, 6vw, 3.5rem);
  font-weight: 500;
  line-height: 1.08;
}

.pay-home__hero-copy span {
  display: block;
  margin-top: 0.9rem;
  max-width: 34rem;
  color: rgb(var(--color-text-muted));
  font-size: 1rem;
  font-weight: 700;
  line-height: 1.6;
}

.pay-home__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-top: 1.4rem;
}

.pay-home__summary {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
}

.pay-home__summary div,
.pay-home__stats article,
.pay-home__entry,
.pay-home__points,
.pay-home__notice {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
}

.pay-home__summary div {
  display: grid;
  gap: 0.45rem;
  min-height: 7rem;
  align-content: center;
  padding: 1rem;
}

.pay-home__summary span,
.pay-home__stats span {
  color: rgb(var(--color-text-muted));
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  line-height: 1;
  text-transform: uppercase;
}

.pay-home__summary strong {
  color: rgb(var(--color-text));
  font-size: 1.35rem;
  font-weight: 800;
  line-height: 1.2;
}

.pay-home__stats {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1rem;
}

.pay-home__stats article {
  display: grid;
  gap: 0.6rem;
  padding: 1.1rem;
}

.pay-home__stats strong {
  color: rgb(var(--color-primary));
  font-size: 1.6rem;
  font-weight: 800;
  line-height: 1;
}

.pay-home__notice {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1rem;
}

.pay-home__notice p {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.9rem;
  font-weight: 700;
  line-height: 1.5;
}

.pay-home__grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1rem;
}

.pay-home__entry {
  display: grid;
  gap: 0.55rem;
  padding: 1rem;
  color: rgb(var(--color-text));
  text-decoration: none;
  transition:
    border-color 0.2s ease,
    transform 0.2s ease;
}

.pay-home__entry:hover {
  border-color: rgb(var(--color-primary) / 0.42);
  transform: translateY(-1px);
}

.pay-home__entry-icon {
  display: inline-grid;
  width: 2.5rem;
  height: 2.5rem;
  place-items: center;
  border-radius: 8px;
  background: rgb(var(--color-primary-soft) / 0.28);
  color: rgb(var(--color-primary));
}

.pay-home__entry strong {
  font-size: 1rem;
  font-weight: 800;
  line-height: 1.2;
}

.pay-home__entry small {
  color: rgb(var(--color-text-muted));
  font-size: 0.82rem;
  font-weight: 650;
  line-height: 1.5;
}

.pay-home__points {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.15rem;
}

.pay-home__points strong {
  display: block;
  margin-top: 0.55rem;
  color: rgb(var(--color-text));
  font-size: 1rem;
  font-weight: 800;
  line-height: 1.35;
}

.pay-home__points p {
  margin: 0.45rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.88rem;
  font-weight: 650;
  line-height: 1.5;
}

@media (max-width: 900px) {
  .pay-home__hero,
  .pay-home__summary,
  .pay-home__stats,
  .pay-home__grid {
    grid-template-columns: 1fr;
  }

  .pay-home__points,
  .pay-home__notice {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>
