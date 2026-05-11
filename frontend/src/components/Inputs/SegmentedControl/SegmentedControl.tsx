import { CSSProperties, useEffect, useRef, useState } from 'react';

import { Typography } from '~components/Typography';
import colors from '~styles/colors';

import styles from './styles.module.scss';

export interface SegmentedControlElement<T> {
  label: string;
  value: T;
  color: string;
  colorActive: string;
}

interface Props<T> {
  elements: SegmentedControlElement<T>[];
  onChange: (value: T) => void;
  active: T;
}

export const SegmentedControl = <
  T extends string | number | readonly string[],
>({
  elements,
  active,
  onChange,
}: Props<T>) => {
  const refContainer = useRef<HTMLDivElement | null>(null);
  const [stylesActiveBg, setStylesActiveBg] = useState<{
    width: number;
    transformLeft: number;
  } | null>();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    const el = refContainer.current;

    if (el) {
      const observer = new ResizeObserver((e) => {
        const element = e[0]?.target;

        if (element) {
          const widthContainer = (element as HTMLElement).offsetWidth || 0;
          const width = widthContainer / elements.length;

          const index = elements.findIndex((e) => e.value === active);

          const indexSave = index === -1 ? 0 : index;

          setStylesActiveBg({
            width,
            transformLeft: width * indexSave,
          });
          setTimeout(() => {
            setMounted(true);
          });
        }
      });

      observer.observe(el);

      return () => {
        observer.unobserve(el);
      };
    }

    return () => {};
  }, [elements, active]);

  return (
    <div className={styles.segmentedControl} ref={refContainer}>
      <div
        style={
          {
            ['--active-color']:
              elements.find((e) => e.value === active)?.colorActive ||
              colors.primary,
            width: stylesActiveBg?.width || 0,
            transform: `translateX(${stylesActiveBg?.transformLeft}px)`,
          } as CSSProperties
        }
        className={styles.segmentedControl__activeBg}
      />
      <div className={styles.segmentedControl__defaultElements}>
        {elements.map((el, i) => (
          <div
            className={styles.segmentedControl__element}
            key={i}
            data-active={false}
            onClick={() => onChange(el.value)}
            style={
              {
                ['--active-color']: el.colorActive,
              } as CSSProperties
            }
          >
            <Typography text={el.label} />
          </div>
        ))}
      </div>
      <div
        className={styles.segmentedControl__activeElements}
        style={{
          transition: mounted ? undefined : 'none',
          clipPath: `inset(0% calc(100% - ${(stylesActiveBg?.width || 0) + (stylesActiveBg?.transformLeft || 0)}px) 0% ${stylesActiveBg?.transformLeft || 0}px)`,
        }}
      >
        {elements.map((el, i) => (
          <div
            className={styles.segmentedControl__element}
            key={i}
            data-active={true}
            onClick={() => onChange(el.value)}
          >
            <Typography text={el.label} />
          </div>
        ))}
      </div>
    </div>
  );
};
