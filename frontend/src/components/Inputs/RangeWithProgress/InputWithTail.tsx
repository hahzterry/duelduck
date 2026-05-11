import { ChangeEvent, forwardRef, memo, ReactNode, useRef } from 'react';
import cx from 'classnames';

import { Typography } from '~components/Typography';
import colors from '~styles/colors';

import styles from './styles.module.scss';

interface RangeWithProgressProps {
  left: number;
  currencyIcon?: ReactNode | string;
  handleInputValueRange: (event: ChangeEvent<HTMLInputElement>) => void;
  value: number;
  handleFocus: () => void;
  handleBlur: () => void;
  isInputFocused: boolean;
  isDouble?: boolean;
  handleInputValueRangeTwo?: (event: ChangeEvent<HTMLInputElement>) => void;
  valueTwo?: number;
  isThin?: boolean;
  isResizeNoDouble?: boolean;
  haveCurrency?: boolean;
  transformXValue: string;
  isOnlyText?: boolean;
  havePercent?: boolean;
}

export const InputWithTail = memo(
  forwardRef<HTMLDivElement, RangeWithProgressProps>(
    (
      {
        left,
        currencyIcon,
        value,
        handleInputValueRange,
        handleInputValueRangeTwo,
        handleBlur,
        handleFocus,
        isInputFocused,
        isDouble,
        valueTwo,
        isThin,
        haveCurrency = true,
        isResizeNoDouble,
        transformXValue,
        isOnlyText,
        havePercent,
      }: RangeWithProgressProps,
      ref,
    ) => {
      const inputRef = useRef<HTMLInputElement>(null);
      const handleFocusInput = () => {
        if (inputRef.current) {
          inputRef.current.focus();
        }
      };

      return (
        <div
          className={cx({
            [`${styles['rangeWithProgress__progress']}`]: true,
            [`${styles['rangeWithProgress__progress--isThin']}`]: isThin,
            [`${styles['rangeWithProgress__progress--isOnlyText']}`]:
              isOnlyText,
          })}
          style={{
            left: `calc(${left}px`,
            transform: transformXValue,
          }}
          ref={ref}
        >
          {isOnlyText ? (
            <Typography
              text={`${value}${havePercent ? '%' : ''}`}
              className={styles['rangeWithProgress__progressText']}
            />
          ) : (
            <>
              <div
                className={cx({
                  [`${styles['rangeWithProgress__inputProgress']}`]: true,
                  [`${styles['rangeWithProgress__inputProgress--isInputFocused']}`]:
                    isInputFocused,
                  [`${styles['rangeWithProgress__inputProgress--isDouble']}`]:
                    isDouble,
                  [`${styles['rangeWithProgress__inputProgress--isResizeNoDouble']}`]:
                    isResizeNoDouble,
                })}
                onClick={handleFocusInput}
              >
                <div
                  className={cx(styles.rangeWithProgress__containerCurrency, {
                    [`${styles['rangeWithProgress__containerCurrency--isDouble']}`]:
                      isDouble,
                  })}
                  style={{ display: !haveCurrency ? 'none' : '' }}
                >
                  {typeof currencyIcon === 'string' ? (
                    <Typography
                      text="%"
                      color={colors['primary']}
                      variant="body-small"
                      lineHeight={1.4}
                    />
                  ) : (
                    currencyIcon
                  )}
                </div>
                {isDouble ? (
                  <div
                    className={styles.rangeWithProgress__containerDoubleInput}
                  >
                    <input
                      onFocus={handleFocus}
                      onBlur={handleBlur}
                      type="text"
                      style={{ width: `${valueTwo?.toString().length}ch` }}
                      value={valueTwo}
                      inputMode={'decimal'}
                      pattern={'^[0-9]*[.,]?[0-9]*$'}
                      onChange={handleInputValueRangeTwo}
                    />
                    <Typography
                      text="-"
                      variant="body-small"
                      lineHeight={1.4}
                      color={colors['textSecondary']}
                    />
                    <input
                      onFocus={handleFocus}
                      onBlur={handleBlur}
                      type="text"
                      style={{ width: `${value?.toString().length}ch` }}
                      value={value}
                      onChange={handleInputValueRange}
                    />
                  </div>
                ) : (
                  <input
                    onFocus={handleFocus}
                    onBlur={handleBlur}
                    ref={inputRef}
                    type="text"
                    value={value}
                    style={{
                      width: isResizeNoDouble
                        ? `${value?.toString().length}ch`
                        : undefined,
                    }}
                    onChange={handleInputValueRange}
                  />
                )}
              </div>
              <div
                className={cx({
                  [`${styles.rangeWithProgress__containerTail}`]: true,
                  [`${styles['rangeWithProgress__containerTail--isInputFocused']}`]:
                    isInputFocused,
                })}
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="16"
                  height="4"
                  viewBox="0 0 16 4"
                  fill="none"
                >
                  <path
                    d="M7.47555 3.59615L5.90683 2.04715C5.1417 1.29163 4.75913 0.913871 4.31267 0.643723C3.91685 0.404211 3.48531 0.227709 3.0339 0.120699C2.52476 0 1.98372 0 0.901656 0H15.6136C14.5315 0 13.9905 0 13.4813 0.120699C13.0299 0.227709 12.5984 0.404211 12.2026 0.643723C11.7561 0.913871 11.3735 1.29163 10.6084 2.04715L9.03967 3.59615C8.60775 4.02264 7.90747 4.02264 7.47555 3.59615Z"
                    fill="currentColor"
                  />
                </svg>
              </div>
            </>
          )}
        </div>
      );
    },
  ),
);
