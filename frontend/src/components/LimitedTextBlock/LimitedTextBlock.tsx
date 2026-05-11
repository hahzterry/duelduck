import { Typography } from '~components/Typography';

import styles from './styles.module.scss';

export const LimitedTextBlock = ({
  text,
  maxHeight,
}: {
  maxHeight?: string;
  text: string;
}) => {
  return (
    <Typography
      text={text}
      element={'pre'}
      className={styles.limitedTextBlock}
      customStyles={{
        maxHeight: maxHeight || '120px',
      }}
    />
  );
};

// <div
//       className={styles.limitedTextBlock}
//       style={{
//         maxHeight: maxHeight || '120px',
//       }}
//     >
//       <Typography text={text} />
//     </div>
