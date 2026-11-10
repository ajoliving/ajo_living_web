/*
 * QRCode 套件型別宣告。
 * 1. 補足本地缺少 @types/qrcode 時的 toDataURL 型別。
 * 2. 供二維碼圖片元件在嚴格模式下使用。
 */
declare module 'qrcode' {
  export interface QRCodeToDataURLOptions {
    width?: number;
    margin?: number;
    errorCorrectionLevel?: 'L' | 'M' | 'Q' | 'H';
  }

  // 1. 將文字內容轉為 data URL
  export function toDataURL(text: string, options?: QRCodeToDataURLOptions): Promise<string>;
}
