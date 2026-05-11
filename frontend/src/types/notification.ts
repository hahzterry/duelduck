import { LINK_FAQ_SELF_RESOLVE } from '~constants/main';
import { ErrorToastTextPart } from '~store/errorToastStore';
import colors from '~styles/colors';
import { CURRENCY } from '~types/general';

export enum NOTIFICATION_TYPE {
  SOLAutoswapSuccess = 0,
  SOLAutoswapFailed = 1,
  USDCAutoswapSuccess = 2,
  USDCAutoswapFailed = 3,

  DuelResolve = 10,
  DuelRefund = 11,
  VotedFor = 12,
  DuelModeration = 13,
  DuelPlayersJoined = 14,
  DuelEndingSoon = 15,
  AccountActivationFee = 16,
  NotificationAutoswapCompleted = 17,
  NotificationDisputeReportApproved = 18,
  NotificationDisputeReportIgnored = 19,
  NotificationIncorrectResolve = 20,
  NotificationSelfResolveRestricted = 21,
  NotificationSelfResolveRestored = 22,
}

export interface UnreadCountResponse {
  notifications: number;
}

// ================= BASE =================
export interface TxNotification {
  id: string;
  user_id: string;
  notification_type: NOTIFICATION_TYPE;
  created_at: string;
  is_read: boolean;
  data?: unknown[];
}

// ================= DATA TYPES =================
export interface VotedForNotificationData {
  duel_id: string;
  duel_name: string;
  voted_for: number;
}

export interface DuelResultNotificationData {
  duel_id: string;
  duel_name: string;
  voted_for: number;
  symbol?: string;
  payment_type?: CURRENCY;
  token_image_url?: string;
  amount: number;
  status: number;
}

export interface DuelModerationNotificationData {
  duel_id: string;
  duel_name: string;
  is_approved: boolean;
  cancellation_reason: string;
}

export interface DuelPlayersJoinedNotificationData {
  duel_id: string;
  duel_name: string;
  players_count: number;
}

export interface DuelEndingSoonNotificationData {
  duel_id: string;
  duel_name: string;
  deadline: number;
}

export interface AccountActivationFeeData {
  tokens_adjusted: number;
  sol_equivalent: number;
  usdc_equivalent: number;
  duel_id?: string;
}

export interface NotificationAutoswapCompletedData {
  from_tokens: {
    mint: string;
    symbol: string;
    amount: number;
  }[];
  to_symbol: string;
  to_amount: number;
  fee_usd: number;
}

export interface NotificationDisputeReportIgnoredData {
  duel_id: string;
  duel_question: string;
  admin_comment?: string;
}

export interface NotificationDisputeReportApprovedData {
  duel_id: string;
  duel_question: string;
  admin_comment?: string;
}

export interface NotificationIncorrectResolveData {
  duel_id: string;
  duel_question: string;
  admin_comment?: string;
}

// ================= MAP =================
export type NotificationDataMap = {
  [NOTIFICATION_TYPE.VotedFor]: VotedForNotificationData;
  [NOTIFICATION_TYPE.NotificationAutoswapCompleted]: NotificationAutoswapCompletedData;
  [NOTIFICATION_TYPE.NotificationDisputeReportIgnored]: NotificationDisputeReportIgnoredData;
  [NOTIFICATION_TYPE.NotificationDisputeReportApproved]: NotificationDisputeReportApprovedData;
  [NOTIFICATION_TYPE.NotificationIncorrectResolve]: NotificationIncorrectResolveData;
  [NOTIFICATION_TYPE.NotificationSelfResolveRestored]: null;
  [NOTIFICATION_TYPE.NotificationSelfResolveRestricted]: null;
  [NOTIFICATION_TYPE.DuelResolve]: DuelResultNotificationData;
  [NOTIFICATION_TYPE.AccountActivationFee]: AccountActivationFeeData;
  [NOTIFICATION_TYPE.DuelRefund]: DuelResultNotificationData;
  [NOTIFICATION_TYPE.DuelModeration]: DuelModerationNotificationData;
  [NOTIFICATION_TYPE.DuelPlayersJoined]: DuelPlayersJoinedNotificationData;
  [NOTIFICATION_TYPE.DuelEndingSoon]: DuelEndingSoonNotificationData;
};

