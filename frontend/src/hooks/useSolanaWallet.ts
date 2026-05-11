'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import type { WalletAdapter } from '@solana/wallet-adapter-base';

import { RPC_URL } from '~constants/api';
import { useRefAndState } from '~hooks/useRefAndState';
import { CUSTOM_EVENT_KEYS } from '~types/general';
import {
  convertToBase64,
  createAppKitInstance,
  disconnect,
  extractWalletName,
  getWagmiAdapter,
  isSolWallet,
} from '~utils/appKitLazy';
import { customEvent } from '~utils/customEvent';
import { loadSolanaWeb3, loadWagmi } from '~utils/lazyWeb3';

const ASSETS = {
  SOL: {
    RPC_URL: RPC_URL,
    USDT: 'Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB',
    USDC: 'EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v',
    DECIMALS: 10 ** 6,
  },
};

export const useSolanaWallet = (withRehydrate?: boolean) => {
  const [connectedAddress, setConnectedAddress] = useState<any>(null);
  const [getIsLoading, setIsLoading, isLoading] = useRefAndState(false);
  const [walletProvider, setWalletProvider] = useState<WalletAdapter | null>(
    null,
  );

  const getAddressFromIdentityCache = (): string | null => {
    if (typeof window === 'undefined') return null;

    try {
      const identityCacheString = localStorage.getItem(
        '@appkit/identity_cache',
      );

      if (!identityCacheString) return null;

      const parsed = JSON.parse(identityCacheString);
      const address = Object.keys(parsed)?.[0];

      return address ?? null;
    } catch (error) {
      console.error('[getAddressFromIdentityCache] Failed to parse:', error);

      return null;
    }
  };

  const isWalletConnectFound = useRef(false);
  const lastRehydrateTime = useRef(0);

  const rehydrateFromAppKit = useCallback(async () => {
    const address = getAddressFromIdentityCache();

    if (!address) return;
    const isSolana = isSolWallet(address);

    try {
      const appKit = await createAppKitInstance();

      // Try for ~4s (20 * 200ms)
      for (let i = 0; i < 20; i++) {
        // 'wbt' | 'solana' | null
        const readEvm = async () => {
          try {
            const nsAddr =
              appKit.getAddressByChainNamespace?.('eip155') ??
              appKit.getAddress?.('eip155');
            const caip = appKit.getCaipAddress?.('eip155');
            const caipAddr = caip?.split(':').pop();
            const { getAccount: getWagmiAccount } = await loadWagmi();
            const wagmiAdapt = await getWagmiAdapter();
            const wagmiAcc = getWagmiAccount((wagmiAdapt as any).wagmiConfig);
            const evmAddress =
              nsAddr ??
              caipAddr ??
              (wagmiAcc.isConnected ? wagmiAcc.address : undefined);

            const evmProvider = appKit.getProvider?.('eip155');

            return {
              address: evmAddress,
              provider: evmProvider as unknown as WalletAdapter | undefined,
            };
          } catch {
            return { address: undefined, provider: undefined };
          }
        };

        const readSol = () => {
          const solAddress =
            appKit.getAddressByChainNamespace?.('solana') ??
            appKit.getAddress?.('solana');
          const solProvider = appKit.getProvider?.('solana');

          return {
            address: solAddress,
            provider: solProvider as unknown as WalletAdapter | undefined,
          };
        };

        let pick;

        if (isSolana) {
          pick = readSol();
        } else {
          pick = await readEvm();
        }

        const { address, provider } = pick;

        if (address && provider) {
          setConnectedAddress(address);
          setWalletProvider(provider);
          break;
        }

        await new Promise((r) => setTimeout(r, 200));
      }
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (window.self !== window.top || !withRehydrate) return;

    lastRehydrateTime.current = Date.now();
    void rehydrateFromAppKit();
    const onVis = () => {
      if (document.visibilityState !== 'visible' || getIsLoading('ref')) return;

      const now = Date.now();

      if (now - lastRehydrateTime.current < 3000) return;
      lastRehydrateTime.current = now;
      void rehydrateFromAppKit();
    };

    document.addEventListener('visibilitychange', onVis);

    return () => document.removeEventListener('visibilitychange', onVis);
  }, [withRehydrate]);

  function findWalletConnect() {
    const modal = document.querySelector('w3m-modal');

    const modalShadow = modal?.shadowRoot;
    const overlay = modalShadow?.querySelector(
      '[data-testid="w3m-modal-overlay"]',
    );
    const card = overlay?.querySelector('[data-testid="w3m-modal-card"]');

    if (card) {
      const router = card?.querySelector('w3m-router');
      const routerShadow = router?.shadowRoot;

      const routerContainer = routerShadow?.querySelector(
        'w3m-router-container',
      );

      const accountView = routerContainer?.querySelector('w3m-account-view');

      if (accountView) {
        const shadow = accountView?.shadowRoot;
        const defView = shadow?.querySelector('w3m-account-default-widget');
        const shadoView = defView?.shadowRoot;
        const wuiFlexElements = shadoView?.querySelectorAll('wui-flex');

        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-expect-error
        const secondWuiFlex = wuiFlexElements[2];
        const disconnectButton = secondWuiFlex?.querySelector(
          '[data-testid="disconnect-button"]',
        );
        const buttonShadow = disconnectButton?.shadowRoot;
        const button = buttonShadow?.querySelector('button');

        if (button) {
          button.click();
        }
      }

      const whiteWalletEl = routerContainer
        ?.querySelector('w3m-connect-view')
        ?.shadowRoot?.querySelector(
          'wui-flex wui-flex wui-flex w3m-wallet-login-list',
        )
        ?.shadowRoot?.querySelector('wui-flex w3m-connector-list')
        ?.shadowRoot?.querySelector(
          'wui-flex w3m-list-wallet[name="Whitewallet"]',
        );

      if (!whiteWalletEl) return;

      const button = whiteWalletEl.shadowRoot
        ?.querySelector('wui-list-wallet')
        ?.shadowRoot?.querySelector('button');

      if (button) {
        button.click();
        isWalletConnectFound.current = true;
      }
    }
  }

  const handleCreateSecret = useCallback(
    async (provider: WalletAdapter, address: string) => {
      const isSolana = isSolWallet(address);

      if (isSolana) {
        const encodedMessage = new TextEncoder().encode(address);
        // @ts-expect-error signMessage is present on Phantom/Solflare adapters
        const signature = await provider.signMessage(encodedMessage);

        return Buffer.from(signature).toString('base64');
      }

      const { signMessage } = await loadWagmi();
      const wagmiAdapt = await getWagmiAdapter();
      const signature = await signMessage((wagmiAdapt as any).wagmiConfig, {
        message: address,
      });

      return convertToBase64(signature);
    },
    [],
  );

  const connectWalletConnect = useCallback(async (noSignIn?: boolean) => {
    setIsLoading(true);
    await disconnect();

    const appKit = await createAppKitInstance();

    isWalletConnectFound.current = false;
    customEvent.emit(CUSTOM_EVENT_KEYS.TOGGLE_SCROLL, false);
    await appKit.open();

    const interval = setInterval(async () => {
      if (!isWalletConnectFound.current && appKit.isOpen()) findWalletConnect();

      const nsAddr =
        appKit.getAddressByChainNamespace?.('eip155') ??
        appKit.getAddress?.('eip155');
      const caip = appKit.getCaipAddress?.('eip155');
      const caipAddr = caip?.split(':').pop();
      const { getAccount: getWagmiAccount } = await loadWagmi();
      const wagmiAdapt = await getWagmiAdapter();
      const wagmiAcc = getWagmiAccount((wagmiAdapt as any).wagmiConfig);
      const evmAddress =
        nsAddr ??
        caipAddr ??
        (wagmiAcc.isConnected ? wagmiAcc.address : undefined);

      const evmProvider = appKit.getProvider?.('eip155');

      if (evmAddress && evmProvider) {
        if (appKit.isOpen()) await appKit.close();
        customEvent.emit(CUSTOM_EVENT_KEYS.TOGGLE_SCROLL, true);
        clearInterval(interval);
        clearTimeout(timeout);

        setConnectedAddress(evmAddress);
        setWalletProvider(evmProvider as unknown as WalletAdapter);

        setIsLoading(false);
      }

      customEvent.emit(CUSTOM_EVENT_KEYS.TOGGLE_SCROLL, !appKit.isOpen());
    }, 400);

    const timeout = setTimeout(
      async () => {
        if (appKit.isOpen()) {
          clearInterval(interval);
          setIsLoading(false);
          customEvent.emit(CUSTOM_EVENT_KEYS.TOGGLE_SCROLL, true);
          await appKit.close();
          connectWalletConnect(noSignIn);
        }
      },
      3 * 60 * 1000,
    );

    return () => {
      clearTimeout(timeout);
      clearInterval(interval);
      setIsLoading(false);
    };
  }, []);

  const connect = useCallback(async () => {
    return new Promise((resolve, reject) => {
      (async () => {
        setIsLoading(true);
        await disconnect();

        const appKit = await createAppKitInstance();

        customEvent.emit(CUSTOM_EVENT_KEYS.TOGGLE_SCROLL, false);
        await appKit.open();

        let countIteration = 0;

        const interval = setInterval(async () => {
          const address = appKit.getAddress?.();
          const provider = appKit.getWalletProvider?.();

          countIteration++;

          if (countIteration >= 60) {
            reject('error connect');
          }

          if (address && provider) {
            if (appKit.isOpen()) await appKit.close();
            customEvent.emit(CUSTOM_EVENT_KEYS.TOGGLE_SCROLL, true);
            clearInterval(interval);

            setConnectedAddress(address);
            setWalletProvider(provider as WalletAdapter);
            resolve(() => {
              clearInterval(interval);
              setIsLoading(false);
            });

            setIsLoading(false);
          }

          const nsAddr =
            appKit.getAddressByChainNamespace?.('eip155') ??
            appKit.getAddress?.('eip155');
          const caip = appKit.getCaipAddress?.('eip155');
          const caipAddr = caip?.split(':').pop();
          const { getAccount: getWagmiAccount } = await loadWagmi();
          const wagmiAdapt = await getWagmiAdapter();
          const wagmiAcc = getWagmiAccount((wagmiAdapt as any).wagmiConfig);
          const evmAddress =
            nsAddr ??
            caipAddr ??
            (wagmiAcc.isConnected ? wagmiAcc.address : undefined);

          const evmProvider = appKit.getProvider?.('eip155');

          if (evmAddress && evmProvider) {
            if (appKit.isOpen()) await appKit.close();
            customEvent.emit(CUSTOM_EVENT_KEYS.TOGGLE_SCROLL, true);
            clearInterval(interval);
            resolve(() => {
              clearInterval(interval);
              setIsLoading(false);
            });

            setConnectedAddress(evmAddress);
            setWalletProvider(evmProvider as unknown as WalletAdapter);

            setIsLoading(false);
          }

          customEvent.emit(CUSTOM_EVENT_KEYS.TOGGLE_SCROLL, !appKit.isOpen());
        }, 1000);
      })();
    });
  }, []);

  const signAndSendTransactionFromBackend = useCallback(
    async (base64Tx: string): Promise<string | null> => {
      console.log('[tx] signAndSendTransactionFromBackend: initiated');

      if (!connectedAddress) {
        console.log('[tx] No connected address, reconnecting');
        await connect();
      }

      if (!walletProvider) {
        console.error('[tx] No wallet provider available');

        return null;
      }

      if (typeof (walletProvider as any).signTransaction !== 'function') {
        console.error('[tx] Provider does not support signTransaction');

        return null;
      }

      try {
        setIsLoading(true);

        const { Connection, Transaction } = await loadSolanaWeb3();
        const connection = new Connection(ASSETS.SOL.RPC_URL);
        const rawTx =
          typeof base64Tx === 'string'
            ? Buffer.from(base64Tx, 'base64')
            : Buffer.from(base64Tx as unknown as Uint8Array);
        const transaction = Transaction.from(rawTx);

        // On mobile external browsers, try to open wallet app for signing
        const env = getEnv();

        if (env.isMobile && !env.isWalletInApp) {
          console.log('[tx] Mobile external browser detected, nudging wallet');
          const walletName = extractWalletName(walletProvider).toLowerCase();

          if (walletName.includes('phantom')) {
            window.open(
              'https://phantom.app/ul/browse/' +
                encodeURIComponent(window.location.href),
              '_blank',
            );
          } else if (walletName.includes('solflare')) {
            window.open(
              'https://solflare.com/ul/v1/browse/' +
                encodeURIComponent(window.location.href),
              '_blank',
            );
          }
        }

        console.log('[tx] Requesting wallet signature');

        const signedTx = await Promise.race([
          // @ts-expect-error signTransaction exists on wallet providers
          walletProvider.signTransaction(transaction),
          new Promise<never>((_, reject) =>
            setTimeout(
              () => reject(new Error('Transaction signing timed out')),
              90_000,
            ),
          ),
        ]);

        console.log('[tx] Wallet signature received, broadcasting');

        const txHash = await connection.sendRawTransaction(
          signedTx.serialize(),
        );

        console.log('[tx] Transaction sent:', txHash);

        return txHash;
      } catch (error) {
        console.error('[signAndSendTransactionFromBackend error]', error);

        return null;
      } finally {
        setIsLoading(false);
      }
    },
    [walletProvider, connectedAddress],
  );

  const waitForSolanaConnection = useCallback(
    async (timeoutMs = 12000, tickMs = 250) => {
      const appKit = await createAppKitInstance();
      const start = Date.now();

      while (Date.now() - start < timeoutMs) {
        const address =
          appKit.getAddressByChainNamespace?.('solana') ??
          appKit.getAddress?.('solana') ??
          appKit.getAddress?.();

        const provider =
          (appKit.getProvider?.('solana') as unknown as
            | WalletAdapter
            | undefined) ??
          (appKit.getWalletProvider?.() as WalletAdapter | undefined);

        if (address && provider) return { address, provider };
        await new Promise((r) => setTimeout(r, tickMs));
      }

      throw new Error('Timed out waiting for Solana connection');
    },
    [],
  );

  const connectByWalletId = useCallback(async (walletId: string) => {
    try {
      setIsLoading(true);
      await disconnect();

      const appKit = await createAppKitInstance();

      const solAdapter: any = (appKit as any).chainAdapters?.solana;

      await solAdapter?.connect?.({
        id: walletId,
        rpcUrl: ASSETS.SOL.RPC_URL,
      });

      const { address, provider } = await waitForSolanaConnection();

      setConnectedAddress(address);
      setWalletProvider(provider);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const connectPhantom = useCallback(
    () =>
      connectByWalletId(
        'a797aa35c0fadbfc1a53e7f675162ed5226968b44a19ee3d24385c64d1d3c393',
      ),
    [],
  );

  const connectSolflare = useCallback(
    () =>
      connectByWalletId(
        '1ca0bdd4747578705b1939af023d120677c64fe6ca76add81fda36e350605e79',
      ),
    [],
  );

  // --- utils: env detection (place near top of hook) ---
  const getEnv = () => {
    const ua =
      (typeof navigator !== 'undefined' ? navigator.userAgent : '') || '';
    const isAndroid = /Android/i.test(ua);
    const isIphone = /iPhone/i.test(ua);
    const isIpad =
      /iPad/i.test(ua) ||
      (typeof navigator !== 'undefined' &&
        (navigator as any).platform === 'MacIntel' &&
        (navigator as any).maxTouchPoints > 1);
    const isIOS = isIphone || isIpad;
    const isMobile = isAndroid || isIOS;

    // Wallet in-app browsers (UA + injected provider on mobile)
    const hasMobileWindow = typeof window !== 'undefined' && isMobile;
    const isPhantomInApp =
      /Phantom/i.test(ua) ||
      !!(hasMobileWindow && (window as any).phantom?.solana);
    const isSolflareInApp =
      /Solflare/i.test(ua) || !!(hasMobileWindow && (window as any).solflare);
    const isWalletInApp = isPhantomInApp || isSolflareInApp;

    // Common social in-app browsers
    const isFB = /\bFBAN|FBAV|FB_IAB\b/i.test(ua);
    const isIG = /\bInstagram\b/i.test(ua);
    const isTG = /\bTelegram\b/i.test(ua);
    const isDiscord = /\bDiscord\b/i.test(ua);
    const isTikTok = /\bTikTok|Musical\.ly\b/i.test(ua);
    const isChromeWV = /\bwv\b/i.test(ua) || /; wv\)/i.test(ua);

    const isInAppBrowser =
      isWalletInApp ||
      isFB ||
      isIG ||
      isTG ||
      isDiscord ||
      isTikTok ||
      isChromeWV;

    return {
      ua,
      isMobile,
      isAndroid,
      isIOS,
      isIpad,
      isIphone,
      isInAppBrowser,
      isWalletInApp,
    };
  };

  // --- utils: open wallet in-app browser via deeplink (must be called from a user gesture) ---
  const openWalletDeeplink = (wallet: 'phantom' | 'solflare', url?: string) => {
    const target = encodeURIComponent(url ?? window.location.href);
    const ref = encodeURIComponent(window.location.origin);

    const deeplink =
      wallet === 'phantom'
        ? // https://phantom.app/ul/browse/<url>?ref=<ref>
          `https://phantom.app/ul/browse/${target}?ref=${ref}`
        : // https://solflare.com/ul/v1/browse/<url>?ref=<ref>
          `https://solflare.com/ul/v1/browse/${target}?ref=${ref}`;

    // Must be a direct navigation from the click handler for iOS universal links
    window.location.href = deeplink;
  };

  // --- helper: try deeplink first; fallback to programmatic connect if page didn’t leave ---
  const openWithFallback = async (
    wallet: 'phantom' | 'solflare',
    programmaticConnect: () => Promise<any>,
  ) => {
    const env = getEnv();

    if (env.isMobile && !env.isInAppBrowser) {
      openWalletDeeplink(wallet);

      return;
    }

    return programmaticConnect();
  };

  const connectPhantomMobileAware = useCallback(() => {
    return openWithFallback('phantom', () => connectPhantom());
  }, []);

  const connectSolflareMobileAware = useCallback(() => {
    return openWithFallback('solflare', () => connectSolflare());
  }, []);

  return {
    isLoading,
    connectedAddress: connectedAddress || getAddressFromIdentityCache(),
    walletProvider,
    connect,
    connectPhantom: connectPhantomMobileAware,
    connectSolflare: connectSolflareMobileAware,
    connectWalletConnect,
    signAndSendTransactionFromBackend,
    handleCreateSecret,
    disconnect,
  };
};
