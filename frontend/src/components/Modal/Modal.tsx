import { ReactNode, useEffect, useMemo, useRef, useState } from 'react';
import cx from 'classnames';
import dynamic from 'next/dynamic';

import { useOuterClick } from '~hooks/useOuterClick';
import { Cross } from '~icons/JsxSvg/Cross';
import useModalStore, { MODAL_CONTENT } from '~store/modalStore';

import styles from './styles.module.scss';

const modalLabels: Record<MODAL_CONTENT, string> = {
  [MODAL_CONTENT.LOGIN]: 'Log in',
};

const Login = dynamic(
  () => import('~components/Modal/components/Login').then((m) => m.Login),
  { ssr: true, loading: () => <div /> },
);

export const Modal = () => {
  const { isOpen, content, closeModal } = useModalStore();
  const modalContentRef = useRef(null);

  const contentsMap: Record<MODAL_CONTENT, ReactNode> = {
    [MODAL_CONTENT.LOGIN]: <Login />,
  };

  const [animationClass, setAnimationClass] = useState(styles.hidden);

  const arrayNotCloseOuter: MODAL_CONTENT[] = [];

  useEffect(() => {
    setAnimationClass(isOpen ? styles.active : styles.hidden);
  }, [isOpen]);

  useOuterClick(
    modalContentRef,
    () => {
      if (content && !arrayNotCloseOuter.includes(content)) closeModal();
    },
    'click',
    [content],
  );

  const isDontHaveModalContainer: MODAL_CONTENT[] = [];

  const hideCloseButtonArray: (MODAL_CONTENT | null)[] = [MODAL_CONTENT.LOGIN];

  const maxContentWidthArray: (MODAL_CONTENT | null)[] = [];

  const modalTopArr = [MODAL_CONTENT.LOGIN];

  const modalContainer = useMemo(
    () => (
      <div
        className={cx(styles.modalContent, {
          [styles.maxContent as string]: maxContentWidthArray.includes(content),
        })}
        ref={modalContentRef}
        onClick={(e) => e.stopPropagation()}
      >
        {!hideCloseButtonArray.includes(content) && (
          <button
            className={styles.closeButton}
            onClick={closeModal}
            aria-label="Close modal"
          >
            <Cross />
          </button>
        )}
        {contentsMap[content as keyof typeof MODAL_CONTENT]}
      </div>
    ),
    [content],
  );

  if (!isOpen) return null;

  return (
    <div
      className={cx(styles.modalBackdrop, animationClass, {
        [`${styles['modalBackdrop--modalTop']}`]:
          content && modalTopArr.includes(content),
      })}
      role="dialog"
      aria-modal="true"
      aria-label={content ? modalLabels[content] : 'Dialog'}
      onMouseDown={(e) => {
        if (e.target === e.currentTarget) {
          if (content && !arrayNotCloseOuter.includes(content)) closeModal();
        }
      }}
    >
      {content && isDontHaveModalContainer.includes(content)
        ? contentsMap[content as keyof typeof MODAL_CONTENT]
        : modalContainer}
    </div>
  );
};
