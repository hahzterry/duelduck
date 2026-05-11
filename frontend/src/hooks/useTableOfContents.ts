import { RefObject, useEffect, useMemo, useState } from 'react';

import { TocLink } from '~components/TableOfContents/TableOfContents';

const useTableOfContents = (
  root: RefObject<HTMLElement | null>,
  selector: string,
): [TocLink[], string | null] => {
  const [elements, setElements] = useState<HTMLElement[]>([]);
  const [activeLinkId, setActiveLinkId] = useState<string | null>(null);

  const links = useMemo<(TocLink & { element: HTMLElement })[]>(() => {
    return elements.map((element) => {
      const label = element.textContent;
      const id = label
        .toLowerCase()
        .replace(/[^a-z\s]/g, '')
        .trim()
        .replace(/\s/g, '-');

      return { label, id, element };
    });
  }, [elements]);

  useEffect(() => {
    if (!root.current) {
      return;
    }

    setElements(Array.from(root.current.querySelectorAll(selector)));
  }, [root, selector]);

  useEffect(() => {
    links.forEach((link) => (link.element.id = link.id));
  }, [links]);

  useEffect(() => {
    let requestedAnimationFrame: number | null = null;

    const handleScroll = () => {
      if (requestedAnimationFrame) {
        return;
      }

      requestedAnimationFrame = requestAnimationFrame(() => {
        let closestToViewportCenterElement: HTMLElement | null = null;
        let smallestViewportOffset = Infinity;

        for (const link of links) {
          const viewportOffset = Math.abs(
            link.element.getBoundingClientRect().top - window.innerHeight / 4,
          );

          if (viewportOffset < smallestViewportOffset) {
            closestToViewportCenterElement = link.element;
            smallestViewportOffset = viewportOffset;
          }
        }

        setActiveLinkId(closestToViewportCenterElement?.id ?? null);
        requestedAnimationFrame = null;
      });
    };

    document.body.addEventListener('scroll', handleScroll);

    return () => {
      document.body.removeEventListener('scroll', handleScroll);

      if (requestedAnimationFrame) {
        cancelAnimationFrame(requestedAnimationFrame);
      }
    };
  }, [links]);

  return [links, activeLinkId];
};

export default useTableOfContents;
