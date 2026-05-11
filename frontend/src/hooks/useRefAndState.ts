import { useRef, useState } from 'react';

export type GetStateType = 'ref' | 'state';

type SetArg<T> = T | ((prev: T) => T);

export const useRefAndState = <T>(
  initialState: T,
): [(type?: GetStateType) => T, (newState: SetArg<T>) => void, T] => {
  const ref = useRef<T>(initialState);
  const [state, setState] = useState<T>(initialState);

  const setStateAndRef = (newState: SetArg<T>) => {
    if (typeof newState === 'function') {
      const newValue = (newState as (prev: T) => T)(ref.current);

      ref.current = newValue;
      setState(newValue);
    } else {
      ref.current = newState;
      setState(newState);
    }
  };

  const getState: (type?: GetStateType) => T = (type = 'state') =>
    type === 'state' ? state : ref.current;

  return [getState, setStateAndRef, state];
};
