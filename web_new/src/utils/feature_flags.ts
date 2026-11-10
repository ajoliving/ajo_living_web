/*
 * 前端建置期功能開關。
 * 1. 以 Vite 環境變數控制只在特定環境開放的功能入口。
 * 2. 未配置時採用預設值，避免本地開發與測試環境意外關閉功能。
 */

// 1. 判斷主題色選擇器是否開放（生產建置以 VITE_ENABLE_THEME_PICKER=0 關閉）
export const isThemePickerEnabled = (): boolean => {
  const raw = String(import.meta.env.VITE_ENABLE_THEME_PICKER ?? '').trim().toLowerCase();

  if (raw === '') {
    return true;
  }

  return !['0', 'false', 'off', 'no'].includes(raw);
};
