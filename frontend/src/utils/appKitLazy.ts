'use client';

import { loadAppKit } from './lazyWeb3';

const PROJECT_ID = '94fc571c191af03a8324d2f18225cd47';

const METADATA = {
  name: 'Duel Duck',
  description: 'Duel Duck',
  url: 'https://duelduck.com/',
  icons: ['https://duelduck.com/logo.svg'],
};

let appKitDual: any = null;
let wagmiAdapter: any = null;
let solanaAdapter: any = null;

export async function getWagmiAdapter() {
  if (wagmiAdapter) return wagmiAdapter;

  return wagmiAdapter;
}

export async function getSolanaAdapter() {
  if (solanaAdapter) return solanaAdapter;

  const { SolanaAdapter, PhantomWalletAdapter, SolflareWalletAdapter } =
    await loadAppKit();

  solanaAdapter = new SolanaAdapter({
    wallets: [new PhantomWalletAdapter(), new SolflareWalletAdapter()],
  });

  return solanaAdapter;
}

export async function createAppKitDualInstance() {
  if (appKitDual) return appKitDual;

  const { createAppKit, solana } = await loadAppKit();
  const solAdapter = await getSolanaAdapter();

  appKitDual = createAppKit({
    themeVariables: { '--w3m-font-family': 'Poppins' },
    adapters: [solAdapter],
    networks: [solana],
    projectId: PROJECT_ID,
    enableReconnect: false,
    metadata: METADATA,
    showWallets: true,
    allWallets: 'SHOW',
    featuredWalletIds: [
      'a797aa35c0fadbfc1a53e7f675162ed5226968b44a19ee3d24385c64d1d3c393',
      '1ca0bdd4747578705b1939af023d120677c64fe6ca76add81fda36e350605e79',
      '8b68d7428a3a80a7390a21c1a0019348d58b3d29bd4efbd9da3cb9f231251933',
    ],
    includeWalletIds: [
      'a797aa35c0fadbfc1a53e7f675162ed5226968b44a19ee3d24385c64d1d3c393',
      '1ca0bdd4747578705b1939af023d120677c64fe6ca76add81fda36e350605e79',
      '8b68d7428a3a80a7390a21c1a0019348d58b3d29bd4efbd9da3cb9f231251933',
    ],
    features: {
      swaps: false,
      onramp: false,
      email: false,
      socials: false,
      analytics: false,
    },
  });

  return appKitDual;
}

export async function createAppKitInstance() {
  return createAppKitDualInstance();
}

export const ensureCorrectAppKitNetwork = (
  chain: 'solana' | 'wbt',
  appKit: any,
) => {
  try {
    if (!appKit) throw new Error('AppKit instance not provided');

    if (chain === 'solana') {
      const caip = 'solana:5eykt4UsFv8P8NJdTREpY1vzqKqZKvdp';

      (appKit as any).setActiveNamespace?.('solana');
      (appKit as any).setActiveCaipNetworkId?.(caip);
      (appKit as any).networkController?.setActiveCaipNetwork?.(caip);
      (appKit as any).networkController?.setActiveNamespace?.('solana');
      (appKit as any).state?.setActiveCaipNetworkId?.(caip);
      (appKit as any).state?.setActiveNamespace?.('solana');

      localStorage.setItem('@appkit/active_namespace', 'solana');
      localStorage.setItem('@appkit/active_caip_network_id', caip);
      localStorage.setItem('chain', 'solana');
    } else {
      const caip = 'eip155:1875';

      (appKit as any).setActiveNamespace?.('eip155');
      (appKit as any).setActiveCaipNetworkId?.(caip);
      (appKit as any).networkController?.setActiveCaipNetwork?.(caip);
      (appKit as any).networkController?.setActiveNamespace?.('eip155');
      (appKit as any).state?.setActiveCaipNetworkId?.(caip);
      (appKit as any).state?.setActiveNamespace?.('eip155');

      localStorage.setItem('@appkit/active_namespace', 'eip155');
      localStorage.setItem('@appkit/active_caip_network_id', caip);
      localStorage.setItem('chain', 'wbt');
    }

    console.info('[AppKit] Active network fully enforced:', chain);
  } catch (err) {
    console.warn('[AppKit] Failed to enforce network', chain, err);
  }
};

export const convertToBase64 = (sig: string): string | null => {
  try {
    const sigWithoutPrefix = sig.startsWith('0x') ? sig.slice(2) : sig;

    if (sigWithoutPrefix.length !== 130) {
      throw new Error(
        `Invalid hex length: ${sigWithoutPrefix.length}, expected 130`,
      );
    }

    const sigBytes = Buffer.from(sigWithoutPrefix, 'hex');

    if (sigBytes.length !== 65) {
      throw new Error(`Invalid bytes length: ${sigBytes.length}, expected 65`);
    }

    const base64Sig = sigBytes.toString('base64');

    if (base64Sig.length !== 88) {
      throw new Error(
        `Invalid Base64 length: ${base64Sig.length}, expected 88`,
      );
    }

    return base64Sig;
  } catch (err) {
    console.error('Conversion error:', err);

    return null;
  }
};

export const extractWalletName = (provider: any): string => {
  if (!provider) return 'Unknown';
  if (provider.name) return provider.name;

  const wcName =
    provider.session?.peer?.metadata?.name ||
    provider.session?.peerMetadata?.name;

  return wcName || 'Unknown Wallet';
};

export const clearAppKitStorage = () => {
  Object.keys(localStorage)
    .filter(
      (k) =>
        k.startsWith('@appkit/') ||
        k.startsWith('wagmi') ||
        k.startsWith('@reown/'),
    )
    .forEach((k) => localStorage.removeItem(k));

  localStorage.removeItem('chain');
  localStorage.removeItem('wallet');
  localStorage.removeItem('withWallet');

  localStorage.removeItem('cbwsdk.store');

  Object.keys(localStorage)
    .filter(
      (k) =>
        k.startsWith('wc@') ||
        k.startsWith('WALLETCONNECT_') ||
        k.startsWith('wc_') ||
        k.includes('walletconnect'),
    )
    .forEach((k) => localStorage.removeItem(k));

  Object.keys(localStorage)
    .filter((k) => k.includes('-walletlink:'))
    .forEach((k) => localStorage.removeItem(k));

  Object.keys(localStorage)
    .filter(
      (k) =>
        k.includes('phantom') || k.includes('solflare') || k.includes('solana'),
    )
    .forEach((k) => localStorage.removeItem(k));

  sessionStorage.clear();

  console.log('[clearAppKitStorage] All storage cleared');
};

export const disconnect = async () => {
  try {
    const appKitDual = await createAppKitDualInstance();

    await (appKitDual as any)?.disconnect?.().catch(() => {});

    const solAdapterDual: any = (appKitDual as any)?.chainAdapters?.solana;

    if (solAdapterDual && typeof solAdapterDual.disconnect === 'function') {
      try {
        await solAdapterDual.disconnect();
      } catch (err) {
        console.warn('[disconnect] solAdapterDual failed', err);
      }
    }

    const evmAdapter: any = (appKitDual as any)?.chainAdapters?.eip155;

    if (evmAdapter && typeof evmAdapter.disconnect === 'function') {
      try {
        await evmAdapter.disconnect();
      } catch (err) {
        console.warn('[disconnect] evmAdapter failed', err);
      }
    }

    clearAppKitStorage();

    console.log('[disconnect] All instances disconnected and caches cleared');
  } catch (err) {
    console.error('[disconnect error]', err);
  }
};

export const isSolWallet = (address: string): boolean => {
  return !address.startsWith('0x');
};
