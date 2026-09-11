/*
 * 支付 domain API 入口。
 * 1. 匯出 AJO Point 錢包與廣告任務 API。
 * 2. 匯出 POS 繳費 API。
 */

// 1. 匯出錢包 API
export * from './wallet';

// 2. 匯出 POS 繳費 API
export * from './payments-pos';
