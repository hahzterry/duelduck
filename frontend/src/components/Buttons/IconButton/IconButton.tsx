import { CSSProperties, MouseEventHandler, ReactNode, useState } from 'react';
import cx from 'classnames';
import Link from 'next/link';

import { SwitchAnimation } from '~components/Animations/SwitchAnimation';

import styles from './styles.module.scss';

interface IconButtonProps {
  children: ReactNode;
  keyAnimation?: string;
  variant?: 'primary' | 'secondary';
  onClick?:
    | MouseEventHandler<HTMLDivElement>
    | MouseEventHandler<HTMLAnchorElement>;
  className?: string;
  bordered?: boolean;
  ariaLabel: string;
  hoverContent?: ReactNode;
  style?: CSSProperties;
  href?: string;
  isCantHover?: boolean;
}

export const IconButton = ({
  children,
  variant = 'primary',
  onClick = () => {},
  className,
  bordered,
  keyAnimation = 'state',
  ariaLabel,
  style,
  href,
  hoverContent,
  isCantHover = true,
}: IconButtonProps) => {
  const [isHovered, setIsHovered] = useState(false);

  const cn = cx(styles.iconButton, {
    [styles.secondary as string]: variant === 'secondary',
    [className as string]: !!className,
    [styles.borderd as string]: bordered,
  });

  const hovered = !!(hoverContent && isHovered);

  const content = (
    <SwitchAnimation
      customWrapperClass={styles.iconButton__elements}
      switchKey={`${keyAnimation}`}
    >
      <div data-hovered={!hovered}>{children}</div>
      <div data-hovered={hovered}>{hoverContent}</div>
    </SwitchAnimation>
  );

  const setHovered = (isHover: boolean) => {
    if (isCantHover) {
      setIsHovered(isHover);
    }
  };

  if (href) {
    return (
      <Link
        aria-label={ariaLabel}
        onMouseEnter={() => setHovered(true)}
        onMouseMove={() => setHovered(true)}
        onMouseLeave={() => setHovered(false)}
        style={style}
        href={href}
        onClick={onClick as MouseEventHandler<HTMLAnchorElement>}
        className={cn}
      >
        {content}
      </Link>
    );
  }

  return (
    <div
      aria-label={ariaLabel}
      onMouseEnter={() => setHovered(true)}
      onMouseMove={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      style={style}
      onClick={onClick as MouseEventHandler<HTMLDivElement>}
      className={cn}
    >
      {content}
    </div>
  );
};
