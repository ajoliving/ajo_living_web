<!--
 * 會員中心 - 地產代理資料。
 * 1. 按帳戶類型管理個人代理或代理公司資料及人工審核版本。
 * 2. 將頭像、Logo、牌照、商業登記證及公司卡分用途上傳至 OSS。
 * 3. 代理公司通過審核後管理公司子帳戶。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import {
  createAgencySubaccount,
  createMyAgencyProfile,
  fetchAgencySubaccounts,
  fetchMyAgencyProfile,
  removeAgencySubaccount,
  submitMyAgencyProfile,
  updateAgencySubaccountStatus,
  updateMyAgencyProfile,
} from '@/httpapis/agency-profiles';
import { completeUpload, createUploadPresign } from '@/httpapis/uploads';
import type {
  AgencyAsset,
  AgencyProfile,
  AgencyProfileStatus,
  AgencyProfileUpsertPayload,
  AgencySubaccount,
  AgencySubaccountPermission,
  AgencySubaccountStatus,
  CreateAgencySubaccountPayload,
} from '@/model/agency-profile';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';
import { useSessionStore } from '@/stores/session';
import { formatDate } from '@/utils/format';
import { buildUploadHeaders } from '@/utils/upload';
import {
  buildAgencyProfileAssetPayload,
  canSubmitAgencyProfile,
  isIndividualAgencyLicenseRequired,
  isAgencyProfileReadonly,
  resolveAgencyProfileSaveMethod,
} from './profile-state';

const MAX_FILE_SIZE = 10 * 1024 * 1024;
type AssetField = 'avatar' | 'wechat_qr' | 'eaa_license' | 'company_card' | 'logo' | 'business_registration';
type AgencyProfileField =
  | 'name_zh'
  | 'name_en'
  | 'address_zh'
  | 'address_en'
  | 'license_number'
  | 'phone_1_country_code'
  | 'phone_1_number'
  | 'avatar_asset_id'
  | 'eaa_license_asset_id'
  | 'business_registration_asset_id';
type AgencySubaccountField = 'display_name' | 'phone_country_code' | 'phone_number' | 'email' | 'password' | 'permissions';

const { t } = useI18n();
const router = useRouter();
const feedbackStore = useFeedbackStore();
const preferenceStore = usePreferenceStore();
const sessionStore = useSessionStore();
const activeProfile = ref<AgencyProfile | null>(null);
const revision = ref<AgencyProfile | null>(null);
const nextEditableAt = ref<string | null>(null);
const isLoading = ref(false);
const isSaving = ref(false);
const isSubmitting = ref(false);
const uploadingField = ref<AssetField | null>(null);
const assets = reactive<Partial<Record<AssetField, AgencyAsset>>>({});
const subaccounts = ref<AgencySubaccount[]>([]);
const showSubaccountForm = ref(false);
const subaccountSaving = ref(false);
const isSigningOut = ref(false);
const savedPayloadSnapshot = ref('');
const profileFieldErrors = reactive<Record<AgencyProfileField, string>>({
  name_zh: '',
  name_en: '',
  address_zh: '',
  address_en: '',
  license_number: '',
  phone_1_country_code: '',
  phone_1_number: '',
  avatar_asset_id: '',
  eaa_license_asset_id: '',
  business_registration_asset_id: '',
});
const subaccountFieldErrors = reactive<Record<AgencySubaccountField, string>>({
  display_name: '',
  phone_country_code: '',
  phone_number: '',
  email: '',
  password: '',
  permissions: '',
});

const profileType = computed<'individual' | 'company'>(() =>
  sessionStore.me?.account_type === 'agency_company' ? 'company' : 'individual',
);
const profileHeaderIcon = computed<'building' | 'user'>(() => profileType.value === 'company' ? 'building' : 'user');

const form = reactive({
  name_zh: '', name_en: '', phone_1_country_code: '+852', phone_1_number: '', phone_1_whatsapp: false,
  phone_2_country_code: '+852', phone_2_number: '', phone_2_whatsapp: false, wechat_id: '', wechat_url: '',
  signature_zh: '', signature_en: '', default_avatar: 'male' as 'male' | 'female' | 'custom', avatar_asset_id: '', wechat_qr_asset_id: '',
  is_overseas: false, license_number: '', eaa_license_asset_id: '', company_card_asset_id: '',
  address_zh: '', address_en: '', is_big_four: false,
  logo_asset_id: '', business_registration_asset_id: '',
});

const subaccountForm = reactive<CreateAgencySubaccountPayload>({
  display_name: '', phone_country_code: '+852', phone_number: '', email: '', password: '', permissions: ['property_publish'],
});

// 1. 統一讀取接口錯誤
const readErrorMessage = (error: unknown, fallback: string): string =>
  axios.isAxiosError<{ message?: string }>(error) ? error.response?.data?.message ?? fallback : fallback;

// 2. 將代理資料同步至表單
const syncForm = (profile: AgencyProfile | null): void => {
  const value = profile ?? ({} as AgencyProfile);
  Object.assign(form, {
    name_zh: value.name_zh ?? '', name_en: value.name_en ?? '',
    phone_1_country_code: value.phone_1_country_code || '+852', phone_1_number: value.phone_1_number ?? '',
    phone_1_whatsapp: value.phone_1_whatsapp ?? false, phone_2_country_code: value.phone_2_country_code || '+852',
    phone_2_number: value.phone_2_number ?? '', phone_2_whatsapp: value.phone_2_whatsapp ?? false,
    wechat_id: value.wechat_id ?? '', wechat_url: value.wechat_url ?? '', signature_zh: value.signature_zh ?? '',
    signature_en: value.signature_en ?? '', default_avatar: value.default_avatar || 'male', avatar_asset_id: value.avatar_asset?.media_asset_id ?? '',
    wechat_qr_asset_id: value.wechat_qr_asset?.media_asset_id ?? '',
    is_overseas: value.is_overseas ?? false, license_number: value.license_number ?? '',
    eaa_license_asset_id: value.eaa_license_asset?.media_asset_id ?? '', company_card_asset_id: value.company_card_asset?.media_asset_id ?? '',
    address_zh: value.address_zh ?? '', address_en: value.address_en ?? '', is_big_four: value.is_big_four ?? false,
    logo_asset_id: value.logo_asset?.media_asset_id ?? '', business_registration_asset_id: value.business_registration_asset?.media_asset_id ?? '',
  });
  assets.avatar = value.avatar_asset ?? undefined;
  assets.wechat_qr = value.wechat_qr_asset ?? undefined;
  assets.eaa_license = value.eaa_license_asset ?? undefined;
  assets.company_card = value.company_card_asset ?? undefined;
  assets.logo = value.logo_asset ?? undefined;
  assets.business_registration = value.business_registration_asset ?? undefined;
};

