/*
 * 前端格式化工具。
 * 1. 集中處理貨幣、日期與簡短時間文案。
 * 2. 避免頁面重複撰寫格式化邏輯。
 */

// 1. 格式化港幣金額
export const formatPrice = (value: number, locale = 'zh-HK') =>
  new Intl.NumberFormat(locale, {
    style: 'currency',
    currency: 'HKD',
    maximumFractionDigits: 0,
  }).format(value);

// 2. 格式化短日期
export const formatDate = (value: string, locale = 'zh-HK') =>
  new Intl.DateTimeFormat(locale, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  }).format(new Date(value));

// 3. 輸出相對時間提示
export const formatRelativeDay = (value: string, locale = 'zh-HK') => {
  const target = new Date(value).getTime();
  const diffDay = Math.round((target - Date.now()) / (1000 * 60 * 60 * 24));
  const formatter = new Intl.RelativeTimeFormat(locale, { numeric: 'auto' });

  return formatter.format(diffDay, 'day');
};
