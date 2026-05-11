'use client';
import { ChangeEventHandler, memo, useEffect, useRef, useState } from 'react';
import cx from 'classnames';
import Image from 'next/image';

import { FadeAnimation } from '~components/Animations/FadeAnimation';
import { SwitchAnimation } from '~components/Animations/SwitchAnimation';
import { Typography } from '~components/Typography';
import { regexLatinAndSymbols } from '~constants/main';
import check from '~icons/check.svg';
import { EditIcon } from '~icons/JsxSvg/EditIcon';
import { Tail } from '~icons/JsxSvg/Tail';
import colors from '~styles/colors';
import { validateUsername } from '~utils/stringFormat';

import { SpinAnimation } from '../Animations/SpinAnimation';

import styles from './styles.module.scss';

export type EditState = 'default' | 'canSave' | 'saving';

const Save = memo(
  ({ onClick, isError }: { onClick: () => void; isError: boolean }) => {
    return (
      <div
        className={styles.mainContent__check}
        onClick={onClick}
        aria-label="Save image"
        style={{
          backgroundColor: isError ? colors.textSecondary : colors.primary,
          cursor: isError ? 'not-allowed' : 'pointer',
        }}
      >
        <Image src={check} width={24} height={24} alt="Edit" />
      </div>
    );
  },
);

const IconSave = memo(
  ({
    editState,
    onClick,
    error,
  }: {
    editState: EditState;
    onClick: () => void;
    error: string;
  }) => {
    const contentMap = {
      ['default']: <EditIcon onClick={onClick} />,
      ['canSave']: <Save isError={!!error} onClick={onClick} />,
      ['saving']: <SpinAnimation key="saving" />,
    };

    return (
      <SwitchAnimation
        switchKey={`${editState}`}
        customWrapperClass={cx(styles.mainContent__editButtonContainer, {
          [`${styles['mainContent__editButtonContainer--canHover']}`]:
            editState === 'default' || (!error && editState === 'canSave'),
        })}
      >
        {contentMap[editState]}
      </SwitchAnimation>
    );
  },
);

type UsernameProps = {
  currentUsername?: string;
  value: string;
  editState: EditState;
  error: string;

  onChange: (value: string, error: string) => void;

  onEditStateChange: (state: EditState) => void;

  onSave: (value: string) => Promise<void | string> | (void | string);

  validate?: (value: string) => string;

  classNameContainer?: string;

  beforeValueSymbol?: string;

  usernamePattern?: RegExp;
};

export const EditableInlineField = memo(
  ({
    currentUsername,
    value,
    editState,
    error,
    onChange,
    onEditStateChange,
    onSave,
    beforeValueSymbol,
    classNameContainer,
    validate = validateUsername,
    usernamePattern = regexLatinAndSymbols,
  }: UsernameProps) => {
    const inputRef = useRef<HTMLInputElement>(null);
    const mirrorRef = useRef<HTMLSpanElement>(null);

    const [showError, setShowError] = useState<string>('');

    useEffect(() => {
      if (error) setShowError(error);
    }, [error]);

    const onExited = () => setShowError('');

    const handleChange: ChangeEventHandler<HTMLInputElement> = (e) => {
      const v = e.target.value;

      if (v.length <= 24) {
        const nextError = validate(v);

        onChange(usernamePattern.test(v) ? v : value, nextError);
      }
    };

    const handleSave = async () => {
      if (!error) {
        try {
          onEditStateChange('saving');
          await onSave(value.trim());
          onEditStateChange('default');
          setTimeout(() => onChange('', ''), 300);
        } catch (e: any) {
          const data = e?.response?.data;
          const msg =
            typeof data === 'string'
              ? data
              : (data?.message ?? 'Error network');

          onChange(value, String(msg));
          onEditStateChange('default');
        }
      }
    };

    // автофокус та підгонка ширини інпуту під контент
    useEffect(() => {
      if (mirrorRef.current && inputRef.current) {
        mirrorRef.current.textContent = value || '';
        const length = mirrorRef.current.getBoundingClientRect().width;

        inputRef.current.style.setProperty('--input-length', `${length}px`);
        if (editState === 'canSave') inputRef.current.focus();
      }
    }, [value, editState]);

    return (
      <div className={styles.mainContent__containerUserName}>
        <FadeAnimation
          isVisible={!!error}
          onExited={onExited}
          customClassNames={{
            enter: styles.fadeTextEnter,
            enterActive: styles.fadeTextEnterActive,
            exit: styles.fadeTextExit,
            exitActive: styles.fadeTextExitActive,
          }}
          customWrapperClass={styles.mainContent__error}
        >
          <Tail />
          <div className={styles.mainContent__errorMsg}>
            <Typography text={showError} />
          </div>
        </FadeAnimation>

        <div
          className={cx(
            styles.mainContent__userName,
            {
              [`${styles['mainContent__userName--visibleInput']}`]:
                editState === 'canSave',
            },
            classNameContainer,
          )}
        >
          <span>
            {beforeValueSymbol ?? <></>}
            {editState === 'canSave' ? (
              <>
                <span
                  ref={mirrorRef}
                  className={styles.mainContent__inputMirror}
                  aria-hidden="true"
                />
                <input
                  ref={inputRef}
                  className={styles.mainContent__input}
                  name="username"
                  value={value}
                  onChange={handleChange}
                />
              </>
            ) : (
              <Typography text={`${value || currentUsername || ''}`} />
            )}
          </span>

          <IconSave
            editState={editState}
            error={error}
            onClick={
              editState === 'canSave'
                ? handleSave
                : () => {
                    onChange(
                      currentUsername || '',
                      validate(currentUsername || ''),
                    );
                    onEditStateChange('canSave');
                  }
            }
          />
        </div>
      </div>
    );
  },
);
