/*
 * 代理資料頁 - 修訂狀態規則。
 * 1. 已通過資料首次修改建立新修訂。
 * 2. 草稿或被拒修訂繼續更新，審核中修訂保持只讀。
 */
import type {
  AgencyAvatarType,
  AgencyProfile,
  AgencyProfileStatus,
  AgencyProfileType,
  AgencyProfileUpsertPayload,
} from '@/model/agency-profile';

interface AgencyProfileAssetDraft {
  avatar_asset_id: string;
  wechat_qr_asset_id: string;
  logo_asset_id: string;
  eaa_license_asset_id: string;
  business_registration_asset_id: string;
  company_card_asset_id: string;
}

// 1. 判斷儲存時建立或更新修訂
export const resolveAgencyProfileSaveMethod = (revision: AgencyProfile | null): 'create' | 'update' =>
  revision ? 'update' : 'create';

// 2. 判斷目前修訂是否只讀
export const isAgencyProfileReadonly = (status: AgencyProfileStatus): boolean => status === 'pending';

// 3. 判斷個人代理是否必須提供香港牌照及 EAA 證明
export const isIndividualAgencyLicenseRequired = (isOverseas: boolean): boolean => !isOverseas;

// 4. 判斷草稿是否可提交審核
export const canSubmitAgencyProfile = (
  hasRevision: boolean,
  status: AgencyProfileStatus,
  isDirty: boolean,
  canEdit: boolean,
): boolean => {
  if (status === 'pending') return false;
  if (isDirty) return canEdit;
  return hasRevision;
};

// 5. 建立與代理類型一致的媒體請求欄位
export const buildAgencyProfileAssetPayload = (
  profileType: AgencyProfileType,
  defaultAvatar: AgencyAvatarType,
  fields: AgencyProfileAssetDraft,
): Pick<AgencyProfileUpsertPayload,
  | 'default_avatar'
  | 'avatar_asset_id'
  | 'wechat_qr_asset_id'
  | 'logo_asset_id'
  | 'eaa_license_asset_id'
  | 'business_registration_asset_id'
  | 'company_card_asset_id'
> => ({
  default_avatar: profileType === 'individual' ? defaultAvatar : '',
  avatar_asset_id: profileType === 'individual' ? fields.avatar_asset_id || null : null,
  wechat_qr_asset_id: profileType === 'individual' ? fields.wechat_qr_asset_id || null : null,
  logo_asset_id: profileType === 'company' ? fields.logo_asset_id || null : null,
  eaa_license_asset_id: fields.eaa_license_asset_id || null,
  business_registration_asset_id: profileType === 'company'
    ? fields.business_registration_asset_id || null
    : null,
  company_card_asset_id: fields.company_card_asset_id || null,
});
