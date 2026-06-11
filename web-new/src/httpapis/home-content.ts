/*
 * 首頁內容設定 API。
 * 1. 讀取公開首頁內容。
 * 2. 管理首頁輪播、三個主模組大圖與登入背景圖設定。
 */
import httpClient from '@/httpapis';
import type { ApiResponse } from '@/model/api';
import type {
  HomeCarouselImage,
  HomeCarouselSaveItem,
  HomeContentResponse,
  LoginHeroImageSetting,
  LoginHeroSaveItem,
  HomeModuleCard,
  HomeModuleCardSaveItem,
} from '@/model/home-content';

// 1. 取得公開首頁內容
export const fetchHomeContent = () =>
  httpClient.get<ApiResponse<HomeContentResponse>>('/home/content');

// 2. 取得首頁輪播設定
export const fetchHomeCarouselSettings = () =>
  httpClient.get<ApiResponse<{ items: HomeCarouselImage[] }>>('/home/settings/carousel');

// 3. 儲存首頁輪播設定
export const saveHomeCarouselSettings = (items: HomeCarouselSaveItem[]) =>
  httpClient.put<ApiResponse<{ items: HomeCarouselImage[] }>>('/home/settings/carousel', {
    items,
  });

// 4. 取得首頁三大圖設定
export const fetchHomeModuleCardSettings = () =>
  httpClient.get<ApiResponse<{ cards: HomeModuleCard[] }>>('/home/settings/module-cards');

// 5. 儲存首頁三大圖設定
export const saveHomeModuleCardSettings = (cards: HomeModuleCardSaveItem[]) =>
  httpClient.put<ApiResponse<{ cards: HomeModuleCard[] }>>('/home/settings/module-cards', {
    cards,
  });

// 6. 取得公開登入背景圖
export const fetchLoginHero = () =>
  httpClient.get<ApiResponse<{ items: LoginHeroImageSetting[] }>>('/home/login-hero');

// 7. 取得登入背景圖設定
export const fetchLoginHeroSettings = () =>
  httpClient.get<ApiResponse<{ items: LoginHeroImageSetting[] }>>('/home/settings/login-hero');

// 8. 儲存登入背景圖設定
export const saveLoginHeroSettings = (items: LoginHeroSaveItem[]) =>
  httpClient.put<ApiResponse<{ items: LoginHeroImageSetting[] }>>('/home/settings/login-hero', {
    items,
  });
