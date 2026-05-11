import { MouseEventHandler, ReactNode } from 'react';
import cx from 'classnames';
import Link from 'next/link';

import styles from './styles.module.scss';

interface Props {
  text: ReactNode;
  onClick?: () => void;
  disabled?: boolean;
  className?: string;
  href?: string;
  onMouseEnter?: MouseEventHandler<HTMLButtonElement | HTMLAnchorElement>;
}

export const ButtonBorderText = ({
  text,
  onClick,
  className,
  disabled,
  href,
  onMouseEnter,
}: Props) => {
  const cn = cx(
    styles.buttonBorderText,
    { [`${styles['buttonBorderText--disabled']}`]: disabled },
    className,
  );

  if (href) {
    return (
      <Link
        className={cn}
        onClick={onClick}
        href={href}
        onMouseEnter={onMouseEnter}
      >
        {text}
      </Link>
    );
  }

  return (
    <button
      onClick={onClick}
      onMouseEnter={onMouseEnter}
      disabled={disabled}
      className={cn}
    >
      {text}
    </button>
  );
};
