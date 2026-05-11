import './Counter.scss';

import { memo, useEffect, useState } from 'react';
import cx from 'classnames';

import { SlotCounter } from '~components/Slot';
import { CoinIcon } from '~icons/JsxSvg/CoinIcon';
import { DuckIcon } from '~icons/JsxSvg/DuckIcon';
import { TokenDisabledIcon } from '~icons/JsxSvg/TokenDisabledIcon';
import { formatNumber } from '~utils/numbers';

export type CounterColor = 'secondary' | 'dark' | 'primary';

interface CounterProps {
  value: number | string;
  isBase?: boolean;
  color?: CounterColor;
  backgroundColor?: string;
  size?:
    | 'default'
    | 'medium'
    | 'large'
    | 'isDuelPageCardSize'
    | 'isMainPageStatsSize'
    | 'stats';
  hasNoCoinIcon?: boolean;
  isDuckCounter?: boolean;
  isGreyDuckCounter?: boolean;
  withoutEmptySlots?: boolean;
  notAnimatedCounter?: boolean;
  isRepeat?: boolean;
}

export const Counter = memo(
  ({
    value = 0,
    isBase,
    color,
    backgroundColor,
    size = 'default',
    hasNoCoinIcon,
    isDuckCounter,
    isGreyDuckCounter,
    withoutEmptySlots,
    notAnimatedCounter,
    isRepeat,
  }: CounterProps) => {
    const [isVisible, setIsVisible] = useState(false);

    let stringValueArr = value.toString().split('');

    const isLong = stringValueArr.length > 6;

    if (isLong) {
      // eslint-disable-next-line @typescript-eslint/ban-ts-comment
      // @ts-ignore
      const valueFormatted = Number(value);

      stringValueArr = formatNumber(valueFormatted, undefined, 1e6)
        .toString()
        .split('');
    }

    let addedZeroes = 0;

    if (stringValueArr.length < 6) {
      const diff = 6 - stringValueArr.length;

      addedZeroes = diff;
      stringValueArr = [...Array(diff).fill(0), ...stringValueArr];
    }

    useEffect(() => {
      setIsVisible(false);
      setTimeout(() => {
        setIsVisible(true);
      }, 60);
    }, []);

    if (withoutEmptySlots) {
      stringValueArr = value.toString().split('');
    }

    return (
      <div
        className={cx('counter', {
          'counter-win': +value > 0 && !isBase,
          'counter-lose': +value < 0 && !isBase,
          'counter-colorSecondary': color === 'secondary',
          'counter-colorDark': color === 'dark',
          'counter-colorPrimary': color === 'primary',
          'counter-isLargeSize': size == 'large',
          'counter-isMediumSize': size == 'medium',
          'counter-stats': size == 'stats',
          'counter-isDuelPageCardSize': size == 'isDuelPageCardSize',
          'counter-isMainPageStatsSize': size === 'isMainPageStatsSize',
        })}
      >
        {!hasNoCoinIcon && (
          <div style={{ backgroundColor }}>
            {isDuckCounter ? (
              isGreyDuckCounter ? (
                <TokenDisabledIcon />
              ) : (
                <DuckIcon />
              )
            ) : (
              <CoinIcon />
            )}
          </div>
        )}
        {stringValueArr.map((val, index) => (
          <div
            key={index}
            style={{ backgroundColor }}
            className={cx({
              'counter-added-zero': addedZeroes - 1 >= index,
            })}
          >
            {isVisible && (
              <>
                {notAnimatedCounter ? (
                  <span>{val}</span>
                ) : (
                  <SlotCounter text={val} triggerOnce={isRepeat} />
                )}
              </>
            )}
          </div>
        ))}
      </div>
    );
  },
);
