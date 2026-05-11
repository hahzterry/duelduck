import {
  CSSProperties,
  MouseEventHandler,
  ReactNode,
  useLayoutEffect,
  useRef,
  useState,
} from 'react';
import { createPortal } from 'react-dom';
import cx from 'classnames';

import { FadeAnimation } from '~components/Animations/FadeAnimation';
import { Typography } from '~components/Typography';
import { useOuterClick } from '~hooks/useOuterClick';
import colors from '~styles/colors';

import styles from './styles.module.scss';

type Variant = 'top' | 'bottom';
type VariantHorizontal = 'left' | 'right' | 'center';

export const PositionPopup = ({
  children,
  className,
  textPopup,
  variant = 'top',
  isShow,
  onMouseLeave,
  onMouseEnter,
  popupClassName,
  horizontalPosition,
  onOutsideClick,
  onClick,
  colorBorder = colors.strokeOpacity,
}: {
  children: ReactNode;
  className?: string;
  textPopup?: string | ReactNode;
  variant?: Variant;
  isShow: boolean;
  onMouseEnter?: MouseEventHandler<HTMLDivElement>;
  onClick?: MouseEventHandler<HTMLDivElement>;
  onMouseLeave?: MouseEventHandler<HTMLDivElement>;
  popupClassName?: string;
  onOutsideClick?: () => void;
  horizontalPosition?: VariantHorizontal;
  colorBorder?: string;
}) => {
  const anchorRef = useRef<HTMLDivElement | null>(null);
  const [coords, setCoords] = useState<{ top: number; left: number } | null>(
    null,
  );
  const roRef = useRef<ResizeObserver | null>(null);
  const popupRef = useRef<HTMLDivElement | null>(null);

  const updateCoords = () => {
    if (!anchorRef?.current || !isShow) return;
    const rect = anchorRef.current.getBoundingClientRect();
    const widthPopup = popupRef.current?.getBoundingClientRect()?.width || 0;

    const divider = {
      left: 1,
      right: rect.width - widthPopup,
      center: rect.width / 2 - widthPopup / 2,
    };

    const left = rect.left + divider[horizontalPosition || 'center'];

    const screenWidth = window.innerWidth - 16;

    const maxLeft =
      screenWidth > left + widthPopup ? left : screenWidth - widthPopup - 8;

    const minLeft = 16 > maxLeft ? 16 : maxLeft;

    const top = variant === 'bottom' ? rect.bottom : rect.top;

    setCoords({ top, left: minLeft });
  };

  useLayoutEffect(() => {
    if (isShow) updateCoords();

    const onScroll = () => {
      updateCoords();
    };

    document.body.addEventListener('scroll', onScroll, true);
    window.addEventListener('resize', onScroll);

    if (anchorRef?.current) {
      roRef.current = new ResizeObserver(() => updateCoords());
      roRef.current.observe(anchorRef.current);
    }

    return () => {
      document.body.removeEventListener('scroll', onScroll, true);
      window.removeEventListener('resize', onScroll);
      if (roRef.current && anchorRef?.current) {
        roRef.current.unobserve(anchorRef.current);
        roRef.current.disconnect();
        roRef.current = null;
      }
    };
  }, [isShow, variant, anchorRef]);

  const portalContent = (
    <FadeAnimation
      isVisible={isShow}
      customRef={popupRef}
      customWrapperClass={cx(
        styles.container__popup,
        styles[`container__popup--${variant}`],
        popupClassName,
      )}
      customClassNames={{
        enter: styles.fadeEnter,
        enterActive: styles.fadeEnterActive,
        exit: styles.fadeExit,
        exitActive: styles.fadeExitActive,
      }}
      style={
        {
          ...(coords
            ? {
                top: coords.top,
                left: coords.left,
              }
            : {}),
          ['--color-border']: colorBorder,
        } as CSSProperties
      }
    >
      {typeof textPopup === 'string' ? (
        <Typography text={textPopup} />
      ) : (
        textPopup
      )}
    </FadeAnimation>
  );

  useOuterClick(anchorRef, () => onOutsideClick?.(), 'click');

  return (
    <div
      className={cx(styles.container, className)}
      onMouseEnter={onMouseEnter}
      ref={anchorRef}
      onMouseLeave={onMouseLeave}
      onClick={onClick}
    >
      {children}
      {typeof window !== 'undefined'
        ? createPortal(portalContent, document.body)
        : null}
    </div>
  );
};
