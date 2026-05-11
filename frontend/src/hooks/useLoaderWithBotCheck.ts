import { useEffect } from 'react';
import { isbot } from 'isbot';

export function useLoaderWithBotCheck(setIsLoaded: (value: boolean) => void) {
  useEffect(() => {
    const userAgent = navigator.userAgent;
    const crawler = isbot(userAgent);

    if (crawler) {
      setIsLoaded(true);
    }
  }, []);
}
