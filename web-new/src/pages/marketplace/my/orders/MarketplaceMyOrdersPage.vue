<!--
 * 二手交易我的訂單頁。
 * 1. 提供會員中心內的交易訂單工作頁外觀。
 * 2. 先展示近期交易與付款狀態，保持與設計稿一致的正式資訊層級。
-->
<script setup lang="ts">
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
</script>

<template>
  <section class="orders-page">
    <header class="orders-header">
      <div>
        <p class="orders-kicker">Orders</p>
        <h1>{{ t('marketplace.myHub.ordersTitle') }}</h1>
        <p>{{ t('marketplace.myHub.ordersDescription') }}</p>
      </div>
    </header>

    <section class="orders-metrics">
      <article class="orders-metric-card">
        <strong>3</strong>
        <span>近期交易</span>
      </article>
      <article class="orders-metric-card">
        <strong>1</strong>
        <span>待處理</span>
      </article>
      <article class="orders-metric-card">
        <strong>2</strong>
        <span>已完成</span>
      </article>
    </section>

    <section class="orders-panel">
      <div class="orders-panel__title">
        <h2>訂單列表</h2>
      </div>

      <div class="orders-table-wrap">
        <table class="orders-table">
          <thead>
            <tr>
              <th>訂單</th>
              <th>類型</th>
              <th>金額</th>
              <th>狀態</th>
              <th>日期</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>AJO-20260605-01</td>
              <td>管理費</td>
              <td>HK$2,850</td>
              <td><span class="orders-chip orders-chip--warn">待繳</span></td>
              <td>2026年6月5日</td>
            </tr>
            <tr>
              <td>AJO-20260522-03</td>
              <td>樓盤發布</td>
              <td>HK$1,000</td>
              <td><span class="orders-chip orders-chip--good">已完成</span></td>
              <td>2026年5月22日</td>
            </tr>
            <tr>
              <td>AJO-20260510-02</td>
              <td>家具交易</td>
              <td>HK$2,400</td>
              <td><span class="orders-chip">處理中</span></td>
              <td>2026年5月10日</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </section>
</template>

<style scoped>
.orders-page {
  display: grid;
  gap: 16px;
}

.orders-header {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 16px;
  border-bottom: 1px solid rgb(var(--color-border));
  padding-bottom: 16px;
}

.orders-kicker {
  margin: 0;
  color: rgb(var(--color-primary));
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.orders-header h1 {
  margin: 6px 0 0;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: clamp(1.85rem, 2.6vw, 2.35rem);
  font-weight: 400;
  line-height: 1.15;
}

.orders-header p:not(.orders-kicker) {
  margin: 8px 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 13px;
  line-height: 1.6;
}

.orders-metrics {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.orders-metric-card,
.orders-panel {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
}

.orders-metric-card {
  display: grid;
  gap: 8px;
  min-height: 120px;
  padding: 16px;
}

.orders-metric-card strong {
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 30px;
  font-weight: 400;
  line-height: 1;
}

.orders-metric-card span {
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  font-weight: 700;
}

.orders-panel {
  display: grid;
  gap: 14px;
  padding: 16px;
}

.orders-panel__title {
  border-bottom: 1px solid rgb(var(--color-border));
  padding-bottom: 14px;
}

.orders-panel__title h2 {
  margin: 0;
  color: rgb(var(--color-primary));
  font-size: 14px;
  font-weight: 700;
}

.orders-table-wrap {
  overflow-x: auto;
}

.orders-table {
  width: 100%;
  min-width: 640px;
  border-collapse: collapse;
  font-size: 13px;
}

.orders-table th,
.orders-table td {
  border-top: 1px solid rgb(var(--color-border));
  padding: 12px 10px;
  text-align: left;
  vertical-align: middle;
}

.orders-table thead th {
  border-top: 0;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  font-weight: 700;
}

.orders-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 28px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 999px;
  background: rgb(var(--color-surface-raised));
  padding: 0 10px;
  color: rgb(var(--color-text));
  font-size: 11px;
  font-weight: 700;
}

.orders-chip--good {
  border-color: rgb(var(--color-success) / 0.2);
  background: rgb(var(--color-success) / 0.1);
  color: rgb(var(--color-success));
}

.orders-chip--warn {
  border-color: rgb(var(--color-warning) / 0.2);
  background: rgb(var(--color-warning) / 0.12);
  color: rgb(var(--color-warning));
}

@media (max-width: 767px) {
  .orders-metrics {
    grid-template-columns: 1fr;
  }
}
</style>
