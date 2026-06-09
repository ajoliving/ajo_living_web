<!--
 * 狀態與可見性標籤。
 * 1. 統一 active、expired、sold 與可見性樣式。
 * 2. 讓列表、詳情與我的發布頁可直接復用。
-->
<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

interface StatusBadgeProps {
  status?: string;
  visibility?: string;
}

const props = defineProps<StatusBadgeProps>();
const { t } = useI18n();

// 1. 組合需要輸出的標籤陣列
const chips = computed(() => {
  const values: Array<{ key: string; label: string; className: string }> = [];

  if (props.status) {
    const statusMap: Record<string, { label: string; className: string }> = {
      active: {
        label: t('common.state.active'),
        className: 'border-success/20 bg-success/10 text-success',
      },
      available: {
        label: t('common.state.available'),
        className: 'border-success/20 bg-success/10 text-success',
      },
      expired: {
        label: t('common.state.expired'),
        className: 'border-warning/20 bg-warning/10 text-warning',
      },
      sold: {
        label: t('common.state.sold'),
        className: 'border-text-muted/20 bg-text-muted/10 text-text-muted',
      },
      draft: {
        label: t('common.state.draft'),
        className: 'border-border bg-surface-raised text-text-muted',
      },
      hidden: {
        label: t('common.state.hidden'),
        className: 'border-border bg-surface-raised text-text-muted',
      },
      reserved: {
        label: t('common.state.reserved'),
        className: 'border-warning/20 bg-warning/10 text-warning',
      },
    };

    if (statusMap[props.status]) {
      values.push({
        key: `status-${props.status}`,
        ...statusMap[props.status],
      });
    }
  }

  if (props.visibility) {
    const visibilityMap: Record<string, { label: string; className: string }> = {
      public: {
        label: t('common.state.public'),
        className: 'border-primary/20 bg-primary/10 text-primary',
      },
      building_only: {
        label: t('common.state.buildingOnly'),
        className: 'border-warning/20 bg-warning/10 text-warning',
      },
    };

    if (visibilityMap[props.visibility]) {
      values.push({
        key: `visibility-${props.visibility}`,
        ...visibilityMap[props.visibility],
      });
    }
  }

  return values;
});
</script>

<template>
  <div class="flex flex-wrap gap-2">
    <span
      v-for="chip in chips"
      :key="chip.key"
      class="status-pill"
      :class="chip.className"
    >
      {{ chip.label }}
    </span>
  </div>
</template>
