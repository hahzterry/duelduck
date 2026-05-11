import React, {
  memo,
  RefObject,
  useCallback,
  useEffect,
  useImperativeHandle,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from 'react';

import { useRefAndState } from '~hooks/useRefAndState';

import styles from './styles.module.scss';

interface SlotCounterProps {
  text: string;
  triggerOnce?: boolean;
  deleteAfterEndAnimation?: boolean;
  refActionRunAnimation?: RefObject<((triggerOnce?: boolean) => void) | null>;
}

interface ColumnData {
  values: string[];
  shouldAnimate: boolean;
}

export const SlotCounter: React.FC<SlotCounterProps> = memo(
  ({
    text,
    triggerOnce = true,
    refActionRunAnimation,
    deleteAfterEndAnimation: deleteAfterEndAnimationProps,
  }) => {
    const containerRef = useRef<HTMLSpanElement | null>(null);
    const [getHasAnimated, setHasAnimated] = useRefAndState(false);
    const [columnWidths, setColumnWidths] = useState<number[]>([]);
    const prevTextRef = useRef(text);
    const animationFrameId = useRef<number>(0);
    const deleteAfterEndAnimation =
      typeof deleteAfterEndAnimationProps === 'undefined'
        ? triggerOnce
        : deleteAfterEndAnimationProps;

    const iterations = 25;
    const getRandomDigit = useCallback((): string => {
      const digitCharacters = '0123456789';

      const digit =
        digitCharacters[Math.floor(Math.random() * digitCharacters.length)];

      return digit ?? '0';
    }, []);

    const [isMounted, setIsMounted] = useState(false);

    useEffect(() => {
      setIsMounted(true);
    }, []);

    const columnsData = useMemo<ColumnData[]>(() => {
      return text.split('').map((finalChar) => {
        const isDigit = /^\d$/.test(finalChar);

        if (!isDigit) {
          return {
            values: [finalChar],
            shouldAnimate: false,
          };
        }

        // Before mount, render only the final char to avoid hydration mismatch
        // (Math.random() produces different values on server vs client)
        if (!isMounted) {
          return {
            values: [finalChar],
            shouldAnimate: false,
          };
        }

        const randomDigits = Array.from({ length: iterations }, () =>
          getRandomDigit(),
        );

        return {
          values: [finalChar, getRandomDigit(), ...randomDigits, finalChar],
          shouldAnimate: true,
        };
      });
    }, [text, isMounted]);

    useEffect(() => {
      if (prevTextRef.current !== text) {
        prevTextRef.current = text;
        if (animationFrameId.current) {
          cancelAnimationFrame(animationFrameId.current);
          animationFrameId.current = 0;
        }

        if (containerRef.current) {
          containerRef.current
            .querySelectorAll<HTMLElement>(`.${styles.charContainer}`)
            .forEach((c) => c.scrollTo(0, 0));
        }

        setHasAnimated(false);
      }
    }, [text]);

    const easeOutQuart = (x: number): number => 1 - Math.pow(1 - x, 4);

    const measureColumnWidths = useCallback((): void => {
      if (!containerRef.current) return;

      const rootStyles = getComputedStyle(containerRef.current);
      const measurer = document.createElement('span');

      measurer.style.position = 'absolute';
      measurer.style.visibility = 'hidden';
      measurer.style.pointerEvents = 'none';
      measurer.style.whiteSpace = 'pre';
      measurer.style.left = '-9999px';
      measurer.style.top = '0';
      measurer.style.fontFamily = rootStyles.fontFamily;
      measurer.style.fontSize = rootStyles.fontSize;
      measurer.style.fontStyle = rootStyles.fontStyle;
      measurer.style.fontWeight = rootStyles.fontWeight;
      measurer.style.letterSpacing = rootStyles.letterSpacing;
      measurer.style.textTransform = rootStyles.textTransform;

      document.body.appendChild(measurer);

      const nextWidths = text.split('').map((char) => {
        const isDigit = /\d/.test(char);

        measurer.style.letterSpacing = isDigit ? '0' : rootStyles.letterSpacing;
        measurer.textContent = char;

        return Math.max(1, Math.ceil(measurer.getBoundingClientRect().width));
      });

      measurer.remove();

      setColumnWidths((prevWidths) => {
        const isSame =
          prevWidths.length === nextWidths.length &&
          prevWidths.every((width, index) => width === nextWidths[index]);

        return isSame ? prevWidths : nextWidths;
      });
    }, [text]);

    const animateColumn = (
      container: HTMLElement,
      targetElement: HTMLElement,
      delay: number,
    ): void => {
      const duration = 2500;
      const targetY = targetElement.offsetTop;

      let startTime: number | null = null;

      function step(timestamp: number): void {
        if (!startTime) startTime = timestamp;
        const elapsed = timestamp - startTime - delay;

        if (elapsed < 0) {
          animationFrameId.current = requestAnimationFrame(step);

          return;
        }

        const progress = Math.min(elapsed / duration, 1);
        const easedProgress = easeOutQuart(progress);

        container.scrollTo(0, easedProgress * targetY);

        if (progress < 1) {
          animationFrameId.current = requestAnimationFrame(step);
        } else {
          if (!deleteAfterEndAnimation) return;

          const column = container.children.item(0);

          if (!column) return;

          Array.from(column.children).forEach((child) => {
            if (
              child instanceof HTMLElement &&
              child.dataset.target !== 'true'
            ) {
              child.remove();
            }
          });
        }
      }

      animationFrameId.current = requestAnimationFrame(step);
    };

    const runAnimation = useCallback(
      (triggerOnceLocal?: boolean) => {
        const containers =
          containerRef.current?.querySelectorAll<HTMLElement>(
            `.${styles.charContainer}`,
          ) || [];

        containers.forEach((container, index) => {
          if (container.dataset.animate !== 'true') return;

          const target = container.querySelector<HTMLElement>(
            '[data-target="true"]',
          );

          if (target) {
            animateColumn(container, target, index * 100);
          }
        });

        if (
          typeof triggerOnceLocal !== 'undefined'
            ? triggerOnceLocal
            : triggerOnce
        )
          setHasAnimated(true);
      },
      [triggerOnce, columnsData, deleteAfterEndAnimationProps],
    );

    useImperativeHandle(refActionRunAnimation, () => runAnimation);

    useEffect(() => {
      const observer = new IntersectionObserver(
        ([entry]) => {
          if (containerRef.current) {
            if (
              entry?.isIntersecting &&
              (!triggerOnce || !getHasAnimated('ref'))
            ) {
              runAnimation();
            } else if (!entry?.isIntersecting && !triggerOnce) {
              const containers =
                containerRef.current.querySelectorAll<HTMLElement>(
                  `.${styles.charContainer}`,
                );

              containers.forEach((c) => {
                if (c.dataset.animate === 'true') c.scrollTo(0, 0);
              });
            }
          }
        },
        { threshold: 0.6 },
      );

      if (containerRef.current) {
        observer.observe(containerRef.current);
      }

      return () => observer.disconnect();
    }, [triggerOnce, columnsData, deleteAfterEndAnimationProps]);

    useLayoutEffect(() => {
      measureColumnWidths();
    }, [columnsData]);

    useEffect(() => {
      if (typeof window !== 'undefined') {
        const handleResize = () => measureColumnWidths();

        window.addEventListener('resize', handleResize);

        if ('fonts' in document) {
          void (
            document as Document & { fonts?: FontFaceSet }
          ).fonts?.ready.then(handleResize);
        }

        return () => window.removeEventListener('resize', handleResize);
      }

      return;
    }, []);

    return (
      <span
        ref={containerRef}
        className={`notranslate ${styles.slotText}`}
        translate="no"
      >
        {columnsData.map((column, colIndex) => (
          <div
            key={colIndex}
            className={styles.charContainer}
            data-animate={column.shouldAnimate ? 'true' : undefined}
            style={
              columnWidths[colIndex]
                ? { width: `${columnWidths[colIndex]}px` }
                : undefined
            }
          >
            <div className={styles.charColumn}>
              {column.values.map((char, charIndex) => (
                <div
                  key={charIndex}
                  className={styles.char}
                  suppressHydrationWarning
                  data-target={
                    column.shouldAnimate &&
                    charIndex === column.values.length - 1
                      ? 'true'
                      : undefined
                  }
                >
                  {char}
                </div>
              ))}
            </div>
          </div>
        ))}
      </span>
    );
  },
);
