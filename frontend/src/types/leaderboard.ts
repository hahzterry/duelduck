export interface Leaderboard {
  claim_reward_date: string;
  created_at: string;
  description: string;
  finish_date: string;
  id: string;
  is_rewarded: boolean;
  is_whitelist_opened: boolean;
  metadata: MetadataLeaderboard;
  pool_address: string;
  reward_pool_usdc: number;
  season_id: number;
  start_date: string;
  title: string;
  users_count: number;
}

export interface LeaderboardUserRank {
  claim_tx_hash?: string;
  image_url: string;
  invite_score: number;
  is_banned: boolean;
  is_claimed: boolean;
  leaderboard_id: string;
  overall_score: number;
  rank: number;
  total_score: number;
  usdc_reward: number;
  user_id: string;
  user_share: number;
  username: string;
  commission_score: number;
}

export interface LeaderboardStats {
  leader_count: number;
  leaders: LeaderboardUserRank[];
}

export interface JoinLeaderboardResponse {
  success: boolean;
}

export interface ClaimLeaderboardResponse {
  tx_hash: string;
}

export interface MetadataLeaderboard {
  tags?: string[];
  main_page_card_url?: string;
  active?: {
    to: number;
    from: number;
  };
  rewards?: {
    from: number;
  };
  waitlist?: {
    to: number;
    from: number;
  };
  background_url?: string;
  image_url?: string;
  mobile_bg?: string;
  partnerName?: string;
  partnerXUrl?: string;
  partnerXUsername?: string;
  shareXCardImageUrl?: string;
  claimRewardModalImageUrl?: string;
}
