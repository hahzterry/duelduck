declare global {
  interface Window {
    gtag?: (
      command: string,
      action: string,
      params?: Record<string, string | number | boolean>,
    ) => void;
    twq?: (...args: unknown[]) => void;
  }
}

export function trackEvent(
  action: string,
  params?: Record<string, string | number | boolean>,
) {
  if (typeof window !== 'undefined' && typeof window.gtag === 'function') {
    window.gtag('event', action, params);
  }
}

export function trackXEvent(
  eventId: string,
  params?: Record<string, string | number | boolean | null>,
) {
  if (typeof window !== 'undefined' && typeof window.twq === 'function') {
    window.twq('event', eventId, params);
  }
}
