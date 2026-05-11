'use client';

import { useCallback, useEffect, useRef, useState } from 'react';

import { FullScreenLoader } from '~components/FullScreenLoader';
import { useLoaderWithBotCheck } from '~hooks/useLoaderWithBotCheck';
import { useAppStore } from '~store/appStore';

interface LoaderSuspenseProps {
  homePage?: boolean;
}

const MAX_LOADER_TIME = 800;
const FADE_DURATION = 300;

const LoaderSuspense = ({ homePage }: LoaderSuspenseProps) => {
  const [progress, setProgress] = useState(0);
  const [isHidden, setIsHidden] = useState(false);
  const { setIsLoaded } = useAppStore();

  const completed = useRef(false);

  const onLoaded = useCallback(() => {
    if (completed.current) return;
    completed.current = true;

    setProgress(100);

    // Start fade-out, then hide completely
    setTimeout(() => {
      setIsHidden(true);
      setIsLoaded(true);
      document.body.classList.remove('noBodyScroll');
    }, FADE_DURATION);
  }, []);

  useLoaderWithBotCheck((value) => {
    if (!homePage || !value) return;
    onLoaded();
  });

  // Fast progress animation (800ms)
  useEffect(() => {
    if (typeof window === 'undefined') return;

    const duration = MAX_LOADER_TIME;
    const start = performance.now();

    let frameId: number;

    const tick = () => {
      if (completed.current) return;

      const elapsed = performance.now() - start;
      const t = Math.min(elapsed / duration, 1);
      // Ease-out cubic for smooth progress
      const eased = 1 - Math.pow(1 - t, 3);
      const value = eased * 100;

      setProgress(value);

      if (t < 1 && !completed.current) {
        frameId = requestAnimationFrame(tick);
      } else if (t >= 1) {
        // Max time reached, complete immediately
        onLoaded();
      }
    };

    frameId = requestAnimationFrame(tick);

    return () => cancelAnimationFrame(frameId);
  }, []);

  // Listen for FCP - complete early if FCP happens
  useEffect(() => {
    if (typeof window === 'undefined') return;
    if (typeof PerformanceObserver === 'undefined') {
      // Fallback: complete after max time
      const timeout = setTimeout(onLoaded, MAX_LOADER_TIME);

      return () => clearTimeout(timeout);
    }

    let paintObserver: PerformanceObserver | null = null;

    try {
      paintObserver = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if (entry.entryType !== 'paint') continue;

          const name = (entry as any).name;

          if (name === 'first-contentful-paint') {
            // FCP happened, we can complete now
            if (!completed.current) {
              onLoaded();
            }

            paintObserver?.disconnect();
            break;
          }
        }
      });

      paintObserver.observe({ type: 'paint', buffered: true as any });
    } catch {
      paintObserver?.disconnect();
      paintObserver = null;
    }

    return () => {
      paintObserver?.disconnect();
    };
  }, []);

  if (isHidden) return null;

  return <FullScreenLoader progress={progress} />;
};

export default LoaderSuspense;
