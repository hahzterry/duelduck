import { useEffect, useRef, useState } from 'react';
import cx from 'classnames';
import { signIn } from 'next-auth/react';

import { DigitsInput } from '~components/Inputs/DigitsInput';
import { Typography } from '~components/Typography';
import { useTurnstile } from '~hooks/useTurnStile';
import { useSidebarStore } from '~store/sidebarStore';
import colors from '~styles/colors';
import { getTrackingTokens } from '~utils/cookieManager';

import styles from './styles.module.scss';

export const ConfirmEmail = () => {
  const [error, setError] = useState(false);
  const { closeSidebar, email, onConfirmEmailCode, setEmail} = useSidebarStore();
  const [cooldown, setCooldown] = useState(0);
  const resendRef = useRef(null);
  const [isShowResend, setIsShowResend] = useState(false);
  const { cloudFlareContainerRef, isSubmitButtonDisabled } = useTurnstile();

  useEffect(() => {
    let interval: ReturnType<typeof setTimeout>;

    if (cooldown > 0) {
      interval = setInterval(() => {
        setCooldown((prev) => prev - 1);
      }, 1000);
    }

    return () => clearInterval(interval);
  }, [cooldown]);

  const handleClose = () => {
    closeSidebar();
    setEmail(null);
  };

  const handleInputChange = async (v: string) => {
    setError(false);
    if (v.length === 6 && !isNaN(+v)) {
      if (!email) {
        setError(true);

        return;
      }

      try {
        const res = await signIn('email-code', {
          redirect: false,
          email,
          code: v,
          ...getTrackingTokens(),
        });

        if (!res?.ok) {
          setError(true);

          return;
        }

        handleClose();
        onConfirmEmailCode?.();
      } catch {
        setError(true);
      }
    }
  };

  async function handleSendAgain() {
    setError(false);
    setCooldown(3);
    // await signSendCode(email!);
  }

  useEffect(() => {
    const nesState = cooldown !== 0;

    if (nesState !== isShowResend) setIsShowResend(nesState);
  }, [cooldown, isShowResend]);

  return (
    <div className={styles.contentConfirmEmail}>
      <div className={styles.contentConfirmEmail__confirmYourEmailTexts}>
        <Typography
          text="Enter Confirmation Code"
          className={styles.contentConfirmEmail__title}
        />
        <div className={styles.contentConfirmEmail__descContainer}>
          <Typography
            color={colors['textSecondary']}
            text="Enter the 6-digit code that your received on "
          />
          <Typography color={colors['textPrimary']} text={email || ''} />
        </div>
      </div>

      <div
        ref={cloudFlareContainerRef}
        style={{
          margin: '0 auto',
        }}
      />

      <div
        className={styles.contentConfirmEmail__confirmYourEmailInputText}
        style={{
          opacity: isSubmitButtonDisabled ? '0.5' : undefined,
          pointerEvents: isSubmitButtonDisabled ? 'none' : undefined,
        }}
      >
        <DigitsInput
          onChange={handleInputChange}
          errorMsg={error ? 'please enter a valid code' : ''}
          isError={error}
        />
        {!isSubmitButtonDisabled && (
          <Typography
            text="Resend Code"
            ref={resendRef}
            onClick={handleSendAgain}
            className={cx(styles.contentConfirmEmail__resendCode, {
              [`${styles.fadeResendEnterActive}`]: !isShowResend,
              [`${styles.fadeResendExitActive}`]: isShowResend,
            })}
          />
        )}
      </div>
    </div>
  );
};
