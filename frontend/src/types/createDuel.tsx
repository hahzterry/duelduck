import { StaticImport } from 'next/dist/shared/lib/get-img-props';

import bigDuck from '~icons/bigduck.svg';
import cryptoDuck from '~icons/cryptoDuck.svg';
import customDuck from '~icons/customDuck.svg';
import gamingDuck from '~icons/gamingDuck.svg';
import sportDuck from '~icons/sportDuck.svg';
import trendingDuck from '~icons/trendingDuck.svg';

export enum CREATE_DUEL_PAGE_STATE {
  CRYPTO = 'crypto',
  SPORT = 'sport',
  GAMING = 'gaming',
  CUSTOM = 'custom',
  ALL = 'all',
  TRENDING = 'trending',
}

export const images: Record<CREATE_DUEL_PAGE_STATE, string | StaticImport> = {
  [CREATE_DUEL_PAGE_STATE.CRYPTO]: cryptoDuck,
  [CREATE_DUEL_PAGE_STATE.GAMING]: gamingDuck,
  [CREATE_DUEL_PAGE_STATE.SPORT]: sportDuck,
  [CREATE_DUEL_PAGE_STATE.CUSTOM]: customDuck,
  [CREATE_DUEL_PAGE_STATE.TRENDING]: trendingDuck,
  [CREATE_DUEL_PAGE_STATE.ALL]: bigDuck,
};

export enum DUEL_FIELDS {
  PRICE_DUEL = 'priceDuel',
  COMMISSION = 'commission',
  QUESTION = 'question',
  DATE = 'deadline',
  LINK = 'link',
  CURRENCY = 'currency',
  IMAGE = 'image',
  BACKGROUND = 'background',
  IS_USER_RESOlVER = 'resolver',
}

export enum PRICE_CHANGE_DIRECTION_TYPE {
  DOWN,
  UP,
}

export interface CustomDuelType {
  [DUEL_FIELDS.CURRENCY]: { symbol: string; mint?: string };
  [DUEL_FIELDS.LINK]?: string;
  [DUEL_FIELDS.QUESTION]?: string;
  [DUEL_FIELDS.DATE]?: number;
  [DUEL_FIELDS.PRICE_DUEL]?: string;
  [DUEL_FIELDS.COMMISSION]?: number;
  [DUEL_FIELDS.BACKGROUND]?: File;
  [DUEL_FIELDS.IMAGE]?: File;
  [DUEL_FIELDS.IS_USER_RESOlVER]: boolean;
}
