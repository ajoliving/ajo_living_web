<!--
 * 會員中心手機返回按鈕。
 * 1. 在會員中心面板標題旁返回功能首頁。
-->
<script setup lang="ts">
import { inject } from 'vue';
import { useI18n } from 'vue-i18n';
import AppIcon from '@/shared/components/base/AppIcon.vue';

const returnToHome = inject<(() => void) | null>('accountReturnToHome', null)
  ?? inject<(() => void) | null>('managementReturnToHome', null);
const props = withDefaults(defineProps<{ labelKey?: string }>(), {
  labelKey: 'account.center.nav.backToHome',
});
const { t } = useI18n();
</script>

<template>
  <button v-if="returnToHome" type="button" class="account-mobile-back" :aria-label="t(props.labelKey)" @click="returnToHome">
    <AppIcon name="chevron-down" :size="20" />
  </button>
</template>

<style scoped>
.account-mobile-back {
  display: none;
}

@media (max-width: 1023px) {
  .account-mobile-back {
    display: inline-flex;
    width: 40px;
    min-height: 44px;
    align-items: center;
    justify-content: center;
    flex: 0 0 auto;
    border: 0;
    background: transparent;
    color: inherit;
    cursor: pointer;
    padding: 0;
    vertical-align: middle;
  }

  .account-mobile-back :deep(.app-icon) {
    transform: rotate(90deg);
  }
}
</style>
