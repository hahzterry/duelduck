import { CSSProperties, ReactNode } from 'react';
import { StaticImport } from 'next/dist/shared/lib/get-img-props';

import { LazyImage } from '~/components/Lazy/LazyImage';
import cryptoDuck from '~icons/cryptoDuck.svg';
import customDuck from '~icons/customDuck.svg';
import duck from '~icons/duckCoin.svg';
import gamingDuck from '~icons/gamingDuck.svg';
import { BigDuck } from '~icons/JsxSvg/BigDuck';
import { CryptoDuck } from '~icons/JsxSvg/CryptoDuck';
import { CustomDuck } from '~icons/JsxSvg/CustomDuck';
import { GamingDuck } from '~icons/JsxSvg/GamingDuck';
import { KingDuck } from '~icons/JsxSvg/KingDuck';
import { SportDuck } from '~icons/JsxSvg/SportDuck';
import { TrendingDuck } from '~icons/JsxSvg/TrendingDuck';
import solana from '~icons/solana.svg';
import sportDuck from '~icons/sportDuck.svg';
import usdc from '~icons/usdc.svg';
import whiteBitCoin from '~icons/white-bit-coin.svg';
import { CURRENCY } from '~types/general';

export enum CREATE_DUEL_PAGE_STATE {
  CRYPTO = 'crypto',
  SPORT = 'sport',
  GAMING = 'gaming',
  CUSTOM = 'custom',
}

export interface SubTopicType {
  id: number;
  topic_id: number;
  name: string;
  image_url: string;
}

export enum ENTITY_TYPE {
  PUBLIC_FIGURE = 1,
  EVENT = 2,
  TEAM = 3,
}

export interface EntityType {
  id: number;
  subtopic_id: number;
  name: string;
  entity_type_id: ENTITY_TYPE;
  image_url: string;
}

export type DuelState =
  | 'resolved'
  | 'pending'
  | 'ended'
  | 'in-review'
  | 'launched'
  | 'cancelled'
  | 'user-win';

export enum DUEL_TOPIC {
  ALL = 'all',
  HISTORY = 'history',
  SPORTS = 'sport',
  GAMING = 'gaming',
  CRYPTO = 'crypto',
  CUSTOM = 'custom',
  TRENDING = 'trending',
}

export const ducksIconMap: Record<DUEL_TOPIC, ReactNode> = {
  [DUEL_TOPIC.CRYPTO]: <CryptoDuck />,
  [DUEL_TOPIC.GAMING]: <GamingDuck />,
  [DUEL_TOPIC.SPORTS]: <SportDuck />,
  [DUEL_TOPIC.CUSTOM]: <CustomDuck />,
  [DUEL_TOPIC.TRENDING]: <TrendingDuck />,
  [DUEL_TOPIC.ALL]: <BigDuck />,
  [DUEL_TOPIC.HISTORY]: <KingDuck />,
};

export interface QuestionType {
  type: string;
  cost: number;
  by: Date;
  will?: string;
}

interface DuckListItem {
  text: string;
  value: DUEL_TOPIC;
}

export const leaderBoard: DuckListItem[] = [
  { text: 'All', value: DUEL_TOPIC.ALL },
  {
    text: 'Crypto',
    value: DUEL_TOPIC.CRYPTO,
  },
  {
    text: 'Sports',
    value: DUEL_TOPIC.SPORTS,
  },
  {
    text: 'Gaming',
    value: DUEL_TOPIC.GAMING,
  },
  { text: 'Custom', value: DUEL_TOPIC.CUSTOM },
];

export enum USER_VOTE {
  DID_NOT_VOTE = '',
  NO = 'no',
  YES = 'yes',
}

export enum EMBED_TYPE {
  INITIAL = 'initial',
  CONNECT = 'connect',
  WAIT = 'wait',
  ABOUT = 'about',
}

