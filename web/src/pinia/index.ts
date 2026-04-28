/*
 * Pinia 實例出口。
 * 1. 集中建立前端全域狀態容器。
 * 2. 供 Router、App 與 HTTP 層共用。
 */
import { createPinia } from 'pinia';

// 1. 建立唯一 Pinia 實例
export const pinia = createPinia();
