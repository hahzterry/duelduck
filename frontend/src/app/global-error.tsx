'use client';

import { useEffect } from 'react';

import { Typography } from '~components/Typography';

export default function GlobalError(props: {
  error: Error & { digest?: string };
}) {
  useEffect(() => {
    console.error('Global error boundary caught:', props.error);
  }, [props]);

  return <Typography text={props.error.message} />;
}
