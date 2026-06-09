/*
 * 上傳 API。
 * 1. 對齊後端 `/uploads/presign` 與 `/uploads/complete` 契約。
 * 2. 提供媒體資產列表與刪除能力。
 */
import httpClient from '@/shared/utils/http';
import type { ApiResponse, PaginatedResult } from '@/shared/utils/http/model';
import type {
  MediaAssetResponse,
  PresignUploadPayload,
  PresignUploadResult,
} from '@/domains/marketplace/model';

interface ListMediaAssetsParams {
  page?: number;
  page_size?: number;
}

interface CompleteUploadPayload {
  object_key: string;
  mime_type: string;
  file_size: number;
  width?: number;
  height?: number;
  checksum_sha256?: string;
}

// 1. 取得上傳預簽名資訊
export const createUploadPresign = (payload: PresignUploadPayload) =>
  httpClient.post<ApiResponse<PresignUploadResult>>('/oss/presign', payload);

// 2. 回寫上傳完成結果
export const completeUpload = (payload: CompleteUploadPayload) =>
  httpClient.post<ApiResponse<MediaAssetResponse>>('/oss/complete', payload);

// 3. 取得媒體資產列表
export const fetchMediaAssets = (params: ListMediaAssetsParams = {}) =>
  httpClient.get<ApiResponse<PaginatedResult<MediaAssetResponse>>>('/oss/assets', { params });

// 4. 刪除媒體資產
export const deleteMediaAsset = (mediaAssetId: string) =>
  httpClient.delete<ApiResponse<{ media_asset_id: string; deleted: boolean }>>(
    `/oss/assets/${mediaAssetId}`,
  );
