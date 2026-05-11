import { ReactNode } from 'react';
import cx from 'classnames';

import { DuckIcon } from '~icons/JsxSvg/DuckIcon';
import { Solana } from '~icons/JsxSvg/Solana';
import { Usdc } from '~icons/JsxSvg/usdc';
import { WbtIcon } from '~icons/JsxSvg/WbtIcon';

import styles from './styles.module.scss';

interface CoinIconProps {
  isUsdc?: boolean;
  isAll?: boolean;
  customIcon?: ReactNode;
  small?: boolean;
  isWbt?: boolean;
  isSol?: boolean;
  isFullWidthIcons?: boolean;
}

export const CoinTypeIcon = ({
  isUsdc,
  isWbt,
  isSol,
  small,
  customIcon,
  isAll,
  isFullWidthIcons,
}: CoinIconProps) => {
  if (isAll) {
    return (
      <div className={styles.wrapper}>
        <div
          className={cx(styles.coinIcon, {
            [`${styles.coinIcon__usdc}`]: false,
            [`${styles['coinIcon--fullWidth']}`]: isFullWidthIcons,
            [`${styles.small}`]: true,
          })}
        >
          <Usdc />
        </div>
        <div
          className={cx(styles.coinIcon, {
            [`${styles.small}`]: true,
            [`${styles['coinIcon--fullWidth']}`]: isFullWidthIcons,
          })}
        >
          <div className={styles.background} />
          <DuckIcon />
        </div>
      </div>
    );
  }

  return (
    <div
      className={cx(styles.coinIcon, {
        [`${styles['coinIcon--fullWidth']}`]: isFullWidthIcons,
        [`${styles.coinIcon__usdc}`]: isUsdc,
        [`${styles.small}`]: small,
      })}
    >
      {customIcon ??
        (isSol ? (
          <Solana />
        ) : isUsdc ? (
          <Usdc />
        ) : isWbt ? (
          <WbtIcon />
        ) : (
          <DuckIcon />
        ))}
    </div>
  );
};
