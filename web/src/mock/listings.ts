/*
 * 二手帖子 Mock 資料。
 * 1. 提供列表、詳情、首頁與我的發布原型資料。
 * 2. 補足分類大類、小類、商品類型與地區篩選所需欄位。
 * 3. 覆蓋 public、building_only、active、expired、sold 等一期狀態。
 */
import type { Listing } from '@/model/listing';

import { communities } from '@/mock/communities';
import { buildMockImage } from '@/mock/image';
import { currentUserProfile, sellerProfiles } from '@/mock/user';

interface ListingCategoryLeaf {
  key: string;
  label_zh_hk: string;
  label_en: string;
}

interface ListingCategoryGroup {
  key: string;
  label_zh_hk: string;
  label_en: string;
  items: ListingCategoryLeaf[];
}

const [harbourGate, midtownResidence, seasideVista, pineCrest, islandBay] = communities;
const [carmenLee, ryanHo, minaWong, kevinChan, ivyLau] = sellerProfiles;

export const listingCategoryGroups: ListingCategoryGroup[] = [
  {
    key: 'home_furniture',
    label_zh_hk: '家居傢俬',
    label_en: 'Home Furniture',
    items: [
      { key: 'cabinet', label_zh_hk: '櫃', label_en: 'Cabinets' },
      { key: 'table-chair', label_zh_hk: '桌椅', label_en: 'Tables & Chairs' },
      { key: 'bed', label_zh_hk: '床', label_en: 'Beds' },
      { key: 'sofa', label_zh_hk: '梳化', label_en: 'Sofas' },
      { key: 'coffee-table', label_zh_hk: '茶几', label_en: 'Coffee Tables' },
    ],
  },
  {
    key: 'home_appliance',
    label_zh_hk: '家庭電器',
    label_en: 'Home Appliances',
    items: [
      { key: 'dehumidifier', label_zh_hk: '抽濕機', label_en: 'Dehumidifiers' },
      { key: 'washer', label_zh_hk: '洗衣機', label_en: 'Washing Machines' },
      { key: 'fridge', label_zh_hk: '雪櫃', label_en: 'Fridges' },
      { key: 'small-appliance', label_zh_hk: '小家電', label_en: 'Small Appliances' },
    ],
  },
  {
    key: 'electronics',
    label_zh_hk: '電子產品',
    label_en: 'Electronics',
    items: [
      { key: 'computer', label_zh_hk: '電腦', label_en: 'Computers' },
      { key: 'mobile-tablet', label_zh_hk: '手機平板', label_en: 'Mobile & Tablet' },
      { key: 'projector', label_zh_hk: '投影設備', label_en: 'Projectors' },
      { key: 'accessory', label_zh_hk: '電子配件', label_en: 'Accessories' },
    ],
  },
  {
    key: 'music',
    label_zh_hk: '樂器',
    label_en: 'Musical Instruments',
    items: [
      { key: 'keyboard', label_zh_hk: '鍵盤樂器', label_en: 'Keyboards' },
      { key: 'guitar', label_zh_hk: '結他', label_en: 'Guitars' },
      { key: 'other-instrument', label_zh_hk: '其他樂器', label_en: 'Other Instruments' },
    ],
  },
  {
    key: 'baby_goods',
    label_zh_hk: 'BB 用品',
    label_en: 'Baby Goods',
    items: [
      { key: 'stroller', label_zh_hk: '嬰兒車', label_en: 'Strollers' },
      { key: 'crib', label_zh_hk: '嬰兒床', label_en: 'Cribs' },
      { key: 'kids-furniture', label_zh_hk: '兒童傢俬', label_en: 'Kids Furniture' },
    ],
  },
  {
    key: 'office_furniture',
    label_zh_hk: '辦公室傢俬',
    label_en: 'Office Furniture',
    items: [
      { key: 'office-chair', label_zh_hk: '辦公椅', label_en: 'Office Chairs' },
      { key: 'office-desk', label_zh_hk: '辦公桌', label_en: 'Office Desks' },
      { key: 'storage-rack', label_zh_hk: '收納架', label_en: 'Storage Racks' },
    ],
  },
  {
    key: 'home_decor',
    label_zh_hk: '家居裝飾',
    label_en: 'Home Decor',
    items: [
      { key: 'lighting', label_zh_hk: '燈飾', label_en: 'Lighting' },
      { key: 'rug-textile', label_zh_hk: '地毯布藝', label_en: 'Rugs & Textiles' },
      { key: 'decorative-object', label_zh_hk: '裝飾擺設', label_en: 'Decor Objects' },
    ],
  },
  {
    key: 'other',
    label_zh_hk: '其他',
    label_en: 'Others',
    items: [{ key: 'miscellaneous', label_zh_hk: '其他', label_en: 'Miscellaneous' }],
  },
];

