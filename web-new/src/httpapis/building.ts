/*
 * 大廈與 POS 樓宇 API。
 * 1. 讀取 POS relay 樓宇清單。
 * 2. 依大廈讀取可選單位清單。
 * 3. 讀取目前會員已綁定的 iSmart 大廈資料。
 * 4. 讀取目前會員已綁定大廈的 iSmart 智能門禁資料。
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
  floorplan?: IsmartBuildingDocument[];
  auditreport?: IsmartBuildingDocument[];
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

// 6. 取得目前會員 iSmart 大廈資料
export const fetchMemberIsmartBuildingInfo = async (buildingID?: string): Promise<IsmartBuildingInfoResponse> => {
  const params = buildingID ? { building_id: buildingID } : undefined;
  const { data } = await httpClient.get<ApiResponse<IsmartBuildingInfoResponse>>('/me/ismart/building-info', { params });
  return data.data;
};

// 7. 取得目前會員 iSmart 智能門禁資料
export const fetchMemberIsmartBuildingAccess = async (buildingID?: string): Promise<IsmartBuildingAccessResponse> => {
  const params = buildingID ? { building_id: buildingID } : undefined;
  const { data } = await httpClient.get<ApiResponse<IsmartBuildingAccessResponse>>('/me/ismart/building-access', { params });
  return data.data;
};

// 8. 發送目前會員 iSmart 開門指令
export const openMemberIsmartDoor = async (payload: OpenIsmartDoorPayload): Promise<IsmartDoorOpenResponse> => {
  const { data } = await httpClient.post<ApiResponse<IsmartDoorOpenResponse>>('/me/ismart/building-access/open-door', payload);
  return data.data;
};

// 9. 生成目前會員 iSmart 門禁二維碼
export const generateMemberIsmartDoorQRCode = async (payload: GenerateIsmartQRCodePayload): Promise<IsmartQRCodeResponse> => {
  const { data } = await httpClient.post<ApiResponse<IsmartQRCodeResponse>>('/me/ismart/building-access/qrcode', payload);
  return data.data;
};
