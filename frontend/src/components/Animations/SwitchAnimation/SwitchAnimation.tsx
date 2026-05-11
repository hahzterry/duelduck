'use client';
import {
  CSSProperties,
  MouseEventHandler,
  ReactNode,
  Ref,
  useEffect,
  useRef,
  useState,
} from 'react';
import { CSSTransition, SwitchTransition } from 'react-transition-group';
import { CSSTransitionClassNames } from 'react-transition-group/CSSTransition';

import styles from './styles.module.scss';

interface Props {
  children: ReactNode;
  switchKey: string;
  customClassNames?: CSSTransitionClassNames;
  customTimeout?: number;
  customWrapperClass?: string;
  onClick?: MouseEventHandler<HTMLDivElement>;
  onMouseEnter?: () => void;
  onMouseLeave?: () => void;
  ref?: Ref<HTMLDivElement>;
  mode?: 'out-in' | 'in-out';
  style?: CSSProperties;
  id?: string;
}

export const SwitchAnimation = ({
  children,
  switchKey,
  customClassNames = {
    enter: styles.fadeEnter,
    enterActive: styles.fadeEnterActive,
    exit: styles.fadeExit,
    exitActive: styles.fadeExitActive,
  },
  style,
  customTimeout = 300,
  customWrapperClass = '',
  onClick,
  onMouseLeave,
  onMouseEnter,
  ref,
  mode = 'out-in',
  id,
}: Props) => {
  const [isMounted, setIsMounted] = useState(false);
  const internalRef = useRef(null);
  const finalRef = ref || internalRef;

  useEffect(() => {
    setIsMounted(true);
  }, []);

  if (!isMounted) {
    return (
      <div
        style={
          {
            '--switch-duration': `${customTimeout}ms`,
            ...style,
          } as CSSProperties
        }
        id={id}
        className={customWrapperClass}
      >
        {children}
      </div>
    );
  }

  return (
    <SwitchTransition mode={mode}>
      <CSSTransition
        key={switchKey}
        nodeRef={finalRef}
        timeout={customTimeout}
        classNames={customClassNames}
        unmountOnExit
      >
        <div
          onMouseLeave={onMouseLeave}
          style={
            {
              '--switch-duration': `${customTimeout}ms`,
              ...style,
            } as CSSProperties
          }
          onMouseEnter={onMouseEnter}
          id={id}
          ref={finalRef}
          onClick={onClick}
          className={customWrapperClass}
        >
          {children}
        </div>
      </CSSTransition>
    </SwitchTransition>
  );
};
