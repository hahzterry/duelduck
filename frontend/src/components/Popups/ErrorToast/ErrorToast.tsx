import { CSSProperties, memo, useEffect, useRef, useState } from 'react';
import cx from 'classnames';

import { FadeAnimation } from '~components/Animations/FadeAnimation';
import { TextToast } from '~components/Popups/ErrorToast/components/TextToast';
import { CircleProgress } from '~components/Progresses/CircleProgress';
import { Cross } from '~icons/JsxSvg/Cross';
import { useErrorToastStore } from '~store/errorToastStore';

import styles from './styles.module.scss';

const CloseButton = memo(
  ({
    close,
    haveProgress,
    progress,
  }: {
    close: () => void;
    haveProgress: boolean;
    progress: number;
  }) => {
    const [displayProgress, setDisplayProgress] = useState(progress);
    const requestRef = useRef<number>(0);

    useEffect(() => {
      if (!progress) return;
      const animate = () => {
        setDisplayProgress((prev) => {
          const diff = progress - prev;

          if (Math.abs(diff) < 0.1) return progress;

          return prev + diff * 0.2;
        });

        requestRef.current = requestAnimationFrame(animate);
      };

      requestRef.current = requestAnimationFrame(animate);

      return () => {
        if (requestRef.current) cancelAnimationFrame(requestRef.current);
      };
    }, [progress]);

    return (
      <button
        onClick={close}
        className={cx(styles.closeBtn, {
          [`${styles['closeBtn--haveProgress']}`]: haveProgress,
        })}
      >
        <Cross />
        {haveProgress && (
          <>
            <div className={styles.closeBtn__backgroundProgress} />
            <div className={styles.closeBtn__progress}>
              <CircleProgress
                progress={displayProgress}
                animationOnHover={false}
                baseStroke={2}
              />
            </div>
          </>
        )}
      </button>
    );
  },
);

export const ErrorToast = () => {
  const { dataToast, isShow, close, resetMsg } = useErrorToastStore();
  const [progress, setProgress] = useState(0);
  const isHovered = useRef(false);
  const duration = dataToast.timeClose || 3100;

  useEffect(() => {
    if (!isShow) {
      isHovered.current = false;
      setProgress(0);

      return;
    }

    setProgress(0);
    let elapsed = 0;
    let lastTick = Date.now();

    const interval = setInterval(() => {
      const now = Date.now();

      if (!isHovered.current) {
        elapsed += now - lastTick;

        if (elapsed >= duration) {
          setProgress(100);
          close();

          return;
        }

        setProgress((elapsed / duration) * 100);
      }

      lastTick = now;
    }, 100);

    return () => {
      clearInterval(interval);
    };
  }, [isShow, duration]);

  const customToastStyle = {
    ['--error-toast-bg']: dataToast.backgroundColor,
    ['--error-toast-text']: dataToast.textColor,
    ['--error-toast-accent']: dataToast.accentTextColor,
    ['--error-toast-progress']: dataToast.progressColor,
    ...dataToast.stylesContainer,
  } as CSSProperties;

  const hasTitle = Boolean(dataToast.titleParts?.length);
  const hasDescription = Boolean(dataToast.textParts?.length);

  return (
    <FadeAnimation
      customClassNames={{
        enter: styles.toastEnter,
        enterActive: styles.toastEnterActive,
        exit: styles.toastExit,
        exitActive: styles.toastExitActive,
      }}
      customWrapperClass={cx(styles.toast__container, {
        [`${styles[`${dataToast.positionClose}`]}`]: dataToast.positionClose,
      })}
      onExited={resetMsg}
      isVisible={isShow}
      style={customToastStyle}
      onMouseEnter={() => {
        isHovered.current = true;
      }}
      onMouseLeave={() => {
        isHovered.current = false;
      }}
    >
      <div className={styles.toast} style={dataToast.stylesToast}>
        <div className={styles.content}>
          {hasTitle && (
            <div className={styles.title}>
              <TextToast parts={dataToast.titleParts} />
            </div>
          )}
          {hasDescription && (
            <div className={styles.description}>
              <TextToast parts={dataToast.textParts} />
            </div>
          )}
          {dataToast.additionalContent}
        </div>
        {(dataToast.positionClose === 'inside-right' ||
          typeof dataToast.positionClose === 'undefined') && (
          <CloseButton
            close={close}
            progress={progress}
            haveProgress={dataToast.progressType === 'cross-circle'}
          />
        )}
        {dataToast.progressType === 'line' && (
          <div className={styles.progressWrapper}>
            <div
              className={styles.progressBar}
              style={{ width: `${progress}%` }}
            />
          </div>
        )}
      </div>
      {dataToast.positionClose === 'outside-bottom' && (
        <CloseButton
          close={close}
          progress={progress}
          haveProgress={dataToast.progressType === 'cross-circle'}
        />
      )}
    </FadeAnimation>
  );
};