export enum DUEL_TYPES {
  DEFAULT = 'default',
  VOTE_YES = 'voteYes',
  VOTE_NO = 'voteNo',
  WAITING_FOR_RESULTS = 'waitingForResults',
  WAITING_FOR_RESOLVE = 'waitingForResolve',
  IN_REVIEW = 'inReview',
  WON_YES = 'wonYes',
  WON_NO = 'wonNo',
  LOST_YES = 'lostYes',
  LOST_NO = 'lostNo',
  COMMISSION_EARNED = 'commissionEarned',
  DENIED_PUBLISHING = 'deniedPublishing',
  REFUNDED = 'refunded',
  DUEL_ENDED_AND_USER_NOT_VOTE = 'deniedAndUserNotVoted',
}

export enum USER_VOTE_TYPE {
  NO = 0,
  YES = 1,
}

export const images: Record<CREATE_DUEL_PAGE_STATE, string | StaticImport> = {
  [CREATE_DUEL_PAGE_STATE.CRYPTO]: cryptoDuck,
  [CREATE_DUEL_PAGE_STATE.GAMING]: gamingDuck,
  [CREATE_DUEL_PAGE_STATE.SPORT]: sportDuck,
  [CREATE_DUEL_PAGE_STATE.CUSTOM]: customDuck,
};

export enum STATUS_DUEL {
  CREATED = 0,
  IN_REVIEW,
  AUTO_CANCELLED,
  CANCELLED_BY_ADMIN,
  IN_PROCESS,
  RESOLVED = 5,
  REFUND,
}

export interface DuelInfoType {
  coin_id?: number;
  direction?: number;
  target_price?: number;
  team?: string[];
  public_figure?: string[];
  event?: string[];
  token_image_url?: string;
  token_mint?: string;
  token_is_verified?: boolean;
}

export enum PLAYER_STATUS_TYPE {
  REFUNDED = 2,
}

export type ReportDuelStatusType = 'pending' | 'ignored' | 'approved' | '';

export interface DisputeReportReporterView {
  id: number;
  status: ReportDuelStatusType;
  reporter_comment: string;
  admin_comment?: string;
  created_at: string;
  resolved_at?: string;
}

export interface Duel {
  id: string;
  tournament_id: string;
  owner_id: string;
  duel_type?: string;
  slug?: string;
  user_answer?: USER_VOTE_TYPE | null;
  subtopic?: string;
  entities: number[];
  room_number: number;
  symbol: string;
  refunded_players_count: number;
  resolved_by: string;
  resolved_at: string | null;
  winners_count: number;
  approved_by: string;
  players_count: number;
  username: string;
  status: STATUS_DUEL;
  image_url: string;
  bg_url: string;
  player_status?: PLAYER_STATUS_TYPE | number;
  topic: CREATE_DUEL_PAGE_STATE;
  owner_image_url?: string;
  category?: string;
  sub_category?: string;
  question: string;
  source_of_truth: string;
  deadline: string;
  duel_price: number;
  commission: number;
  duel_info: DuelInfoType | any | null;
  cancellation_reason: string;
  dispute_report: DisputeReportReporterView | null;
  tx_hash?: string | null;
  final_result: USER_VOTE_TYPE | null;
  created_at: string;
  updated_at: string;
  total_count?: number;
  yes_count?: number;
  no_count?: number;
  joined?: boolean;
  author_reputation?: number;
  your_answer: USER_VOTE_TYPE | null;
  is_external_wallet_based: boolean;
  is_owner_resolving: boolean;
  room_token_pda: string;
  is_app_duel: boolean;
  multi_duel_id?: string | null;
  multi_duel_slug?: string | null;
}

export interface DuelsWithTotalResponse {
  duels: Duel[];
  total: number;
}

export interface DuelCounts {
  topic: CREATE_DUEL_PAGE_STATE;
  total_duels?: number;
}

export interface UnsignedWhitechainTransaction {
  chainId: string;
  data: string;
  gas: string;
  maxFeePerGas: string;
  maxPriorityFeePerGas: string;
  nonce: string;
  to: string;
  value: string;
}

// Symbol constants for common tokens
export const TOKEN_SYMBOLS = {
  DUCKIES: 'DDPOINTS',
  USDC: 'USDC',
  WBT: 'WBT',
  SOL: 'SOL',
  BEER2: 'BEER2',
} as const;

