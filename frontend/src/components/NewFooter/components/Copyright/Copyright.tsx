import { useState } from 'react';
import Link from 'next/link';

import { Overdone } from '~icons/JsxSvg/Overdone';
import { OverdoneMonochrome } from '~icons/JsxSvg/OverdoneMonochrome';

import styles from './styles.module.scss';

export default function Copyright() {
  const [isDeveloperLinkHovered, setIsDeveloperLinkHovered] =
    useState<boolean>(false);

  return (
    <div
      className={styles.copyrights}
      onMouseEnter={() => setIsDeveloperLinkHovered(true)}
      onMouseLeave={() => setIsDeveloperLinkHovered(false)}
    >
      <p>{new Date().getFullYear()} © DUELDUCK CR LIMITADA</p>
      <Link
        href="https://overdone.it/"
        target="_blank"
        rel="nofollow noopener noreferrer"
      >
        Developed by{' '}
        {isDeveloperLinkHovered ? (
          <Overdone width={24} height={24} />
        ) : (
          <OverdoneMonochrome width={24} height={24} />
        )}
      </Link>
    </div>
  );
}
