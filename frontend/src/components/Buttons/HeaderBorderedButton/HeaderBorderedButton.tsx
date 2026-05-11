import { ReactNode, RefObject } from 'react';
import cx from 'classnames';
import Link from 'next/link';

import styles from './styles.module.scss';

interface Props {
  text: string;
  onClick?: () => void;
  icon?: ReactNode;
  className?: string;
  link?: string;
  onMouseEnter?: () => void;
  onMouseLeave?: () => void;
  ref?: RefObject<HTMLButtonElement | HTMLAnchorElement | null>;
  onMouseMove?: () => void;
}

export const HeaderBorderedButton = ({
  text,
  onClick,
  icon,
  className,
  link,
  onMouseLeave,
  onMouseMove,
  onMouseEnter,
  ref,
}: Props) => {
  if (link) {
    return (
      <Link
        className={cx(styles.container, className)}
        href={link}
        onClick={onClick}
        ref={ref as RefObject<HTMLAnchorElement>}
        onMouseMove={onMouseMove}
        onMouseLeave={onMouseLeave}
        onMouseEnter={onMouseEnter}
        aria-label={text}
      >
        {text}
        {icon && icon}
      </Link>
    );
  }

  return (
    <button
      className={cx(styles.container, className)}
      ref={ref as RefObject<HTMLButtonElement>}
      onClick={onClick}
    >
      {text}
      {icon && icon}
    </button>
  );
};
