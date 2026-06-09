/*
 * 樓盤放售展示資料。
 * 1. 提供高保真 UI 展示用樓盤、地圖、配套、代理與圖則資料。
 * 2. 在真實資料不足時支撐發現、篩選與詳情頁骨架。
 */
import type { ContactAccessResult, PropertyListingDetailResponse, PropertyListingSummaryResponse } from '@/domains/property/model';

const buildDemoSalePayload = (
  value: Pick<NonNullable<PropertyListingSummaryResponse['property_sale']>,
    'transaction_type' |
    'property_type' |
    'rental_type' |
    'estate_name' |
    'address_text' |
    'asking_price_hkd' |
    'monthly_rent_hkd' |
    'price_reference_only' |
    'price_negotiable' |
    'area_mode' |
    'usable_area_sqft' |
    'gross_area_sqft' |
    'bedroom_count' |
    'living_room_count' |
    'bathroom_count' |
    'floor_level' |
    'direction' |
    'building_age' |
    'feature_tags' |
    'contact_method' |
    'publisher_role_label'
  >,
): NonNullable<PropertyListingSummaryResponse['property_sale']> => ({
  ...value,
  property_no: 'DEMO-PS',
  location_scope: 'local',
  listing_category: 'standard',
  multi_unit_project: false,
  address_text_en: value.address_text,
  block_name: '',
  unit_name: 'A',
  show_unit: true,
  price_reference_only: value.price_reference_only,
  price_negotiable: value.price_negotiable,
  annual_prepay_discount: false,
  annual_prepay_option: 'none',
  lease_start_date: '',
  rent_included: '',
  floor_raw: '',
  floor_zone: value.floor_level === '高層' ? 'high' : value.floor_level === '低層' ? 'low' : 'middle',
  floor_display_range: '',
  total_floors: 0,
  public_location_text: value.floor_level,
  kitchen_type: 'NA',
  cooking_mode: 'NA',
  management_fee_hkd: 0,
  video_url: '',
  vr_url: '',
  private_note: '',
  title_en: '',
  description_en: '',
  ad_package_code: 'basic',
  ad_weight: 0,
  ad_price_hkd: 600,
  ad_price_points: 600,
  ad_duration_days: 30,
  ad_expires_at: null,
});

export interface PropertyShowcaseMetric {
  label: string;
  value: string;
  caption: string;
}

export interface PropertyQuickFilter {
  label: string;
  value: string;
  to: string;
}

export interface PropertyNearbyPlace {
  category: 'transport' | 'mall' | 'dining' | 'school' | 'bank' | 'medical';
  categoryLabel: string;
  name: string;
  type: string;
  walkMinutes: number;
}

export interface PropertyAgentProfile {
  companyName: string;
  companyLicense: string;
  officeAddress: string;
  agentName: string;
  agentLicense: string;
  title: string;
  languages: string[];
  verifiedBadges: string[];
}

export interface PropertyFloorPlan {
  id: string;
  title: string;
  imageUrl: string;
}

export const propertyShowcaseMetrics: PropertyShowcaseMetric[] = [
  { label: '活躍放盤', value: '128', caption: '覆蓋港九新界主要屋苑' },
  { label: '平均查詢回覆', value: '18 分鐘', caption: '代理與業主聯絡入口分層' },
  { label: '資料完整度', value: '92%', caption: '相片、圖則與附近配套可視化' },
];

export const propertyQuickFilters: PropertyQuickFilter[] = [
  { label: '全新一手', value: 'brand_new', to: '/properties?feature=brand_new' },
  { label: '三房家庭', value: 'family_three_bed', to: '/properties?bedrooms=3' },
  { label: '景觀單位', value: 'view', to: '/properties?feature=view' },
  { label: '業主直售', value: 'owner_direct', to: '/properties?publisher=owner' },
  { label: '已裝修', value: 'renovated', to: '/properties?feature=renovated' },
];

