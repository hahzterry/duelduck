import { Duel, TOKEN_SYMBOLS, USER_VOTE } from '~types/duel';

export const MAX_COMMISSION = 10;

export const MIN_COMMISSION = 0;

export const mockDuel: Duel = {
  id: '12345678-90ab-cdef-1234-567890abcdef',
  owner_id: 'aboba-id-228-1488',
  resolved_by: '00000000-0000-0000-0000-000000000000',
  approved_by: '00000000-0000-0000-0000-000000000000',
  room_number: 101,
  symbol: TOKEN_SYMBOLS.USDC,
  players_count: 5,
  username: 'mock_user',
  status: 3,
  image_url: '',
  bg_url: '',
  topic: 'gaming',
  subtopic: 'Esports',
  duel_type: 'Top Scorer',
  entities: [8, 22, 19],
  question: 'Will Team Alpha outscore Team Beta?',
  source_of_truth: 'https://www.mockesports.com/match/12345',
  deadline: '2024-12-10T12:00:00Z',
  duel_price: 500,
  commission: 5,
  duel_info: {
    direction: 1,
    event: ['Esports Championship'],
    team: ['Team Alpha', 'Team Beta'],
  },
  tx_hash: '0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890',
  final_result: null,
  cancellation_reason: '',
  created_at: '2024-12-01T10:00:00Z',
  updated_at: '2024-12-01T10:00:00Z',
  yes_count: 3,
  no_count: 2,
  joined: true,
  your_answer: USER_VOTE.YES,
} as unknown as Duel;
