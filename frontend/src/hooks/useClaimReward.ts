'use client';

import { useCallback, useEffect, useRef, useState } from 'react';

import { LINK_DD } from '~constants/api';
import { ClaimRewardData } from '~types/claimReward';

type ClaimStep = 'share' | 'claiming' | 'success' | 'error';

export const CLAIM_REWARD_SHARE_CARD_ID = 'claim-reward-share-card';

interface UseClaimRewardParams {
  claimRewardData: ClaimRewardData | null;
  onClose: () => void;
}

const getClaimTxHash = (response: unknown): string | undefined => {
  if (!response || typeof response !== 'object') return undefined;

  if ('data' in response) {
    const axiosResponse = response as { data?: { claim_tx_hash?: string } };

    return axiosResponse.data?.claim_tx_hash;
  }

  const leaderboardResponse = response as { tx_hash?: string };

  return leaderboardResponse.tx_hash;
};

export const useClaimReward = ({
  claimRewardData,
  onClose,
}: UseClaimRewardParams) => {
  const rewardAmount = claimRewardData?.rewardAmount ?? 0;
  const formattedReward = rewardAmount.toLocaleString('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });

  const [step, setStep] = useState<ClaimStep>('claiming');
  const [secondsLeft, setSecondsLeft] = useState(60);
  const [errorMsg, setErrorMsg] = useState('');
  const [isClaiming, setIsClaiming] = useState(false);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const isClaimingRef = useRef(false);
  const isInitialClaimRewardDataEffectRef = useRef(true);

  const clearTimer = useCallback(() => {
    if (timerRef.current) {
      clearInterval(timerRef.current);
      timerRef.current = null;
    }
  }, []);

  const handleClaim = useCallback(async () => {
    if (isClaimingRef.current) return;

    isClaimingRef.current = true;
    setIsClaiming(true);

    try {
      if (claimRewardData?.isMock) {
        clearTimer();
        claimRewardData.onClaimSuccess?.({ txHash: 'mock-tx-hash' });
        setStep('share');

        return;
      }

      if (!claimRewardData?.claimId) {
        setStep('error');
        setErrorMsg('Reward source not found');

        return;
      }

      if (!claimRewardData.claimRequest) {
        setStep('error');
        setErrorMsg('Claim request is not configured');

        return;
      }

      const response = await claimRewardData.claimRequest(
        claimRewardData.claimId,
      );

      const txHash = claimRewardData.getClaimTxHash
        ? claimRewardData.getClaimTxHash(response)
        : getClaimTxHash(response);

      claimRewardData.onClaimSuccess?.({ txHash });
      clearTimer();
      setStep('share');
    } catch {
      clearTimer();
      setStep('error');
      setErrorMsg('Failed to claim reward. Please try again.');
    } finally {
      isClaimingRef.current = false;
      setIsClaiming(false);
    }
  }, [claimRewardData, clearTimer]);

  const startClaiming = useCallback(() => {
    if (isClaimingRef.current) return;

    clearTimer();
    setStep('claiming');
    setErrorMsg('');
    setSecondsLeft(60);

    timerRef.current = setInterval(() => {
      setSecondsLeft((prev) => {
        if (prev <= 1) {
          clearTimer();

          return 0;
        }

        return prev - 1;
      });
    }, 1000);

    void handleClaim();
  }, []);

  useEffect(() => {
    return clearTimer;
  }, []);

  useEffect(() => {
    const timeoutId = setTimeout(startClaiming, 0);

    return () => {
      clearTimeout(timeoutId);
    };
  }, []);

  useEffect(() => {
    clearTimer();
    setStep('claiming');
    setErrorMsg('');
    setSecondsLeft(60);
    isClaimingRef.current = false;
    setIsClaiming(false);

    if (isInitialClaimRewardDataEffectRef.current) {
      isInitialClaimRewardDataEffectRef.current = false;

      return;
    }

    if (!claimRewardData) return;

    const timeoutId = setTimeout(startClaiming, 0);

    return () => {
      clearTimeout(timeoutId);
    };
  }, [claimRewardData]);

  const handleSkip = useCallback(() => {
    clearTimer();
    setStep('success');
  }, []);

  const handleShareToX = useCallback(async () => {
    try {
      const html2canvasModule = await import('html2canvas');
      const html2canvas = html2canvasModule.default;
      const element =
        document.getElementById(CLAIM_REWARD_SHARE_CARD_ID) ||
        document.getElementById('mention-share-card');

      if (element) {
        const canvas = await html2canvas(element, {
          scale: 2,
          useCORS: true,
        });
        const dataUrl = canvas.toDataURL('image/png');
        const downloadLink = document.createElement('a');

        downloadLink.href = dataUrl;
        downloadLink.download = 'duel-duck-reward.png';
        downloadLink.setAttribute('data-ignore-outer-click', 'true');
        document.body.appendChild(downloadLink);
        downloadLink.click();
        downloadLink.remove();
      }
    } catch (e) {
      console.error('Failed to generate share card', e);
    }

    const pageLink = claimRewardData?.sharePageUrl || `${LINK_DD}/leaderboard`;

    const contextName = claimRewardData?.contextName || 'Leaderboard';

    const tweetText = claimRewardData?.sharePostText
      ? claimRewardData.sharePostText.replace('{amount}', formattedReward)
      : `I just claimed $${formattedReward} USDC` +
        ` from @duel_duck ${contextName}!\n\n` +
        `Join here 👉 ${pageLink}`;

    const shareUrl =
      `https://x.com/intent/tweet?text=` + `${encodeURIComponent(tweetText)}`;

    window.open(shareUrl, '_blank');
    clearTimer();
    setStep('success');
  }, [formattedReward, claimRewardData]);

  const handleBack = useCallback(() => {
    onClose();
  }, [onClose]);

  return {
    step,
    secondsLeft,
    errorMsg,
    rewardAmount,
    formattedReward,
    isClaiming,
    handleShareToX,
    handleSkip,
    startClaiming,
    handleBack,
  };
};
