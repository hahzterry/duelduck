'use client';
import dynamic from 'next/dynamic';

import LoaderSuspense from '~components/LoaderSuspense';

const NotFound = dynamic(
  () => import('~components/PageContents/NotFound/NotFound'),
  {
    loading: () => <LoaderSuspense />,
    ssr: true,
  },
);

export default NotFound;
