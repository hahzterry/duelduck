import { Typography } from '~components/Typography';
import useMediaQuery from '~hooks/useMediaQuery';
import { AnswerDuck } from '~icons/JsxSvg/AnswerDuck';

import styles from './styles.module.scss';

interface FAQItem {
  question: string;
  answer: string;
}

interface FAQListProps {
  faqs: FAQItem[];
}

export const FAQList = ({ faqs }: FAQListProps) => {
  const { isPhone } = useMediaQuery();

  return (
    <div className={styles.faqList}>
      {faqs.map((faq, index) => (
        <div key={index} className={styles.faqItem}>
          <Typography element="h4" className={styles.question}>
            {faq.question}
          </Typography>
          {!isPhone && (
            <Typography element="p" className={'p2'}>
              Answer
            </Typography>
          )}
          <div className={styles.answer}>
            {!isPhone && <AnswerDuck />}
            <Typography element="p" className={isPhone ? 'p2' : 'p1'}>
              {faq.answer}
            </Typography>
          </div>
          {isPhone && (
            <div className={styles.answer}>
              <AnswerDuck />
              <Typography element="p" className={'p1'}>
                by duelduck
              </Typography>
            </div>
          )}
        </div>
      ))}
    </div>
  );
};
