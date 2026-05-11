import {
  CSSProperties,
  FC,
  forwardRef,
  RefObject,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
} from 'react';
import { createPortal } from 'react-dom';
import cx from 'classnames';

import { useDateTimePicker } from '~components/Inputs/DatePicker/useDatePicker';
import { InputDropdownSelect } from '~components/Inputs/InputDropdownSelect';
import { Typography } from '~components/Typography';
import { useOnClickOutside } from '~hooks/useClickOutside';
import useMediaQuery from '~hooks/useMediaQuery';
import { Chevron } from '~icons/Chevrons/Chevron';
import { CalendarIcon } from '~icons/JsxSvg/CalendarIcon';
import { pxToRem } from '~utils/general';

import styles from './styles.module.scss';

type Placement = 'bottom' | 'top' | 'right' | 'left';

function useSmartDropdownPosition(
  buttonRef: RefObject<HTMLElement | null>,
  panelRef: RefObject<HTMLElement | null>,
  isOpen: boolean,
) {
  const [style, setStyle] = useState<CSSProperties>({});
  const { isMedium, isSmallTablet, isTablet, isPhone } = useMediaQuery();
  const [placement, setPlacement] = useState<Placement>('bottom');

  useLayoutEffect(() => {
    const changePosition = () => {
      if (!isOpen || !buttonRef.current || !panelRef.current) return;

      const buttonRect = buttonRef.current.getBoundingClientRect();
      const panelRect = panelRef.current.getBoundingClientRect();
      const { innerWidth, innerHeight } = window;

      const space = {
        bottom: innerHeight - buttonRect.bottom,
        top: buttonRect.top,
        right: innerWidth - buttonRect.right,
        left: buttonRect.left,
      };

      let best: Placement = 'bottom';
      let maxSpace = -Infinity;

      (Object.keys(space) as Placement[]).forEach((side) => {
        if (space[side] > maxSpace) {
          best = side;
          maxSpace = space[side];
        }
      });

      setPlacement(best);

      let top = 0;
      let left = 0;

      switch (best as Placement) {
        case 'bottom': {
          top = buttonRect.bottom;
          left = buttonRect.left;
          break;
        }

        case 'top': {
          top = buttonRect.top - panelRect.height;
          left = buttonRect.left;
          break;
        }

        case 'right':
          top = buttonRect.top;
          left = buttonRect.right;
          break;
        case 'left':
          top = buttonRect.top;
          left = buttonRect.left - panelRect.width;
          break;
      }

      setStyle({
        position: 'absolute',
        top: Math.max(0, Math.min(top, innerHeight - panelRect.height)),
        left: Math.max(0, Math.min(left, innerWidth - panelRect.width)),
        zIndex: 9999,
      });
    };

    changePosition();

    document.body.addEventListener('scroll', changePosition);

    return () => {
      document.body.removeEventListener('scroll', changePosition);
    };
  }, [isOpen, isMedium, isSmallTablet, isTablet, isPhone]);

  return { style, placement };
}

type CustomButtonOpenProps = {
  value?: string;
  onClick?: () => void;
  handleClick?: () => void;
  idButton: string;
  isDuelVariant?: boolean;
  isCalendarOpen: boolean;
  classNameButtonOpen?: string;
};

const CustomButtonOpen = forwardRef<HTMLButtonElement, CustomButtonOpenProps>(
  (
    {
      onClick,
      isCalendarOpen,
      value,
      handleClick,
      idButton,
      classNameButtonOpen,
    },
    ref,
  ) => {
    return (
      <button
        className={cx(styles.containerButtonOpen, classNameButtonOpen)}
        ref={ref}
        id={idButton}
        onClick={(e) => {
          e.stopPropagation();
          onClick && onClick();
          handleClick && handleClick();
        }}
      >
        <Typography text={`${value}`} />
        <div
          className={cx(styles.buttonOpen, {
            [`${styles['buttonOpen--active']}`]: isCalendarOpen,
          })}
        >
          <CalendarIcon />
          <Chevron />
        </div>
      </button>
    );
  },
);

type ButtonOpenWithDateProps = {
  value?: string;
  onClick?: () => void;
  isCalendarOpen?: boolean;
  handleClick?: () => void;
  placeholder?: string;
  idButton: string;
  havePlaceholder?: boolean;
  classNameButtonOpen?: string;
};

const ButtonOpenWithDate = forwardRef<
  HTMLButtonElement,
  ButtonOpenWithDateProps
