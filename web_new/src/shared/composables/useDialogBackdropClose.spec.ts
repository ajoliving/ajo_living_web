/*
 * 彈窗遮罩關閉守衛測試。
 * 1. 驗證完整遮罩點擊會執行關閉。
 * 2. 驗證由內容區拖動到遮罩不會誤關閉。
 */
import { describe, expect, it, vi } from 'vitest';

import { useDialogBackdropClose } from './useDialogBackdropClose';

// 1. 建立測試用指標事件
const createPointerEvent = (
  pointerId: number,
  target: EventTarget,
  currentTarget: EventTarget,
) => ({ currentTarget, pointerId, target });

describe('useDialogBackdropClose', () => {
  it('closes only when the pointer starts and ends on the backdrop', () => {
    const close = vi.fn();
    const backdrop = new EventTarget();
    const panel = new EventTarget();
    const { handleBackdropPointerDown, handleBackdropPointerUp } = useDialogBackdropClose(close);

    handleBackdropPointerDown(createPointerEvent(1, backdrop, backdrop));
    handleBackdropPointerUp(createPointerEvent(1, backdrop, backdrop));
    expect(close).toHaveBeenCalledTimes(1);

    handleBackdropPointerDown(createPointerEvent(2, panel, backdrop));
    handleBackdropPointerUp(createPointerEvent(2, backdrop, backdrop));
    expect(close).toHaveBeenCalledTimes(1);
  });

  it('does not close after leaving the backdrop or cancelling the pointer', () => {
    const close = vi.fn();
    const backdrop = new EventTarget();
    const panel = new EventTarget();
    const {
      handleBackdropPointerCancel,
      handleBackdropPointerDown,
      handleBackdropPointerUp,
    } = useDialogBackdropClose(close);

    handleBackdropPointerDown(createPointerEvent(3, backdrop, backdrop));
    handleBackdropPointerUp(createPointerEvent(3, panel, backdrop));
    handleBackdropPointerDown(createPointerEvent(4, backdrop, backdrop));
    handleBackdropPointerCancel(createPointerEvent(4, backdrop, backdrop));
    handleBackdropPointerUp(createPointerEvent(4, backdrop, backdrop));
    handleBackdropPointerDown(createPointerEvent(5, backdrop, backdrop));
    handleBackdropPointerDown(createPointerEvent(5, panel, backdrop));
    handleBackdropPointerUp(createPointerEvent(5, backdrop, backdrop));

    expect(close).not.toHaveBeenCalled();
  });
});
