<!--
 * 自定義按鈕元件。
 * 1. 統一主要、次要、幽靈與危險按鈕樣式。
 * 2. 提供尺寸、全寬與左右圖示插槽。
-->
<script setup lang="ts">
import { computed } from 'vue';

type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger';
type ButtonSize = 'sm' | 'md' | 'lg';

interface BaseButtonProps {
  variant?: ButtonVariant;
  size?: ButtonSize;
  type?: 'button' | 'submit' | 'reset';
  block?: boolean;
  disabled?: boolean;
}

const props = withDefaults(defineProps<BaseButtonProps>(), {
  variant: 'secondary',
  size: 'md',
  type: 'button',
  block: false,
  disabled: false,
});

// 1. 組合按鈕樣式
const buttonClassName = computed(() => {
  const variantClassMap: Record<ButtonVariant, string> = {
    primary: 'app-button-primary',
    secondary: 'app-button-secondary',
    ghost: 'app-button-ghost',
    danger: 'app-button-danger',
  };

  const sizeClassMap: Record<ButtonSize, string> = {
    sm: 'px-3 py-2 text-xs',
    md: 'px-4 py-2.5 text-sm',
    lg: 'px-5 py-3 text-base',
  };

  return [
    'app-button',
    variantClassMap[props.variant],
    sizeClassMap[props.size],
    props.block ? 'w-full' : '',
  ];
});
</script>

<template>
  <button
    :type="props.type"
    :disabled="props.disabled"
    :class="buttonClassName"
  >
    <slot name="leading" />
    <span>
      <slot />
    </span>
    <slot name="trailing" />
  </button>
</template>
