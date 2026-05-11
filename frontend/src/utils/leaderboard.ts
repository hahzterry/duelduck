import { Leaderboard } from '~types/leaderboard';
import { LeaderStats, LeaderStatsWithoutRank } from '~types/leaders';

export function calculateRank(
  index: number,
  sortType: string,
  currentPage: number,
  itemPerPage: number,
  totalCount: number,
): number {
  const minPageValue = itemPerPage * (currentPage - 1);
  const descPos = minPageValue + index + 1;
  const ascPos = totalCount - (minPageValue + index);

  return sortType === 'asc' ? ascPos : descPos;
}

export function updateLeadersRanks(
  leaders: LeaderStatsWithoutRank[],
  sortType: string,
  currentPage: number,
  itemPerPage: number,
  totalCount: number,
): LeaderStats[] {
  return leaders.map((e, i) => {
    return {
      ...e,
      rank: calculateRank(i, sortType, currentPage, itemPerPage, totalCount),
    } as LeaderStats;
  });
}

export const getFinishedLeaderboards = (
  leaderboards: Leaderboard[],
): Leaderboard[] => {
  const now = new Date();

  return leaderboards.filter((lb) => {
    const finishDate = new Date(lb.start_date);

    return finishDate <= now;
  });
};
