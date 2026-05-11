'use client';

import {
  MouseEvent,
  MouseEventHandler,
  useEffect,
  useRef,
  useState,
} from 'react';

import colors from '~styles/colors';

import styles from './styles.module.scss';

const isInsideRing = (
  distance: number,
  radius: number,
  ringThickness: number,
) => {
  const innerRadius = radius - ringThickness;
  const outerRadius = radius + ringThickness;

  return distance >= innerRadius && distance <= outerRadius;
};

const getMousePosition = (
  e: MouseEvent<SVGSVGElement, globalThis.MouseEvent>,
) => {
  const rect = e.currentTarget.getBoundingClientRect();

  const center = rect.width / 2;

  const x = e.clientX - rect.left;
  const y = e.clientY - rect.top;
  const dx = x - center;
  const dy = y - center;
  const distance = Math.sqrt(dx * dx + dy * dy);

  return { distance, dx, dy, center };
};

interface ProgressProps {
  progress: number;
  onClickProgress?: (percent: number) => void;
  onClickOutsideProgress?: () => void;
  onClickSvg?: MouseEventHandler<SVGSVGElement>;
  animationOnHover?: boolean;
  hoverStroke?: number;
  baseStroke?: number;
}

export const CircleProgress = ({
  progress,
  onClickSvg,
  onClickProgress,
  onClickOutsideProgress,
  animationOnHover,
  hoverStroke = 6,
  baseStroke = 4,
}: ProgressProps) => {
  const [width, setWidth] = useState(0);
  const [strokewidth, setStrokeWidth] = useState(baseStroke);
  const wrapperRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const el = wrapperRef.current;

    if (!el) return;

    const observer = new ResizeObserver((entries) => {
      for (const entry of entries) {
        setWidth(entry.contentRect.width);
      }
    });

    observer.observe(el);

    return () => observer.disconnect();
  }, []);

  const radius = (width - 6) / 2;
  const center = width / 2;
  const circumference = 2 * Math.PI * radius;
  const dashoffset = circumference - (progress / 100) * circumference;

  const handleMouseMove: MouseEventHandler<SVGSVGElement> = (e) => {
    if (!animationOnHover) return;
    const { distance } = getMousePosition(e);

    const strokeWidth = (e.target as SVGElement).getAttribute('stroke-width');

    const onRing =
      strokeWidth &&
      isInsideRing(distance, center, +(strokeWidth?.match(/^\d+$/) || 0));

    setStrokeWidth(onRing ? hoverStroke : baseStroke);
  };

  const handleMouseEnter = handleMouseMove;

  const handleMouseLeave = () => {
    setStrokeWidth(baseStroke);
  };

  const handleSvgClick: MouseEventHandler<SVGSVGElement> = (e) => {
    const { distance, dx, dy, center } = getMousePosition(e);
    const strokeWidth = (e.target as SVGElement).getAttribute('stroke-width');

    const onRing =
      strokeWidth &&
      isInsideRing(distance, center, +(strokeWidth?.match(/^\d+$/) || 0));

    if (onRing) {
      let angle = Math.atan2(dy, dx);

      angle = angle < -Math.PI / 2 ? 2 * Math.PI + angle : angle;
      const angleFromTop = (angle + Math.PI / 2) % (2 * Math.PI);
      const percent = (angleFromTop / (2 * Math.PI)) * 100;

      onClickProgress?.(percent);
    } else {
      onClickOutsideProgress?.();
    }
  };

  return (
    <div ref={wrapperRef} className={styles.progressCircle__container}>
      <svg
        className={styles.progressCircle}
        width={width}
        height={width}
        viewBox={`0 0 ${width} ${width}`}
        onClick={(e) => {
          onClickSvg?.(e);
          handleSvgClick(e);
        }}
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
        onMouseMove={handleMouseMove}
      >
        <circle
          className="background"
          stroke={`${colors.primaryOpacity}`}
          strokeWidth={strokewidth}
          r={radius}
          cx={center}
          cy={center}
          fill="none"
        />
        <circle
          className="progress"
          stroke={`${colors.primary}`}
          strokeWidth={strokewidth}
          r={radius}
          cx={center}
          cy={center}
          fill="none"
          strokeDasharray={circumference}
          strokeDashoffset={dashoffset}
        />
      </svg>
    </div>
  );
};
