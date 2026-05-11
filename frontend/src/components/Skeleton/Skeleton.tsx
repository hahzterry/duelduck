import 'react-loading-skeleton/dist/skeleton.css';

import SkeletonConponent from 'react-loading-skeleton';

import colors from '~styles/colors';

interface SkeletonProps {
  cirle?: boolean;
  className?: string;
  height?: number | string;
  width?: number | string;
}

export const Skeleton = ({
  cirle,
  className,
  height,
  width,
}: SkeletonProps) => {
  return (
    <SkeletonConponent
      baseColor={colors['background']}
      highlightColor={colors['backgroundDark']}
      circle={cirle}
      className={className}
      containerClassName={className}
      height={height}
      width={width}
    />
  );
};
