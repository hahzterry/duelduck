import { Fragment, ReactNode, useRef, useState } from 'react';
import cx from 'classnames';
import { Property } from 'csstype';

import { FadeAnimation } from '~components/Animations/FadeAnimation';
import { LazyImage } from '~components/Lazy/LazyImage';
import { Typography } from '~components/Typography';
import { useOuterClick } from '~hooks/useOuterClick';
import { Chevron } from '~icons/Chevrons/Chevron';

import styles from './styles.module.scss';

interface Element {
  value: string;
  label: string;
  icon?: ReactNode;
  isSelected?: boolean;
  onClick?: () => void;
}

interface Props {
  title: string;
  width?: Property.Width<string | number>;
  elements: Element[];
  isButton?: boolean;
  onClick?: () => void;
  customIcon?: ReactNode;
  prefixIcon?: ReactNode;
  sideDropdown?: 'left' | 'right';
  className?: string;
  triggerClassName?: string;
  parentWidth?: number;
}

export const SelectSecondary = ({
  title,
  width = 196,
  elements,
  onClick,
  isButton,
  customIcon,
  prefixIcon,
  sideDropdown = 'right',
  className,
  triggerClassName,
  parentWidth,
}: Props) => {
  const [isOpen, setIsOpen] = useState(false);
  const ref = useRef<HTMLDivElement | null>(null);

  useOuterClick(ref, () => {
    setIsOpen(false);
  });

  return (
    <div
      className={cx(styles.buttonOpen__container, {
        [className as string]: !!className,
      })}
      ref={ref}
    >
      <div
        className={cx(styles.buttonOpen, {
          [triggerClassName || '']: triggerClassName,
        })}
        onClick={() => {
          onClick?.();
          !isButton && setIsOpen(true);
        }}
      >
        <div className={styles.buttonOpen__left}>
          {prefixIcon}
          <Typography text={title} />
        </div>
        {customIcon || <Chevron />}
      </div>
      <FadeAnimation
        isVisible={isOpen && !isButton}
        style={{ width: parentWidth || width }}
        customWrapperClass={cx(
          styles.buttonOpen__dropdown,
          styles[`buttonOpen__dropdown--${sideDropdown}`],
          {
            [triggerClassName || '']: triggerClassName,
          },
        )}
      >
        <div
          className={cx(styles.buttonOpen__dropdownTitle, {
            [triggerClassName || '']: triggerClassName,
          })}
          onClick={() => {
            !isButton && setIsOpen(false);
          }}
        >
          <Typography text={title} />
          <Chevron />
        </div>
        <div className={styles.buttonOpen__dropdownScroll}>
          {elements.map((e, i, arr) => (
            <Fragment key={i}>
              <div
                className={cx(styles.buttonOpen__dropdownElement, {
                  [styles.buttonOpen__dropdownElement__selected as string]:
                    e.isSelected,
                })}
                onClick={() => {
                  !isButton && setIsOpen(false);
                  e.onClick?.();
                }}
              >
                <Typography text={e.label} />
                {typeof e.icon !== 'string' ? (
                  e.icon
                ) : (
                  <LazyImage
                    src={e.icon}
                    width={20}
                    height={20}
                    alt={e.label}
                  />
                )}
              </div>
              {arr.length - 1 !== i && (
                <div className={styles.buttonOpen__line} />
              )}
            </Fragment>
          ))}
        </div>
      </FadeAnimation>
    </div>
  );
};
