import { ComponentType, SVGProps } from 'react';

import colors from '~styles/colors';

export type Styles = Record<string, string | number>;

export type SVGElementProps = SVGProps<SVGSVGElement>;

export type SVGElement = ComponentType<
  SVGElementProps & { title?: string | undefined }
>;

export enum SIDEBAR_CONTENT {
  SELECT_TYPE_SIGN_IN = 'selectTypeSignIn',
  SIGN_IN_WITH_EMAIL = 'signInWithEmail',
  CONFIRM_EMAIL = 'confirmEmail',
}

export enum CURRENCY {
  DUCKIES,
  USDC,
  WBT,
  BEER_TWO,
  SOLANA,
}

export const currencyNames: Record<CURRENCY, string> = {
  [CURRENCY.DUCKIES]: 'Duckies',
  [CURRENCY.USDC]: 'USDC',
  [CURRENCY.WBT]: 'WBT',
  [CURRENCY.BEER_TWO]: 'BEER2',
  [CURRENCY.SOLANA]: 'SOL',
};

export enum QUERY_PARAMS {
  TOPIC = 'topic',
  CATEGORIES = 'categories',
  SEARCH_QUARY = 'search_quary',
}

export type QueryParameters = {
  [K in QUERY_PARAMS]?: string;
};

export enum CUSTOM_EVENT_KEYS {
  MENU_CLICK = 'menuClick',
  REPLY_FAQ = 'replyFAQ',
  OPEN_POPUP = 'openPopUp',
  CLOSE_POPUP = 'closePopUp',
  OPEN_DONT_HAVE_MONEY = 'openDontsHaveMoney',
  LOGOUT = 'logout',
  SCROLL_TO = 'scrollTo',
  TOGGLE_SCROLL = 'toggleScroll',
  TOGGLE_SEARCH_OPEN = 'toggleSearchOpen',
  ADD_REPLY = 'addReply',
  REMOVE_SEARCH_PARAMS = 'removeSearchParams',
  CREATE_DUEL_CHANGE_OPENED_ITEM = 'createDUEL_CHANGE_OPENED_ITEM',
  CHANGE_ROUTE = 'change_route',
  EXPORT_CARD = 'export_card',
  MOUNT_EXPORT_CARD = 'mount_export_card',
  BOUNCING_GUIDE_CREATE = 'BOUNCING_GUIDE_CREATE',
  CHANGE_DUEL_VOTE = 'change_votes',
  RESET_CREATE_DUEL = 'reset_create_duel',
  HIDE_MAIN_PAGE_DUCKS = 'hide_main_page_ducks',
  LOGIN_FIREBASE = 'login_firebase',
  LOGIN = 'login',
  CONNECT = 'connect',
  CREATE_DUEL_ACTION = 'create_duel_action',
  SWITCH_WALLET = 'switch_wallet',
  REMOVE_WALLET = 'remove_wallet',
  LINK_WALLET = 'link_wallet',
  RELOAD_BALANCES = 'reload_balances',
  RELOAD_MENTIONBOARD_DATA = 'reload_mentionboard_data',
  OPEN_CLAIM_REWARD = 'open_claim_reward',
}

export type ColorType = keyof typeof colors;