export const demoPropertyListings: PropertyListingSummaryResponse[] = [
  {
    listing_id: 'demo-mei-king-01',
    module: 'property_sale',
    title: '美景花園 高層三房向南海景',
    summary: '高層開揚，三房兩廳，鄰近巴士總站及商場，適合家庭買家預約睇樓。',
    district_code: 'new_territories',
    publisher_identity_type: 'agent',
    publication_status: 'active',
    business_status: 'available',
    expire_at: '2026-06-14T10:00:00Z',
    published_at: '2026-05-12T10:00:00Z',
    updated_at: '2026-05-14T09:30:00Z',
    cover_image: {
      media_asset_id: 'demo-property-photo-01',
      url: '/home-stage/property-sale.webp',
      sort_order: 1,
      is_cover: true,
    },
    community: null,
    owner: {
      user_id: 'demo-agent-01',
      public_id: 'demo-agent-01',
      display_name: '陳小姐',
      avatar_url: '',
      publisher_identity_type: 'agent',
    },
    property_sale: buildDemoSalePayload({
      transaction_type: 'sale',
      property_type: 'residential',
      rental_type: '',
      estate_name: '美景花園',
      address_text: '青衣青康路 2 號',
    asking_price_hkd: 6380000,
    monthly_rent_hkd: 0,
    price_reference_only: false,
    price_negotiable: false,
      area_mode: 'usable',
      usable_area_sqft: 532,
      gross_area_sqft: 680,
      bedroom_count: 3,
      living_room_count: 2,
      bathroom_count: 1,
      floor_level: '高層',
      direction: '南',
      building_age: '約 38 年',
      feature_tags: ['view', 'renovated', 'parking_included'],
      contact_method: 'both',
      publisher_role_label: '持牌地產代理',
    }),
    serviced_apartment: null,
  },
  {
    listing_id: 'demo-taikoo-01',
    module: 'property_sale',
    title: '太古城 實用兩房 近港鐵站',
    summary: '屋苑管理成熟，步行可達港鐵及大型商場，室內保養良好。',
    district_code: 'hong_kong_island',
    publisher_identity_type: 'agent',
    publication_status: 'active',
    business_status: 'available',
    expire_at: '2026-06-10T10:00:00Z',
    published_at: '2026-05-10T10:00:00Z',
    updated_at: '2026-05-13T11:30:00Z',
    cover_image: {
      media_asset_id: 'demo-property-photo-02',
      url: '/home-stage/carousel/building.jpeg',
      sort_order: 1,
      is_cover: true,
    },
    community: null,
    owner: {
      user_id: 'demo-agent-02',
      public_id: 'demo-agent-02',
      display_name: '何先生',
      avatar_url: '',
      publisher_identity_type: 'agent',
    },
    property_sale: buildDemoSalePayload({
      transaction_type: 'sale',
      property_type: 'residential',
      rental_type: '',
      estate_name: '太古城',
      address_text: '太古城道 18 號',
    asking_price_hkd: 10380000,
    monthly_rent_hkd: 0,
    price_reference_only: false,
    price_negotiable: false,
      area_mode: 'usable',
      usable_area_sqft: 580,
      gross_area_sqft: 700,
      bedroom_count: 2,
      living_room_count: 2,
      bathroom_count: 1,
      floor_level: '中層',
      direction: '東南',
      building_age: '約 42 年',
      feature_tags: ['brand_new', 'renovated', 'appliances'],
      contact_method: 'both',
      publisher_role_label: '代理發布',
    }),
    serviced_apartment: null,
  },
  {
    listing_id: 'demo-harbour-01',
    module: 'property_sale',
    title: '西九龍臨海單位 會所配套齊全',
    summary: '臨海景觀，屋苑會所及交通配套完善，適合作自住或長線配置。',
    district_code: 'kowloon',
    publisher_identity_type: 'owner',
    publication_status: 'active',
    business_status: 'available',
    expire_at: '2026-06-08T10:00:00Z',
    published_at: '2026-05-08T10:00:00Z',
    updated_at: '2026-05-11T15:10:00Z',
    cover_image: {
      media_asset_id: 'demo-property-photo-03',
      url: '/images/pexels-jimmy-teoh-294331-35774007.jpg',
      sort_order: 1,
      is_cover: true,
    },
    community: null,
    owner: {
      user_id: 'demo-owner-01',
      public_id: 'demo-owner-01',
      display_name: '業主直售',
      avatar_url: '',
      publisher_identity_type: 'owner',
    },
    property_sale: buildDemoSalePayload({
      transaction_type: 'rent',
      property_type: 'residential',
      rental_type: 'short_term',
      estate_name: '海庭道屋苑',
      address_text: '西九龍海庭道 8 號',
    asking_price_hkd: 12800000,
    monthly_rent_hkd: 36000,
    price_reference_only: false,
    price_negotiable: false,
      area_mode: 'usable',
      usable_area_sqft: 712,
      gross_area_sqft: 890,
      bedroom_count: 3,
      living_room_count: 2,
      bathroom_count: 2,
      floor_level: '高層',
      direction: '西南',
      building_age: '約 12 年',
      feature_tags: ['view', 'furnished', 'prepay_discount'],
      contact_method: 'chat_or_whatsapp',
      publisher_role_label: '業主直售',
    }),
    serviced_apartment: null,
  },
];

