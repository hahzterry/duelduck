export const stringToHash = (title: string) => {
  return encodeURIComponent(
    (title || '')
      .toString()
      .trim()
      .toLowerCase()
      .normalize('NFKD')
      .replace(/[\u0300-\u036f]/g, '')
      .replace(/[^\p{L}\p{N}]+/gu, ' ')
      .trim()
      .replace(/\s+/g, '-'),
  );
};

export function normalizeCanonicalUrl(url: string): string {
  if (url === '/' || url.endsWith('://')) return url;

  return url.endsWith('/') ? url.slice(0, -1) : url;
}

export function normalizeRoute(route: string): string {
  if (route === '/') {
    return '/';
  }

  return route.endsWith('/') ? route.slice(0, -1) : route;
}