>(
  (
    {
      value,
      havePlaceholder,
      onClick,
      handleClick,
      isCalendarOpen,
      placeholder,
      idButton,
      classNameButtonOpen,
    },
    ref,
  ) => {
    const { isMedium } = useMediaQuery();

    return (
      <button
        className={cx(
          styles.button,
          {
            [`${styles['button--onFocus']}`]: isCalendarOpen,
          },
          classNameButtonOpen,
        )}
        id={idButton}
        onClick={() => {
          onClick && onClick();
          handleClick && handleClick();
        }}
        ref={ref}
      >
        <div className={styles.containerCalendarIconAndValue}>
          <CalendarIcon />
          <Typography
            text={havePlaceholder ? placeholder : value!}
            variant={isMedium ? 'body-small' : 'body'}
            customStyles={{ whiteSpace: 'nowrap' }}
            lineHeight={1.3}
          />
        </div>
        <Chevron />
      </button>
    );
  },
);

export type ButtonOpenDatepickerProps = {
  idButton: string;
  value: string;
  isCalendarOpen: boolean;
  classNameButtonOpen?: string;
  handleClick: () => void;
  refButton?: RefObject<HTMLButtonElement | null>;
  havePlaceholder?: boolean;
  selectedDate: Date;
  placeholder?: string;
};

interface Props {
  isButtonOpenWithDate?: boolean;
  havePlaceholder?: boolean;
  id?: string;
  classNameButtonOpen?: string;
  minDate?: Date;
  maxDate?: Date;
  value?: Date;
  onChangeConfirm?: (date?: Date) => void;
  usePortal?: boolean;
  placeholder?: string;
  isDefaultPicker?: boolean;
  RenderButtonOpen?: FC<ButtonOpenDatepickerProps>;
  classNameContainer?: string;
  hideTime?: boolean;
}

function getWeeksOfMonth(year: number, month: number) {
  const firstOfMonth = new Date(year, month, 1);
  const lastOfMonth = new Date(year, month + 1, 0);

  const firstDate = new Date(firstOfMonth);

  while (firstDate.getDay() !== 1) {
    firstDate.setDate(firstDate.getDate() - 1);
  }

  const lastDate = new Date(lastOfMonth);

  while (lastDate.getDay() !== 0) {
    lastDate.setDate(lastDate.getDate() + 1);
  }

  const weeks = [];

  const current = new Date(firstDate);

  while (current <= lastDate) {
    const week = [];

    for (let i = 0; i < 7; i++) {
      week.push(new Date(current));
      current.setDate(current.getDate() + 1);
    }

    weeks.push(week);
  }

  return weeks;
}

const shortDayNames = ['Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa', 'Su'];
const monthsNames = [
  'Jan',
  'Feb',
  'Mar',
  'Apr',
  'May',
  'Jun',
  'Jul',
  'Aug',
  'Sep',
  'Oct',
  'Nov',
  'Dec',
];

const minDateDefault = new Date(1970, 0, 1, 23, 59, 59);
const maxDateDefault = new Date(2100, 11, 31, 23, 59, 59);
const valueDefault = new Date();

