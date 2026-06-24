/*
 * 物業頻道展示工具。
 * 1. 統一樓盤放售與服務式住宅列表、詳情欄位解析。
 * 2. 提供價格、圖片、屋苑與狀態等頁面共用邏輯。
 */
import type { PropertyChannel, PropertyListingSummaryResponse } from '@/model/property';
import type { AppLocale } from '@/stores/preferences';
import { getPropertyDistrictLabel, getPropertyTypeLabel } from '@/constants/property';
import { humanizeCodeLabel } from '@/utils/marketplace';

interface PropertyListingWithImages extends PropertyListingSummaryResponse {
  images?: Array<{
    media_asset_id: string;
    url: string;
    is_cover: boolean;
  }>;
}

// 1. 取得物業頻道
export const resolvePropertyChannel = (
  listing: PropertyListingSummaryResponse,
): PropertyChannel => (listing.module === 'serviced_apartment' ? 'serviced' : 'sale');

// 2. 取得物業詳情路徑
export const resolvePropertyDetailPath = (
  listing: PropertyListingSummaryResponse,
) => `${resolvePropertyChannel(listing) === 'serviced' ? '/serviced-residences' : '/properties'}/${listing.listing_id}`;

// 3. 取得物業標題
export const resolvePropertyTitle = (listing: PropertyListingSummaryResponse) => listing.title;

// 4. 取得物業摘要
export const resolvePropertySummary = (listing: PropertyListingSummaryResponse) => listing.summary;

// 5. 取得物業價格
export const resolvePropertyPrice = (listing: PropertyListingSummaryResponse) =>
  listing.serviced_apartment?.lowest_monthly_rent_hkd ||
  listing.property_sale?.asking_price_hkd ||
  0;

// 6. 取得物業面積
export const resolvePropertyArea = (listing: PropertyListingSummaryResponse) =>
  listing.serviced_apartment?.room_types?.[0]?.usable_area_sqft ||
  listing.property_sale?.usable_area_sqft ||
  0;

// 7. 取得物業房間摘要
export const resolvePropertyRooms = (listing: PropertyListingSummaryResponse) => {
  const sale = listing.property_sale;
  if (sale) {
    return `${sale.bedroom_count}房 ${sale.living_room_count}廳 ${sale.bathroom_count}廁`;
  }

  const roomCount = listing.serviced_apartment?.room_types.length ?? 0;
  return roomCount > 0 ? `${roomCount} 種房型` : '-';
};

// 8. 取得地區標籤
export const resolvePropertyDistrict = (
  listing: PropertyListingSummaryResponse,
  locale: AppLocale,
) => getPropertyDistrictLabel(listing.district_code, locale);

// 9. 取得物業類型或項目類型
export const resolvePropertyTypeLabel = (
  listing: PropertyListingSummaryResponse,
  locale: AppLocale,
) => listing.property_sale
  ? getPropertyTypeLabel(listing.property_sale.property_type, locale)
  : locale === 'zh-HK' ? '服務式住宅' : 'Serviced residence';

// 10. 取得物業封面
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

// 11. 取得物業全部圖片
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

// 12. 取得屋苑或項目名稱
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

// 13. 取得狀態標籤
export const resolvePropertyStatus = (listing: PropertyListingSummaryResponse) => {
  if (listing.business_status === 'sold') {
    return 'sold';
  }
  if (listing.publication_status === 'active' && listing.business_status === 'available') {
    return 'available';
  }

  return listing.publication_status || 'draft';
};

// 14. 取得聯絡角色
export const resolvePropertyPublisherRole = (listing: PropertyListingSummaryResponse) =>
  listing.property_sale?.publisher_role_label ||
  listing.serviced_apartment?.publisher_role_label ||
  resolvePublisherRoleLabel(listing.publisher_identity_type);

// 15. 取得發布身份標籤
const resolvePublisherRoleLabel = (identityType: string) => {
  const labels: Record<string, string> = {
    agent: '代理人',
    owner: '業主',
    professional_seller: '代理人',
    property_manager: '物業管理公司',
  };

  return labels[identityType] ?? humanizeCodeLabel(identityType);
};
