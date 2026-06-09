/*
 * Vite 建置設定。
 * 1. 啟用 Vue 單檔元件支援。
 * 2. 配置 `@/` 路徑別名、開發代理與基礎 vendor 分包。
 */
import { fileURLToPath, URL } from 'node:url';

import { defineConfig, loadEnv } from 'vite';
import vue from '@vitejs/plugin-vue';

// 1. 輸出前端建置設定
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const apiProxyTarget = env.VITE_API_PROXY_TARGET || 'http://localhost:8081';

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
            if (
              id.includes('node_modules/vue/') ||
              id.includes('node_modules/vue-router') ||
              id.includes('node_modules/pinia') ||
              id.includes('node_modules/vue-i18n')
            ) {
              return 'vue-vendor';
            }

            return undefined;
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
