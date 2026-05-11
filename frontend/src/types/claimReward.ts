export type ClaimRewardType =
  | 'leaderboard'
  | 'mention_challenge'
  | 'multi_duel_shareboard';

export interface ClaimRewardData {
  claimId: string;
  claimType: ClaimRewardType;
  rewardAmount: number;
  isMock?: boolean;
  claimRequest?: (claimId: string) => Promise<unknown>;
  getClaimTxHash?: (response: unknown) => string | undefined;
  contextName?: string;
  shareStepText?: string;
  successStepText?: string;
  sharePostText?: string;
  sharePageUrl?: string;
  shareCardBannerUrl?: string;
  modalBannerUrl?: string;
  logoUrl?: string;
  destinationUrl?: string;
  destinationLabel?: string;
  backButtonText?: string;
  onClaimSuccess?: (params: { txHash?: string }) => void;
}
