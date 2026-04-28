/*
 * Axios 實例設定。
 * 1. 統一處理 baseURL、Authorization 與 401 響應。
 * 2. 供後續真實 API 接入時直接復用。
 */
import axios from 'axios';

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api/v1';
const ACCESS_TOKEN_STORAGE_KEY = 'ajoliving.access-token';
const REFRESH_TOKEN_STORAGE_KEY = 'ajoliving.refresh-token';

// 1. 建立 HTTP 用戶端實例
const httpClient = axios.create({
  baseURL: BASE_URL,
  timeout: 10000,
});

// 2. 統一注入示範登入 token
httpClient.interceptors.request.use((config) => {
  const token = window.localStorage.getItem(ACCESS_TOKEN_STORAGE_KEY);

  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }

  return config;
});

// 3. 統一處理 401 響應
httpClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      window.localStorage.removeItem(ACCESS_TOKEN_STORAGE_KEY);
      window.localStorage.removeItem(REFRESH_TOKEN_STORAGE_KEY);
    }

    return Promise.reject(error);
  },
);

export default httpClient;
