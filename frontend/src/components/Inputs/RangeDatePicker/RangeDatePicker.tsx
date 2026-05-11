import 'react-datepicker/dist/react-datepicker.css';
import './styles.scss';

import {
  forwardRef,
  MouseEventHandler,
  useEffect,
  useRef,
  useState,
} from 'react';
import DatePicker from 'react-datepicker';
import cx from 'classnames';

import { Typography } from '~components/Typography';
import useMediaQuery from '~hooks/useMediaQuery';
import { Chevron } from '~icons/Chevrons/Chevron';
import { CalendarIcon } from '~icons/JsxSvg/CalendarIcon';

type ButtonOpenWithDateProps = {
  value?: string;
  onClick?: () => void;

  isCalendarOpen?: boolean;
  handleClick?: () => void;
  placeholder?: string;

  idButton: string;
  havePlaceholder?: boolean;
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
    },
    ref,
  ) => {
    const { isMedium } = useMediaQuery();

    return (
      <button
        className={cx(
          'CustomDatePiker__button CustomDatePiker__button--rangePicker',
          {
            'CustomDatePiker__button--onFocus': isCalendarOpen,
          },
        )}
        id={idButton}
        onClick={() => {
          onClick && onClick();
          handleClick && handleClick();
        }}
        ref={ref}
      >
        <div className="CustomDatePiker__containerCalendarIconAndValue">
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

interface CustomDatePikerProps {
  onChange?: (dates: [Date | null, Date | null]) => void;
  startDate?: [Date | null, Date | null];
  maxDate?: Date;
  minDate?: Date;
  isLeft?: boolean;
  id: string;
  placement?: string;
  havePortal?: boolean;
  placeholder?: string;
  havePlaceholder?: boolean;
  autoOpen?: boolean;
  hideButton?: boolean;
}

export const CustomDatePikerRange = ({
  onChange,
  startDate,
  maxDate = new Date('12-31-2030'),
  minDate = new Date('01-01-2000'),
  isLeft,
  id,
  placement = 'left-start',
  havePortal,
  placeholder,
  autoOpen = false,
  hideButton = false,
}: CustomDatePikerProps) => {
  const ref = useRef<HTMLDivElement>(null);

  const [rootDate, setRootDate] = useState<[Date | null, Date | null]>(
    startDate || [null, null],
  );
  const [isCalendarOpen, setIsCalendarOpen] = useState(autoOpen);

  useEffect(() => {
    setRootDate(startDate || [null, null]);
  }, [startDate]);

  const handleConfirm: MouseEventHandler<HTMLButtonElement> = (e) => {
    e.preventDefault();
    setIsCalendarOpen(false);
    onChange && onChange(rootDate);
  };

  const handleCancel: MouseEventHandler<HTMLButtonElement> = (e) => {
    e.preventDefault();
    setRootDate([null, null]);
    setIsCalendarOpen(false);
    onChange && onChange([null, null]);
  };

  return (
    <div
      className="customRangePicker"
      style={
        autoOpen
          ? {
              position: 'absolute',
              right: 0,
              bottom: 0,
            }
          : undefined
      }
    >
      <DatePicker
        selectsRange
        startDate={rootDate[0]}
        id={'customRangePicker'}
        endDate={rootDate[1]}
        placeholderText={placeholder}
        portalId={havePortal ? `${id}-portal` : undefined}
        customInput={
          hideButton ? (
            <div
              id={`${id}-button-open`}
              style={{
                position: 'absolute',
                left: 0,
                top: 0,
                width: 0,
                height: 0,
                opacity: 0,
                pointerEvents: 'none',
              }}
            />
          ) : (
            <ButtonOpenWithDate
              havePlaceholder={!rootDate[0] && !rootDate[1]}
              idButton={`${id}-button-open`}
              isCalendarOpen={isCalendarOpen}
              handleClick={() => {
                setIsCalendarOpen((s) => !s);
              }}
            />
          )
        }
        calendarStartDay={1}
        dateFormat={'d MMM, yyyy'}
        open={isCalendarOpen}
        shouldCloseOnSelect={false}
        onClickOutside={(event) => {
          const buttonElement = document.getElementById(`${id}-button-open`);

          if (!buttonElement || !buttonElement.contains(event.target as Node)) {
            setRootDate(startDate || [null, null]);
            setIsCalendarOpen(false);
          }
        }}
        minDate={minDate}
        maxDate={maxDate}
        calendarContainer={({ className, children }) => {
          return (
            <div
              className={className}
              style={{
                position: isLeft ? 'relative' : 'static',
                left: isLeft ? '25%' : '',
                top: '0px',
              }}
              ref={ref}
            >
              {children}
              <div className="CustomDatePiker__footer">
                <button
                  className="CustomDatePiker__cancelButton"
                  onClick={handleCancel}
                >
                  <Typography
                    text="Cancel"
                    isBungee
                    variant="body"
                    lineHeight={1.4}
                  />
                </button>
                <button
                  className={cx('CustomDatePiker__confirmButton', {
                    ['CustomDatePiker__confirmButton--disabled']:
                      (rootDate[0] &&
                        rootDate[1] &&
                        rootDate[0].getTime() > rootDate[1].getTime()) ||
                      (rootDate[0] &&
                        rootDate[0].getTime() < minDate.getTime()) ||
                      (rootDate[1] &&
                        rootDate[1].getTime() < minDate.getTime()),
                  })}
                  onClick={handleConfirm}
                  disabled={
                    !!(
                      (rootDate[0] &&
                        rootDate[1] &&
                        rootDate[0].getTime() > rootDate[1].getTime()) ||
                      (rootDate[0] &&
                        rootDate[0].getTime() < minDate.getTime()) ||
                      (rootDate[1] && rootDate[1].getTime() < minDate.getTime())
                    )
                  }
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
          );
        }}
        highlightDates={[new Date()]}
        onChange={(dates) => {
          setRootDate(dates);
        }}
        showPopperArrow={false}
        popperPlacement={placement as any}
      />
    </div>
  );
};
