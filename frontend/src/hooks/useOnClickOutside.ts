import { RefObject, useEffect } from 'react';

export function useOnClickOutside<T extends HTMLElement>(
  ref: RefObject<T | null>,
  handler: (event: MouseEvent | TouchEvent) => void,
) {
  useEffect(() => {
    const listener = (event: MouseEvent | TouchEvent) => {
      const el = ref.current;

      if (!el || el.contains(event.target as Node)) {
        return; // click inside -> ignore
      }

      handler(event); // click outside -> trigger
    };

    document.body.addEventListener('mousedown', listener);
    document.body.addEventListener('touchstart', listener);

    return () => {
      document.body.removeEventListener('mousedown', listener);
      document.body.removeEventListener('touchstart', listener);
    };
  }, [ref, handler]);
}
