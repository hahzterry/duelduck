'use client';

import {
  CSSProperties,
  memo,
  RefObject,
  useCallback,
  useImperativeHandle,
  useRef,
  useState,
  VideoHTMLAttributes,
} from 'react';

import styles from './styles.module.scss';

/**
 * Generates a minimal SVG data URI placeholder with exact dimensions.
 * This ensures the video container has the same dimensions as the final video,
 * preventing Cumulative Layout Shift (CLS).
 */
const generatePlaceholderSvg = (width: number, height: number): string => {
  const svg = `<svg width="${width}" height="${height}" viewBox="0 0 ${width} ${height}" fill="none" xmlns="http://www.w3.org/2000/svg"><rect width="${width}" height="${height}" fill="#212121"/></svg>`;

  return `data:image/svg+xml,${encodeURIComponent(svg)}`;
};

type LazyVideoProps = VideoHTMLAttributes<HTMLVideoElement> & {
  src: string;
  ref?: RefObject<HTMLVideoElement | null>;
  /**
   * Poster image to show while video loads.
   * Can be a URL string. Uses native video poster attribute.
   */
  poster?: string;
  /**
   * Base64 blur data URL for the poster placeholder.
   * Shows immediately while poster image loads.
   */
  blurDataURL?: string;
  /**
   * Aspect ratio for the video container (e.g., "16/9", "4/3", "1/1").
   * Helps prevent CLS by reserving space before video loads.
   */
  aspectRatio?: string;
  /**
   * Width of the video in pixels.
   * Used to generate placeholder dimensions if aspectRatio is not provided.
   */
  width?: number;
  /**
   * Height of the video in pixels.
   * Used to generate placeholder dimensions if aspectRatio is not provided.
   */
  height?: number;
  /**
   * When true, renders without a wrapper container.
   * Use this when the parent already handles positioning (e.g., absolute positioning).
   * When false/undefined, wraps in a container div for CLS prevention.
   */
  noWrapper?: boolean;
};

export const LazyVideo = memo(
  ({
    src,
    ref,
    poster,
    blurDataURL,
    aspectRatio,
    width,
    height,
    className,
    style,
    autoPlay,
    noWrapper,
    ...props
  }: LazyVideoProps) => {
    const refVideo = useRef<HTMLVideoElement | null>(null);
    const [isLoaded, setIsLoaded] = useState(false);
    const [hasError, setHasError] = useState(false);

    useImperativeHandle(ref, () => refVideo.current!, []);

    // Calculate placeholder dimensions
    const placeholderWidth = width || 640;
    const placeholderHeight = height || 360;

    // Generate fallback placeholder if no poster provided
    const fallbackPlaceholder = generatePlaceholderSvg(
      placeholderWidth,
      placeholderHeight,
    );

    // Determine which poster to use
    const effectivePoster = poster || fallbackPlaceholder;
    const showBlurPlaceholder = blurDataURL && !isLoaded;

    const handleCanPlay = useCallback(() => {
      setIsLoaded(true);
    }, []);

    const handleError = useCallback(() => {
      setHasError(true);
    }, []);

    const handleLoadedData = useCallback(() => {
      setIsLoaded(true);
    }, []);

    // For noWrapper mode, render just the video element
    if (noWrapper) {
      return (
        <video
          ref={refVideo}
          src={src}
          poster={effectivePoster}
          className={className}
          style={style}
          autoPlay={autoPlay}
          onCanPlay={handleCanPlay}
          onLoadedData={handleLoadedData}
          onError={handleError}
          {...props}
        />
      );
    }

    // Build container styles for CLS prevention
    const containerStyle: CSSProperties = {
      ...style,
      aspectRatio: aspectRatio,
    };

    // Only add explicit width/height if provided and no aspect-ratio
    if (!aspectRatio && width) {
      containerStyle.width = width;
    }

    if (!aspectRatio && height) {
      containerStyle.height = height;
    }

    return (
      <div
        className={`${styles.lazyVideoContainer} ${className || ''}`}
        style={containerStyle}
        data-loaded={isLoaded}
        data-error={hasError}
      >
        {/* Blur placeholder layer - shows immediately */}
        {showBlurPlaceholder && (
          <div
            className={styles.blurPlaceholder}
            style={{
              backgroundImage: `url(${blurDataURL})`,
            }}
            aria-hidden="true"
          />
        )}

        {/* Video element with native poster */}
        <video
          ref={refVideo}
          src={src}
          poster={effectivePoster}
          className={styles.video}
          autoPlay={autoPlay}
          onCanPlay={handleCanPlay}
          onLoadedData={handleLoadedData}
          onError={handleError}
          {...props}
        />
      </div>
    );
  },
);

LazyVideo.displayName = 'LazyVideo';