// 3. 讀取代理資料及公司子帳戶
const loadProfile = async (): Promise<void> => {
  isLoading.value = true;
  try {
    const { data } = await fetchMyAgencyProfile();
    activeProfile.value = data.data.active_profile;
    revision.value = data.data.revision;
    nextEditableAt.value = data.data.next_editable_at ?? null;
    syncForm(revision.value ?? activeProfile.value);
    savedPayloadSnapshot.value = JSON.stringify(buildPayload());
    if (profileType.value === 'company' && activeProfile.value?.status === 'approved') {
      const subaccountResponse = await fetchAgencySubaccounts();
      subaccounts.value = subaccountResponse.data.data.items;
    }
  } catch (error) {
    feedbackStore.pushToast(readErrorMessage(error, t('account.agencyCompany.loadError')), 'error');
  } finally {
    isLoading.value = false;
  }
};

const workingProfile = computed(() => revision.value ?? activeProfile.value);
const status = computed<AgencyProfileStatus>(() => workingProfile.value?.status ?? 'draft');
const isReadonly = computed(() => isAgencyProfileReadonly(status.value));
const canEdit = computed(() => !isReadonly.value && (!nextEditableAt.value || Date.now() >= new Date(nextEditableAt.value).getTime()));
const individualLicenseRequired = computed(() => isIndividualAgencyLicenseRequired(form.is_overseas));
const isDirty = computed(() => JSON.stringify(buildPayload()) !== savedPayloadSnapshot.value);
const canSubmit = computed(() => canSubmitAgencyProfile(
  Boolean(revision.value), status.value, isDirty.value, canEdit.value,
));
const statusLabel = computed(() => t(`account.agencyCompany.status.${status.value}`));

// 4. 讓受限審核頁仍可安全結束目前登入
const handleSignOut = async (): Promise<void> => {
  if (isSigningOut.value) {
    return;
  }

  isSigningOut.value = true;
  try {
    await sessionStore.signOut();
    await router.push('/login');
  } finally {
    isSigningOut.value = false;
  }
};

// 5. 清除代理資料指定欄位錯誤
const clearProfileFieldError = (field: AgencyProfileField): void => {
  profileFieldErrors[field] = '';
};

// 6. 清除子帳戶指定欄位錯誤
const clearSubaccountFieldError = (field: AgencySubaccountField): void => {
  subaccountFieldErrors[field] = '';
};

// 7. 處理條件必填欄位狀態
const handleOverseasChange = (): void => {
  clearProfileFieldError('license_number');
  clearProfileFieldError('eaa_license_asset_id');
};
const handleDefaultAvatarChange = (): void => {
  clearProfileFieldError('avatar_asset_id');
};

// 8. 檢查電話及可選電郵格式
const isValidPhone = (countryCode: string, phoneNumber: string): boolean =>
  /^\+\d{1,7}$/.test(countryCode.trim()) && /^\d{4,32}$/.test(phoneNumber.trim());
const isValidOptionalEmail = (email: string): boolean => {
  const value = email.trim();
  return value === '' || (value.includes('@') && value.includes('.') && value.length <= 255);
};

// 9. 顯示第一個欄位校驗原因
const showFirstFieldError = (errors: Record<string, string>): void => {
  const message = Object.values(errors).find((value) => value !== '');
  if (message) {
    feedbackStore.pushToast(message, 'error');
  }
};

// 10. 設定必填欄位錯誤
const setProfileRequiredError = (field: AgencyProfileField, labelKey: string, value: string): void => {
  profileFieldErrors[field] = value.trim() === ''
    ? t('account.agencyCompany.fieldRequired', { field: t(labelKey) })
    : '';
};

// 11. 校驗提交審核所需資料
const validateProfileForReview = (): boolean => {
  Object.keys(profileFieldErrors).forEach((field) => {
    profileFieldErrors[field as AgencyProfileField] = '';
  });
  setProfileRequiredError('name_zh', profileType.value === 'individual'
    ? 'account.agencyCompany.nameZh'
    : 'account.agencyCompany.companyNameZh', form.name_zh);
  setProfileRequiredError('name_en', profileType.value === 'individual'
    ? 'account.agencyCompany.nameEn'
    : 'account.agencyCompany.companyNameEn', form.name_en);

  if (!isValidPhone(form.phone_1_country_code, form.phone_1_number)) {
    const message = t('account.agencyCompany.invalidPhone');
    profileFieldErrors.phone_1_country_code = message;
    profileFieldErrors.phone_1_number = message;
  }

  const licenseRequired = profileType.value === 'company' || individualLicenseRequired.value;
  if (licenseRequired) {
    setProfileRequiredError('license_number', 'account.agencyCompany.licenseNumber', form.license_number);
    profileFieldErrors.eaa_license_asset_id = form.eaa_license_asset_id
      ? ''
      : t('account.agencyCompany.fieldRequired', { field: t('account.agencyCompany.eaaLicense') });
  }

  if (profileType.value === 'individual' && form.default_avatar === 'custom') {
    profileFieldErrors.avatar_asset_id = form.avatar_asset_id
      ? ''
      : t('account.agencyCompany.fieldRequired', { field: t('account.agencyCompany.avatarCustom') });
  }

  return Object.values(profileFieldErrors).every((value) => value === '');
};

// 12. 校驗公司子帳戶資料
const validateSubaccount = (): boolean => {
  Object.keys(subaccountFieldErrors).forEach((field) => {
    subaccountFieldErrors[field as AgencySubaccountField] = '';
  });
  subaccountFieldErrors.display_name = subaccountForm.display_name.trim()
    ? ''
    : t('account.agencyCompany.fieldRequired', { field: t('account.agencyCompany.subaccountName') });
  if (!isValidPhone(subaccountForm.phone_country_code, subaccountForm.phone_number)) {
    const message = t('account.agencyCompany.invalidPhone');
    subaccountFieldErrors.phone_country_code = message;
    subaccountFieldErrors.phone_number = message;
  }
  subaccountFieldErrors.email = isValidOptionalEmail(subaccountForm.email)
    ? ''
    : t('account.agencyCompany.invalidEmail');
  subaccountFieldErrors.password = subaccountForm.password.length >= 8
    ? ''
    : t('account.agencyCompany.passwordMinLength');
  subaccountFieldErrors.permissions = subaccountForm.permissions.length > 0
    ? ''
    : t('account.agencyCompany.subaccountPermissionRequired');

  return Object.values(subaccountFieldErrors).every((value) => value === '');
};

// 13. 建立符合後端統一欄位的儲存請求
const buildPayload = (): AgencyProfileUpsertPayload => ({
  profile_type: profileType.value, name_zh: form.name_zh.trim(), name_en: form.name_en.trim(),
  address_zh: profileType.value === 'company' ? form.address_zh.trim() : '',
  address_en: profileType.value === 'company' ? form.address_en.trim() : '',
  license_number: form.license_number.trim(), is_overseas: profileType.value === 'individual' && form.is_overseas,
  is_big_four: profileType.value === 'company' && form.is_big_four,
  phone_1_country_code: form.phone_1_country_code, phone_1_number: form.phone_1_number.trim(), phone_1_whatsapp: form.phone_1_whatsapp,
  phone_2_country_code: form.phone_2_country_code, phone_2_number: form.phone_2_number.trim(), phone_2_whatsapp: form.phone_2_whatsapp,
  wechat_id: profileType.value === 'individual' ? form.wechat_id.trim() : '',
  wechat_url: profileType.value === 'individual' ? form.wechat_url.trim() : '',
  signature_zh: profileType.value === 'individual' ? form.signature_zh.trim() : '',
  signature_en: profileType.value === 'individual' ? form.signature_en.trim() : '',
  ...buildAgencyProfileAssetPayload(profileType.value, form.default_avatar, form),
});

