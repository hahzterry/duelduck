'use client';

import { ComponentType, ReactNode } from 'react';
import dynamic from 'next/dynamic';

import { DUEL_TOPIC } from '~types/duel';

const DuckIconSkeleton = () => (
  <div
    style={{
      width: 24,
      height: 24,
      background:
        'linear-gradient(90deg, #2a2a2a 25%, #3a3a3a 50%, #2a2a2a 75%)',
      backgroundSize: '200% 100%',
      animation: 'shimmer 1.5s infinite',
      borderRadius: 4,
    }}
  />
);

const LazyBigDuck = dynamic(
  () =>
    import('~icons/JsxSvg/BigDuck').then((mod) => mod.BigDuck as ComponentType),
  { loading: DuckIconSkeleton, ssr: false },
);

const LazyCryptoDuck = dynamic(
  () =>
    import('~icons/JsxSvg/CryptoDuck').then(
      (mod) => mod.CryptoDuck as ComponentType,
    ),
  { loading: DuckIconSkeleton, ssr: false },
);

const LazyCustomDuck = dynamic(
  () =>
    import('~icons/JsxSvg/CustomDuck').then(
      (mod) => mod.CustomDuck as ComponentType,
    ),
  { loading: DuckIconSkeleton, ssr: false },
);

const LazyGamingDuck = dynamic(
  () =>
    import('~icons/JsxSvg/GamingDuck').then(
      (mod) => mod.GamingDuck as ComponentType,
    ),
  { loading: DuckIconSkeleton, ssr: false },
);

const LazyKingDuck = dynamic(
  () =>
    import('~icons/JsxSvg/KingDuck').then(
      (mod) => mod.KingDuck as ComponentType,
    ),
  { loading: DuckIconSkeleton, ssr: false },
);

const LazySportDuck = dynamic(
  () =>
    import('~icons/JsxSvg/SportDuck').then(
      (mod) => mod.SportDuck as ComponentType,
    ),
  { loading: DuckIconSkeleton, ssr: false },
);

const LazyTrendingDuck = dynamic(
  () =>
    import('~icons/JsxSvg/TrendingDuck').then(
      (mod) => mod.TrendingDuck as ComponentType,
    ),
  { loading: DuckIconSkeleton, ssr: false },
);

export const lazyDucksIconMap: Record<DUEL_TOPIC, ReactNode> = {
  [DUEL_TOPIC.CRYPTO]: <LazyCryptoDuck />,
  [DUEL_TOPIC.GAMING]: <LazyGamingDuck />,
  [DUEL_TOPIC.SPORTS]: <LazySportDuck />,
  [DUEL_TOPIC.CUSTOM]: <LazyCustomDuck />,
  [DUEL_TOPIC.TRENDING]: <LazyTrendingDuck />,
  [DUEL_TOPIC.ALL]: <LazyBigDuck />,
  [DUEL_TOPIC.HISTORY]: <LazyKingDuck />,
};
