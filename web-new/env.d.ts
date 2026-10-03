/// <reference types="vite/client" />

/*
 * 前端建置期環境變數型別。
 * 1. 介面位址與 WebSocket 位址由部署腳本注入。
 * 2. 功能開關只在指定環境關閉，未配置時保持預設開放。
 */
interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string;
  readonly VITE_WS_BASE_URL?: string;
  readonly VITE_ENABLE_THEME_PICKER?: string;
}

declare module '*.vue' {
  import type { DefineComponent } from 'vue';

  const component: DefineComponent<Record<string, never>, Record<string, never>, unknown>;

  export default component;
}