// 14. 儲存草稿
const saveDraft = async (): Promise<boolean> => {
  if (!canEdit.value) {
    feedbackStore.pushToast(t('account.agencyCompany.rateLimited'), 'error');
    return false;
  }
  isSaving.value = true;
  try {
    const request = resolveAgencyProfileSaveMethod(revision.value) === 'update'
      ? updateMyAgencyProfile
      : createMyAgencyProfile;
    const { data } = await request(buildPayload());
    activeProfile.value = data.data.active_profile;
    revision.value = data.data.revision;
    nextEditableAt.value = data.data.next_editable_at ?? null;
    savedPayloadSnapshot.value = JSON.stringify(buildPayload());
    feedbackStore.pushToast(t('account.agencyCompany.saveSuccess'), 'success');
    return true;
  } catch (error) {
    feedbackStore.pushToast(readErrorMessage(error, t('account.agencyCompany.saveError')), 'error');
    return false;
  } finally {
    isSaving.value = false;
  }
};

// 15. 提交人工審核
const submitReview = async (): Promise<void> => {
  if (!validateProfileForReview()) {
    showFirstFieldError(profileFieldErrors);
    return;
  }
  if (isDirty.value && !(await saveDraft())) return;
  if (!revision.value) return;
  isSubmitting.value = true;
  try {
    const { data } = await submitMyAgencyProfile();
    activeProfile.value = data.data.active_profile;
    revision.value = data.data.revision;
    nextEditableAt.value = data.data.next_editable_at ?? null;
    await sessionStore.loadCurrentUser();
    feedbackStore.pushToast(t('account.agencyCompany.submitSuccess'), 'success');
  } catch (error) {
    feedbackStore.pushToast(readErrorMessage(error, t('account.agencyCompany.submitError')), 'error');
  } finally {
    isSubmitting.value = false;
  }
};

// 16. 上傳指定用途圖片
const uploadAsset = async (event: Event, field: AssetField): Promise<void> => {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  input.value = '';
  if (!file) return;
  if (!file.type.startsWith('image/') || file.size > MAX_FILE_SIZE) {
    feedbackStore.pushToast(file.size > MAX_FILE_SIZE ? t('account.agencyCompany.fileTooLarge') : t('account.agencyCompany.invalidFile'), 'error');
    return;
  }
  uploadingField.value = field;
  try {
    const purposeMap: Record<AssetField, string> = profileType.value === 'individual'
      ? {
          avatar: 'agency_individual_avatar', wechat_qr: 'agency_individual_wechat_qr',
          eaa_license: 'agency_individual_eaa', company_card: 'agency_individual_company_card',
          logo: 'agency_individual_avatar', business_registration: 'agency_individual_eaa',
        }
      : {
          avatar: 'agency_company_logo', wechat_qr: 'agency_company_logo', eaa_license: 'agency_company_eaa',
          company_card: 'agency_company_company_card', logo: 'agency_company_logo',
          business_registration: 'agency_company_business_registration',
        };
    const purpose = purposeMap[field];
    const { data } = await createUploadPresign({ file_name: file.name, mime_type: file.type, file_size: file.size, purpose });
    const presign = data.data;
    const response = await fetch(presign.upload_url, { method: 'PUT', headers: buildUploadHeaders(presign.headers, file.type), body: file });
    if (!response.ok) throw new Error('upload_failed');
    const complete = await completeUpload({ object_key: presign.object_key, upload_token: presign.upload_token, mime_type: file.type, file_size: file.size });
    const asset = complete.data.data;
    assets[field] = asset;
    if (field === 'avatar') form.avatar_asset_id = asset.media_asset_id;
    if (field === 'wechat_qr') form.wechat_qr_asset_id = asset.media_asset_id;
    if (field === 'eaa_license') form.eaa_license_asset_id = asset.media_asset_id;
    if (field === 'company_card') form.company_card_asset_id = asset.media_asset_id;
    if (field === 'logo') form.logo_asset_id = asset.media_asset_id;
    if (field === 'business_registration') form.business_registration_asset_id = asset.media_asset_id;
    const profileFieldMap: Partial<Record<AssetField, AgencyProfileField>> = {
      avatar: 'avatar_asset_id',
      eaa_license: 'eaa_license_asset_id',
      business_registration: 'business_registration_asset_id',
    };
    const profileField = profileFieldMap[field];
    if (profileField) {
      clearProfileFieldError(profileField);
    }
  } catch (error) {
    feedbackStore.pushToast(readErrorMessage(error, t('account.agencyCompany.uploadError')), 'error');
  } finally {
    uploadingField.value = null;
  }
};

// 17. 移除尚未提交的媒體欄位
const removeAsset = (field: AssetField): void => {
  delete assets[field];
  if (field === 'avatar') form.avatar_asset_id = '';
  if (field === 'wechat_qr') form.wechat_qr_asset_id = '';
  if (field === 'eaa_license') form.eaa_license_asset_id = '';
  if (field === 'company_card') form.company_card_asset_id = '';
  if (field === 'logo') form.logo_asset_id = '';
  if (field === 'business_registration') form.business_registration_asset_id = '';
};

// 18. 建立公司子帳戶
const addSubaccount = async (): Promise<void> => {
  if (!validateSubaccount()) {
    showFirstFieldError(subaccountFieldErrors);
    return;
  }
  subaccountSaving.value = true;
  try {
    const { data } = await createAgencySubaccount({ ...subaccountForm });
    subaccounts.value.unshift(data.data);
    Object.assign(subaccountForm, { display_name: '', phone_country_code: '+852', phone_number: '', email: '', password: '', permissions: ['property_publish'] });
    Object.keys(subaccountFieldErrors).forEach((field) => {
      subaccountFieldErrors[field as AgencySubaccountField] = '';
    });
    showSubaccountForm.value = false;
  } catch (error) {
    feedbackStore.pushToast(readErrorMessage(error, t('account.agencyCompany.subaccountSaveError')), 'error');
  } finally {
    subaccountSaving.value = false;
  }
};

// 19. 更新或移除公司子帳戶
const setSubaccountStatus = async (item: AgencySubaccount, nextStatus: AgencySubaccountStatus): Promise<void> => {
  const { data } = await updateAgencySubaccountStatus(item.public_id, nextStatus);
  Object.assign(item, data.data);
};
const deleteSubaccount = async (item: AgencySubaccount): Promise<void> => {
  await removeAgencySubaccount(item.public_id);
  subaccounts.value = subaccounts.value.filter((value) => value.public_id !== item.public_id);
};
const togglePermission = (permission: AgencySubaccountPermission): void => {
  subaccountForm.permissions = subaccountForm.permissions.includes(permission)
    ? subaccountForm.permissions.filter((value) => value !== permission)
    : [...subaccountForm.permissions, permission];
  clearSubaccountFieldError('permissions');
};

