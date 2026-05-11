import { ReactNode } from 'react';

import { LayoutWrapper } from '~components/Layout/LayoutWrapper';

export default async function RootLayout({
  children,
}: {
  children: ReactNode;
}) {
  return <LayoutWrapper>{children}</LayoutWrapper>;
}
