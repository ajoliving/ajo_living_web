/*
 * 二手交易模組展示工具。
 * 1. 統一處理後端帖子資料與舊原型資料的欄位差異。
 * 2. 供列表卡片、詳情頁與我的發布頁共用。
 */
import type { Listing } from '@/model/listing';
import type { MarketplaceListingLike } from '@/model/marketplace';
import type { AppLocale } from '@/stores/preferences';
import {
  getMarketplaceCategoryLabel,
  getMarketplaceConditionLabel,
  getMarketplaceDistrictLabel,
} from '@/constants/marketplace';

// 1. 判斷是否為後端帖子摘要格式
const isApiListing = (
  listing: MarketplaceListingLike,
): listing is Exclude<MarketplaceListingLike, Listing> =>
  'listing_id' in listing;

// 2. 將 snake_case 或 kebab-case 代碼轉為可讀標籤
export const humanizeCodeLabel = (value: string) =>
  value
    .trim()
    .split(/[_-]+/)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ');

// 3. 取得帖子主鍵
export const resolveListingId = (listing: MarketplaceListingLike) =>
  isApiListing(listing) ? listing.listing_id : String(listing.id);

// 4. 取得帖子主狀態
export const resolveListingStatus = (listing: MarketplaceListingLike) => {
  if (!isApiListing(listing)) {
    return listing.status;
  }

  if (listing.business_status === 'sold') {
    return 'sold';
  }

  if (listing.publication_status === 'active' && listing.business_status === 'available') {
    return 'available';
  }

  return listing.publication_status || 'active';
};

// 5. 取得帖子可見性
export const resolveListingVisibility = (listing: MarketplaceListingLike) =>
  isApiListing(listing) ? listing.visibility_scope : listing.visibility;

// 6. 取得帖子分類標籤
export const resolveListingCategoryLabel = (
  listing: MarketplaceListingLike,
  locale: AppLocale = 'zh-HK',
) => (isApiListing(listing) ? getMarketplaceCategoryLabel(listing.category_code, locale) : listing.category);

// 7. 取得帖子成色標籤
export const resolveListingConditionLabel = (
  listing: MarketplaceListingLike,
  locale: AppLocale = 'zh-HK',
) => (isApiListing(listing) ? getMarketplaceConditionLabel(listing.condition_level, locale) : listing.condition_label);

// 8. 取得帖子價格
export const resolveListingPrice = (listing: MarketplaceListingLike) => {
  if (!isApiListing(listing)) {
    return listing.price_hkd;
  }

  return listing.price_hkd ?? 0;
};

// 9. 取得帖子社區名稱
export const resolveListingCommunityName = (listing: MarketplaceListingLike) => {
  if (!isApiListing(listing)) {
    return listing.community.name;
  }

  if (!listing.community) {
    return getMarketplaceDistrictLabel(listing.district_code, 'zh-HK');
  }

  return (
    listing.community.name_zh.trim() ||
    listing.community.name_en.trim() ||
    listing.community.address_text.trim() ||
    getMarketplaceDistrictLabel(listing.community.district_code, 'zh-HK')
  );
};

// 10. 取得帖子社區次標題
export const resolveListingCommunitySecondary = (listing: MarketplaceListingLike) => {
  if (!isApiListing(listing)) {
    return listing.community.district;
  }

  return getMarketplaceDistrictLabel(listing.community?.district_code || listing.district_code, 'zh-HK');
};

// 11. 取得帖子發布者名稱
export const resolveListingOwnerName = (listing: MarketplaceListingLike) => {
  if (!isApiListing(listing)) {
    return listing.seller.display_name;
  }

  return (
    listing.owner?.display_name?.trim() ||
    humanizeCodeLabel(
      listing.owner?.publisher_identity_type || listing.publisher_identity_type || 'member',
    )
  );
};

// 12. 取得帖子封面圖
export const resolveListingCoverImage = (listing: MarketplaceListingLike) => {
  if (!isApiListing(listing)) {
    return listing.images.find((image) => image.is_cover) ?? listing.images[0];
  }

  const cover = listing.cover_image ?? ('images' in listing ? listing.images[0] : undefined);
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

// 13. 取得帖子全部圖片
export const resolveListingImages = (listing: MarketplaceListingLike) => {
  if (!isApiListing(listing)) {
    return listing.images;
  }

  if (!('images' in listing)) {
    const cover = resolveListingCoverImage(listing);
    return cover ? [cover] : [];
  }

  return listing.images.map((image) => ({
    id: image.media_asset_id,
    url: image.url,
    alt: listing.title,
    is_cover: image.is_cover,
  }));
};

// 14. 取得帖子描述
export const resolveListingDescription = (listing: MarketplaceListingLike) => {
  if (!isApiListing(listing)) {
    return listing.description;
  }

  return 'description' in listing ? listing.description : '';
};

// 15. 取得帖子摘要
export const resolveListingSummary = (listing: MarketplaceListingLike) => listing.summary;

// 16. 取得帖子標題
export const resolveListingTitle = (listing: MarketplaceListingLike) => listing.title;

// 17. 取得帖子發布時間
export const resolveListingPublishedAt = (listing: MarketplaceListingLike) => {
  if (!isApiListing(listing)) {
    return listing.published_at;
  }

  return listing.published_at || listing.updated_at;
};

// 18. 取得帖子到期時間
export const resolveListingExpireAt = (listing: MarketplaceListingLike) => {
  if (!isApiListing(listing)) {
    return listing.expires_at;
  }

  return listing.expire_at || listing.updated_at;
};

// 19. 取得帖子標籤
export const resolveListingTags = (listing: MarketplaceListingLike) => {
  if (!isApiListing(listing)) {
    return listing.tags;
  }

  if ('delivery_tags' in listing && listing.delivery_tags.length > 0) {
    return listing.delivery_tags;
  }

  return [humanizeCodeLabel(listing.category_code), humanizeCodeLabel(listing.visibility_scope)];
};
