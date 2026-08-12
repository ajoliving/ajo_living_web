/*
 * 全站主題色配置。
 * 1. 提供四十組細分亮橙方案與二十組正式替代主題。
 * 2. 以統一語意化 token 同步既有 Tailwind 與 scoped 元件配色。
 * 3. 保留白色內容表面與固定狀態色，避免主題影響業務狀態辨識。
 */

export type AppThemeName =
  | 'orange-01'
  | 'orange-02'
  | 'orange-03'
  | 'orange-04'
  | 'orange-05'
  | 'orange-06'
  | 'orange-07'
  | 'orange-08'
  | 'orange-09'
  | 'orange-10'
  | 'orange-11'
  | 'orange-12'
  | 'orange-13'
  | 'orange-14'
  | 'orange-15'
  | 'orange-16'
  | 'orange-17'
  | 'orange-18'
  | 'orange-19'
  | 'orange-20'
  | 'orange-21'
  | 'orange-22'
  | 'orange-23'
  | 'orange-24'
  | 'orange-25'
  | 'orange-26'
  | 'orange-27'
  | 'orange-28'
  | 'orange-29'
  | 'orange-30'
  | 'orange-31'
  | 'orange-32'
  | 'orange-33'
  | 'orange-34'
  | 'orange-35'
  | 'orange-36'
  | 'orange-37'
  | 'orange-38'
  | 'orange-39'
  | 'orange-40'
  | 'crimson'
  | 'rose'
  | 'berry'
  | 'plum'
  | 'violet'
  | 'indigo'
  | 'cobalt'
  | 'azure'
  | 'cyan'
  | 'teal'
  | 'jade'
  | 'emerald'
  | 'forest'
  | 'lime'
  | 'gold'
  | 'amber'
  | 'coral'
  | 'terracotta'
  | 'slate'
  | 'graphite';

export interface ThemeOption {
  value: AppThemeName;
  labelKey: string;
  primary: string;
  primaryContrast: 'light' | 'dark';
  group: 'orange' | 'other';
  orangeTone?: 'flame' | 'classic' | 'solar' | 'peach';
}

