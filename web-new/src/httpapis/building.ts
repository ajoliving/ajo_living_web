/*
 * 大廈與 POS 樓宇 API。
 * 1. 讀取 POS relay 樓宇清單。
 * 2. 依大廈讀取可選單位清單。
 * 3. 讀取目前會員已綁定的 iSmart 大廈資料。
 * 4. 讀取目前會員已綁定大廈的 iSmart 智能門禁資料。
 * 5. 讀取目前會員已綁定大廈的 iCCTV 視像監控資料。
 * 6. 讀取目前會員已綁定大廈的 iSmart 管理費與其他費用資料。
 * 7. 讀取目前會員已綁定大廈的 iSmart 有效通告資料。
 * 8. 提交目前會員 iSmart 業主角色綁定申請。
 * 9. 讀取與管理目前會員 iSmart 授權副戶資料。
 */
import httpClient from '@/httpapis';
import type { ApiResponse } from '@/model/api';
import type { PosBuilding, PosBuildingUnit } from '@/model/community';

interface ApiListData<T> {
  items: T[];
}

export interface IsmartBuildingOptionResponse {
  building_options: string[];
  is_staff?: boolean;
}

export interface IsmartBuildingSummary {
  building_id?: string;
  buildname_chi?: string;
  buildname?: string;
  building_type?: string;
  area?: string;
  district?: string;
  street?: string;
  street_no?: string;
  court?: string;
  block?: string;
}

export interface IsmartBuildingInfo {
  year_built?: number | string | null;
  total_floor?: number | string | null;
  total_unit?: number | string | null;
  total_carpark?: number | string | null;
  google_map_url?: string | null;
  owners_corporation_name?: string | null;
  management_office_phone?: string | null;
  management_company_name?: string | null;
  management_company_phone?: string | null;
  management_company_email?: string | null;
  management_company_fax?: string | null;
  home_affairs_department_phone?: string | null;
}

export interface IsmartBuildingDocument {
  id?: number | string;
  title?: string;
  type?: string;
  file_url?: string;
  file_date?: string | null;
  file_month?: string | null;
  created_date?: string | null;
}

export interface IsmartBuildingDocuments {
  forms?: IsmartBuildingDocument[];
  building_info_files?: IsmartBuildingDocument[];
  floorplans?: IsmartBuildingDocument[];
  audit_reports?: IsmartBuildingDocument[];
  financial_reports?: IsmartBuildingDocument[];
  floorplan?: IsmartBuildingDocument[];
  auditreport?: IsmartBuildingDocument[];
  audition?: IsmartBuildingDocument[];
  audit_report?: IsmartBuildingDocument[];
  auditreports?: IsmartBuildingDocument[];
  auditions?: IsmartBuildingDocument[];
  mfinreport?: IsmartBuildingDocument[];
}

export interface IsmartBuildingInfoResponse extends IsmartBuildingOptionResponse {
  selected_building_id?: string;
  building?: IsmartBuildingSummary;
  building_info?: IsmartBuildingInfo;
  documents?: IsmartBuildingDocuments;
}

export interface IsmartAccessBuilding {
  requested_building_id?: string;
  access_building_id?: string;
  buildname_chi?: string;
  buildname?: string;
}

export interface IsmartAccessCamera {
  id?: number | string;
  title?: string;
  url?: string;
  source?: string;
}

export interface IsmartDoorPassword {
  record_id?: number | string;
  value?: string;
  start_time?: string;
  end_time?: string;
}

export interface IsmartDoorQRCode {
  record_id?: number | string;
  start_time?: string;
  end_time?: string;
}

export interface IsmartAccessDoor {
  door_id?: number | string;
  title?: string;
  serial?: string;
  door_no?: number | string;
  building_id?: string;
  is_public?: boolean;
  is_qrcode_enabled?: boolean;
  has_permission?: boolean;
  camera?: IsmartAccessCamera | null;
  password?: IsmartDoorPassword | null;
  qrcode?: IsmartDoorQRCode | null;
}

