import {
  CSSProperties,
  RefObject,
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useState,
} from 'react';

type Placement = 'bottom' | 'top';
type Align = 'start' | 'center' | 'end';

export function useAnchoredPosition<T extends HTMLElement | null>(
  anchorRef: RefObject<T>,
  {
    placement = 'bottom',
    align = 'start',
    offset = 8,
  }: { placement?: Placement; align?: Align; offset?: number } = {},
) {
  const [pos, setPos] = useState<{ left: number; top: number } | null>(null);
  const [actualPlacement, setActualPlacement] = useState<Placement>(placement);

  const update = useCallback(() => {
    const el = anchorRef.current;

    if (!el) return;

    const rect = el.getBoundingClientRect();
    const vw = window.innerWidth;
    const vh = window.innerHeight;

    let top = rect.bottom + offset;
    let usedPlacement: Placement = 'bottom';

    if (
      placement === 'bottom' &&
      rect.bottom + offset > vh &&
      rect.top - offset >= 0
    ) {
      top = rect.top - offset;
      usedPlacement = 'top';
    } else if (placement === 'top') {
      top = rect.top - offset;
      usedPlacement = 'top';
      if (rect.top - offset < 0) {
        top = rect.bottom + offset;
        usedPlacement = 'bottom';
      }
    }

    let left = rect.left;

    if (align === 'center') left = rect.left + rect.width / 2;
    if (align === 'end') left = rect.right;

    const clamp = (v: number, min: number, max: number) =>
      Math.min(Math.max(v, min), max);

    left = clamp(left, 8, vw - 8);

    setPos({ left, top });
    setActualPlacement(usedPlacement);
  }, [anchorRef, placement, align, offset]);

  useLayoutEffect(() => {
    update();
  }, []);

  useEffect(() => {
    const ro = new ResizeObserver(update);
    const el = anchorRef.current;

    update();

    const mutationCallback = () => {
      update();
    };

    const mutationConfig = {
      attributes: true,
      childList: true,
      subtree: true,
    };

    const mo = new MutationObserver(mutationCallback);

    if (el) {
      ro.observe(el);

      mo.observe(el, mutationConfig);
    }

    let rafId: number | null = null;

    const batchedUpdate = () => {
      if (rafId !== null) return;
      rafId = requestAnimationFrame(() => {
        rafId = null;
        update();
      });
    };

    window.addEventListener('scroll', batchedUpdate, true);
    window.addEventListener('resize', batchedUpdate);

    return () => {
      ro.disconnect();
      mo.disconnect();
      if (rafId !== null) cancelAnimationFrame(rafId);
      window.removeEventListener('scroll', batchedUpdate, true);
      window.removeEventListener('resize', batchedUpdate);
    };
  }, [anchorRef]);

  const style = useMemo<CSSProperties>(() => {
    if (!pos) return { visibility: 'hidden' };
    let translateX = '0';

    if (align === 'center') translateX = '-50%';
    if (align === 'end') translateX = '-100%';

    const translateY = actualPlacement === 'top' ? '-100%' : '0';

    return {
      position: 'fixed',
      left: pos.left,
      top: pos.top,
      transform: `translate(${translateX}, ${translateY})`,
      zIndex: 2147483648,
    };
  }, [pos, align, actualPlacement]);

  return { style, placement: actualPlacement, update };
}