export const DatePicker = ({
  isButtonOpenWithDate,
  id,
  classNameButtonOpen,
  havePlaceholder,
  minDate = minDateDefault,
  maxDate = maxDateDefault,
  value = valueDefault,
  onChangeConfirm,
  isDefaultPicker,
  usePortal = true,
  RenderButtonOpen,
  placeholder,
  classNameContainer,
  hideTime,
}: Props) => {
  const [isCalendarOpen, setIsCalendarOpen] = useState(false);
  const buttonRef = useRef<HTMLButtonElement>(null);
  const panelRef = useRef<HTMLDivElement>(null);
  const [timeSlots, setTimeSlots] = useState<
    { stringValue: string; h: number; m: number }[]
  >([]);
  const { isSmallTablet } = useMediaQuery();
  const { style: dropdownStyle } = useSmartDropdownPosition(
    buttonRef,
    panelRef,
    isCalendarOpen,
  );

  const {
    setDateTime,
    selected,
    getSelected,
    years,
    months,
    checkDate,
    hours,
    minutes,
  } = useDateTimePicker({
    minDate,
    maxDate,
    value,
  });

  useEffect(() => {
    if (value?.getTime() !== getSelected('ref')?.getTime()) {
      setDateTime({
        month: value?.getMonth(),
        year: value?.getFullYear(),
        day: value?.getDate(),
        hour: value?.getHours(),
        minute: value?.getMinutes(),
        second: value?.getSeconds(),
      });
    }
  }, [value]);

  const isDateValid = checkDate(selected, minDate, maxDate);

  const handleConfirm = () => {
    if (isDateValid) {
      onChangeConfirm && onChangeConfirm(selected);
      setIsCalendarOpen(false);
    }
  };

  useEffect(() => {
    const slots: { stringValue: string; h: number; m: number }[] = [];

    hours.forEach((h) => {
      minutes
        .filter((m) => m % 15 === 0)
        .forEach((m) => {
          const period = h < 12 ? 'AM' : 'PM';
          const hour12 = h % 12 === 0 ? 12 : h % 12;
          const minuteStr = m.toString().padStart(2, '0');

          slots.push({
            stringValue: `${hour12}:${minuteStr} ${period}`,
            h,
            m,
          });
        });
    });

    setTimeSlots(slots);
  }, [hours, minutes, minDate]);

  useOnClickOutside(panelRef, (e) => {
    if (e.target && !buttonRef.current?.contains(e.target as Node))
      setIsCalendarOpen(false);
  });

  const renderDropdown = (
    <>
      {isSmallTablet && (
        <div className={styles.containerDropdown__mobileBgBlur} />
      )}
      <div
        className={cx(styles.containerDropdown, {
          [`${styles['containerDropdown--hideTime']}`]: hideTime,
        })}
        ref={panelRef}
        style={usePortal && !isSmallTablet ? dropdownStyle : undefined}
      >
        <div className={styles.containerDropdown__mainContent}>
          <div
            className={cx(styles.containerDropdown__datePicker, {
              [`${styles['containerDropdown__datePicker--hideTime']}`]:
                hideTime,
            })}
          >
            <div className={styles.containerDropdown__monthAndYear}>
              <div className={styles.containerDropdown__headerSelects}>
                <InputDropdownSelect
                  options={months.map((e, i) => ({
                    value: `${e}`,
                    label: `${monthsNames[e]}`,
                    id: i,
                  }))}
                  widthDropDown={pxToRem(185)}
                  isDropDownValueCenter={false}
                  variant="isDropDownCalendar"
                  dropDownPositionX="left"
                  visibleItems
                  value={{
                    value: `${selected.getMonth()}`,
                    label: `${monthsNames[selected.getMonth()]}`,
                    id: 1,
                  }}
                  onChange={(v) => {
                    console.log(
                      new Date(new Date(selected).setMonth(+v.value)),
                      minDate,
                      maxDate,
                      checkDate(
                        new Date(new Date(selected).setMonth(+v.value)),
                        minDate,
                        maxDate,
                      ),
                      minDate?.getDate() || selected.getDate(),
                    );

                    setDateTime({
                      month: +v.value,
                      day: checkDate(
                        new Date(new Date(selected).setMonth(+v.value)),
                        minDate,
                        maxDate,
                      )
                        ? undefined
                        : minDate?.getDate() || selected.getDate(),
                    });
                  }}
                />
                <InputDropdownSelect
                  options={years.map((e, i) => ({
                    value: `${e}`,
                    id: i,
                    label: `${e}`,
                  }))}
                  widthDropDown={pxToRem(185)}
                  isDropDownValueCenter={false}
                  variant="isDropDownCalendar"
                  dropDownPositionX="right"
                  value={{
                    value: `${selected.getFullYear()}`,
                    label: `${selected.getFullYear()}`,
                    id: 1,
                  }}
                  onChange={(v) => {
                    setDateTime({
                      year: +v.value,
                    });
                  }}
                />
              </div>
              <div className={styles.containerDropdown__headerButtons}>
                <button
                  onClick={() => {
                    setDateTime({
                      month: (selected?.getMonth() || 0) - 1,
                    });
                  }}
                  className={cx(
                    styles.containerDropdown__buttonHeader,
                    styles['containerDropdown__buttonHeader--left'],
                    {
                      [`${styles['containerDropdown__buttonHeader--disabled']}`]:
                        !checkDate(
                          new Date(
                            selected.getFullYear(),
                            selected.getMonth() - 1,
                            selected.getDate(),
                            selected.getHours(),
                            selected.getMinutes(),
                            selected.getSeconds(),
                          ),
                          minDate,
                          maxDate,
                        ),
                    },
                  )}
                >
                  <Chevron />
                </button>
                <button
                  onClick={() => {
                    setDateTime({
                      month: (selected?.getMonth() || 0) + 1,
                    });
                  }}
                  className={cx(
                    styles.containerDropdown__buttonHeader,
                    styles['containerDropdown__buttonHeader--right'],
                    {
                      [`${styles['containerDropdown__buttonHeader--disabled']}`]:
                        !checkDate(
                          new Date(
                            selected.getFullYear(),
                            selected.getMonth() + 1,
                            selected.getDate(),
                            selected.getHours(),
                            selected.getMinutes(),
                            selected.getSeconds(),
                          ),
                          minDate,
                          maxDate,
                        ),
                    },
                  )}
                >
                  <Chevron />
                </button>
              </div>
            </div>
            <div className={styles.containerDropdown__tableDays}>
              <div className={styles.containerDropdown__week}>
                {shortDayNames.map((e, i) => (
                  <div key={i} className={styles.containerDropdown__dayName}>
                    {e}
                  </div>
                ))}
              </div>
              {getWeeksOfMonth(
                selected?.getFullYear() || 0,
                selected?.getMonth() || 0,
              ).map((week, weekIndex) => (
                <div key={weekIndex} className={styles.containerDropdown__week}>
                  {week.map((day, i) => {
                    return (
                      <div
                        key={i}
                        className={cx(styles.containerDropdown__day, {
                          [`${styles['containerDropdown__day--active']}`]:
                            selected?.getDate() === day.getDate() &&
                            (selected?.getMonth() || 0) === day.getMonth(),
                          [`${styles['containerDropdown__day--disabled']}`]: !(
                            new Date(day).setHours(0, 0, 0, 0) >=
                              new Date(minDate).setHours(0, 0, 0, 0) &&
                            new Date(day).setHours(0, 0, 0, 0) <=
                              new Date(maxDate).setHours(0, 0, 0, 0)
                          ),
                          [`${styles['containerDropdown__day--anotherMonth']}`]:
                            (selected?.getMonth() || 0) !== day.getMonth(),
                        })}
                        onClick={() => {
                          setDateTime({
                            month: day.getMonth(),
                            day: day.getDate(),
                          });
                        }}
                      >
                        {day.getDate()}
                      </div>
                    );
                  })}
                </div>
              ))}
            </div>
          </div>
          {!hideTime && (
            <div className={styles.containerDropdown__timePicker}>
              <div className={styles.containerDropdown__titleTime}>
                <Typography text={'Time'} />
              </div>
              <div className={styles.containerDropdown__timeContainer}>
                {timeSlots.map((time, index) => (
                  <div
                    className={cx(styles.containerDropdown__time, {
                      [`${styles['containerDropdown__time--active']}`]:
                        time.h === selected?.getHours() &&
                        time.m === selected.getMinutes(),
                    })}
                    key={index}
                    onClick={() => {
                      setDateTime({
                        hour: time.h,
                        minute: time.m,
                      });
                    }}
                  >
                    {time.stringValue}
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
        <div className={styles.containerDropdown__footer}>
          <button
            className={styles.containerDropdown__cancelButton}
            onClick={() => setIsCalendarOpen(false)}
          >
            <Typography
              text="Cancel"
              isBungee
              variant="body"
              lineHeight={1.4}
            />
          </button>
          <button
            className={cx(styles.containerDropdown__confirmButton, {
              [`${styles['containerDropdown__confirmButton--disabled']}`]:
                !isDateValid,
            })}
            disabled={!isDateValid}
            onClick={handleConfirm}
          >
            <Typography
              text="Confirm"
              isBungee
              variant={'body'}
              lineHeight={1.4}
            />
          </button>
        </div>
      </div>
    </>
  );

  const formattedValue = selected
    .toLocaleString('en-US', {
      day: 'numeric',
      month: 'short',
      year: 'numeric',
      hour: 'numeric',
      minute: '2-digit',
      hour12: true,
    })
    .replace(',', '')
    .replace(/([A-Za-z]{3})(?=\s)/, '$1.');

  const buttonOpenProps: ButtonOpenDatepickerProps = {
    idButton: `${id}-button-open`,
    value: formattedValue,
    selectedDate: selected,
    isCalendarOpen,
    classNameButtonOpen,
    handleClick: () => setIsCalendarOpen((p) => !p),
    havePlaceholder,
    placeholder,
  };

  return (
    <div className={cx(styles.containerDatePicker, classNameContainer)}>
      {RenderButtonOpen ? (
        <RenderButtonOpen {...buttonOpenProps} refButton={buttonRef} />
      ) : isButtonOpenWithDate ? (
        <ButtonOpenWithDate
          {...buttonOpenProps}
          ref={buttonRef}
          havePlaceholder={havePlaceholder}
        />
      ) : (
        <CustomButtonOpen {...buttonOpenProps} ref={buttonRef} />
      )}

      {isDefaultPicker && (
        <input
          type="datetime-local"
          id="date"
          onChange={(e) => {
            const value = new Date(e.target.value);

            onChangeConfirm && onChangeConfirm(value);
            setDateTime({
              month: value.getMonth(),
              day: value.getDate(),
              year: value.getFullYear(),
              hour: value.getHours(),
              minute: value.getMinutes(),
              second: value.getSeconds(),
            });
          }}
          className={styles.defaultPicker}
          min={minDate?.getTime()}
          max={maxDate?.getTime()}
          value={
            value ? new Date(value!).toISOString().slice(0, 16) : undefined
          }
        />
      )}

      {isCalendarOpen &&
        !isDefaultPicker &&
        (usePortal
          ? createPortal(renderDropdown, document.body)
          : renderDropdown)}
    </div>
  );
};
