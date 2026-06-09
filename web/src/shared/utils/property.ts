/*
 * 物業頻道展示工具。
 * 1. 統一樓盤放售與服務式住宅列表、詳情欄位解析。
 * 2. 提供價格、圖片、屋苑與狀態等頁面共用邏輯。
 */
import type { PropertyChannel, PropertyListingSummaryResponse } from '@/domains/property/model';
import type { AppLocale } from '@/app/stores/preferences';
import { getPropertyDistrictLabel, getPropertyTagLabel, getPropertyTypeLabel } from '@/domains/property/constants';
import { formatPrice } from '@/shared/utils/format';
import { humanizeCodeLabel } from '@/shared/utils/marketplace';

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
export const mergePropertyFeatureTags = (tags: string[], sale?: PropertyAutoTagSource | null): string[] => {
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
export const resolvePropertyTitle = (listing: PropertyListingSummaryResponse) => listing.title;

// 6. 取得物業摘要
export const resolvePropertySummary = (listing: PropertyListingSummaryResponse) => listing.summary;

// 7. 取得物業價格
export const resolvePropertyPrice = (listing: PropertyListingSummaryResponse, priceMode: 'monthly' | 'daily' = 'monthly') =>
  (priceMode === 'daily'
    ? listing.serviced_apartment?.lowest_daily_rent_hkd
    : listing.serviced_apartment?.lowest_monthly_rent_hkd) ||
  (listing.property_sale?.transaction_type === 'rent' ? listing.property_sale.monthly_rent_hkd : 0) ||
  listing.property_sale?.asking_price_hkd ||
  0;

// 8. 取得物業價格展示標記
export const resolvePropertyPriceFlags = (listing: PropertyListingSummaryResponse) => ({
  referenceOnly: listing.serviced_apartment?.price_reference_only ?? listing.property_sale?.price_reference_only ?? false,
  negotiable: listing.serviced_apartment?.price_negotiable ?? listing.property_sale?.price_negotiable ?? false,
});

// 9. 取得物業價格文案
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

  const priceText = formatPrice(price, locale);
  return flags.referenceOnly ? `${priceText}${locale === 'en' ? ' from' : ' 起'}` : priceText;
};

// 10. 取得物業面積
export const resolvePropertyArea = (listing: PropertyListingSummaryResponse) =>
  listing.serviced_apartment?.min_usable_area_sqft ||
  listing.serviced_apartment?.room_types?.[0]?.usable_area_sqft ||
  listing.property_sale?.usable_area_sqft ||
  0;

// 11. 取得物業房間摘要
export const resolvePropertyRooms = (listing: PropertyListingSummaryResponse) => {
  const sale = listing.property_sale;
  if (sale) {
    const location = sale.public_location_text ? `${sale.public_location_text} · ` : '';
    return `${location}${sale.bedroom_count}房 ${sale.living_room_count}廳 ${sale.bathroom_count}廁`;
  }

  const roomCount = listing.serviced_apartment?.room_types.length ?? 0;
  return roomCount > 0 ? `${roomCount} 種房型` : '-';
};

// 12. 取得地區標籤
export const resolvePropertyDistrict = (
  listing: PropertyListingSummaryResponse,
  locale: AppLocale,
) => getPropertyDistrictLabel(listing.district_code, locale);

// 13. 取得物業類型或項目類型
export const resolvePropertyTypeLabel = (
  listing: PropertyListingSummaryResponse,
  locale: AppLocale,
) => listing.property_sale
  ? getPropertyTypeLabel(listing.property_sale.property_type, locale)
  : 'Serviced residence';

// 14. 取得樓盤交易類型
export const resolvePropertyTransactionType = (listing: PropertyListingSummaryResponse) =>
  listing.property_sale?.transaction_type === 'rent' ? 'rent' : 'sale';

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

// 14. 取得物業全部圖片
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

// 15. 取得屋苑或項目名稱
export const resolvePropertyCommunityName = (listing: PropertyListingSummaryResponse) => {
  if (listing.serviced_apartment?.project_name) {
    return listing.serviced_apartment.project_name;
  }
  if (listing.property_sale?.estate_name) {
    return listing.property_sale.estate_name;
  }
  if (listing.community) {
    return (
      listing.community.name_zh.trim() ||
      listing.community.name_en.trim() ||
      listing.community.address_text.trim()
    );
  }

  return humanizeCodeLabel(listing.district_code);
};

// 16. 取得狀態標籤
export const resolvePropertyStatus = (listing: PropertyListingSummaryResponse) => {
  if (listing.business_status === 'sold') {
    return 'sold';
  }
  if (listing.publication_status === 'active' && listing.business_status === 'available') {
    return 'available';
  }

  return listing.publication_status || 'draft';
};

// 17. 取得聯絡角色
export const resolvePropertyPublisherRole = (listing: PropertyListingSummaryResponse) =>
  listing.property_sale?.publisher_role_label ||
  listing.serviced_apartment?.publisher_role_label ||
  humanizeCodeLabel(listing.publisher_identity_type);

// 18. 取得樓盤特色標籤文案
export const resolvePropertyTagLabels = (
  listing: PropertyListingSummaryResponse,
  locale: AppLocale,
  limit = 5,
) => mergePropertyFeatureTags(listing.property_sale?.feature_tags ?? [], listing.property_sale)
  .concat(listing.serviced_apartment?.facility_tags ?? [], listing.serviced_apartment?.service_tags ?? [])
  .slice(0, limit)
  .map((tag) => getPropertyTagLabel(tag, locale));
