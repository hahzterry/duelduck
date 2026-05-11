import { Typography } from '~components/Typography';
import { Styles } from '~types/general';

import styles from './styles.module.scss';

export const TextInBlock = ({
  text,
  bgColor,
  color,
}: {
  text: string;
  bgColor?: string;
  color?: string;
}) => {
  return (
    <Typography
      text={text}
      className={styles.textInBlock}
      customStyles={
        {
          ['--background-color']: bgColor,
          ['--color']: color,
        } as Styles
      }
    />
  );
};
