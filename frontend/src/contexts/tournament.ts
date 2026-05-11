import { createContext } from 'react';

import { MentionChallenge } from '~types/leaders';
import { Tournament, TournamentRewardList } from '~types/tournament';

const TournamentContext = createContext<{
  tournament: Tournament | null;
  mentionChallenge: MentionChallenge | null;
  tournamentRewards: TournamentRewardList;
}>({
  tournament: null,
  mentionChallenge: null,
  tournamentRewards: [],
});

export default TournamentContext;
