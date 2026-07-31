/*
 * 彈窗遮罩關閉守衛。
 * 1. 只允許按下與放開都發生在遮罩本身時執行關閉。
 * 2. 阻止由彈窗內容拖動到遮罩後放開造成的誤關閉。
 */

interface DialogBackdropPointerEvent {
  currentTarget: EventTarget | null;
  pointerId: number;
  target: EventTarget | null;
}

// 1. 建立彈窗遮罩指標事件處理器
export const useDialogBackdropClose = (onClose: () => void) => {
  const backdropPointerIds = new Set<number>();

  // 1.1 記錄由遮罩本身開始的指標操作
  const handleBackdropPointerDown = (event: DialogBackdropPointerEvent): void => {
    backdropPointerIds.delete(event.pointerId);
    if (event.target === event.currentTarget) {
      backdropPointerIds.add(event.pointerId);
    }
  };

  // 1.2 只在同一次指標操作完整發生於遮罩時關閉
  const handleBackdropPointerUp = (event: DialogBackdropPointerEvent): void => {
    const startedOnBackdrop = backdropPointerIds.delete(event.pointerId);
    if (startedOnBackdrop && event.target === event.currentTarget) {
      onClose();
    }
  };

  // 1.3 指標操作取消時清除記錄
  const handleBackdropPointerCancel = (event: DialogBackdropPointerEvent): void => {
    backdropPointerIds.delete(event.pointerId);
  };

  return {
    handleBackdropPointerCancel,
    handleBackdropPointerDown,
    handleBackdropPointerUp,
  };
};
