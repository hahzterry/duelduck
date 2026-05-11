import { create } from 'zustand';

import CookieManager from '~utils/cookieManager';

interface State {
  isLoaded: boolean;
  cloudflareToken: string | null;
}

interface Actions {
  setIsLoaded: (isLoaded: boolean) => void;
  setCloudflareToken: (token: string | null) => void;
}

const initialState: State = {
  isLoaded: false,
  cloudflareToken: null,
};

export const useAppStore = create<State & Actions>((set) => ({
  ...initialState,
  setIsLoaded: (isLoaded) => set({ isLoaded }),
  setCloudflareToken: (cloudflareToken) => {
    if (cloudflareToken)
      CookieManager.setItem(
        'cloudflare_token',
        cloudflareToken,
        Date.now() + 5 * 60 * 1000,
      );
    set({ cloudflareToken });
  },
}));
