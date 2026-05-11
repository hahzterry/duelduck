import { StaticImport } from 'next/dist/shared/lib/get-img-props';

export enum TRANSACTION_HISTORY_TYPE {
  DEPOSIT = 5,
  WITHDRAWAL = 6,
  DUEL_PREDICTION = 1,
  DUEL_REFUND = 2,
  DUEL_COMMISSION = 3,
  DUEL_REWARD = 4,
}

export const transactionTypesMap: Record<TRANSACTION_HISTORY_TYPE, string> = {
  [TRANSACTION_HISTORY_TYPE.DEPOSIT]: 'Deposit',
  [TRANSACTION_HISTORY_TYPE.WITHDRAWAL]: 'Withdrawal',
  [TRANSACTION_HISTORY_TYPE.DUEL_PREDICTION]: 'Duel Prediction',
  [TRANSACTION_HISTORY_TYPE.DUEL_REFUND]: 'Duel Refund',
  [TRANSACTION_HISTORY_TYPE.DUEL_COMMISSION]: 'Duel Commission',
  [TRANSACTION_HISTORY_TYPE.DUEL_REWARD]: 'Duel Reward',
};

export interface TransactionHistory {
  amount: number;
  decimals: number;
  balanceChange: number;
  transactionType: TRANSACTION_HISTORY_TYPE;
  txHash: string;
  date: number;
  currency: string;
  tokenName: string;
  tokenLogo: string;
  mintToken: string;
}

export interface TransactionHistoryDuckies {
  sender_address: string;
  recipient_address: string;
  amount: number;
  created_at: string; // ISO 8601 date string
}

export interface Token {
  tokenAddress: string;
  tokenName: string;
  tokenSymbol: string;
  tokenLogo: string | StaticImport;
  amount: number;
  balance: number;
  price: number;
}

export type Tokens = Token[];

export interface TokenMetadata {
  tokenName: string;
  tokenSymbol: string;
  tokenLogo: string | StaticImport;
  tokenAddress: string;
}

export type TokensMetaData = TokenMetadata[];

export enum NOTIFICATION_TYPE {
  SOLAutoswapSuccess = 0,
  SOLAutoswapFailed = 1,
  USDCAutoswapSuccess = 2,
  USDCAutoswapFailed = 3,
}