onMounted(loadProfile);
</script>

<template>
  <section class="agency-profile-page">
    <header class="agency-profile-header">
      <div class="agency-profile-heading">
        <span class="agency-profile-heading-icon"><AppIcon :name="profileHeaderIcon" :size="21" /></span>
        <div>
          <p>{{ t('account.agencyCompany.kicker') }}</p>
          <h1>{{ profileType === 'individual' ? t('account.agencyCompany.individualTitle') : t('account.agencyCompany.companyTitle') }}</h1>
          <span>{{ t('account.agencyCompany.reviewNotice') }}</span>
        </div>
      </div>
      <div class="agency-profile-actions">
        <button
          type="button"
          class="agency-profile-signout"
          :disabled="isSigningOut"
          :aria-label="t('common.action.signOut')"
          :title="t('common.action.signOut')"
          @click="handleSignOut"
        >
          <AppIcon name="logout" :size="17" />
          <span>{{ t('common.action.signOutShort') }}</span>
        </button>
        <span class="status-pill" :class="`status-pill--${status}`">{{ statusLabel }}</span>
      </div>
    </header>

    <div v-if="isLoading" class="agency-panel">{{ t('common.status.loading') }}</div>
    <template v-else>
      <div v-if="activeProfile && revision" class="agency-version-note">
        <AppIcon name="clock" :size="17" />
        <span>{{ t('account.agencyCompany.activeRevisionNotice') }}</span>
      </div>
      <section v-if="activeProfile && revision" class="agency-panel agency-active-summary">
        <div>
          <span>{{ t('account.agencyCompany.activeKicker') }}</span>
          <strong>{{ activeProfile.name_zh || activeProfile.name_en }}</strong>
        </div>
        <div>
          <span>{{ t('account.agencyCompany.licenseNumber') }}</span>
          <strong>{{ activeProfile.license_number || t('account.agencyCompany.notAvailable') }}</strong>
        </div>
        <span class="status-pill status-pill--approved">{{ t('account.agencyCompany.status.approved') }}</span>
      </section>
      <div v-if="status === 'rejected' && workingProfile?.review_note" class="agency-rejected">
        <AppIcon name="close" :size="17" />
        <span><strong>{{ t('account.agencyCompany.reviewNote') }}</strong> {{ workingProfile.review_note }}</span>
      </div>
      <div v-if="!canEdit && nextEditableAt && status !== 'pending'" class="agency-version-note">
        <AppIcon name="clock" :size="17" />
        <span>{{ t('account.agencyCompany.rateLimitUntil', { date: formatDate(nextEditableAt, preferenceStore.locale) }) }}</span>
      </div>

      <form class="agency-panel agency-profile-form" novalidate @submit.prevent="submitReview">
        <fieldset :disabled="isReadonly || isSaving || isSubmitting">
          <template v-if="profileType === 'individual'">
            <div class="agency-section-heading"><AppIcon name="user" :size="18" /><h2>{{ t('account.agencyCompany.personalDetails') }}</h2></div>
            <div class="form-grid">
              <label :class="{ 'agency-field--error': profileFieldErrors.name_zh }"><span>{{ t('account.agencyCompany.nameZh') }} <i class="agency-required-mark" aria-hidden="true">*</i></span><input v-model.trim="form.name_zh" :aria-invalid="profileFieldErrors.name_zh ? 'true' : undefined" :aria-describedby="profileFieldErrors.name_zh ? 'agency-name-zh-error' : undefined" maxlength="120" required @input="clearProfileFieldError('name_zh')"><small v-if="profileFieldErrors.name_zh" id="agency-name-zh-error" class="agency-field-error" role="alert">{{ profileFieldErrors.name_zh }}</small></label>
              <label :class="{ 'agency-field--error': profileFieldErrors.name_en }"><span>{{ t('account.agencyCompany.nameEn') }} <i class="agency-required-mark" aria-hidden="true">*</i></span><input v-model.trim="form.name_en" :aria-invalid="profileFieldErrors.name_en ? 'true' : undefined" :aria-describedby="profileFieldErrors.name_en ? 'agency-name-en-error' : undefined" maxlength="120" required @input="clearProfileFieldError('name_en')"><small v-if="profileFieldErrors.name_en" id="agency-name-en-error" class="agency-field-error" role="alert">{{ profileFieldErrors.name_en }}</small></label>
            </div>
            <div class="form-grid">
              <label :class="{ 'agency-field--error': profileFieldErrors.phone_1_country_code || profileFieldErrors.phone_1_number }"><span>{{ t('account.agencyCompany.phoneOne') }} <i class="agency-required-mark" aria-hidden="true">*</i></span><div class="phone-row"><input v-model.trim="form.phone_1_country_code" :aria-invalid="profileFieldErrors.phone_1_country_code ? 'true' : undefined" :aria-describedby="profileFieldErrors.phone_1_country_code ? 'agency-phone-error' : undefined" maxlength="8" placeholder="+852" required @input="clearProfileFieldError('phone_1_country_code')"><input v-model.trim="form.phone_1_number" :aria-invalid="profileFieldErrors.phone_1_number ? 'true' : undefined" :aria-describedby="profileFieldErrors.phone_1_number ? 'agency-phone-error' : undefined" inputmode="tel" required @input="clearProfileFieldError('phone_1_number')"></div><small v-if="profileFieldErrors.phone_1_country_code || profileFieldErrors.phone_1_number" id="agency-phone-error" class="agency-field-error" role="alert">{{ profileFieldErrors.phone_1_country_code || profileFieldErrors.phone_1_number }}</small><small><input v-model="form.phone_1_whatsapp" type="checkbox"> WhatsApp</small></label>
              <label><span>{{ t('account.agencyCompany.phoneTwo') }}</span><div class="phone-row"><input v-model.trim="form.phone_2_country_code" maxlength="8" placeholder="+852"><input v-model.trim="form.phone_2_number" inputmode="tel"></div><small><input v-model="form.phone_2_whatsapp" type="checkbox"> WhatsApp</small></label>
              <label><span>{{ t('account.agencyCompany.wechatId') }}</span><input v-model.trim="form.wechat_id" maxlength="120"></label>
              <label><span>{{ t('account.agencyCompany.wechatQrUrl') }}</span><input v-model.trim="form.wechat_url" type="url" placeholder="https://"></label>
            </div>
            <label><span>{{ t('account.agencyCompany.signatureZh') }}</span><textarea v-model="form.signature_zh" maxlength="300"></textarea><small>{{ form.signature_zh.length }}/300</small></label>
            <label><span>{{ t('account.agencyCompany.signatureEn') }}</span><textarea v-model="form.signature_en" maxlength="1000"></textarea><small>{{ form.signature_en.length }}/1000</small></label>
            <label class="check-row"><input v-model="form.is_overseas" type="checkbox" @change="handleOverseasChange"> {{ t('account.agencyCompany.overseasAgent') }}</label>
            <p v-if="form.is_overseas" class="field-hint">{{ t('account.agencyCompany.overseasLicenseOptional') }}</p>
            <div class="form-grid">
              <label><span>{{ t('account.agencyCompany.defaultAvatar') }} <i class="agency-required-mark" aria-hidden="true">*</i></span><select v-model="form.default_avatar" @change="handleDefaultAvatarChange"><option value="male">{{ t('account.agencyCompany.avatarMale') }}</option><option value="female">{{ t('account.agencyCompany.avatarFemale') }}</option><option value="custom">{{ t('account.agencyCompany.avatarCustom') }}</option></select></label>
              <label :class="{ 'agency-field--error': profileFieldErrors.license_number }"><span>{{ t('account.agencyCompany.licenseNumber') }} <i v-if="individualLicenseRequired" class="agency-required-mark" aria-hidden="true">*</i><small v-else>{{ t('account.agencyCompany.optionalLabel') }}</small></span><input v-model.trim="form.license_number" :aria-invalid="profileFieldErrors.license_number ? 'true' : undefined" :aria-describedby="profileFieldErrors.license_number ? 'agency-license-error' : undefined" maxlength="120" :required="individualLicenseRequired" @input="clearProfileFieldError('license_number')"><small v-if="profileFieldErrors.license_number" id="agency-license-error" class="agency-field-error" role="alert">{{ profileFieldErrors.license_number }}</small></label>
            </div>
          </template>

          <template v-else>
            <div class="agency-section-heading"><AppIcon name="building" :size="18" /><h2>{{ t('account.agencyCompany.companyDetails') }}</h2></div>
            <div class="form-grid">
              <label :class="{ 'agency-field--error': profileFieldErrors.name_zh }"><span>{{ t('account.agencyCompany.companyNameZh') }} <i class="agency-required-mark" aria-hidden="true">*</i></span><input v-model.trim="form.name_zh" :aria-invalid="profileFieldErrors.name_zh ? 'true' : undefined" :aria-describedby="profileFieldErrors.name_zh ? 'agency-name-zh-error' : undefined" maxlength="120" required @input="clearProfileFieldError('name_zh')"><small v-if="profileFieldErrors.name_zh" id="agency-name-zh-error" class="agency-field-error" role="alert">{{ profileFieldErrors.name_zh }}</small></label>
              <label :class="{ 'agency-field--error': profileFieldErrors.name_en }"><span>{{ t('account.agencyCompany.companyNameEn') }} <i class="agency-required-mark" aria-hidden="true">*</i></span><input v-model.trim="form.name_en" :aria-invalid="profileFieldErrors.name_en ? 'true' : undefined" :aria-describedby="profileFieldErrors.name_en ? 'agency-name-en-error' : undefined" maxlength="160" required @input="clearProfileFieldError('name_en')"><small v-if="profileFieldErrors.name_en" id="agency-name-en-error" class="agency-field-error" role="alert">{{ profileFieldErrors.name_en }}</small></label>
              <label :class="{ 'agency-field--error': profileFieldErrors.license_number }"><span>{{ t('account.agencyCompany.licenseNumber') }} <i class="agency-required-mark" aria-hidden="true">*</i></span><input v-model.trim="form.license_number" :aria-invalid="profileFieldErrors.license_number ? 'true' : undefined" :aria-describedby="profileFieldErrors.license_number ? 'agency-license-error' : undefined" maxlength="120" required @input="clearProfileFieldError('license_number')"><small v-if="profileFieldErrors.license_number" id="agency-license-error" class="agency-field-error" role="alert">{{ profileFieldErrors.license_number }}</small></label>
              <label><span>{{ t('account.agencyCompany.bigFour') }} <i class="agency-required-mark" aria-hidden="true">*</i></span><select v-model="form.is_big_four"><option :value="false">{{ t('account.agencyCompany.no') }}</option><option :value="true">{{ t('account.agencyCompany.yes') }}</option></select></label>
              <label :class="{ 'agency-field--error': profileFieldErrors.phone_1_country_code || profileFieldErrors.phone_1_number }"><span>{{ t('account.agencyCompany.phoneOne') }} <i class="agency-required-mark" aria-hidden="true">*</i></span><div class="phone-row"><input v-model.trim="form.phone_1_country_code" :aria-invalid="profileFieldErrors.phone_1_country_code ? 'true' : undefined" :aria-describedby="profileFieldErrors.phone_1_country_code ? 'agency-phone-error' : undefined" maxlength="8" placeholder="+852" required @input="clearProfileFieldError('phone_1_country_code')"><input v-model.trim="form.phone_1_number" :aria-invalid="profileFieldErrors.phone_1_number ? 'true' : undefined" :aria-describedby="profileFieldErrors.phone_1_number ? 'agency-phone-error' : undefined" inputmode="tel" required @input="clearProfileFieldError('phone_1_number')"></div><small v-if="profileFieldErrors.phone_1_country_code || profileFieldErrors.phone_1_number" id="agency-phone-error" class="agency-field-error" role="alert">{{ profileFieldErrors.phone_1_country_code || profileFieldErrors.phone_1_number }}</small><small><input v-model="form.phone_1_whatsapp" type="checkbox"> WhatsApp</small></label>
              <label><span>{{ t('account.agencyCompany.phoneTwo') }}</span><div class="phone-row"><input v-model.trim="form.phone_2_country_code" maxlength="8" placeholder="+852"><input v-model.trim="form.phone_2_number" inputmode="tel"></div><small><input v-model="form.phone_2_whatsapp" type="checkbox"> WhatsApp</small></label>
            </div>
          </template>

          <div class="agency-section-heading agency-documents-heading"><AppIcon name="picture" :size="18" /><h2>{{ t('account.agencyCompany.documents') }}</h2></div>
          <div class="asset-grid">
            <label v-if="profileType === 'individual' && form.default_avatar === 'custom'" class="asset-field" :class="{ 'agency-field--error': profileFieldErrors.avatar_asset_id }"><span>{{ t('account.agencyCompany.avatarCustom') }} <i class="agency-required-mark" aria-hidden="true">*</i></span><img v-if="assets.avatar" :src="assets.avatar.url" alt=""><input type="file" accept="image/*" :aria-invalid="profileFieldErrors.avatar_asset_id ? 'true' : undefined" :aria-describedby="profileFieldErrors.avatar_asset_id ? 'agency-avatar-error' : undefined" @change="uploadAsset($event, 'avatar')"><small v-if="profileFieldErrors.avatar_asset_id" id="agency-avatar-error" class="agency-field-error" role="alert">{{ profileFieldErrors.avatar_asset_id }}</small><button v-if="assets.avatar" type="button" @click="removeAsset('avatar')">{{ t('account.agencyCompany.removeFile') }}</button></label>
            <label v-if="profileType === 'individual'" class="asset-field"><span>{{ t('account.agencyCompany.wechatQrImage') }}</span><img v-if="assets.wechat_qr" :src="assets.wechat_qr.url" alt=""><input type="file" accept="image/*" @change="uploadAsset($event, 'wechat_qr')"><button v-if="assets.wechat_qr" type="button" @click="removeAsset('wechat_qr')">{{ t('account.agencyCompany.removeFile') }}</button></label>
            <label v-if="profileType === 'company'" class="asset-field"><span>{{ t('account.agencyCompany.logo') }}</span><img v-if="assets.logo" :src="assets.logo.url" alt=""><input type="file" accept="image/*" @change="uploadAsset($event, 'logo')"><button v-if="assets.logo" type="button" @click="removeAsset('logo')">{{ t('account.agencyCompany.removeFile') }}</button></label>
            <label class="asset-field" :class="{ 'agency-field--error': profileFieldErrors.eaa_license_asset_id }"><span>{{ t('account.agencyCompany.eaaLicense') }} <i v-if="profileType === 'company' || individualLicenseRequired" class="agency-required-mark" aria-hidden="true">*</i><small v-else>{{ t('account.agencyCompany.optionalLabel') }}</small></span><img v-if="assets.eaa_license" :src="assets.eaa_license.url" alt=""><input type="file" accept="image/*" :aria-invalid="profileFieldErrors.eaa_license_asset_id ? 'true' : undefined" :aria-describedby="profileFieldErrors.eaa_license_asset_id ? 'agency-eaa-error' : undefined" @change="uploadAsset($event, 'eaa_license')"><small v-if="profileFieldErrors.eaa_license_asset_id" id="agency-eaa-error" class="agency-field-error" role="alert">{{ profileFieldErrors.eaa_license_asset_id }}</small><button v-if="assets.eaa_license" type="button" @click="removeAsset('eaa_license')">{{ t('account.agencyCompany.removeFile') }}</button></label>
            <label class="asset-field"><span>{{ t('account.agencyCompany.companyCard') }}</span><img v-if="assets.company_card" :src="assets.company_card.url" alt=""><input type="file" accept="image/*" @change="uploadAsset($event, 'company_card')"><button v-if="assets.company_card" type="button" @click="removeAsset('company_card')">{{ t('account.agencyCompany.removeFile') }}</button></label>
          </div>
        </fieldset>
        <div v-if="!isReadonly" class="form-actions">
          <button type="button" :disabled="!canEdit || !isDirty || isSaving" @click="saveDraft"><AppIcon name="check-circle" :size="16" />{{ isSaving ? t('account.agencyCompany.saving') : t('account.agencyCompany.saveDraftAction') }}</button>
          <button class="primary" type="submit" :disabled="!canSubmit || isSaving || isSubmitting || Boolean(uploadingField)"><AppIcon name="send" :size="16" />{{ isSubmitting ? t('account.agencyCompany.submitting') : t('account.agencyCompany.submitAction') }}</button>
        </div>
        <p v-else class="field-hint agency-pending-hint"><AppIcon name="clock" :size="16" />{{ t('account.agencyCompany.pendingHint') }}</p>
      </form>

      <section v-if="profileType === 'company' && activeProfile?.status === 'approved'" class="agency-panel">
        <header class="subaccount-header"><div><h2>{{ t('account.agencyCompany.subaccounts') }}</h2><p>{{ t('account.agencyCompany.subaccountsHint') }}</p></div><button class="primary" type="button" @click="showSubaccountForm = !showSubaccountForm"><AppIcon name="plus-square" :size="16" />{{ t('account.agencyCompany.addSubaccount') }}</button></header>
        <form v-if="showSubaccountForm" class="subaccount-form" novalidate @submit.prevent="addSubaccount">
          <div class="form-grid">
            <label :class="{ 'agency-field--error': subaccountFieldErrors.display_name }"><span>{{ t('account.agencyCompany.subaccountName') }} <i class="agency-required-mark" aria-hidden="true">*</i></span><input v-model.trim="subaccountForm.display_name" :aria-invalid="subaccountFieldErrors.display_name ? 'true' : undefined" :aria-describedby="subaccountFieldErrors.display_name ? 'subaccount-name-error' : undefined" required @input="clearSubaccountFieldError('display_name')"><small v-if="subaccountFieldErrors.display_name" id="subaccount-name-error" class="agency-field-error" role="alert">{{ subaccountFieldErrors.display_name }}</small></label>
            <label :class="{ 'agency-field--error': subaccountFieldErrors.email }"><span>{{ t('account.agencyCompany.subaccountEmail') }}</span><input v-model.trim="subaccountForm.email" :aria-invalid="subaccountFieldErrors.email ? 'true' : undefined" :aria-describedby="subaccountFieldErrors.email ? 'subaccount-email-error' : undefined" type="email" @input="clearSubaccountFieldError('email')"><small v-if="subaccountFieldErrors.email" id="subaccount-email-error" class="agency-field-error" role="alert">{{ subaccountFieldErrors.email }}</small></label>
            <label :class="{ 'agency-field--error': subaccountFieldErrors.phone_country_code || subaccountFieldErrors.phone_number }"><span>{{ t('account.agencyCompany.phoneOne') }} <i class="agency-required-mark" aria-hidden="true">*</i></span><div class="phone-row"><input v-model.trim="subaccountForm.phone_country_code" :aria-invalid="subaccountFieldErrors.phone_country_code ? 'true' : undefined" :aria-describedby="subaccountFieldErrors.phone_country_code ? 'subaccount-phone-error' : undefined" maxlength="8" placeholder="+852" required @input="clearSubaccountFieldError('phone_country_code')"><input v-model.trim="subaccountForm.phone_number" :aria-invalid="subaccountFieldErrors.phone_number ? 'true' : undefined" :aria-describedby="subaccountFieldErrors.phone_number ? 'subaccount-phone-error' : undefined" inputmode="tel" required @input="clearSubaccountFieldError('phone_number')"></div><small v-if="subaccountFieldErrors.phone_country_code || subaccountFieldErrors.phone_number" id="subaccount-phone-error" class="agency-field-error" role="alert">{{ subaccountFieldErrors.phone_country_code || subaccountFieldErrors.phone_number }}</small></label>
            <label :class="{ 'agency-field--error': subaccountFieldErrors.password }"><span>{{ t('account.agencyCompany.subaccountPassword') }} <i class="agency-required-mark" aria-hidden="true">*</i></span><input v-model="subaccountForm.password" :aria-invalid="subaccountFieldErrors.password ? 'true' : undefined" :aria-describedby="subaccountFieldErrors.password ? 'subaccount-password-error' : undefined" type="password" minlength="8" required @input="clearSubaccountFieldError('password')"><small v-if="subaccountFieldErrors.password" id="subaccount-password-error" class="agency-field-error" role="alert">{{ subaccountFieldErrors.password }}</small></label>
          </div>
          <div class="permission-row" :class="{ 'agency-field--error': subaccountFieldErrors.permissions }"><span>{{ t('account.agencyCompany.permissionPublish') }} / {{ t('account.agencyCompany.permissionManage') }} <i class="agency-required-mark" aria-hidden="true">*</i></span><label><input :checked="subaccountForm.permissions.includes('property_publish')" type="checkbox" @change="togglePermission('property_publish')">{{ t('account.agencyCompany.permissionPublish') }}</label><label><input :checked="subaccountForm.permissions.includes('property_manage')" type="checkbox" @change="togglePermission('property_manage')">{{ t('account.agencyCompany.permissionManage') }}</label><small v-if="subaccountFieldErrors.permissions" class="agency-field-error" role="alert">{{ subaccountFieldErrors.permissions }}</small></div>
          <div class="form-actions"><button type="button" @click="showSubaccountForm = false">{{ t('account.agencyCompany.cancelAction') }}</button><button class="primary" :disabled="subaccountSaving" type="submit">{{ t('account.agencyCompany.createSubaccount') }}</button></div>
        </form>
        <div class="subaccount-list"><article v-for="item in subaccounts" :key="item.public_id"><div><strong>{{ item.display_name }}</strong><span>{{ item.phone_country_code }} {{ item.phone_number }} · {{ item.email || '-' }}</span></div><span class="status-pill">{{ item.status }}</span><div class="subaccount-actions"><button type="button" @click="setSubaccountStatus(item, item.status === 'active' ? 'disabled' : 'active')">{{ item.status === 'active' ? t('account.agencyCompany.disableSubaccount') : t('account.agencyCompany.enableSubaccount') }}</button><button type="button" @click="deleteSubaccount(item)">{{ t('account.agencyCompany.removeSubaccount') }}</button></div></article><p v-if="subaccounts.length === 0" class="field-hint">{{ t('account.agencyCompany.noSubaccounts') }}</p></div>
      </section>
    </template>
  </section>
