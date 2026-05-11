import Cookies from 'js-cookie';

import {
  INVITED_BY_QUERY_KEY,
  MULTI_DUEL_REFERRAL_CODE_QUERY_KEY,
  REFERRAL_TOKEN_QUERY_KEY,
} from '~constants/referralTokens';
import { saveToken } from '~utils/saveToken';

class CookieManager {
  static setItem(
    name: string,
    value: string,
    expires: number = Date.now() + 31449600, // 1 year by default
  ) {
    Cookies.set(name, value, {
      expires: new Date(expires),
      secure: true,
      // TODO - for embed page, but when "none", can steal from other site
      sameSite: 'none', // strict
    });
  }

  static getItem(name: string): string | undefined {
    return Cookies.get(name);
  }

  static removeItem(name: string) {
    Cookies.remove(name);
  }

  static clear() {
    const cookies = Object.keys(Cookies.get());

    for (const cookie of cookies) {
      Cookies.remove(cookie);
    }

    localStorage.removeItem('withWallet');
    localStorage.removeItem('tokens');
    localStorage.removeItem('lastFetchTimestamp');
  }

  static getAll(): Record<string, string> {
    return Cookies.get();
  }
}

export default CookieManager;

export interface TrackingTokens {
  referrer_token?: string;
  multi_duel_referral_code?: string;
  advertiser_link_token?: string;
}

export function getTrackingTokens(): TrackingTokens {
  const referrer_token = saveToken.get(REFERRAL_TOKEN_QUERY_KEY);
  const advertiser_link_token = saveToken.get(INVITED_BY_QUERY_KEY);
  const multi_duel_referral_code = saveToken.get(
    MULTI_DUEL_REFERRAL_CODE_QUERY_KEY,
  );

  return {
    ...(referrer_token && { referrer_token }),
    ...(multi_duel_referral_code && { multi_duel_referral_code }),
    ...(advertiser_link_token && { advertiser_link_token }),
  };
}
