import { memo, useCallback, useEffect, useRef } from 'react';
import { StaticImport } from 'next/dist/shared/lib/get-img-props';

import { LazyImage } from '~components/Lazy/LazyImage';
import { Typography } from '~components/Typography';
import duck from '~icons/animation-duck.svg';
import duck2 from '~icons/animation-duck-2.svg';
import duck3 from '~icons/animation-duck-3.svg';
import duck4 from '~icons/animation-duck-4.svg';

import styles from './styles.module.scss';

interface ConfettiPiece {
  x: number;
  y: number;
  rotation: number;
  size: number;
  velocity: number;
  rotationSpeed: number;
  horizontalDrift: number;
  image: string | StaticImport;
  randomBlur: boolean;
}

const generateRandom = (min: number, max: number) =>
  Math.random() * (max - min) + min;

const getSrc = (img: unknown): string =>
  typeof img === 'string' ? img : ((img as { src?: string })?.src ?? '');

interface ConfettiAnimationProps {
  winnerPrice?: number;
  onEndAnimation?: () => void;
  isInfinite?: boolean;
  position?: 'absolute' | 'relative' | 'fixed' | 'sticky';
  top?: string;
  left?: string;
  width?: string;
  height?: string;
  sizeMin?: number;
  sizeMax?: number;
  velocityMax?: number;
  velocityMin?: number;
  horizontalDriftMax?: number;
  horizontalDriftMin?: number;
  countConfetti?: number;
  maxX?: number;
  minX?: number;
  duckies?: (string | StaticImport)[];
  rotationSpeedMin?: number;
  rotationSpeedMax?: number;
  countBlurEl?: number;
}

export const ConfettiAnimationDown = memo(
  ({
    winnerPrice = 0,
    isInfinite,
    onEndAnimation,
    height = '100vh',
    top = '0px',
    position = 'fixed',
    left = '0px',
    width = `100vw`,
    sizeMin = 20,
    sizeMax = 50,
    velocityMax = 1.4,
    velocityMin = 0.6,
    horizontalDriftMin = -0.5,
    horizontalDriftMax = 0.5,
    countConfetti = 30,
    minX = 0,
    maxX = 100,
    duckies = [duck, duck2, duck3, duck4],
    rotationSpeedMax = 3,
    rotationSpeedMin = 1,
    countBlurEl = 0,
  }: ConfettiAnimationProps) => {
    const piecesRef = useRef<ConfettiPiece[]>([]);
    const elemsRef = useRef<(HTMLDivElement | null)[]>([]);
    const rafRef = useRef<number>(0);

    const generatePiece = useCallback(
      (id: number): ConfettiPiece => ({
        x: generateRandom(minX, maxX),
        y: generateRandom(-100, -50),
        rotation: generateRandom(0, 360),
        size: generateRandom(sizeMin, sizeMax),
        velocity: generateRandom(velocityMin, velocityMax),
        rotationSpeed: generateRandom(rotationSpeedMin, rotationSpeedMax),
        horizontalDrift: generateRandom(horizontalDriftMin, horizontalDriftMax),
        image: getSrc(
          duckies[Math.floor(Math.random() * duckies.length)] || duck,
        ),
        randomBlur: countBlurEl >= id,
      }),
      [
        minX,
        maxX,
        sizeMin,
        sizeMax,
        velocityMin,
        velocityMax,
        rotationSpeedMin,
        rotationSpeedMax,
        horizontalDriftMin,
        horizontalDriftMax,
        duckies,
        countBlurEl,
      ],
    );

    useEffect(() => {
      piecesRef.current = Array.from({ length: countConfetti }, (_, i) =>
        generatePiece(i),
      );

      elemsRef.current = elemsRef.current.slice(0, countConfetti);
    }, [countConfetti, generatePiece]);

    useEffect(() => {
      let allDone = false;

      const tick = () => {
        const pieces = piecesRef.current;

        let belowCount = 0;

        for (let i = 0; i < pieces.length; i++) {
          const p = pieces[i];

          if (!p) continue;

          const newY = p.y + p.velocity;

          if (newY > 100 && isInfinite) {
            p.y = generateRandom(-100, -50);
            p.x = generateRandom(minX, maxX);
          } else {
            p.y = newY;
            p.x = p.x + p.horizontalDrift;

            if (p.x < -10) p.x = 100;
            else if (p.x > 100) p.x = -30;
          }

          p.rotation = (p.rotation + p.rotationSpeed) % 360;

          if (p.y >= 100) belowCount++;

          const el = elemsRef.current[i];

          if (el) {
            el.style.transform = `translate(${p.x}vw, ${p.y}vh) rotate(${p.rotation}deg)`;
          }
        }

        if (belowCount === pieces.length && !isInfinite) {
          allDone = true;
          onEndAnimation?.();

          return;
        }

        if (!allDone) {
          rafRef.current = requestAnimationFrame(tick);
        }
      };

      rafRef.current = requestAnimationFrame(tick);

      return () => {
        cancelAnimationFrame(rafRef.current);
      };
    }, [isInfinite, minX, maxX, onEndAnimation]);

    return (
      <div
        className={styles.confettiContainer}
        style={{ width, height, position, top, left }}
      >
        {!!winnerPrice && <Typography text={`+${winnerPrice}`} />}
        {Array.from({ length: countConfetti }, (_, i) => {
          const p = piecesRef.current[i];
          const size = p?.size ?? sizeMin;

          return (
            <div
              key={i}
              ref={(el) => {
                elemsRef.current[i] = el;
              }}
              className={styles.confettiPiece}
              style={{
                width: `${size}px`,
                height: `${size}px`,
              }}
            >
              {p?.randomBlur && <div className={styles.blur} />}
              <LazyImage
                src={p?.image ?? getSrc(duck)}
                className={styles.duckImage}
                width={size}
                height={size}
                alt="duck"
              />
            </div>
          );
        })}
      </div>
    );
  },
);