// ================= FIXED GENERIC =================
export type TypedNotification<T extends keyof NotificationDataMap> = Omit<
  TxNotification,
  'data' | 'notification_type'
> & {
  notification_type: T;
  data: NotificationDataMap[T];
};

export interface CustomNotification {
  title: string;
  description: string | ErrorToastTextPart[];
  id: string;
}

// ================= UNION =================
export type DuelNotificationUnion =
  | TypedNotification<NOTIFICATION_TYPE.VotedFor>
  | TypedNotification<NOTIFICATION_TYPE.NotificationAutoswapCompleted>
  | TypedNotification<NOTIFICATION_TYPE.DuelResolve>
  | TypedNotification<NOTIFICATION_TYPE.DuelRefund>
  | TypedNotification<NOTIFICATION_TYPE.DuelModeration>
  | TypedNotification<NOTIFICATION_TYPE.DuelPlayersJoined>
  | TypedNotification<NOTIFICATION_TYPE.AccountActivationFee>
  | TypedNotification<NOTIFICATION_TYPE.NotificationSelfResolveRestricted>
  | TypedNotification<NOTIFICATION_TYPE.NotificationDisputeReportApproved>
  | TypedNotification<NOTIFICATION_TYPE.NotificationDisputeReportIgnored>
  | TypedNotification<NOTIFICATION_TYPE.NotificationIncorrectResolve>
  | TypedNotification<NOTIFICATION_TYPE.NotificationSelfResolveRestored>
  | TypedNotification<NOTIFICATION_TYPE.DuelEndingSoon>;

export const signTransactionNotification: CustomNotification = {
  title: 'CONFIRM YOU TRANSACTION',
  description: 'Please confirm transaction in your Wallet',
  id: 'confirm',
};

export const signatureTransactionNotification: CustomNotification = {
  title: 'CONFIRM YOU SIGNATURE',
  description: 'Please confirm creating signature in your Wallet',
  id: 'confirm',
};

export const switchWalletNotification: CustomNotification = {
  title: 'SWITCHING WALLET',
  description: 'Please approve wallet connection in your Wallet',
  id: 'switch-wallet',
};

export const walletConnectedNotification: CustomNotification = {
  title: 'WALLET CONNECTED',
  description: 'Your wallet has been successfully connected',
  id: 'wallet-connected',
};

export const fallbackSwitchNotification: CustomNotification = {
  title: 'CUSTODIAL WALLET ACTIVATED',
  description:
    'Failed to connect your wallet. You’ve been switched to a custodial wallet.',
  id: 'fallback-wallet',
};

export const switchedToCustodialNotification: CustomNotification = {
  title: 'SWITCHED TO CUSTODIAL WALLET',
  description: 'You’ve been switched to your custodial wallet.',
  id: 'switched-custodial',
};

export const reconnectingWalletNotification: CustomNotification = {
  title: 'RECONNECTING WALLET',
  description:
    'You are not connected right now. Reconnecting to your wallet...',
  id: 'reconnect-wallet',
};

export const failedToLinkWalletNotification: CustomNotification = {
  title: 'FAILED LINKING WALLET TO ACCOUNT',
  description: 'During linking process something went wrong.',
  id: 'failedToLinkWalletNotification',
};

export const maxWalletsNotification: CustomNotification = {
  title: 'MAX WALLET LIMIT REACHED',
  description: 'You have reached the maximum number of linked wallets.',
  id: 'max-wallets',
};

export const switchWalletBeforeRemoveNotification: CustomNotification = {
  title: 'SWITCH WALLET BEFORE REMOVING',
  description: 'Please switch to another wallet before removing current one.',
  id: 'switch-before-remove',
};

export const removeWalletSuccess: CustomNotification = {
  title: 'WALLET REMOVED SUCCESSFULLY',
  description: 'Your wallet has been removed from your account.',
  id: 'remove-wallet-success',
};

