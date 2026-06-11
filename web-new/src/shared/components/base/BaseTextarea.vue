<!--
 * 自定義多行輸入元件。
 * 1. 統一多行文字欄位外觀。
 * 2. 支援可配置列數與字數限制。
-->
<script setup lang="ts">
interface BaseTextareaProps {
  modelValue: string;
  placeholder?: string;
  rows?: number;
  maxlength?: number;
}

const props = withDefaults(defineProps<BaseTextareaProps>(), {
  placeholder: '',
  rows: 4,
  maxlength: undefined,
});

const emit = defineEmits<{
  (event: 'update:modelValue', value: string): void;
}>();

// 1. 同步文字區內容
const handleInput = (event: Event) => {
  emit('update:modelValue', (event.target as HTMLTextAreaElement).value);
};
</script>

<template>
  <textarea
    :value="props.modelValue"
    :rows="props.rows"
    :maxlength="props.maxlength"
    :placeholder="props.placeholder"
    class="app-textarea"
    @input="handleInput"
  />
</template>
