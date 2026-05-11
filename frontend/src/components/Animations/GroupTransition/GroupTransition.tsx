'use client';
import {
  ComponentType,
  createRef,
  CSSProperties,
  MutableRefObject,
  useRef,
} from 'react';
import { CSSTransition, TransitionGroup } from 'react-transition-group';

import styles from './styles.module.scss';

interface Props<T> {
  data: T[];
  getKey: (item: T, index: number) => string;
  timeout?: number;
  CardComponent: ComponentType<{ item: T; index: number }>;
  getRowStyle?: (item: T, index: number) => CSSProperties;
  classNames?: {
    appear?: string;
    appearActive?: string;
    appearDone?: string;
    enter?: string;
    enterActive?: string;
    enterDone?: string;
    exit?: string;
    exitActive?: string;
    exitDone?: string;
  };
  unmountOnExit?: boolean;
}

export function GroupTransition<T>({
  data,
  getKey,
  CardComponent,
  timeout,
  classNames = {
    enter: styles.duelCardEnter,
    enterActive: styles.duelCardEnterActive,
    exit: styles.duelCardExit,
    exitActive: styles.duelCardExitActive,
  },
  getRowStyle,
  unmountOnExit = true,
}: Props<T>) {
  const cardRefs = useRef<
    Record<string, MutableRefObject<HTMLDivElement | null>>
  >({});

  return (
    <TransitionGroup component={null}>
      {data.map((item, index) => {
        const key = getKey(item, index);

        if (!cardRefs.current[key]) {
          cardRefs.current[key] = createRef<HTMLDivElement>();
        }

        return (
          <CSSTransition
            key={key}
            timeout={timeout ?? 700}
            classNames={classNames}
            appear={false}
            unmountOnExit={unmountOnExit}
            nodeRef={cardRefs.current[key]}
          >
            <div
              ref={cardRefs.current[key]}
              style={{
                height: 'max-content',
                transitionDuration: `${timeout}ms`,
                width: '100%',
                ...getRowStyle?.(item, index),
              }}
            >
              <CardComponent item={item} index={index} />
            </div>
          </CSSTransition>
        );
      })}
    </TransitionGroup>
  );
}
