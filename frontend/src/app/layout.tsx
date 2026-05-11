import { ReactNode } from 'react';
import type { Metadata, Viewport } from 'next';
import { headers } from 'next/headers';

import { Layout } from '~components/Layout';
import { IS_STAGE } from '~constants/api';
import { DeviceHintsProvider } from '~contexts/deviceHints';
import { bungee, poppins, spaceMono } from '~fonts/fonts';

import { AuthProvider } from '../components/Layout/components/AuthProvider';

import { SerwistProvider } from './serwist';

const baseMetadata: Metadata = {
  title: '',
  description: '',
};

export async function generateMetadata(): Promise<Metadata> {
  const h = await headers();
  const host = (h.get('host') || '').toLowerCase();
  const isStageHost = host.startsWith('stage.');

  if (isStageHost || IS_STAGE) {
    return {
      ...baseMetadata,
    };
  }

  return {
    ...baseMetadata,
  };
}

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  // maximumScale: 1,
  // userScalable: false,
};

export default async function RootLayout({
  children,
}: {
  children: ReactNode;
}) {
  const h = await headers();
  const deviceHints = h.get('x-device-hints');

  return (
    <html
      lang="en"
      suppressHydrationWarning
      className={`${poppins.variable} ${bungee.variable} ${spaceMono.variable}`}
    >
      <body suppressHydrationWarning>
        <SerwistProvider disable={IS_STAGE} swUrl="/serwist/sw.js">
          <AuthProvider>
            <DeviceHintsProvider hints={deviceHints}>
              <Layout>
                <div suppressHydrationWarning id="root">
                  <main id="main-content">{children}</main>
                </div>
              </Layout>
            </DeviceHintsProvider>
          </AuthProvider>
        </SerwistProvider>
      </body>
    </html>
  );
}
