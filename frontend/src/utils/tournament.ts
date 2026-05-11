export function getPartnerXUsername(xUrl?: string): string {
  if (!xUrl) return '';
  try {
    const url = new URL(xUrl);

    return url.pathname.replace(/^\//, '').replace(/\/$/, '');
  } catch {
    return xUrl.replace(/.*\//, '');
  }
}
