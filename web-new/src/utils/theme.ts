/*
 * 主題設計 Token 定義。
 * 1. 維護色彩、圓角、陰影、間距、字體與邊框 token。
 * 2. 提供多套 AJO Living 前台主題的語意化配置。
 * 3. 將 token 寫入 CSS variables 供 Tailwind 與自定義元件共用。
 */

// 1. 定義支援的主題名稱
export type AppThemeName =
  | 'default'
  | 'html-fidelity'
  | 'copper-sun'
  | 'dark-neutral';

interface ThemeColorTokens {
  canvas: string;
  pageTint: string;
  surface: string;
  surfaceMuted: string;
  surfaceRaised: string;
  topbarSurface: string;
  toolbarSurface: string;
  fieldSurface: string;
  dropdownSurface: string;
  primary: string;
  primaryContrast: string;
  primarySoft: string;
  text: string;
  textMuted: string;
  border: string;
  borderStrong: string;
  success: string;
  warning: string;
  danger: string;
}

interface ThemeShadowTokens {
  soft: string;
  raised: string;
  floating: string;
}

interface ThemeFontTokens {
  sans: string;
  display: string;
}

interface ThemePreset {
  labelKey: string;
  color: ThemeColorTokens;
  shadow: ThemeShadowTokens;
  font: ThemeFontTokens;
}

interface ThemePaletteSeed {
  canvas: string;
  pageTint: string;
  surface: string;
  surfaceMuted: string;
  surfaceRaised: string;
  primary: string;
  primaryContrast: string;
  primarySoft: string;
  text: string;
  textMuted: string;
  border: string;
  borderStrong: string;
  success: string;
  warning: string;
  danger: string;
}

const SANS_FONT_STACK = '"DM Sans", "Avenir Next", "PingFang TC", "Noto Sans TC", sans-serif';
const WARM_SERIF_FONT_STACK = '"DM Serif Display", "Noto Serif TC", serif';
const BLACK_RGB = '0 0 0';

const SHARED_RADIUS_TOKENS = {
  sm: '0.125rem',
  md: '0.25rem',
  lg: '0.375rem',
  xl: '0.5rem',
  xxl: '0.5rem',
  pill: '999px',
} as const;

const SHARED_SPACING_TOKENS = {
  shell: 'clamp(1rem, 2vw, 1.5rem)',
  section: 'clamp(4rem, 8vw, 6rem)',
  stackSm: '0.75rem',
  stackMd: '1.25rem',
  stackLg: '2rem',
  card: '1.5rem',
} as const;

const SHARED_TYPE_SCALE = {
  bodySm: '0.875rem',
  body: '1rem',
  bodyLg: '1.125rem',
  titleSm: '1.5rem',
  title: '2rem',
  hero: '3.25rem',
} as const;

const SHARED_BORDER_SCALE = {
  thin: '1px',
  strong: '1.5px',
} as const;

type ThemeToneMode = 'light' | 'dark';

// 2.1 解析與混合 RGB token，讓層級色可由主題種子推導
const parseRgbValue = (value: string): [number, number, number] => {
  const [r, g, b] = value.split(' ').map((segment) => Number(segment));
  return [r, g, b];
};

const toRgbValue = ([r, g, b]: [number, number, number]) =>
  `${Math.round(r)} ${Math.round(g)} ${Math.round(b)}`;

const mixRgbValue = (from: string, to: string, ratio: number) => {
  const [fromR, fromG, fromB] = parseRgbValue(from);
  const [toR, toG, toB] = parseRgbValue(to);
  const clampedRatio = Math.min(Math.max(ratio, 0), 1);

  return toRgbValue([
    fromR * (1 - clampedRatio) + toR * clampedRatio,
    fromG * (1 - clampedRatio) + toG * clampedRatio,
    fromB * (1 - clampedRatio) + toB * clampedRatio,
  ]);
};

// 2.2 依主題深淺推導導航與互動表面層級色
const buildDerivedColorTokens = (seed: ThemePaletteSeed, mode: ThemeToneMode = 'light'): ThemeColorTokens => {
  const topbarSurface =
    mode === 'dark'
      ? mixRgbValue(seed.canvas, BLACK_RGB, 0.18)
      : mixRgbValue(seed.canvas, seed.pageTint, 0.58);
  const toolbarSurface =
    mode === 'dark'
      ? mixRgbValue(seed.canvas, BLACK_RGB, 0.1)
      : mixRgbValue(topbarSurface, seed.pageTint, 0.34);
  const fieldSurface =
    mode === 'dark'
      ? mixRgbValue(seed.surface, seed.canvas, 0.34)
      : mixRgbValue(seed.surface, seed.canvas, 0.22);
  const dropdownSurface =
    mode === 'dark'
      ? mixRgbValue(seed.surfaceMuted, seed.surface, 0.34)
      : mixRgbValue(seed.surfaceRaised, seed.surface, 0.38);

  return {
    ...seed,
    topbarSurface,
    toolbarSurface,
    fieldSurface,
    dropdownSurface,
  };
};

