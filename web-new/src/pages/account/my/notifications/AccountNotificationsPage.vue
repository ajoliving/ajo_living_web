<!--
 * 會員通知兼容跳轉頁。
 * 1. 將舊通知中心路徑導向聊天中的通知會話。
 * 2. 不承載獨立通知中心介面。
-->
<script setup lang="ts">
import { onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

const router = useRouter();
const { t } = useI18n();

// 1. 導向系統通知會話
onMounted(() => {
  void router.replace({
    path: '/account/marketplace/my/chat',
    query: { target: 'system_notice' },
  });
});
</script>

<template>
  <main class="section-shell py-8">
    <div class="panel-surface p-6 text-sm text-text-muted">
      {{ t('chat.loadingConversations') }}
    </div>
  </main>
</template>
