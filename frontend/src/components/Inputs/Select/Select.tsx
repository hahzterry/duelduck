'use client';
import { Fragment, ReactNode, useEffect, useRef, useState } from 'react';
import cx from 'classnames';

import { FadeAnimation } from '~components/Animations/FadeAnimation';
import { CountElement } from '~components/CountElement';
import { Typography } from '~components/Typography';
import { Chevron } from '~icons/Chevrons/Chevron';
import { Check2 } from '~icons/JsxSvg/Check2';

import styles from './styles.module.scss';

export interface SelectOption<T> {
  id: number;
  label: string;
  value: T;
  icon?: ReactNode;
  count?: number;
}

interface SelectProps<T> {
  options: SelectOption<T>[];
  value?: T;
  onChange: (value: T) => void;
  optionsClass?: string;
  triggerClass?: string;
  iconOnly?: boolean;
  getLabelSelected?: (item?: SelectOption<T>) => ReactNode;
  wrapperClassName?: string;
  isLeft?: boolean;
}

export const Select = <T,>({
  options,
  value,
  onChange,
  optionsClass,
  triggerClass,
  iconOnly,
  getLabelSelected = (item) => item?.label || 'Select',
  wrapperClassName,
  isLeft,
}: SelectProps<T>) => {
  const dropdownRef = useRef<HTMLDivElement | null>(null);
  const buttonRef = useRef<HTMLDivElement>(null);
  const [isOpen, setIsOpen] = useState(false);
  const [paddingTop, setPaddingTop] = useState(0);
  const observerRef = useRef<ResizeObserver | null>(null);

  useEffect(() => {
    if (buttonRef.current) {
      observerRef.current = new ResizeObserver((e) => {
        e.forEach((e) => {
          setPaddingTop(e.target.clientHeight / 2);
        });
      });

      observerRef.current.observe(buttonRef.current);

      return () => {
        if (observerRef.current) {
          observerRef.current.disconnect();
          observerRef.current = null;
        }
      };
    }

    return () => {};
  }, []);

  useEffect(() => {
    const handleOutsideClick = (event: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node) &&
        buttonRef.current &&
        !buttonRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleOutsideClick);

    return () => {
      document.removeEventListener('mousedown', handleOutsideClick);
    };
  }, []);

  const renderOptions = () =>
    options.map(({ value: optionValue, label, id, icon, count }, index) => (
      <Fragment key={id}>
        <div
          className={cx(styles.option, optionsClass, {
            [`${styles['option--selected']}`]: optionValue === value,
          })}
          onClick={() => {
            setIsOpen(false);
            onChange(optionValue);
          }}
          style={{ justifyContent: 'center' }}
        >
          {icon}
          <Typography className={styles.label}>{label}</Typography>
          {optionValue === value && <Check2 />}
          {typeof count !== 'undefined' && <CountElement count={count} />}
        </div>
        {index !== options.length - 1 && (
          <div className={styles.separatorList} />
        )}
      </Fragment>
    ));

  const selectedValue = options.find((option) => option.value === value);

  const labelSelected = getLabelSelected(selectedValue);

  return (
    <div className={cx(styles.wrapper, wrapperClassName)}>
      <div
        ref={buttonRef}
        data-open={isOpen}
        className={cx(styles.dropdownTrigger, triggerClass, {
          [styles.iconOnly as string]: iconOnly,
        })}
        onClick={() => setIsOpen((prev) => !prev)}
      >
        {selectedValue?.icon}
        <Typography>{labelSelected}</Typography>
        <Chevron />
      </div>
      <FadeAnimation
        customRef={dropdownRef}
        customClassNames={{
          enter: styles.fadeEnter,
          enterActive: styles.fadeEnterActive,
          exit: styles.fadeExit,
          exitActive: styles.fadeExitActive,
        }}
        isVisible={isOpen}
        style={{
          paddingTop,
        }}
        customWrapperClass={cx(styles.dropdown, {
          [styles.dropdownLeft as string]: isLeft,
        })}
      >
        <div className={styles.list}>{renderOptions()}</div>
      </FadeAnimation>
    </div>
  );
};
