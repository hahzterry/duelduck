import { ChangeEventHandler } from 'react';
import cx from 'classnames';

import { LazyImage } from '~/components/Lazy/LazyImage';
import { Typography } from '~components/Typography';
import colors from '~styles/colors';

import styles from './styles.module.scss';

interface Props {
  onChange?: ChangeEventHandler<HTMLInputElement>;
  icon?: string;
  date?: string;
  activeColor?: string;
  isNoRounded?: boolean;
  min?: string;
  max?: string;
  error?: string;
  disabled?: boolean;
}

export const CustomDate = ({
  onChange,
  icon,
  isNoRounded,
  activeColor = colors['textPrimary'],
  date,
  min,
  max,
  disabled,
  error,
}: Props) => {
  return (
    <div
      className={cx(styles.customDate, {
        [`${styles['customDate--disabled']}`]: disabled,
      })}
    >
      <input
        type="datetime-local"
        id="date"
        disabled={disabled}
        onChange={onChange}
        min={min}
        max={max}
        value={date ? new Date(date!).toISOString().slice(0, 16) : undefined}
      />
      {icon && (
        <LazyImage
          loading="lazy"
          src={icon}
          width={22}
          height={22}
          style={{
            borderRadius: !isNoRounded ? '100%' : undefined,
          }}
          alt={'icon'}
        />
      )}
      <Typography
        text={
          error
            ? error
            : date
              ? new Date(date)
                  .toLocaleString('en-US', {
                    day: 'numeric',
                    month: 'short',
                    year: 'numeric',
                    hour: 'numeric',
                    minute: '2-digit',
                    hour12: true,
                  })
                  .replace(',', '')
                  .replace(/([A-Za-z]{3})(?=\s)/, '$1.')
              : 'Chose date'
        }
        color={
          error ? colors['error'] : date ? activeColor : colors['textSecondary']
        }
      />
    </div>
  );
};
