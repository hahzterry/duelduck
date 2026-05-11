import { useEffect, useState } from 'react';
import cx from 'classnames';

import { Typography } from '~components/Typography';
import { ArrowTopRight } from '~icons/JsxSvg/ArrowTopRight';
import { CUSTOM_EVENT_KEYS } from '~types/general';
import { customEvent } from '~utils/customEvent';

import styles from './styles.module.scss';

interface Props {
  scrollHeight?: number;
  variant?: 'default' | 'primary';
  text?: string;
}

export const ScrollToTopButton = ({
  scrollHeight,
  variant = 'default',
  text,
}: Props) => {
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    const handleScroll = () => {
      const screenHeight = window.innerHeight;

      if (document.body.scrollTop > (scrollHeight ?? screenHeight / 4)) {
        setIsVisible(true);
      } else {
        setIsVisible(false);
      }
    };

    document.body.addEventListener('scroll', handleScroll);

    return () => {
      document.body.removeEventListener('scroll', handleScroll);
    };
  }, []);

  const scrollToTop = () => {
    customEvent.emit(CUSTOM_EVENT_KEYS.SCROLL_TO, 0);
  };

  return (
    <button
      onClick={scrollToTop}
      className={cx(styles.buttonContainer, {
        [`${styles[`buttonContainer--visible`]}`]: isVisible,
        [`${styles[`buttonContainer--${variant}`]}`]: variant !== 'default',
      })}
    >
      <div>
        <ArrowTopRight />
      </div>
      {text && (
        <Typography className={styles.buttonContainer__text} text={text} />
      )}
    </button>
  );
};
