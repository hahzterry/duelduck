'use client';

import { useEffect, useLayoutEffect, useState } from 'react';
import { usePathname, useSearchParams } from 'next/navigation';

import { useInitMediaQueryListeners } from '~hooks/useInitMediaQueries';
import { useLayoutEffects } from '~hooks/useLayoutEffects';
import { useSaveLocalStorageLinkToken } from '~hooks/useSaveLocalStorageLinkToken';
import { CUSTOM_EVENT_KEYS } from '~types/general';
import { customEvent } from '~utils/customEvent';

export const useRouteEffects = () => {
  const pathname = usePathname() as string;
  const searchParams = useSearchParams();

  const [isThrowError, setIsThrowError] = useState(false);

  useSaveLocalStorageLinkToken(searchParams);
  useInitMediaQueryListeners();
  useLayoutEffects();

  useLayoutEffect(() => {
    if ('scrollRestoration' in window.history) {
      window.history.scrollRestoration = 'manual';
    }
  }, []);

  useLayoutEffect(() => {
    document.body.scrollTo(0, 0);
  }, [pathname]);

  useEffect(() => {
    const scrollBodyTo = (top: number, behavior: ScrollBehavior = 'smooth') => {
      const safeTop = Math.max(0, Math.floor(top));

      document.body.scrollTo({ top: safeTop, behavior });
    };

    customEvent.on(CUSTOM_EVENT_KEYS.SCROLL_TO, scrollBodyTo);

    return () => {
      customEvent.off(CUSTOM_EVENT_KEYS.SCROLL_TO, scrollBodyTo);
    };
  }, []);

  useEffect(() => {
    customEvent.emit(CUSTOM_EVENT_KEYS.CHANGE_ROUTE, pathname, searchParams);
  }, [pathname, searchParams]);

  useEffect(() => {
    if (isThrowError && !(typeof window === 'undefined')) {
      setIsThrowError(false);
      throw new Error('Test error');
    }
  }, [isThrowError]);
};
