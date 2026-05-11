import {
  ChangeEventHandler,
  CSSProperties,
  FocusEventHandler,
  HTMLInputTypeAttribute,
  KeyboardEventHandler,
  MouseEventHandler,
  ReactNode,
  RefObject,
  useEffect,
  useRef,
  useState,
} from 'react';
import cx from 'classnames';

import { FadeAnimation } from '~components/Animations/FadeAnimation';
import { Cross } from '~icons/JsxSvg/Cross';

import styles from './style.module.scss';

interface BaseInputProps {
  placeholder?: string;
  value?: number | string;
  onChange?: ChangeEventHandler<HTMLInputElement | HTMLTextAreaElement>;
  onFocus?: FocusEventHandler<HTMLInputElement | HTMLTextAreaElement>;
  onBlur?: FocusEventHandler<HTMLInputElement | HTMLTextAreaElement>;
  isFullWidth?: boolean;
  type?: HTMLInputTypeAttribute;
  customStyles?: CSSProperties;
  haveCross?: boolean;
  onClickCross?: () => void;
  height?: string;
  inputMode?:
    | 'email'
    | 'search'
    | 'tel'
    | 'text'
    | 'url'
    | 'numeric'
    | 'none'
    | 'decimal';
  pattern?: string;
  lastIcon?: ReactNode;
  onKeyDown?: KeyboardEventHandler<HTMLInputElement | HTMLTextAreaElement>;
  min?: number | string;
  max?: number | string;
  className?: string;
  error?: string;
  haveError?: boolean;
  startIcon?: ReactNode;
  startIconClassName?: string;
  lastIconClassName?: string;
  lastIconRefContainer?: RefObject<HTMLDivElement | null>;
  onMouseLeave?: MouseEventHandler<HTMLDivElement>;
  onMouseEnter?: MouseEventHandler<HTMLDivElement>;
  id?: string;
  isTextArea?: boolean;
}

const ERROR_ANIMATION_DURATION = 300;

/**
 * Hook to autosize textarea based on content
 */
const useAutosizeTextarea = (value: string | number | undefined) => {
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    const el = textareaRef.current;

    if (!el) return;
    el.style.height = '0px';
    const scrollHeight = el.scrollHeight;

    el.style.height = `${scrollHeight}px`;
  }, [value]);

  return textareaRef;
};

export const BaseInput = ({
  placeholder,
  type,
  isFullWidth,
  height,
  error,
  value,
  onChange,
  onFocus,
  onBlur,
  customStyles,
  haveCross,
  onClickCross,
  inputMode,
  pattern,
  id,
  lastIcon,
  onKeyDown,
  max,
  min,
  haveError,
  className,
  startIcon,
  startIconClassName,
  lastIconClassName,
  lastIconRefContainer,
  onMouseLeave,
  onMouseEnter,
  isTextArea,
}: BaseInputProps) => {
  const [displayedError, setDisplayedError] = useState<string | undefined>(
    error,
  );
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | undefined>(null);
  const textareaRef = useAutosizeTextarea(value);

  useEffect(() => {
    if (timeoutRef.current) clearTimeout(timeoutRef.current);

    if (error) {
      setDisplayedError(error);
    } else {
      timeoutRef.current = setTimeout(() => {
        setDisplayedError(undefined);
      }, ERROR_ANIMATION_DURATION);
    }

    return () => {
      if (timeoutRef.current) clearTimeout(timeoutRef.current);
    };
  }, [error]);

  return (
    <div
      className={cx(
        styles['containerBaseInput'],
        {
          [styles['containerBaseInput--fullWidth'] as string]: isFullWidth,
        },
        className,
      )}
      onMouseLeave={onMouseLeave}
      onMouseEnter={onMouseEnter}
      id={id}
      style={{ height }}
    >
      <div>
        {startIcon && (
          <div
            className={cx(
              styles['containerBaseInput__startIcon'],
              startIconClassName,
            )}
            onClick={onClickCross}
            id="input-base-start-icon"
          >
            {startIcon}
          </div>
        )}

        {isTextArea ? (
          <textarea
            ref={textareaRef}
            className={cx(styles['containerBaseInput__roundedInput'])}
            placeholder={placeholder}
            value={value}
            onFocus={onFocus}
            inputMode={inputMode}
            onKeyDown={onKeyDown as KeyboardEventHandler<HTMLTextAreaElement>}
            onBlur={onBlur}
            onChange={onChange as ChangeEventHandler<HTMLTextAreaElement>}
            style={{
              ...customStyles,
              overflow: 'hidden',
              resize: 'none',
            }}
            rows={1}
          />
        ) : (
          <input
            className={cx(styles['containerBaseInput__roundedInput'])}
            placeholder={placeholder}
            type={type}
            value={value}
            onFocus={onFocus}
            inputMode={inputMode}
            pattern={pattern}
            onKeyDown={onKeyDown as KeyboardEventHandler<HTMLInputElement>}
            min={min}
            max={max}
            onBlur={onBlur}
            onChange={onChange as ChangeEventHandler<HTMLInputElement>}
            style={{
              ...customStyles,
              ...(!(lastIcon || haveCross)
                ? {
                    height: 'unset',
                  }
                : {}),
            }}
          />
        )}
        {(lastIcon || haveCross) && (
          <div
            className={cx(
              styles['containerBaseInput__cross'],
              lastIconClassName,
            )}
            onClick={onClickCross}
            data-last-icon={!!lastIcon}
            ref={lastIconRefContainer}
          >
            {lastIcon ?? (haveCross && <Cross />)}
          </div>
        )}
      </div>

      <FadeAnimation
        isVisible={!!haveError}
        customWrapperClass={styles.containerBaseInput__errorText}
      >
        {displayedError}
      </FadeAnimation>
    </div>
  );
};
