'use client';

import { createContext, ReactNode, useContext } from 'react';

export interface DeviceHints {
  isPhone: boolean;
  isSmallTablet: boolean;
  isTablet: boolean;
  isMedium: boolean;
}

const defaults: DeviceHints = {
  isPhone: false,
  isSmallTablet: false,
  isTablet: false,
  isMedium: false,
};

const DeviceHintsContext = createContext<DeviceHints>(defaults);

export const useDeviceHintsContext = () => useContext(DeviceHintsContext);

export function DeviceHintsProvider({
  hints,
  children,
}: {
  hints: string | null;
  children: ReactNode;
}) {
  const value: DeviceHints = hints ? JSON.parse(hints) : defaults;

  return (
    <DeviceHintsContext.Provider value={value}>
      {children}
    </DeviceHintsContext.Provider>
  );
}