const listingImages = {
  loungeSet: [
    {
      id: 'img-101-a',
      url: buildMockImage('Nordic Lounge Set', '#1C4ED8', '#8FB7FF'),
      alt: 'Nordic Lounge Set',
      is_cover: true,
    },
    {
      id: 'img-101-b',
      url: buildMockImage('Oak Coffee Table', '#3B82F6', '#BBD0FF'),
      alt: 'Oak Coffee Table',
      is_cover: false,
    },
  ],
  babyStroller: [
    {
      id: 'img-102-a',
      url: buildMockImage('Buggy with Rain Cover', '#B4532C', '#F2C7AE'),
      alt: 'Buggy with Rain Cover',
      is_cover: true,
    },
  ],
  dehumidifier: [
    {
      id: 'img-103-a',
      url: buildMockImage('Smart Dehumidifier', '#334155', '#D0A560'),
      alt: 'Smart Dehumidifier',
      is_cover: true,
    },
  ],
  diningSet: [
    {
      id: 'img-104-a',
      url: buildMockImage('Walnut Dining Set', '#7C3AED', '#C4B5FD'),
      alt: 'Walnut Dining Set',
      is_cover: true,
    },
  ],
  lampPair: [
    {
      id: 'img-105-a',
      url: buildMockImage('Sculptural Table Lamps', '#0F766E', '#67E8F9'),
      alt: 'Sculptural Table Lamps',
      is_cover: true,
    },
  ],
  storageCabinet: [
    {
      id: 'img-106-a',
      url: buildMockImage('Storage Cabinet', '#475569', '#CBD5E1'),
      alt: 'Storage Cabinet',
      is_cover: true,
    },
  ],
  officeChair: [
    {
      id: 'img-107-a',
      url: buildMockImage('Ergonomic Office Chair', '#2563EB', '#93C5FD'),
      alt: 'Ergonomic Office Chair',
      is_cover: true,
    },
  ],
  digitalKeyboard: [
    {
      id: 'img-108-a',
      url: buildMockImage('Digital Keyboard', '#7C3AED', '#DDD6FE'),
      alt: 'Digital Keyboard',
      is_cover: true,
    },
  ],
  babyCrib: [
    {
      id: 'img-109-a',
      url: buildMockImage('Bedside Baby Crib', '#B4532C', '#FCD9C4'),
      alt: 'Bedside Baby Crib',
      is_cover: true,
    },
  ],
  projector: [
    {
      id: 'img-110-a',
      url: buildMockImage('Portable Projector', '#111827', '#9CA3AF'),
      alt: 'Portable Projector',
      is_cover: true,
    },
  ],
  washingMachine: [
    {
      id: 'img-111-a',
      url: buildMockImage('Washing Machine', '#0F172A', '#94A3B8'),
      alt: 'Washing Machine',
      is_cover: true,
    },
  ],
  sideboard: [
    {
      id: 'img-112-a',
      url: buildMockImage('Oak Sideboard Cabinet', '#92400E', '#FCD34D'),
      alt: 'Oak Sideboard Cabinet',
      is_cover: true,
    },
  ],
  workDesk: [
    {
      id: 'img-113-a',
      url: buildMockImage('Work Desk', '#166534', '#86EFAC'),
      alt: 'Work Desk',
      is_cover: true,
    },
  ],
  laptopDock: [
    {
      id: 'img-114-a',
      url: buildMockImage('Laptop Dock and Stand', '#1D4ED8', '#BFDBFE'),
      alt: 'Laptop Dock and Stand',
      is_cover: true,
    },
  ],
};

