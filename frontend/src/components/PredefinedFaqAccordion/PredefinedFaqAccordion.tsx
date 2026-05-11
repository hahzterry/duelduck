'use client';

import { createElement, ReactNode, useEffect, useState } from 'react';
import cx from 'classnames';

import { LazyImage } from '~components/Lazy/LazyImage';
import { Typography } from '~components/Typography';
import { Chevron } from '~icons/Chevrons/Chevron';

export interface PredefinedFaqItem {
  id?: string | number;
  question: string;
  answer: ReactNode;
}

type AccordionStyles = Record<string, string>;

interface PredefinedFaqAccordionProps {
  faqs: PredefinedFaqItem[];
  styles: AccordionStyles;
  title?: string;
  subtitle?: string;
  titleAlign?: 'left' | 'center' | 'right' | 'start';
  sectionId?: string;
  initialOpenIndex?: number | null;
  openItemId?: string | number | null;
  answerIdPrefix?: string;
  renderMode?: 'typography' | 'native';
  answerLabel?: string;
  useHeader?: boolean;
}

const renderText = ({
  renderMode,
  element,
  className,
  text,
  textAlign,
}: {
  renderMode: 'typography' | 'native';
  element: 'h2' | 'h3' | 'p' | 'span';
  className?: string;
  text: string;
  textAlign?: 'left' | 'center' | 'right' | 'start';
}) => {
  if (renderMode === 'typography') {
    return (
      <Typography
        element={element}
        className={className}
        text={text}
        textAlign={textAlign}
      />
    );
  }

  return createElement(
    element,
    {
      className,
      style: textAlign ? { textAlign } : undefined,
    },
    text,
  );
};

export const PredefinedFaqAccordion = ({
  faqs,
  styles,
  title,
  subtitle,
  titleAlign,
  sectionId,
  initialOpenIndex = null,
  openItemId = null,
  answerIdPrefix = 'faq-answer',
  renderMode = 'typography',
  answerLabel = 'Answer',
  useHeader = false,
}: PredefinedFaqAccordionProps) => {
  const getOpenIndex = () => {
    if (openItemId === null || typeof openItemId === 'undefined') {
      return initialOpenIndex;
    }

    const foundIndex = faqs.findIndex(
      (faq) =>
        typeof faq.id !== 'undefined' && String(faq.id) === String(openItemId),
    );

    return foundIndex >= 0 ? foundIndex : initialOpenIndex;
  };

  const [openIndex, setOpenIndex] = useState<number | null>(getOpenIndex);

  useEffect(() => {
    setOpenIndex(getOpenIndex());
  }, [openItemId, faqs, initialOpenIndex]);

  const handleToggle = (index: number) => {
    setOpenIndex((prev) => (prev === index ? null : index));
  };

  return (
    <section className={styles.faq} id={sectionId}>
      {title &&
        (useHeader ? (
          <div className={styles.faq__header}>
            {renderText({
              renderMode,
              element: 'h2',
              className: styles.faq__title,
              text: title,
              textAlign: titleAlign,
            })}
            {subtitle &&
              renderText({
                renderMode,
                element: 'p',
                className: styles.faq__subtitle,
                text: subtitle,
              })}
          </div>
        ) : (
          <>
            {renderText({
              renderMode,
              element: 'h2',
              className: styles.faq__title,
              text: title,
              textAlign: titleAlign,
            })}
            {subtitle &&
              renderText({
                renderMode,
                element: 'p',
                className: styles.faq__subtitle,
                text: subtitle,
              })}
          </>
        ))}

      <div className={styles.faq__list} role="list">
        {faqs.map((item, index) => {
          const isOpen = openIndex === index;
          const answerId = `${answerIdPrefix}-${index}`;

          return (
            <div
              key={index}
              className={cx(styles.faq__item, {
                [`${styles['faq__item--open']}`]: isOpen,
              })}
              role="listitem"
            >
              <button
                className={styles.faq__question}
                onClick={() => handleToggle(index)}
                aria-expanded={isOpen}
                aria-controls={answerId}
              >
                {renderText({
                  renderMode,
                  element: 'h3',
                  className: styles.faq__questionText,
                  text: item.question,
                })}
                <Chevron
                  className={cx(styles.faq__chevron, {
                    [`${styles['faq__chevron--open']}`]: isOpen,
                  })}
                />
              </button>
              <div
                id={answerId}
                role="region"
                aria-hidden={!isOpen}
                className={cx(styles.faq__answer, {
                  [`${styles['faq__answer--open']}`]: isOpen,
                })}
              >
                <div className={styles.faq__answerInner}>
                  {renderText({
                    renderMode,
                    element: renderMode === 'native' ? 'span' : 'p',
                    className: styles.faq__answerLabel,
                    text: answerLabel,
                  })}
                  <div className={styles.faq__answerContent}>
                    <div className={styles.faq__duckWrapper} aria-hidden="true">
                      <LazyImage
                        src="/logo.svg"
                        alt=""
                        width={40}
                        height={40}
                        className={styles.faq__duck}
                      />
                    </div>
                    {typeof item.answer === 'string'
                      ? renderText({
                          renderMode,
                          element: 'p',
                          className: styles.faq__answerText,
                          text: item.answer,
                        })
                      : createElement(
                          'div',
                          {
                            className: styles.faq__answerText,
                          },
                          item.answer,
                        )}
                  </div>
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </section>
  );
};
