/*
 * 支付 domain 型別入口。
 * 1. 匯出訂單型別。
 * 2. 匯出錢包、積分與廣告任務型別。
 */

// 1. 匯出訂單型別
export * from './order-model';

// 2. 匯出錢包型別
export * from './wallet-model';

// 3. 匯出 POS 物業繳費型別
export * from './pos-model';
