/*
 * 代理資料修訂狀態測試。
 * 1. 驗證已通過資料建立修訂，草稿或被拒修訂繼續更新。
 * 2. 驗證審核中修訂不可編輯。
 */
import { describe, expect, it } from 'vitest';

import type { AgencyProfile } from '@/model/agency-profile';

import {
  buildAgencyProfileAssetPayload,
  canSubmitAgencyProfile,
  isIndividualAgencyLicenseRequired,
  isAgencyProfileReadonly,
  resolveAgencyProfileSaveMethod,
} from './profile-state';

const revision = (status: AgencyProfile['status']): AgencyProfile => ({ status } as AgencyProfile);

describe('agency profile revision state', () => {
  it('makes the Hong Kong licence optional for overseas individual agents', () => {
    expect(isIndividualAgencyLicenseRequired(true)).toBe(false);
    expect(isIndividualAgencyLicenseRequired(false)).toBe(true);
  });

  it('creates a revision when only an approved active profile exists', () => {
    expect(resolveAgencyProfileSaveMethod(null)).toBe('create');
  });

  it('updates existing draft and rejected revisions', () => {
    expect(resolveAgencyProfileSaveMethod(revision('draft'))).toBe('update');
    expect(resolveAgencyProfileSaveMethod(revision('rejected'))).toBe('update');
  });

  it('keeps pending revisions read only', () => {
    expect(isAgencyProfileReadonly('pending')).toBe(true);
    expect(isAgencyProfileReadonly('rejected')).toBe(false);
  });

  it('submits a saved draft during edit cooldown without saving it again', () => {
    expect(canSubmitAgencyProfile(true, 'draft', false, false)).toBe(true);
    expect(canSubmitAgencyProfile(true, 'rejected', false, false)).toBe(true);
  });

  it('requires edit availability when the form has unsaved changes', () => {
    expect(canSubmitAgencyProfile(true, 'draft', true, false)).toBe(false);
    expect(canSubmitAgencyProfile(true, 'draft', true, true)).toBe(true);
    expect(canSubmitAgencyProfile(false, 'draft', false, true)).toBe(false);
  });

  it('keeps a custom individual avatar in the saved payload', () => {
    const payload = buildAgencyProfileAssetPayload('individual', 'custom', {
      avatar_asset_id: 'avatar-1',
      wechat_qr_asset_id: 'wechat-1',
      logo_asset_id: '',
      eaa_license_asset_id: 'eaa-1',
      business_registration_asset_id: '',
      company_card_asset_id: 'card-1',
    });

    expect(payload).toMatchObject({
      default_avatar: 'custom',
      avatar_asset_id: 'avatar-1',
      wechat_qr_asset_id: 'wechat-1',
      eaa_license_asset_id: 'eaa-1',
      company_card_asset_id: 'card-1',
      logo_asset_id: null,
      business_registration_asset_id: null,
    });
  });

  it('keeps company-only media separate from individual avatar fields', () => {
    const payload = buildAgencyProfileAssetPayload('company', 'male', {
      avatar_asset_id: 'stale-avatar',
      wechat_qr_asset_id: 'stale-wechat',
      logo_asset_id: 'logo-1',
      eaa_license_asset_id: 'eaa-1',
      business_registration_asset_id: 'br-1',
      company_card_asset_id: 'card-1',
    });

    expect(payload).toMatchObject({
      default_avatar: '',
      avatar_asset_id: null,
      wechat_qr_asset_id: null,
      logo_asset_id: 'logo-1',
      business_registration_asset_id: 'br-1',
    });
  });
});
