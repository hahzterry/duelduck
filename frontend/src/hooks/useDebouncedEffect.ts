import { DependencyList, useCallback, useEffect } from 'react';

type EffectFunction = () => void;

export const useDebouncedEffect = (
  fn: EffectFunction,
  deps: DependencyList,
  time: number,
) => {
  const stableFn = useCallback(fn, deps);

  useEffect(() => {
    const timeout = setTimeout(stableFn, time);

    return () => {
      clearTimeout(timeout);
    };
  }, [time, ...deps]);
};
