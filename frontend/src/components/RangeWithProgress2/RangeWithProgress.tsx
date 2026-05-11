import './styles.scss';

import { ChangeEvent, ReactNode, useEffect, useRef, useState } from 'react';
import cx from 'classnames';

import { LazyImage } from '~/components/Lazy/LazyImage';
import { Typography } from '~components/Typography';
import arrow from '~icons/arrowRightActive.svg';
import colors from '~styles/colors';

import { InputWithTail } from './InputWithTail';

interface Props {
  min: number;
  max: number;
  step: number;
  valueRange: number;
  onChangeValue?: (value: number) => void;
  currencyIcon?: ReactNode | string;
  valueUnlockPremium?: number;

  maxValue?: number;
  isDouble?: boolean;
  valueRangeTwo?: number;
  isHaveMinMaxValues?: boolean;
  isThin?: boolean;
  onChangeDoubleValue?: (valueFirst: number, valueSecond: number) => void;
  isResizeNoDouble?: boolean;
  havePercent?: boolean;
  textBetweenMaxMin?: string;
}

export const RangeWithProgress2 = ({
  min,
  max,
  step,
  valueRange = min,
  onChangeValue,
  valueUnlockPremium,
  currencyIcon = '%',
  maxValue = 1000000000,
  isDouble,
  valueRangeTwo,
  isHaveMinMaxValues = true,
  onChangeDoubleValue,
  isThin,
  isResizeNoDouble,
  havePercent,
  textBetweenMaxMin,
}: Props) => {
  const [isInputFocused, setIsInputFocused] = useState(false);

  const DISTANCE_BETWEEN_DASHES = 11;
  const WEBKIT_SLIDER_THUMB = 32;
  const BUTTON_UNLOCK_PREMIUM_WIDTH = 150;
  const refTail = useRef<HTMLDivElement | null>(null);
  const myRef = useRef<HTMLDivElement | null>(null);
  const [width, setWidth] = useState(0);

  useEffect(() => {
    setTimeout(() => {
      setWidth(myRef.current?.offsetWidth || 0);
    }, 50);
  }, []);

  const handleFocus = () => {
    setIsInputFocused(true);
  };

  const handleBlur = () => {
    setIsInputFocused(false);
  };

  const setNewValue = (newValue: string) => {
    if (/^\d*$/.test(newValue)) {
      const numericValue = Number(newValue);

      const setRange = (value: number) => {
        onChangeValue && onChangeValue(value);
        if (isDouble && onChangeDoubleValue)
          onChangeDoubleValue(valueRangeTwo!, value);
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

  const handleInputValueRange = (event: ChangeEvent<HTMLInputElement>) => {
    if (isDouble) {
      if (Number(event.target.value) > valueRangeTwo!)
        setNewValue(event.target.value);
    } else setNewValue(event.target.value);
  };

  const handleInputValueRangeTwo = (event: ChangeEvent<HTMLInputElement>) => {
    const value = Number(event.target.value);

    if (value < valueRange! && value >= min) {
      onChangeDoubleValue && onChangeDoubleValue(value!, valueRange);
    }
  };

  const yellowProgressWidthFormula = () => {
    const range = (max - min) / step;
    const widthPerStep = (width - WEBKIT_SLIDER_THUMB / 2) / range;

    const calculateValue = (value: number) => {
      return widthPerStep * (value - min);
    };

    if (isDouble) {
      const valueOne = calculateValue(valueRange);
      const valueTwo = calculateValue(valueRangeTwo!);

      return `${valueOne - valueTwo}px`;
    }

    const adjustedValueRange = valueRange >= max ? max : valueRange;
    const countSteps = adjustedValueRange / step;

    return widthPerStep * countSteps;
  };

  const yellowProgressLeftFormula = () => {
    if (isDouble) {
      const percentage = (valueRangeTwo! - min) / (max - min);
      const offset = -5 * percentage;

      return `${(width / ((max - min) / step)) * (valueRangeTwo! - min) + offset}px`;
    }

    const adjustedValueRange = valueRange >= max ? max : valueRange;
    const countSteps = adjustedValueRange / step;

    return countSteps <= 1 ? `${2}px` : `${0}px`;
  };

  const inputWithTailLeftFormula = () => {
    const calculateSliderPosition = (value: number) => {
      const scaleFactor = width / ((max - min) / step);
      const thumbOffset = (WEBKIT_SLIDER_THUMB / ((max - min) / step)) * value;
      const minOffset = scaleFactor * min;

      return (
        scaleFactor * value + WEBKIT_SLIDER_THUMB - thumbOffset - minOffset
      );
    };

    if (isDouble) {
      const positionOne = calculateSliderPosition(valueRange);
      const positionTwo = calculateSliderPosition(valueRangeTwo!);
      const difference = positionOne - positionTwo;

      return (
        ((step * (width - WEBKIT_SLIDER_THUMB)) / (max - min)) *
          (valueRange - min) +
        WEBKIT_SLIDER_THUMB / 2 -
        difference / 2
      );
    }

    const range = (max - min) / step;
    const widthPerStep =
      (width - (refTail.current?.offsetWidth || 0) - WEBKIT_SLIDER_THUMB) /
      range;

    const adjustedValueRange = valueRange >= max ? max : valueRange;
    const countSteps = adjustedValueRange / step;

    return widthPerStep * countSteps;
  };

  return (
    <div className="rangeWithProgress2">
      <InputWithTail
        currencyIcon={currencyIcon}
        handleBlur={handleBlur}
        handleFocus={handleFocus}
        isResizeNoDouble={isResizeNoDouble}
        isInputFocused={isInputFocused}
        handleInputValueRange={handleInputValueRange}
        valueTwo={valueRangeTwo}
        handleInputValueRangeTwo={handleInputValueRangeTwo}
        value={valueRange}
        isDouble={isDouble}
        left={inputWithTailLeftFormula()}
        isThin={isThin}
        refContainer={refTail}
        havePercent={havePercent}
      />
      <div
        className={cx('rangeWithProgress2__range', {
          ['rangeWithProgress2__range--isThin']: isThin,
        })}
        ref={myRef}
      >
        <div
          className="rangeWithProgress2__yellowProgress"
          style={{
            width: yellowProgressWidthFormula(),
            left: yellowProgressLeftFormula(),
          }}
        />
        <div
          className={cx('rangeWithProgress2__inputsRange', {
            ['rangeWithProgress2__inputsRange--isDouble']: isDouble,
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
          className={cx('rangeWithProgress2__buttonUnlockPremium', {
            'rangeWithProgress2__buttonUnlockPremium--visible':
              valueUnlockPremium && valueRange === valueUnlockPremium,
            'rangeWithProgress2__buttonUnlockPremium--noVisible': !(
              valueUnlockPremium && valueRange === valueUnlockPremium
            ),
          })}
          style={{
            right: `${
              (width / 2) * (1 - (step * (valueRange - min)) / (max - min)) -
              BUTTON_UNLOCK_PREMIUM_WIDTH / 2
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
          <LazyImage
            loading="lazy"
            src={arrow}
            width={24}
            height={24}
            alt={'Arrow'}
          />
        </div>
        {!isThin && (
          <div className="rangeWithProgress2__containerVerticalLine">
            {Array.from(
              { length: Math.floor((width + 17) / DISTANCE_BETWEEN_DASHES) },
              (_, i) => (
                <div
                  className={cx('rangeWithProgress2__verticalLine')}
                  key={i}
                  style={{ left: DISTANCE_BETWEEN_DASHES * i - 18 }}
                />
              ),
            )}
          </div>
        )}
      </div>
      {isHaveMinMaxValues && (
        <>
          <div className="rangeWithProgress2__maxMinContainer">
            <Typography
              text={`${min}${havePercent ? '%' : ''} `}
              className={'rangeWithProgress2__textMaxMin'}
              color={colors['textSecondary']}
            />
            {textBetweenMaxMin && (
              <Typography
                color={colors['textSecondary']}
                text={textBetweenMaxMin}
                className={'rangeWithProgress2__commission'}
              />
            )}
            <Typography
              text={`${max}${havePercent ? '%' : ''} `}
              className={'rangeWithProgress2__textMaxMin'}
              color={colors['textSecondary']}
            />
          </div>
        </>
      )}
    </div>
  );
};
