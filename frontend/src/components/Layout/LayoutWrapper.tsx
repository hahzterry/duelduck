'use client';

import { ReactNode } from 'react';

import HeaderClient from '~components/Header';
import NewFooter from '~components/NewFooter';

interface LayoutWrapperProps {
  children: ReactNode;
}

export const LayoutWrapper = ({ children }: LayoutWrapperProps) => {
  return (
    <>
      <HeaderClient />
      <div className={'mainContent'}>
        <div className={'mainContent__pageContainer'}>{children}</div>
        <NewFooter />
      </div>
    </>
  );
};
