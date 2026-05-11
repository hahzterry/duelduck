'use client';

import dynamic from 'next/dynamic';

import { PopupOnElement } from '~components/PopupOnElement';
import { useAppStore } from '~store/appStore';

const Modal = dynamic(() => import('~components/Modal'), {
  ssr: false,
  loading: () => <div />,
});

const Sidebar = dynamic(() => import('~components/Sidebar'), {
  ssr: false,
  loading: () => <div />,
});

const BottomSheet = dynamic(() => import('~components/BottomSheet'), {
  ssr: false,
  loading: () => <div />,
});

const Cookies = dynamic(() => import('~components/Cookies'), {
  ssr: false,
  loading: () => <div />,
});

const ErrorToast = dynamic(() => import('~components/Popups/ErrorToast'), {
  ssr: false,
  loading: () => <div />,
});

export const LayoutOverlays = () => {
  const { isLoaded } = useAppStore();

  return (
    <>
      {isLoaded && (
        <>
          <Modal />
          <Sidebar />
          <PopupOnElement />
          <Cookies />
          <ErrorToast />
          <BottomSheet />
        </>
      )}
    </>
  );
};
