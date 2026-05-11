'use client';

import { useEffect } from 'react';
import { ReadonlyURLSearchParams } from 'next/navigation';

import { REFERRAL_QUERY_PARAM_KEYS } from '~constants/referralTokens';
import { saveToken } from '~utils/saveToken';

export const useSaveLocalStorageLinkToken = (
  searchParams: ReadonlyURLSearchParams | null,
) => {
  useEffect(() => {
    if (!searchParams) return;

    for (const key of REFERRAL_QUERY_PARAM_KEYS) {
      const value = searchParams.get(key);

      if (value) {
        saveToken.add(key, value);
      }
    }
  }, [searchParams]);
};
