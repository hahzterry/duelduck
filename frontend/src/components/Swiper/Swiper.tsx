import {
  CSSProperties,
  JSX,
  MouseEvent,
  ReactNode,
  TouchEvent,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
import cx from 'classnames';

import { Chevron } from '~icons/Chevrons/Chevron';
import colors from '~styles/colors';
import { clamp } from '~utils/numbers';

import styles from './styles.module.scss';

const getActive = (
  countActiveElements: number,
  initialActiveProps?: number | RangeActive,
) => {
  const initialActive =
    typeof initialActiveProps === 'undefined'
      ? {
          start: 0,
          end: countActiveElements - 1,
        }
      : initialActiveProps;

  return typeof initialActive === 'number'
    ? {
        end: initialActive,
        start: initialActive + (countActiveElements - 1),
      }
    : initialActive;
};

export enum SWIPER_MODE {
  FREE = 'free',
}

type RangeActive = {
  start: number;
  end: number;
};

interface SwiperProps {
  children: ReactNode;
  className?: string;
  styles?: CSSProperties;
  onIndexChange?: (index: number) => void;

  isMaxContentHeight?: boolean;

  showDots?: boolean;
  showArrows?: boolean;
  countActiveElements?: number;

  overflowVisible?: boolean;
  isShowShadow?: boolean;
  shadowWidthMultiplier?: number;

  initialActive?: number | RangeActive;

  prevArrow?: ReactNode;
  nextArrow?: ReactNode;

  arrowsPosition?: 'inside' | 'outside' | 'bottom';

  colorDot?: {
    backgroundActive?: string;
    backgroundDefault?: string;
  };

  classNamesDots?: string;
  gap?: number;

  autoplay?: boolean;
  autoplayInterval?: number;
  loop?: boolean;

  mode?: SWIPER_MODE;

  classNameElement?: string | ((index: number) => string);

  element?: keyof JSX.IntrinsicElements;
}

export const Swiper = ({
  children,
  className,
  onIndexChange,
  showDots = true,
  showArrows = true,
  element = 'div',
  prevArrow,
  nextArrow,
  autoplay = false,
  autoplayInterval = 5000,
  arrowsPosition,
  colorDot,
  loop = true,
  classNamesDots,
  gap = 0,
  mode,
  classNameElement,
  overflowVisible,
  isShowShadow = false,
  initialActive: initialActiveProps,
  countActiveElements = 1,
  isMaxContentHeight,
  styles: customStyles,
  shadowWidthMultiplier = 1,
}: SwiperProps) => {
  const Element: keyof JSX.IntrinsicElements = element;
  const slides = useMemo(
    () =>
      (Array.isArray(children) ? children : [children]).filter(
        Boolean,
      ) as ReactNode[],
    [children],
  );

  const arrowPositionClass =
    styles[`swiper__arrow--${arrowsPosition ?? 'inside'}`];

  const countActiveElementsRef = useRef(countActiveElements);

  const [activeIndex, setActiveIndex] = useState<RangeActive>(
    getActive(countActiveElements, initialActiveProps),
  );

  useEffect(() => {
    setActiveIndex(
      getActive(countActiveElementsRef.current, initialActiveProps),
    );
  }, [initialActiveProps]);

  useEffect(() => {
    countActiveElementsRef.current = countActiveElements;
  }, [countActiveElements]);

  const [isDragging, setIsDragging] = useState(false);
  const containerRef = useRef<any | null>(null);
  const [width, setWidth] = useState<number | undefined>(undefined);
  const [widthTrack, setWidthTrack] = useState<number | undefined>(undefined);

  const startXRef = useRef<number | null>(null);
  const deltaXRef = useRef(0);

  const minWidthElement = useRef(0);
  const observerElements = useRef<ResizeObserver | null>(null);
  const itemsRef = useRef(new Map<number, HTMLDivElement>());
  const trackRef = useRef<HTMLDivElement | null>(null);
  const autoplayRef = useRef<number | null>(null);
  const freeOffsetRef = useRef(0);
  const freeCurrentOffsetRef = useRef(0);
  const observerTrackWidth = useRef<ResizeObserver | null>(null);
  const slideCount = slides.length;

  useEffect(() => {
    const track = trackRef.current;

    if (track) {
      observerTrackWidth.current = new ResizeObserver((entries) => {
        const el = entries[0];

        if (el) {
          setWidthTrack(el.target.clientWidth);
        }
      });

      observerTrackWidth.current.observe(track);
    }

    return () => {
      observerTrackWidth.current?.disconnect();
    };
  }, []);

  useEffect(() => {
    if (!slideCount) return;

    onIndexChange?.(clamp(activeIndex.start, 0, slideCount - 1));
  }, [activeIndex, onIndexChange, slideCount]);

  const onObserve: ResizeObserverCallback = (els) => {
    els.forEach((el) => {
      const newWidth = el.target.getBoundingClientRect().width;
      const oldWidth = minWidthElement.current;

      if (!oldWidth || newWidth < oldWidth) {
        minWidthElement.current = newWidth;
      }
    });
  };

  useEffect(() => {
    return () => {
      if (observerElements.current) {
        observerElements.current.disconnect();
        observerElements.current = null;
      }
    };
  }, []);

  const isLoop = loop && slideCount > 1;
  const activePositionIndex = isLoop
    ? activeIndex.start + countActiveElements
    : activeIndex.start;

  const goTo = (index: number) => {
    if (!slideCount) return;

    if (isLoop) {
      const start =
        (index + (slideCount - (countActiveElements - 1))) %
        (slideCount - (countActiveElements - 1));

      const end = start + countActiveElements - 1;

      setActiveIndex({
        start,
        end,
      });

      onIndexChange?.(clamp(start, 0, slideCount - 1));

      return;
    }

    const start = Math.max(
      0,
      Math.min(index, slideCount - countActiveElements),
    );

    setActiveIndex({
      start,
      end: start + countActiveElements - 1,
    });

    onIndexChange?.(clamp(start, 0, slideCount - 1));
  };

  const goNext = () => goTo(activeIndex.start + 1);
  const goPrev = () => goTo(activeIndex.start - 1);

  const isPrevDisabled = !isLoop && activeIndex.start === 0;
  const isNextDisabled = !isLoop && activeIndex.end === slideCount - 1;

  useEffect(() => {
    if (!containerRef.current) {
      setWidth(undefined);

      return;
    }

    const observer = new ResizeObserver((e) => {
      const el = e[0];

      if (el) {
        const width = el.target.clientWidth;

        setWidth(
          (width -
            gap * (countActiveElements === 1 ? 0 : countActiveElements)) /
            countActiveElements,
        );
      }
    });

    observer.observe(containerRef.current);

    return () => {
      observer.disconnect();
    };
  }, [activeIndex, countActiveElements, children, gap]);

  const getCanScroll = useCallback(() => {
    return (widthTrack || 0) < (gap + (width || 0)) * slideCount;
  }, [widthTrack, slideCount, gap, width]);

  useEffect(() => {
    if (!autoplay || !getCanScroll()) return;

    if (autoplayRef.current) {
      window.clearInterval(autoplayRef.current);
    }

    autoplayRef.current = window.setInterval(() => {
      goNext();
    }, autoplayInterval);

    return () => {
      if (autoplayRef.current) {
        window.clearInterval(autoplayRef.current);
      }
    };
  }, [autoplay, autoplayInterval, slideCount, activeIndex, isLoop]);

  const stopAutoplay = () => {
    if (autoplayRef.current) {
      window.clearInterval(autoplayRef.current);
      autoplayRef.current = null;
    }
  };

  const startDrag = (clientX: number, mode?: SWIPER_MODE) => {
    if (!getCanScroll()) return;
    startXRef.current = clientX;
    if (mode !== SWIPER_MODE.FREE) deltaXRef.current = 0;
    stopAutoplay();
  };

  const freeMove = (clientX: number) => {
    if (startXRef.current === null) return;
    const dx = clientX - startXRef.current;

    deltaXRef.current = dx;

    if (!trackRef.current) return;
    const width = trackRef.current.offsetWidth || 1;
    const percentage = (dx / width) * 100;

    if (width >= trackRef.current.scrollWidth) return;

    const minPercent = ((trackRef.current.scrollWidth - width) / width) * 100;

    const newValue = clamp(freeOffsetRef.current + percentage, -minPercent, 0);

    trackRef.current.style.transition = 'none';
    trackRef.current.style.transform = `translateX(calc(${newValue}%))`;
    freeCurrentOffsetRef.current = percentage;
  };

  const moveDrag = (clientX: number, mode?: SWIPER_MODE) => {
    if (startXRef.current === null) return;

    setIsDragging(true);
    if (mode === SWIPER_MODE.FREE) {
      freeMove(clientX);

      return;
    }

    const dx = clientX - startXRef.current;

    deltaXRef.current = dx;

    if (!trackRef.current) return;

    const widthTrack = trackRef.current.offsetWidth || 1;
    const percentage = (dx / widthTrack) * 100;

    trackRef.current.style.transition = 'none';
    trackRef.current.style.transform = `translateX(calc(${
      -activePositionIndex * (width || 0)
    }px + ${percentage}% - ${gap * activePositionIndex}px))`;
  };

  const endDrag = (mode?: SWIPER_MODE) => {
    setIsDragging(false);

    if (!trackRef.current) return;

    if (mode === SWIPER_MODE.FREE) {
      const width = trackRef.current.offsetWidth || 1;

      freeOffsetRef.current = clamp(
        freeCurrentOffsetRef.current + freeOffsetRef.current,
        -(((trackRef.current.scrollWidth - width) / width) * 100),
        0,
      );

      freeCurrentOffsetRef.current = 0;

      return;
    }

    trackRef.current.style.transition = '';

    const widthTrack = minWidthElement.current || 1;
    const threshold = widthTrack * 0.15;

    if (deltaXRef.current > threshold) {
      if (!isPrevDisabled) {
        goPrev();
      } else {
        trackRef.current.style.transform = `translateX(calc(-${
          activePositionIndex * (width || 0)
        }px - ${gap * activePositionIndex}px))`;
      }
    } else if (deltaXRef.current < -threshold) {
      if (!isNextDisabled) {
        goNext();
      } else {
        trackRef.current.style.transform = `translateX(calc(-${
          activePositionIndex * (width || 0)
        }px - ${gap * activePositionIndex}px))`;
      }
    } else {
      trackRef.current.style.transform = `translateX(calc(-${
        activePositionIndex * (width || 0)
      }px - ${gap * activePositionIndex}px))`;
    }

    startXRef.current = null;
    deltaXRef.current = 0;
  };

  const handleTouchStart = (e: TouchEvent<HTMLDivElement>) => {
    const touch = e.touches[0];

    startDrag(touch?.clientX || 0, mode);

    const handleMouseMove = (ev: TouchEvent<Document>) => {
      const touch = ev.touches[0];

      moveDrag(touch?.clientX || 0, mode);
    };

    const handleMouseUp = () => {
      endDrag(mode);
      window.removeEventListener('touchmove', handleMouseMove as any);
      window.removeEventListener('touchend', handleMouseUp as any);
    };

    window.addEventListener('touchmove', handleMouseMove as any);
    window.addEventListener('touchend', handleMouseUp as any);
  };

  const handleMouseDown = (e: MouseEvent<HTMLDivElement>) => {
    e.preventDefault();
    startDrag(e.clientX, mode);

    const handleMouseMove = (ev: MouseEvent<Document>) => {
      moveDrag(ev.clientX, mode);
    };

    const handleMouseUp = () => {
      endDrag(mode);
      window.removeEventListener('mousemove', handleMouseMove as any);
      window.removeEventListener('mouseup', handleMouseUp as any);
    };

    window.addEventListener('mousemove', handleMouseMove as any);
    window.addEventListener('mouseup', handleMouseUp as any);
  };

  const containerClassName = cx(
    styles.swiper,
    !getCanScroll() ? styles['swiper--single'] : '',
    className || '',
  );

  const viewportClassName = cx(styles.swiper__viewport, {
    [`${styles['swiper__viewport--visibleOverflow']}`]: overflowVisible,
    [`${styles['swiper__viewport--dragging']}`]: isDragging,
  });

  const showClones = isLoop;
  const firsts = showClones ? slides.slice(0, countActiveElements) : null;
  const lasts = showClones ? slides.slice(-countActiveElements) : null;
  const shadowWidth = minWidthElement.current || width || 0;

  const slideClassName = (index: number) =>
    cx(
      styles.swiper__slide,
      mode === 'free' && styles['swiper__slide--free'],
      typeof classNameElement === 'string'
        ? classNameElement
        : classNameElement?.(index),
    );

  return (
    <Element
      // eslint-disable-next-line @typescript-eslint/ban-ts-comment
      // @ts-ignore
      className={containerClassName}
      ref={containerRef}
      style={{
        ...customStyles,
        height: isMaxContentHeight ? 'max-content' : undefined,
      }}
    >
      {showArrows && slideCount > 1 && (
        <button
          type="button"
          className={cx(
            styles.swiper__arrow,
            styles['swiper__arrow--prev'],
            arrowPositionClass,
            {
              [`${styles['swiper__arrow--disabled']}`]: isPrevDisabled,
            },
          )}
          onClick={isPrevDisabled ? undefined : goPrev}
          disabled={isPrevDisabled}
          aria-label="Previous slide"
        >
          {prevArrow ?? <Chevron />}
        </button>
      )}

      <div
        className={viewportClassName}
        onTouchStart={handleTouchStart}
        onMouseDown={handleMouseDown}
      >
        <div
          ref={trackRef}
          className={styles.swiper__track}
          style={{
            gap,
            transform: `translateX(calc(-${
              activePositionIndex * (width || 0)
            }px - ${gap * activePositionIndex}px))`,
          }}
        >
          {showClones &&
            lasts &&
            lasts.map((e, i) => (
              <div
                key={`lasts-${i}`}
                data-index={i}
                style={{
                  pointerEvents: isDragging ? 'none' : undefined,
                  flexBasis:
                    mode === SWIPER_MODE.FREE
                      ? 'max-content'
                      : (width ?? '100%'),
                }}
                className={slideClassName(-2)}
              >
                {e}
              </div>
            ))}

          {slides.map((slide, index) => (
            <div
              key={index}
              style={{
                pointerEvents: isDragging ? 'none' : undefined,
                flexBasis: mode === SWIPER_MODE.FREE ? 'max-content' : width,
              }}
              ref={(node) => {
                if (node) {
                  itemsRef.current.set(index, node);
                  if (!observerElements.current)
                    observerElements.current = new ResizeObserver(onObserve);
                  observerElements.current?.observe(node);

                  const newWidth = node.getBoundingClientRect().width;
                  const oldWidth = minWidthElement.current;

                  if (!oldWidth || newWidth < oldWidth) {
                    minWidthElement.current = newWidth;
                  }
                } else {
                  const item = itemsRef.current.get(index);

                  if (item) observerElements.current?.unobserve(item);
                  itemsRef.current.delete(index);
                }
              }}
              className={slideClassName(index)}
            >
              {slide}
            </div>
          ))}

          {showClones &&
            firsts &&
            firsts.map((e, i) => (
              <div
                key={`firsts-${i}`}
                data-index={i}
                style={{
                  pointerEvents: isDragging ? 'none' : undefined,
                  flexBasis: mode === SWIPER_MODE.FREE ? 'max-content' : width,
                }}
                className={slideClassName(-1)}
              >
                {e}
              </div>
            ))}
        </div>
      </div>

      {isShowShadow && shadowWidth > 0 && (
        <>
          <div
            className={cx(
              styles.swiper__shadow,
              styles['swiper__shadow--left'],
            )}
            style={{ width: shadowWidth * shadowWidthMultiplier }}
          />
          <div
            className={cx(
              styles.swiper__shadow,
              styles['swiper__shadow--right'],
            )}
            style={{ width: shadowWidth * shadowWidthMultiplier }}
          />
        </>
      )}

      {showArrows && slideCount > 1 && (
        <button
          type="button"
          className={cx(
            styles.swiper__arrow,
            styles['swiper__arrow--next'],
            arrowPositionClass,
            {
              [`${styles['swiper__arrow--disabled']}`]: isNextDisabled,
            },
          )}
          onClick={isNextDisabled ? undefined : goNext}
          disabled={isNextDisabled}
          aria-label="Next slide"
        >
          {nextArrow ?? <Chevron />}
        </button>
      )}

      {showDots && slideCount > 1 && (
        <div className={cx(styles.swiper__dots, classNamesDots)}>
          {slides.map((_, index) => (
            <button
              key={index}
              type="button"
              style={{
                background:
                  index <= activeIndex.end && index >= activeIndex.start
                    ? colorDot?.backgroundActive || colors.primary
                    : colorDot?.backgroundDefault || 'rgba(255, 197, 77, 0.20)',
              }}
              className={styles.swiper__dot}
              onClick={() => goTo(index)}
              aria-label={`Go to slide ${index + 1}`}
            />
          ))}
        </div>
      )}
    </Element>
  );
};
