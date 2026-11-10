<!--
 * 自定義輸入框元件。
 * 1. 統一文字輸入外觀與焦點狀態。
 * 2. 支援前置圖示與多種原生輸入型別。
-->
<script setup lang="ts">
import type { HTMLAttributes, InputTypeHTMLAttribute } from 'vue';

interface BaseInputProps {
  modelValue: string | number;
  type?: InputTypeHTMLAttribute;
  placeholder?: string;
  inputmode?: HTMLAttributes['inputmode'];
  autocomplete?: string;
  readonly?: boolean;
  disabled?: boolean;
}

const props = withDefaults(defineProps<BaseInputProps>(), {
  type: 'text',
  placeholder: '',
  inputmode: undefined,
  autocomplete: 'off',
  readonly: false,
  disabled: false,
});

const emit = defineEmits<{
  (event: 'update:modelValue', value: string): void;
}>();

// 1. 同步輸入內容
const handleInput = (event: Event) => {
  emit('update:modelValue', (event.target as HTMLInputElement).value);
};
</script>

<template>
  <label class="app-input-shell">
    <slot name="leading" />
    <input
      :value="props.modelValue"
      :type="props.type"
      :placeholder="props.placeholder"
      :inputmode="props.inputmode"
      :autocomplete="props.autocomplete"
      :readonly="props.readonly"
      :disabled="props.disabled"
      class="app-input"
      @input="handleInput"
    />
  </label>
</template>
