import { ReactNode } from 'react';
import cx from 'classnames';
import Link from 'next/link';

import { LazyImage } from '~/components/Lazy/LazyImage';
import { SwitchAnimation } from '~components/Animations/SwitchAnimation';
import { Typography } from '~components/Typography';
import ArrowRight from '~icons/arrowRight.svg';

import styles from './styles.module.scss';

interface Props {
  text: string;
  width?: string;
  onClick?: () => void;
  isDisabled?: boolean;
  className?: string;
  keyIcon?: string;
  icon?: ReactNode;
  beforeIcon?: ReactNode;
  href?: string;
}

export const SlidingButton = ({
  text,
  onClick = () => {},
  width,
  className,
  isDisabled,
  keyIcon = 'default',
  icon,
  beforeIcon,
  href,
}: Props) => {
  const cn = cx(
    styles['sliding-button'],
    {
      [`${styles['sliding-button-disabled']}`]: isDisabled,
    },
    className,
  );

  const content = (
    <>
      {beforeIcon && beforeIcon}
      <Typography
        text={text}
        isBungee
        textTransform={'uppercase'}
        customStyles={{
          width: 'max-content',
          whiteSpace: 'nowrap',
        }}
      />
      <div>
        <SwitchAnimation
          switchKey={keyIcon}
          customWrapperClass={styles['sliding-button__arrow']}
        >
          {icon ?? (
            <LazyImage
              loading="lazy"
              src={ArrowRight}
              width={32}
              alt={'Arrow image'}
              height={32}
            />
          )}
        </SwitchAnimation>
      </div>
    </>
  );

  if (href) {
    const isExternal =
      href.startsWith('http://') || href.startsWith('https://');

    return (
      <Link
        className={cn}
        style={{ width }}
        href={href}
        aria-label={text}
        {...(isExternal && {
          target: '_blank',
          rel: 'nofollow noopener noreferrer',
        })}
      >
        {content}
      </Link>
    );
  }

  return (
    <div className={cn} style={{ width }} onClick={onClick} aria-label={text}>
      {content}
    </div>
  );
};
