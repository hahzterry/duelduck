import { useCallback, useRef, useState } from 'react';

import useMediaQuery from '~hooks/useMediaQuery';

export const useShare = () => {
  const [isCopied, setIsCopied] = useState(false);
  const timerCopy = useRef<ReturnType<typeof setTimeout> | null>(null);
  const { isSmallTablet } = useMediaQuery();

  const share = useCallback(
    async (shareData: ShareData, forceValue?: 'share' | 'copy') => {
      if (
        ((isSmallTablet && navigator?.share) ||
          (forceValue === 'share' && navigator?.share)) &&
        forceValue !== 'copy'
      ) {
        await navigator.share(shareData);
      } else {
        await navigator.clipboard.writeText(
          shareData?.text || shareData?.url || '',
        );
        setIsCopied(true);
        timerCopy.current = setTimeout(() => {
          setIsCopied(false);
        }, 3000);
      }
    },
    [isSmallTablet],
  );

  return { share, isCopied };
};
