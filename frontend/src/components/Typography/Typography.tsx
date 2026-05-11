import { forwardRef, JSX, ReactNode } from 'react';
import cx from 'classnames';

import { Styles } from '~types/general';

import styles from './Typography.module.scss';

export type TypographyVariants =
  | 'subtitle'
  | 'body'
  | 'body-small'
  | 'body-big'
  | 'title-1'
  | 'title-2'
  | 'title-big'
  | 'h-1'
  | 'h-2'
  | 'h-3';

export const fontWeightsClassesMap = {
  r: 'regular',
  m: 'medium',
  sb: 'semibold',
  b: 'bold',
} as const;

export type FontWeightsValues = keyof typeof fontWeightsClassesMap;

interface TypographyProps {
  text?: string;
  element?: keyof JSX.IntrinsicElements;
  level?: string;
  lineHeight?: number;
  fontWeight?: FontWeightsValues;
  textTransform?: 'uppercase' | 'lowercase' | 'capitalize' | 'none';
  textAlign?: 'left' | 'center' | 'right' | 'start';
  color?: string;
  customStyles?: Styles;
  className?: string;
  variant?: TypographyVariants;
  isBungee?: boolean;
  onClick?: (e: MouseEvent) => void;
  children?: ReactNode;
  noSelection?: boolean;
  ellipsis?: boolean;
  id?: string;
  onMouseEnter?: () => void;
  onMouseLeave?: () => void;
}

const pxToRem = (px: number): string => `${px / 16}rem`;

export const Typography = forwardRef<
  HTMLSpanElement | HTMLHeadingElement | any,
  TypographyProps
>(
  (
    {
      element = 'span',
      text,
      level,
      fontWeight,
      lineHeight,
      color,
      textTransform,
      textAlign,
      customStyles = {},
      variant,
      className,
      isBungee,
      onClick,
      children,
      noSelection,
      ellipsis,
      id,
      onMouseEnter,
      onMouseLeave,
      ...props
    },
    ref,
  ) => {
    const Element: keyof JSX.IntrinsicElements = element; // Ensure correct element type

    const baseStyles: Styles = {};

    if (level) {
      const [fontSizeStr, lineHeightStr] = level.split('-');

      if (fontSizeStr) baseStyles.fontSize = pxToRem(+fontSizeStr);
      if (lineHeightStr) {
        baseStyles.lineHeight = pxToRem(parseFloat(lineHeightStr));
      }
    }

    if (lineHeight) {
      baseStyles.lineHeight = lineHeight;
    }

    if (color) {
      baseStyles.color = color;
    }

    if (textTransform) {
      baseStyles.textTransform = textTransform;
    }

    if (textAlign) {
      baseStyles.textAlign = textAlign;
    }

    const combinedClassName = cx(
      styles.typography,
      className,
      fontWeight && styles[fontWeightsClassesMap[fontWeight]],
      isBungee && styles.bungee,
      variant && styles[variant],
      noSelection && styles.noSelection,
      ellipsis && styles.ellipsis,
    );

    const handleClick = (e: MouseEvent) => {
      if (onClick) {
        onClick(e);
      }
    };

    return (
      <Element
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore
        ref={ref}
        className={combinedClassName}
        id={id}
        style={{ ...baseStyles, ...customStyles }}
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore
        onClick={(e: MouseEvent) => handleClick(e)}
        onMouseEnter={onMouseEnter}
        onMouseLeave={onMouseLeave}
        {...props}
      >
        {text || children}
      </Element>
    );
  },
);
