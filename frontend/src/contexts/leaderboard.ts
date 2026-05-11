import { createContext, Dispatch, SetStateAction } from 'react';

import { LeaderboardUserRank } from '~types/leaderboard';

interface LeaderboardContextValue {
  userRank: LeaderboardUserRank | null;
  setUserRank: Dispatch<SetStateAction<LeaderboardUserRank | null>>;
}

const LeaderboardContext = createContext<LeaderboardContextValue>({
  userRank: null,
  setUserRank: () => undefined,
});

export default LeaderboardContext;
