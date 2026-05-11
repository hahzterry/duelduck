import { CSSProperties, ReactNode, useState } from 'react';
import cx from 'classnames';

import { AnimatedChangeTextComponent } from '~components/Animations/AnimatedChangeTextComponent';
import { Typography } from '~components/Typography';
import { Duck } from '~icons/JsxSvg/Duck';
import colors from '~styles/colors';
import { ColorType } from '~types/general';

import styles from './styles.module.scss';

interface PointsWithIconProps {
  text?: string | ReactNode;
  colorText?: ColorType;
  isHaveCheck?: boolean;
  isActiveCheck?: boolean;
  bgColor?: ColorType;
  withMinus?: boolean;
  withPlus?: boolean;
  width?: string;
  isTextPositionAbsolute?: boolean;
  icon?: ReactNode;
  haveImageBackground?: boolean;
  isBungee?: boolean;
  isExport?: boolean;
  className?: string;
  hoverText?: string;
  haveAnimationOnHoverText?: boolean;
  iconAroundColor?: string;
}

export const PointsWithIcon = ({
  text = '0',
  colorText,
  isHaveCheck,
  isActiveCheck,
  bgColor,
  withPlus,
  withMinus,
  width,
  icon,
  isTextPositionAbsolute,
  isBungee,
  className,
  isExport,
  iconAroundColor,
  hoverText,
  haveAnimationOnHoverText,
}: PointsWithIconProps) => {
  const [isHovered, setHovered] = useState(false);
  const textLocal = isHovered ? (hoverText ?? text) : text;

  return (
    <div
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      onMouseMove={() => setHovered(true)}
      className={cx(
        styles.pointsWithIcon,
        {
          [`${styles['pointsWithIcon--isTextPositionAbsolute']}`]:
            isTextPositionAbsolute,
        },
        className,
      )}
      data-export={isExport}
      style={
        {
          background: bgColor ? colors[bgColor as ColorType] : undefined,
          width: width || 'max-content',
        } as CSSProperties
      }
    >
      <div
        className={styles.pointsWithIcon__icon}
        style={{
          background: iconAroundColor
            ? iconAroundColor
            : bgColor
              ? colors[bgColor as ColorType]
              : undefined,
        }}
      >
        {icon ? icon : <Duck />}
      </div>
      {typeof textLocal === 'string' ? (
        haveAnimationOnHoverText ? (
          <AnimatedChangeTextComponent
            text={`${withMinus ? '-' : ''}${withPlus ? '+' : ''}${textLocal}`}
            font={{
              fontFamily: 'Bungee',
              fontSize: '14px',
              fontWeight: '400',
            }}
          />
        ) : (
          <Typography
            isBungee={isBungee}
            color={colors[(colorText || '') as ColorType]}
            text={`${withMinus ? '-' : ''}${withPlus ? '+' : ''}${text}`}
          />
        )
      ) : (
        textLocal
      )}
      {isHaveCheck && (
        <svg
          version="1.1"
          xmlns="http://www.w3.org/2000/svg"
          viewBox="0 0 130.2 130.2"
        >
          <circle
            fill="none"
            stroke={`${isActiveCheck ? '#b0d356' : '#a7a7a7'}`}
            strokeWidth="7"
            strokeMiterlimit="10"
            cx="65.1"
            cy="65.1"
            r="62.1"
          />
          {isActiveCheck && (
            <polyline
              fill="none"
              stroke="#b0d356"
              strokeWidth="8"
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeMiterlimit="10"
              points="100.2,40.2 51.5,88.8 29.8,67.5 "
            />
          )}
        </svg>
      )}
    </div>
  );
};
