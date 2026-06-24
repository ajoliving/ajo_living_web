/*
 * Tailwind CSS 設定。
 * 1. 對接語意化色彩與版面 token。
 * 2. 提供字體、圓角、間距、陰影與邊框的主題化映射。
 */

/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,ts}'],
  theme: {
    extend: {
      colors: {
        canvas: 'rgb(var(--color-canvas) / <alpha-value>)',
        'page-tint': 'rgb(var(--color-page-tint) / <alpha-value>)',
        surface: 'rgb(var(--color-surface) / <alpha-value>)',
        'surface-muted': 'rgb(var(--color-surface-muted) / <alpha-value>)',
        'surface-raised': 'rgb(var(--color-surface-raised) / <alpha-value>)',
        'surface-2': 'rgb(var(--color-surface-2) / <alpha-value>)',
        'surface-3': 'rgb(var(--color-surface-3) / <alpha-value>)',
        'topbar-surface': 'rgb(var(--color-topbar-surface) / <alpha-value>)',
        'toolbar-surface': 'rgb(var(--color-toolbar-surface) / <alpha-value>)',
        'field-surface': 'rgb(var(--color-field-surface) / <alpha-value>)',
        'dropdown-surface': 'rgb(var(--color-dropdown-surface) / <alpha-value>)',
        primary: 'rgb(var(--color-primary) / <alpha-value>)',
        'primary-contrast': 'rgb(var(--color-primary-contrast) / <alpha-value>)',
        'primary-soft': 'rgb(var(--color-primary-soft) / <alpha-value>)',
        'brand-mid': 'rgb(var(--color-brand-mid) / <alpha-value>)',
        'brand-dark': 'rgb(var(--color-brand-dark) / <alpha-value>)',
        text: 'rgb(var(--color-text) / <alpha-value>)',
        'text-muted': 'rgb(var(--color-text-muted) / <alpha-value>)',
        'ink-2': 'rgb(var(--color-ink-2) / <alpha-value>)',
        'ink-3': 'rgb(var(--color-ink-3) / <alpha-value>)',
        'ink-4': 'rgb(var(--color-ink-4) / <alpha-value>)',
        border: 'rgb(var(--color-border) / <alpha-value>)',
        'border-strong': 'rgb(var(--color-border-strong) / <alpha-value>)',
        'border-2': 'rgb(var(--color-border-2) / <alpha-value>)',
        success: 'rgb(var(--color-success) / <alpha-value>)',
        'success-bg': 'rgb(var(--color-success-bg) / <alpha-value>)',
        warning: 'rgb(var(--color-warning) / <alpha-value>)',
        'warning-bg': 'rgb(var(--color-warning-bg) / <alpha-value>)',
        danger: 'rgb(var(--color-danger) / <alpha-value>)',
        'danger-bg': 'rgb(var(--color-danger-bg) / <alpha-value>)',
      },
      fontFamily: {
        sans: ['var(--font-sans)'],
        display: ['var(--font-display)'],
      },
      fontSize: {
        'body-sm': ['var(--font-size-body-sm)', { lineHeight: '1.55' }],
        body: ['var(--font-size-body)', { lineHeight: '1.7' }],
        'body-lg': ['var(--font-size-body-lg)', { lineHeight: '1.75' }],
        'title-sm': ['var(--font-size-title-sm)', { lineHeight: '1.2' }],
        title: ['var(--font-size-title)', { lineHeight: '1.15' }],
        hero: ['var(--font-size-hero)', { lineHeight: '1.02' }],
      },
      borderRadius: {
        sm: 'var(--radius-sm)',
        md: 'var(--radius-md)',
        lg: 'var(--radius-lg)',
        xl: 'var(--radius-xl)',
        '2xl': 'var(--radius-xl)',
        '3xl': 'var(--radius-2xl)',
        field: 'var(--radius-md)',
        panel: 'var(--radius-lg)',
        feature: 'var(--radius-xl)',
        pill: 'var(--radius-pill)',
      },
      boxShadow: {
        soft: 'var(--shadow-soft)',
        raised: 'var(--shadow-raised)',
        floating: 'var(--shadow-floating)',
      },
      spacing: {
        shell: 'var(--space-shell)',
        section: 'var(--space-section)',
        card: 'var(--space-card)',
        'stack-sm': 'var(--space-stack-sm)',
        'stack-md': 'var(--space-stack-md)',
        'stack-lg': 'var(--space-stack-lg)',
      },
      borderWidth: {
        thin: 'var(--border-width-thin)',
        strong: 'var(--border-width-strong)',
      },
      backgroundImage: {
        'shell-glow':
          'radial-gradient(circle at top left, rgba(var(--color-page-tint), 0.55), transparent 46%), radial-gradient(circle at bottom right, rgba(var(--color-primary-soft), 0.26), transparent 42%)',
      },
    },
  },
  plugins: [],
};
