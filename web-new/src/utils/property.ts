/*
 * 物業頻道展示工具。
 * 1. 統一樓盤放售與服務式住宅列表、詳情欄位解析。
 * 2. 提供價格、圖片、屋苑與狀態等頁面共用邏輯。
 */
import type { PropertyChannel, PropertyListingSummaryResponse } from '@/model/property';
import type { AppLocale } from '@/stores/preferences';
import { getPropertyDistrictLabel, getPropertyTagLabel, getPropertyTypeLabel } from '@/constants/property';
import { formatPrice } from '@/utils/format';
import { humanizeCodeLabel } from '@/utils/marketplace';

interface PropertyListingWithImages extends PropertyListingSummaryResponse {
  images?: Array<{
    media_asset_id: string;
    url: string;
    is_cover: boolean;
  }>;
}

interface PropertyAutoTagSource {
  video_url?: string;
  vr_url?: string;
  annual_prepay_discount?: boolean;
  annual_prepay_option?: string;
}

// 1. 取得樓盤自動標籤
export const resolvePropertyAutoTags = (sale?: PropertyAutoTagSource | null): string[] => {
  if (!sale) {
    return [];
  }

  return [
    sale.video_url?.trim() || sale.vr_url?.trim() ? 'vr_video' : '',
    sale.annual_prepay_option && sale.annual_prepay_option !== 'none' || sale.annual_prepay_discount ? 'prepay_discount' : '',
  ].filter(Boolean);
};

// 2. 合併人工與自動標籤
export const mergePropertyFeatureTags = (
  tags: string[],
  sale?: PropertyAutoTagSource | null,
): string[] => {
  const manualTags = tags.filter((tag) => tag !== 'vr_video' && tag !== 'prepay_discount');
  return [...new Set([...manualTags, ...resolvePropertyAutoTags(sale)])];
};

// 3. 取得物業頻道
export const resolvePropertyChannel = (
  listing: PropertyListingSummaryResponse,
): PropertyChannel => (listing.module === 'serviced_apartment' ? 'serviced' : 'sale');

// 4. 取得物業詳情路徑
export const resolvePropertyDetailPath = (
  listing: PropertyListingSummaryResponse,
) => `${resolvePropertyChannel(listing) === 'serviced' ? '/serviced-residences' : '/properties'}/${listing.listing_id}`;

// 5. 取得物業標題
export const resolvePropertyTitle = (
  listing: PropertyListingSummaryResponse,
  locale: AppLocale = 'zh-HK',
) => {
  if (locale === 'en') {
    return listing.property_sale?.title_en ||
      listing.serviced_apartment?.project_name_en ||
      listing.title;
  }

  return listing.title;
};

// 6. 取得物業摘要
export const resolvePropertySummary = (
  listing: PropertyListingSummaryResponse,
  locale: AppLocale = 'zh-HK',
) => {
  if (locale === 'en') {
    return listing.property_sale?.description_en ||
      listing.serviced_apartment?.description_en ||
      listing.summary;
  }

  return listing.summary;
};

// 7. 取得樓盤交易類型
export const resolvePropertyTransactionType = (listing: PropertyListingSummaryResponse) =>
  listing.property_sale?.transaction_type === 'rent' ? 'rent' : 'sale';

// 8. 取得物業價格
export const resolvePropertyPrice = (
  listing: PropertyListingSummaryResponse,
  priceMode: 'monthly' | 'daily' = 'monthly',
) =>
  (priceMode === 'daily'
    ? listing.serviced_apartment?.lowest_daily_rent_hkd
    : listing.serviced_apartment?.lowest_monthly_rent_hkd) ||
  (listing.property_sale?.transaction_type === 'rent' ? listing.property_sale.monthly_rent_hkd : 0) ||
  listing.property_sale?.asking_price_hkd ||
  0;

// 9. 取得物業價格展示標記
export const resolvePropertyPriceFlags = (listing: PropertyListingSummaryResponse) => ({
  referenceOnly: listing.serviced_apartment?.price_reference_only ?? listing.property_sale?.price_reference_only ?? false,
  negotiable: listing.serviced_apartment?.price_negotiable ?? listing.property_sale?.price_negotiable ?? false,
});

// 10. 取得物業價格文案
export const resolvePropertyPriceText = (
  listing: PropertyListingSummaryResponse,
  locale: AppLocale,
  priceMode: 'monthly' | 'daily' = 'monthly',
): string => {
  const flags = resolvePropertyPriceFlags(listing);
  if (flags.negotiable) {
    return locale === 'en' ? 'Negotiable' : '面議';
  }

  const price = resolvePropertyPrice(listing, priceMode);
  if (flags.referenceOnly && price <= 1) {
    return locale === 'en' ? 'Reference price' : '價格僅供參考';
  }
  if (listing.serviced_apartment && priceMode === 'monthly') {
    const highestPrice = Number(listing.serviced_apartment.highest_monthly_rent_hkd || 0);
    if (highestPrice > price && price > 0) {
      return `${formatPrice(price, locale)}-${formatPrice(highestPrice, locale)}`;
    }
  }

  const priceText = formatPrice(price, locale);
  return flags.referenceOnly ? `${priceText}${locale === 'en' ? ' from' : ' 起'}` : priceText;
};

// 11. 取得物業面積
export const resolvePropertyArea = (listing: PropertyListingSummaryResponse) =>
  listing.serviced_apartment?.min_usable_area_sqft ||
  listing.serviced_apartment?.room_types?.[0]?.usable_area_sqft ||
  listing.property_sale?.usable_area_sqft ||
  0;

