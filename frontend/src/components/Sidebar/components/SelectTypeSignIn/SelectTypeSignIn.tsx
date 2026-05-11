'use client';
import './styles.scss';

import { Fragment } from 'react';

import { ButtonWithIcons } from '~components/Buttons/ButtonWithIcons';
import { ResponsibleGamingBanner } from '~components/ResponsibleGamingBanner';
import { Typography } from '~components/Typography';
import { useTurnstile } from '~hooks/useTurnStile';
import { ArrowRight } from '~icons/JsxSvg/ArrowRight';
import { EmailIcon } from '~icons/JsxSvg/EmailIcon';
import { GoogleIcon } from '~icons/JsxSvg/GoogleIcon';
import { useSidebarStore } from '~store/sidebarStore';
import { SIDEBAR_CONTENT } from '~types/general';
import { signInWithGooglePopup } from '~utils/firebase';

import styles from './styles.module.scss';

export const SelectTypeSignIn = () => {
  const { closeSidebar, openSidebar } = useSidebarStore();
  const { cloudFlareContainerRef, isSubmitButtonDisabled } = useTurnstile();

  const buttonsConfig = [
    {
      label: 'gmail',
      startIcon: <GoogleIcon />,
      endIcon: <ArrowRight />,
      onClick: async () => {
        if (isSubmitButtonDisabled) return;

        const isSignedIn = await signInWithGooglePopup();

        if (isSignedIn) closeSidebar();
      },
    },
    {
      label: 'E-MAIL',
      startIcon: <EmailIcon />,
      endIcon: <ArrowRight />,
      onClick: () => {
        if (isSubmitButtonDisabled) return;
        openSidebar(SIDEBAR_CONTENT.SIGN_IN_WITH_EMAIL);
      },
    },
  ];

  return (
    <div className={'contentSelectTypeSignIn'}>
      <div className={'contentSelectTypeSignIn__buttonsAndTexts'}>
        <div className={'contentSelectTypeSignIn__texts'}>
          <Typography
            className={'contentSelectTypeSignIn__title'}
            text="join duel duck via"
          />
        </div>
        <div ref={cloudFlareContainerRef} className={styles.cloudflare} />
        <div
          className={'contentSelectTypeSignIn__buttons'}
          style={{
            opacity: isSubmitButtonDisabled ? '0.5' : undefined,
            pointerEvents: isSubmitButtonDisabled ? 'none' : undefined,
          }}
        >
          {buttonsConfig.map(
            ({ label, startIcon, endIcon, onClick }, index) => (
              <Fragment key={index}>
                {(index + 1) % 2 !== 0 && (
                  <div className="contentSelectTypeSignIn__divider" />
                )}
                <ButtonWithIcons
                  startIcon={startIcon}
                  endIcon={endIcon}
                  className="contentSelectTypeSignIn__buttonLogin"
                  isBungee
                  textVariant={'body'}
                  onClick={onClick}
                >
                  {label}
                </ButtonWithIcons>
              </Fragment>
            ),
          )}
        </div>
        <ResponsibleGamingBanner variant="compact" />
      </div>
    </div>
  );
};
