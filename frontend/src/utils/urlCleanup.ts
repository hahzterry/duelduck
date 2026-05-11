export const TRACKING_PARAMS = [
  'utm_source',
  'utm_medium',
  'utm_campaign',
  'utm_term',
  'utm_content',
  'gclid',
  'fbclid',
  'msclkid',
  'twclid',
  '_ga',
  'mc_cid',
  'mc_eid',
] as const;

export type TrackingParam = (typeof TRACKING_PARAMS)[number];

export const isTrackingParam = (param: string): boolean => {
  return TRACKING_PARAMS.includes(param as TrackingParam);
};

export const cleanTrackingParamsFromUrl = (): void => {
  if (typeof window === 'undefined') return;

  const url = new URL(window.location.href);
  const params = new URLSearchParams(url.search);

  let hasTrackingParams = false;

  for (const param of Array.from(params.keys())) {
    if (isTrackingParam(param)) {
      params.delete(param);
      hasTrackingParams = true;
    }
  }

  if (hasTrackingParams) {
    const cleanUrl =
      url.origin +
      url.pathname +
      (params.toString() ? `?${params.toString()}` : '') +
      url.hash;

    window.history.replaceState({}, '', cleanUrl);
  }
};

export const getCleanUrl = (url: string): string => {
  try {
    const urlObj = new URL(url);
    const params = new URLSearchParams(urlObj.search);

    for (const param of Array.from(params.keys())) {
      if (isTrackingParam(param)) {
        params.delete(param);
      }
    }

    return (
      urlObj.origin +
      urlObj.pathname +
      (params.toString() ? `?${params.toString()}` : '')
    );
  } catch {
    return url;
  }
};
