import { ChangeEvent, ReactNode, useEffect, useRef, useState } from 'react';
import cx from 'classnames';

import { Typography } from '~components/Typography';
import useMediaQuery from '~hooks/useMediaQuery';
import { ArrowRight } from '~icons/JsxSvg/ArrowRight';
import colors from '~styles/colors';

import { InputWithTail } from './InputWithTail';
import styles from './styles.module.scss';

interface RangeWithProgressProps {
  min: number;
  max: number;
  step: number;
  valueRange: number;
  onChange?: (value: number) => void;
  currencyIcon?: ReactNode | string;
  currency?: string;
  valueUnlockPremium?: number;
  maxValue?: number;
  isDouble?: boolean;
  valueRangeTwo?: number;
  isFloatNumbers?: boolean;
  isCurrencyBeforeValue?: boolean;
  isHaveMinMaxValues?: boolean;
  isThin?: boolean;
  currencySecond?: string;
  onChangeDoubleValue?: (valueFirst: number, valueSecond: number) => void;
  haveCurrency?: boolean;
  isResizeNoDouble?: boolean;
  isOnlyText?: boolean;
  haveTail?: boolean;
  havePercent?: boolean;
}

export const RangeWithProgress = ({
  min,
  max,
  step,
  valueRange = min,
  onChange,
  valueUnlockPremium,
  currency = '%',
  isFloatNumbers,
  currencySecond,
  currencyIcon = '%',
  maxValue = 1000000000,
  isDouble,
  haveTail = true,
  valueRangeTwo,
  isCurrencyBeforeValue,
  isHaveMinMaxValues = true,
  onChangeDoubleValue,
  isThin,
  haveCurrency,
  isResizeNoDouble,
  isOnlyText,
  havePercent,
}: RangeWithProgressProps) => {
  const { isTablet, isPhone } = useMediaQuery();
  const [isInputFocused, setIsInputFocused] = useState(false);

  const [tailValueTwo, setTailValueTwo] = useState(valueRangeTwo);
  const [tailValue, setTailValue] = useState(valueRange);

  useEffect(() => {
    if (valueRangeTwo !== tailValueTwo) setTailValueTwo(valueRangeTwo);
    if (valueRange !== tailValue) setTailValue(valueRange);
  }, [valueRangeTwo, valueRange]);

  const myRef = useRef<HTMLDivElement | null>(null);
  const [width, setWidth] = useState(0);
  const [distanceBetweenDashes, setDistanceBetweenDashes] = useState(11);
  const [webkitSliderThumb, setWebkitSliderThumb] = useState(
    isPhone ? 32 : isTablet && !isPhone ? 56 : 32,
  );
  const [buttonUnlockPremiumWidth, setButtonUnlockPremiumWidth] = useState(150);

  useEffect(() => {
    const updateDimensions = () => {
      if (myRef.current) {
        const currentWidth = myRef.current.offsetWidth;

        setWidth(currentWidth);

        setDistanceBetweenDashes(11);
        setWebkitSliderThumb(isPhone ? 32 : isTablet && !isPhone ? 56 : 32);
        setButtonUnlockPremiumWidth(150);
      }
    };

    const resizeObserver = new ResizeObserver(() => updateDimensions());

    if (myRef.current) {
      resizeObserver.observe(myRef.current);
      updateDimensions(); // Початкове оновлення
    }

    return () => {
      if (myRef.current) {
        resizeObserver.unobserve(myRef.current);
      }
    };
  }, []);

  const handleFocus = () => {
    setIsInputFocused(true);
  };

  const handleBlur = () => {
    setIsInputFocused(false);
  };

  const setNewValue = (newValue: string) => {
    if (/^\d*$/.test(newValue) || isFloatNumbers) {
      const numericValue = +newValue;

      const setRange = (value: number) => {
        onChange && onChange(value);
        if (isDouble && onChangeDoubleValue) {
          onChangeDoubleValue(valueRangeTwo!, value);
        }
      };

      if (valueUnlockPremium) {
        setRange(
          numericValue <= min
            ? min
            : Math.min(numericValue, valueUnlockPremium || max),
        );
      } else {
        if (isDouble) {
          setRange(numericValue <= min ? min : Math.min(numericValue, max));
        } else {
          setRange(
            numericValue <= min ? min : Math.min(numericValue, maxValue),
          );
        }
      }
    }
  };

  const handleInputTailTwo = (e: ChangeEvent<HTMLInputElement>) => {
    const value = +e.target.value;

    if (Number.isNaN(value)) return;

    setTailValueTwo(value);

    onChangeDoubleValue &&
      onChangeDoubleValue(
        value < valueRange ? (value < min ? min : value) : valueRange,
        valueRange,
      );
  };

  const handleInputTail = (e: ChangeEvent<HTMLInputElement>) => {
    const value = +e.target.value;

    if (Number.isNaN(value)) return;

    setTailValue(value);
    setNewValue(`${value < (valueRangeTwo || 0) ? valueRangeTwo || 0 : value}`);
  };

  const handleInputValueRange = (event: ChangeEvent<HTMLInputElement>) => {
    const value = event.target.value;

    if (isDouble) {
      if (+value > valueRangeTwo!) {
        setTailValue(+value);
        setNewValue(value);
      }
    } else {
      haveTail && setTailValue(+value);
      setNewValue(value);
    }
  };

  const handleInputValueRangeTwo = (event: ChangeEvent<HTMLInputElement>) => {
    const value = +event.target.value;

    if (value < valueRange! && value >= min) {
      setTailValueTwo(value);
      onChangeDoubleValue && onChangeDoubleValue(value!, valueRange);
    }
  };

  const yellowProgressWidthFormula = () => {
    const range = (max - min) / step;
    const widthPerStep = width / range;

    const calculateValue = (value: number) => {
      return widthPerStep * (value - min);
    };

    if (isDouble) {
      const valueOne = calculateValue(valueRange);
      const valueTwo = calculateValue(valueRangeTwo!);
      const newWidth = valueOne - valueTwo;

      return `${Math.max(newWidth, 0)}px`;
    }

    const adjustedValueRange = valueRange >= max ? max : valueRange;

    const countSteps = (adjustedValueRange - min) / step;

    return (
      widthPerStep * countSteps +
      (webkitSliderThumb / 2) * (1 - countSteps / range)
    );
  };

  const yellowProgressLeftFormula = () => {
    if (isDouble) {
      const percentage = (valueRangeTwo! - min) / (max - min);
      const offset = -5 * percentage;

      return `${(width / ((max - min) / step)) * (valueRangeTwo! - min) + offset}px`;
    }

    return valueRange === 0 ? `${2}px` : `${0}px`;
  };

  const inputWithTailLeftFormula = () => {
    const calculateSliderPosition = (value: number) => {
      const scaleFactor = width / ((max - min) / step);
      const thumbOffset = (webkitSliderThumb / ((max - min) / step)) * value;
      const minOffset = scaleFactor * min;

      return scaleFactor * value + webkitSliderThumb - thumbOffset - minOffset;
    };

    if (isDouble) {
      const positionOne = calculateSliderPosition(valueRange);
      const positionTwo = calculateSliderPosition(valueRangeTwo!);
      const difference = positionOne - positionTwo;

      return (
        ((step * (width - webkitSliderThumb)) / (max - min)) *
          (valueRange - min) +
        webkitSliderThumb / 2 -
        difference / 2
      );
    }

    const range = (max - min) / step;
    const widthPerStep = width / range;

    const adjustedValueRange = valueRange >= max ? max : valueRange;
    const countSteps = (adjustedValueRange - min) / step;

    return widthPerStep * countSteps;
  };

  const inputWithTailTranslateFormula = () => {
    const range = (max - min) / step;

    const adjustedValueRange = valueRange >= max ? max : valueRange;
    const countSteps = (adjustedValueRange - min) / step;

    return `translateX(calc(-50% + ${webkitSliderThumb / 2 - webkitSliderThumb * (countSteps / range)}px))`;
  };

  return (
    <div className={styles.rangeWithProgress}>
      {haveTail && (
        <InputWithTail
          currencyIcon={currencyIcon}
          handleBlur={handleBlur}
          handleFocus={handleFocus}
          isResizeNoDouble={isResizeNoDouble}
          isInputFocused={isInputFocused}
          handleInputValueRange={handleInputTail}
          valueTwo={tailValueTwo}
          handleInputValueRangeTwo={handleInputTailTwo}
          value={tailValue}
          isDouble={isDouble}
          left={inputWithTailLeftFormula()}
          transformXValue={inputWithTailTranslateFormula()}
          isThin={isThin}
          haveCurrency={haveCurrency}
          isOnlyText={isOnlyText}
          havePercent={havePercent}
        />
      )}
      <div
        className={cx(styles.rangeWithProgress__range, {
          [`${styles['rangeWithProgress__range--isThin']}`]: isThin,
        })}
        ref={myRef}
      >
        <div className={styles.rangeWithProgress__yellowLineContainer}>
          <div
            className={styles.rangeWithProgress__yellowProgress}
            style={{
              width: yellowProgressWidthFormula(),
              left: yellowProgressLeftFormula(),
            }}
          />
        </div>
        <div
          className={cx(styles.rangeWithProgress__inputsRange, {
            [`${styles['rangeWithProgress__inputsRange--isDouble']}`]: isDouble,
          })}
        >
          {isDouble && (
            <input
              type="range"
              min={min}
              max={max}
              step={step}
              value={valueRangeTwo}
              onChange={handleInputValueRangeTwo}
            />
          )}
          <input
            type="range"
            min={min}
            max={max}
            step={step}
            value={valueRange}
            onChange={handleInputValueRange}
          />
        </div>
        <div
          className={cx(styles.rangeWithProgress__buttonUnlockPremium, {
            [`${styles['rangeWithProgress__buttonUnlockPremium--visible']}`]:
              valueUnlockPremium && valueRange === valueUnlockPremium,
            [`${styles['rangeWithProgress__buttonUnlockPremium--noVisible']}`]:
              !(valueUnlockPremium && valueRange === valueUnlockPremium),
          })}
          style={{
            right: `${
              (width / 2) * (1 - (step * (valueRange - min)) / (max - min)) -
              buttonUnlockPremiumWidth / 2
            }px`,
          }}
        >
          <Typography
            text="Unlock Premium"
            color={colors['primary']}
            variant="body-small"
            lineHeight={1.4}
            fontWeight="sb"
          />
          <ArrowRight />
        </div>
        {!isThin && (
          <div className={styles.rangeWithProgress__containerVerticalLine}>
            {Array.from(
              { length: Math.floor((width + 17) / distanceBetweenDashes) },
              (_, i) => (
                <div
                  className={cx(styles.rangeWithProgress__verticalLine)}
                  key={i}
                  style={{ left: distanceBetweenDashes * i - 18 }}
                />
              ),
            )}
          </div>
        )}
      </div>
      {isHaveMinMaxValues && (
        <>
          <div className={styles.rangeWithProgress__maxMinContainer}>
            <span>
              {isCurrencyBeforeValue && (
                <Typography
                  text={`${currency?.toUpperCase()} `}
                  variant={'body-big'}
                  lineHeight={1.4}
                  fontWeight="sb"
                  color={colors['textSecondary']}
                />
              )}
              <Typography
                text={`${min} `}
                variant={'body-big'}
                lineHeight={1.4}
                fontWeight="sb"
                color={colors['textSecondary']}
              />
              {!isCurrencyBeforeValue && (
                <Typography
                  text={`${currency?.toUpperCase()} `}
                  variant={'body-big'}
                  lineHeight={1.4}
                  fontWeight="sb"
                  color={colors['textSecondary']}
                />
              )}
            </span>
            <span>
              {isCurrencyBeforeValue && (
                <Typography
                  text={`${currencySecond?.toUpperCase() || currency?.toUpperCase()} `}
                  variant={'body-big'}
                  lineHeight={1.4}
                  fontWeight="sb"
                  color={colors['textSecondary']}
                />
              )}
              <Typography
                text={`${max} `}
                variant="body-big"
                lineHeight={1.4}
                fontWeight="sb"
                color={colors['textSecondary']}
              />
              {!isCurrencyBeforeValue && (
                <Typography
                  text={`${currencySecond?.toUpperCase() || currency?.toUpperCase()} `}
                  variant={'body-big'}
                  lineHeight={1.4}
                  fontWeight="sb"
                  color={colors['textSecondary']}
                />
              )}
            </span>
          </div>
        </>
      )}
    </div>
  );
};
