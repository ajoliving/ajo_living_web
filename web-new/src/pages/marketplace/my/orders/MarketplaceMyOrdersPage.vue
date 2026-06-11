<!--
 * 二手交易我的訂單頁。
 * 1. 承載未來訂單與交易紀錄。
 * 2. 保持我的中心右側內容區樣式一致。
-->
<script setup lang="ts">
import { useI18n } from 'vue-i18n';

import BaseEmpty from '@/shared/components/feedback/BaseEmpty.vue';

const { t } = useI18n();
</script>

<template>
  <BaseEmpty
    :title="t('marketplace.myHub.ordersTitle')"
    :description="t('marketplace.myHub.ordersDescription')"
  />
</template>
