export enum USER_PROFILE_TAB {
  PROFILE = 'profile',
  NOTIFICATION = 'notification',
  AMBASSADOR = 'ambassador',
  REFERRAL_PROGRAM = 'referral-program',
  MY_TOURNAMENTS = 'my-tournaments',
  WALLETS = 'wallets',
}

export enum USER_PROFILE_DUELS_TAB {
  CREATED = 'created',
  PLAYING = 'playing',
  ENDED = 'ended',
  REFUNDED = 'refunded',
}

export const PROFILE_TAB_QUERY_PARAM = 'tab';

const profileTabs = new Set<string>(Object.values(USER_PROFILE_TAB));
const profileDuelsTabs = new Set<string>(Object.values(USER_PROFILE_DUELS_TAB));

export const isUserProfileTab = (value: string): value is USER_PROFILE_TAB =>
  profileTabs.has(value);

export const isUserProfileDuelTab = (
  value: string,
): value is USER_PROFILE_DUELS_TAB => profileDuelsTabs.has(value);

export const getUserProfileTabHref = (
  tab: USER_PROFILE_TAB | USER_PROFILE_DUELS_TAB,
): string => {
  if (tab === USER_PROFILE_TAB.PROFILE) return '/profile';

  return `/profile/${encodeURIComponent(tab)}`;
};
