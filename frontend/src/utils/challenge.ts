import { MentionChallenge } from '~types/leaders';

export function getActiveChallenge(seasons: MentionChallenge[]) {
  if (!seasons || seasons.length === 0) {
    return {
      active: null,
      index: -1,
      list: [],
    };
  }

  const now = Date.now();

  let index = seasons.findIndex(({ start_date, finish_date }) => {
    const start = new Date(start_date).getTime();
    const finish = new Date(finish_date).getTime();

    return now >= start && now < finish;
  });

  const list = [...seasons];

  if (index === -1) {
    const past = list.filter((s) => new Date(s.finish_date).getTime() <= now);

    if (past.length > 0 && past[0]) {
      past.sort(
        (a, b) =>
          new Date(b.finish_date).getTime() - new Date(a.finish_date).getTime(),
      );
      index = list.indexOf(past[0]);
    } else {
      list.sort(
        (a, b) =>
          new Date(a.start_date).getTime() - new Date(b.start_date).getTime(),
      );
      index = 0;
    }
  }

  return {
    active: list[index],
    index,
    list,
  };
}

const ALLOWED_ORIGINS = new Set([
  'https://duelduck.com',
  'https://stage.duelduck.com',
  'http://localhost:3000',
]);

const ALLOWED_STATIC_PATHS = new Set(['/', '/profile', '/mentionboard']);

const TOURNAMENT_BASE_PATH = /^\/tournament\/[a-z0-9-]+$/i;
const TOURNAMENT_TAB_PATH =
  /^\/tournament\/[a-z0-9-]+\/(duels|mentionboard|how-it-works|ranking)$/i;

export function resolveUrl(url = window.location.href) {
  const u = new URL(url);

  u.search = '';
  u.hash = '';

  const normalizedOrigin = u.origin.toLowerCase();
  const normalizedPath = (u.pathname.replace(/\/+$/, '') || '/').toLowerCase();
  const normalizedUrl = `${normalizedOrigin}${normalizedPath}`;

  if (!ALLOWED_ORIGINS.has(normalizedOrigin)) {
    return window.location.origin;
  }

  if (
    ALLOWED_STATIC_PATHS.has(normalizedPath) ||
    TOURNAMENT_BASE_PATH.test(normalizedPath) ||
    TOURNAMENT_TAB_PATH.test(normalizedPath)
  ) {
    return normalizedUrl;
  }

  return window.location.origin;
}

/*
 * Used as the X (Twitter) OAuth redirect_uri. Must be deterministic and
 * match exactly what is registered in the X app dashboard — otherwise the
 * token exchange fails with "redirect uri doesn't match". We always
 * return just the current origin (no path, no trailing slash) so the
 * value is identical whether the user triggers login from `/`, `/duels`,
 * a tournament page, or anywhere else, and identical between the
 * authorize call and the code-exchange call.
 */
export const getMentionLinkRedirect = () => {
  const origin = window.location.origin;

  if (ALLOWED_ORIGINS.has(origin)) return origin;

  return window.location.origin;
};
