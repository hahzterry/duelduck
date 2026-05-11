import { Referral } from '~types/referral';

export const mockReferralUsers: Referral[] = [...new Array(10)]
  .map(() => ({
    referrer_id: '11111111-aaaa-bbbb-cccc-000000000001',
    referral_id: '22222222-aaaa-bbbb-cccc-000000000001',
    username: 'john_doe',
    income_usdc: 12.5,
    income_dp: 300,
    referral_commission_end: '2025-07-01T00:00:00Z',
    created_at: '2025-06-01T12:00:00Z',
  }))
  .map((user) => ({
    rank: 1,
    username: user.username,
    earnedUsdc: user.income_usdc,
    earnedDdp: user.income_dp,
    date: user.created_at,
  }));