export interface IsmartAccessRecord {
  open_time?: string;
  open_type?: string;
  is_success?: boolean | number | string;
}

export interface IsmartRecentAccessGroup {
  door?: {
    id?: number | string;
    title?: string;
  };
  records?: IsmartAccessRecord[];
}

export interface IsmartBuildingAccessResponse extends IsmartBuildingOptionResponse {
  selected_building_id?: string;
  building?: IsmartAccessBuilding;
  doors?: IsmartAccessDoor[];
  recent_records?: IsmartRecentAccessGroup[];
}

export interface IsmartDoorOpenResponse extends IsmartBuildingOptionResponse {
  selected_building_id?: string;
  door_id?: number | string;
  building_id?: string;
  is_success?: boolean;
  upstream_message?: string;
}

export interface IsmartQRCodeResponse extends IsmartBuildingOptionResponse {
  selected_building_id?: string;
  qrcode_record_id?: number | string;
  door_id?: number | string;
  term?: string;
  qrcode_value?: string;
  expires_at?: string;
  upstream_message?: string;
}

export interface OpenIsmartDoorPayload {
  building_id?: string;
  door_id: number;
}

export interface GenerateIsmartQRCodePayload {
  building_id?: string;
  qrcode_record_id: number;
  term?: string;
}

export interface IsmartOwnerBindingRequestPayload {
  building_id: string;
  ownedflat: string[];
  cli_role?: string;
  ownernote?: string;
  is_receive_email?: boolean;
  reg_tel?: string;
  reg_email?: string;
  cli_name?: string;
  cli_id_card?: string;
  cli_tel?: string;
}

export interface IsmartOwnerBindingResponse extends IsmartBuildingOptionResponse {
  selected_building_id?: string;
  request_id?: number | string;
  id?: number | string;
  upstream_message?: string;
}

export interface IsmartSubaccountRow {
  relation_info_id?: number | string;
  building_id?: string;
  building_name?: string;
  unit_id?: string;
  unit_display?: string;
  target_user_id?: number | string;
  target_username?: string;
  target_client_id?: string;
  target_name?: string;
  target_phone?: string;
  target_email?: string;
  role?: string;
  history_count?: number;
}

export interface IsmartSubaccountsResponse extends IsmartBuildingOptionResponse {
  items?: IsmartSubaccountRow[];
  count?: number;
  upstream_message?: string;
}

export interface IsmartSubaccountMutationPayload {
  unit_id: string;
  target_user_id: number;
  remark?: string;
}

export interface IsmartSubaccountMutationResponse extends IsmartBuildingOptionResponse {
  history_id?: number | string;
  relation_info_id?: number | string;
  building_id?: string;
  unit_id?: string;
  target_user_id?: number | string;
  target_username?: string;
  role?: string;
  active_subaccount_count?: number;
  upstream_message?: string;
}

export interface ICCTVOrangePiSummary {
  orangepi_id: number;
  orangepi_name: string;
  is_active: boolean;
  camera_count: number;
}

export interface ICCTVCameraSummary {
  id: string;
  title: string;
  channel: string;
  url: string;
  orangepi_id: number;
  orangepi_name: string;
  is_active: boolean;
}

export interface ICCTVPublicCameraResponse extends IsmartBuildingOptionResponse {
  selected_building_id: string;
  orangepis: ICCTVOrangePiSummary[];
  cameras: ICCTVCameraSummary[];
}

export interface IsmartBuildingNotice {
  id?: number | string;
  mess_code?: string;
  mess_title?: string;
  mess_type?: string;
  mess_date?: string | null;
  mess_down?: string | null;
  mess_file?: string | null;
}

export interface IsmartBuildingNoticesResponse extends IsmartBuildingOptionResponse {
  selected_building_id?: string;
  upstream_message?: string;
  result?: IsmartBuildingNotice[];
}

