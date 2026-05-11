import { CSSProperties } from 'react';
import cx from 'classnames';
import colors from 'src/styles/colors';

import styles from './styles.module.scss';

export const DoubleProgress = ({
  isNoAnimation,
  firstPercentProgress,
  showAnimation,
  firstColor = colors.success,
  lastColor = colors.error,
  borderRadius,
  styles: customStyles = {},
}: {
  firstPercentProgress: number;
  isNoAnimation?: boolean;
  showAnimation?: boolean;
  firstColor?: string;
  lastColor?: string;
  borderRadius?: string;
  styles?: CSSProperties;
}) => {
  return (
    <div
      className={cx(styles.linePercent, {
        [`${styles.animated}`]: !isNoAnimation && showAnimation,
      })}
      data-export={isNoAnimation}
      style={
        {
          ['--percent-no']: `${100 - firstPercentProgress}%`,
          ['--first-color']: lastColor,
          ['--second-color']: firstColor,
          ['--border-radius']: borderRadius,
          ...customStyles,
        } as CSSProperties
      }
    />
  );
};
