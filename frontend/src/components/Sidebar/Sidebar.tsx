import './styles.scss';

import { ReactNode, useEffect, useRef } from 'react';
import cx from 'classnames';
import dynamic from 'next/dynamic';

import { FadeAnimation } from '~components/Animations/FadeAnimation';
import useMediaQuery from '~hooks/useMediaQuery';
import { useOuterClick } from '~hooks/useOuterClick';
import { useSolanaWallet } from '~hooks/useSolanaWallet';
import { ArrowRight } from '~icons/JsxSvg/ArrowRight';
import { Close } from '~icons/JsxSvg/Close';
import { useSidebarStore } from '~store/sidebarStore';
import { SIDEBAR_CONTENT } from '~types/general';

const SelectTypeSignIn = dynamic(
  () => import('./components/SelectTypeSignIn').then((m) => m.SelectTypeSignIn),
  { ssr: true, loading: () => <div /> },
);

const SignInWithEmail = dynamic(
  () => import('./components/SignInWithEmail').then((m) => m.SignInWithEmail),
  { ssr: true, loading: () => <div /> },
);

const ConfirmEmail = dynamic(
  () =>
    import('~components/Sidebar/components/ConfirmEmail').then(
      (m) => m.ConfirmEmail,
    ),
  { ssr: true, loading: () => <div /> },
);

const sidebarContentsMap: Record<SIDEBAR_CONTENT, ReactNode> = {
  [SIDEBAR_CONTENT.SELECT_TYPE_SIGN_IN]: <SelectTypeSignIn />,
  [SIDEBAR_CONTENT.SIGN_IN_WITH_EMAIL]: <SignInWithEmail />,
  [SIDEBAR_CONTENT.CONFIRM_EMAIL]: <ConfirmEmail />,
};

export const Sidebar = () => {
  const {
    isSidebarOpen,
    content: contentType,
    closeSidebar,
    back,
  } = useSidebarStore();
  const { isTablet } = useMediaQuery();
  const { isLoading } = useSolanaWallet();
  const contentRef = useRef<HTMLDivElement>(null);
  const isLoadingRef = useRef(isLoading);

  useEffect(() => {
    isLoadingRef.current = isLoading;
  }, [isLoading]);

  const content = sidebarContentsMap[contentType as SIDEBAR_CONTENT];

  useOuterClick(contentRef, () => {
    if (isLoadingRef.current) return;
    closeSidebar();
  });

  const signInEmailArray = [
    SIDEBAR_CONTENT.SIGN_IN_WITH_EMAIL,
    SIDEBAR_CONTENT.CONFIRM_EMAIL,
  ];

  const hideBlur: SIDEBAR_CONTENT[] = [];
  const dontHavePadding: SIDEBAR_CONTENT[] = [];

  const haveBack: SIDEBAR_CONTENT[] = [...signInEmailArray];

  return (
    <>
      <div
        className={cx('sideBarBlurryBackground', {
          ['sideBarBlurryBackground--active']:
            isSidebarOpen &&
            !!(!isTablet && contentType && !hideBlur.includes(contentType)),
        })}
        onClick={() => closeSidebar()}
      />
      <div
        className={cx('containerSidebar', {
          ['containerSidebar--active']: isSidebarOpen,
        })}
        role="dialog"
        aria-modal="true"
        aria-label="Sidebar"
        style={{
          backgroundColor: undefined,
        }}
        ref={contentRef}
      >
        <FadeAnimation
          isVisible={!!contentType && haveBack.includes(contentType)}
        >
          <button className={'containerSidebar__back'} onClick={back}>
            <ArrowRight />
            Back
          </button>
        </FadeAnimation>
        <div
          className={'contentSidebar'}
          style={{
            padding:
              contentType && dontHavePadding.includes(contentType)
                ? 'unset'
                : undefined,
          }}
        >
          {content}
        </div>
        <button
          className={'containerSidebar__closeSidebarButton'}
          onClick={() => closeSidebar()}
          aria-label="Close sidebar"
        >
          <Close />
        </button>
      </div>
    </>
  );
};
