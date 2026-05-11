'use client';
import { MutableRefObject, useEffect } from 'react';

export const useOuterClick = (
  ref: MutableRefObject<HTMLElement | null>,
  callback: (event: Event) => void,
  eventType?: string,
  deps?: any[],
): void => {
  useEffect(() => {
    const handleClickOutside: EventListener = (event: Event) => {
      const mouseEvent = event as MouseEvent;
      const target = mouseEvent.target;

      if (
        target instanceof Element &&
        target.closest('[data-ignore-outer-click="true"]')
      ) {
        return;
      }

      if (ref.current && !ref.current.contains(mouseEvent.target as Node)) {
        callback(event);
      }
    };

    const events = eventType ? [eventType] : ['mousedown', 'touchstart'];

    events.forEach((e) => document.addEventListener(e, handleClickOutside));

    return () => {
      events.forEach((e) =>
        document.removeEventListener(e, handleClickOutside),
      );
    };
  }, [ref, callback, ...(deps || [])]);
};
