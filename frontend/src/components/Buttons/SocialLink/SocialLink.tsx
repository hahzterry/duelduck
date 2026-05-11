'use client';

import { CSSProperties, ReactNode } from 'react';
import cx from 'classnames';
import Link from 'next/link';

import { Typography } from '~components/Typography';

import styles from './styles.module.scss';

type SocialLinkProps = {
  href?: string;
  icon: ReactNode;
  hoverIcon?: ReactNode;
  hoverIconColor?: string;
  hoverBackgroundColor?: string;
  text?: string;
  ariaLabelText?: string;
  className?: string;
  onClick?: () => void;
};

export const SocialLink = ({
  href,
  icon,
  hoverIcon,
  hoverIconColor,
  hoverBackgroundColor,
  text,
  className,
  ariaLabelText,
  onClick,
}: SocialLinkProps) => {
  if (!href) {
    return (
      <button
        className={cx(styles.link, className)}
        onClick={onClick}
        aria-label={ariaLabelText}
        style={
          {
            '--hover-color-background': hoverBackgroundColor,
            '--hover-color-icon': hoverIconColor,
          } as CSSProperties
        }
      >
        <div>
          {icon}
          {hoverIcon}
        </div>
        {text && <Typography text={text} />}
      </button>
    );
  }

  return (
    <Link
      href={href}
      onClick={onClick}
      className={cx(styles.link, className)}
      target="_blank"
      aria-label={ariaLabelText}
      rel="noopener noreferrer nofollow"
      style={
        {
          '--hover-color-background': hoverBackgroundColor,
          '--hover-color-icon': hoverIconColor,
        } as CSSProperties
      }
    >
      <div>
        {icon}
        {hoverIcon}
      </div>
      {text && <Typography text={text} />}
    </Link>
  );
};
