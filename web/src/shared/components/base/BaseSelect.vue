<!--
 * 自定義選擇器元件。
 * 1. 提供統一的原生 select 外觀。
 * 2. 支援字串與數值選項。
-->
<script setup lang="ts">
interface SelectOption {
  label: string;
  value: string | number;
}

interface BaseSelectProps {
  modelValue: string | number;
  options: SelectOption[];
  disabled?: boolean;
}

const props = withDefaults(defineProps<BaseSelectProps>(), {
  disabled: false,
});

const emit = defineEmits<{
  (event: 'update:modelValue', value: string | number): void;
}>();

// 1. 同步選取值
const handleChange = (event: Event) => {
  const nextValue = (event.target as HTMLSelectElement).value;
  const matchedOption = props.options.find((item) => String(item.value) === nextValue);

  emit('update:modelValue', matchedOption ? matchedOption.value : nextValue);
};
</script>

<template>
  <select
    :value="props.modelValue"
    :disabled="props.disabled"
    class="app-select"
    @change="handleChange"
  >
    <option
      v-for="option in props.options"
      :key="String(option.value)"
      :value="option.value"
    >
      {{ option.label }}
    </option>
  </select>
</template>
