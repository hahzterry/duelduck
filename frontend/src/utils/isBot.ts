/**
 * Lightweight user-agent test. Used to gate analytics/ads tracking so
 * crawlers don't get flagged as real users and don't waste our ad budget.
 *
 * SSR note: returns true when `navigator` is unavailable (server render) so
 * no tracking script ever runs before the client takes over.
 */
const BOT_RE =
  /bot|crawl|spider|slurp|Googlebot|Bingbot|DuckDuckBot|YandexBot|Baiduspider|GPTBot|ChatGPT-User|ClaudeBot|anthropic-ai|Claude-Web|PerplexityBot|Google-Extended|FacebookBot|facebookexternalhit|LinkedInBot|Applebot|Twitterbot|AhrefsBot|SemrushBot/i;

export const isBot = (userAgent?: string): boolean => {
  const ua =
    userAgent ??
    (typeof navigator !== 'undefined' ? navigator.userAgent : null);

  if (!ua) return true;

  return BOT_RE.test(ua);
};
