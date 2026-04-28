<!--
 * 二手交易個人資料入口頁。
 * 1. 提供我的中心內的個人資料入口。
 * 2. 引導使用者前往帳戶資料頁維護真實會員資料。
-->
<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { RouterLink } from 'vue-router';

import BaseEmpty from '@/shared/components/feedback/BaseEmpty.vue';

const { t } = useI18n();
</script>

<template>
  <BaseEmpty
    :title="t('marketplace.myHub.profileTitle')"
    :description="t('marketplace.myHub.profileDescription')"
  >
    <RouterLink
      to="/account/profile"
      class="marketplace-profile-link"
    >
      {{ t('marketplace.myHub.profileAction') }}
    </RouterLink>
  </BaseEmpty>
</template>

<style scoped>
.marketplace-profile-link {
  display: inline-flex;
  min-height: 2.75rem;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: #002727;
  padding: 0.75rem 1.15rem;
  color: #ffffff;
  font-size: 0.875rem;
  font-weight: 800;
  line-height: 1;
}
</style>