</template>

<style scoped>
.agency-profile-page{display:grid;gap:1rem;min-width:0}.agency-profile-header,.subaccount-header{display:flex;align-items:flex-start;justify-content:space-between;gap:1rem}.agency-profile-header p{margin:0;color:rgb(var(--color-primary));font-size:.72rem;font-weight:800;text-transform:uppercase}.agency-profile-header h1,.agency-panel h2{margin:.2rem 0;color:rgb(var(--color-text))}.agency-profile-header span,.subaccount-header p,.field-hint{color:rgb(var(--color-text-muted));font-size:.82rem}.agency-panel{display:grid;gap:1rem;border:1px solid rgb(var(--color-border));border-radius:6px;background:rgb(var(--color-surface));padding:1rem}.agency-panel fieldset{display:grid;gap:1rem;margin:0;border:0;padding:0}.form-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:1rem}.agency-panel label{display:grid;gap:.4rem;color:rgb(var(--color-text));font-size:.8rem;font-weight:700}.agency-panel input:not([type=checkbox]):not([type=file]),.agency-panel select,.agency-panel textarea{width:100%;min-height:42px;border:1px solid rgb(var(--color-border));border-radius:4px;background:rgb(var(--color-surface));color:rgb(var(--color-text));padding:.6rem .7rem}.agency-panel textarea{min-height:90px;resize:vertical}.phone-row{display:grid;grid-template-columns:90px minmax(0,1fr);gap:.5rem}.check-row{display:flex!important;align-items:center}.agency-version-note,.agency-rejected{border-left:3px solid rgb(var(--color-primary));background:rgb(var(--color-surface-muted));padding:.75rem 1rem;color:rgb(var(--color-text));font-size:.82rem}.agency-rejected{border-color:rgb(var(--color-danger));color:rgb(var(--color-danger))}.status-pill{flex:none;border:1px solid rgb(var(--color-border));border-radius:999px;background:rgb(var(--color-surface-muted));padding:.35rem .65rem;color:rgb(var(--color-text));font-size:.72rem;font-weight:800}.asset-grid{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:.75rem}.asset-field{border:1px dashed rgb(var(--color-border));padding:.75rem}.asset-field img{width:100%;aspect-ratio:4/3;object-fit:contain;background:rgb(var(--color-surface-muted))}.asset-field input{max-width:100%;font-size:.72rem}.agency-panel button{min-height:38px;border:1px solid rgb(var(--color-border));border-radius:4px;background:rgb(var(--color-surface));color:rgb(var(--color-text));padding:.5rem .8rem;font-weight:750}.agency-panel button.primary{border-color:rgb(var(--color-primary));background:rgb(var(--color-primary));color:white}.form-actions,.permission-row,.subaccount-actions{display:flex;justify-content:flex-end;gap:.6rem;flex-wrap:wrap}.permission-row{justify-content:flex-start}.permission-row label{display:flex;align-items:center}.subaccount-form{display:grid;gap:1rem;border-top:1px solid rgb(var(--color-border));padding-top:1rem}.subaccount-list{display:grid;gap:.6rem}.subaccount-list article{display:grid;grid-template-columns:minmax(0,1fr) auto auto;align-items:center;gap:.8rem;border-top:1px solid rgb(var(--color-border));padding-top:.7rem}.subaccount-list article div:first-child{display:grid;gap:.2rem}.subaccount-list article span{color:rgb(var(--color-text-muted));font-size:.75rem}@media(max-width:767px){.agency-profile-header,.subaccount-header{align-items:stretch;flex-direction:column}.form-grid,.asset-grid{grid-template-columns:1fr}.form-actions button,.subaccount-header button{flex:1}.subaccount-list article{grid-template-columns:minmax(0,1fr) auto}.subaccount-actions{grid-column:1/-1}.subaccount-actions button{flex:1}}
.agency-active-summary{grid-template-columns:minmax(0,1fr) minmax(0,1fr) auto;align-items:center}.agency-active-summary div{display:grid;gap:.2rem}.agency-active-summary div span{color:rgb(var(--color-text-muted));font-size:.72rem}.agency-active-summary strong{color:rgb(var(--color-text));font-size:.88rem}@media(max-width:767px){.agency-active-summary{grid-template-columns:1fr}.agency-active-summary .status-pill{justify-self:start}}

