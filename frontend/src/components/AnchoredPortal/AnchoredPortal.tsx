import { PropsWithChildren, RefObject } from 'react';
import { createPortal } from 'react-dom';

import { useAnchoredPosition } from '~hooks/useAnchoredPosition';

type PortalProps = PropsWithChildren<{
  anchorRef: RefObject<HTMLElement | null>;
  placement?: 'bottom' | 'top';
  align?: 'start' | 'center' | 'end';
  offset?: number;
  className?: string;
}>;

export function AnchoredPortal({
  anchorRef,
  placement = 'bottom',
  align = 'start',
  offset = 8,
  className,
  children,
}: PortalProps) {
  const { style } = useAnchoredPosition(anchorRef, {
    placement,
    align,
    offset,
  });

  return createPortal(
    <div style={style} className={className} data-anchored-portal>
      {children}
    </div>,
    document.body,
  );
}
