import { AnchorHTMLAttributes, forwardRef } from 'react';
import Link, { LinkProps } from 'next/link';

interface SafeLinkProps
  extends
    Omit<AnchorHTMLAttributes<HTMLAnchorElement>, keyof LinkProps>,
    LinkProps {
  ugc?: boolean;
}

function isInternal(href: string): boolean {
  if (typeof href !== 'string') return true;
  if (href.startsWith('/') || href.startsWith('#')) return true;

  try {
    const hostname = new URL(href).hostname;

    return hostname === 'duelduck.com' || hostname.endsWith('.duelduck.com');
  } catch {
    return true;
  }
}

export const SafeLink = forwardRef<HTMLAnchorElement, SafeLinkProps>(
  ({ ugc, href, ...rest }, ref) => {
    if (isInternal(typeof href === 'string' ? href : '/')) {
      return <Link ref={ref} href={href} {...rest} />;
    }

    const relValue = ugc
      ? 'ugc nofollow noopener noreferrer'
      : 'nofollow noopener noreferrer';

    return (
      <Link ref={ref} href={href} rel={relValue} target="_blank" {...rest} />
    );
  },
);

SafeLink.displayName = 'SafeLink';
