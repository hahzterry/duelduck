export enum CHAIN {
  SOLANA = 0,
  WHITECHAIN = 1,
}

export enum WALLET_TYPE {
  CUSTODIAL = 'custodial',
  EXTERNAL = 'external',
}

export interface Wallet {
  id: string;
  user_id: string;
  address: string;
  label?: string;
  name: string;
  chain: CHAIN; // 0 = SOLANA, 1 = WHITECHAIN
  type: WALLET_TYPE; // "custodial" | "external"
  verified_at: string | null; // ISO string or null if not verified
  created_at: string; // ISO string
  updated_at: string; // ISO string
}

export interface WalletListResponse {
  wallets: Wallet[];
}

export interface WalletResponse {
  wallet: Wallet;
}

export interface LinkWalletRequest {
  address: string;
  secret: string;
  chain: number;
  label: string;
}

export interface LinkWalletResponse {
  wallet: Wallet;
}

export interface ActivateWalletRequest {
  wallet_id: string;
}

export interface ActivateWalletResponse {
  user: {
    id: string;
    active_wallet_id: string;
    role: string;
    is_premium: boolean;
  };
  jwt_info: {
    access_token: string;
    refresh_token: string;
  };
  active_wallet: string;
}

export interface UpdateWalletRequest {
  name?: string;
  label?: string;
}

export interface UpdateWalletResponse {
  wallet: Wallet;
}
