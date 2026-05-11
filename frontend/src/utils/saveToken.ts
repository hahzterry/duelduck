import CookieManager from '~utils/cookieManager';

class SaveToken {
  tokens: { [key: string]: string } = {};

  add(key: string, value: string, expires?: number) {
    CookieManager.setItem(key, value, expires);

    if (typeof window !== 'undefined') {
      try {
        localStorage.setItem(key, value);
      } catch {
        // Storage access can be blocked in embedded/private contexts.
      }
    }

    this.tokens[key] = value;
  }

  get(key: string) {
    const searchValue =
      typeof window !== 'undefined'
        ? new URLSearchParams(window.location.search).get(key)
        : null;

    let localStorageValue: string | null = null;

    if (typeof window !== 'undefined') {
      try {
        localStorageValue = localStorage.getItem(key);
      } catch {
        localStorageValue = null;
      }
    }

    return (
      this.tokens[key] ||
      searchValue ||
      CookieManager.getItem(key) ||
      localStorageValue
    );
  }
}

export const saveToken = new SaveToken();