// 2. 定義主題配置
export const THEME_PRESETS: Record<AppThemeName, ThemePreset> = {
  default: {
    labelKey: 'common.theme.default',
    color: buildDerivedColorTokens({
      canvas: '255 255 255',
      pageTint: '244 244 244',
      surface: '255 255 255',
      surfaceMuted: '248 248 248',
      surfaceRaised: '255 255 255',
      primary: '240 90 0',
      primaryContrast: '255 255 255',
      primarySoft: '255 240 230',
      text: '26 26 26',
      textMuted: '102 102 102',
      border: '228 228 228',
      borderStrong: '192 70 0',
      success: '46 132 100',
      warning: '188 135 78',
      danger: '178 77 73',
    }),
    shadow: {
      soft: '0 3px 16px rgba(240, 90, 0, 0.12)',
      raised: '0 10px 28px rgba(26, 26, 26, 0.08)',
      floating: '0 18px 48px rgba(26, 26, 26, 0.14)',
    },
    font: {
      sans: SANS_FONT_STACK,
      display: WARM_SERIF_FONT_STACK,
    },
  },
  'html-fidelity': {
    labelKey: 'common.theme.htmlFidelity',
    color: buildDerivedColorTokens({
      canvas: '255 255 255',
      pageTint: '244 244 244',
      surface: '255 255 255',
      surfaceMuted: '244 244 244',
      surfaceRaised: '255 255 255',
      primary: '240 90 0',
      primaryContrast: '255 255 255',
      primarySoft: '255 240 230',
      text: '26 26 26',
      textMuted: '119 119 119',
      border: '228 228 228',
      borderStrong: '192 70 0',
      success: '46 132 100',
      warning: '188 135 78',
      danger: '178 77 73',
    }),
    shadow: {
      soft: '0 3px 14px rgba(240, 90, 0, 0.15)',
      raised: '0 3px 16px rgba(240, 90, 0, 0.15)',
      floating: '0 4px 18px rgba(240, 90, 0, 0.20)',
    },
    font: {
      sans: SANS_FONT_STACK,
      display: WARM_SERIF_FONT_STACK,
    },
  },
  'copper-sun': {
    labelKey: 'common.theme.copperSun',
    color: buildDerivedColorTokens({
      canvas: '255 250 245',
      pageTint: '255 240 230',
      surface: '255 255 255',
      surfaceMuted: '255 247 240',
      surfaceRaised: '255 252 248',
      primary: '204 76 0',
      primaryContrast: '255 255 255',
      primarySoft: '255 230 210',
      text: '55 38 28',
      textMuted: '120 92 75',
      border: '235 213 198',
      borderStrong: '204 76 0',
      success: '46 132 100',
      warning: '188 135 78',
      danger: '178 77 73',
    }),
    shadow: {
      soft: '0 3px 16px rgba(204, 76, 0, 0.12)',
      raised: '0 10px 28px rgba(55, 38, 28, 0.08)',
      floating: '0 18px 48px rgba(55, 38, 28, 0.14)',
    },
    font: {
      sans: SANS_FONT_STACK,
      display: WARM_SERIF_FONT_STACK,
    },
  },
  'dark-neutral': {
    labelKey: 'common.theme.darkNeutral',
    color: buildDerivedColorTokens({
      canvas: '0 0 0',
      pageTint: '18 18 18',
      surface: '10 10 10',
      surfaceMuted: '18 18 18',
      surfaceRaised: '28 28 28',
      primary: '214 172 103',
      primaryContrast: '10 10 10',
      primarySoft: '78 63 35',
      text: '242 245 248',
      textMuted: '154 154 154',
      border: '52 52 52',
      borderStrong: '208 165 96',
      success: '73 190 145',
      warning: '221 176 100',
      danger: '235 113 121',
    }, 'dark'),
    shadow: {
      soft: '0 18px 40px rgba(0, 0, 0, 0.24)',
      raised: '0 24px 64px rgba(0, 0, 0, 0.30)',
      floating: '0 32px 84px rgba(0, 0, 0, 0.40)',
    },
    font: {
      sans: SANS_FONT_STACK,
      display: WARM_SERIF_FONT_STACK,
    },
  },
};