/* 12. 代理資料頁視覺層級與操作狀態 */
.agency-profile-page {
  gap: 0.9rem;
}

.agency-profile-header {
  align-items: center;
  border-bottom: 1px solid rgb(var(--color-border));
  padding: 0 0 0.95rem;
}

.agency-profile-heading {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 0.8rem;
}

.agency-profile-heading-icon {
  display: grid;
  width: 2.8rem;
  height: 2.8rem;
  flex: 0 0 auto;
  place-items: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: rgb(var(--color-surface-muted));
  color: rgb(var(--color-primary));
}

.agency-profile-heading > div {
  min-width: 0;
}

.agency-profile-header h1 {
  margin: 0.15rem 0 0;
  font-family: var(--font-display);
  font-size: clamp(1.35rem, 2vw, 1.85rem);
  font-weight: 650;
  line-height: 1.22;
}

.agency-profile-header span:not(.agency-profile-heading-icon):not(.status-pill) {
  display: block;
  margin-top: 0.25rem;
  line-height: 1.5;
}

.agency-panel {
  gap: 0;
  overflow: hidden;
  border-radius: 8px;
  background: rgb(var(--color-surface-raised));
  padding: 0;
}

.agency-profile-form fieldset {
  gap: 0.95rem;
  padding: 1rem 1.1rem 1.1rem;
}

