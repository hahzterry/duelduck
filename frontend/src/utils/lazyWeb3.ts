/**
 * Lazy-loaded web3 utilities to reduce initial bundle size
 * These modules are only loaded when wallet connection is initiated
 */

// Lazy load ethers
export const loadEthers = async () => {
  const { ethers, formatEther, JsonRpcProvider } = await import('ethers');

  return { ethers, formatEther, JsonRpcProvider };
};

// Lazy load Solana web3
export const loadSolanaWeb3 = async () => {
  const { Connection, PublicKey, Transaction } =
    await import('@solana/web3.js');

  return { Connection, PublicKey, Transaction };
};

// Lazy load Solana SPL Token
export const loadSolanaToken = async () => {
  const {
    createAssociatedTokenAccountInstruction,
    createTransferInstruction,
    getAccount,
    getAssociatedTokenAddress,
  } = await import('@solana/spl-token');

  return {
    createAssociatedTokenAccountInstruction,
    createTransferInstruction,
    getAccount,
    getAssociatedTokenAddress,
  };
};

// Lazy load Wagmi
export const loadWagmi = async () => {
  const { getAccount, signMessage } = await import('@wagmi/core');

  return { getAccount, signMessage };
};

// Lazy load AppKit
export const loadAppKit = async () => {
  const { createAppKit } = await import('@reown/appkit');
  const { solana } = await import('@reown/appkit/networks');
  // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  // @ts-expect-error
  const { BaseWalletAdapter, SolanaAdapter } =
    await import('@reown/appkit-adapter-solana');
  const { PhantomWalletAdapter, SolflareWalletAdapter } =
    await import('@solana/wallet-adapter-wallets');

  return {
    createAppKit,
    solana,
    BaseWalletAdapter,
    SolanaAdapter,
    PhantomWalletAdapter,
    SolflareWalletAdapter,
  };
};

// Lazy load base64 utilities
export const loadBase64Utils = async () => {
  const { base64 } = await import('@scure/base');

  return { base64 };
};
