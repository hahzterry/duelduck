import cx from 'classnames';
import Link from 'next/link';

import { LazyImage } from '~/components/Lazy/LazyImage';
import { Typography } from '~components/Typography';
import { BottomRightArrow } from '~icons/JsxSvg/BottomRightArrow';

import styles from './styles.module.scss';

const bgDuck = '/bgDuck.webp';

export const RightCard = () => {
  return (
    <Link href={'/create-duel'} className={cx(styles.card)}>
      <div className={styles.card__logo}>
        <LazyImage
          src={bgDuck}
          width={280}
          height={200}
          alt={'Create your first duel'}
        />
      </div>
      <div>
        <Typography
          className={styles.card__text}
          text={'Create your first duel'}
        />
        <BottomRightArrow />
      </div>
    </Link>
  );
};
