/*
 * 通訊 domain API 入口。
 * 1. 匯出聊天、訊息與通知 API。
 * 2. 為站內信、公告與 iBoard 接入提供穩定引用邊界。
 */

// 1. 匯出聊天會話 API
export * from './chats-api';

// 2. 匯出聊天訊息 API
export * from './messages-api';

// 3. 匯出通知 API
export * from './notifications-api';
