export interface DailyReward {
  day_streak: number;
  claimed: boolean;
}

export interface JwtInfo {
  refresh_token: string;
  refresh_exp_time: string;
  access_token: string;
  access_exp_time: string;
}

export interface User {
  id: string;
  telegram_id: string;
  username: string;
  email: string;
  image_url: string;
  bg_url: string;
  bio: string;
  is_verified: boolean;
  has_unread_self_resolve_restriction?: boolean;
  website: string;
  role: number;
  adult_confirmed: boolean;
  is_premium: boolean;
  balance: number;
  referral_token: string;
  daily_reward_streak: number;
  last_claimed_reward: string;
  created_at: string;
  updated_at: string;
  x_url: string;
  x_id: string;
  x_username?: string;
  youtube_url: string;
  telegram_url: string;
  instagram_url: string;
  reputation?: number;
  discord_url: string;
  is_self_resolve_blocked?: boolean;
  public_address: string;
  last_completed_streak: string;
  level: number;
  current_xp: number;
  active_wallet_id: string;
}

export interface AuthData {
  daily_reward: DailyReward;
  jwt_info: JwtInfo;
  user: User;
}

export interface Session {
  id: string;
  user: User;
  accessToken: any;
  refreshToken: any;
  expiresAt: number;
}

export type SessionType = Session | null;
