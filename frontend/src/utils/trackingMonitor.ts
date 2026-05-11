import { isTrackingParam } from './urlCleanup';

export const detectTrackingParamsInUrl = (): string[] => {
  if (typeof window === 'undefined') return [];

  const params = new URLSearchParams(window.location.search);
  const trackingParams: string[] = [];

  for (const key of Array.from(params.keys())) {
    if (isTrackingParam(key)) {
      trackingParams.push(key);
    }
  }

  return trackingParams;
};

export const logTrackingParamsDetected = (): void => {
  const trackingParams = detectTrackingParamsInUrl();

  if (trackingParams.length > 0) {
    console.warn(
      '[SEO Warning] Tracking parameters detected in URL:',
      trackingParams,
      'URL:',
      window.location.href,
    );
  }
};

export const setupTrackingParamsMonitoring = (): void => {
  if (typeof window === 'undefined') return;

  const checkInterval = setInterval(() => {
    const trackingParams = detectTrackingParamsInUrl();

    if (trackingParams.length > 0) {
      console.error(
        '[SEO ALERT] Tracking parameters still present after 5 seconds:',
        trackingParams,
        'This should not happen. Check URL cleanup logic.',
      );
    }
  }, 5000);

  setTimeout(() => {
    clearInterval(checkInterval);
  }, 30000);
};
