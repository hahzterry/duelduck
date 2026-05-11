'use client';

import '~components/Layout/main.scss';

import { useState } from 'react';

import { LazyImage } from '~/components/Lazy/LazyImage';
import { SwitchAnimation } from '~components/Animations/SwitchAnimation';
import { HeaderBorderedButton } from '~components/Buttons/HeaderBorderedButton';
import bigDuck from '~icons/big-angry-duck.svg';
import bigDuckGray from '~icons/big-duck-gray.svg';
import { ArrowRight } from '~icons/JsxSvg/ArrowRight';

import styles from './styles.module.scss';

export default () => {
  const [hovered, setHovered] = useState(false);

  return (
    <section className={styles.notFound}>
      <div className={styles.notFound__content}>
        <h1 className={styles.notFound__404Text}>404</h1>
        <div className={styles.notFound__titleAndButton}>
          <div className={styles.notFound__titles}>
            <span>All the ducks flew to another page</span>
            <span>Hop on me, and I’ll take you there!</span>
          </div>
          <HeaderBorderedButton
            text={'GO TO DUELS'}
            link={'/duels'}
            onMouseEnter={() => setHovered(true)}
            onMouseMove={() => setHovered(true)}
            onMouseLeave={() => setHovered(false)}
            className={styles.notFound__buttonCreate}
            icon={<ArrowRight />}
          />
        </div>
      </div>
      <SwitchAnimation switchKey={hovered ? 'hovered' : 'not-hovered'}>
        <LazyImage
          loading="lazy"
          src={hovered ? bigDuck : bigDuckGray}
          width={300}
          height={300}
          alt={'Duel Duck 404'}
          className={styles.notFound__bigDuck}
        />
      </SwitchAnimation>
    </section>
  );
};
