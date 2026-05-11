import { CSSProperties, useEffect, useRef, useState } from 'react';
import cx from 'classnames';

import { Typography } from '~components/Typography';
import colors from '~styles/colors';

import styles from './styles.module.scss';

interface Props {
  text: string;
  onClick?: () => void;
  mainColor?: string;
  progress?: number;
  onMouseEnter?: () => void;
  onMouseLeave?: () => void;
  onMouseMove?: () => void;
  disabled?: boolean;
}

export const ProgressButton = ({
  mainColor,
  onClick,
  text,
  progress,
  onMouseMove,
  onMouseLeave,
  onMouseEnter,
  disabled,
}: Props) => {
  const [animatedProgress, setAnimatedProgress] = useState(progress ?? 0);
  const requestRef = useRef<number | null>(null);
  const progressRef = useRef(progress ?? 0);

  const animate = () => {
    const target = progress ?? 0;
    const diff = target - progressRef.current;

    if (Math.abs(diff) > 0.1) {
      progressRef.current += diff / 10;
      setAnimatedProgress(progressRef.current);
      requestRef.current = requestAnimationFrame(animate);
    } else {
      progressRef.current = target;
      setAnimatedProgress(target);
    }
  };

  useEffect(() => {
    requestRef.current = requestAnimationFrame(animate);

    return () => {
      if (requestRef.current) cancelAnimationFrame(requestRef.current);
    };
  }, [progress]);

  const noProgress = typeof progress === 'undefined' || disabled;

  const displayProgress = noProgress ? 0 : animatedProgress;

  return (
    <div
      className={cx(styles.progressButton, {
        [`${styles['progressButton--progress']}`]:
          typeof progress !== 'undefined',
      })}
      style={
        {
          '--main-button-color': disabled ? colors.textSecondary : mainColor,
        } as CSSProperties
      }
      data-disabled={!!disabled}
      onClick={onClick}
      onMouseLeave={onMouseLeave}
      onMouseMove={onMouseMove}
      onMouseEnter={onMouseEnter}
    >
      <div className={styles.progressButton__backgroundContainer}>
        <div
          className={styles.progressButton__background}
          style={{
            transform: `scaleX(${noProgress ? 100 : displayProgress / 100})`,
            transformOrigin: 'left top',
          }}
        />
      </div>
      <Typography className={styles.progressButton__text} text={text} />
      <div
        className={styles.progressButton__textProgressContainer}
        style={{
          clipPath: `inset(0 ${100 - displayProgress}% 0 0)`,
        }}
      >
        <Typography className={styles.progressButton__text} text={text} />
      </div>
    </div>
  );
};
