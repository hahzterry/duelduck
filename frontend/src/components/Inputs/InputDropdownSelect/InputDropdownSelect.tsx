'use client';
import './style.scss';

import { ReactNode, useEffect, useRef, useState } from 'react';
import cx from 'classnames';

import { FadeAnimation } from '~components/Animations/FadeAnimation';
import { Typography } from '~components/Typography';
import { Chevron } from '~icons/Chevrons/Chevron';
import { SearchIcon } from '~icons/JsxSvg/SearchIcon';
import colors from '~styles/colors';

import { InputDropdownSelectOption } from './InputDropdownSelectOption';

export interface InputDropdownSelectOptionValue {
  value: string;
  label: string;
  id: number;
  icon?: ReactNode;
}

interface InputDropdownSelectProps {
  options: InputDropdownSelectOptionValue[];
  value?: InputDropdownSelectOptionValue;
  inputPlaceholder?: string;
  isHaveSearch?: boolean;
  onChange?: (v: InputDropdownSelectOptionValue) => void;
  widthDropDown?: string;
  isDropDownValueCenter?: boolean;
  variant?: 'default' | 'isDropDownCalendar';
  dropDownPositionX?: 'center' | 'left' | 'right';
  backgroundColorDropDown?: string;
  selectPlaceholder?: string;
  visibleItems?: boolean;
  minWidth?: string;
  className?: string;
}

export const InputDropdownSelect = ({
  options,
  value = { value: '', label: '', id: 0 },
  inputPlaceholder,
  isHaveSearch,
  onChange,
  widthDropDown,
  isDropDownValueCenter,
  variant,
  dropDownPositionX = 'center',
  backgroundColorDropDown = colors['background'],
  selectPlaceholder,
  visibleItems = true,
  className,
}: InputDropdownSelectProps) => {
  const [isFocus, setIsFocus] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const containerOptionsRef = useRef<HTMLDivElement>(null);
  const [localOptions, setLocalOption] = useState(options);
  const [inputSearchValue, setInputSearchValue] = useState('');

  useEffect(() => {
    if (containerOptionsRef.current)
      containerOptionsRef.current.scrollTop =
        ((containerOptionsRef.current?.scrollHeight || 0) /
          localOptions.length -
          1) *
        localOptions.findIndex((e) => e.value === value.value);
  }, []);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        containerRef.current &&
        event.target instanceof Node &&
        !containerRef.current.contains(event.target)
      ) {
        setIsFocus(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  useEffect(() => {
    setLocalOption(
      options.filter((elem) =>
        elem.value.toLowerCase().includes(inputSearchValue.toLowerCase()),
      ),
    );
  }, [inputSearchValue, options]);

  return (
    <div className="inputDropdownSelect" ref={containerRef}>
      {variant == 'isDropDownCalendar' ? (
        <>
          <button
            className={cx({
              ['inputDropdownSelect__inputDropDownCalendar']: true,
              ['inputDropdownSelect__inputDropDownCalendar--onFocus']: isFocus,
            })}
            onClick={() => setIsFocus((s) => !s)}
          >
            <Typography
              text={value.label || selectPlaceholder}
              variant={'title-2'}
              lineHeight={1.3}
              fontWeight="sb"
            />
            <Chevron />
          </button>
        </>
      ) : (
        <>
          <div
            className={cx({
              ['inputDropdownSelect__input']: true,
              ['inputDropdownSelect__input--haveIcon']: value.icon,
              [className as string]: className,
            })}
            onClick={() => setIsFocus((s) => !s)}
          >
            <div className="inputDropdownSelect__containerSelectedValue">
              {value.icon}
            </div>
            <Typography
              text={value.label || selectPlaceholder}
              variant={'body-big'}
              fontWeight="r"
              lineHeight={1.3}
            />
          </div>
          <div
            className={cx({
              ['inputDropdownSelect__containerSvg']: true,
              ['inputDropdownSelect__containerSvg--onFocus']: isFocus,
            })}
          >
            <Chevron />
          </div>
        </>
      )}

      <FadeAnimation
        customWrapperClass={cx({
          ['inputDropdownSelect__containerDropdown']: true,
          ['inputDropdownSelect__containerDropdown--isDropDownRight']:
            dropDownPositionX == 'right',
          ['inputDropdownSelect__containerDropdown--isDropDownLeft']:
            dropDownPositionX == 'left',
        })}
        customClassNames={{
          enter: 'inputDropdownSelect__enter',
          enterActive: 'inputDropdownSelect__enterActive',
          exit: 'inputDropdownSelect__exit',
          exitActive: 'inputDropdownSelect__exitActive',
        }}
        isVisible={isFocus}
        style={{
          width: widthDropDown,
          backgroundColor: backgroundColorDropDown,
        }}
      >
        <div
          className="inputDropdownSelect__dropdown"
          ref={containerOptionsRef}
        >
          <div className="inputDropdownSelect__dropdownContent">
            {isHaveSearch && (
              <div className={`inputDropdownSelect__customOption`}>
                <div
                  className={cx({
                    ['inputDropdownSelect__customOptionContainerValue']: true,
                  })}
                >
                  <SearchIcon />
                  <input
                    value={inputSearchValue}
                    type="text"
                    placeholder={inputPlaceholder}
                    onChange={(e) => setInputSearchValue(e.target.value)}
                  />
                </div>
                <hr className={'inputDropdownSelect__horizontalLine'} />
              </div>
            )}
            {visibleItems &&
              localOptions.map((e, i) => (
                <InputDropdownSelectOption
                  id={i}
                  key={i}
                  value={e.label}
                  icon={e.icon}
                  isCenter={isDropDownValueCenter}
                  isSelected={e.value == value.value}
                  onClick={() => {
                    onChange && onChange(e);

                    setIsFocus(false);
                  }}
                />
              ))}
          </div>
        </div>
      </FadeAnimation>
    </div>
  );
};