// 1. 建立可供甲方直接選擇的正式主題色清單
export const THEME_OPTIONS: ThemeOption[] = [
  { value: 'orange-01', labelKey: 'common.theme.orange01', primary: '#FF4D00', primaryContrast: 'dark', group: 'orange', orangeTone: 'flame' },
  { value: 'orange-02', labelKey: 'common.theme.orange02', primary: '#FF5500', primaryContrast: 'dark', group: 'orange', orangeTone: 'flame' },
  { value: 'orange-03', labelKey: 'common.theme.orange03', primary: '#FF5D00', primaryContrast: 'dark', group: 'orange', orangeTone: 'flame' },
  { value: 'orange-04', labelKey: 'common.theme.orange04', primary: '#FF6500', primaryContrast: 'dark', group: 'orange', orangeTone: 'flame' },
  { value: 'orange-05', labelKey: 'common.theme.orange05', primary: '#FF5A00', primaryContrast: 'dark', group: 'orange', orangeTone: 'flame' },
  { value: 'orange-06', labelKey: 'common.theme.orange06', primary: '#FF7000', primaryContrast: 'dark', group: 'orange', orangeTone: 'classic' },
  { value: 'orange-07', labelKey: 'common.theme.orange07', primary: '#FF7800', primaryContrast: 'dark', group: 'orange', orangeTone: 'classic' },
  { value: 'orange-08', labelKey: 'common.theme.orange08', primary: '#FF8000', primaryContrast: 'dark', group: 'orange', orangeTone: 'classic' },
  { value: 'orange-09', labelKey: 'common.theme.orange09', primary: '#FF8800', primaryContrast: 'dark', group: 'orange', orangeTone: 'classic' },
  { value: 'orange-10', labelKey: 'common.theme.orange10', primary: '#FF9000', primaryContrast: 'dark', group: 'orange', orangeTone: 'classic' },
  { value: 'orange-11', labelKey: 'common.theme.orange11', primary: '#FF9800', primaryContrast: 'dark', group: 'orange', orangeTone: 'solar' },
  { value: 'orange-12', labelKey: 'common.theme.orange12', primary: '#FFA000', primaryContrast: 'dark', group: 'orange', orangeTone: 'solar' },
  { value: 'orange-13', labelKey: 'common.theme.orange13', primary: '#FFA800', primaryContrast: 'dark', group: 'orange', orangeTone: 'solar' },
  { value: 'orange-14', labelKey: 'common.theme.orange14', primary: '#FFB000', primaryContrast: 'dark', group: 'orange', orangeTone: 'solar' },
  { value: 'orange-15', labelKey: 'common.theme.orange15', primary: '#FFB800', primaryContrast: 'dark', group: 'orange', orangeTone: 'solar' },
  { value: 'orange-16', labelKey: 'common.theme.orange16', primary: '#FF6F1A', primaryContrast: 'dark', group: 'orange', orangeTone: 'peach' },
  { value: 'orange-17', labelKey: 'common.theme.orange17', primary: '#FF7924', primaryContrast: 'dark', group: 'orange', orangeTone: 'peach' },
  { value: 'orange-18', labelKey: 'common.theme.orange18', primary: '#FF832E', primaryContrast: 'dark', group: 'orange', orangeTone: 'peach' },
  { value: 'orange-19', labelKey: 'common.theme.orange19', primary: '#FF8D38', primaryContrast: 'dark', group: 'orange', orangeTone: 'peach' },
  { value: 'orange-20', labelKey: 'common.theme.orange20', primary: '#FF9742', primaryContrast: 'dark', group: 'orange', orangeTone: 'peach' },
  { value: 'orange-21', labelKey: 'common.theme.orange21', primary: '#FF4700', primaryContrast: 'dark', group: 'orange', orangeTone: 'flame' },
  { value: 'orange-22', labelKey: 'common.theme.orange22', primary: '#FF4F00', primaryContrast: 'dark', group: 'orange', orangeTone: 'flame' },
  { value: 'orange-23', labelKey: 'common.theme.orange23', primary: '#FF5700', primaryContrast: 'dark', group: 'orange', orangeTone: 'flame' },
  { value: 'orange-24', labelKey: 'common.theme.orange24', primary: '#FF5F00', primaryContrast: 'dark', group: 'orange', orangeTone: 'flame' },
  { value: 'orange-25', labelKey: 'common.theme.orange25', primary: '#FF6700', primaryContrast: 'dark', group: 'orange', orangeTone: 'flame' },
  { value: 'orange-26', labelKey: 'common.theme.orange26', primary: '#FF6C00', primaryContrast: 'dark', group: 'orange', orangeTone: 'classic' },
  { value: 'orange-27', labelKey: 'common.theme.orange27', primary: '#FF7400', primaryContrast: 'dark', group: 'orange', orangeTone: 'classic' },
  { value: 'orange-28', labelKey: 'common.theme.orange28', primary: '#FF7C00', primaryContrast: 'dark', group: 'orange', orangeTone: 'classic' },
  { value: 'orange-29', labelKey: 'common.theme.orange29', primary: '#FF8400', primaryContrast: 'dark', group: 'orange', orangeTone: 'classic' },
  { value: 'orange-30', labelKey: 'common.theme.orange30', primary: '#FF8C00', primaryContrast: 'dark', group: 'orange', orangeTone: 'classic' },
  { value: 'orange-31', labelKey: 'common.theme.orange31', primary: '#FF9400', primaryContrast: 'dark', group: 'orange', orangeTone: 'solar' },
  { value: 'orange-32', labelKey: 'common.theme.orange32', primary: '#FF9C00', primaryContrast: 'dark', group: 'orange', orangeTone: 'solar' },
  { value: 'orange-33', labelKey: 'common.theme.orange33', primary: '#FFA400', primaryContrast: 'dark', group: 'orange', orangeTone: 'solar' },
  { value: 'orange-34', labelKey: 'common.theme.orange34', primary: '#FFAC00', primaryContrast: 'dark', group: 'orange', orangeTone: 'solar' },
  { value: 'orange-35', labelKey: 'common.theme.orange35', primary: '#FFB400', primaryContrast: 'dark', group: 'orange', orangeTone: 'solar' },
  { value: 'orange-36', labelKey: 'common.theme.orange36', primary: '#FF6912', primaryContrast: 'dark', group: 'orange', orangeTone: 'peach' },
  { value: 'orange-37', labelKey: 'common.theme.orange37', primary: '#FF731C', primaryContrast: 'dark', group: 'orange', orangeTone: 'peach' },
  { value: 'orange-38', labelKey: 'common.theme.orange38', primary: '#FF7D26', primaryContrast: 'dark', group: 'orange', orangeTone: 'peach' },
  { value: 'orange-39', labelKey: 'common.theme.orange39', primary: '#FF872F', primaryContrast: 'dark', group: 'orange', orangeTone: 'peach' },
  { value: 'orange-40', labelKey: 'common.theme.orange40', primary: '#FF9139', primaryContrast: 'dark', group: 'orange', orangeTone: 'peach' },
  { value: 'crimson', labelKey: 'common.theme.crimson', primary: '#C72C41', primaryContrast: 'light', group: 'other' },
  { value: 'rose', labelKey: 'common.theme.rose', primary: '#D93670', primaryContrast: 'dark', group: 'other' },
  { value: 'berry', labelKey: 'common.theme.berry', primary: '#A82B5E', primaryContrast: 'light', group: 'other' },
  { value: 'plum', labelKey: 'common.theme.plum', primary: '#7F3568', primaryContrast: 'light', group: 'other' },
  { value: 'violet', labelKey: 'common.theme.violet', primary: '#7046C5', primaryContrast: 'light', group: 'other' },
  { value: 'indigo', labelKey: 'common.theme.indigo', primary: '#4D58C9', primaryContrast: 'light', group: 'other' },
  { value: 'cobalt', labelKey: 'common.theme.cobalt', primary: '#2856BE', primaryContrast: 'light', group: 'other' },
  { value: 'azure', labelKey: 'common.theme.azure', primary: '#087BB5', primaryContrast: 'light', group: 'other' },
  { value: 'cyan', labelKey: 'common.theme.cyan', primary: '#087F8C', primaryContrast: 'light', group: 'other' },
  { value: 'teal', labelKey: 'common.theme.teal', primary: '#08796B', primaryContrast: 'light', group: 'other' },
  { value: 'jade', labelKey: 'common.theme.jade', primary: '#0D8B64', primaryContrast: 'dark', group: 'other' },
  { value: 'emerald', labelKey: 'common.theme.emerald', primary: '#17824A', primaryContrast: 'light', group: 'other' },
  { value: 'forest', labelKey: 'common.theme.forest', primary: '#35703C', primaryContrast: 'light', group: 'other' },
  { value: 'lime', labelKey: 'common.theme.lime', primary: '#607D1C', primaryContrast: 'light', group: 'other' },
  { value: 'gold', labelKey: 'common.theme.gold', primary: '#A86900', primaryContrast: 'dark', group: 'other' },
  { value: 'amber', labelKey: 'common.theme.amber', primary: '#C57400', primaryContrast: 'dark', group: 'other' },
  { value: 'coral', labelKey: 'common.theme.coral', primary: '#D94F45', primaryContrast: 'dark', group: 'other' },
  { value: 'terracotta', labelKey: 'common.theme.terracotta', primary: '#A94B32', primaryContrast: 'light', group: 'other' },
  { value: 'slate', labelKey: 'common.theme.slate', primary: '#466071', primaryContrast: 'light', group: 'other' },
  { value: 'graphite', labelKey: 'common.theme.graphite', primary: '#3E4651', primaryContrast: 'light', group: 'other' },
];

