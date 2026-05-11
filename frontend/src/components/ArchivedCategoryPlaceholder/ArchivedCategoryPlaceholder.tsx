'use client';

import Link from 'next/link';

import { Typography } from '~components/Typography';

import styles from './styles.module.scss';

interface Props {
  categoryTitle: string;
  parentSlug?: string;
}

export const ArchivedCategoryPlaceholder = ({
  categoryTitle,
  parentSlug,
}: Props) => {
  return (
    <div className={styles.archived}>
      <Typography element="h1" className={styles.archived__title}>
        This category has been archived
      </Typography>
      <Typography element="p" className={styles.archived__text}>
        The &quot;{categoryTitle}&quot; category is no longer active. Browse
        other prediction markets on DuelDuck.
      </Typography>
      <div className={styles.archived__links}>
        {parentSlug && (
          <Link href={`/duels/${parentSlug}`} className={styles.archived__link}>
            Back to {parentSlug}
          </Link>
        )}
        <Link href="/" className={styles.archived__link}>
          Go to Home
        </Link>
        <Link href="/duels" className={styles.archived__link}>
          Browse All Duels
        </Link>
      </div>
    </div>
  );
};
