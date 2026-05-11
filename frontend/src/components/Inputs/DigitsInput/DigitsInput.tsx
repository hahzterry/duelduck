import './styles.scss';

import {
  ChangeEvent,
  ClipboardEvent,
  KeyboardEvent,
  useRef,
  useState,
} from 'react';
import cx from 'classnames';

import { Typography } from '~components/Typography';
import colors from '~styles/colors';

interface DigitsInputProps {
  onChange?: (value: string) => void;
  errorMsg?: string;
  isError?: boolean;
  length?: number;
  inputClass?: string;
}

export const DigitsInput = ({
  onChange,
  errorMsg,
  isError,
  length = 6,
  inputClass,
}: DigitsInputProps) => {
  const inputRefs = useRef<HTMLInputElement[]>([]);
  const [hasNumber, setHasNumber] = useState<boolean[]>(
    Array(length).fill(false),
  );

  const buildValueFromRefs = () =>
    inputRefs.current.map((ref) => ref?.value || '').join('');

  const handleInputChange = (
    index: number,
    e: ChangeEvent<HTMLInputElement>,
  ) => {
    const input = e.target;

    // залишаємо тільки цифри і максимум 1 символ
    const digit = input.value.replace(/\D/g, '').slice(0, 1);

    input.value = digit;

    const newValue = buildValueFromRefs();

    onChange?.(newValue);

    // оновлюємо стейт для класу hasNumber
    setHasNumber((prev) => prev.map((v, i) => (i === index ? !!digit : v)));

    // якщо ввели цифру — фокус на наступний інпут
    if (digit && index < inputRefs.current.length - 1) {
      const next = inputRefs.current[index + 1];

      if (next && next.value === '') {
        next.focus();
      }
    }
  };

  const handleKeyDown = (index: number, e: KeyboardEvent<HTMLInputElement>) => {
    const allowedControlKeys = [
      'Backspace',
      'Tab',
      'ArrowLeft',
      'ArrowRight',
      'Delete',
    ];

    if (e.ctrlKey || e.metaKey) {
      return;
    }

    if (e.key.length === 1 && !/[0-9]/.test(e.key)) {
      e.preventDefault();

      return;
    }

    if (e.key === 'Backspace') {
      const target = e.target as HTMLInputElement;

      if (target.value) {
        target.value = '';
        const newValue = buildValueFromRefs();

        onChange?.(newValue);

        setHasNumber((prev) => prev.map((v, i) => (i === index ? false : v)));
      } else if (index > 0) {
        e.preventDefault();
        const prevInput = inputRefs.current[index - 1];

        if (prevInput) {
          prevInput.focus();
          prevInput.value = '';
        }

        const newValue = buildValueFromRefs();

        onChange?.(newValue);

        setHasNumber((prev) =>
          prev.map((v, i) => (i === index - 1 ? false : v)),
        );
      }
    }

    if (!allowedControlKeys.includes(e.key) && e.key.length > 1) {
      return;
    }
  };

  const handlePaste = (e: ClipboardEvent<HTMLInputElement>) => {
    e.preventDefault();
    const clipboardData = e.clipboardData?.getData('text') || '';

    const digits = clipboardData.replace(/\D/g, '').slice(0, length);

    if (!digits) return;
    for (let i = 0; i < length; i++) {
      const inputRef = inputRefs.current[i];
      const char = digits[i];

      if (!inputRef) continue;

      inputRef.value = char ?? '';
      hasNumber[i] = !!char;
    }

    setHasNumber([...hasNumber]);

    const newValue = buildValueFromRefs();

    onChange?.(newValue);

    const firstEmptyIndex = inputRefs.current.findIndex(
      (ref) => ref && ref.value === '',
    );

    if (firstEmptyIndex !== -1) {
      inputRefs.current[firstEmptyIndex]?.focus();
    }
  };

  return (
    <div className={cx('containerDigitsInput')}>
      <div className={cx('containerDigitsInput__inputs')}>
        {[...Array(length)].map((_, index) => (
          <input
            key={index}
            type="tel"
            inputMode="numeric"
            pattern="\d*"
            placeholder="0"
            className={cx({
              ['containerDigitsInput__input']: true,
              ['containerDigitsInput__input--hasNumber']: hasNumber[index],
              ['containerDigitsInput__input--error']: isError,
              [inputClass as string]: inputClass,
            })}
            maxLength={1}
            // eslint-disable-next-line @typescript-eslint/ban-ts-comment
            // @ts-ignore
            ref={(el) => (inputRefs.current[index] = el!)}
            onChange={(e) => handleInputChange(index, e)}
            onKeyDown={(e) => handleKeyDown(index, e)}
            onPaste={handlePaste}
          />
        ))}
      </div>
      <Typography
        color={isError ? colors.error : ''}
        className={'containerDigitsInput__errorMsg'}
        text={errorMsg}
      />
    </div>
  );
};
