import Link from 'next/link';

import { LazyImage } from '~/components/Lazy/LazyImage';

import styles from './styles.module.scss';

interface Props {
  variant?: 'footer' | 'compact';
}

export const ResponsibleGamingBanner = ({ variant = 'footer' }: Props) => {
  return (
    <div
      className={
        variant === 'compact' ? styles.wrapperCompact : styles.wrapperFooter
      }
    >
      <div className={styles.content}>
        <LazyImage
          src="/duck-18-plus.webp"
          alt="18+ platform"
          width={variant === 'compact' ? 56 : 40}
          height={variant === 'compact' ? 56 : 40}
          className={
            variant === 'compact' ? styles.imageCompact : styles.imageFooter
          }
        />
        {variant === 'compact' ? (
          <span className={styles.textCompact}>
            This platform is intended for 18+ users. Please follow{' '}
            <Link
              href="/responsible-gaming-policy"
              className={styles.linkCompact}
            >
              Responsible Gaming Policy
            </Link>
          </span>
        ) : (
          <>
            <span className={styles.textFooter}>
              This platform is intended for 18+ users:
            </span>
            <Link
              href="/responsible-gaming-policy"
              className={styles.linkFooter}
            >
              Responsible Gaming Policy
            </Link>
          </>
        )}
      </div>
    </div>
  );
};
