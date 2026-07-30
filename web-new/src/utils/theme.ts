/*
 * 全站視覺基線。
 * 1. 固定白色頁面底色與鮮橙主操作色。
 * 2. 將舊版多皮膚偏好統一回落至唯一正式皮膚。
 */

// 1. 保留主題型別，前端只使用唯一正式皮膚
export type AppThemeName = 'default';

// 2. 定義白色與鮮橙皮膚的語意化 CSS token
const DEFAULT_THEME_VARIABLES: Record<string, string> = {
  '--color-canvas': '255 255 255',
  '--color-page-tint': '255 255 255',
  '--color-surface': '255 255 255',
  '--color-surface-muted': '247 247 247',
  '--color-surface-raised': '255 255 255',
  '--color-topbar-surface': '255 255 255',
  '--color-toolbar-surface': '255 255 255',
  '--color-field-surface': '255 255 255',
  '--color-dropdown-surface': '255 255 255',
  '--color-primary': '255 106 0',
  '--color-primary-contrast': '255 255 255',
  '--color-primary-soft': '255 241 232',
  '--color-text': '26 26 26',
  '--color-text-muted': '102 102 102',
  '--color-border': '228 228 228',
  '--color-border-strong': '217 87 0',
  '--color-success': '26 122 58',
  '--color-warning': '224 123 0',
  '--color-danger': '192 57 43',
  '--color-ink-2': '68 68 68',
  '--color-ink-3': '119 119 119',
  '--color-ink-4': '170 170 170',
  '--color-surface-2': '247 247 247',
  '--color-surface-3': '239 239 239',
  '--color-border-2': '204 204 204',
  '--color-brand-mid': '255 172 112',
  '--color-brand-dark': '217 87 0',
  '--color-success-bg': '230 244 236',
  '--color-warning-bg': '255 243 220',
  '--color-danger-bg': '253 236 234',
  '--brand': 'rgb(255 106 0)',
  '--brand-dark': 'rgb(217 87 0)',
  '--brand-light': 'rgb(255 241 232)',
  '--brand-mid': 'rgb(255 172 112)',
  '--ink': 'rgb(26 26 26)',
  '--ink-2': 'rgb(68 68 68)',
  '--ink-3': 'rgb(119 119 119)',
  '--ink-4': 'rgb(170 170 170)',
  '--sur': 'rgb(255 255 255)',
  '--sur-2': 'rgb(247 247 247)',
  '--sur-3': 'rgb(239 239 239)',
  '--bdr': 'rgb(228 228 228)',
  '--bdr-2': 'rgb(204 204 204)',
  '--success': 'rgb(26 122 58)',
  '--success-bg': 'rgb(230 244 236)',
  '--error': 'rgb(192 57 43)',
  '--error-bg': 'rgb(253 236 234)',
  '--warning': 'rgb(224 123 0)',
  '--warning-bg': 'rgb(255 243 220)',
  '--black': 'rgb(26 26 26)',
  '--white': 'rgb(255 255 255)',
  '--g1': 'rgb(247 247 247)',
  '--g2': 'rgb(228 228 228)',
  '--g3': 'rgb(170 170 170)',
  '--g4': 'rgb(119 119 119)',
  '--g5': 'rgb(68 68 68)',
  '--accent': 'rgb(255 106 0)',
  '--accent-light': 'rgb(255 241 232)',
  '--accent-dark': 'rgb(217 87 0)',
};

// 3. 定義唯一正式皮膚名稱
export const DEFAULT_THEME_NAME: AppThemeName = 'default';

// 4. 套用白色與鮮橙全站視覺基線
export const applyTheme = (): void => {
  if (typeof document === 'undefined') {
    return;
  }

  const root = document.documentElement;
  root.setAttribute('data-theme', DEFAULT_THEME_NAME);

  Object.entries(DEFAULT_THEME_VARIABLES).forEach(([key, value]) => {
    root.style.setProperty(key, value);
  });
};