export const DEFAULT_THEME_NAME: AppThemeName = 'orange-05';

const BASE_THEME_VARIABLES: Record<string, string> = {
  '--color-canvas': '255 255 255',
  '--color-page-tint': '247 248 250',
  '--color-surface': '255 255 255',
  '--color-surface-muted': '247 248 250',
  '--color-surface-raised': '255 255 255',
  '--color-topbar-surface': '255 255 255',
  '--color-toolbar-surface': '247 248 250',
  '--color-field-surface': '255 255 255',
  '--color-dropdown-surface': '255 255 255',
  '--color-text': '26 26 26',
  '--color-text-muted': '102 102 102',
  '--color-border': '223 227 232',
  '--color-success': '26 122 58',
  '--color-warning': '224 123 0',
  '--color-danger': '192 57 43',
  '--color-ink-2': '68 68 68',
  '--color-ink-3': '119 119 119',
  '--color-ink-4': '170 170 170',
  '--color-surface-2': '247 248 250',
  '--color-surface-3': '238 241 244',
  '--color-border-2': '200 205 211',
  '--color-success-bg': '230 244 236',
  '--color-warning-bg': '255 243 220',
  '--color-danger-bg': '253 236 234',
  '--ink': 'rgb(26 26 26)',
  '--ink-2': 'rgb(68 68 68)',
  '--ink-3': 'rgb(119 119 119)',
  '--ink-4': 'rgb(170 170 170)',
  '--sur': 'rgb(255 255 255)',
  '--sur-2': 'rgb(247 248 250)',
  '--sur-3': 'rgb(238 241 244)',
  '--bdr': 'rgb(223 227 232)',
  '--bdr-2': 'rgb(200 205 211)',
  '--success': 'rgb(26 122 58)',
  '--success-bg': 'rgb(230 244 236)',
  '--error': 'rgb(192 57 43)',
  '--error-bg': 'rgb(253 236 234)',
  '--warning': 'rgb(224 123 0)',
  '--warning-bg': 'rgb(255 243 220)',
  '--black': 'rgb(26 26 26)',
  '--white': 'rgb(255 255 255)',
  '--g1': 'rgb(247 248 250)',
  '--g2': 'rgb(223 227 232)',
  '--g3': 'rgb(170 170 170)',
  '--g4': 'rgb(119 119 119)',
  '--g5': 'rgb(68 68 68)',
};

