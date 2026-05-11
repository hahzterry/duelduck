import { CSSProperties, MouseEventHandler, ReactElement } from 'react';
import cx from 'classnames';
import Link from 'next/link';

import { Typography, TypographyVariants } from '~components/Typography';
import colors from '~styles/colors';

import styles from './styles.module.scss';

interface ButtonWithIconsProps {
  children?: string;
  startIcon?: ReactElement;
  endIcon?: ReactElement;
  textVariant?: TypographyVariants;
  href?: string;
  target?: string;
  rel?: string;
  prefetch?: boolean;
  onClick?: MouseEventHandler<any>;
  color?: string;
  hoverColor?: string;

  isBungee?: boolean;
  backgroundColor?: string;
  width?: string;
  className?: string;
  disabled?: boolean;
}

export const ButtonWithIcons = ({
  children,
  startIcon,
  textVariant,
  endIcon,
  onClick,
  href,
  target,
  rel,
  prefetch,
  isBungee = false,
  backgroundColor,
  width,
  className,
  color = colors['textSecondary'],
  hoverColor = colors['textPrimary'],
  disabled,
}: ButtonWithIconsProps) => {
  const commonProps = {
    className: cx(styles.buttonWithIcons, className),
    style: { backgroundColor, width },
    onClick: (e: any) => {
      if (!disabled) {
        onClick?.(e);
      }
    },
  };

  const computedRel =
    target === '_blank' ? (rel ?? 'noopener noreferrer') : rel;

  if (href) {
    return (
      <Link
        href={href}
        {...commonProps}
        target={target}
        rel={computedRel}
        style={
          {
            ['--hover-color']: hoverColor,
            ['--color']: color,
            opacity: disabled ? 0.7 : undefined,
            cursor: disabled ? 'not-allowed' : undefined,
          } as CSSProperties
        }
        prefetch={prefetch}
      >
        <div className={styles.buttonWithIcons__iconAndChildren}>
          {startIcon}
          <Typography
            text={children as string}
            variant={textVariant}
            isBungee={isBungee}
          />
        </div>
        {endIcon}
      </Link>
    );
  }

  return (
    <button
      {...commonProps}
      type="button"
      style={
        { ['--hover-color']: hoverColor, ['--color']: color } as CSSProperties
      }
    >
      <div className={styles.buttonWithIcons__iconAndChildren}>
        {startIcon}
        <Typography
          text={children as string}
          variant={textVariant}
          isBungee={isBungee}
        />
      </div>
      {endIcon}
    </button>
  );
};
