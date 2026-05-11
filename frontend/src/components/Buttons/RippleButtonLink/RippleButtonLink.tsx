import { JSX, ReactNode, useState } from 'react';
import cx from 'classnames';
import Link from 'next/link';

import { FadeAnimation } from '~components/Animations/FadeAnimation';
import { Typography } from '~components/Typography';

import styles from './styles.module.scss';

interface RippleButtonLinkProps {
  className?: string;
  text: string;
  onClick?: () => void;
  href?: string;
  isLink?: boolean;
  icon?: ReactNode;
  activeIcon?: ReactNode;
  count?: number;
}

export const RippleButtonLink = ({
  text,
  isLink,
  className,
  icon,
  activeIcon,
  href,
  onClick,
  count,
}: RippleButtonLinkProps) => {
  const [hovered, setHovered] = useState(false);
  const Element: keyof JSX.IntrinsicElements = isLink ? 'span' : 'button';

  if (isLink) {
    return (
      <Link
        aria-label={text}
        href={href || '/'}
        className={cx(styles.button, className)}
        onClick={onClick}
        onMouseEnter={() => setHovered(true)}
        onMouseLeave={() => setHovered(false)}
      >
        {icon && (
          <span className={styles.button__iconWrapper}>
            <span
              className={cx(styles.button__icon, {
                [styles.visible as string]: !hovered,
              })}
            >
              {icon}
            </span>
            <span
              className={cx(styles.button__icon, {
                [styles.visible as string]: hovered,
              })}
            >
              {activeIcon}
            </span>
          </span>
        )}
        <span className={styles.button__text}>{text}</span>
        <FadeAnimation
          customWrapperClass={styles.button__count}
          isVisible={!!count}
        >
          <Typography text={`${count}`} />
        </FadeAnimation>
      </Link>
    );
  }

  return (
    <Element
      aria-label={text}
      className={cx(styles.button, className)}
      onClick={onClick}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
    >
      <span className={styles.button__iconWrapper}>
        <span
          className={cx(styles.button__icon, {
            [styles.visible as string]: !hovered,
          })}
        >
          {icon}
        </span>
        <span
          className={cx(styles.button__icon, {
            [styles.visible as string]: hovered,
          })}
        >
          {activeIcon}
        </span>
      </span>
      <span className={styles.button__text}>{text}</span>
      <FadeAnimation
        customWrapperClass={styles.button__count}
        isVisible={!!count}
      >
        <Typography text={`${count}`} />
      </FadeAnimation>
    </Element>
  );
};
