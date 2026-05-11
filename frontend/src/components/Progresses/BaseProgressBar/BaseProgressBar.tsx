import './styles.scss';

import { useEffect, useRef, useState } from 'react';
import cx from 'classnames';
import { pxToRem } from 'src/utils/general';

interface BaseProgressBarProps {
  progress: number;
  className?: string;
  styles?: any;
}

export const BaseProgressBar = ({
  progress,
  className,
  styles,
}: BaseProgressBarProps) => {
  const myRef = useRef<HTMLDivElement | null>(null);
  const [width, setWidth] = useState(0);

  useEffect(() => {
    const updateWidth = () => {
      setWidth(myRef.current?.offsetWidth || 0);
    };

    const resizeObserver = new ResizeObserver(updateWidth);

    if (myRef.current) {
      resizeObserver.observe(myRef.current);
    }

    // Initial width update
    updateWidth();

    return () => {
      if (myRef.current) {
        resizeObserver.unobserve(myRef.current);
      }
    };
  }, []);

  return (
    <div
      className={cx(`baseProgressBar`, className)}
      ref={myRef}
      style={styles ? styles : {}}
    >
      <div
        className="baseProgressBar__progress"
        style={{ width: pxToRem((width / 100) * progress) }}
      >
        {[...Array(Math.floor((width + 18) / 8))].map((_, i) => (
          <div
            className="baseProgressBar__verticalLine"
            key={i}
            style={{ left: 8 * i - 18 }}
          />
        ))}
      </div>
    </div>
  );
};
