import type { User as DuelDuckUser } from '~types/auth';

declare module 'next-auth' {
  interface Session {
    id: string;
    user: DuelDuckUser;
    accessToken?: unknown;
    refreshToken?: unknown;
    expiresAt: number;
  }

  interface User {
    user?: DuelDuckUser;
    accessToken?: string;
    refreshToken?: string;
    expiresAt?: number;
  }
}

declare module 'next-auth/jwt' {
  interface JWT {
    id?: string;
    user?: DuelDuckUser;
    accessToken?: unknown;
    refreshToken?: unknown;
    expiresAt?: number;
  }
}
