export const getExternalSource = (): string | null => {
  try {
    if (window.self !== window.top) {
      if (document.referrer) {
        try {
          const url = new URL(document.referrer);

          return url.hostname;
        } catch {
          return document.referrer;
        }
      }

      return 'unknown';
    }

    return null;
  } catch {
    return null;
  }
};
