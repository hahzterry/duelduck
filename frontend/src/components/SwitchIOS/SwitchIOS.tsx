import { ChangeEventHandler, CSSProperties } from 'react';
import cx from 'classnames';

import colors from '~styles/colors';
import { ColorType } from '~types/general';

import styles from './styles.module.scss';

interface Props {
  isChecked: boolean;
  onChange?: ChangeEventHandler<HTMLInputElement>;
  id: string;
  customColors?: {
    backgroundNoChecked?: ColorType;
    backgroundChecked?: ColorType;
    knob?: ColorType;
  };
  disabled?: boolean;
  secondary?: boolean;
}

export const SwitchIOS = ({
  onChange,
  isChecked,
  id,
  customColors,
  disabled,
  secondary,
}: Props) => {
  const style: CSSProperties = {
    ['--switch-bg-no-checked' as any]: secondary
      ? '#565656'
      : colors[customColors?.backgroundNoChecked ?? 'disabled'],
    ['--switch-bg-checked' as any]:
      colors[customColors?.backgroundChecked ?? 'success'],
    ['--switch-knob-color' as any]:
      colors[customColors?.knob || 'backgroundDark'],
  };

  return (
    <div
      className={cx(
        styles.switch__container,
        disabled && styles['switch__container--disabled'],
      )}
      style={style}
    >
      <input
        id={id}
        disabled={disabled}
        className={cx(
          styles.switch,
          styles['switch--shadow'],
          disabled && styles['switch--disabled'],
        )}
        type="checkbox"
        checked={isChecked}
        onChange={(val) => onChange && onChange(val)}
      />
      <label htmlFor={id}></label>
    </div>
  );
};
