import { type Duel, USER_VOTE_TYPE } from '~types/duel';

export interface MultiDuelOutcome {
  id: string;
  duelId: string;
  title: string;
  tournamentId: string | null;
  status: Duel['status'];
  finalResult: USER_VOTE_TYPE | null;
  chance: number | null;
  yesCount: number | null;
  noCount: number | null;
  duelPrice: number | null;
  currency: string | null;
  imageUrl: string | null;
  tokenImageUrl: string | null;
  usdPrice: number;
  tokenMint: string | null;
  userVote: USER_VOTE_TYPE | null;
  isWin: boolean | null;
  resultAmount: number | null;
  commission: number | null;
  deadline: string | null;
}

export interface MultiDuelDetails {
  id: string;
  slug: string | null;
  question: string;
  imageUrl: string | null;
  bgUrl: string | null;
  isRewarded?: boolean;
  createdAt: string | null;
  deadline: string | null;
  createdBy: string | null;
  createdByAvatarUrl: string | null;
  updatedAt: string | null;
  updatedBy: string | null;
  playersCount: number | null;
  joined: boolean;
  shareboardJoined: boolean;
  rewardPoolAmount: number | null;
  rewardPoolCurrency: string | null;
  totalSpent: number | null;
  tvl: number | null;
  isCompleted: boolean;
  isClaimed?: boolean;
  rewardAmount?: number;
  referralCode?: string;
  outcomes: MultiDuelOutcome[];
}

export interface MultiDuelHistory {
  duel_id: string;
  image_url: string;
  is_current_user: boolean;
  keyword: string;
  token_image_url: string;
  multi_duel_id: string;
  outcome_id: string;
  payout_status: 'paid' | 'unpaid';
  reward_amount: number;
  status: 'won' | 'lost' | 'pending' | 'refunded';
  tx_hash: string;
  user_id: string;
  username: string;
  vote: USER_VOTE_TYPE;
  voted_at: string;
}

export interface ShareboardMultiDuelTableData {
  created_at: string;
  id: string;
  image_url: string;
  invited_count: number;
  is_current_user: boolean;
  multi_duel_id: string;
  rank: number;
  referral_code: string;
  reward_amount: number;
  reward_tx_hash: string;
  share_percent: number;
  status: 'claimed' | 'unclaimed';
  total_score: number;
  total_volume_usdc: number;
  updated_at: string;
  user_id: string;
  username: string;
}

export interface MultiDuelShareboardRes {
  me?: ShareboardMultiDuelTableData;
  participants: ShareboardMultiDuelTableData[];
  total: number;
}