.agency-section-heading {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: rgb(var(--color-primary));
}

.agency-section-heading h2 {
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 0.95rem;
  font-weight: 780;
}

.agency-documents-heading {
  margin-top: 0.35rem;
  border-top: 1px solid rgb(var(--color-border));
  padding-top: 1rem;
}

.agency-panel label {
  gap: 0.42rem;
  font-size: 0.76rem;
  line-height: 1.35;
}

.agency-panel label > span {
  color: rgb(var(--color-text-muted));
  font-weight: 760;
}

.agency-panel input:not([type=checkbox]):not([type=file]),
.agency-panel select,
.agency-panel textarea {
  min-height: 2.7rem;
  border-radius: 4px;
  background: rgb(var(--color-surface));
  font-size: 0.875rem;
  font-weight: 620;
  outline: none;
  transition: border-color 0.18s ease, box-shadow 0.18s ease;
}

.agency-panel input:not([type=checkbox]):not([type=file]):focus,
.agency-panel select:focus,
.agency-panel textarea:focus {
  border-color: rgb(var(--color-primary));
  box-shadow: 0 0 0 2px rgb(var(--color-primary) / 0.1);
}

.agency-panel textarea {
  min-height: 6.25rem;
  line-height: 1.55;
}

.agency-panel input[type='checkbox'] {
  width: 1rem;
  height: 1rem;
  accent-color: rgb(var(--color-primary));
}

