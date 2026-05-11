import cx from 'classnames';

import { LazyImage } from '~/components/Lazy/LazyImage';
import { FadeAnimation } from '~components/Animations/FadeAnimation';
import { Typography } from '~components/Typography';
import check from '~icons/check.svg';
import colors from '~styles/colors';
import { ColorType } from '~types/general';

import styles from './styles.module.scss';

export type AnimationState = 'loading' | 'default' | 'checked';

const Animation = ({
  width,
  height,
  state,
  backgroundColor,
  padding,
}: {
  width?: string | number;
  height?: string | number;
  state?: AnimationState;
  backgroundColor?: string;
  padding?: string;
}) => (
  <div
    className={styles.deadlineLoadingCheck__animation}
    style={{
      width,
      height,
      padding,
    }}
  >
    <div
      className={cx(styles.deadlineLoadingCheck__loadingDiv, {
        [`${styles['deadlineLoadingCheck__loadingDiv--checked']}`]:
          state === 'checked',
      })}
      data-loading={state === 'loading'}
    >
      <div />
      <div />
      <div
        style={{
          backgroundColor,
        }}
      />
    </div>
    <FadeAnimation
      customWrapperClass={styles.deadlineLoadingCheck__checkWrapper}
      isVisible={state === 'checked'}
      customDelay={state === 'checked' || state === 'loading' ? 1500 : 0}
    >
      <LazyImage loading="lazy" src={check} alt={''} />
    </FadeAnimation>
  </div>
);

interface Props {
  state?: AnimationState;
  text?: string;
  colorText?: ColorType;
  width?: string | number;
  height?: string | number;
  backgroundColor?: string;
  padding?: string;
}

export const DeadlineLoadingCheck = ({
  text,
  state = 'default',
  height,
  width,
  backgroundColor,
  colorText = 'textPrimary',
  padding,
}: Props) => {
  if (text) {
    return (
      <div className={styles.deadlineLoadingCheck}>
        <Animation
          width={width}
          backgroundColor={backgroundColor}
          height={height}
          state={state}
          padding={padding}
        />
        {text && <Typography text={text} color={colors[colorText]} />}
      </div>
    );
  }

  return (
    <Animation
      width={width}
      backgroundColor={backgroundColor}
      height={height}
      state={state}
      padding={padding}
    />
  );
};
