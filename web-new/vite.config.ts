/*
 * Vite 建置設定。
 * 1. 啟用 Vue 單檔元件支援。
 * 2. 配置 `@/` 路徑別名、開發代理與基礎 vendor 分包。
 */
import { fileURLToPath, URL } from 'node:url';

import { defineConfig, loadEnv } from 'vite';
import vue from '@vitejs/plugin-vue';

// 1. 解析第三方依賴分包名稱
const resolveVendorChunk = (id: string): string | undefined => {
  const normalizedID = id.replaceAll('\\', '/');

  if (
    normalizedID.includes('/node_modules/vue/') ||
    normalizedID.includes('/node_modules/@vue/') ||
    normalizedID.includes('/node_modules/vue-router') ||
    normalizedID.includes('/node_modules/pinia') ||
    normalizedID.includes('/node_modules/vue-i18n')
  ) {
    return 'vue-vendor';
  }

  if (normalizedID.includes('/node_modules/axios')) {
    return 'http-vendor';
  }

  if (
    normalizedID.includes('/node_modules/qrcode') ||
    normalizedID.includes('/node_modules/dijkstrajs') ||
    normalizedID.includes('/node_modules/encode-utf8')
  ) {
    return 'media-vendor';
  }

  if (
    normalizedID.includes('/node_modules/gsap') ||
    normalizedID.includes('/node_modules/@studio-freight/lenis')
  ) {
    return 'motion-vendor';
  }

  return undefined;
};

// 2. 輸出前端建置設定
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const apiProxyTarget = env.VITE_API_PROXY_TARGET || 'http://127.0.0.1:8080';

  return {
    plugins: [vue()],
    server: {
      proxy: {
        '/api/v1': {
          target: apiProxyTarget,
          changeOrigin: true,
        },
      },
    },
    build: {
      rollupOptions: {
        output: {
          manualChunks(id) {
            return resolveVendorChunk(id);
          },
        },
      },
    },
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
  };
});
