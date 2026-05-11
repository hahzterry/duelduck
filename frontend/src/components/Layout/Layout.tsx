'use client';

import './main.scss';

import { ReactNode } from 'react';
import { SkeletonTheme } from 'react-loading-skeleton';

import { useRouteEffects } from '~hooks/useRouteEffects';

import { LayoutOverlays } from './LayoutOverlays';

const config = {
  width: '100%',
  baseColor: '#000000',
  highlightColor: '#151515',
  duration: 1,
};

export const Layout = ({ children }: { children: ReactNode }) => {
  useRouteEffects();

  return (
    <SkeletonTheme {...config}>
      {children}
      <LayoutOverlays />
    </SkeletonTheme>
  );
};
