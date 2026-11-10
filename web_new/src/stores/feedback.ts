/*
 * 全域提示狀態。
 * 1. 管理輕量成功與錯誤提示。
 * 2. 為發布、登入與重新發布流程提供視覺回饋。
 */
import { defineStore } from 'pinia';

type FeedbackTone = 'success' | 'error' | 'info';

interface FeedbackToast {
  id: number;
  message: string;
  tone: FeedbackTone;
}

let toastId = 0;

// 1. 建立全域提示 Store
export const useFeedbackStore = defineStore('feedback', {
  state: () => ({
    toasts: [] as FeedbackToast[],
  }),
  actions: {
    // 2. 新增提示項目
    pushToast(message: string, tone: FeedbackTone = 'info') {
      const id = ++toastId;

      this.toasts.push({ id, message, tone });

      window.setTimeout(() => {
        this.dismissToast(id);
      }, 2800);
    },

    // 3. 移除指定提示
    dismissToast(id: number) {
      this.toasts = this.toasts.filter((item) => item.id !== id);
    },
  },
});
