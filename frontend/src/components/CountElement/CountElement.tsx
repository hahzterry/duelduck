import { Typography } from '~components/Typography';

import styles from './styles.module.scss';

interface Props {
  count: number;
}

export const CountElement = ({ count }: Props) => {
  return (
    <div className={styles.option__count}>
      <Typography text={`${count}`} />
    </div>
  );
};
