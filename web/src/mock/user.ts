/*
 * 會員 Mock 資料。
 * 1. 提供目前登入示範會員與賣家摘要。
 * 2. 用於登入、詳情、聊天與我的發布頁。
 */
import type { UserProfile, UserSummary } from '@/domains/account/model';

import { communities } from '@/mock/communities';
import { buildMockImage } from '@/mock/image';

const currentCommunity = communities[0];

// 1. 定義公開會員摘要
export const sellerProfiles: UserSummary[] = [
  {
    id: 201,
    display_name: 'Carmen Lee',
    avatar_url: buildMockImage('Carmen', '#3B82F6', '#60A5FA'),
    member_since: '2023-04-10T10:00:00.000Z',
    community_id: 101,
  },
  {
    id: 202,
    display_name: 'Ryan Ho',
    avatar_url: buildMockImage('Ryan', '#B4532C', '#F59E0B'),
    member_since: '2022-11-08T09:20:00.000Z',
    community_id: 102,
  },
  {
    id: 203,
    display_name: 'Mina Wong',
    avatar_url: buildMockImage('Mina', '#374151', '#D0A560'),
    member_since: '2024-01-12T13:40:00.000Z',
    community_id: 103,
  },
  {
    id: 204,
    display_name: 'Kevin Chan',
    avatar_url: buildMockImage('Kevin', '#14532D', '#86EFAC'),
    member_since: '2022-06-18T10:10:00.000Z',
    community_id: 104,
  },
  {
    id: 205,
    display_name: 'Ivy Lau',
    avatar_url: buildMockImage('Ivy', '#7C2D12', '#FDBA74'),
    member_since: '2023-09-01T15:15:00.000Z',
    community_id: 105,
  },
];

// 2. 定義當前示範會員
export const currentUserProfile: UserProfile = {
  id: 1,
  display_name: 'AJO Member',
  avatar_url: buildMockImage('Member', '#1C4ED8', '#60A5FA'),
  member_since: '2021-08-20T08:30:00.000Z',
  community_id: currentCommunity.id,
  phone: '+852 6123 4567',
  email: 'member@ajoliving.hk',
  preferred_locale: 'zh-HK',
  preferred_theme: 'default',
  primary_community: currentCommunity,
};
