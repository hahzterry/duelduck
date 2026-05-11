import { ChangeEvent, KeyboardEvent, useEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import cx from 'classnames';

import { FadeAnimation } from '~components/Animations/FadeAnimation';
import { Typography } from '~components/Typography';
import { useOnClickOutside } from '~hooks/useOnClickOutside';
import { Clock } from '~icons/JsxSvg/Clock';

import styles from './styles.module.scss';

const ButtonSelect = ({
  isActive,
  text,
  onClick,
}: {
  isActive: boolean;
  text: string;
  onClick: () => void;
}) => {
  return (
    <button
      className={cx(styles.timePicker__buttonCheck, {
        [`${styles['timePicker__buttonCheck--active']}`]: isActive,
      })}
      onClick={onClick}
    >
      {text}
    </button>
  );
};

type FormatTime = '24h' | 'PM' | 'AM';

export interface TimeValue {
  hour: string;
  minute: string;
  format: FormatTime;
}

interface TimePickerProps {
  value: TimeValue;
  onChange: (value: TimeValue) => void;
  disabled?: boolean;
  className?: string;
}

const minutesList = ['00', '15', '30', '45'];
const formatOptions: FormatTime[] = ['24h', 'AM', 'PM'];

export const TimePicker = ({
  value,
  onChange,
  disabled = false,
  className,
}: TimePickerProps) => {
  const { hour, minute, format: formatType } = value;

  const valueRef = useRef(value);
  const hourInputRef = useRef<HTMLInputElement>(null);
  const minutesInputRef = useRef<HTMLInputElement>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const buttonRef = useRef<HTMLDivElement>(null);

  const [isOpen, setIsOpen] = useState(false);
  const [position, setPosition] = useState<{
    width: number;
    top: number;
    left: number;
  } | null>(null);

  const updateTime = (updates: Partial<TimeValue>) => {
    valueRef.current = { ...valueRef.current, ...updates };
    onChange({ ...value, ...updates });
  };

  const getVisibleHours = () => {
    const is24Format = formatType === '24h';

    return [...new Array(is24Format ? 24 : 12)].map((_, i) =>
      is24Format ? `${i}` : `${i + 1}`,
    );
  };

  const handleHourChange = (e: ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value.replace(/\D/g, '').slice(0, 2);

    if (val === '') {
      updateTime({ hour: '' });

      return;
    }

    const num = parseInt(val, 10);

    if (num <= 23) {
      const nextFormat = num > 11 ? '24h' : formatType;

      updateTime({ hour: val, format: nextFormat });

      if (val.length === 2 || num >= 3) {
        minutesInputRef.current?.focus();
      }
    } else {
      minutesInputRef.current?.focus();
    }
  };

  const handleKeyDown = (
    e: KeyboardEvent<HTMLInputElement>,
    currentType: 'hours' | 'minutes',
  ) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      if (currentType === 'hours') {
        minutesInputRef.current?.focus();
      } else {
        minutesInputRef.current?.blur();
      }
    }

    if (
      (e.key === 'Backspace' || e.key === 'Delete') &&
      currentType === 'minutes' &&
      !minute
    ) {
      e.preventDefault();
      hourInputRef.current?.focus();
    }

    if (
      e.key === 'ArrowRight' &&
      currentType === 'hours' &&
      e.currentTarget.selectionStart === valueRef.current.hour.length
    ) {
      e.preventDefault();
      minutesInputRef.current?.focus();
    }

    if (
      e.key === 'ArrowLeft' &&
      currentType === 'minutes' &&
      e.currentTarget.selectionStart === 0
    ) {
      e.preventDefault();
      hourInputRef.current?.focus();
    }
  };

  const handleFormatChange = (newFormat: FormatTime) => {
    const currentH = parseInt(hour || '0', 10);

    let nextHour = hour;

    if (newFormat !== '24h' && currentH >= 12) {
      nextHour = String(currentH - 12).padStart(2, '0');
    }

    updateTime({ format: newFormat, hour: nextHour });
  };

  const handleBlur = (field: 'hour' | 'minute') => {
    const val = valueRef.current[field];

    if (val && val.length === 1) {
      updateTime({ [field]: `0${val}` });
    }
  };

  useEffect(() => {
    if (typeof window === 'undefined') return;
    const el = buttonRef.current;

    if (!el) return;

    const changePosition = () => {
      const el = buttonRef.current;

      if (!el) return;
      const rect = el.getBoundingClientRect();

      setPosition({
        width: rect.width,
        top: rect.y + rect.height - 13,
        left: rect.left,
      });
    };

    changePosition();

    document.body.addEventListener('scroll', changePosition);

    return () => {
      document.body.removeEventListener('scroll', changePosition);
    };
  }, [isOpen]);

  useOnClickOutside(buttonRef, (e) => {
    if (!dropdownRef.current?.contains(e.target as Node)) setIsOpen(false);
  });

  return (
    <div
      className={cx(styles.timePicker, className, {
        [`${styles.disabled}`]: disabled,
      })}
      ref={buttonRef}
      onClick={(e) => {
        if (disabled) return;
        const isMinutesClick = minutesInputRef.current?.contains(
          e.target as Node,
        );

        if (!isMinutesClick) hourInputRef.current?.focus();
        setIsOpen(true);
      }}
    >
      <span className={styles.timePicker__inputs}>
        <input
          type="text"
          ref={hourInputRef}
          value={hour}
          onChange={handleHourChange}
          onFocus={(e) => e.target.select()}
          onBlur={() => handleBlur('hour')}
          placeholder="00"
          onKeyDown={(e) => handleKeyDown(e, 'hours')}
          disabled={disabled}
        />
        <Typography text={':'} />
        <input
          type="text"
          ref={minutesInputRef}
          value={minute}
          onKeyDown={(e) => handleKeyDown(e, 'minutes')}
          onChange={(e) => {
            const val = e.target.value.replace(/\D/g, '').slice(0, 2);

            if (val === '' || parseInt(val, 10) <= 59) {
              updateTime({ minute: val });
              if (val.length === 2) minutesInputRef.current?.blur();
            }
          }}
          onFocus={(e) => e.target.select()}
          onBlur={() => handleBlur('minute')}
          placeholder="00"
          disabled={disabled}
        />
      </span>
      <Clock />

      {isOpen &&
        typeof window !== 'undefined' &&
        createPortal(
          <FadeAnimation
            isVisible={isOpen}
            customClassNames={{
              enter: styles.fadeEnter,
              enterActive: styles.fadeEnterActive,
              exit: styles.fadeExit,
              exitActive: styles.fadeExitActive,
            }}
            customRef={dropdownRef}
            customWrapperClass={styles.timePicker__dropdown}
            style={{ ...position }}
          >
            <div className={styles.timePicker__listTimes}>
              {getVisibleHours().map((h) => (
                <ButtonSelect
                  key={h}
                  text={h.padStart(2, '0')}
                  isActive={parseInt(hour) === parseInt(h)}
                  onClick={() => {
                    updateTime({ hour: h.padStart(2, '0') });
                    minutesInputRef.current?.focus();
                  }}
                />
              ))}
            </div>

            <div className={styles.timePicker__listTimes}>
              {minutesList.map((m) => (
                <ButtonSelect
                  key={m}
                  text={m}
                  isActive={minute === m}
                  onClick={() => {
                    updateTime({ minute: m });
                    setIsOpen(false);
                  }}
                />
              ))}
            </div>

            <div className={styles.timePicker__listTimes}>
              {formatOptions.map((f) => (
                <ButtonSelect
                  key={f}
                  text={f}
                  isActive={formatType === f}
                  onClick={() => handleFormatChange(f)}
                />
              ))}
            </div>
          </FadeAnimation>,
          document.body,
        )}
    </div>
  );
};
