import NextAuth from 'next-auth';
import type { JWT } from 'next-auth/jwt';
import Credentials from 'next-auth/providers/credentials';

import type { AuthData, User } from '~types/auth';

const ACCESS_TOKEN_TTL_MS = 30 * 60 * 1000;
const REFRESH_TOKEN_TTL_MS = 7 * 24 * 60 * 60 * 1000;

type LoginProvider = 'email-code' | 'google-firebase';

type AuthUser = {
  id: string;
  user: User;
  accessToken: string;
  refreshToken: string;
  expiresAt: number;
};

type AuthSession<TSession extends object> = Omit<
  TSession,
  'accessToken' | 'expiresAt' | 'id' | 'refreshToken' | 'user'
> & {
  id: string;
  user: User;
  accessToken?: unknown;
  refreshToken?: unknown;
  expiresAt: number;
};

const getCredentialValue = (value: unknown) =>
  typeof value === 'string' ? value.trim() : '';

const createMockJwtInfo = (provider: LoginProvider) => {
  const now = Date.now();

  return {
    access_token: `mock-${provider}-access-${now}`,
    refresh_token: `mock-${provider}-refresh-${now}`,
    access_exp_time: new Date(now + ACCESS_TOKEN_TTL_MS).toISOString(),
    refresh_exp_time: new Date(now + REFRESH_TOKEN_TTL_MS).toISOString(),
  };
};

const createMockUser = (provider: LoginProvider, email?: string): User => {
  const now = new Date().toISOString();
  const normalizedEmail = email || `${provider}@duelduck.mock`;

  return {
    id: `mock-${provider}-user`,
    telegram_id: '',
    username: provider === 'google-firebase' ? 'Google user' : 'Email user',
    email: normalizedEmail,
    image_url: '',
    bg_url: '',
    bio: '',
    is_verified: true,
    website: '',
    role: 0,
    adult_confirmed: true,
    is_premium: false,
    balance: 0,
    referral_token: 'mock-referral-token',
    daily_reward_streak: 0,
    last_claimed_reward: '',
    created_at: now,
    updated_at: now,
    x_url: '',
    x_id: '',
    youtube_url: '',
    telegram_url: '',
    instagram_url: '',
    discord_url: '',
    public_address: '',
    last_completed_streak: '',
    level: 1,
    current_xp: 0,
    active_wallet_id: '',
  };
};

const createMockAuthData = (
  provider: LoginProvider,
  email?: string,
): AuthData => ({
  daily_reward: {
    day_streak: 0,
    claimed: false,
  },
  jwt_info: createMockJwtInfo(provider),
  user: createMockUser(provider, email),
});

async function fetchMe(token: JWT) {
  return token.user || createMockUser('email-code');
}

async function refreshAccessToken(token: JWT) {
  const now = Date.now();

  return {
    ...token,
    accessToken: `mock-refreshed-access-${now}`,
    refreshToken:
      typeof token.refreshToken === 'string'
        ? token.refreshToken
        : `mock-refresh-${now}`,
    expiresAt: now + ACCESS_TOKEN_TTL_MS,
  };
}

function mapAuthDataToUser(data: AuthData): AuthUser {
  const accessExpTime = Date.parse(data.jwt_info.access_exp_time);

  return {
    id: data.user.id,
    user: data.user,
    accessToken: data.jwt_info.access_token,
    refreshToken: data.jwt_info.refresh_token,
    expiresAt: Number.isFinite(accessExpTime)
      ? accessExpTime
      : Date.now() + ACCESS_TOKEN_TTL_MS,
  };
}

const handler = NextAuth({
  session: { strategy: 'jwt' },

  providers: [
    Credentials({
      id: 'email-code',
      name: 'Email code',
      credentials: {
        email: { type: 'text' },
        code: { type: 'text' },
        referrer_token: { type: 'text' },
        advertiser_link_token: { type: 'text' },
      },
      async authorize(credentials) {
        const email = getCredentialValue(credentials?.email);
        const code = getCredentialValue(credentials?.code);

        if (!email || !code) return null;

        return mapAuthDataToUser(createMockAuthData('email-code', email));
      },
    }),

    Credentials({
      id: 'google-firebase',
      name: 'Google (Firebase)',
      credentials: {
        firebase_id_token: { type: 'text' },
        referrer_token: { type: 'text' },
        advertiser_link_token: { type: 'text' },
      },
      async authorize(credentials) {
        const firebaseIdToken = getCredentialValue(
          credentials?.firebase_id_token,
        );

        if (!firebaseIdToken) return null;

        return mapAuthDataToUser(createMockAuthData('google-firebase'));
      },
    }),
  ],

  callbacks: {
    async jwt({ token, user }) {
      if (user) {
        const authUser = user as AuthUser;

        token.user = authUser.user;
        token.id = authUser.id;
        token.accessToken = authUser.accessToken;
        token.refreshToken = authUser.refreshToken;
        token.expiresAt = authUser.expiresAt;

        return token;
      }

      const expiresAt =
        typeof token.expiresAt === 'number' ? token.expiresAt : 0;

      if (expiresAt && Date.now() < expiresAt - 10_000) return token;

      return refreshAccessToken(token);
    },

    async session({ session, token }) {
      const authSession = session as AuthSession<typeof session>;

      authSession.id = typeof token.id === 'string' ? token.id : '';
      authSession.user = await fetchMe(token);
      authSession.accessToken = token.accessToken;
      authSession.refreshToken = token.refreshToken;
      authSession.expiresAt =
        typeof token.expiresAt === 'number' ? token.expiresAt : 0;

      return authSession;
    },
  },
});

export { handler as GET, handler as POST };
