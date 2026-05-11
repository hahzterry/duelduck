import { parseHtmlToChildren } from '~components/seoComponents/utils';
import { Typography } from '~components/Typography';

import styles from './styles.module.scss';

interface BulletListProps {
  items: string[];
  className?: string;
}

export const BulletList = ({ items, className }: BulletListProps) => {
  return (
    <ul className={`${styles.bulletList} ${className ?? ''}`}>
      {items.map((item, i) => (
        <li key={i} className={styles.bulletItem}>
          <Typography element="span" className={`${styles.p2custom} p2`}>
            {parseHtmlToChildren(item)}
          </Typography>
        </li>
      ))}
    </ul>
  );
};
