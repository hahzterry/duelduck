export interface Leader {
  date: string;
  name: string;
  tx_hash: string;
  victories: number;
  win_sum: number;
}

export type Leaders = Leader[];

export interface TableLeadersRow {
  id: number;
  name: string;
  duels: number;
  victories: number;
  spent: number;
  earned: number;
  pnl: number;
  isCurrentUser: boolean;
  pnlChange: {
    percent: string;
    direction: 'up' | 'down';
  } | null;
}