type SymbolIconMap = {
  (symbol: string, isExport?: boolean): ReactNode | null;
} & Record<string, ReactNode>;

interface SymbolIconConfig {
  src: string | StaticImport;
  alt: string;
  width?: number;
  height?: number;
  style?: CSSProperties;
}

const ICON_CONFIG: Record<string, SymbolIconConfig> = {
  [TOKEN_SYMBOLS.DUCKIES]: {
    src: duck,
    alt: 'duck',
    width: 40,
    height: 40,
  },

  [TOKEN_SYMBOLS.USDC]: {
    src: usdc,
    alt: 'usdc',
    width: 40,
    height: 40,
  },

  [TOKEN_SYMBOLS.WBT]: {
    src: whiteBitCoin,
    alt: 'wbt',
    width: 40,
    height: 40,
    style: { borderRadius: '50%' },
  },

  [TOKEN_SYMBOLS.SOL]: {
    src: solana,
    alt: 'sol',
    width: 40,
    height: 40,
    style: { borderRadius: '50%' },
  },

  [TOKEN_SYMBOLS.BEER2]: {
    src: 'https://dd.dexscreener.com/ds-data/tokens/solana/GCKTFLxkMeUTUbuQM1k6f1cy4o4tC2V13ye2ebL3p52r.webp',
    alt: 'beer2',
    width: 40,
    height: 40,
    style: { borderRadius: '50%' },
  },
};

// Use Image for SVG files, LazyImage for remote images
const lazyNodeMap: Record<string, ReactNode> = Object.fromEntries(
  Object.entries(ICON_CONFIG).map(([symbol, config]) => {
    return [
      symbol,
      <LazyImage
        key={symbol}
        src={config.src}
        alt={config.alt}
        width={config.width}
        height={config.height}
        style={config.style}
      />,
    ];
  }),
) as Record<string, ReactNode>;

export const currencyIconMap: SymbolIconMap = Object.assign(
  (symbol: string, isExport = false) => {
    const config = ICON_CONFIG[symbol];

    if (!config) return null;

    if (isExport) {
      return (
        <LazyImage
          src={config.src}
          alt={config.alt}
          offLazyLoad={isExport}
          width={config.width}
          height={config.height}
          style={config.style}
        />
      );
    }

    return lazyNodeMap[symbol];
  },

  lazyNodeMap,
);

export const SYMBOL_MINT_MAP: Record<string, string> = {
  [TOKEN_SYMBOLS.DUCKIES]: '',
  [TOKEN_SYMBOLS.USDC]: 'EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v',
  [TOKEN_SYMBOLS.WBT]: '0x925206b8a707096Ed26ae47C84747fE0bb734F59',
  [TOKEN_SYMBOLS.BEER2]: 'GCKTFLxkMeUTUbuQM1k6f1cy4o4tC2V13ye2ebL3p52r',
};

// Backwards compatibility - map CURRENCY enum to symbol strings
export const CURRENCY_TO_SYMBOL: Partial<Record<CURRENCY, string>> = {
  [CURRENCY.DUCKIES]: TOKEN_SYMBOLS.DUCKIES,
  [CURRENCY.USDC]: TOKEN_SYMBOLS.USDC,
  [CURRENCY.WBT]: TOKEN_SYMBOLS.WBT,
  [CURRENCY.BEER_TWO]: TOKEN_SYMBOLS.BEER2,
};

// Reverse mapping - map symbol strings to CURRENCY enum
export const SYMBOL_TO_CURRENCY: Record<string, CURRENCY> = {
  [TOKEN_SYMBOLS.DUCKIES]: CURRENCY.DUCKIES,
  [TOKEN_SYMBOLS.USDC]: CURRENCY.USDC,
  [TOKEN_SYMBOLS.WBT]: CURRENCY.WBT,
  [TOKEN_SYMBOLS.BEER2]: CURRENCY.BEER_TWO,
};

// Legacy export for backwards compatibility
export const PAYMENT_TYPE_MINT_MAP = SYMBOL_MINT_MAP;
