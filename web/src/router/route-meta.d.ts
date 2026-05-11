/*
 * Vue Router meta 型別補充。
 * 1. 定義登入與 Staff 權限守衛欄位。
 * 2. 定義文件標題與頂部導航顯示欄位。
 */
import 'vue-router';

declare module 'vue-router' {
  interface RouteMeta {
    requiresAuth?: boolean;
    requiresStaff?: boolean;
    titleKey?: string;
    headerMode?: 'transparent';
  }
}
