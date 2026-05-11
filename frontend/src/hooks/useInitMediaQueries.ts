// ~hooks/useInitMediaQueryListeners.ts
'use client';
import { useEffect } from 'react';

import { useMediaQueryStore } from '~store/mediaStore';

type MediaQueryListener = (event: MediaQueryListEvent) => void;

export const useInitMediaQueryListeners = () => {
  const setState = useMediaQueryStore((s) => s.setMediaQueryState);
  const setHydrated = useMediaQueryStore((s) => s.setHydrated);

  // Set real values after hydration - store starts with desktop defaults
  // to prevent SSR/client mismatch
  useEffect(() => {
    const queries = {
      isPhone: window.matchMedia('(max-width: 600px)'),
      isSmallTablet: window.matchMedia('(max-width: 834px)'),
      isTablet: window.matchMedia('(max-width: 1052px)'),
      isMedium: window.matchMedia('(max-width: 1360px)'),
    };

    // Update state to ensure sync (values should already match from store init)
    setState({
      isPhone: queries.isPhone.matches,
      isSmallTablet: queries.isSmallTablet.matches,
      isTablet: queries.isTablet.matches,
      isMedium: queries.isMedium.matches,
    });
    setHydrated();

    const listeners: { [key in keyof typeof queries]: MediaQueryListener } = {
      isPhone: (e) => setState({ isPhone: e.matches }),
      isSmallTablet: (e) => setState({ isSmallTablet: e.matches }),
      isTablet: (e) => setState({ isTablet: e.matches }),
      isMedium: (e) => setState({ isMedium: e.matches }),
    };

    for (const key in queries) {
      queries[key as keyof typeof queries].addEventListener(
        'change',
        listeners[key as keyof typeof listeners],
      );
    }

    return () => {
      for (const key in queries) {
        queries[key as keyof typeof queries].removeEventListener(
          'change',
          listeners[key as keyof typeof listeners],
        );
      }
    };
  }, [setState, setHydrated]);
};
