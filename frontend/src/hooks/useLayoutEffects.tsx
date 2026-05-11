'use client';

import { useEffect } from 'react';
import { usePathname } from 'next/navigation';

import { CUSTOM_EVENT_KEYS } from '~types/general';
import { customEvent } from '~utils/customEvent';
import { openLoginSelect } from '~utils/loginService';

export const useLayoutEffects = () => {
  const pathname = usePathname() as string;

  useEffect(() => {
    const handleOpenLogin = () => {
      openLoginSelect();
    };

    customEvent.on(CUSTOM_EVENT_KEYS.LOGIN, handleOpenLogin);

    return () => {
      customEvent.off(CUSTOM_EVENT_KEYS.LOGIN, handleOpenLogin);
    };
  }, [pathname]);
};
