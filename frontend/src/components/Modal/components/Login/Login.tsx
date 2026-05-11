import cx from 'classnames';

import { ButtonWithIcons } from '~components/Buttons/ButtonWithIcons';
import { Typography } from '~components/Typography';
import { useTurnstile } from '~hooks/useTurnStile';
import { ArrowRight } from '~icons/JsxSvg/ArrowRight';
import { Cross } from '~icons/JsxSvg/Cross';
import { EmailIcon } from '~icons/JsxSvg/EmailIcon';
import { GoogleIcon } from '~icons/JsxSvg/GoogleIcon';
import useModalStore from '~store/modalStore';
import { useSidebarStore } from '~store/sidebarStore';
import colors from '~styles/colors';
import { SIDEBAR_CONTENT } from '~types/general';
import { signInWithGooglePopup } from '~utils/firebase';

import styles from './styles.module.scss';

export const Login = () => {
  const { closeModal } = useModalStore();
  const { openSidebar } = useSidebarStore();
  const { cloudFlareContainerRef, isSubmitButtonDisabled } = useTurnstile();

  const buttons = [
    {
      label: 'gmail',
      startIcon: <GoogleIcon />,
      onClick: async () => {
        if (isSubmitButtonDisabled) return;

        const isSignedIn = await signInWithGooglePopup();

        if (isSignedIn) closeModal();
      },
    },
    {
      label: 'E-MAIL',
      startIcon: <EmailIcon />,
      onClick: () => {
        if (isSubmitButtonDisabled) return;
        openSidebar(SIDEBAR_CONTENT.SIGN_IN_WITH_EMAIL);
        closeModal();
      },
    },
  ];

  return (
    <div className={styles.login}>
      <div className={styles.login__header}>
        <Typography text={'Log in methods'} />
        <button
          onClick={() => {
            closeModal();
          }}
          aria-label="Close login"
        >
          <Cross />
        </button>
      </div>
      <div className={styles.login__buttonsContainer}>
        <div
          className={styles.login__cloudflare}
          ref={cloudFlareContainerRef}
        />
        {buttons.map(({ label, startIcon, onClick }, index) => (
          <ButtonWithIcons
            key={index}
            startIcon={startIcon}
            endIcon={<ArrowRight />}
            className={cx(styles.login__buttonLogin, {
              [`${styles['login__buttonLogin--disabled']}`]:
                isSubmitButtonDisabled,
            })}
            backgroundColor={colors.backgroundDark}
            isBungee
            textVariant="body"
            width="100%"
            onClick={onClick}
          >
            {label}
          </ButtonWithIcons>
        ))}
      </div>
    </div>
  );
};
