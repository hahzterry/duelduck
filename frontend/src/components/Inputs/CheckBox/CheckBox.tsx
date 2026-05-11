import { ChangeEventHandler, RefObject } from 'react';
import Image from 'next/image';

import check from '~icons/check.svg';

import styles from './styles.module.scss';

interface CheckboxProps {
  checked: boolean;
  onChange?: ChangeEventHandler<HTMLInputElement>;
  ref?: RefObject<HTMLLabelElement | null>;
}

export const Checkbox = ({
  checked,
  onChange = () => {},
  ref,
}: CheckboxProps) => {
  return (
    <label ref={ref} className={styles.checkboxContainer}>
      <input
        type="checkbox"
        checked={checked}
        onChange={onChange}
        className={styles.checkboxInput}
      />
      <Image
        src={check}
        width={18}
        height={18}
        alt={`${checked ? '' : 'No '} checked`}
      />
    </label>
  );
};
