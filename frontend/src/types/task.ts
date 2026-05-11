import { CURRENCY } from '~types/general';

export interface Task {
  id: number;
  name: string;
  description: string;
  completed: boolean;
  auto_completion: boolean;
  reward_claimed: number;
  completion_count: number;
  reward: number;
  currency: CURRENCY;
  xp: number;
  completion_limit: number;
  completion_limit_period: number;
  deadline: number | null;
  task_notes: string;
  link: string;
  created_at: string;
}

export type TaskStatus = 'active' | 'completed';

export interface Stats {
  id: string;
  reward_theme: string;
  level: number;
  current_xp: number;
  next_level_xp: number;
  ddp_earned: number;
  usdc_earned: number;
}
