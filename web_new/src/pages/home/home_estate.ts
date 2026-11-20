/*
 * 首頁屋苑選擇。
 * 1. 列出三個公開參考屋苑。
 * 2. 沒有明確選擇時優先使用會員已綁定的大廈。
 */

export const HOME_ESTATES = ['seaview', 'summit', 'fleur'] as const;
export type HomeEstate = (typeof HOME_ESTATES)[number];

// 1. 只接受已知參考屋苑，不從查詢字串讀取會員大廈。
export const resolveHomeEstate = (value: unknown, buildingID?: string): HomeEstate | undefined =>
  HOME_ESTATES.find((estate) => estate === value) ?? (buildingID ? undefined : HOME_ESTATES[0]);
