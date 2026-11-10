<!--
 * 自定義分段切換元件。
 * 1. 提供列表狀態與可見性等二選或多選切換。
 * 2. 維持一致的高保真切換樣式。
-->
<script setup lang="ts">
interface SegmentedOption {
  label: string;
  value: string;
}

interface BaseSegmentedProps {
  modelValue: string;
  options: SegmentedOption[];
  block?: boolean;
}

const props = withDefaults(defineProps<BaseSegmentedProps>(), {
  block: false,
});

const emit = defineEmits<{
  (event: 'update:modelValue', value: string): void;
}>();

// 1. 更新目前選項
const handleSelect = (value: string) => {
  emit('update:modelValue', value);
};
</script>

<template>
  <div
    class="app-segmented"
    :class="props.block ? 'flex w-full' : ''"
  >
    <button
      v-for="option in props.options"
      :key="option.value"
      type="button"
      class="app-segmented-item"
      :class="[
        props.modelValue === option.value ? 'app-segmented-item-active' : '',
        props.block ? 'flex-1' : '',
      ]"
      @click="handleSelect(option.value)"
    >
      {{ option.label }}
    </button>
  </div>
</template>
