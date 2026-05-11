import { Typography } from '~components/Typography';

import styles from './styles.module.scss';

interface BadgeListProps {
  title: string;
  badges: string[];
}
export const BadgeList = ({ title, badges }: BadgeListProps) => {
  return (
    <div className={styles.badgeList}>
      <Typography element="h4" className={styles['badgeList__title']}>
        {title}:
      </Typography>
      <div className={styles['badgeList__list']}>
        {badges.map((badge, index) => (
          <div className={styles['badgeList__item']} key={index}>
            <Typography>{badge}</Typography>
          </div>
        ))}
      </div>
    </div>
  );
};
