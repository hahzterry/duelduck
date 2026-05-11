import { CSSProperties, ReactNode, useState } from 'react';
import cx from 'classnames';

import { PositionPopup } from '~components/Popups/PositionPopup';
import useMediaQuery from '~hooks/useMediaQuery';
import { InfoBorder } from '~icons/JsxSvg/InfoBorder';
import { InfoBorderBackground } from '~icons/JsxSvg/InfoBorderBackground';
import { Warning } from '~icons/JsxSvg/Warning';
import colors from '~styles/colors';

import styles from './styles.module.scss';

const icons = {
  ['info']: <InfoBorder />,
  ['info-background']: <InfoBorderBackground />,
  ['warning']: <Warning />,
};

export const Tooltip = ({
  text,
  triggerClassName = '',
  horizontalPosition,
  type = 'info',
  hovered: hoveredProps,
  color = colors.textSecondary,
  colorBorder = colors.strokeOpacity,
  onClose,
  customTrigger,
}: {
  text: string | ReactNode;
  triggerClassName?: string;
  horizontalPosition?: 'left' | 'right' | 'center';
  type?: 'info' | 'warning' | 'info-background';
  color?: string;
  colorBorder?: string;
  hovered?: boolean;
  onClose?: () => void;
  customTrigger?: ReactNode;
}) => {
  const [hovered, setHovered] = useState(false);
  const { isSmallTablet } = useMediaQuery();

  return (
    <PositionPopup
      isShow={typeof hoveredProps !== 'undefined' ? hoveredProps : hovered}
      textPopup={<div className={styles.tooltip}>{text}</div>}
      variant={'bottom'}
      horizontalPosition={horizontalPosition}
      onOutsideClick={() => {
        onClose?.();
        setHovered(false);
      }}
      colorBorder={colorBorder}
      onMouseEnter={() => {
        if (!isSmallTablet) setHovered(true);
      }}
      onClick={(e) => {
        if (isSmallTablet) {
          e.stopPropagation();

          setHovered((p) => !p);
        }
      }}
      onMouseLeave={() => {
        onClose?.();
        if (!isSmallTablet) {
          setHovered(false);
        }
      }}
    >
      {customTrigger ?? (
        <div
          className={cx(styles.trigger, triggerClassName)}
          style={
            {
              ['--color-icon']: color,
            } as CSSProperties
          }
        >
          {icons[type]}
        </div>
      )}
    </PositionPopup>
  );
};
