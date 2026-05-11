'use client';

import {
  type ComponentProps,
  memo,
  useEffect,
  useImperativeHandle,
  useRef,
  useState,
} from 'react';
import { StaticImport } from 'next/dist/shared/lib/get-img-props';

type LazyImageProps = Omit<ComponentProps<'img'>, 'src'> & {
  src: string | StaticImport | Blob;
  srcSet?: string;
  offLazyLoad?: boolean;
  priority?: boolean;
  alt: string;
};

export const LazyImage = memo(
  ({ ref, offLazyLoad, priority, ...props }: LazyImageProps) => {
    const imgRef = useRef<HTMLImageElement | null>(null);
    const [visible, setVisible] = useState(false);

    useImperativeHandle(ref, () => imgRef.current!, []);

    useEffect(() => {
      const img = imgRef.current;

      if (!img || visible) return;

      const observer = new IntersectionObserver(([entry]) => {
        if (entry?.isIntersecting) {
          setVisible(true);
          observer.disconnect();
        }
      });

      observer.observe(img);

      return () => observer.disconnect();
    }, [visible, props]);

    const getSrc = (
      source: string | StaticImport | Blob | undefined | null,
    ): string => {
      if (!source) return '/skeletonImage.svg';

      if (typeof source === 'string') return source;

      if (source instanceof Blob) return URL.createObjectURL(source);

      if (typeof source === 'object' && 'src' in source) {
        return source.src;
      }

      return '';
    };

    const finalSrc = getSrc(props.src);

    const isReady = visible || offLazyLoad || priority;

    return (
      <img
        {...props}
        ref={imgRef}
        width={props?.width ?? 16}
        height={props?.height ?? 16}
        loading={offLazyLoad || priority ? 'eager' : 'lazy'}
        fetchPriority={priority ? 'high' : undefined}
        style={{
          ...props.style,
          objectFit: isReady ? props.style?.objectFit : 'cover',
        }}
        alt={props.alt}
        title={props.title || props.alt}
        srcSet={`${isReady ? (props.srcSet ?? finalSrc) : '/skeletonImage.svg'} 1x`}
        src={finalSrc}
      />
    );
  },
);

LazyImage.displayName = 'LazyImage';
