import './styles.scss';

import { ChangeEvent, FocusEventHandler } from 'react';
import cx from 'classnames';
import { Property } from 'csstype';

interface TextAreaProps {
  value?: string;
  onChange?: (value: string) => void;
  placeholder?: string;
  resize?: Property.Resize;
  rows?: number;
  maxText?: number;
  className?: string;
  onFocus?: FocusEventHandler<HTMLTextAreaElement>;
  onBlur?: FocusEventHandler<HTMLTextAreaElement>;
}

export const TextArea = ({
  value = '',
  onChange,
  placeholder = '',
  resize,
  rows,
  maxText = 100000,
  className = '',
  onFocus,
  onBlur,
}: TextAreaProps) => {
  const handleChange = (event: ChangeEvent<HTMLTextAreaElement>) => {
    const newValue = event.target.value;

    if (newValue.length <= maxText) {
      onChange?.(event.target.value);
    }
  };

  return (
    <textarea
      value={value}
      onChange={handleChange}
      placeholder={placeholder}
      onFocus={onFocus}
      onBlur={onBlur}
      className={cx('customTextArea', className)}
      style={{ resize }}
      rows={rows}
    />
  );
};
