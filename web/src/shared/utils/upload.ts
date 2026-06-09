/*
 * OSS 上傳工具。
 * 1. 建立瀏覽器 PUT 直傳需要的 headers。
 * 2. 避免重複傳入瀏覽器禁止設定的 header。
 */
const blockedUploadHeaders = new Set(['host', 'content-length']);

// 1. 建立 OSS 直傳 headers
export const buildUploadHeaders = (headers: Record<string, string>, mimeType: string): Headers => {
  const result = new Headers();

  Object.entries(headers).forEach(([key, value]) => {
    if (!blockedUploadHeaders.has(key.toLowerCase())) {
      result.set(key, value);
    }
  });

  if (!result.has('Content-Type')) {
    result.set('Content-Type', mimeType);
  }

  return result;
};
