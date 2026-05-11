import { ReactNode, RefObject } from 'react';
import cx from 'classnames';

import styles from './styles.module.scss';

export const ButtonCheckTextAndIcon = ({
  children,
  onClick,
  active,
  width,
  className,
  ref,
}: {
  children: ReactNode;
  onClick?: () => void;
  active?: boolean;
  width?: string;
  className?: string;
  ref?: RefObject<HTMLButtonElement | null>;
}) => {
  return (
    <button
      className={cx(styles.button, {
        [`${styles['button--active']}`]: active,
        [className as string]: className,
      })}
      ref={ref}
      style={{ width }}
      onClick={onClick}
    >
      {children}
    </button>
  );
};
