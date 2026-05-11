import { useCallback, useEffect, useRef, useState } from 'react';
import { usePathname } from 'next/navigation';

import {
  MessagePopup,
  PopupMessageTexts,
} from '~components/Popups/MessagePopup';
import { CUSTOM_EVENT_KEYS } from '~types/general';
import { customEvent } from '~utils/customEvent';

const PopupWrapper = ({
  texts,
  id,
}: {
  texts?: PopupMessageTexts;
  id: string;
}) => {
  const element = useRef<HTMLDivElement | null>(null);
  const requestFrameAnimationRef = useRef(0);

  const updatePosition = useCallback(() => {
    const popup = element.current;

    if (!popup) return;

    const target = document.getElementById(id);

    if (!target) return;

    const rect = target.getBoundingClientRect();

    const top = rect.bottom + window.scrollY;

    let left = rect.left + window.scrollX;

    const popupWidth = popup.offsetWidth || 264;
    const viewportWidth = window.innerWidth;

    if (left < 0) {
      left = 0;
    }

    if (left + popupWidth > viewportWidth) {
      left = viewportWidth - popupWidth;
    }

    popup.style.top = `${top}px`;
    popup.style.left = `${left}px`;
  }, [id]);

  useEffect(() => {
    const requestUpdate = () => {
      updatePosition();

      requestFrameAnimationRef.current = requestAnimationFrame(requestUpdate);
    };

    requestFrameAnimationRef.current = requestAnimationFrame(requestUpdate);

    return () => {
      if (requestFrameAnimationRef.current)
        cancelAnimationFrame(requestFrameAnimationRef.current);
    };
  }, []);

  return <MessagePopup ref={element} texts={texts} isOpen />;
};

export const PopupOnElement = () => {
  const [popups, setPopups] = useState<
    {
      id: string;
      texts: PopupMessageTexts;
    }[]
  >([]);
  const pathname = usePathname();

  useEffect(() => {
    const handleOpen = (id: string, texts: PopupMessageTexts) => {
      setPopups((p) => {
        const newObj = { id, texts };

        return p.findIndex((e) => e.id === id) !== -1
          ? p.map((e) => (e.id === id ? newObj : e))
          : [...p, newObj];
      });
    };

    const handleClose = (id: string) => {
      setPopups((p) => {
        return p.filter((e) => e.id !== id);
      });
    };

    customEvent.on(CUSTOM_EVENT_KEYS.OPEN_POPUP, handleOpen);
    customEvent.on(CUSTOM_EVENT_KEYS.CLOSE_POPUP, handleClose);

    return () => {
      customEvent.off(CUSTOM_EVENT_KEYS.OPEN_POPUP, handleOpen);
      customEvent.off(CUSTOM_EVENT_KEYS.CLOSE_POPUP, handleClose);
    };
  }, []);

  useEffect(() => {
    return () => {
      setPopups([]);
    };
  }, [pathname]);

  return (
    <>
      {popups.map((e, i) => (
        <PopupWrapper id={e.id} texts={e.texts} key={`${e.id}-${i}`} />
      ))}
    </>
  );
};

export const handleOpenPopup = (id: string, texts: PopupMessageTexts) => {
  customEvent.emit(CUSTOM_EVENT_KEYS.OPEN_POPUP, id, texts);
};

export const handleClosePopup = (id: string) => {
  customEvent.emit(CUSTOM_EVENT_KEYS.CLOSE_POPUP, id);
};
