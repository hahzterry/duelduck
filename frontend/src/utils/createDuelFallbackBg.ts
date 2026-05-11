export const CREATE_DUEL_FALLBACK_BGS = [
  '/duel-fallback-bg/june_ticket.webp',
  '/duel-fallback-bg/may_ticket.webp',
  '/duel-fallback-bg/july_ticket.webp',
] as const;

export const pickFallbackBg = (seed: number = Date.now()): string => {
  const list = CREATE_DUEL_FALLBACK_BGS;
  const idx = Math.abs(Math.floor(seed)) % list.length;

  return list[idx] ?? list[0];
};
