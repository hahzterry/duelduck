import { create } from 'zustand';

interface MediaQueryState {
  isMedium: boolean;
  isTablet: boolean;
  isSmallTablet: boolean;
  isPhone: boolean;
  isHydrated: boolean;
  setMediaQueryState: (state: Partial<MediaQueryState>) => void;
  setHydrated: () => void;
}

function getInitialMediaState() {
  if (typeof window !== 'undefined' && (window as any).__DEVICE_HINTS__) {
    const h = (window as any).__DEVICE_HINTS__;

    return {
      isPhone: !!h.isPhone,
      isSmallTablet: !!h.isSmallTablet,
      isTablet: !!h.isTablet,
      isMedium: !!h.isMedium,
    };
  }

  return {
    isMedium: false,
    isTablet: false,
    isSmallTablet: false,
    isPhone: false,
  };
}

const initial = getInitialMediaState();

export const useMediaQueryStore = create<MediaQueryState>((set) => ({
  ...initial,
  isHydrated: false,
  setMediaQueryState: (state) => set(state),
  setHydrated: () => set({ isHydrated: true }),
}));
