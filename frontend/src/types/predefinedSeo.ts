import type { ReactNode } from 'react';

export interface SeoSectionHeaderData {
  title: string;
  highlight: string;
  description: string;
}

export interface SeoStepCardData {
  step: string;
  title: string;
  description: string;
}

export interface SeoHubCardData {
  icon: ReactNode;
  title: string;
  description: string;
  tags: string[];
}

export type IconTheme = 'primary' | 'success' | 'warning';

export interface SeoFeatureCardData {
  icon: ReactNode;
  title: string;
  description: string;
  linkText?: string;
  linkHref?: string;
  iconTheme?: IconTheme;
}

export interface SeoInterstitialCard {
  title: string;
  description: string;
  isActive?: boolean;
  linkText?: string;
  linkHref?: string;
  icon?: ReactNode;
}

export interface SeoInterstitialData {
  title: string;
  highlight: string;
  cards: SeoInterstitialCard[];
  buttonText?: string;
  buttonHref?: string;
  tip?: string;
}

export interface SeoAboutData {
  title: string;
  highlight: string;
  description: string;
}

export interface SeoStatItem {
  value: string;
  label: string;
}

export interface SeoStatsHeaderData {
  title: string;
  highlight: string;
  description: string;
  stats: SeoStatItem[];
  takeaways: string[];
  avgAccuracy?: string;
}

export interface SeoGroupedHubCard {
  icon: string;
  title: string;
  description: string;
}

export interface SeoGroupedHubGroup {
  region: string;
  cards: SeoGroupedHubCard[];
}

export interface SeoGroupedHubsData {
  header: SeoSectionHeaderData;
  groups: SeoGroupedHubGroup[];
}

export interface PredefinedSeoContent {
  statsHeader?: SeoStatsHeaderData | null;
  howItWorks: {
    header: SeoSectionHeaderData;
    steps: SeoStepCardData[];
  };
  hubs: {
    header: SeoSectionHeaderData;
    cards: SeoHubCardData[];
  };
  groupedHubs?: SeoGroupedHubsData | null;
  interstitial?: SeoInterstitialData | null;
  whyForecast: {
    header: SeoSectionHeaderData;
    cards: SeoFeatureCardData[];
  };
  connections: {
    header: SeoSectionHeaderData;
    cards: SeoFeatureCardData[];
  };
  about?: SeoAboutData | null;
  embed?: {
    header: SeoSectionHeaderData;
    buttonText: string;
    buttonHref: string;
  } | null;
  embedAfterFaq?: boolean;
  cta?: {
    header: SeoSectionHeaderData;
    primaryButton: { text: string; href: string };
    secondaryButton: { text: string; href: string };
  } | null;
}