.check-row {
  width: fit-content;
  min-height: 2rem;
  flex-direction: row;
  gap: 0.5rem !important;
  border: 1px solid rgb(var(--color-border));
  border-radius: 4px;
  background: rgb(var(--color-surface-muted));
  padding: 0.4rem 0.65rem;
}

.asset-grid {
  gap: 0.7rem;
}

.asset-field {
  min-width: 0;
  align-content: start;
  border-style: solid;
  border-radius: 6px;
  background: rgb(var(--color-surface));
  padding: 0.7rem;
}

.asset-field img {
  border: 1px solid rgb(var(--color-border));
  border-radius: 4px;
}

.asset-field input[type='file'] {
  width: 100%;
  color: rgb(var(--color-text-muted));
}

.asset-field input[type='file']::file-selector-button {
  min-height: 2.1rem;
  margin-right: 0.45rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 4px;
  background: rgb(var(--color-surface-muted));
  padding: 0 0.6rem;
  color: rgb(var(--color-text));
  font-size: 0.72rem;
  font-weight: 700;
}

.agency-panel button {
  display: inline-flex;
  min-height: 2.35rem;
  align-items: center;
  justify-content: center;
  gap: 0.42rem;
  border-radius: 4px;
  font-size: 0.78rem;
  line-height: 1;
  transition: border-color 0.18s ease, background 0.18s ease, opacity 0.18s ease;
}

.agency-panel button:hover:not(:disabled) {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-surface-muted));
}

.agency-panel button.primary:hover:not(:disabled) {
  border-color: rgb(var(--color-text));
  background: rgb(var(--color-text));
}

.agency-panel button:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.form-actions {
  border-top: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface-muted));
  padding: 0.85rem 1.1rem;
}

.agency-pending-hint {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  margin: 0;
  border-top: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-warning-bg));
  padding: 0.85rem 1.1rem;
  color: rgb(var(--color-warning));
  font-weight: 700;
}

.agency-version-note,
.agency-rejected {
  display: flex;
  align-items: flex-start;
  gap: 0.55rem;
  border: 1px solid rgb(var(--color-border));
  border-left-width: 3px;
  border-radius: 4px;
  background: rgb(var(--color-surface-muted));
  line-height: 1.55;
}

.agency-version-note > svg,
.agency-rejected > svg {
  flex: 0 0 auto;
  margin-top: 0.05rem;
}

.agency-rejected {
  border-color: rgb(var(--color-danger));
  background: rgb(var(--color-danger-bg));
}

.status-pill {
  min-height: 1.75rem;
  display: inline-flex;
  align-items: center;
  border: 0;
  border-radius: 999px;
  padding: 0 0.65rem;
}

.status-pill--pending {
  background: rgb(var(--color-warning-bg));
  color: rgb(var(--color-warning));
}

.status-pill--approved {
  background: rgb(var(--color-success-bg));
  color: rgb(var(--color-success));
}

.status-pill--rejected {
  background: rgb(var(--color-danger-bg));
  color: rgb(var(--color-danger));
}

.status-pill--draft {
  background: rgb(var(--color-surface-muted));
  color: rgb(var(--color-text-muted));
}

.agency-active-summary {
  padding: 0.85rem 1rem;
}

.subaccount-header,
.subaccount-form,
.subaccount-list {
  padding: 1rem 1.1rem;
}

.subaccount-header {
  align-items: center;
  border-bottom: 1px solid rgb(var(--color-border));
}

.subaccount-header h2,
.subaccount-header p {
  margin: 0;
}

.subaccount-header p {
  margin-top: 0.25rem;
}

.subaccount-form {
  border-top: 0;
  border-bottom: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface-muted));
}

.subaccount-list article {
  border: 1px solid rgb(var(--color-border));
  border-radius: 4px;
  background: rgb(var(--color-surface));
  padding: 0.7rem 0.75rem;
}

.agency-required-mark {
  color: rgb(var(--color-primary));
  font-style: normal;
}

.agency-field--error input:not([type='checkbox']):not([type='file']),
.agency-field--error select,
.agency-field--error textarea {
  border-color: rgb(var(--color-danger));
  box-shadow: 0 0 0 2px rgb(var(--color-danger) / 0.1);
}

.agency-field-error {
  color: rgb(var(--color-danger));
  font-size: 0.75rem;
  font-weight: 700;
  line-height: 1.4;
}

.permission-row > span {
  width: 100%;
  color: rgb(var(--color-text));
  font-size: 0.8rem;
  font-weight: 700;
}

.permission-row .agency-field-error {
  width: 100%;
}

.agency-profile-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.agency-profile-signout {
  display: inline-flex;
  min-height: 2rem;
  align-items: center;
  gap: 0.35rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 4px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  padding: 0.35rem 0.55rem;
  font: inherit;
  font-size: 0.78rem;
  cursor: pointer;
}

.agency-profile-signout span {
  display: inline;
  margin: 0;
  color: inherit;
  line-height: 1;
}

.agency-profile-signout:disabled {
  cursor: wait;
  opacity: 0.65;
}

@media (max-width: 767px) {
  .agency-profile-header {
    align-items: flex-start;
  }

  .agency-profile-heading {
    align-items: flex-start;
  }

  .agency-profile-heading-icon {
    width: 2.45rem;
    height: 2.45rem;
  }

  .agency-profile-header .status-pill {
    align-self: flex-start;
  }

  .agency-profile-actions {
    width: 100%;
    justify-content: space-between;
  }

  .agency-profile-form fieldset,
  .subaccount-header,
  .subaccount-form,
  .subaccount-list {
    padding-right: 0.85rem;
    padding-left: 0.85rem;
  }

  .phone-row {
    grid-template-columns: 5.25rem minmax(0, 1fr);
  }

  .form-actions {
    position: sticky;
    z-index: 3;
    bottom: calc(var(--app-safe-bottom) + 4rem);
    padding-right: 0.85rem;
    padding-left: 0.85rem;
  }

  .form-actions button {
    min-height: 2.65rem;
  }
}
</style>
