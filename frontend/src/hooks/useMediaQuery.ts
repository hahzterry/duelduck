import { useDeviceHintsContext } from '~contexts/deviceHints';
import { useMediaQueryStore } from '~store/mediaStore';

const useMediaQuery = () => {
  const { isPhone, isTablet, isMedium, isSmallTablet, isHydrated } =
    useMediaQueryStore();
  const hints = useDeviceHintsContext();

  // Server: isHydrated=false → use UA-based hints from React Context
  // Client: isHydrated=true → use Zustand store (window.__DEVICE_HINTS__ then matchMedia)
  if (isHydrated) {
    return { isPhone, isTablet, isMedium, isSmallTablet };
  }

  return {
    isPhone: hints.isPhone,
    isSmallTablet: hints.isSmallTablet,
    isTablet: hints.isTablet,
    isMedium: hints.isMedium,
  };
};

export default useMediaQuery;
