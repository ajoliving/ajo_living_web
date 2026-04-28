<!--
 * 二手交易商品收藏頁。
 * 1. 承載未來收藏商品列表。
 * 2. 保持我的中心右側內容區樣式一致。
-->
<script setup lang="ts">
import { useI18n } from 'vue-i18n';

import BaseEmpty from '@/shared/components/feedback/BaseEmpty.vue';

const { t } = useI18n();
</script>

<template>
  <BaseEmpty
    :title="t('marketplace.myHub.favoritesTitle')"
    :description="t('marketplace.myHub.favoritesDescription')"
  />
</template>