// 12. 取得物業房間摘要
export const resolvePropertyRooms = (
  listing: PropertyListingSummaryResponse,
  locale: AppLocale = 'zh-HK',
) => {
  const sale = listing.property_sale;
  if (sale) {
    const locationText = locale === 'en'
      ? sale.address_text_en || sale.public_location_text
      : sale.public_location_text;
    const location = locationText ? `${locationText} · ` : '';
    const bedroomText = sale.bedroom_count < 0
      ? 'N/A'
      : locale === 'en'
        ? `${sale.bedroom_count} bed`
        : `${sale.bedroom_count}房`;
    const bathroomLabel = locale === 'en'
      ? ['industrial', 'shop'].includes(sale.property_type) ? 'toilet' : 'bath'
      : ['industrial', 'shop'].includes(sale.property_type) ? '廁' : '浴室';
    return `${location}${bedroomText} ${sale.bathroom_count} ${bathroomLabel}`;
  }

  const roomCount = listing.serviced_apartment?.room_types.length ?? 0;
  return roomCount > 0
    ? locale === 'en' ? `${roomCount} room types` : `${roomCount} 種房型`
    : '-';
};

// 13. 取得地區標籤
export const resolvePropertyDistrict = (
  listing: PropertyListingSummaryResponse,
  locale: AppLocale,
) => getPropertyDistrictLabel(listing.district_code, locale);

// 14. 取得物業類型或項目類型
export const resolvePropertyTypeLabel = (
  listing: PropertyListingSummaryResponse,
  locale: AppLocale,
) => listing.property_sale
  ? getPropertyTypeLabel(listing.property_sale.property_type, locale)
  : locale === 'zh-HK' ? '服務式住宅' : 'Serviced residence';

// 15. 取得物業封面
export const resolvePropertyCoverImage = (listing: PropertyListingSummaryResponse) => {
  const listingWithImages = listing as PropertyListingWithImages;
  const cover = listing.cover_image ?? listingWithImages.images?.[0];
  if (!cover) {
    return undefined;
  }

  return {
    id: cover.media_asset_id,
    url: cover.url,
    alt: listing.title,
    is_cover: cover.is_cover,
  };
};

// 16. 取得物業全部圖片
export const resolvePropertyImages = (listing: PropertyListingSummaryResponse) => {
  const listingWithImages = listing as PropertyListingWithImages;
  if (Array.isArray(listingWithImages.images)) {
    return listingWithImages.images.map((image) => ({
      id: image.media_asset_id,
      url: image.url,
      alt: listing.title,
      is_cover: image.is_cover,
    }));
  }

  const cover = resolvePropertyCoverImage(listing);
  return cover ? [cover] : [];
};

// 17. 取得屋苑或項目名稱
export const resolvePropertyCommunityName = (
  listing: PropertyListingSummaryResponse,
  locale: AppLocale = 'zh-HK',
) => {
  if (listing.serviced_apartment?.project_name) {
    return locale === 'en'
      ? listing.serviced_apartment.project_name_en || listing.serviced_apartment.project_name
      : listing.serviced_apartment.project_name;
  }
  if (locale === 'en' && listing.community) {
    return (
      listing.community.name_en.trim() ||
      listing.property_sale?.address_text_en?.trim() ||
      listing.property_sale?.estate_name?.trim() ||
      listing.community.name_zh.trim()
    );
  }
  if (listing.property_sale?.estate_name) {
    return listing.property_sale.estate_name;
  }
  if (listing.community) {
    return (
      (locale === 'en' ? listing.community.name_en : listing.community.name_zh).trim() ||
      (locale === 'en' ? listing.community.name_zh : listing.community.name_en).trim() ||
      listing.community.address_text.trim()
    );
  }

  return humanizeCodeLabel(listing.district_code);
};

// 18. 取得狀態標籤
export const resolvePropertyStatus = (listing: PropertyListingSummaryResponse) => {
  if (listing.business_status === 'sold') {
    return 'sold';
  }
  if (listing.publication_status === 'active' && listing.business_status === 'available') {
    return 'available';
  }

  return listing.publication_status || 'draft';
};

// 19. 取得聯絡角色
export const resolvePropertyPublisherRole = (
  listing: PropertyListingSummaryResponse,
  locale: AppLocale = 'zh-HK',
) => {
  if (locale === 'en') {
    return resolvePublisherRoleLabel(listing.publisher_identity_type, locale);
  }

  return listing.property_sale?.publisher_role_label ||
    listing.serviced_apartment?.publisher_role_label ||
    resolvePublisherRoleLabel(listing.publisher_identity_type, locale);
};

// 20. 取得樓盤特色標籤文案
export const resolvePropertyTagLabels = (
  listing: PropertyListingSummaryResponse,
  locale: AppLocale,
  limit = 5,
) => mergePropertyFeatureTags(listing.property_sale?.feature_tags ?? [], listing.property_sale)
  .concat(listing.serviced_apartment?.facility_tags ?? [], listing.serviced_apartment?.service_tags ?? [])
  .slice(0, limit)
  .map((tag) => getPropertyTagLabel(tag, locale));

// 21. 取得發布身份標籤
const resolvePublisherRoleLabel = (identityType: string, locale: AppLocale) => {
  const labels: Record<AppLocale, Record<string, string>> = {
    'zh-HK': {
      agent: '代理人',
      owner: '業主',
      professional_seller: '代理人',
      property_manager: '物業管理公司',
    },
    en: {
      agent: 'Agent',
      owner: 'Owner',
      professional_seller: 'Agent',
      property_manager: 'Property manager',
    },
  };

  return labels[locale][identityType] ?? humanizeCodeLabel(identityType);
};
