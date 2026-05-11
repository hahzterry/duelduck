import { useCallback, useEffect, useRef, useState } from 'react';

import { IS_STAGE } from '~constants/api';
import { cloudflareCaptchaSiteKey } from '~constants/main';
import { useAppStore } from '~store/appStore';

declare global {
  interface Window {
    turnstile?: {
      render: (
        container: HTMLElement | string,
        options: {
          sitekey: string;
          callback: (token: string) => void;
        },
      ) => string;
      reset: (id: string) => void;
    };
  }
}

export const useTurnstile = () => {
  const [isLoaded, setIsLoaded] = useState(false);
  const [renderVersion, setRenderVersion] = useState(0);
  const { setCloudflareToken } = useAppStore();

  const containerRef = useRef<HTMLDivElement | null>(null);
  const widgetIdRef = useRef<string | null>(null);

  const [isDisabled, setIsDisabled] = useState(true);

  const renderTurnstile = useCallback(() => {
    if (!containerRef.current) return;

    containerRef.current.innerHTML = '';
    setCloudflareToken(null);
    setIsDisabled(true);

    let attempts = 0;
    let retryTimer: ReturnType<typeof setTimeout>;

    const tryRender = () => {
      if (containerRef.current && window.turnstile) {
        widgetIdRef.current = window.turnstile.render(containerRef.current, {
          sitekey: cloudflareCaptchaSiteKey,
          callback: (token: string) => {
            setCloudflareToken(token);
            setIsDisabled(false);
          },
        });

        return;
      }

      if (attempts < 5) {
        attempts++;
        retryTimer = setTimeout(tryRender, 300);
      }
    };

    tryRender();

    return () => clearTimeout(retryTimer);
  }, []);

  useEffect(() => {
    if (window.turnstile) {
      setIsLoaded(true);

      return;
    }

    const script = document.createElement('script');

    script.src =
      'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit';
    script.async = true;
    script.defer = true;
    script.onload = () => setIsLoaded(true);

    document.body.appendChild(script);

    return () => {
      script.remove();
    };
  }, []);

  useEffect(() => {
    if (isLoaded) {
      renderTurnstile();
    }
  }, [isLoaded, renderVersion]);

  const rerenderWidget = useCallback(() => {
    setRenderVersion((v) => v + 1);
  }, []);

  const resetWidget = useCallback(() => {
    setCloudflareToken(null);
    setIsDisabled(true);

    if (widgetIdRef.current && window.turnstile) {
      window.turnstile.reset(widgetIdRef.current);
    }

    rerenderWidget();
  }, []);

  return {
    cloudFlareContainerRef: containerRef,
    resetWidget,
    rerenderWidget,
    renderTurnstile,
    isSubmitButtonDisabled: isDisabled && !IS_STAGE,
    isCloudFlareLoaded: isLoaded,
  };
};
