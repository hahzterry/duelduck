import { MouseEventHandler, ReactElement } from 'react';
import cx from 'classnames';
import Link from 'next/link';

import {
  fontWeightsClassesMap,
  Typography,
  TypographyVariants,
} from '~components/Typography';

import styles from './styles.module.scss';

export interface TextButtonProps {
  textVariant?: TypographyVariants;
  text: string;
  prefixIcon?: ReactElement;
  postfixIcon?: ReactElement;
  onClick?: MouseEventHandler<HTMLAnchorElement | HTMLDivElement>;
  activeColor?: 'primary' | 'secondary';
  isActive?: boolean;
  uppercase?: boolean;
  href?: string;
  target?: '_blank' | '_self' | '_parent' | '_top';
  rel?: string;
  className?: string;
  fontWeight?: keyof typeof fontWeightsClassesMap;
}

export const TextButton = ({
  textVariant = 'body',
  text,
  prefixIcon,
  onClick,
  activeColor,
  isActive,
  uppercase,
  href,
  fontWeight,
  target = '_self',
  postfixIcon,
  rel,
  className,
}: TextButtonProps) => {
  const isLink = Boolean(href);
  const commonProps = {
    className: cx(styles.textButton, {
      [styles.activeColorPrimary as string]: activeColor === 'primary',
      [styles.activePrimary as string]: isActive && activeColor === 'primary',
      [className as string]: !!className,
    }),
  };

  if (isLink) {
    return (
      <Link
        {...commonProps}
        href={href || ''}
        target={target}
        rel={rel || (target === '_blank' ? 'noopener noreferrer' : undefined)}
        onClick={onClick}
        aria-label={text}
      >
        {prefixIcon && prefixIcon}
        <Typography
          element="span"
          fontWeight={fontWeight}
          variant={textVariant}
          textTransform={uppercase ? 'uppercase' : 'none'}
          text={text}
        />
        {postfixIcon && postfixIcon}
      </Link>
    );
  }

  return (
    <div {...commonProps} onClick={onClick}>
      {prefixIcon && prefixIcon}
      <Typography
        element="span"
        variant={textVariant}
        fontWeight={fontWeight}
        textTransform={uppercase ? 'uppercase' : 'none'}
        text={text}
      />
      {postfixIcon && postfixIcon}
    </div>
  );
};