// 2. 將 hex 色值轉為既有 token 使用的 RGB 字串
const hexToRgb = (hex: string): [number, number, number] => [
  Number.parseInt(hex.slice(1, 3), 16),
  Number.parseInt(hex.slice(3, 5), 16),
  Number.parseInt(hex.slice(5, 7), 16),
];

// 3. 混合主色與白色，產出互動淺色表面
const mixWithWhite = (rgb: [number, number, number], amount: number): string =>
  rgb.map((value) => Math.round(value + (255 - value) * amount)).join(' ');

// 4. 取得主題定義，未知舊值安全回落至預設方案
export const getThemeOption = (themeName: string | null | undefined): ThemeOption =>
  THEME_OPTIONS.find((option) => option.value === themeName) ??
  THEME_OPTIONS.find((option) => option.value === DEFAULT_THEME_NAME)!;

// 5. 套用選定方案到全站語意化與既有相容 token
export const applyTheme = (themeName: AppThemeName = DEFAULT_THEME_NAME): void => {
  if (typeof document === 'undefined') {
    return;
  }

  const theme = getThemeOption(themeName);
  const primaryRgb = hexToRgb(theme.primary);
  const softRgb = mixWithWhite(primaryRgb, 0.88);
  const midRgb = mixWithWhite(primaryRgb, 0.44);
  const darkRgb = primaryRgb.map((value) => Math.round(value * 0.78)).join(' ');
  const primary = primaryRgb.join(' ');
  const contrast = theme.primaryContrast === 'light' ? '255 255 255' : '26 26 26';
  const root = document.documentElement;

  root.setAttribute('data-theme', theme.value);
  Object.entries(BASE_THEME_VARIABLES).forEach(([key, value]) => root.style.setProperty(key, value));
  root.style.setProperty('--color-primary', primary);
  root.style.setProperty('--color-primary-contrast', contrast);
  root.style.setProperty('--color-primary-soft', softRgb);
  root.style.setProperty('--color-brand-mid', midRgb);
  root.style.setProperty('--color-brand-dark', darkRgb);
  root.style.setProperty('--color-border-strong', darkRgb);
  root.style.setProperty('--brand', `rgb(${primary})`);
  root.style.setProperty('--brand-dark', `rgb(${darkRgb})`);
  root.style.setProperty('--brand-light', `rgb(${softRgb})`);
  root.style.setProperty('--brand-mid', `rgb(${midRgb})`);
  root.style.setProperty('--accent', `rgb(${primary})`);
  root.style.setProperty('--accent-light', `rgb(${softRgb})`);
  root.style.setProperty('--accent-dark', `rgb(${darkRgb})`);
  root.style.setProperty('--shadow-brand', `0 8px 24px rgb(${primary} / 0.24)`);
};
