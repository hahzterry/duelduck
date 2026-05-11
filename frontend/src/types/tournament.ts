import { MentionChallenge } from '~types/leaders';

export interface Tournament {
  id: string;
  owner_id: string;
  reward_pool_dp: number;
  reward_pool_usdc: number;
  reward_pool_wbt: number;
  reward_pool_sol: number;
  usd_price_equivalent: number;
  name: string;
  description: string;
  author: string;
  usdc_tvl?: number;
  chain_type: number;
  slug?: string;
  players_count: number;
  image_url: string;
  bg_url: string;
  x_url: string;
  bg_card_url?: string;
  instagram_url: string;
  mobile_bg_url?: string;
  youtube_url: string;
  url: string;
  start_date: string;
  finish_date: string;
  created_at: string;
  tournament_type: number;
  mention_challenge_id?: string;
}

export interface TournamentAndMention {
  tournament: Tournament;
  mention_challenge?: MentionChallenge;
}

export type TournamentList = TournamentAndMention[];

export interface TournamentReward {
  tournament_id: string;
  from_rank: number;
  to_rank: number;
  wbt_amount: number;
  sol_amount: number;
  dp_amount: number;
  usdc_amount: number;
}

export type TournamentRewardList = TournamentReward[];

export interface MyTournament {
  id: string;
  pnl: number;
  duels_count: number;
  bg_card_url: string;
  victories: number;
  symbol: string;
  players_count: number;
  name: string;
  image_url: string;
  url: string;
  start_date: string;
  finish_date: string;
  slug?: string;
}

export interface MyTournamentAndMention {
  tournament: MyTournament;
  mention_challenge?: MentionChallenge;
}

export type MyTournaments = MyTournamentAndMention[];
