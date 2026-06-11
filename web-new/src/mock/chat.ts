/*
 * 聊天 Mock 資料。
 * 1. 提供會話列表與訊息內容。
 * 2. 支援聊天頁輸入互動與詳情頁聊天入口演示。
 */
import type { ChatConversation, Message } from '@/model/chat';

import { marketplaceListings } from '@/mock/listings';
import { sellerProfiles } from '@/mock/user';

const [carmenLee, ryanHo] = sellerProfiles;
const [loungeSet, , dehumidifier] = marketplaceListings;

// 1. 定義聊天會話列表
export const chatConversations: ChatConversation[] = [
  {
    id: 'chat-101',
    listing: {
      id: loungeSet.id,
      title: loungeSet.title,
      price_hkd: loungeSet.price_hkd,
      images: loungeSet.images,
      status: loungeSet.status,
    },
    peer: carmenLee,
    unread_count: 2,
    last_message: 'Saturday 11am pickup works for me.',
    last_message_at: '2026-04-16T08:10:00.000Z',
  },
  {
    id: 'chat-102',
    listing: {
      id: dehumidifier.id,
      title: dehumidifier.title,
      price_hkd: dehumidifier.price_hkd,
      images: dehumidifier.images,
      status: dehumidifier.status,
    },
    peer: ryanHo,
    unread_count: 0,
    last_message: 'Marked as sold yesterday, thank you.',
    last_message_at: '2026-04-15T19:40:00.000Z',
  },
];

// 2. 定義每個會話的訊息內容
export const chatMessages: Record<string, Message[]> = {
  'chat-101': [
    {
      id: 'msg-101-1',
      chat_id: 'chat-101',
      sender_role: 'peer',
      body: 'Hi, the sofa set is still available if you want to view it tonight.',
      sent_at: '2026-04-16T07:20:00.000Z',
    },
    {
      id: 'msg-101-2',
      chat_id: 'chat-101',
      sender_role: 'self',
      body: 'Great, could I arrange pickup on Saturday morning instead?',
      sent_at: '2026-04-16T07:32:00.000Z',
    },
    {
      id: 'msg-101-3',
      chat_id: 'chat-101',
      sender_role: 'peer',
      body: 'Saturday 11am pickup works for me.',
      sent_at: '2026-04-16T08:10:00.000Z',
    },
  ],
  'chat-102': [
    {
      id: 'msg-102-1',
      chat_id: 'chat-102',
      sender_role: 'self',
      body: 'Is the dehumidifier still under warranty?',
      sent_at: '2026-04-15T18:10:00.000Z',
    },
    {
      id: 'msg-102-2',
      chat_id: 'chat-102',
      sender_role: 'peer',
      body: 'The one-year warranty already ended. I have the original receipt though.',
      sent_at: '2026-04-15T18:22:00.000Z',
    },
    {
      id: 'msg-102-3',
      chat_id: 'chat-102',
      sender_role: 'peer',
      body: 'Marked as sold yesterday, thank you.',
      sent_at: '2026-04-15T19:40:00.000Z',
    },
  ],
};
