/*
 * 社區 Mock 資料。
 * 1. 提供一期示範用屋苑與大廈清單。
 * 2. 供會員資料、列表與發布流程共用。
 */
import type { Community } from '@/domains/building/model';

// 1. 建立固定社區清單
export const communities: Community[] = [
  {
    id: 101,
    slug: 'harbour-gate',
    name: 'Harbour Gate',
    district: 'Hung Hom',
    region: 'kowloon',
    building_count: 6,
  },
  {
    id: 102,
    slug: 'midtown-residence',
    name: 'Midtown Residence',
    district: 'Tsim Sha Tsui',
    region: 'kowloon',
    building_count: 4,
  },
  {
    id: 103,
    slug: 'seaside-vista',
    name: 'Seaside Vista',
    district: 'Sai Wan Ho',
    region: 'hong-kong-island',
    building_count: 8,
  },
  {
    id: 104,
    slug: 'pine-crest',
    name: 'Pine Crest',
    district: 'Sha Tin',
    region: 'new-territories',
    building_count: 5,
  },
  {
    id: 105,
    slug: 'island-bay',
    name: 'Island Bay',
    district: 'Tung Chung',
    region: 'outlying-islands',
    building_count: 3,
  },
];
