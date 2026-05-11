import cx from 'classnames';
import { Property } from 'csstype';
import colors from 'src/styles/colors';

import styles from './styles.module.scss';

interface Props {
  colorSpin?: Property.BorderColor;
  className?: string;
}

export const SpinAnimation = ({
  colorSpin = colors.primary,
  className,
}: Props) => {
  return (
    <div className={cx(styles.spinElement__container, className)}>
      <div
        className={styles.spinElement}
        style={{ borderTopColor: colorSpin }}
      />
    </div>
  );
};
