import { useEffect, useState } from 'react';

import { LazyImage } from '~/components/Lazy/LazyImage';

import styles from './styles.module.scss';

interface Props {
  url: string;
}

export const Avatar = ({ url }: Props) => {
  const [imgSrc, setImgSrc] = useState(url);

  useEffect(() => {
    setImgSrc(url);
  }, [url]);

  return (
    <div
      className={styles.avatar}
      style={{
        borderRadius: imgSrc === '/duckProfile.svg' ? 'unset' : '50%',
      }}
    >
      <LazyImage
        src={imgSrc}
        width={324}
        height={424}
        loading="lazy"
        decoding="async"
        fetchPriority="low"
        style={{
          borderRadius: imgSrc === '/duckProfile.svg' ? 'unset' : '50%',
        }}
        alt={'Avatar'}
        onError={() => setImgSrc('/duckProfile.svg')}
      />
    </div>
  );
};
