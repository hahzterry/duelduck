'use client';
import './style.scss';

import { Key, MouseEventHandler, ReactNode, useState } from 'react';
import cx from 'classnames';

import { Typography } from '~components/Typography';
import colors from '~styles/colors';

interface InputDropdownSelectOptionProps {
  id?: Key;
  value: string;
  icon?: ReactNode;
  onClick?: MouseEventHandler<HTMLButtonElement>;
  isSelected?: boolean;
  isCenter?: boolean;
}

export const InputDropdownSelectOption = ({
  id,
  value,
  icon,
  onClick,
  isSelected,
  isCenter,
}: InputDropdownSelectOptionProps) => {
  const [isHovered, setIsHovered] = useState(false);

  const handleMouseEnter = () => {
    setIsHovered(true);
  };

  const handleMouseLeave = () => {
    setIsHovered(false);
  };

  return (
    <div key={id} className={`inputDropdownSelect__customOption`}>
      <button
        className={cx({
          ['inputDropdownSelect__customOptionContainerValue']: true,
          ['inputDropdownSelect__customOptionContainerValue--isCenter']:
            isCenter,
        })}
        onClick={onClick}
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
      >
        {icon}
        <Typography
          color={
            isHovered || isSelected ? colors['primary'] : colors['textPrimary']
          }
          text={value}
          variant={'body-big'}
          lineHeight={1.3}
        />
      </button>
      <hr className={'inputDropdownSelect__horizontalLine'} />
    </div>
  );
};
