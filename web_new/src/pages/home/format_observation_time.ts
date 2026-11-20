/*
 * 首頁觀測時間格式。
 * 1. 所有即時資訊統一按香港本地時間顯示。
 */

// 1. 無效或缺少的時間不顯示，也不以頁面載入時間代替。
export const formatObservationTime = (value: string | undefined, locale: string): string => {
  if (!value) return '';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '';
  return new Intl.DateTimeFormat(locale, {
    timeZone: 'Asia/Hong_Kong',
    hour: '2-digit',
    minute: '2-digit',
    hourCycle: 'h23',
  }).format(date);
};
