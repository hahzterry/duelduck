import './styles.scss';

import React, { FormEvent, useState } from 'react';
import cx from 'classnames';

import { Typography } from '~components/Typography';
import { ArrowRight } from '~icons/JsxSvg/ArrowRight';
import colors from '~styles/colors';

interface RoundedInputProps {
  placeholder?: string;
  value?: string;
  onChange?: React.ChangeEventHandler<HTMLInputElement>;
  onSubmit?: React.FormEventHandler<HTMLFormElement>;
  isFullWidth?: boolean;
  isError?: boolean;
  errorMsg?: string;
}

export const RoundedInput = ({
  placeholder,
  value,
  onChange,
  isFullWidth,
  errorMsg,
  isError,
  onSubmit,
}: RoundedInputProps) => {
  const [hasText, setHasText] = useState(false);

  const handleInput = (event: FormEvent<HTMLInputElement>) => {
    event.currentTarget.value = event.currentTarget.value.replace(
      /[^A-Za-z0-9.@]/g,
      '',
    );
  };

  return (
    <form
      onSubmit={onSubmit}
      className={cx({
        ['containerInput']: true,
        ['containerInput--fullWidth']: isFullWidth,
      })}
    >
      <input
        className={cx({
          ['containerInput__roundedInput']: true,
          ['containerInput__roundedInput--fullWidth']: isFullWidth,
          ['containerInput__roundedInput--inputError']: isError,
        })}
        placeholder={placeholder}
        type="email"
        onInput={handleInput}
        value={value}
        onChange={(e) => {
          setHasText(!!e.target.value);
          if (onChange) onChange(e);
        }}
      />
      <button
        type="submit"
        disabled={isError}
        className={cx({
          ['containerInput__arrowRight']: true,
          ['containerInput__arrowRight--arrowRightYellow']: hasText && !isError,
        })}
      >
        <ArrowRight />
      </button>
      {errorMsg ? (
        <Typography
          color={isError ? colors['error'] : ''}
          className={'containerInput__errorMsg'}
          variant="subtitle"
          element="span"
          fontWeight="m"
          lineHeight={1.2}
          text={errorMsg as string}
        />
      ) : (
        <></>
      )}
    </form>
  );
};