// 1. 定義市場帖子清單
export const marketplaceListings: Listing[] = [
  {
    id: 101,
    slug: 'nordic-lounge-set',
    title: 'Nordic lounge set with oak coffee table',
    category: 'Living room furniture',
    category_group_key: 'home_furniture',
    category_key: 'sofa',
    listing_type: 'used',
    condition_label: 'Gently used',
    price_hkd: 3200,
    summary: 'Three-seat sofa with oak coffee table, ideal for first home setup.',
    description:
      'A bright, clean lounge set kept in a smoke-free flat. Includes sofa, oak coffee table, and two neutral cushions. Available for self-pickup this weekend.',
    visibility: 'public',
    status: 'active',
    published_at: '2026-04-12T09:00:00.000Z',
    expires_at: '2026-05-12T09:00:00.000Z',
    tags: ['Self pickup', 'Near MTR', 'Well kept'],
    featured: true,
    community: harbourGate,
    seller: carmenLee,
    images: listingImages.loungeSet,
    contact: {
      phone: '+852 6777 2201',
      whatsapp: 'https://wa.me/85267772201',
      allow_chat: true,
      gate_hint: '公開帖子可先查看聯絡入口，再決定是否開啟站內聊天。',
    },
    stats: {
      view_count: 182,
      save_count: 26,
      chat_count: 9,
    },
  },
  {
    id: 102,
    slug: 'baby-stroller-rain-cover',
    title: 'Compact baby stroller with rain cover',
    category: 'Baby stroller',
    category_group_key: 'baby_goods',
    category_key: 'stroller',
    listing_type: 'personal',
    condition_label: 'Excellent',
    price_hkd: 980,
    summary: 'Lightweight stroller from same community owner, easy folding design.',
    description:
      'Used for six months and always stored indoors. Includes rain cover, snack tray, and side hooks. Perfect for quick lift access and neighbourhood errands.',
    visibility: 'building_only',
    status: 'active',
    published_at: '2026-04-14T07:30:00.000Z',
    expires_at: '2026-05-14T07:30:00.000Z',
    tags: ['Building only', 'Quick fold', 'Baby gear'],
    featured: true,
    community: harbourGate,
    seller: currentUserProfile,
    images: listingImages.babyStroller,
    contact: {
      phone: '+852 6123 4567',
      whatsapp: 'https://wa.me/85261234567',
      allow_chat: true,
      gate_hint: '同屋苑會員登入後才會完整顯示聯絡方式。',
    },
    stats: {
      view_count: 91,
      save_count: 14,
      chat_count: 4,
    },
  },
  {
    id: 103,
    slug: 'smart-dehumidifier-20l',
    title: 'Smart dehumidifier 20L with laundry mode',
    category: 'Dehumidifier',
    category_group_key: 'home_appliance',
    category_key: 'dehumidifier',
    listing_type: 'used',
    condition_label: 'Very good',
    price_hkd: 1650,
    summary: 'High-capacity dehumidifier with app scheduling and quiet sleep mode.',
    description:
      'Used for one wet season only. Includes original manual, removable water tank, and laundry drying mode. Seller prefers weekend pickup.',
    visibility: 'public',
    status: 'sold',
    published_at: '2026-04-02T12:10:00.000Z',
    expires_at: '2026-04-28T12:10:00.000Z',
    tags: ['Quiet mode', 'Laundry mode'],
    featured: false,
    community: midtownResidence,
    seller: ryanHo,
    images: listingImages.dehumidifier,
    contact: {
      phone: '+852 6888 1534',
      whatsapp: 'https://wa.me/85268881534',
      allow_chat: true,
      gate_hint: '帖子已售出，聊天入口仍可查看過往訊息。',
    },
    stats: {
      view_count: 264,
      save_count: 33,
      chat_count: 18,
    },
  },
  {
    id: 104,
    slug: 'walnut-dining-set',
    title: 'Walnut dining set for four seats',
    category: 'Dining table set',
    category_group_key: 'home_furniture',
    category_key: 'table-chair',
    listing_type: 'used',
    condition_label: 'Good',
    price_hkd: 2400,
    summary: 'Table and four chairs, warm walnut tone with compact footprint.',
    description:
      'A sturdy dining set from a pet-free household. Minor wear on one chair edge. Best for buyers looking for same-day self pickup in Sai Wan Ho.',
    visibility: 'building_only',
    status: 'expired',
    published_at: '2026-03-08T14:20:00.000Z',
    expires_at: '2026-04-08T14:20:00.000Z',
    tags: ['Pickup only', 'For family', 'Expired'],
    featured: false,
    community: seasideVista,
    seller: minaWong,
    images: listingImages.diningSet,
    contact: {
      phone: '+852 6555 9981',
      whatsapp: 'https://wa.me/85265559981',
      allow_chat: true,
      gate_hint: '帖子已過期，重新發布後才會回到上架列表。',
    },
    stats: {
      view_count: 77,
      save_count: 11,
      chat_count: 3,
    },
  },
  {
    id: 105,
    slug: 'sculptural-table-lamps',
    title: 'Pair of sculptural bedside table lamps',
    category: 'Lighting decor',
    category_group_key: 'home_decor',
    category_key: 'lighting',
    listing_type: 'new',
    condition_label: 'Like new',
    price_hkd: 620,
    summary: 'Warm light pair suitable for bedside, study corner, or entry cabinet.',
    description:
      'Bought for styling a show flat and no longer needed. Each lamp includes dimmable bulb and textured linen shade.',
    visibility: 'public',
    status: 'active',
    published_at: '2026-04-15T16:40:00.000Z',
    expires_at: '2026-05-15T16:40:00.000Z',
    tags: ['Decor', 'Warm tone', 'Ready to pickup'],
    featured: false,
    community: midtownResidence,
    seller: ryanHo,
    images: listingImages.lampPair,
    contact: {
      phone: '+852 6888 1534',
      whatsapp: 'https://wa.me/85268881534',
      allow_chat: true,
      gate_hint: '公開帖子可直接開啟站內聊天。',
    },
    stats: {
      view_count: 56,
      save_count: 9,
      chat_count: 2,
    },
  },
  {
    id: 106,
    slug: 'tall-storage-cabinet-free-pickup',
    title: 'Tall storage cabinet, free pickup this week',
    category: 'Storage cabinet',
    category_group_key: 'home_furniture',
    category_key: 'cabinet',
    listing_type: 'personal',
    condition_label: 'Used with minor marks',
    price_hkd: 0,
    summary: 'Free standing storage cabinet, good for utility room or balcony supplies.',
    description:
      'Owner is moving out and prefers to gift the cabinet to someone who can collect this week. Minor scratches on one side, but structure is still solid.',
    visibility: 'public',
    status: 'active',
    published_at: '2026-04-11T10:00:00.000Z',
    expires_at: '2026-05-11T10:00:00.000Z',
    tags: ['願意捐贈', 'Free pickup', 'Storage'],
    featured: false,
    community: pineCrest,
    seller: kevinChan,
    images: listingImages.storageCabinet,
    contact: {
      phone: '+852 6333 1288',
      whatsapp: 'https://wa.me/85263331288',
      allow_chat: true,
      gate_hint: '公開帖子可直接聯絡賣家安排交收。',
    },
    stats: {
      view_count: 68,
      save_count: 10,
      chat_count: 5,
    },
  },
  {
    id: 107,
    slug: 'ergonomic-office-chair',
    title: 'Ergonomic office chair with adjustable lumbar support',
    category: 'Office chair',
    category_group_key: 'office_furniture',
    category_key: 'office-chair',
    listing_type: 'used',
    condition_label: 'Good working condition',
    price_hkd: 480,
    summary: 'Mesh back office chair suitable for WFH setup and compact study rooms.',
    description:
      'Seat cushion remains supportive and wheel base moves smoothly. One armrest has a light surface mark but no crack.',
    visibility: 'public',
    status: 'active',
    published_at: '2026-04-13T19:10:00.000Z',
    expires_at: '2026-05-13T19:10:00.000Z',
    tags: ['WFH', 'Mesh back', 'Adjustable'],
    featured: true,
    community: pineCrest,
    seller: kevinChan,
    images: listingImages.officeChair,
    contact: {
      phone: '+852 6333 1288',
      whatsapp: 'https://wa.me/85263331288',
      allow_chat: true,
      gate_hint: '公開帖子可直接聯絡賣家安排交收。',
    },
    stats: {
      view_count: 134,
      save_count: 21,
      chat_count: 7,
    },
  },
  {
    id: 108,
    slug: 'digital-keyboard-61-keys',
    title: 'Digital keyboard 61 keys with sustain pedal',
    category: 'Digital keyboard',
    category_group_key: 'music',
    category_key: 'keyboard',
    listing_type: 'used',
    condition_label: 'Very good',
    price_hkd: 1680,
    summary: 'Entry-level keyboard with stand, pedal, and foldable stool included.',
    description:
      'Ideal for beginners or children starting lessons. Seller kept the original adapter and music stand.',
    visibility: 'public',
    status: 'active',
    published_at: '2026-04-10T11:20:00.000Z',
    expires_at: '2026-05-10T11:20:00.000Z',
    tags: ['Music lesson', 'Beginner friendly', 'With stand'],
    featured: false,
    community: islandBay,
    seller: ivyLau,
    images: listingImages.digitalKeyboard,
    contact: {
      phone: '+852 6222 7788',
      whatsapp: 'https://wa.me/85262227788',
      allow_chat: true,
      gate_hint: '公開帖子可直接聯絡賣家安排交收。',
    },
    stats: {
      view_count: 83,
      save_count: 15,
      chat_count: 4,
    },
  },
  {
    id: 109,
    slug: 'bedside-baby-crib',
    title: 'Bedside baby crib with mattress and mosquito net',
    category: 'Baby crib',
    category_group_key: 'baby_goods',
    category_key: 'crib',
    listing_type: 'new',
    condition_label: 'Almost new',
    price_hkd: 720,
    summary: 'Foldable bedside crib, easy to move between bedroom and living area.',
    description:
      'Used for a short newborn stage only. Includes mattress protector, side storage pouch, and mosquito net frame.',
    visibility: 'building_only',
    status: 'active',
    published_at: '2026-04-16T08:45:00.000Z',
    expires_at: '2026-05-16T08:45:00.000Z',
    tags: ['Baby room', 'Foldable', 'Same building'],
    featured: false,
    community: harbourGate,
    seller: currentUserProfile,
    images: listingImages.babyCrib,
    contact: {
      phone: '+852 6123 4567',
      whatsapp: 'https://wa.me/85261234567',
      allow_chat: true,
      gate_hint: '同屋苑會員登入後才會完整顯示聯絡方式。',
    },
    stats: {
      view_count: 47,
      save_count: 8,
      chat_count: 3,
    },
  },
  {
    id: 110,
    slug: 'portable-projector-home-cinema',
    title: 'Portable projector for home cinema and bedroom wall',
    category: 'Portable projector',
    category_group_key: 'electronics',
    category_key: 'projector',
    listing_type: 'personal',
    condition_label: 'Good',
    price_hkd: 1280,
    summary: 'Compact projector with HDMI dongle support and built-in speaker.',
    description:
      'Previously used for movie nights. Image remains bright in a dim room and the unit still includes carrying pouch and remote.',
    visibility: 'public',
    status: 'active',
    published_at: '2026-04-09T17:30:00.000Z',
    expires_at: '2026-05-09T17:30:00.000Z',
    tags: ['Movie night', 'Portable', 'Compact'],
    featured: true,
    community: islandBay,
    seller: ivyLau,
    images: listingImages.projector,
    contact: {
      phone: '+852 6222 7788',
      whatsapp: 'https://wa.me/85262227788',
      allow_chat: true,
      gate_hint: '公開帖子可直接聯絡賣家安排交收。',
    },
    stats: {
      view_count: 149,
      save_count: 18,
      chat_count: 6,
    },
  },
  {
    id: 111,
    slug: 'front-load-washing-machine-7kg',
    title: 'Front-load washing machine 7kg, suitable for small flats',
    category: 'Washing machine',
    category_group_key: 'home_appliance',
    category_key: 'washer',
    listing_type: 'used',
    condition_label: 'Working well',
    price_hkd: 900,
    summary: 'Compact depth machine, good for smaller utility balconies and family laundry.',
    description:
      'Machine is still in daily use and can be tested before pickup. Seller will share measurements for lift and corridor planning.',
    visibility: 'public',
    status: 'active',
    published_at: '2026-04-07T09:50:00.000Z',
    expires_at: '2026-05-07T09:50:00.000Z',
    tags: ['Appliance', 'Pickup planning', 'Family use'],
    featured: false,
    community: pineCrest,
    seller: kevinChan,
    images: listingImages.washingMachine,
    contact: {
      phone: '+852 6333 1288',
      whatsapp: 'https://wa.me/85263331288',
      allow_chat: true,
      gate_hint: '公開帖子可直接聯絡賣家安排交收。',
    },
    stats: {
      view_count: 98,
      save_count: 16,
      chat_count: 5,
    },
  },
  {
    id: 112,
    slug: 'oak-sideboard-cabinet',
    title: 'Oak sideboard cabinet for dining area or hallway',
    category: 'Sideboard cabinet',
    category_group_key: 'home_furniture',
    category_key: 'cabinet',
    listing_type: 'used',
    condition_label: 'Very good',
    price_hkd: 650,
    summary: 'Warm oak storage piece with three drawers and two hidden side compartments.',
    description:
      'Works well as dining sideboard, hallway shoe cabinet, or TV console. One owner only and stored indoors throughout.',
    visibility: 'public',
    status: 'active',
    published_at: '2026-04-06T13:15:00.000Z',
    expires_at: '2026-05-06T13:15:00.000Z',
    tags: ['Oak finish', 'Storage', 'Living room'],
    featured: false,
    community: seasideVista,
    seller: minaWong,
    images: listingImages.sideboard,
    contact: {
      phone: '+852 6555 9981',
      whatsapp: 'https://wa.me/85265559981',
      allow_chat: true,
      gate_hint: '公開帖子可直接聯絡賣家安排交收。',
    },
    stats: {
      view_count: 73,
      save_count: 12,
      chat_count: 4,
    },
  },
  {
    id: 113,
    slug: 'compact-work-desk',
    title: 'Compact work desk with cable slot and side shelf',
    category: 'Work desk',
    category_group_key: 'office_furniture',
    category_key: 'office-desk',
    listing_type: 'personal',
    condition_label: 'Lightly used',
    price_hkd: 300,
    summary: 'Slim desk footprint that fits study corners, bedrooms, or part-time office setup.',
    description:
      'Desk includes a side shelf for router or printer and a cut-out for cable routing. Easy to disassemble for moving.',
    visibility: 'public',
    status: 'active',
    published_at: '2026-04-05T18:25:00.000Z',
    expires_at: '2026-05-05T18:25:00.000Z',
    tags: ['Study room', 'Compact', 'Easy move'],
    featured: false,
    community: midtownResidence,
    seller: ryanHo,
    images: listingImages.workDesk,
    contact: {
      phone: '+852 6888 1534',
      whatsapp: 'https://wa.me/85268881534',
      allow_chat: true,
      gate_hint: '公開帖子可直接聯絡賣家安排交收。',
    },
    stats: {
      view_count: 58,
      save_count: 9,
      chat_count: 3,
    },
  },
  {
    id: 114,
    slug: 'laptop-dock-stand-set',
    title: 'Laptop dock and stand set for dual-screen desk setup',
    category: 'Laptop accessory',
    category_group_key: 'electronics',
    category_key: 'accessory',
    listing_type: 'new',
    condition_label: 'Brand new in box',
    price_hkd: 180,
    summary: 'Dock and aluminium stand combo, good for WFH desk or student laptop use.',
    description:
      'Unused set purchased for an office setup that was later cancelled. Includes USB-C hub, stand, and original packaging.',
    visibility: 'public',
    status: 'active',
    published_at: '2026-04-08T15:00:00.000Z',
    expires_at: '2026-05-08T15:00:00.000Z',
    tags: ['New in box', 'Desk setup', 'Accessory'],
    featured: false,
    community: harbourGate,
    seller: carmenLee,
    images: listingImages.laptopDock,
    contact: {
      phone: '+852 6777 2201',
      whatsapp: 'https://wa.me/85267772201',
      allow_chat: true,
      gate_hint: '公開帖子可直接聯絡賣家安排交收。',
    },
    stats: {
      view_count: 64,
      save_count: 13,
      chat_count: 2,
    },
  },
];

// 2. 定義首頁推薦清單
export const featuredListings = marketplaceListings.filter((listing) => listing.featured);

// 3. 定義我的發布清單
export const myListings = marketplaceListings.filter(
  (listing) => listing.seller.id === currentUserProfile.id,
);

// 4. 依據 ID 取得帖子詳情
export const getListingById = (listingId: number) =>
  marketplaceListings.find((listing) => listing.id === listingId);
