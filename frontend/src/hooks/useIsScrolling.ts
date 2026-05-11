import { useEffect, useRef, useState } from 'react';

export const useIsScrolling = (
  callback: (isScrolling: boolean) => void,
  delay = 150,
) => {
  const [isScrolling, setIsScrolling] = useState(false);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    const onScroll = () => {
      if (!isScrolling) {
        setIsScrolling(true);
      }

      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }

      timeoutRef.current = setTimeout(() => {
        setIsScrolling(false);
      }, delay);
    };

    document.body.addEventListener('scroll', onScroll, { passive: true });

    return () => {
      document.body.removeEventListener('scroll', onScroll);
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, [isScrolling, delay]);

  useEffect(() => {
    callback(isScrolling);
  }, [isScrolling]);

  return isScrolling;
};
