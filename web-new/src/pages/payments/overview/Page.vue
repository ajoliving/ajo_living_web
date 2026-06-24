<!--
 * 支付中心首頁。
 * 1. 呈現本月待繳、費用拆分與最近繳費記錄。
 * 2. 對齊支付工作台的正式標題、卡片與表格密度。
 * 3. 保持首頁獨立於各子功能頁資料流程。
-->
<script setup lang="ts">
import BaseButton from '@/shared/components/base/BaseButton.vue';

const feeItems = [
  { label: '管理費', value: 'HK$2,800' },
  { label: '水費', value: 'HK$480' },
  { label: '電費', value: 'HK$1,200' },
  { label: '其他費用', value: 'HK$200' },
];

const paymentRecords = [
  { date: '2025/05/15', item: '5月管理費', amount: 'HK$2,800', status: '已繳', tone: 'ok' },
  { date: '2025/05/12', item: '4月水費', amount: 'HK$450', status: '已繳', tone: 'ok' },
  { date: '2025/05/10', item: '4月電費', amount: 'HK$1,150', status: '已繳', tone: 'ok' },
  { date: '2025/04/15', item: '4月管理費', amount: 'HK$2,800', status: '已繳', tone: 'ok' },
  { date: '2025/04/03', item: '停車場月費', amount: 'HK$1,200', status: '待確認', tone: 'pending' },
];
</script>

<template>
  <main class="pos-page pos-overview">
    <section class="pos-page__header">
      <div>
        <p>Payments / Overview</p>
        <h1>概覽</h1>
        <span>本月待繳 HK$4,680 · 最近繳費 5 筆</span>
      </div>
      <BaseButton
        variant="primary"
        size="sm"
      >
        立即繳費
      </BaseButton>
    </section>

    <section class="pos-page__panel pos-overview__summary-panel">
      <div class="pos-section-head">
        <div>
          <strong>本月待繳</strong>
          <span>HK$4,680</span>
        </div>
      </div>

      <div class="pos-overview__summary-row">
        <div class="pos-overview__summary-amount">
          <strong>HK$4,680</strong>
          <span>到期日：2025年6月15日</span>
        </div>
        <p class="pos-overview__summary-note">
          已入賬項目以最新賬單為準。
        </p>
      </div>

      <div class="pos-overview__stats-grid">
        <article
          v-for="item in feeItems"
          :key="item.label"
          class="pos-overview__stat-card"
        >
          <span>{{ item.label }}</span>
          <strong>{{ item.value }}</strong>
        </article>
      </div>
    </section>

    <section class="pos-page__panel">
      <div class="pos-section-head">
        <div>
          <strong>最近繳費記錄</strong>
          <span>5 筆</span>
        </div>
      </div>

      <div class="pos-table-wrap">
        <table class="pos-table pos-table--overview">
          <thead>
            <tr>
              <th>日期</th>
              <th>項目</th>
              <th>金額</th>
              <th>狀態</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="record in paymentRecords"
              :key="`${record.date}-${record.item}`"
            >
              <td class="pos-table__nowrap">{{ record.date }}</td>
              <td>{{ record.item }}</td>
              <td class="pos-table__amount">{{ record.amount }}</td>
              <td>
                <span
                  class="pos-overview__status"
                  :class="`pos-overview__status--${record.tone}`"
                >
                  {{ record.status }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </main>
</template>

<style scoped>
@import '../payment-page.css';

.pos-overview {
  display: grid;
  gap: 1rem;
}

.pos-overview__summary-panel {
  gap: 0.9rem;
}

.pos-overview__summary-row {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 1rem;
}

.pos-overview__summary-amount {
  display: grid;
  align-content: start;
  gap: 0.35rem;
}

.pos-overview__summary-amount strong {
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 1.5rem;
  font-weight: 500;
  line-height: 1.1;
}

.pos-overview__summary-amount span,
.pos-overview__summary-note {
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 700;
  line-height: 1.5;
}

.pos-overview__summary-note {
  max-width: 22rem;
  margin: 0;
}

.pos-overview__stats-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.75rem;
}

.pos-overview__stat-card {
  display: grid;
  gap: 0.3rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface-raised));
  padding: 0.9rem 1rem;
}

.pos-overview__stat-card span {
  color: rgb(var(--color-text-muted));
  font-size: 0.72rem;
  font-weight: 800;
  line-height: 1.2;
}

.pos-overview__stat-card strong {
  color: rgb(var(--color-text));
  font-size: 0.98rem;
  font-weight: 750;
  line-height: 1.35;
}

.pos-overview__status {
  display: inline-flex;
  min-height: 1.65rem;
  align-items: center;
  border-radius: 999px;
  padding: 0 0.6rem;
  font-size: 0.76rem;
  font-weight: 800;
  line-height: 1;
}

.pos-overview__status--ok {
  background: rgb(var(--color-success) / 0.12);
  color: rgb(var(--color-success));
}

.pos-overview__status--pending {
  background: rgb(var(--color-warning) / 0.12);
  color: rgb(var(--color-warning));
}

@media (max-width: 900px) {
  .pos-overview__stats-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .pos-overview__stats-grid {
    grid-template-columns: 1fr;
  }

  .pos-overview__summary-row {
    align-items: flex-start;
  }
}
</style>
