import { CSSProperties, memo } from 'react';
import cx from 'classnames';
import Link from 'next/link';

import { ErrorToastTextPart } from '~store/errorToastStore';

import styles from './styles.module.scss';

export const TextToast = memo(({ parts }: { parts?: ErrorToastTextPart[] }) => {
  if (parts?.length) {
    return parts.map((part, index) => {
      const partStyle = {
        cursor: part.href || part.onClick ? 'pointer' : undefined,
        fontSize: part.fontSize,
        '--hover-color': part?.hoverColor || part.color,
        '--color': part.color,
      } as CSSProperties;
      const normalizedText = part.text.replace(/\\n/g, '\n');

      if (part.href) {
        return (
          <Link
            key={`${part.text}-${index}`}
            href={part.href}
            className={cx(styles.textToast, styles.textToast__link)}
            style={partStyle}
            onClick={() => part.onClick?.()}
          >
            {normalizedText}
          </Link>
        );
      }

      return (
        <span
          key={`${part.text}-${index}`}
          className={styles.textToast}
          style={partStyle}
          onClick={() => part.onClick?.()}
        >
          {normalizedText}
        </span>
      );
    });
  }

  return null;
});
