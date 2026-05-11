import type { Metadata } from 'next';

import { MainPage } from '~components/PageContents/MainPage';
import { IS_STAGE, LINK_DD } from '~constants/api';
import { normalizeCanonicalUrl } from '~utils/url';

const defaultImage = `/logoShareLink.webp`;
const PROD_LINK_DD = 'https://duelduck.com';

export const generateMetadata = async (): Promise<Metadata> => {
  const isStageHost = IS_STAGE;

  const title = '';
  const description = '';

  if (isStageHost) {
    return {
      title,
      description,
      alternates: {
        canonical: PROD_LINK_DD,
        languages: {
          en: PROD_LINK_DD,
          'x-default': PROD_LINK_DD,
        },
      },
      robots: {
        index: false,
        follow: false,
        noarchive: true,
        nosnippet: true,
      },
      openGraph: {
        title: 'DuelDuck – The On-Chain Prediction Market Platform',
        description:
          'Forecast politics, crypto, elections, and world events with transparent yes/no duels on Solana.',
        url: PROD_LINK_DD,
        type: 'website',
        siteName: 'DuelDuck',
        locale: 'en_US',
        images: [
          {
            url: defaultImage,
            width: 1200,
            height: 630,
            alt: 'DuelDuck Prediction Market Platform',
          },
        ],
      },
      twitter: {
        card: 'summary_large_image',
        title: 'DuelDuck – The On-Chain Prediction Market Platform',
        description:
          'Forecast politics, crypto, elections, and world events with on-chain duels.',
        site: '@duel_duck',
        images: [defaultImage],
      },
      metadataBase: new URL(PROD_LINK_DD),
    };
  }

  return {
    title,
    description,
    alternates: {
      canonical: normalizeCanonicalUrl(LINK_DD),
      languages: {
        en: LINK_DD,
        'x-default': LINK_DD,
      },
    },
    robots: {
      index: true,
      follow: true,
      'max-image-preview': 'large' as const,
      'max-snippet': -1,
      'max-video-preview': -1,
    },
    openGraph: {
      title: 'DuelDuck – The On-Chain Prediction Market Platform',
      description:
        'Forecast politics, crypto, elections, and world events with transparent yes/no duels on Solana.',
      url: LINK_DD,
      type: 'website',
      siteName: 'DuelDuck',
      locale: 'en_US',
      images: [
        {
          url: defaultImage,
          width: 1200,
          height: 630,
          alt: 'DuelDuck Prediction Market Platform',
        },
      ],
    },
    twitter: {
      card: 'summary_large_image',
      title: 'DuelDuck – The On-Chain Prediction Market Platform',
      description:
        'Forecast politics, crypto, elections, and world events with on-chain duels.',
      site: '@duel_duck',
      images: [defaultImage],
    },
    metadataBase: new URL(LINK_DD),
  };
};

export default async function Page() {
  return <MainPage />;
}