const buildThemeCssVariables = (preset: ThemePreset): Record<string, string> => ({
  '--color-canvas': preset.color.canvas,
  '--color-page-tint': preset.color.pageTint,
  '--color-surface': preset.color.surface,
  '--color-surface-muted': preset.color.surfaceMuted,
  '--color-surface-raised': preset.color.surfaceRaised,
  '--color-topbar-surface': preset.color.topbarSurface,
  '--color-toolbar-surface': preset.color.toolbarSurface,
  '--color-field-surface': preset.color.fieldSurface,
  '--color-dropdown-surface': preset.color.dropdownSurface,
  '--color-primary': preset.color.primary,
  '--color-primary-contrast': preset.color.primaryContrast,
  '--color-primary-soft': preset.color.primarySoft,
  '--color-text': preset.color.text,
  '--color-text-muted': preset.color.textMuted,
  '--color-border': preset.color.border,
  '--color-border-strong': preset.color.borderStrong,
  '--color-success': preset.color.success,
  '--color-warning': preset.color.warning,
  '--color-danger': preset.color.danger,
  '--radius-sm': SHARED_RADIUS_TOKENS.sm,
  '--radius-md': SHARED_RADIUS_TOKENS.md,
  '--radius-lg': SHARED_RADIUS_TOKENS.lg,
  '--radius-xl': SHARED_RADIUS_TOKENS.xl,
  '--radius-2xl': SHARED_RADIUS_TOKENS.xxl,
  '--radius-pill': SHARED_RADIUS_TOKENS.pill,
  '--shadow-soft': preset.shadow.soft,
  '--shadow-raised': preset.shadow.raised,
  '--shadow-floating': preset.shadow.floating,
  '--space-shell': SHARED_SPACING_TOKENS.shell,
  '--space-section': SHARED_SPACING_TOKENS.section,
  '--space-stack-sm': SHARED_SPACING_TOKENS.stackSm,
  '--space-stack-md': SHARED_SPACING_TOKENS.stackMd,
  '--space-stack-lg': SHARED_SPACING_TOKENS.stackLg,
  '--space-card': SHARED_SPACING_TOKENS.card,
  '--font-sans': preset.font.sans,
  '--font-display': preset.font.display,
  '--font-size-body-sm': SHARED_TYPE_SCALE.bodySm,
  '--font-size-body': SHARED_TYPE_SCALE.body,
  '--font-size-body-lg': SHARED_TYPE_SCALE.bodyLg,
  '--font-size-title-sm': SHARED_TYPE_SCALE.titleSm,
  '--font-size-title': SHARED_TYPE_SCALE.title,
  '--font-size-hero': SHARED_TYPE_SCALE.hero,
  '--border-width-thin': SHARED_BORDER_SCALE.thin,
  '--border-width-strong': SHARED_BORDER_SCALE.strong,
  '--border-subtle': `${SHARED_BORDER_SCALE.thin} solid rgb(${preset.color.border} / 0.82)`,
  '--border-strong': `${SHARED_BORDER_SCALE.strong} solid rgb(${preset.color.borderStrong} / 0.22)`,
});

// 3. 定義主題切換選項
export const THEME_OPTIONS = (
  Object.entries(THEME_PRESETS) as [AppThemeName, ThemePreset][]
).map(([value, preset]) => ({
  labelKey: preset.labelKey,
  value,
}));

// 4. 定義預設主題名稱
export const DEFAULT_THEME_NAME: AppThemeName = 'default';

const THEME_ROOT_ATTRIBUTE = 'data-theme';

// 5. 將主題 token 寫入根節點
const applyCssVariables = (themeName: AppThemeName) => {
  const root = document.documentElement;
  const cssVariables = buildThemeCssVariables(THEME_PRESETS[themeName]);

  Object.entries(cssVariables).forEach(([key, value]) => {
    root.style.setProperty(key, value);
  });
};

// 6. 套用主題到全域文件
export const applyTheme = (themeName: AppThemeName) => {
  if (typeof document === 'undefined') {
    return;
  }

  document.documentElement.setAttribute(THEME_ROOT_ATTRIBUTE, themeName);
  applyCssVariables(themeName);
};
