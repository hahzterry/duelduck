import { Tournament } from '~types/tournament';

/**
 * tournament_type === 2 means the tournament uses mention challenge flow.
 * tournament_type 0 (default) and 1 (invite) use PnL-based leaderboard.
 */
export const isMentionTournament = (tournament: Tournament): boolean =>
  tournament.tournament_type === 2;