export type IsmartReceivableCellValue = string | number | boolean | null;

export type IsmartReceivableDataStructure = 'table' | 'list' | 'block_dict';

export interface IsmartManagementFeeTableRow {
  [field: string]: IsmartReceivableCellValue | undefined;
}

export interface IsmartOtherFeeRow {
  invoice_no?: string;
  flat_code?: string;
  item_id?: string;
  trs_to?: string;
  trs_val?: number;
  remark?: string;
  [field: string]: IsmartReceivableCellValue | undefined;
}

export interface IsmartReceivableResponse<T> extends IsmartBuildingOptionResponse {
  selected_building_id?: string;
  upstream_message?: string;
  result?: T[];
  blocks?: unknown[];
}

// 1. 取得 POS 樓宇清單
export const fetchPosBuildings = async (): Promise<PosBuilding[]> => {
  const { data } = await httpClient.get<ApiResponse<ApiListData<PosBuilding>>>('/pos/buildings');
  return data.data.items;
};

// 2. 取得 POS 大廈單位清單
export const fetchPosBuildingUnits = async (buildingID: string): Promise<PosBuildingUnit[]> => {
  const { data } = await httpClient.get<ApiResponse<ApiListData<PosBuildingUnit>>>(`/pos/buildings/${encodeURIComponent(buildingID)}/units`);
  return data.data.items;
};

// 3. 取得目前會員可見 POS 樓宇清單
export const fetchMemberPosBuildings = async (): Promise<PosBuilding[]> => {
  const { data } = await httpClient.get<ApiResponse<ApiListData<PosBuilding>>>('/me/pos/buildings');
  return data.data.items;
};

// 4. 取得目前會員可見 POS 大廈單位清單
export const fetchMemberPosBuildingUnits = async (buildingID: string): Promise<PosBuildingUnit[]> => {
  const { data } = await httpClient.get<ApiResponse<ApiListData<PosBuildingUnit>>>(`/me/pos/buildings/${encodeURIComponent(buildingID)}/units`);
  return data.data.items;
};

// 5. 取得目前會員 iSmart 綁定大廈
export const fetchMemberIsmartBuildings = async (): Promise<IsmartBuildingOptionResponse> => {
  const { data } = await httpClient.get<ApiResponse<IsmartBuildingOptionResponse>>('/me/ismart/buildings');
  return data.data;
};

// 6. 提交目前會員 iSmart 業主角色綁定申請
export const submitMemberIsmartOwnerBindingRequest = async (
  payload: IsmartOwnerBindingRequestPayload,
): Promise<IsmartOwnerBindingResponse> => {
  const { data } = await httpClient.post<ApiResponse<IsmartOwnerBindingResponse>>('/me/ismart/owner-binding-requests', payload);
  return data.data;
};

// 7. 取得目前會員 iSmart 授權副戶
export const fetchMemberIsmartSubaccounts = async (unitID?: string): Promise<IsmartSubaccountsResponse> => {
  const params = unitID ? { unit_id: unitID } : undefined;
  const { data } = await httpClient.get<ApiResponse<IsmartSubaccountsResponse>>('/me/ismart/subaccounts', { params });
  return data.data;
};

// 8. 新增目前會員 iSmart 授權副戶
export const grantMemberIsmartSubaccount = async (
  payload: IsmartSubaccountMutationPayload,
): Promise<IsmartSubaccountMutationResponse> => {
  const { data } = await httpClient.post<ApiResponse<IsmartSubaccountMutationResponse>>('/me/ismart/subaccounts/grant', payload);
  return data.data;
};

// 9. 撤銷目前會員 iSmart 授權副戶
export const revokeMemberIsmartSubaccount = async (
  payload: IsmartSubaccountMutationPayload,
): Promise<IsmartSubaccountMutationResponse> => {
  const { data } = await httpClient.post<ApiResponse<IsmartSubaccountMutationResponse>>('/me/ismart/subaccounts/revoke', payload);
  return data.data;
};

