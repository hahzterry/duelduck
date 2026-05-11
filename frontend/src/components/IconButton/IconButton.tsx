import { CSSProperties, MouseEventHandler, ReactNode } from 'react';
import cx from 'classnames';

import styles from './styles.module.scss';

interface IconButtonProps {
  children: ReactNode;
  variant?: 'primary' | 'secondary';
  onClick?: MouseEventHandler<HTMLButtonElement>;
  style?: CSSProperties;
  disabled?: boolean;
  isLarge?: boolean;
  iconColor?: string;
  className?: string;
  ariaLabel: string;
}

export const IconButton = ({
  children,
  variant = 'primary',
  onClick = () => {},
  style,
  disabled,
  isLarge,
  className,
  ariaLabel,
}: IconButtonProps) => {
  return (
    <button
      type="button"
      style={style}
      onClick={onClick}
      disabled={disabled}
      className={cx(styles.iconButton, {
        [styles.secondary as string]: variant === 'secondary',
        [styles.disabled as string]: disabled,
        [styles.large as string]: isLarge,
        [className as string]: className,
      })}
      aria-label={ariaLabel}
    >
      {children}
    </button>
  );
};
