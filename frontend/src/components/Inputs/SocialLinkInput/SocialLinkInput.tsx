import {
  ChangeEvent,
  CSSProperties,
  ReactNode,
  useCallback,
  useMemo,
  useRef,
} from 'react';
import cx from 'classnames';

import { FadeAnimation } from '~components/Animations/FadeAnimation';
import { Cross } from '~icons/JsxSvg/Cross';

import styles from './styles.module.scss';

interface Props {
  value: string;
  error?: string;
  icon: ReactNode;
  placeholder: string;
  iconColor?: string;
  activeIconColor?: string;
  iconBackgroundColor?: string;
  activeIconBackgroundColor?: string;
  validate?: (value: string) => string | undefined;
  onChange: (value: string, newError?: string) => void;
  disabled?: boolean;
  className?: string;
}

const normalizeError = (error?: string) => {
  const trimmed = error?.trim();

  return trimmed ? trimmed : undefined;
};

export const SocialLinkInput = ({
  value,
  error,
  icon,
  placeholder,
  iconColor,
  activeIconColor,
  iconBackgroundColor,
  activeIconBackgroundColor,
  validate,
  onChange,
  disabled = false,
  className,
}: Props) => {
  const input = useRef<HTMLInputElement | null>(null);
  const activeError = useMemo(() => normalizeError(error), [error]);
  const variables = useMemo(
    () =>
      ({
        '--social-link-input-icon-color': iconColor,
        '--social-link-input-active-icon-color': activeIconColor,
        '--social-link-input-icon-bg': iconBackgroundColor,
        '--social-link-input-active-icon-bg': activeIconBackgroundColor,
      }) as CSSProperties,
    [
      iconColor,
      activeIconColor,
      iconBackgroundColor,
      activeIconBackgroundColor,
    ],
  );

  const emitChange = useCallback((nextValue: string) => {
    const validatedError = normalizeError(validate?.(nextValue));

    onChange(nextValue, validatedError);
  }, []);

  const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
    emitChange(event.target.value);
  };

  const handleClear = (e: any) => {
    e.stopPropagation();
    emitChange('');
  };

  return (
    <div
      className={cx(
        styles.socialLinkInput,
        {
          [styles.socialLinkInput__error as string]: !!activeError,
          [styles.socialLinkInput__disabled as string]: disabled,
        },
        className,
      )}
      style={variables}
      onClick={() => {
        input.current?.focus();
      }}
    >
      <div className={styles.socialLinkInput__field}>
        <div className={styles.socialLinkInput__icon}>{icon}</div>
        <input
          type="text"
          value={value}
          ref={input}
          placeholder={placeholder}
          disabled={disabled}
          onChange={handleChange}
          className={styles.socialLinkInput__input}
        />
        <FadeAnimation
          onClick={handleClear}
          isVisible={!!value && !disabled}
          customClassNames={{
            enter: styles.fadeEnter,
            enterActive: styles.fadeEnterActive,
            exit: styles.fadeExit,
            exitActive: styles.fadeExitActive,
          }}
          customWrapperClass={styles.socialLinkInput__clear}
        >
          <Cross />
        </FadeAnimation>
      </div>
      <FadeAnimation
        isVisible={!!activeError}
        customWrapperClass={styles.socialLinkInput__errorText}
      >
        {activeError}
      </FadeAnimation>
    </div>
  );
};
