export interface LeaderStats {
  rank: number;
  user_id: string;
  username: string;
  total_duels: number;
  victories: number;
  spent: number;
  earned: number;
  pnl: number;
  symbol?: string;
  players_invited?: number;
  approved_duels_spent?: number;
}

export interface LeaderStatsWithoutRank {
  username: string;
  total_duels: number;
  victories: number;
  spent: number;
  user_id: string;
  earned: number;
  pnl: number;
}

export interface MentionChallenge {
  id: string;
  title: string;
  description: string;
  season_id?: number;
  reward_pool_contract_address?: string | number;
  reward_pool_usdc: number;
  users_count: number;
  start_date: string;
  finish_date: string;
  is_rewarded?: boolean;
  is_whitelist_opened?: boolean;
  created_at: string;
  claim_reward_date: string;
  metadata?: ChallengeMetadata;
}

export interface MentionChallengeLeader {
  challenge_id: string;
  user_id: string;
  x_id: string;
  username: string;
  is_claimed: boolean;
  x_score: number;
  author_score: number;
  total_score: number;
  is_banned: boolean;
  dd_score: number;
  usdc_reward: number;
  rank: number;
  x_username: string;
  overall_score: number;
  user_share: number;
  image_url?: string;
  claim_tx_hash?: string;
}

export interface MentionChallengeLeaderboardResponse {
  leader_count: number;
  leaders: MentionChallengeLeader[];
}

export const buildMentionSeasons = (challenges: MentionChallenge[]) => {
  if (!challenges || challenges.length === 0) {
    return [];
  }

  const sortedChallenges = [...challenges].sort((a, b) => {
    return new Date(a.start_date).getTime() - new Date(b.start_date).getTime();
  });

  const renamedChallenges = sortedChallenges.map((challenge, index) => ({
    ...challenge,
    title: `Duck Season #${index + 1}`,
    reward_pool_usdc: index === 0 ? 10000 : index === 1 ? 20000 : 30000,
  }));

  const mockSeasons: MentionChallenge[] = [
    {
      id: 'mock-season-3',
      title: 'Duck Season #3',
      description: '',
      reward_pool_usdc: 30000,
      users_count: 0,
      start_date: 'TBD',
      finish_date: 'TBD',
      created_at: '',
      claim_reward_date: '',
    },
    {
      id: 'mock-season-4',
      title: 'Duck Season #4',
      description: '',
      reward_pool_usdc: 50000,
      users_count: 0,
      start_date: 'TBD',
      finish_date: 'TBD',
      created_at: '',
      claim_reward_date: '',
    },
  ];

  return [...renamedChallenges, ...mockSeasons];
};

export interface ChallengeMetadata {
  active?: { from: number; to: number };
  rewards?: { from: number };
  waitlist?: { from: number; to: number };
  shareXCardImageUrl?: string;
  claimRewardModalImageUrl?: string;
  mobile_bg?: string;
  partnerName: string;
  partnerXUsername: string;
  partnerXUrl: string;
}

export function parseChallengeMetadata(
  metadata?: ChallengeMetadata,
): ChallengeMetadata | null {
  if (!metadata) return null;

  return metadata;
}

export interface InviteCodeData {
  challenge_id: string;
  invite_code: string;
  usages: number;
  user_id: string;
  is_whitelisted: boolean;
}

export enum MENTION_TABS {
  LEADERBOARD = 'leaderboard',
  HOW_IT_WORKS = 'how-it-works',
  CALCULATE_REWARDS = 'calculate-rewards',
  REPORT_BUG = 'report-bug',
}
