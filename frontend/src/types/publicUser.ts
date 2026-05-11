export enum USER_SOCIAL_LINKS_TYPE {
  X = 'X',
  TELEGRAM = 'Telegram',
  INSTAGRAM = 'Instagram',
  DISCORD = 'Discord',
  YOUTUBE = 'YouTube',
  WEBSITE = 'Website',
}

export interface PublicProfileSocialLink {
  platform: USER_SOCIAL_LINKS_TYPE;
  url: string;
}

export interface PublicProfileDuelEarnings {
  creator_fee: number;
  duel_wins: number;
  total_earnings: number;
}

export interface PublicProfileStatistic {
  duel_earnings: PublicProfileDuelEarnings;
  duels_played: number;
  win_rate: number;
}

export interface ReputationPublicProfile {
  incorrect_resolves_count: 0;
  is_self_resolve_blocked: true;
  score: 0;
  leaderboard_rank?: number;
  x_score?: number;
}

export interface PublicProfileResponse {
  bg_url: string;
  image_url: string;
  social_links: PublicProfileSocialLink[];
  statistic: PublicProfileStatistic;
  username: string;
  reputation: ReputationPublicProfile;
}

export interface CountsDuels {
  created: number;
  lost: number;
  playing: number;
  ended?: number;
  refunded?: number;
  won: number;
}
