import { RefCallback, useCallback, useLayoutEffect, useState } from 'react';

type DimensionObject = {
  width?: number;
  height?: number;
  top?: number;
  left?: number;
  x?: number;
  y?: number;
  right?: number;
  bottom?: number;
};

type UseDimensionsArgs = {
  liveMeasure?: boolean;
};

type UseDimensionsHook = [
  RefCallback<HTMLElement | null>,
  DimensionObject,
  HTMLElement | null,
];

function getDimensionObject(node: HTMLElement): DimensionObject {
  return node.getBoundingClientRect() as DOMRect;
}

export const useDimensions = ({
  liveMeasure = true,
}: UseDimensionsArgs = {}): UseDimensionsHook => {
  const [dimensions, setDimensions] = useState<DimensionObject>(
    {} as DimensionObject,
  );
  const [node, setNode] = useState<HTMLElement | null>(null);

  const ref = useCallback((node: HTMLElement | null) => {
    setNode(node);
  }, []);

  /* eslint-disable-next-line @typescript-eslint/ban-ts-comment */
  /* @ts-expect-error */
  useLayoutEffect(() => {
    if (node) {
      const measure = () =>
        window.requestAnimationFrame(() =>
          setDimensions(getDimensionObject(node)),
        );

      measure();

      if (liveMeasure) {
        const resizeObserver = new ResizeObserver(measure);

        resizeObserver.observe(node);

        return () => {
          resizeObserver.disconnect();
        };
      }
    }
  }, [node, liveMeasure]);

  return [ref, dimensions, node];
};
