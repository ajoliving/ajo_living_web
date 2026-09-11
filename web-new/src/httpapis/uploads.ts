/*
 * 上傳 API。
 * 1. 對齊後端 `/uploads/presign` 與 `/uploads/complete` 契約。
 * 2. 提供媒體資產列表與刪除能力。
 */
import httpClient from '@/httpapis';
import type { ApiResponse, PaginatedResult } from '@/model/api';
import type {
  MediaAssetResponse,
  PresignUploadPayload,
  PresignUploadResult,
} from '@/model/marketplace';

interface ListMediaAssetsParams {
  page?: number;
  page_size?: number;
}

interface CompleteUploadPayload {
  object_key: string;
  upload_token: string;
  mime_type: string;
  file_size: number;
  width?: number;
  height?: number;
  checksum_sha256?: string;
}

export interface MultipartInitiateResult {
  upload_id: string;
  object_key: string;
  upload_token: string;
  part_size: number;
  part_count: number;
  expires_at: string;
}

export interface MultipartPartResult {
  upload_url: string;
  object_key: string;
  upload_id: string;
  part_number: number;
  headers: Record<string, string>;
}

export interface MultipartCompletePayload {
  object_key: string;
  upload_id: string;
  upload_token: string;
  mime_type: string;
  file_size: number;
  parts: Array<{ part_number: number; etag: string }>;
  width?: number;
  height?: number;
  checksum_sha256?: string;
}

export interface MultipartPartPayload {
  object_key: string;
  upload_id: string;
  upload_token: string;
  mime_type: string;
  file_size: number;
  part_number: number;
}

export interface MultipartAbortPayload {
  object_key: string;
  upload_id: string;
  upload_token: string;
  mime_type: string;
  file_size: number;
}

// 1. 取得上傳預簽名資訊
export const createUploadPresign = (payload: PresignUploadPayload) =>
  httpClient.post<ApiResponse<PresignUploadResult>>('/oss/presign', payload);

// 2. 回寫上傳完成結果
export const completeUpload = (payload: CompleteUploadPayload) =>
  httpClient.post<ApiResponse<MediaAssetResponse>>('/oss/complete', payload);

// 3.1 建立可續傳分片上傳會話
export const initiateMultipartUpload = (payload: PresignUploadPayload) =>
  httpClient.post<ApiResponse<MultipartInitiateResult>>('/oss/multipart/initiate', payload);

// 3.2 取得單個分片預簽名地址
export const presignMultipartPart = (payload: MultipartPartPayload) =>
  httpClient.post<ApiResponse<MultipartPartResult>>('/oss/multipart/part', payload);

// 3.3 完成分片上傳並登記媒體
export const completeMultipartUpload = (payload: MultipartCompletePayload) =>
  httpClient.post<ApiResponse<MediaAssetResponse>>('/oss/multipart/complete', payload);

// 3.4 查詢已完成分片，支援斷線續傳
export const listMultipartParts = (payload: MultipartAbortPayload) =>
  httpClient.post<ApiResponse<{ parts: Array<{ part_number: number; etag: string }> }>>('/oss/multipart/parts', payload);

// 3.5 取消未完成分片上傳
export const abortMultipartUpload = (payload: MultipartAbortPayload) =>
  httpClient.post<ApiResponse<{ aborted: boolean }>>('/oss/multipart/abort', payload);

// 3. 取得媒體資產列表
export const fetchMediaAssets = (params: ListMediaAssetsParams = {}) =>
  httpClient.get<ApiResponse<PaginatedResult<MediaAssetResponse>>>('/oss/assets', { params });

// 4. 刪除媒體資產
export const deleteMediaAsset = (mediaAssetId: string) =>
  httpClient.delete<ApiResponse<{ media_asset_id: string; deleted: boolean }>>(
    `/oss/assets/${mediaAssetId}`,
  );
