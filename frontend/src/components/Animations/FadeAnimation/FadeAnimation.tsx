'use client';
import {
  CSSProperties,
  HTMLAttributes,
  ReactNode,
  RefObject,
  useEffect,
  useRef,
  useState,
} from 'react';
import { CSSTransition } from 'react-transition-group';
import { CSSTransitionClassNames } from 'react-transition-group/CSSTransition';

import styles from './styles.module.scss';

interface Props extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
  children: ReactNode;
  isVisible: boolean;
  customClassNames?: CSSTransitionClassNames;
  customTimeout?: number;
  customWrapperClass?: string;
  customDelay?: number;
  onExited?: () => void;
  customRef?: RefObject<HTMLDivElement | null>;
  style?: CSSProperties;
  unmountOnExit?: boolean;
}

export const FadeAnimation = ({
  children,
  isVisible,
  customClassNames,
  unmountOnExit = true,
  customTimeout = 300,
  customWrapperClass,
  customDelay = 0,
  onExited,
  onClick,
  customRef,
  style,
  ...rest
}: Props) => {
  const [inProp, setInProp] = useState(false);
  const refLocal = useRef<HTMLDivElement>(null);

  const ref = customRef || refLocal;

  useEffect(() => {
    if (!customDelay) {
      setInProp(isVisible);
    } else {
      const timeout = setTimeout(() => {
        setInProp(isVisible);
      }, customDelay);

      return () => clearTimeout(timeout);
    }

    return () => {};
  }, [customDelay, isVisible]);

  return (
    <CSSTransition
      in={inProp}
      timeout={customTimeout}
      classNames={
        customClassNames ?? {
          enter: styles.fadeEnter,
          enterActive: styles.fadeEnterActive,
          exit: styles.fadeExit,
          exitActive: styles.fadeExitActive,
        }
      }
      nodeRef={ref}
      onExited={onExited}
      unmountOnExit={unmountOnExit}
    >
      <div
        {...rest}
        onClick={onClick}
        className={customWrapperClass ?? styles.wrapper}
        ref={ref}
        style={
          {
            '--switch-duration': `${customTimeout}ms`,
            ...style,
          } as CSSProperties
        }
      >
        {children}
      </div>
    </CSSTransition>
  );
};
