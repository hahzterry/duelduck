export const DUCK_POINTS_IN_RATE = 0.0004;

export const WBT_PRICE = 54.5;

export const usdcMint = process.env.NEXT_PUBLIC_MINT_ADDRESS as string;

export const solanaMint = process.env.NEXT_PUBLIC_SOLANA_ADDRESS as string;

export const LINK_FAQ_SELF_RESOLVE =
  'https://duelduck.com/blog/duelduck-resolution-and-reputation-system-how-fair-outcomes-are-enforced';

export const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;

export const regexLatinAndSymbols = /^[a-zA-Z\p{N}\p{P}\p{S}\p{Zs}]*$/u;

export const telegramLinkRegex =
  /^(https:\/\/)?(www\.)?(t\.me|telegram\.me)(\/.*)?$/;

export const discordLinkReportBug = 'https://discord.gg/bDdTRheD2a';

export const cloudflareCaptchaSiteKey = process.env
  .NEXT_PUBLIC_CLOUD_FLARE_CAPTCHA_SITE_KEY as string;
