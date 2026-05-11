import React, { forwardRef, ReactNode, Ref } from 'react';
import cx from 'classnames';
import Link from 'next/link';

import { TypographyVariants } from '~components/Typography';

import styles from './styles.module.scss';

interface ButtonBaseProps {
  text: string | ReactNode;
  textVariant?: TypographyVariants;
  onClick?:
    | React.MouseEventHandler<HTMLAnchorElement>
    | React.MouseEventHandler<HTMLButtonElement>;
  className?: string;
  id?: string;
  onMouseEnter?: () => void;
  onMouseLeave?: () => void;
  isMainPage?: boolean;
  isDisabled?: boolean;
  typeButton?: 'button' | 'submit' | 'reset';
  isLink?: boolean;
  variant?: 'border' | 'close';
  href?: string;
  icon?: ReactNode;
  prevIcon?: ReactNode;
  style?: React.CSSProperties;
  target?: string;
  rel?: string;
}

export const ButtonBase = forwardRef<
  HTMLButtonElement | HTMLAnchorElement,
  ButtonBaseProps
>(
  (
    {
      text,
      textVariant = 'body',
      onClick = () => {},
      isMainPage,
      onMouseEnter,
      onMouseLeave,
      className,
      id,
      isDisabled,
      isLink,
      typeButton,
      href = '#',
      variant,
      icon,
      prevIcon,
      style,
      target,
      rel,
    },
    ref,
  ) => {
    const buttonClassName = cx(styles.buttonBase, textVariant, {
      [`${styles['buttonBase--disabled']}`]: isDisabled,
      [`${styles.mainPageButton}`]: isMainPage,
      [`${styles[`buttonBase--${variant}`]}`]: variant,
      [className as string]: !!className,
    });

    if (isLink) {
      const isExternal =
        href.startsWith('http://') || href.startsWith('https://');
      const computedTarget = target ?? (isExternal ? '_blank' : undefined);
      const computedRel =
        rel ?? (isExternal ? 'nofollow noopener noreferrer' : undefined);

      return (
        <Link
          ref={ref as Ref<HTMLAnchorElement>}
          href={isDisabled ? '#' : href}
          className={buttonClassName}
          id={id}
          target={computedTarget}
          rel={computedRel}
          onMouseEnter={onMouseEnter}
          onMouseLeave={onMouseLeave}
          onClick={(e) => {
            if (isDisabled) {
              e.preventDefault();
            } else {
              (onClick as React.MouseEventHandler<HTMLAnchorElement>)(e);
            }
          }}
          style={style}
        >
          {prevIcon && prevIcon}
          {text}
          {icon && icon}
        </Link>
      );
    }

    return (
      <button
        ref={ref as Ref<HTMLButtonElement>}
        onClick={onClick as React.MouseEventHandler<HTMLButtonElement>}
        id={id}
        onMouseEnter={onMouseEnter}
        type={typeButton}
        onMouseLeave={onMouseLeave}
        className={buttonClassName}
        style={style}
      >
        {prevIcon && prevIcon}
        {text}
        {icon && icon}
      </button>
    );
  },
);
