import './styles.scss';

import { ChangeEvent, MutableRefObject, ReactNode } from 'react';
import cx from 'classnames';

import { Typography } from '~components/Typography';

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
  refContainer: MutableRefObject<HTMLDivElement | null>;
  isResizeNoDouble?: boolean;
  havePercent?: boolean;
}

export const InputWithTail = ({
  left,
  value,
  isDouble,
  valueTwo,
  isThin,
  refContainer,
  isResizeNoDouble,
  havePercent,
}: RangeWithProgressProps) => {
  const charLength = (value?: number) => (value || 0).toString().length;

  const charLengthSum = charLength(valueTwo) + charLength(value);
  const baseOffset = isDouble ? 22 : 28;

  let leftStyle;

  if (isDouble) {
    leftStyle = `calc(${left}px - ${charLengthSum / 2}ch - ${baseOffset}px)`;
  } else if (isResizeNoDouble) {
    leftStyle = `calc(${left}px - ${charLength(value) / 2}ch - ${baseOffset}px)`;
  } else {
    leftStyle = `${left}px`;
  }

  return (
    <div
      className={cx({
        ['rangeWithProgress2__progress']: true,
        ['rangeWithProgress2__progress--isThin']: isThin,
      })}
      style={{
        left: leftStyle,
      }}
      ref={refContainer}
    >
      <Typography
        text={`${value}${havePercent ? '%' : ''}`}
        className={'rangeWithProgress2__progressText'}
      />
    </div>
  );
};
