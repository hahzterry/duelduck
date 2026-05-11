import {
  PointerEvent,
  ReactNode,
  TouchEvent,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
import cx from 'classnames';

import {
  BOTTOM_SHEET_STATE,
  useBottomSheetStore,
} from '~store/bottomSheetStore';
import { CUSTOM_EVENT_KEYS } from '~types/general';
import { customEvent } from '~utils/customEvent';

import styles from './styles.module.scss';

const labels: Record<BOTTOM_SHEET_STATE, string> = {};

const contents: Record<BOTTOM_SHEET_STATE, ReactNode> = {};

const isHideDragHandle: BOTTOM_SHEET_STATE[] = [];

export const BottomSheet = () => {
  const { content, setContent } = useBottomSheetStore();
  const [localContent, setLocalContent] = useState<BOTTOM_SHEET_STATE | null>(
    null,
  );
  const isVisible = !!content;
  const lastRouteRef = useRef<string | null>(null);

  const isHideDrag = useMemo(() => {
    return isHideDragHandle.includes(content!);
  }, [content]);

  useEffect(() => {
    const handle = (
      pathname?: string,
      searchParams?: { toString: () => string } | null,
    ) => {
      if (typeof pathname !== 'string') {
        setContent(null);

        return;
      }

      const routeKey = `${pathname}?${searchParams?.toString() || ''}`;

      if (lastRouteRef.current === null) {
        lastRouteRef.current = routeKey;

        return;
      }

      if (lastRouteRef.current === routeKey) {
        return;
      }

      lastRouteRef.current = routeKey;
      setContent(null);
    };

    customEvent.on(CUSTOM_EVENT_KEYS.CHANGE_ROUTE, handle);

    return () => {
      customEvent.off(CUSTOM_EVENT_KEYS.CHANGE_ROUTE, handle);
    };
  }, []);

  const sheetRef = useRef<HTMLDivElement>(null);
  const activeScrollableParentRef = useRef<HTMLElement | null>(null);
  const [offsetY, setOffsetY] = useState(0);
  const [isDragging, setIsDragging] = useState(false);
  const dragDeltaYRef = useRef(0);
  const startedFromDragHandleRef = useRef(false);
  const mousePointerIdRef = useRef<number | null>(null);

  const startYRef = useRef(0);

  const getScrollableParent = (target: HTMLElement | null) => {
    let node = target;

    while (node && node !== sheetRef.current) {
      const style = window.getComputedStyle(node);
      const isScrollableY = /(auto|scroll|overlay)/.test(style.overflowY);

      if (isScrollableY && node.scrollHeight > node.clientHeight) {
        return node;
      }

      node = node.parentElement;
    }

    return null;
  };

  useEffect(() => {
    setOffsetY(0);
    setIsDragging(false);
    startYRef.current = 0;
    dragDeltaYRef.current = 0;
    startedFromDragHandleRef.current = false;
    activeScrollableParentRef.current = null;

    if (content) {
      setLocalContent(content);
    }
  }, [content]);

  const startDrag = (
    target: HTMLElement | null,
    startClientY: number,
    isFromDragHandle: boolean,
  ): boolean => {
    if (isHideDrag) return false;

    const scrollableParent = getScrollableParent(target);

    activeScrollableParentRef.current = scrollableParent;

    if (!isFromDragHandle && scrollableParent) {
      setIsDragging(false);

      return false;
    }

    setIsDragging(true);
    startYRef.current = startClientY;
    dragDeltaYRef.current = 0;

    return true;
  };

  const moveDrag = (currentClientY: number) => {
    if (!isDragging || isHideDrag) return;

    if (activeScrollableParentRef.current) {
      return;
    }

    const deltaY = currentClientY - startYRef.current;

    if (deltaY > 0) setOffsetY(deltaY);
  };

  const endDrag = () => {
    if (isHideDrag || !isDragging) return;

    setIsDragging(false);

    activeScrollableParentRef.current = null;
    if (offsetY > 200) {
      setContent(null);
    } else {
      setOffsetY(0);
    }
  };

  const cancelDrag = () => {
    if (!isDragging) return;

    setIsDragging(false);
    activeScrollableParentRef.current = null;

    setOffsetY(0);
  };

  const handleTouchStart = (e: TouchEvent) => {
    const target = e.target as HTMLElement | null;
    const isFromDragHandle = !!target?.closest(`.${styles.dragHandle}`);
    const clientY = e.touches[0]?.clientY || 0;

    startDrag(target, clientY, isFromDragHandle);
  };

  const handleTouchMove = (e: TouchEvent) => {
    moveDrag(e.touches[0]?.clientY || 0);
  };

  const handleTouchEnd = () => {
    endDrag();
  };

  const handleTouchCancel = () => {
    cancelDrag();
  };

  const handlePointerDown = (e: PointerEvent<HTMLDivElement>) => {
    if (e.pointerType !== 'mouse' || e.button !== 0) return;

    const target = e.target as HTMLElement | null;
    const isFromDragHandle = !!target?.closest(`.${styles.dragHandle}`);
    const didStartDrag = startDrag(target, e.clientY, isFromDragHandle);

    if (!didStartDrag) return;

    mousePointerIdRef.current = e.pointerId;
    e.currentTarget.setPointerCapture(e.pointerId);
  };

  const handlePointerMove = (e: PointerEvent<HTMLDivElement>) => {
    if (e.pointerType !== 'mouse') return;
    if (mousePointerIdRef.current !== e.pointerId) return;

    moveDrag(e.clientY);
  };

  const handlePointerUp = (e: PointerEvent<HTMLDivElement>) => {
    if (e.pointerType !== 'mouse') return;
    if (mousePointerIdRef.current !== e.pointerId) return;

    if (e.currentTarget.hasPointerCapture(e.pointerId)) {
      e.currentTarget.releasePointerCapture(e.pointerId);
    }

    mousePointerIdRef.current = null;
    endDrag();
  };

  const handlePointerCancel = (e: PointerEvent<HTMLDivElement>) => {
    if (e.pointerType !== 'mouse') return;
    if (mousePointerIdRef.current !== e.pointerId) return;

    if (e.currentTarget.hasPointerCapture(e.pointerId)) {
      e.currentTarget.releasePointerCapture(e.pointerId);
    }

    mousePointerIdRef.current = null;
    cancelDrag();
  };

  const backgrounds: Partial<Record<BOTTOM_SHEET_STATE, string>> = {};

  return (
    <div
      ref={sheetRef}
      role="dialog"
      aria-modal="true"
      aria-label={localContent ? labels[localContent] : undefined}
      className={cx(styles.bottomSheet, {
        [`${styles['bottomSheet--active']}`]: isVisible,
        [`${styles['bottomSheet--dragging']}`]: isDragging && !isHideDrag,
      })}
      onTouchStart={handleTouchStart}
      onTouchMove={handleTouchMove}
      onTouchEnd={handleTouchEnd}
      onTouchCancel={handleTouchCancel}
      onPointerDown={handlePointerDown}
      onPointerMove={handlePointerMove}
      onPointerUp={handlePointerUp}
      onPointerCancel={handlePointerCancel}
      style={{
        transform: isVisible ? `translateY(${offsetY}px)` : undefined,
        background: localContent ? backgrounds[localContent] : undefined,
        transition: isDragging
          ? 'none'
          : 'transform 0.25s ease, max-height 0.25s ease, padding 0.25s ease, gap 0.25s ease',
      }}
    >
      {!isHideDrag && <div className={styles.dragHandle} />}
      {localContent ? contents[localContent] : null}
    </div>
  );
};