// 10. 取得目前會員 iSmart 大廈資料
export const fetchMemberIsmartBuildingInfo = async (buildingID?: string): Promise<IsmartBuildingInfoResponse> => {
  const params = buildingID ? { building_id: buildingID } : undefined;
  const { data } = await httpClient.get<ApiResponse<IsmartBuildingInfoResponse>>('/me/ismart/building-info', { params });
  return data.data;
};

// 11. 取得目前會員 iSmart 智能門禁資料
export const fetchMemberIsmartBuildingAccess = async (buildingID?: string): Promise<IsmartBuildingAccessResponse> => {
  const params = buildingID ? { building_id: buildingID } : undefined;
  const { data } = await httpClient.get<ApiResponse<IsmartBuildingAccessResponse>>('/me/ismart/building-access', {
    params,
    timeout: 20000,
  });
  return data.data;
};

// 12. 發送目前會員 iSmart 開門指令
export const openMemberIsmartDoor = async (payload: OpenIsmartDoorPayload): Promise<IsmartDoorOpenResponse> => {
  const { data } = await httpClient.post<ApiResponse<IsmartDoorOpenResponse>>('/me/ismart/building-access/open-door', payload);
  return data.data;
};

// 13. 生成目前會員 iSmart 門禁二維碼
export const generateMemberIsmartDoorQRCode = async (payload: GenerateIsmartQRCodePayload): Promise<IsmartQRCodeResponse> => {
  const { data } = await httpClient.post<ApiResponse<IsmartQRCodeResponse>>('/me/ismart/building-access/qrcode', payload);
  return data.data;
};

// 14. 取得目前會員 iCCTV 視像監控鏡頭
export const fetchMemberICCTVPublicCameras = async (buildingID?: string): Promise<ICCTVPublicCameraResponse> => {
  const params = buildingID ? { building_id: buildingID } : undefined;
  const { data } = await httpClient.get<ApiResponse<ICCTVPublicCameraResponse>>('/me/security/icctv/public-cameras', {
    params,
    timeout: 20000,
  });
  return data.data;
};

// 15. 取得目前會員 iSmart 大廈管理費表
export const fetchMemberIsmartManagementFees = async (
  buildingID?: string,
  dataStructure: IsmartReceivableDataStructure = 'table',
): Promise<IsmartReceivableResponse<IsmartManagementFeeTableRow>> => {
  const params = {
    data_structure: dataStructure,
    ...(buildingID ? { building_id: buildingID } : {}),
  };
  const { data } = await httpClient.get<ApiResponse<IsmartReceivableResponse<IsmartManagementFeeTableRow>>>(
    '/me/ismart/management-fees',
    { params },
  );
  return data.data;
};

// 16. 取得目前會員 iSmart 大廈其他費用列表
export const fetchMemberIsmartOtherFees = async (
  buildingID?: string,
  dataStructure: Extract<IsmartReceivableDataStructure, 'list' | 'block_dict'> = 'list',
): Promise<IsmartReceivableResponse<IsmartOtherFeeRow>> => {
  const params = {
    data_structure: dataStructure,
    ...(buildingID ? { building_id: buildingID } : {}),
  };
  const { data } = await httpClient.get<ApiResponse<IsmartReceivableResponse<IsmartOtherFeeRow>>>(
    '/me/ismart/other-fees',
    { params },
  );
  return data.data;
};

// 17. 取得目前會員 iSmart 大廈有效通告
export const fetchMemberIsmartBuildingNotices = async (buildingID?: string): Promise<IsmartBuildingNoticesResponse> => {
  const params = buildingID ? { building_id: buildingID } : undefined;
  const { data } = await httpClient.get<ApiResponse<IsmartBuildingNoticesResponse>>('/me/ismart/notices', { params });
  return data.data;
};