const demoDetailDescriptions: Record<string, string> = {
  'demo-mei-king-01': '單位間隔實用，客飯廳分明，主人房及兩間睡房均有窗。廚房及浴室已完成基礎翻新，業主可配合預約睇樓。附近有巴士總站、商場、街市及社區配套，適合希望兼顧交通和生活便利的家庭。',
  'demo-taikoo-01': '屋苑管理成熟，交通與商場配套集中。單位保養良好，廳房四正，適合小家庭或專業人士。港鐵站、商場、餐飲及學校均在步行生活圈內。',
  'demo-harbour-01': '臨海景觀開揚，會所及住客配套齊全。屋苑鄰近西九龍商圈，交通網絡成熟，適合作自住或長線持有。業主直售，可安排彈性睇樓時間。',
};

export const demoPropertyDetails: PropertyListingDetailResponse[] = demoPropertyListings.map((listing, index) => ({
  ...listing,
  description: demoDetailDescriptions[listing.listing_id] ?? listing.summary,
  images: [
    listing.cover_image as NonNullable<PropertyListingSummaryResponse['cover_image']>,
    {
      media_asset_id: `demo-property-extra-${index + 1}`,
      url: index === 0 ? '/images/pexels-steppewalker-36596073.jpg' : '/home-stage/carousel/1613383ojE1fL80qbZphhF.png',
      sort_order: 2,
      is_cover: false,
    },
    {
      media_asset_id: `demo-property-extra-${index + 4}`,
      url: index === 2 ? '/home-stage/carousel/building.jpeg' : '/images/pexels-kseniya-kobi-3624194-7820979.jpg',
      sort_order: 3,
      is_cover: false,
    },
  ],
  contact_summary: {
    show_phone: true,
    show_whatsapp: true,
    show_chat: true,
    show_inquiry_form: false,
  },
}));

export const propertyNearbyPlaces: PropertyNearbyPlace[] = [
  { category: 'transport', categoryLabel: '交通', name: '美景花園', type: '巴士站', walkMinutes: 1 },
  { category: 'transport', categoryLabel: '交通', name: '美景花園，青康路', type: '巴士站', walkMinutes: 1 },
  { category: 'transport', categoryLabel: '交通', name: '美景花園巴士總站', type: '巴士站', walkMinutes: 2 },
  { category: 'mall', categoryLabel: '商場', name: '美景商場', type: '購物中心', walkMinutes: 3 },
  { category: 'dining', categoryLabel: '餐飲', name: '青衣生活餐飲帶', type: '餐飲', walkMinutes: 4 },
  { category: 'school', categoryLabel: '學校', name: '區內小學及幼稚園', type: '學校', walkMinutes: 7 },
  { category: 'bank', categoryLabel: '銀行', name: '主要銀行服務點', type: '銀行', walkMinutes: 6 },
  { category: 'medical', categoryLabel: '醫療', name: '社區診所', type: '醫療', walkMinutes: 8 },
];

export const propertyAgentProfile: PropertyAgentProfile = {
  companyName: 'AJO Premier Property Agency Limited',
  companyLicense: 'C-123456',
  officeAddress: '九龍尖沙咀廣東道 30 號 18 樓',
  agentName: '陳小姐',
  agentLicense: 'S-654321',
  title: '高級物業顧問',
  languages: ['廣東話', 'English', '普通話'],
  verifiedBadges: ['公司資料已核驗', '代理牌照已核驗', '電話已核驗'],
};

export const propertyFloorPlans: PropertyFloorPlan[] = [
  { id: 'plan-a', title: '三房兩廳參考圖則', imageUrl: '/home-stage/carousel/city.svg' },
  { id: 'plan-b', title: '同座標準層圖則', imageUrl: '/home-stage/carousel/city copy.svg' },
];

export const demoContactAccess: ContactAccessResult = {
  listing_id: 'demo',
  allowed_channels: {
    phone: true,
    whatsapp: true,
    chat: true,
  },
  contact_payload: {
    phone: '+852 9123 4567',
    whatsapp: '+852 9123 4567',
    agent: '陳小姐',
  },
  contact_access_granted: true,
  contact_already_paid: true,
};

// 1. 取得展示詳情
export const getDemoPropertyDetail = (listingId: string): PropertyListingDetailResponse | undefined =>
  demoPropertyDetails.find((listing) => listing.listing_id === listingId) ?? demoPropertyDetails[0];

// 2. 判斷是否為展示樓盤
export const isDemoPropertyListing = (listingId: string): boolean =>
  listingId.startsWith('demo-');