export const switchToWhitechainNotification: CustomNotification = {
  title: 'SWITCH TO WHITECHAIN WALLET',
  description:
    'This duel requires a Whitechain wallet. Please switch to your Whitechain wallet to continue.',
  id: 'switch-to-whitechain-notification',
};

export const switchToSolanaNotification: CustomNotification = {
  title: 'SWITCH TO SOLANA WALLET',
  description:
    'This duel requires a Solana wallet. Please switch to your Solana wallet to continue.',
  id: 'switch-to-solana-notification',
};

export const swapSuccessNotification: CustomNotification = {
  title: 'SWAP SUCCESSFUL',
  description:
    "Your swap has been completed successfully. Let's vote in duel now!",
  id: 'wbtSwapSuccessNotification',
};

export const failedToSignInXNotification: CustomNotification = {
  title: 'FAILED TO SIGN IN WITH X',
  description: 'During sign in with X something went wrong.',
  id: 'failed-sign-in-x',
};

export const successfulSignInXNotification: CustomNotification = {
  title: 'SUCCESSFUL SIGN IN WITH X',
  description: 'You have successfully signed in with X.',
  id: 'successful-sign-in-x',
};

export const failedToConnectXNotification: CustomNotification = {
  title: 'FAILED TO CONNECT X ACCOUNT',
  description: 'During connecting X account something went wrong.',
  id: 'failed-connect-x',
};

export const successfulConnectXNotification: CustomNotification = {
  title: 'SUCCESSFUL CONNECT X ACCOUNT',
  description: 'You have successfully connected your X account.',
  id: 'successful-connect-x',
};

export const tooManyRequestsNotification: CustomNotification = {
  title: 'TOO MANY REQUESTS',
  description:
    'You have made too many requests in a short period. Please try again later.',
  id: 'too-many-requests',
};

export const failerToCreateDuelNotification: CustomNotification = {
  title: 'FAILED TO CREATE DUEL',
  description: 'During duel creation something went wrong.',
  id: 'failed-create-duel',
};

export const failedToJoinDuelNotification: CustomNotification = {
  title: 'FAILED TO JOIN DUEL',
  description:
    'We couldn’t complete your duel join. Please check that your wallet contains at least 0.03 SOL required for transaction fees.',
  id: 'failed-join-duel',
};

export const failedToLoadDuelLogoNotification: CustomNotification = {
  title: 'LOGO FAILED TO LOAD',
  description:
    'We couldn’t load the duel logo image. Please regenerate or upload your own.',
  id: 'failed-load-duel-logo',
};

export const topUpWalletNotification: CustomNotification = {
  title: 'TOP UP YOUR WALLET',
  description: 'Please top up your wallet to continue.',
  id: 'top-up-wallet',
};

export const selfResolveRestrictedNotification: CustomNotification = {
  title: 'Self-Resolve Restricted',
  description: [
    {
      text: 'We’ve temporarily ',
      color: colors.textSecondary,
    },
    {
      text: 'restricted your ability to resolve duels',
      hoverColor: colors.primary,
      color: colors.textSecondary,
      href: LINK_FAQ_SELF_RESOLVE,
    },
    {
      text: ' following multiple reports of unfair outcomes. But you can still create and join duels.',
      color: colors.textSecondary,
    },
  ],
  id: 'self-resolve-restricted',
};

export const selfResolveRestoredNotification: CustomNotification = {
  title: 'Self-Resolve Restored',
  description:
    'Your ability to resolve duels has been restored. Thank you for your patience and for keeping the competition fair!',
  id: 'self-resolve-restored',
};

export const walletTxFailedNotification: CustomNotification = {
  title: 'TRANSACTION FAILED',
  description:
    'Could not send transaction to your wallet. Please reconnect and try again.',
  id: 'wallet-tx-failed',
};

export const walletTxTimedOutNotification: CustomNotification = {
  title: 'TRANSACTION TIMED OUT',
  description:
    'Wallet did not respond in time. Please open your wallet app and try again.',
  id: 'wallet-tx-timed-out',
};

export const openWalletToSignNotification: CustomNotification = {
  title: 'OPEN YOUR WALLET',
  description:
    'Tap to open your wallet app and confirm the pending transaction.',
  id: 'open-wallet-to-sign',
};
