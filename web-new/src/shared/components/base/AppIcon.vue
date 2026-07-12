<!--
 * 全域圖示元件。
 * 1. 集中管理常用介面圖示。
 * 2. 以單一元件提供一致的描邊風格。
-->
<script setup lang="ts">
import { computed } from 'vue';

type IconName =
  | 'arrow-left'
  | 'arrow-right'
  | 'bell'
  | 'building'
  | 'browse'
  | 'check-circle'
  | 'chevron-down'
  | 'cloud-upload'
  | 'clock'
  | 'close'
  | 'filter'
  | 'globe'
  | 'home'
  | 'inbox'
  | 'layout-grid'
  | 'layout-list'
  | 'location'
  | 'lock'
  | 'login'
  | 'menu'
  | 'message'
  | 'palette'
  | 'phone'
  | 'picture'
  | 'plus-square'
  | 'reload'
  | 'search'
  | 'send'
  | 'shield'
  | 'star'
  | 'user'
  | 'view'
  | 'view-off'
  | 'wallet';

interface AppIconProps {
  name: IconName;
  size?: number;
  strokeWidth?: number;
}

const props = withDefaults(defineProps<AppIconProps>(), {
  size: 20,
  strokeWidth: 1.8,
});

const iconMap: Record<IconName, string> = {
  'arrow-left': 'M15 6 9 12l6 6M9 12h12',
  'arrow-right': 'M9 6l6 6-6 6M15 12H3',
  bell: 'M18 8a6 6 0 0 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9 M10 21h4',
  building: 'M4 21h16 M6 21V8l6-4 6 4v13 M9 12h.01 M15 12h.01 M9 16h.01 M15 16h.01',
  browse:
    'M4 5.5A1.5 1.5 0 0 1 5.5 4h13A1.5 1.5 0 0 1 20 5.5v13a1.5 1.5 0 0 1-1.5 1.5h-13A1.5 1.5 0 0 1 4 18.5z M4 10h16 M10 4v16',
  'check-circle':
    'M22 12c0 5.523-4.477 10-10 10S2 17.523 2 12 6.477 2 12 2s10 4.477 10 10z M8 12.5l2.5 2.5L16.5 9',
  'chevron-down': 'M6 9l6 6 6-6',
  'cloud-upload':
    'M16 16l-4-4-4 4 M12 12v9 M20.39 18.39A5 5 0 0 0 18 9h-1.26A8 8 0 1 0 3 16.3',
  clock: 'M12 7v5l3 2 M22 12c0 5.523-4.477 10-10 10S2 17.523 2 12 6.477 2 12 2s10 4.477 10 10z',
  close: 'M18 6 6 18 M6 6l12 12',
  filter: 'M4 6h16 M7 12h10 M10 18h4',
  globe:
    'M12 2c4.7 0 8.7 3.24 9.73 7.63H2.27C3.3 5.24 7.3 2 12 2z M2.27 14.37h19.46C20.7 18.76 16.7 22 12 22s-8.7-3.24-9.73-7.63z M9 2.6c-1.94 2.12-3.12 5.18-3.26 9.4.14 4.22 1.32 7.28 3.26 9.4 M15 2.6c1.94 2.12 3.12 5.18 3.26 9.4-.14 4.22-1.32 7.28-3.26 9.4',
  home: 'M3 10.5 12 3l9 7.5 M5 9.5V20h14V9.5',
  inbox: 'M4 5h16l-1 11H5z M9 11h6 M7 15h10',
  'layout-grid': 'M4 4h6v6H4z M14 4h6v6h-6z M4 14h6v6H4z M14 14h6v6h-6z',
  'layout-list': 'M8 6h12 M8 12h12 M8 18h12 M4 6h.01 M4 12h.01 M4 18h.01',
  location: 'M12 21s6-5.63 6-11a6 6 0 1 0-12 0c0 5.37 6 11 6 11z M12 12.5a2.5 2.5 0 1 0 0-5 2.5 2.5 0 0 0 0 5z',
  lock: 'M7 11V8a5 5 0 0 1 10 0v3 M6 11h12v9H6z',
  login: 'M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4 M10 17l5-5-5-5 M15 12H3',
  menu: 'M4 7h16 M4 12h16 M4 17h16',
  message:
    'M4 5h16v10H8l-4 4z',
  palette:
    'M12 3c5 0 9 3.58 9 8 0 2.67-2.1 4-4.2 4h-1.3c-.83 0-1.5.67-1.5 1.5 0 .57.33 1.08.33 1.67A1.83 1.83 0 0 1 12.5 20C7.25 20 3 16.42 3 11.9 3 7.04 7.03 3 12 3z M7.5 11h.01 M10.5 7.5h.01 M15.5 8.5h.01 M16.5 13h.01',
  phone:
    'M6.62 10.79a15.07 15.07 0 0 0 6.59 6.59l2.2-2.2a1 1 0 0 1 1.02-.24 11.36 11.36 0 0 0 3.57.57 1 1 0 0 1 1 1V20a1 1 0 0 1-1 1C10.3 21 3 13.7 3 4a1 1 0 0 1 1-1h3.5a1 1 0 0 1 1 1 11.36 11.36 0 0 0 .57 3.57 1 1 0 0 1-.24 1.02z',
  picture:
    'M4 5h16v14H4z M8 11a1.5 1.5 0 1 0 0-3 1.5 1.5 0 0 0 0 3z M20 15l-4.5-4.5L9 17',
  'plus-square': 'M12 8v8 M8 12h8 M5 4h14a1 1 0 0 1 1 1v14a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1V5a1 1 0 0 1 1-1z',
  reload: 'M20 6v6h-6 M4 18v-6h6 M6.93 9A8 8 0 0 1 20 12 M17.07 15A8 8 0 0 1 4 12',
  search: 'm21 21-4.35-4.35 M10.5 18a7.5 7.5 0 1 1 0-15 7.5 7.5 0 0 1 0 15z',
  send: 'M22 2 11 13 M22 2l-7 20-4-9-9-4z',
  shield: 'M12 3l7 3v6c0 5-3.5 8.5-7 9-3.5-.5-7-4-7-9V6z M9.5 12l1.8 1.8L15 10',
  star: 'm12 3 2.78 5.63 6.22.9-4.5 4.39 1.06 6.2L12 17.23 6.44 20.12l1.06-6.2L3 9.53l6.22-.9z',
  user: 'M20 21a8 8 0 1 0-16 0 M12 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8z',
  view: 'M2 12s3.5-6 10-6 10 6 10 6-3.5 6-10 6S2 12 2 12z M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z',
  'view-off': 'M3 3l18 18 M10.58 10.58A3 3 0 0 0 13.42 13.42 M9.88 4.24A10.65 10.65 0 0 1 12 4c6.5 0 10 8 10 8a18.33 18.33 0 0 1-2.19 3.12 M6.61 6.61C3.62 8.42 2 12 2 12s3.5 6 10 6a10.5 10.5 0 0 0 4.39-.95',
  wallet: 'M3 7.5A2.5 2.5 0 0 1 5.5 5h11A2.5 2.5 0 0 1 19 7.5V8h1a1 1 0 0 1 1 1v7a3 3 0 0 1-3 3H5.5A2.5 2.5 0 0 1 3 16.5z M17 13h.01 M3 9h18',
};

// 1. 取得目前圖示 path 定義
const iconPath = computed(() => iconMap[props.name]);
</script>

<template>
  <svg
    class="app-icon"
    :width="props.size"
    :height="props.size"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    :stroke-width="props.strokeWidth"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
  >
    <path :d="iconPath" />
  </svg>
</template>
